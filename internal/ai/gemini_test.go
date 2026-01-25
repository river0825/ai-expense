package ai

import (
	"context"
	"testing"

	"github.com/riverlin/aiexpense/internal/domain"
)

func TestParseExpenseRegex(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectCount int
		expectFirst *domain.ParsedExpense
	}{
		{
			name:        "single expense",
			input:       "早餐$20",
			expectCount: 1,
			expectFirst: &domain.ParsedExpense{
				Description: "早餐",
				Amount:      20,
			},
		},
		{
			name:        "multiple expenses",
			input:       "早餐$20午餐$30加油$200",
			expectCount: 3,
			expectFirst: &domain.ParsedExpense{
				Description: "早餐",
				Amount:      20,
			},
		},
		{
			name:        "decimal amount",
			input:       "咖啡$3.50",
			expectCount: 1,
			expectFirst: &domain.ParsedExpense{
				Description: "咖啡",
				Amount:      3.50,
			},
		},
		{
			name:        "no expenses",
			input:       "random text",
			expectCount: 0,
		},
		{
			name:        "mixed with spaces",
			input:       "早餐 $20 午餐 $30",
			expectCount: 2,
			expectFirst: &domain.ParsedExpense{
				Description: "早餐",
				Amount:      20,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ai := &GeminiAI{apiKey: "test"}
			expenses, err := ai.parseExpenseRegex(tt.input)

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(expenses) != tt.expectCount {
				t.Errorf("expected %d expenses, got %d", tt.expectCount, len(expenses))
			}

			if tt.expectCount > 0 && tt.expectFirst != nil {
				if expenses[0].Description != tt.expectFirst.Description {
					t.Errorf("expected description %q, got %q", tt.expectFirst.Description, expenses[0].Description)
				}
				if expenses[0].Amount != tt.expectFirst.Amount {
					t.Errorf("expected amount %f, got %f", tt.expectFirst.Amount, expenses[0].Amount)
				}
			}
		})
	}
}

func TestSuggestCategoryKeywords(t *testing.T) {
	tests := []struct {
		name             string
		description      string
		expectedCategory string
	}{
		{
			name:             "breakfast",
			description:      "早餐",
			expectedCategory: "Food",
		},
		{
			name:             "lunch",
			description:      "午餐",
			expectedCategory: "Food",
		},
		{
			name:             "gas",
			description:      "加油",
			expectedCategory: "Transport",
		},
		{
			name:             "taxi",
			description:      "計程車",
			expectedCategory: "Transport",
		},
		{
			name:             "clothes",
			description:      "衣服",
			expectedCategory: "Shopping",
		},
		{
			name:             "movie",
			description:      "電影",
			expectedCategory: "Entertainment",
		},
		{
			name:             "unknown",
			description:      "隨機東西",
			expectedCategory: "Other",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ai := &GeminiAI{}
			category := ai.suggestCategoryKeywords(tt.description)

			if category != tt.expectedCategory {
				t.Errorf("expected %q, got %q", tt.expectedCategory, category)
			}
		})
	}
}

func TestNewGeminiAI(t *testing.T) {
	tests := []struct {
		name      string
		apiKey    string
		shouldErr bool
	}{
		{
			name:      "valid api key",
			apiKey:    "test_key_123",
			shouldErr: false,
		},
		{
			name:      "empty api key",
			apiKey:    "",
			shouldErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewGeminiAI(tt.apiKey, "", nil)

			if (err != nil) != tt.shouldErr {
				t.Errorf("expected error: %v, got: %v", tt.shouldErr, err)
			}
		})
	}
}

func TestParseExpense(t *testing.T) {
	ai := &GeminiAI{apiKey: "test"}
	ctx := context.Background()

	text := "早餐$20午餐$30"
	resp, err := ai.ParseExpense(ctx, text, "test_user")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(resp.Expenses) != 2 {
		t.Errorf("expected 2 expenses, got %d", len(resp.Expenses))
	}

	if resp.Expenses[0].Description != "早餐" {
		t.Errorf("expected first description 早餐, got %s", resp.Expenses[0].Description)
	}

	if resp.Expenses[0].Amount != 20 {
		t.Errorf("expected first amount 20, got %f", resp.Expenses[0].Amount)
	}

	// Fallback (regex) should return zero tokens
	if resp.Tokens.TotalTokens != 0 {
		t.Errorf("expected 0 tokens for fallback, got %d", resp.Tokens.TotalTokens)
	}
}

func TestSuggestCategory(t *testing.T) {
	ai := &GeminiAI{apiKey: "test"}
	ctx := context.Background()

	resp, err := ai.SuggestCategory(ctx, "早餐咖啡", "test_user")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.Category != "Food" {
		t.Errorf("expected Food, got %s", resp.Category)
	}

	// Keyword matching returns zero tokens (no API call)
	if resp.Tokens.TotalTokens != 0 {
		t.Errorf("expected 0 tokens for keyword match, got %d", resp.Tokens.TotalTokens)
	}
}
