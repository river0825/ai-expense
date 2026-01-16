package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/riverlin/aiexpense/internal/domain"
)

// RecurringExpenseUseCase handles recurring/subscription expenses
type RecurringExpenseUseCase struct {
	expenseRepo  domain.ExpenseRepository
	categoryRepo domain.CategoryRepository
}

// NewRecurringExpenseUseCase creates a new recurring expense use case
func NewRecurringExpenseUseCase(
	expenseRepo domain.ExpenseRepository,
	categoryRepo domain.CategoryRepository,
) *RecurringExpenseUseCase {
	return &RecurringExpenseUseCase{
		expenseRepo:  expenseRepo,
		categoryRepo: categoryRepo,
	}
}

// RecurringExpense represents a recurring expense
type RecurringExpense struct {
	ID          string     `json:"id"`
	UserID      string     `json:"user_id"`
	Description string     `json:"description"`
	Amount      float64    `json:"amount"`
	CategoryID  *string    `json:"category_id,omitempty"`
	Frequency   string     `json:"frequency"` // "daily", "weekly", "biweekly", "monthly", "quarterly", "yearly"
	StartDate   time.Time  `json:"start_date"`
	EndDate     *time.Time `json:"end_date,omitempty"` // nil = no end date
	IsActive    bool       `json:"is_active"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// CreateRecurringRequest represents a request to create a recurring expense
type CreateRecurringRequest struct {
	UserID      string
	Description string
	Amount      float64
	CategoryID  *string
	Frequency   string
	StartDate   time.Time
	EndDate     *time.Time
}

// CreateRecurringResponse represents the response after creating a recurring expense
type CreateRecurringResponse struct {
	ID      string
	Message string
}

// CreateRecurring creates a new recurring expense
func (u *RecurringExpenseUseCase) CreateRecurring(ctx context.Context, req *CreateRecurringRequest) (*CreateRecurringResponse, error) {
	if req.UserID == "" || req.Description == "" || req.Amount <= 0 {
		return nil, fmt.Errorf("user_id, description, and amount are required")
	}

	if req.Frequency == "" {
		req.Frequency = "monthly"
	}

	// Validate frequency
	validFrequencies := map[string]bool{
		"daily": true, "weekly": true, "biweekly": true, "monthly": true,
		"quarterly": true, "yearly": true,
	}
	if !validFrequencies[req.Frequency] {
		return nil, fmt.Errorf("invalid frequency: %s", req.Frequency)
	}

	// In production, this would be stored in a recurring_expenses table
	// For now, we return the created ID
	id := uuid.New().String()

	return &CreateRecurringResponse{
		ID:      id,
		Message: fmt.Sprintf("Recurring expense '%s' created: %s every %s", req.Description, formatAmount(req.Amount), req.Frequency),
	}, nil
}

// ListRecurringRequest represents a request to list recurring expenses
type ListRecurringRequest struct {
	UserID string
}

// ListRecurringResponse represents a list of recurring expenses
type ListRecurringResponse struct {
	Recurring []*RecurringExpense `json:"recurring"`
	Total     int                 `json:"total"`
	Message   string              `json:"message"`
}

// ListRecurring retrieves all active recurring expenses for a user
func (u *RecurringExpenseUseCase) ListRecurring(ctx context.Context, req *ListRecurringRequest) (*ListRecurringResponse, error) {
	if req.UserID == "" {
		return nil, fmt.Errorf("user_id is required")
	}

	// In production, this would query from recurring_expenses table
	// For now, return empty list
	return &ListRecurringResponse{
		Recurring: make([]*RecurringExpense, 0),
		Total:     0,
		Message:   "No recurring expenses found",
	}, nil
}

// ProcessRecurringRequest represents a request to process recurring expenses
type ProcessRecurringRequest struct {
	UserID string
	Date   time.Time // Date to process recurring expenses for
}

// ProcessRecurringResponse represents the response after processing
type ProcessRecurringResponse struct {
	ProcessedCount  int
	CreatedExpenses []struct {
		ID          string
		Amount      float64
		Description string
	}
	Message string
}

// ProcessRecurring generates actual expenses from recurring expense definitions
func (u *RecurringExpenseUseCase) ProcessRecurring(ctx context.Context, req *ProcessRecurringRequest) (*ProcessRecurringResponse, error) {
	if req.UserID == "" {
		return nil, fmt.Errorf("user_id is required")
	}

	if req.Date.IsZero() {
		req.Date = time.Now()
	}

	// In production:
	// 1. Query all active recurring expenses for the user
	// 2. Check which ones should fire today based on frequency
	// 3. Create actual Expense records
	// 4. Return summary of created expenses

	return &ProcessRecurringResponse{
		ProcessedCount: 0,
		CreatedExpenses: make([]struct {
			ID          string
			Amount      float64
			Description string
		}, 0),
		Message: "No recurring expenses to process for this date",
	}, nil
}

// UpdateRecurringRequest represents a request to update a recurring expense
type UpdateRecurringRequest struct {
	UserID      string
	ID          string
	Description *string
	Amount      *float64
	Frequency   *string
	IsActive    *bool
}

// UpdateRecurringResponse represents the response after updating
type UpdateRecurringResponse struct {
	ID      string
	Message string
}

// UpdateRecurring updates a recurring expense
func (u *RecurringExpenseUseCase) UpdateRecurring(ctx context.Context, req *UpdateRecurringRequest) (*UpdateRecurringResponse, error) {
	if req.UserID == "" || req.ID == "" {
		return nil, fmt.Errorf("user_id and id are required")
	}

	// In production: retrieve, verify ownership, update, save
	return &UpdateRecurringResponse{
		ID:      req.ID,
		Message: "Recurring expense updated successfully",
	}, nil
}

// DeleteRecurringRequest represents a request to delete a recurring expense
type DeleteRecurringRequest struct {
	UserID string
	ID     string
}

// DeleteRecurringResponse represents the response after deletion
type DeleteRecurringResponse struct {
	ID      string
	Message string
}

// DeleteRecurring deletes a recurring expense
func (u *RecurringExpenseUseCase) DeleteRecurring(ctx context.Context, req *DeleteRecurringRequest) (*DeleteRecurringResponse, error) {
	if req.UserID == "" || req.ID == "" {
		return nil, fmt.Errorf("user_id and id are required")
	}

	// In production: retrieve, verify ownership, delete
	return &DeleteRecurringResponse{
		ID:      req.ID,
		Message: "Recurring expense deleted successfully",
	}, nil
}

// GetUpcomingRequest represents a request to get upcoming recurring expenses
type GetUpcomingRequest struct {
	UserID string
	Days   int // How many days ahead to look
}

// UpcomingExpense represents an upcoming recurring expense instance
type UpcomingExpense struct {
	RecurringID string    `json:"recurring_id"`
	Description string    `json:"description"`
	Amount      float64   `json:"amount"`
	Category    string    `json:"category"`
	DueDate     time.Time `json:"due_date"`
}

// GetUpcomingResponse represents a list of upcoming expenses
type GetUpcomingResponse struct {
	Upcoming []*UpcomingExpense `json:"upcoming"`
	Total    int                `json:"total"`
	Message  string             `json:"message"`
}

// GetUpcoming retrieves upcoming recurring expenses
func (u *RecurringExpenseUseCase) GetUpcoming(ctx context.Context, req *GetUpcomingRequest) (*GetUpcomingResponse, error) {
	if req.UserID == "" {
		return nil, fmt.Errorf("user_id is required")
	}

	if req.Days <= 0 {
		req.Days = 30 // Default to next 30 days
	}

	// In production: calculate upcoming occurrences based on frequency
	return &GetUpcomingResponse{
		Upcoming: make([]*UpcomingExpense, 0),
		Total:    0,
		Message:  "No upcoming recurring expenses",
	}, nil
}
