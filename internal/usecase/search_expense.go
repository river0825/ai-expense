package usecase

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/riverlin/aiexpense/internal/domain"
)

// SearchExpenseUseCase handles searching and filtering expenses
type SearchExpenseUseCase struct {
	expenseRepo  domain.ExpenseRepository
	categoryRepo domain.CategoryRepository
}

// NewSearchExpenseUseCase creates a new search expense use case
func NewSearchExpenseUseCase(
	expenseRepo domain.ExpenseRepository,
	categoryRepo domain.CategoryRepository,
) *SearchExpenseUseCase {
	return &SearchExpenseUseCase{
		expenseRepo:  expenseRepo,
		categoryRepo: categoryRepo,
	}
}

// SearchRequest represents a request to search expenses
type SearchRequest struct {
	UserID     string
	Query      string     // Search term for description
	CategoryID *string    // Filter by category
	MinAmount  *float64   // Filter by minimum amount
	MaxAmount  *float64   // Filter by maximum amount
	StartDate  *time.Time // Filter by date range start
	EndDate    *time.Time // Filter by date range end
	SortBy     string     // "date_desc", "date_asc", "amount_desc", "amount_asc"
	Limit      int
	Offset     int
}

// SearchResult represents a search result
type SearchResult struct {
	ID          string    `json:"id"`
	Description string    `json:"description"`
	Amount      float64   `json:"amount"`
	Category    string    `json:"category"`
	Date        time.Time `json:"date"`
}

// SearchResponse represents the response from a search
type SearchResponse struct {
	Results     []*SearchResult `json:"results"`
	Total       int             `json:"total"`
	Limit       int             `json:"limit"`
	Offset      int             `json:"offset"`
	Pages       int             `json:"pages"`
	CurrentPage int             `json:"current_page"`
	Message     string          `json:"message"`
}

// Search performs a search for expenses
func (u *SearchExpenseUseCase) Search(ctx context.Context, req *SearchRequest) (*SearchResponse, error) {
	if req.UserID == "" {
		return nil, fmt.Errorf("user_id is required")
	}

	// Default pagination
	if req.Limit <= 0 {
		req.Limit = 20
	}
	if req.Limit > 100 {
		req.Limit = 100
	}

	// Default sort
	if req.SortBy == "" {
		req.SortBy = "date_desc"
	}

	// Set default date range if not provided
	if req.StartDate == nil {
		start := time.Now().AddDate(-1, 0, 0) // Last year
		req.StartDate = &start
	}
	if req.EndDate == nil {
		end := time.Now()
		req.EndDate = &end
	}

	// Get expenses from repository (in production, would support filtering)
	allExpenses, err := u.expenseRepo.GetByUserIDAndDateRange(ctx, req.UserID, *req.StartDate, *req.EndDate)
	if err != nil {
		return nil, fmt.Errorf("failed to search expenses: %w", err)
	}

	// Filter expenses
	var filtered []*domain.Expense
	for _, exp := range allExpenses {
		// Filter by query (description search)
		if req.Query != "" {
			if !strings.Contains(strings.ToLower(exp.Description), strings.ToLower(req.Query)) {
				continue
			}
		}

		// Filter by category
		if req.CategoryID != nil {
			if exp.CategoryID == nil || *exp.CategoryID != *req.CategoryID {
				continue
			}
		}

		// Filter by amount range
		if req.MinAmount != nil && exp.Amount < *req.MinAmount {
			continue
		}
		if req.MaxAmount != nil && exp.Amount > *req.MaxAmount {
			continue
		}

		filtered = append(filtered, exp)
	}

	// Sort expenses
	u.sortExpenses(filtered, req.SortBy)

	// Apply pagination
	total := len(filtered)
	pages := (total + req.Limit - 1) / req.Limit
	currentPage := (req.Offset / req.Limit) + 1

	start := req.Offset
	end := start + req.Limit
	if end > total {
		end = total
	}

	var paginatedExpenses []*domain.Expense
	if start < total {
		paginatedExpenses = filtered[start:end]
	}

	// Convert to results
	var results []*SearchResult
	for _, exp := range paginatedExpenses {
		categoryName := "Uncategorized"
		if exp.CategoryID != nil {
			cat, _ := u.categoryRepo.GetByID(ctx, *exp.CategoryID)
			if cat != nil {
				categoryName = cat.Name
			}
		}

		results = append(results, &SearchResult{
			ID:          exp.ID,
			Description: exp.Description,
			Amount:      exp.Amount,
			Category:    categoryName,
			Date:        exp.ExpenseDate,
		})
	}

	message := fmt.Sprintf("Found %d expenses", total)
	if req.Query != "" {
		message = fmt.Sprintf("Found %d expenses matching '%s'", total, req.Query)
	}

	return &SearchResponse{
		Results:     results,
		Total:       total,
		Limit:       req.Limit,
		Offset:      req.Offset,
		Pages:       pages,
		CurrentPage: currentPage,
		Message:     message,
	}, nil
}

