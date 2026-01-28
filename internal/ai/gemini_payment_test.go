package ai

import (
	"testing"
)

func TestParseGeminiResponseText_Account(t *testing.T) {
	jsonText := `[
		{
			"description": "Lunch",
			"amount": 200,
			"suggested_category": "Food",
			"date": "2023-01-01",
			"account": "Cash"
		},
		{
			"description": "Gas",
			"amount": 1500,
			"suggested_category": "Transport",
			"account": "Credit Card"
		}
	]`

	expenses, err := parseGeminiResponseText(jsonText)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(expenses) != 2 {
		t.Fatalf("expected 2 expenses, got %d", len(expenses))
	}

	// Test Case 1: Cash
	if expenses[0].Account != "Cash" {
		t.Errorf("Expense 1: expected Account 'Cash', got '%s'", expenses[0].Account)
	}

	// Test Case 2: Credit Card
	if expenses[1].Account != "Credit Card" {
		t.Errorf("Expense 2: expected Account 'Credit Card', got '%s'", expenses[1].Account)
	}
}
