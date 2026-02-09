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

	bubble := BuildExpenseBubble(expenses, 85, "TWD", "zh-TW")

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

	bubble := BuildExpenseBubble(expenses, 150, "TWD", "en")

	require.NotNil(t, bubble)
	assert.Equal(t, "bubble", bubble["type"])
}

func TestBuildExpenseBubble_MultipleExpenses(t *testing.T) {
	expenses := []ExpenseData{
		{Description: "Lunch", HomeAmount: 85, HomeCurrency: "TWD", Category: "Food", Account: "Cash", Date: time.Date(2024, 2, 10, 0, 0, 0, 0, time.UTC)},
		{Description: "Train", HomeAmount: 1065, HomeCurrency: "TWD", Category: "Transport", Account: "Credit", Date: time.Date(2024, 2, 10, 0, 0, 0, 0, time.UTC)},
		{Description: "Coffee", HomeAmount: 100, HomeCurrency: "TWD", Category: "Food", Account: "Cash", Date: time.Date(2024, 2, 10, 0, 0, 0, 0, time.UTC)},
	}

	bubble := BuildExpenseBubble(expenses, 1250, "TWD", "zh-TW")

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
