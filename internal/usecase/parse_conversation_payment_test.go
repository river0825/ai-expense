package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/riverlin/aiexpense/internal/ai"
	"github.com/riverlin/aiexpense/internal/domain"
)

// MockAIForPayment implements ai.Service for testing payment method logic
type MockAIForPayment struct {
	Response *ai.ParseExpenseResponse
}

func (m *MockAIForPayment) ParseExpense(ctx context.Context, text string, userID string) (*ai.ParseExpenseResponse, error) {
	return m.Response, nil
}

func (m *MockAIForPayment) SuggestCategory(ctx context.Context, description string, userID string) (*ai.SuggestCategoryResponse, error) {
	return nil, nil // Not used in this test
}

func TestParseConversation_DefaultAccount(t *testing.T) {
	mockAI := &MockAIForPayment{
		Response: &ai.ParseExpenseResponse{
			Expenses: []*domain.ParsedExpense{
				{
					Description:       "Lunch",
					Amount:            100,
					SuggestedCategory: "Food",
					Date:              time.Now(),
					Account:           "", // Empty, expecting default to "Cash"
				},
				{
					Description:       "Gas",
					Amount:            1000,
					SuggestedCategory: "Transport",
					Date:              time.Now(),
					Account:           "Credit Card", // Explicit, should remain
				},
			},
			Tokens: &ai.TokenMetadata{TotalTokens: 10},
		},
	}

	// Create UseCase with mock AI
	// Repositories can be nil for this logic test
	uc := NewParseConversationUseCase(mockAI, nil, nil, "gemini", "test-model")

	result, err := uc.Execute(context.TODO(), "Conversational text", "user123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Expenses) != 2 {
		t.Fatalf("expected 2 expenses, got %d", len(result.Expenses))
	}

	// Test Case 1: Default to Cash
	if result.Expenses[0].Account != "Cash" {
		t.Errorf("Expense 1: expected account 'Cash' (default), got '%s'", result.Expenses[0].Account)
	}

	// Test Case 2: Preserve explicit value
	if result.Expenses[1].Account != "Credit Card" {
		t.Errorf("Expense 2: expected account 'Credit Card', got '%s'", result.Expenses[1].Account)
	}
}
