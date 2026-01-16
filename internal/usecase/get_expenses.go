package usecase

import (
	"context"
	"time"

	"github.com/riverlin/aiexpense/internal/domain"
)

// GetExpensesUseCase handles retrieving expenses
type GetExpensesUseCase struct {
	expenseRepo  domain.ExpenseRepository
	categoryRepo domain.CategoryRepository
}

// NewGetExpensesUseCase creates a new get expenses use case
func NewGetExpensesUseCase(
	expenseRepo domain.ExpenseRepository,
	categoryRepo domain.CategoryRepository,
) *GetExpensesUseCase {
	return &GetExpensesUseCase{
		expenseRepo:  expenseRepo,
		categoryRepo: categoryRepo,
	}
}

// GetAllRequest represents a request to get all expenses
type GetAllRequest struct {
	UserID string
}

// GetByDateRangeRequest represents a request to get expenses by date range
type GetByDateRangeRequest struct {
	UserID string
	From   time.Time
	To     time.Time
}

// GetByCategoryRequest represents a request to get expenses by category
type GetByCategoryRequest struct {
	UserID     string
	CategoryID string
}

// ExpenseDTO represents an expense in responses
type ExpenseDTO struct {
	ID           string
	Description  string
	Amount       float64
	CategoryID   *string
	CategoryName *string
	Date         time.Time
}

// GetAllResponse represents the response for getting all expenses
type GetAllResponse struct {
	Expenses []*ExpenseDTO
	Total    float64
	Count    int
}

// ExecuteGetAll retrieves all expenses for a user
func (u *GetExpensesUseCase) ExecuteGetAll(ctx context.Context, req *GetAllRequest) (*GetAllResponse, error) {
	expenses, err := u.expenseRepo.GetByUserID(ctx, req.UserID)
	if err != nil {
		return nil, err
	}

	return u.buildResponse(ctx, expenses)
}

// ExecuteGetByDateRange retrieves expenses within a date range
func (u *GetExpensesUseCase) ExecuteGetByDateRange(ctx context.Context, req *GetByDateRangeRequest) (*GetAllResponse, error) {
	expenses, err := u.expenseRepo.GetByUserIDAndDateRange(ctx, req.UserID, req.From, req.To)
	if err != nil {
		return nil, err
	}

	return u.buildResponse(ctx, expenses)
}

// ExecuteGetByCategory retrieves expenses in a category
func (u *GetExpensesUseCase) ExecuteGetByCategory(ctx context.Context, req *GetByCategoryRequest) (*GetAllResponse, error) {
	expenses, err := u.expenseRepo.GetByUserIDAndCategory(ctx, req.UserID, req.CategoryID)
	if err != nil {
		return nil, err
	}

	return u.buildResponse(ctx, expenses)
}

// buildResponse builds response from expenses
func (u *GetExpensesUseCase) buildResponse(ctx context.Context, expenses []*domain.Expense) (*GetAllResponse, error) {
	var dtos []*ExpenseDTO
	var total float64

	for _, expense := range expenses {
		var categoryName *string
		if expense.CategoryID != nil {
			category, _ := u.categoryRepo.GetByID(ctx, *expense.CategoryID)
			if category != nil {
				categoryName = &category.Name
			}
		}

		dto := &ExpenseDTO{
			ID:           expense.ID,
			Description:  expense.Description,
			Amount:       expense.Amount,
			CategoryID:   expense.CategoryID,
			CategoryName: categoryName,
			Date:         expense.ExpenseDate,
		}
		dtos = append(dtos, dto)
		total += expense.Amount
	}

	return &GetAllResponse{
		Expenses: dtos,
		Total:    total,
		Count:    len(dtos),
	}, nil
}
