package usecase

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/riverlin/aiexpense/internal/domain"
)

func TestDataExport_AccountField(t *testing.T) {
	expenseRepo := NewMockExpenseRepository()
	categoryRepo := NewMockCategoryRepository()

	uc := NewDataExportUseCase(expenseRepo, categoryRepo)

	ctx := context.Background()
	userID := "user_export_account"
	now := time.Now()
	expenseDate := now.Add(-time.Hour)
	startDate := now.AddDate(0, 0, -30)
	endDate := now.AddDate(0, 0, 1)

	expenseRepo.Create(ctx, &domain.Expense{
		ID:          "exp_export_cash",
		UserID:      userID,
		Description: "Lunch",
		Amount:      15.00,
		Account:     "Cash",
		ExpenseDate: expenseDate,
		CreatedAt:   now,
		UpdatedAt:   now,
	})
	expenseRepo.Create(ctx, &domain.Expense{
		ID:          "exp_export_credit",
		UserID:      userID,
		Description: "Electronics",
		Amount:      299.99,
		Account:     "Credit Card",
		ExpenseDate: expenseDate,
		CreatedAt:   now,
		UpdatedAt:   now,
	})
	expenseRepo.Create(ctx, &domain.Expense{
		ID:          "exp_export_debit",
		UserID:      userID,
		Description: "Groceries",
		Amount:      85.50,
		Account:     "Debit Card",
		ExpenseDate: expenseDate,
		CreatedAt:   now,
		UpdatedAt:   now,
	})

	t.Run("ExportAsJSON includes account field", func(t *testing.T) {
		req := &ExportRequest{
			UserID:    userID,
			Format:    "json",
			StartDate: startDate,
			EndDate:   endDate,
		}

		jsonData, err := uc.ExportAsJSON(ctx, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		var exportData ExportData
		if err := json.Unmarshal(jsonData, &exportData); err != nil {
			t.Fatalf("failed to parse JSON: %v", err)
		}

		if len(exportData.Data) == 0 {
			t.Fatal("expected expenses in export data")
		}

		accountMap := make(map[string]string)
		for _, exp := range exportData.Data {
			accountMap[exp.ID] = exp.Account
		}

		if accountMap["exp_export_cash"] != "Cash" {
			t.Errorf("expected account 'Cash', got '%s'", accountMap["exp_export_cash"])
		}
		if accountMap["exp_export_credit"] != "Credit Card" {
			t.Errorf("expected account 'Credit Card', got '%s'", accountMap["exp_export_credit"])
		}
		if accountMap["exp_export_debit"] != "Debit Card" {
			t.Errorf("expected account 'Debit Card', got '%s'", accountMap["exp_export_debit"])
		}
	})

	t.Run("ExportAsCSV includes account column", func(t *testing.T) {
		req := &ExportRequest{
			UserID:    userID,
			Format:    "csv",
			StartDate: startDate,
			EndDate:   endDate,
		}

		csvData, err := uc.ExportAsCSV(ctx, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		reader := csv.NewReader(strings.NewReader(string(csvData)))
		records, err := reader.ReadAll()
		if err != nil {
			t.Fatalf("failed to parse CSV: %v", err)
		}

		if len(records) < 2 {
			t.Fatal("expected header and at least one data row")
		}

		header := records[0]
		accountColIndex := -1
		for i, col := range header {
			if col == "Account" {
				accountColIndex = i
				break
			}
		}

		if accountColIndex == -1 {
			t.Fatal("Account column not found in CSV header")
		}

		accountValues := make(map[string]string)
		for _, row := range records[1:] {
			if len(row) > accountColIndex {
				accountValues[row[0]] = row[accountColIndex]
			}
		}

		if accountValues["exp_export_cash"] != "Cash" {
			t.Errorf("expected account 'Cash' in CSV, got '%s'", accountValues["exp_export_cash"])
		}
		if accountValues["exp_export_credit"] != "Credit Card" {
			t.Errorf("expected account 'Credit Card' in CSV, got '%s'", accountValues["exp_export_credit"])
		}
		if accountValues["exp_export_debit"] != "Debit Card" {
			t.Errorf("expected account 'Debit Card' in CSV, got '%s'", accountValues["exp_export_debit"])
		}
	})

	t.Run("Execute includes account in ExportedExpense", func(t *testing.T) {
		req := &ExportRequest{
			UserID:    userID,
			Format:    "json",
			StartDate: startDate,
			EndDate:   endDate,
		}

		data, err := uc.Execute(ctx, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		for _, exp := range data.Data {
			if exp.Account == "" {
				t.Errorf("account should not be empty for expense %s", exp.ID)
			}
		}
	})
}
