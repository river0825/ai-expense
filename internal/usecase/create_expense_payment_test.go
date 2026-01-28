package usecase

import (
	"context"
	"testing"
	"time"
)

func TestCreateExpense_DefaultAccount(t *testing.T) {
	expenseRepo := NewMockExpenseRepository()
	categoryRepo := NewMockCategoryRepository()
	// Other repos can be nil for this test
	uc := NewCreateExpenseUseCase(expenseRepo, categoryRepo, nil, nil, nil, nil, NewMockAIService())

	ctx := context.Background()
	userID := "user123"

	t.Run("Default to Cash", func(t *testing.T) {
		req := &CreateRequest{
			UserID:      userID,
			Description: "Breakfast",
			Amount:      10.0,
			Date:        time.Now(),
			// Account is empty
		}

		resp, err := uc.Execute(ctx, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		expense, _ := expenseRepo.GetByID(ctx, resp.ID)
		if expense.Account != "Cash" {
			t.Errorf("expected account 'Cash', got '%s'", expense.Account)
		}
	})

	t.Run("Preserve explicit value", func(t *testing.T) {
		req := &CreateRequest{
			UserID:      userID,
			Description: "Lunch",
			Amount:      20.0,
			Date:        time.Now(),
			Account:     "Credit Card",
		}

		resp, err := uc.Execute(ctx, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		expense, _ := expenseRepo.GetByID(ctx, resp.ID)
		if expense.Account != "Credit Card" {
			t.Errorf("expected account 'Credit Card', got '%s'", expense.Account)
		}
	})
}
