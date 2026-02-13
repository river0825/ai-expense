package flex

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildExpenseBubble_SingleExpense(t *testing.T) {
	expenses := []ExpenseData{
		{
			Description:  "午餐 便當",
			HomeAmount:   85,
			HomeCurrency: "TWD",
			Category:     "餐飲",
			Account:      "現金",
			Date:         time.Date(2024, 2, 10, 0, 0, 0, 0, time.UTC),
		},
	}

	bubble := BuildExpenseBubble(expenses, 85, "TWD", "zh-TW", "")

	require.NotNil(t, bubble)
	assert.Equal(t, "bubble", bubble["type"])
	assert.NotNil(t, bubble["header"])
	assert.NotNil(t, bubble["body"])
	assert.NotNil(t, bubble["footer"])
}

func TestBuildExpenseBubble_MultiCurrency(t *testing.T) {
	expenses := []ExpenseData{
		{
			Description:      "Coffee",
			HomeAmount:       150,
			HomeCurrency:     "TWD",
			OriginalAmount:   5,
			OriginalCurrency: "USD",
			Category:         "Food",
			Account:          "Credit Card",
			Date:             time.Date(2024, 2, 10, 0, 0, 0, 0, time.UTC),
		},
	}

	bubble := BuildExpenseBubble(expenses, 150, "TWD", "en", "")

	require.NotNil(t, bubble)
	assert.Equal(t, "bubble", bubble["type"])
}

func TestBuildExpenseBubble_MultipleExpenses(t *testing.T) {
	expenses := []ExpenseData{
		{Description: "Lunch", HomeAmount: 85, HomeCurrency: "TWD", Category: "Food", Account: "Cash", Date: time.Date(2024, 2, 10, 0, 0, 0, 0, time.UTC)},
		{Description: "Train", HomeAmount: 1065, HomeCurrency: "TWD", Category: "Transport", Account: "Credit", Date: time.Date(2024, 2, 10, 0, 0, 0, 0, time.UTC)},
		{Description: "Coffee", HomeAmount: 100, HomeCurrency: "TWD", Category: "Food", Account: "Cash", Date: time.Date(2024, 2, 10, 0, 0, 0, 0, time.UTC)},
	}

	bubble := BuildExpenseBubble(expenses, 1250, "TWD", "zh-TW", "")

	require.NotNil(t, bubble)
	body := bubble["body"].(map[string]interface{})
	contents := body["contents"].([]interface{})
	// 3 expenses with separators between: expense1, sep, expense2, sep, expense3 + detail rows
	assert.True(t, len(contents) > 3)
}

func TestBuildReportBubble(t *testing.T) {
	bubble := BuildReportBubble("https://example.com/report/abc123", "zh-TW")

	require.NotNil(t, bubble)
	assert.Equal(t, "bubble", bubble["type"])
	assert.NotNil(t, bubble["header"])
	assert.NotNil(t, bubble["body"])
	assert.NotNil(t, bubble["footer"])
}

func TestBuildReportBubble_English(t *testing.T) {
	bubble := BuildReportBubble("https://example.com/report/abc123", "en")

	require.NotNil(t, bubble)
	assert.Equal(t, "bubble", bubble["type"])
}

func TestBuildErrorBubble(t *testing.T) {
	bubble := BuildErrorBubble("訊息中未偵測到任何支出", "請嘗試輸入如：「午餐 便當 85元」", "zh-TW")

	require.NotNil(t, bubble)
	assert.Equal(t, "bubble", bubble["type"])
	assert.NotNil(t, bubble["header"])
	assert.NotNil(t, bubble["body"])
}

func TestBuildInfoBubble(t *testing.T) {
	bubble := BuildInfoBubble("訊息中未偵測到任何支出", "請嘗試輸入如：「午餐 便當 85元」", "zh-TW")

	require.NotNil(t, bubble)
	assert.Equal(t, "bubble", bubble["type"])
}

func TestBuildErrorBubble_NoHint(t *testing.T) {
	bubble := BuildErrorBubble("Something went wrong", "", "en")

	require.NotNil(t, bubble)
	body := bubble["body"].(map[string]interface{})
	contents := body["contents"].([]interface{})
	assert.Len(t, contents, 1) // Only message, no hint
}

// Tests for Edit Button Feature

func TestBuildExpenseBubble_WithEditButton(t *testing.T) {
	expenses := []ExpenseData{
		{
			ID:           "exp_123",
			Description:  "Lunch",
			HomeAmount:   85,
			HomeCurrency: "TWD",
			Category:     "Food",
			Account:      "Cash",
			Date:         time.Date(2024, 2, 10, 0, 0, 0, 0, time.UTC),
		},
	}

	bubble := BuildExpenseBubble(expenses, 85, "TWD", "en", "https://dashboard.example.com")

	require.NotNil(t, bubble)
	body := bubble["body"].(map[string]interface{})
	contents := body["contents"].([]interface{})

	// Find the expense item box
	var expenseBox map[string]interface{}
	for _, item := range contents {
		if box, ok := item.(map[string]interface{}); ok && box["type"] == "box" {
			expenseBox = box
			break
		}
	}

	require.NotNil(t, expenseBox, "Should have expense box")
	expenseContents := expenseBox["contents"].([]interface{})

	// Check if edit button exists
	var hasEditButton bool
	for _, content := range expenseContents {
		if btn, ok := content.(map[string]interface{}); ok {
			if btn["type"] == "button" {
				action := btn["action"].(map[string]interface{})
				assert.Equal(t, "uri", action["type"])
				assert.Equal(t, "Edit", action["label"])
				assert.Contains(t, action["uri"], "https://dashboard.example.com")
				assert.Contains(t, action["uri"], "?edit=exp_123")
				hasEditButton = true
				break
			}
		}
	}

	assert.True(t, hasEditButton, "Edit button should exist in expense box")
}

