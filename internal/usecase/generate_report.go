package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/riverlin/aiexpense/internal/domain"
)

// GenerateReportUseCase handles generating expense reports
type GenerateReportUseCase struct {
	expenseRepo  domain.ExpenseRepository
	categoryRepo domain.CategoryRepository
	metricsRepo  domain.MetricsRepository
}

// NewGenerateReportUseCase creates a new generate report use case
func NewGenerateReportUseCase(
	expenseRepo domain.ExpenseRepository,
	categoryRepo domain.CategoryRepository,
	metricsRepo domain.MetricsRepository,
) *GenerateReportUseCase {
	return &GenerateReportUseCase{
		expenseRepo:  expenseRepo,
		categoryRepo: categoryRepo,
		metricsRepo:  metricsRepo,
	}
}

// ReportRequest represents a request to generate a report
type ReportRequest struct {
	UserID     string
	ReportType string // "daily", "weekly", "monthly"
	StartDate  time.Time
	EndDate    time.Time
}

// ExpenseDetail represents a single expense in a report
type ExpenseDetail struct {
	ID          string    `json:"id"`
	Description string    `json:"description"`
	Amount      float64   `json:"amount"`
	Category    string    `json:"category"`
	Date        time.Time `json:"date"`
	Account     string    `json:"account"`
}

// CategoryBreakdown represents spending by category
type CategoryBreakdown struct {
	Category   string  `json:"category"`
	Total      float64 `json:"total"`
	Count      int     `json:"count"`
	Percentage float64 `json:"percentage"`
}

// DailyBreakdown represents spending by day
type DailyBreakdown struct {
	Date   time.Time `json:"date"`
	Total  float64   `json:"total"`
	Count  int       `json:"count"`
	Amount float64   `json:"amount"`
}

// ExpenseReport represents a generated expense report
type ExpenseReport struct {
	UserID            string              `json:"user_id"`
	ReportType        string              `json:"report_type"`
	Period            string              `json:"period"`
	StartDate         time.Time           `json:"start_date"`
	EndDate           time.Time           `json:"end_date"`
	TotalExpenses     float64             `json:"total_expenses"`
	TransactionCount  int                 `json:"transaction_count"`
	AverageExpense    float64             `json:"average_expense"`
	HighestExpense    float64             `json:"highest_expense"`
	LowestExpense     float64             `json:"lowest_expense"`
	CategoryBreakdown []CategoryBreakdown `json:"category_breakdown"`
	DailyBreakdown    []DailyBreakdown    `json:"daily_breakdown"`
	TopExpenses       []ExpenseDetail     `json:"top_expenses"`
	GeneratedAt       time.Time           `json:"generated_at"`
}

