package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/riverlin/aiexpense/internal/domain"
)

// BudgetManagementUseCase handles managing user budgets
type BudgetManagementUseCase struct {
	categoryRepo domain.CategoryRepository
	expenseRepo  domain.ExpenseRepository
}

// NewBudgetManagementUseCase creates a new budget management use case
func NewBudgetManagementUseCase(
	categoryRepo domain.CategoryRepository,
	expenseRepo domain.ExpenseRepository,
) *BudgetManagementUseCase {
	return &BudgetManagementUseCase{
		categoryRepo: categoryRepo,
		expenseRepo:  expenseRepo,
	}
}

// Budget represents a user's budget
type Budget struct {
	ID         string    `json:"id"`
	UserID     string    `json:"user_id"`
	CategoryID *string   `json:"category_id,omitempty"`
	Category   string    `json:"category"`
	Limit      float64   `json:"limit"`
	Period     string    `json:"period"`    // "monthly", "weekly", "daily"
	Threshold  float64   `json:"threshold"` // Alert when spending exceeds this %
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// BudgetStatus represents the current status of a budget
type BudgetStatus struct {
	ID             string  `json:"id"`
	Category       string  `json:"category"`
	Limit          float64 `json:"limit"`
	Spent          float64 `json:"spent"`
	Remaining      float64 `json:"remaining"`
	Percentage     float64 `json:"percentage"`
	IsExceeded     bool    `json:"is_exceeded"`
	AlertTriggered bool    `json:"alert_triggered"`
	Message        string  `json:"message"`
}

// SetBudgetRequest represents a request to set a budget
type SetBudgetRequest struct {
	UserID     string
	CategoryID *string
	Category   string
	Limit      float64
	Period     string  // "monthly", "weekly", "daily"
	Threshold  float64 // 0-100, percentage
}

// SetBudgetResponse represents the response after setting a budget
type SetBudgetResponse struct {
	Budget  *Budget `json:"budget"`
	Message string  `json:"message"`
}

// SetBudget creates or updates a budget for a category
func (u *BudgetManagementUseCase) SetBudget(ctx context.Context, req *SetBudgetRequest) (*SetBudgetResponse, error) {
	if req.UserID == "" {
		return nil, fmt.Errorf("user_id is required")
	}

	if req.Limit <= 0 {
		return nil, fmt.Errorf("budget limit must be greater than 0")
	}

	if req.Period == "" {
		req.Period = "monthly"
	}

	if req.Threshold <= 0 {
		req.Threshold = 80 // Default 80%
	}

	// In production, this would be stored in a budget table
	// For now, we're just returning the budget object
	budget := &Budget{
		ID:         uuid.New().String(),
		UserID:     req.UserID,
		CategoryID: req.CategoryID,
		Category:   req.Category,
		Limit:      req.Limit,
		Period:     req.Period,
		Threshold:  req.Threshold,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	return &SetBudgetResponse{
		Budget:  budget,
		Message: fmt.Sprintf("Budget set: %s %s %.2f (alert at %.0f%%)", req.Category, req.Period, req.Limit, req.Threshold),
	}, nil
}

// GetBudgetStatusRequest represents a request to get budget status
type GetBudgetStatusRequest struct {
	UserID     string
	CategoryID *string
}

// GetBudgetStatusResponse represents the response with budget status
type GetBudgetStatusResponse struct {
	Budgets    []BudgetStatus `json:"budgets"`
	TotalLimit float64        `json:"total_limit"`
	TotalSpent float64        `json:"total_spent"`
	Alert      bool           `json:"alert"`
	Message    string         `json:"message"`
}

// GetBudgetStatus retrieves the current budget status
func (u *BudgetManagementUseCase) GetBudgetStatus(ctx context.Context, req *GetBudgetStatusRequest) (*GetBudgetStatusResponse, error) {
	if req.UserID == "" {
		return nil, fmt.Errorf("user_id is required")
	}

	// Get user's expenses for the current month
	now := time.Now()
	startDate := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	endDate := now

	expenses, err := u.expenseRepo.GetByUserIDAndDateRange(ctx, req.UserID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get expenses: %w", err)
	}

	// Calculate spending by category
	categorySpending := make(map[string]float64)
	totalSpent := 0.0

	for _, expense := range expenses {
		categoryName := "Uncategorized"
		if expense.CategoryID != nil {
			cat, _ := u.categoryRepo.GetByID(ctx, *expense.CategoryID)
			if cat != nil {
				categoryName = cat.Name
			}
		}

		categorySpending[categoryName] += expense.Amount
		totalSpent += expense.Amount
	}

	// Get user's categories to determine budgets
	categories, _ := u.categoryRepo.GetByUserID(ctx, req.UserID)

	// Build budget status list
	var budgets []BudgetStatus
	totalLimit := 0.0
	hasAlert := false

	for _, cat := range categories {
		categoryName := cat.Name
		spent := categorySpending[categoryName]

		// Default budget: 100 per category (in production, would be from budget table)
		limit := 100.0
		threshold := 80.0

		remaining := limit - spent
		percentage := 0.0
		if limit > 0 {
			percentage = (spent / limit) * 100
		}

		isExceeded := spent > limit
		alertTriggered := percentage >= threshold

		if alertTriggered {
			hasAlert = true
		}

		message := "On track"
		if isExceeded {
			message = fmt.Sprintf("Exceeded by %.2f", spent-limit)
		} else if alertTriggered {
			message = fmt.Sprintf("%.0f%% of budget used", percentage)
		}

		budgets = append(budgets, BudgetStatus{
			ID:             cat.ID,
			Category:       categoryName,
			Limit:          limit,
			Spent:          spent,
			Remaining:      remaining,
			Percentage:     percentage,
			IsExceeded:     isExceeded,
			AlertTriggered: alertTriggered,
			Message:        message,
		})

		totalLimit += limit
	}

	resp := &GetBudgetStatusResponse{
		Budgets:    budgets,
		TotalLimit: totalLimit,
		TotalSpent: totalSpent,
		Alert:      hasAlert,
	}

	if hasAlert {
		resp.Message = "Budget alert: Some categories have exceeded alerts"
	} else {
		resp.Message = "All budgets on track"
	}

	return resp, nil
}

// CompareToBudgetRequest represents a request to compare spending to budget
type CompareToBudgetRequest struct {
	UserID     string
	CategoryID *string
	Period     string // "daily", "weekly", "monthly"
}

// BudgetComparison represents a comparison of spending to budget
type BudgetComparison struct {
	Category       string  `json:"category"`
	BudgetLimit    float64 `json:"budget_limit"`
	Spent          float64 `json:"spent"`
	Remaining      float64 `json:"remaining"`
	PercentageUsed float64 `json:"percentage_used"`
	Status         string  `json:"status"` // "under", "warning", "exceeded"
	Recommendation string  `json:"recommendation"`
}

// CompareToBudget compares spending to budget for a category
func (u *BudgetManagementUseCase) CompareToBudget(ctx context.Context, req *CompareToBudgetRequest) (*BudgetComparison, error) {
	if req.UserID == "" {
		return nil, fmt.Errorf("user_id is required")
	}

	if req.Period == "" {
		req.Period = "monthly"
	}

	// Calculate period dates
	now := time.Now()
	var startDate, endDate time.Time

	switch req.Period {
	case "daily":
		startDate = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		endDate = startDate.Add(24*time.Hour - time.Nanosecond)
	case "weekly":
		startDate = now.AddDate(0, 0, -int(now.Weekday()))
		endDate = startDate.AddDate(0, 0, 7).Add(-time.Nanosecond)
	case "monthly":
		startDate = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		endDate = startDate.AddDate(0, 1, -1).Add(24*time.Hour - time.Nanosecond)
	}

	// Get category name
	var categoryName string
	if req.CategoryID != nil {
		cat, _ := u.categoryRepo.GetByID(ctx, *req.CategoryID)
		if cat != nil {
			categoryName = cat.Name
		}
	}

	// Get expenses for the category and period
	var expenses []*domain.Expense
	var err error

	if req.CategoryID != nil {
		expenses, err = u.expenseRepo.GetByUserIDAndCategory(ctx, req.UserID, *req.CategoryID)
	} else {
		expenses, err = u.expenseRepo.GetByUserIDAndDateRange(ctx, req.UserID, startDate, endDate)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get expenses: %w", err)
	}

	// Filter by date range if category was specified
	if req.CategoryID != nil {
		var filtered []*domain.Expense
		for _, exp := range expenses {
			if exp.ExpenseDate.After(startDate) && exp.ExpenseDate.Before(endDate) {
				filtered = append(filtered, exp)
			}
		}
		expenses = filtered
	}

	// Calculate spending
	spent := 0.0
	for _, exp := range expenses {
		spent += exp.Amount
	}

	// Default budget (in production, would come from budget table)
	budgetLimit := 100.0
	remaining := budgetLimit - spent
	percentageUsed := 0.0
	if budgetLimit > 0 {
		percentageUsed = (spent / budgetLimit) * 100
	}

	// Determine status and recommendation
	status := "under"
	recommendation := "Keep up the good spending habits!"

	if percentageUsed >= 100 {
		status = "exceeded"
		recommendation = fmt.Sprintf("You've exceeded your budget by %.2f. Try to reduce spending.", spent-budgetLimit)
	} else if percentageUsed >= 80 {
		status = "warning"
		recommendation = fmt.Sprintf("You're at %.0f%% of your budget. Be careful not to exceed it.", percentageUsed)
	}

	return &BudgetComparison{
		Category:       categoryName,
		BudgetLimit:    budgetLimit,
		Spent:          spent,
		Remaining:      remaining,
		PercentageUsed: percentageUsed,
		Status:         status,
		Recommendation: recommendation,
	}, nil
}
