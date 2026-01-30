package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/riverlin/aiexpense/internal/domain"
)

// UpdateExpenseUseCase handles updating existing expenses
type UpdateExpenseUseCase struct {
	expenseRepo  domain.ExpenseRepository
	categoryRepo domain.CategoryRepository
}

// NewUpdateExpenseUseCase creates a new update expense use case
func NewUpdateExpenseUseCase(
	expenseRepo domain.ExpenseRepository,
	categoryRepo domain.CategoryRepository,
) *UpdateExpenseUseCase {
	return &UpdateExpenseUseCase{
		expenseRepo:  expenseRepo,
		categoryRepo: categoryRepo,
	}
}

// UpdateRequest represents a request to update an expense
type UpdateRequest struct {
	ID          string
	UserID      string // For authorization
	Description *string
	Amount      *float64
	CategoryID  *string
	ExpenseDate *time.Time
}

// UpdateResponse represents the response after updating an expense
type UpdateResponse struct {
	ID       string
	Message  string
	Category string
}

// Execute updates an existing expense
func (u *UpdateExpenseUseCase) Execute(ctx context.Context, req *UpdateRequest) (*UpdateResponse, error) {
	// Get the existing expense
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

	// Update fields if provided
	if req.Description != nil {
		expense.Description = *req.Description
	}

	if req.Amount != nil {
		expense.OriginalAmount = *req.Amount
		expense.HomeAmount = *req.Amount
		expense.Amount = expense.HomeAmount
		if expense.ExchangeRate == 0 {
			expense.ExchangeRate = 1.0
		}
	}

	if req.ExpenseDate != nil {
		expense.ExpenseDate = *req.ExpenseDate
	}

	// Handle category update
	var categoryName string
	if req.CategoryID != nil {
		expense.CategoryID = req.CategoryID
		// Get category name for response
		category, _ := u.categoryRepo.GetByID(ctx, *req.CategoryID)
		if category != nil {
			categoryName = category.Name
		}
	} else if expense.CategoryID != nil {
		// Keep existing category, get its name
		category, _ := u.categoryRepo.GetByID(ctx, *expense.CategoryID)
		if category != nil {
			categoryName = category.Name
		}
	}

	// Update timestamp
	expense.UpdatedAt = time.Now()

	// Save the updated expense
	if err := u.expenseRepo.Update(ctx, expense); err != nil {
		return nil, fmt.Errorf("failed to update expense: %w", err)
	}

	// Prepare response message
	message := fmt.Sprintf("Expense updated: %s %s", expense.Description, formatAmount(expense.Amount))
	if categoryName != "" {
		message = fmt.Sprintf("Expense updated: %s %s [%s]", expense.Description, formatAmount(expense.Amount), categoryName)
	}

	return &UpdateResponse{
		ID:       expense.ID,
		Message:  message,
		Category: categoryName,
	}, nil
}
