package usecase

import (
	"context"
	"fmt"

	"github.com/riverlin/aiexpense/internal/domain"
)

// DeleteExpenseUseCase handles deleting expenses
type DeleteExpenseUseCase struct {
	expenseRepo domain.ExpenseRepository
}

// NewDeleteExpenseUseCase creates a new delete expense use case
func NewDeleteExpenseUseCase(
	expenseRepo domain.ExpenseRepository,
) *DeleteExpenseUseCase {
	return &DeleteExpenseUseCase{
		expenseRepo: expenseRepo,
	}
}

// DeleteRequest represents a request to delete an expense
type DeleteRequest struct {
	ID     string
	UserID string // For authorization
}

// DeleteResponse represents the response after deleting an expense
type DeleteResponse struct {
	ID      string
	Message string
}

// Execute deletes an expense
func (u *DeleteExpenseUseCase) Execute(ctx context.Context, req *DeleteRequest) (*DeleteResponse, error) {
	// Get the expense to verify ownership
	expense, err := u.expenseRepo.GetByID(ctx, req.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get expense: %w", err)
	}

	if expense == nil {
		return nil, fmt.Errorf("expense not found")
	}

	// Verify authorization (user owns this expense)
	if expense.UserID != req.UserID {
		return nil, fmt.Errorf("unauthorized: user does not own this expense")
	}

	// Delete the expense
	if err := u.expenseRepo.Delete(ctx, req.ID); err != nil {
		return nil, fmt.Errorf("failed to delete expense: %w", err)
	}

	return &DeleteResponse{
		ID:      req.ID,
		Message: fmt.Sprintf("Expense '%s' deleted successfully", expense.Description),
	}, nil
}