func TestBuildExpenseBubble_EditButtonURL(t *testing.T) {
	expenses := []ExpenseData{
		{
			ID:           "exp_abc123",
			Description:  "Coffee",
			HomeAmount:   50,
			HomeCurrency: "TWD",
		},
	}

	bubble := BuildExpenseBubble(expenses, 50, "TWD", "en", "https://prod.example.com")

	body := bubble["body"].(map[string]interface{})
	contents := body["contents"].([]interface{})

	// Extract button URI
	var buttonURI string
	for _, item := range contents {
		if box, ok := item.(map[string]interface{}); ok && box["type"] == "box" {
			expenseContents := box["contents"].([]interface{})
			for _, content := range expenseContents {
				if btn, ok := content.(map[string]interface{}); ok && btn["type"] == "button" {
					action := btn["action"].(map[string]interface{})
					buttonURI = action["uri"].(string)
					break
				}
			}
		}
	}

	assert.Equal(t, "https://prod.example.com/dashboard/expenses?edit=exp_abc123", buttonURI)
}

func TestBuildExpenseBubble_MultipleExpensesHaveEditButtons(t *testing.T) {
	expenses := []ExpenseData{
		{ID: "exp_1", Description: "Lunch", HomeAmount: 85, HomeCurrency: "TWD"},
		{ID: "exp_2", Description: "Coffee", HomeAmount: 50, HomeCurrency: "TWD"},
	}

	bubble := BuildExpenseBubble(expenses, 135, "TWD", "en", "https://dashboard.example.com")

	body := bubble["body"].(map[string]interface{})
	contents := body["contents"].([]interface{})

	editButtonCount := 0
	for _, item := range contents {
		if box, ok := item.(map[string]interface{}); ok && box["type"] == "box" {
			expenseContents := box["contents"].([]interface{})
			for _, content := range expenseContents {
				if btn, ok := content.(map[string]interface{}); ok && btn["type"] == "button" {
					editButtonCount++
				}
			}
		}
	}

	assert.Equal(t, 2, editButtonCount, "Should have edit button for each expense")
}

func TestBuildExpenseBubble_EditButtonInternationalization(t *testing.T) {
	testCases := []struct {
		locale       string
		expectedText string
	}{
		{"en", "Edit"},
		{"zh-TW", "編輯"},
		{"ja", "編集"},
		{"es", "Editar"},
	}

	for _, tc := range testCases {
		t.Run(tc.locale, func(t *testing.T) {
			expenses := []ExpenseData{
				{ID: "exp_123", Description: "Test", HomeAmount: 100, HomeCurrency: "TWD"},
			}

			bubble := BuildExpenseBubble(expenses, 100, "TWD", tc.locale, "https://example.com")

			body := bubble["body"].(map[string]interface{})
			contents := body["contents"].([]interface{})

			var buttonLabel string
			for _, item := range contents {
				if box, ok := item.(map[string]interface{}); ok && box["type"] == "box" {
					expenseContents := box["contents"].([]interface{})
					for _, content := range expenseContents {
						if btn, ok := content.(map[string]interface{}); ok && btn["type"] == "button" {
							action := btn["action"].(map[string]interface{})
							buttonLabel = action["label"].(string)
							break
						}
					}
				}
			}

			assert.Equal(t, tc.expectedText, buttonLabel)
		})
	}
}

func TestBuildExpenseBubble_NoEditButtonWhenNoID(t *testing.T) {
	expenses := []ExpenseData{
		{
			ID:           "", // No ID
			Description:  "Lunch",
			HomeAmount:   85,
			HomeCurrency: "TWD",
		},
	}

	bubble := BuildExpenseBubble(expenses, 85, "TWD", "en", "https://dashboard.example.com")

	body := bubble["body"].(map[string]interface{})
	contents := body["contents"].([]interface{})

	// Check that no edit button exists
	hasEditButton := false
	for _, item := range contents {
		if box, ok := item.(map[string]interface{}); ok && box["type"] == "box" {
			expenseContents := box["contents"].([]interface{})
			for _, content := range expenseContents {
				if btn, ok := content.(map[string]interface{}); ok && btn["type"] == "button" {
					hasEditButton = true
					break
				}
			}
		}
	}

	assert.False(t, hasEditButton, "Edit button should NOT exist when expense has no ID")
}

func TestBuildExpenseBubble_NoEditButtonWhenNoDashboardURL(t *testing.T) {
	expenses := []ExpenseData{
		{
			ID:           "exp_123",
			Description:  "Lunch",
			HomeAmount:   85,
			HomeCurrency: "TWD",
		},
	}

	bubble := BuildExpenseBubble(expenses, 85, "TWD", "en", "") // Empty dashboard URL

	body := bubble["body"].(map[string]interface{})
	contents := body["contents"].([]interface{})

	// Check that no edit button exists
	hasEditButton := false
	for _, item := range contents {
		if box, ok := item.(map[string]interface{}); ok && box["type"] == "box" {
			expenseContents := box["contents"].([]interface{})
			for _, content := range expenseContents {
				if btn, ok := content.(map[string]interface{}); ok && btn["type"] == "button" {
					hasEditButton = true
					break
				}
			}
		}
	}

	assert.False(t, hasEditButton, "Edit button should NOT exist when dashboard URL is empty")
}