// Execute generates an expense report
func (u *GenerateReportUseCase) Execute(ctx context.Context, req *ReportRequest) (*ExpenseReport, error) {
	if req.UserID == "" {
		return nil, fmt.Errorf("user_id is required")
	}

	// Get all expenses for the user in the date range
	expenses, err := u.expenseRepo.GetByUserIDAndDateRange(ctx, req.UserID, req.StartDate, req.EndDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get expenses: %w", err)
	}

	// Calculate basic statistics
	totalExpenses := 0.0
	highestExpense := 0.0
	lowestExpense := 0.0
	if len(expenses) > 0 {
		lowestExpense = expenses[0].Amount
	}

	categoryMap := make(map[string]*CategoryBreakdown)
	dailyMap := make(map[string]*DailyBreakdown)

	for _, expense := range expenses {
		// Update totals
		totalExpenses += expense.Amount

		if expense.Amount > highestExpense {
			highestExpense = expense.Amount
		}

		if expense.Amount < lowestExpense || lowestExpense == 0 {
			lowestExpense = expense.Amount
		}

		// Category breakdown
		categoryName := "Uncategorized"
		if expense.CategoryID != nil {
			cat, _ := u.categoryRepo.GetByID(ctx, *expense.CategoryID)
			if cat != nil {
				categoryName = cat.Name
			}
		}

		if _, ok := categoryMap[categoryName]; !ok {
			categoryMap[categoryName] = &CategoryBreakdown{
				Category: categoryName,
				Total:    0,
				Count:    0,
			}
		}

		categoryMap[categoryName].Total += expense.Amount
		categoryMap[categoryName].Count += 1

		// Daily breakdown
		dayKey := expense.ExpenseDate.Format("2006-01-02")
		if _, ok := dailyMap[dayKey]; !ok {
			dailyMap[dayKey] = &DailyBreakdown{
				Date:   expense.ExpenseDate,
				Total:  0,
				Count:  0,
				Amount: 0,
			}
		}

		dailyMap[dayKey].Total += expense.Amount
		dailyMap[dayKey].Count += 1
		dailyMap[dayKey].Amount += expense.Amount
	}

	// Convert maps to slices
	var categoryBreakdown []CategoryBreakdown
	for _, cb := range categoryMap {
		cb.Percentage = 0
		if totalExpenses > 0 {
			cb.Percentage = (cb.Total / totalExpenses) * 100
		}
		categoryBreakdown = append(categoryBreakdown, *cb)
	}

	var dailyBreakdown []DailyBreakdown
	for _, db := range dailyMap {
		dailyBreakdown = append(dailyBreakdown, *db)
	}

	// Get all expenses (removed the 10 item limit for comprehensive expense list)
	var topExpenses []ExpenseDetail
	for _, expense := range expenses {
		categoryName := "Uncategorized"
		if expense.CategoryID != nil {
			cat, _ := u.categoryRepo.GetByID(ctx, *expense.CategoryID)
			if cat != nil {
				categoryName = cat.Name
			}
		}

		topExpenses = append(topExpenses, ExpenseDetail{
			ID:          expense.ID,
			Description: expense.Description,
			Amount:      expense.Amount,
			Category:    categoryName,
			Date:        expense.ExpenseDate,
			Account:     expense.Account,
		})
	}

	// Calculate average
	avgExpense := 0.0
	if len(expenses) > 0 {
		avgExpense = totalExpenses / float64(len(expenses))
	}

	period := u.formatPeriod(req.ReportType, req.StartDate, req.EndDate)

	return &ExpenseReport{
		UserID:            req.UserID,
		ReportType:        req.ReportType,
		Period:            period,
		StartDate:         req.StartDate,
		EndDate:           req.EndDate,
		TotalExpenses:     totalExpenses,
		TransactionCount:  len(expenses),
		AverageExpense:    avgExpense,
		HighestExpense:    highestExpense,
		LowestExpense:     lowestExpense,
		CategoryBreakdown: categoryBreakdown,
		DailyBreakdown:    dailyBreakdown,
		TopExpenses:       topExpenses,
		GeneratedAt:       time.Now(),
	}, nil
}

// GenerateMonthlyReport generates a monthly report for the current month
func (u *GenerateReportUseCase) GenerateMonthlyReport(ctx context.Context, userID string) (*ExpenseReport, error) {
	now := time.Now()
	startDate := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	endDate := startDate.AddDate(0, 1, -1).Add(24*time.Hour - time.Nanosecond)

	return u.Execute(ctx, &ReportRequest{
		UserID:     userID,
		ReportType: "monthly",
		StartDate:  startDate,
		EndDate:    endDate,
	})
}

// GenerateWeeklyReport generates a weekly report for the current week
func (u *GenerateReportUseCase) GenerateWeeklyReport(ctx context.Context, userID string) (*ExpenseReport, error) {
	now := time.Now()
	startDate := now.AddDate(0, 0, -int(now.Weekday()))
	endDate := startDate.AddDate(0, 0, 7).Add(-time.Nanosecond)

	return u.Execute(ctx, &ReportRequest{
		UserID:     userID,
		ReportType: "weekly",
		StartDate:  startDate,
		EndDate:    endDate,
	})
}

// GenerateDailyReport generates a daily report for today
func (u *GenerateReportUseCase) GenerateDailyReport(ctx context.Context, userID string) (*ExpenseReport, error) {
	now := time.Now()
	startDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	endDate := startDate.Add(24*time.Hour - time.Nanosecond)

	return u.Execute(ctx, &ReportRequest{
		UserID:     userID,
		ReportType: "daily",
		StartDate:  startDate,
		EndDate:    endDate,
	})
}

// formatPeriod formats the period for the report
func (u *GenerateReportUseCase) formatPeriod(reportType string, startDate, endDate time.Time) string {
	if reportType == "daily" {
		return startDate.Format("2006-01-02")
	} else if reportType == "weekly" {
		return fmt.Sprintf("%s to %s", startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))
	} else if reportType == "monthly" {
		return startDate.Format("2006-01")
	}
	return startDate.Format("2006-01-02")
}
