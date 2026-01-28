package usecase

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"time"

	"github.com/riverlin/aiexpense/internal/domain"
)

// DataExportUseCase handles exporting expense data
type DataExportUseCase struct {
	expenseRepo  domain.ExpenseRepository
	categoryRepo domain.CategoryRepository
}

// NewDataExportUseCase creates a new data export use case
func NewDataExportUseCase(
	expenseRepo domain.ExpenseRepository,
	categoryRepo domain.CategoryRepository,
) *DataExportUseCase {
	return &DataExportUseCase{
		expenseRepo:  expenseRepo,
		categoryRepo: categoryRepo,
	}
}

// ExportRequest represents a request to export data
type ExportRequest struct {
	UserID    string
	Format    string // "csv", "json"
	StartDate time.Time
	EndDate   time.Time
}

// ExportedExpense represents an expense in export format
type ExportedExpense struct {
	ID          string  `json:"id" csv:"ID"`
	Date        string  `json:"date" csv:"Date"`
	Description string  `json:"description" csv:"Description"`
	Amount      float64 `json:"amount" csv:"Amount"`
	Category    string  `json:"category" csv:"Category"`
	Account     string  `json:"account" csv:"Account"`
	CreatedAt   string  `json:"created_at" csv:"CreatedAt"`
	UpdatedAt   string  `json:"updated_at" csv:"UpdatedAt"`
}

// ExportData represents exported data
type ExportData struct {
	Format       string            `json:"format"`
	ExportedAt   time.Time         `json:"exported_at"`
	PeriodStart  time.Time         `json:"period_start"`
	PeriodEnd    time.Time         `json:"period_end"`
	TotalRecords int               `json:"total_records"`
	Data         []ExportedExpense `json:"data"`
}

// Execute exports expense data in the requested format
func (u *DataExportUseCase) Execute(ctx context.Context, req *ExportRequest) (*ExportData, error) {
	if req.UserID == "" {
		return nil, fmt.Errorf("user_id is required")
	}

	if req.Format == "" {
		req.Format = "json"
	}

	// Get all expenses for the user in the date range
	expenses, err := u.expenseRepo.GetByUserIDAndDateRange(ctx, req.UserID, req.StartDate, req.EndDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get expenses: %w", err)
	}

	// Convert to export format
	var exportedExpenses []ExportedExpense

	for _, expense := range expenses {
		categoryName := "Uncategorized"
		if expense.CategoryID != nil {
			cat, _ := u.categoryRepo.GetByID(ctx, *expense.CategoryID)
			if cat != nil {
				categoryName = cat.Name
			}
		}

		exportedExpenses = append(exportedExpenses, ExportedExpense{
			ID:          expense.ID,
			Date:        expense.ExpenseDate.Format("2006-01-02"),
			Description: expense.Description,
			Amount:      expense.Amount,
			Category:    categoryName,
			Account:     expense.Account,
			CreatedAt:   expense.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:   expense.UpdatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	return &ExportData{
		Format:       req.Format,
		ExportedAt:   time.Now(),
		PeriodStart:  req.StartDate,
		PeriodEnd:    req.EndDate,
		TotalRecords: len(exportedExpenses),
		Data:         exportedExpenses,
	}, nil
}

// ExportAsJSON exports expenses as JSON
func (u *DataExportUseCase) ExportAsJSON(ctx context.Context, req *ExportRequest) ([]byte, error) {
	req.Format = "json"
	data, err := u.Execute(ctx, req)
	if err != nil {
		return nil, err
	}

	return json.MarshalIndent(data, "", "  ")
}

// ExportAsCSV exports expenses as CSV
func (u *DataExportUseCase) ExportAsCSV(ctx context.Context, req *ExportRequest) ([]byte, error) {
	req.Format = "csv"
	data, err := u.Execute(ctx, req)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)
	defer writer.Flush()

	// Write header
	headers := []string{"ID", "Date", "Description", "Amount", "Category", "Account", "CreatedAt", "UpdatedAt"}
	if err := writer.Write(headers); err != nil {
		return nil, fmt.Errorf("failed to write CSV header: %w", err)
	}

	// Write data rows
	for _, exp := range data.Data {
		record := []string{
			exp.ID,
			exp.Date,
			exp.Description,
			fmt.Sprintf("%.2f", exp.Amount),
			exp.Category,
			exp.Account,
			exp.CreatedAt,
			exp.UpdatedAt,
		}
		if err := writer.Write(record); err != nil {
			return nil, fmt.Errorf("failed to write CSV row: %w", err)
		}
	}

	return buf.Bytes(), nil
}

// SummaryExportRequest represents a request for summary export
type SummaryExportRequest struct {
	UserID    string
	StartDate time.Time
	EndDate   time.Time
}

// SummaryData represents summary export data
type SummaryData struct {
	UserID           string             `json:"user_id"`
	PeriodStart      time.Time          `json:"period_start"`
	PeriodEnd        time.Time          `json:"period_end"`
	TotalExpenses    float64            `json:"total_expenses"`
	TransactionCount int                `json:"transaction_count"`
	AverageExpense   float64            `json:"average_expense"`
	CategoryTotals   map[string]float64 `json:"category_totals"`
	DailyAverages    float64            `json:"daily_average"`
	ExportedAt       time.Time          `json:"exported_at"`
}

// ExportSummary exports a summary of expenses
func (u *DataExportUseCase) ExportSummary(ctx context.Context, req *SummaryExportRequest) (*SummaryData, error) {
	if req.UserID == "" {
		return nil, fmt.Errorf("user_id is required")
	}

	// Get all expenses for the user in the date range
	expenses, err := u.expenseRepo.GetByUserIDAndDateRange(ctx, req.UserID, req.StartDate, req.EndDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get expenses: %w", err)
	}

	totalExpenses := 0.0
	categoryTotals := make(map[string]float64)

	for _, expense := range expenses {
		totalExpenses += expense.Amount

		categoryName := "Uncategorized"
		if expense.CategoryID != nil {
			cat, _ := u.categoryRepo.GetByID(ctx, *expense.CategoryID)
			if cat != nil {
				categoryName = cat.Name
			}
		}

		categoryTotals[categoryName] += expense.Amount
	}

	// Calculate averages
	avgExpense := 0.0
	if len(expenses) > 0 {
		avgExpense = totalExpenses / float64(len(expenses))
	}

	// Calculate daily average
	daysInPeriod := int(req.EndDate.Sub(req.StartDate).Hours() / 24)
	if daysInPeriod == 0 {
		daysInPeriod = 1
	}

	dailyAverage := totalExpenses / float64(daysInPeriod)

	return &SummaryData{
		UserID:           req.UserID,
		PeriodStart:      req.StartDate,
		PeriodEnd:        req.EndDate,
		TotalExpenses:    totalExpenses,
		TransactionCount: len(expenses),
		AverageExpense:   avgExpense,
		CategoryTotals:   categoryTotals,
		DailyAverages:    dailyAverage,
		ExportedAt:       time.Now(),
	}, nil
}
