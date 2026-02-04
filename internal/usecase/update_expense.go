package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/riverlin/aiexpense/internal/domain"
)

// UpdateExpenseUseCase handles updating existing expenses
type UpdateExpenseUseCase struct {
	expenseRepo     domain.ExpenseRepository
	categoryRepo    domain.CategoryRepository
	exchangeRateSvc domain.ExchangeRateService
}

// NewUpdateExpenseUseCase creates a new update expense use case
func NewUpdateExpenseUseCase(
	expenseRepo domain.ExpenseRepository,
	categoryRepo domain.CategoryRepository,
	exchangeRateSvc domain.ExchangeRateService,
) *UpdateExpenseUseCase {
	return &UpdateExpenseUseCase{
		expenseRepo:     expenseRepo,
		categoryRepo:    categoryRepo,
		exchangeRateSvc: exchangeRateSvc,
	}
}

// UpdateRequest represents a request to update an expense
type UpdateRequest struct {
	ID             string
	UserID         string // For authorization
	Description    *string
	OriginalAmount *float64
	Currency       *string
	CategoryID     *string
	Account        *string
	ExpenseDate    *time.Time
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

	needRecalc := false
	if req.OriginalAmount != nil {
		expense.OriginalAmount = *req.OriginalAmount
		needRecalc = true
	}

	if req.Currency != nil {
		currency := normalizeCurrency(*req.Currency)
		if currency != "" {
			expense.Currency = currency
			needRecalc = true
		}
	}

	if needRecalc {
		u.recalculateHomeAmount(ctx, expense)
	}

	if req.ExpenseDate != nil {
		expense.ExpenseDate = *req.ExpenseDate
	}

	if req.Account != nil {
		expense.Account = *req.Account
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

func (u *UpdateExpenseUseCase) recalculateHomeAmount(ctx context.Context, expense *domain.Expense) {
	if expense.OriginalAmount == 0 {
		expense.HomeAmount = 0
		expense.Amount = 0
		return
	}

	fromCurrency := normalizeCurrency(expense.Currency)
	toCurrency := normalizeCurrency(expense.HomeCurrency)
	if toCurrency == "" {
		toCurrency = fromCurrency
		expense.HomeCurrency = toCurrency
	}

	if fromCurrency == "" {
		fromCurrency = toCurrency
		expense.Currency = fromCurrency
	}

	if u.exchangeRateSvc != nil && fromCurrency != "" && toCurrency != "" && fromCurrency != toCurrency {
		converted, rate, err := u.exchangeRateSvc.Convert(ctx, expense.OriginalAmount, fromCurrency, toCurrency, expense.ExpenseDate)
		if err == nil {
			expense.HomeAmount = converted
			expense.ExchangeRate = rate
			expense.Amount = expense.HomeAmount
			return
		}
	}

	// Fallback to 1:1 conversion
	expense.HomeAmount = expense.OriginalAmount
	if expense.ExchangeRate == 0 {
		expense.ExchangeRate = 1.0
	}
	expense.Amount = expense.HomeAmount
}