// sortExpenses sorts expenses based on the sort parameter
func (u *SearchExpenseUseCase) sortExpenses(expenses []*domain.Expense, sortBy string) {
	switch sortBy {
	case "date_asc":
		// Sort by date ascending (already in order from DB)
	case "date_desc":
		// Reverse order (most recent first)
		for i, j := 0, len(expenses)-1; i < j; i, j = i+1, j-1 {
			expenses[i], expenses[j] = expenses[j], expenses[i]
		}
	case "amount_desc":
		// Sort by amount descending
		for i := 0; i < len(expenses); i++ {
			for j := i + 1; j < len(expenses); j++ {
				if expenses[j].Amount > expenses[i].Amount {
					expenses[i], expenses[j] = expenses[j], expenses[i]
				}
			}
		}
	case "amount_asc":
		// Sort by amount ascending
		for i := 0; i < len(expenses); i++ {
			for j := i + 1; j < len(expenses); j++ {
				if expenses[j].Amount < expenses[i].Amount {
					expenses[i], expenses[j] = expenses[j], expenses[i]
				}
			}
		}
	}
}

// FilterRequest represents a request to filter expenses
type FilterRequest struct {
	UserID     string
	CategoryID string
	Period     string // "today", "this_week", "this_month", "last_30_days", "custom"
	StartDate  *time.Time
	EndDate    *time.Time
}

// FilterResponse represents filtered expenses
type FilterResponse struct {
	Total    float64         `json:"total"`
	Count    int             `json:"count"`
	Average  float64         `json:"average"`
	Min      float64         `json:"min"`
	Max      float64         `json:"max"`
	Expenses []*SearchResult `json:"expenses"`
	Message  string          `json:"message"`
}

// Filter retrieves and aggregates expenses by filters
func (u *SearchExpenseUseCase) Filter(ctx context.Context, req *FilterRequest) (*FilterResponse, error) {
	if req.UserID == "" {
		return nil, fmt.Errorf("user_id is required")
	}

	// Determine date range based on period
	now := time.Now()
	var startDate, endDate time.Time

	switch req.Period {
	case "today":
		startDate = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		endDate = startDate.Add(24*time.Hour - time.Nanosecond)
	case "this_week":
		startDate = now.AddDate(0, 0, -int(now.Weekday()))
		endDate = startDate.AddDate(0, 0, 7).Add(-time.Nanosecond)
	case "this_month":
		startDate = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		endDate = startDate.AddDate(0, 1, -1).Add(24*time.Hour - time.Nanosecond)
	case "last_30_days":
		endDate = now
		startDate = now.AddDate(0, 0, -30)
	case "custom":
		if req.StartDate != nil && req.EndDate != nil {
			startDate = *req.StartDate
			endDate = *req.EndDate
		} else {
			return nil, fmt.Errorf("start_date and end_date required for custom period")
		}
	default:
		// Default to current month
		startDate = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		endDate = startDate.AddDate(0, 1, -1).Add(24*time.Hour - time.Nanosecond)
	}

	// Get expenses
	var expenses []*domain.Expense
	var err error

	if req.CategoryID != "" {
		expenses, err = u.expenseRepo.GetByUserIDAndCategory(ctx, req.UserID, req.CategoryID)
		if err != nil {
			return nil, fmt.Errorf("failed to filter expenses: %w", err)
		}
		// Filter by date
		var filtered []*domain.Expense
		for _, exp := range expenses {
			if exp.ExpenseDate.After(startDate) && exp.ExpenseDate.Before(endDate) {
				filtered = append(filtered, exp)
			}
		}
		expenses = filtered
	} else {
		expenses, err = u.expenseRepo.GetByUserIDAndDateRange(ctx, req.UserID, startDate, endDate)
		if err != nil {
			return nil, fmt.Errorf("failed to filter expenses: %w", err)
		}
	}

	// Calculate statistics
	total := 0.0
	min := 0.0
	max := 0.0
	if len(expenses) > 0 {
		min = expenses[0].Amount
		max = expenses[0].Amount
	}

	var results []*SearchResult
	for _, exp := range expenses {
		total += exp.Amount

		if exp.Amount < min || min == 0 {
			min = exp.Amount
		}
		if exp.Amount > max {
			max = exp.Amount
		}

		categoryName := "Uncategorized"
		if exp.CategoryID != nil {
			cat, _ := u.categoryRepo.GetByID(ctx, *exp.CategoryID)
			if cat != nil {
				categoryName = cat.Name
			}
		}

		results = append(results, &SearchResult{
			ID:          exp.ID,
			Description: exp.Description,
			Amount:      exp.Amount,
			Category:    categoryName,
			Date:        exp.ExpenseDate,
		})
	}

	avg := 0.0
	if len(expenses) > 0 {
		avg = total / float64(len(expenses))
	}

	return &FilterResponse{
		Total:    total,
		Count:    len(expenses),
		Average:  avg,
		Min:      min,
		Max:      max,
		Expenses: results,
		Message:  fmt.Sprintf("Retrieved %d expenses for period", len(expenses)),
	}, nil
}
