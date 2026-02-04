package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/riverlin/aiexpense/internal/domain"
)

func TestGenerateReport_AccountInExpenseDetails(t *testing.T) {
	expenseRepo := NewMockExpenseRepository()
	categoryRepo := NewMockCategoryRepository()

	uc := NewGenerateReportUseCase(expenseRepo, categoryRepo, nil)

	ctx := context.Background()
	userID := "user_report_account"
	now := time.Now()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	endOfMonth := startOfMonth.AddDate(0, 1, 0).Add(-time.Nanosecond)

	t.Run("ExpenseDetail includes account field", func(t *testing.T) {
		expenseRepo.Create(ctx, &domain.Expense{
			ID:          "exp_report_1",
			UserID:      userID,
			Description: "Grocery shopping",
			Amount:      75.50,
			Account:     "Debit Card",
			ExpenseDate: now,
			CreatedAt:   now,
			UpdatedAt:   now,
		})

		req := &ReportRequest{
			UserID:     userID,
			ReportType: "monthly",
			StartDate:  startOfMonth,
			EndDate:    endOfMonth,
		}

		report, err := uc.Execute(ctx, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(report.TopExpenses) == 0 {
			t.Fatal("expected at least one expense in report")
		}

		found := false
		for _, exp := range report.TopExpenses {
			if exp.ID == "exp_report_1" {
				found = true
				if exp.Account != "Debit Card" {
					t.Errorf("expected account 'Debit Card', got '%s'", exp.Account)
				}
			}
		}
		if !found {
			t.Error("expense exp_report_1 not found in report")
		}
	})

	t.Run("ExpenseDetail includes currency metadata", func(t *testing.T) {
		expenseRepo.Create(ctx, &domain.Expense{
			ID:             "exp_currency_meta",
			UserID:         userID,
			Description:    "Business trip",
			Amount:         9000,
			OriginalAmount: 300,
			Currency:       "USD",
			HomeAmount:     9482.39,
			HomeCurrency:   "TWD",
			Account:        "Corporate Card",
			ExpenseDate:    now,
			CreatedAt:      now,
			UpdatedAt:      now,
		})

		req := &ReportRequest{
			UserID:     userID,
			ReportType: "monthly",
			StartDate:  startOfMonth,
			EndDate:    endOfMonth,
		}

		report, err := uc.Execute(ctx, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		for _, exp := range report.TopExpenses {
			if exp.ID == "exp_currency_meta" {
				if exp.OriginalAmount != 300 {
					t.Fatalf("expected original amount 300, got %v", exp.OriginalAmount)
				}
				if exp.OriginalCurrency != "USD" {
					t.Fatalf("expected original currency USD, got %s", exp.OriginalCurrency)
				}
				if exp.HomeAmount != 9482.39 {
					t.Fatalf("expected home amount 9482.39, got %v", exp.HomeAmount)
				}
				if exp.HomeCurrency != "TWD" {
					t.Fatalf("expected home currency TWD, got %s", exp.HomeCurrency)
				}
				return
			}
		}
		t.Fatal("expense exp_currency_meta not found in report")
	})

	t.Run("Multiple expenses with different accounts", func(t *testing.T) {
		expenseRepo.Create(ctx, &domain.Expense{
			ID:          "exp_report_cash",
			UserID:      userID,
			Description: "Coffee",
			Amount:      5.00,
			Account:     "Cash",
			ExpenseDate: now,
			CreatedAt:   now,
			UpdatedAt:   now,
		})
		expenseRepo.Create(ctx, &domain.Expense{
			ID:          "exp_report_credit",
			UserID:      userID,
			Description: "Online purchase",
			Amount:      120.00,
			Account:     "Credit Card",
			ExpenseDate: now,
			CreatedAt:   now,
			UpdatedAt:   now,
		})

		req := &ReportRequest{
			UserID:     userID,
			ReportType: "monthly",
			StartDate:  startOfMonth,
			EndDate:    endOfMonth,
		}

		report, err := uc.Execute(ctx, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		accountMap := make(map[string]string)
		for _, exp := range report.TopExpenses {
			accountMap[exp.ID] = exp.Account
		}

		if accountMap["exp_report_cash"] != "Cash" {
			t.Errorf("expected account 'Cash' for exp_report_cash, got '%s'", accountMap["exp_report_cash"])
		}
		if accountMap["exp_report_credit"] != "Credit Card" {
			t.Errorf("expected account 'Credit Card' for exp_report_credit, got '%s'", accountMap["exp_report_credit"])
		}
	})

	t.Run("Default Cash account preserved in report", func(t *testing.T) {
		expenseRepo.Create(ctx, &domain.Expense{
			ID:          "exp_default_account",
			UserID:      userID,
			Description: "Street food",
			Amount:      8.00,
			Account:     "Cash",
			ExpenseDate: now,
			CreatedAt:   now,
			UpdatedAt:   now,
		})

		req := &ReportRequest{
			UserID:     userID,
			ReportType: "monthly",
			StartDate:  startOfMonth,
			EndDate:    endOfMonth,
		}

		report, err := uc.Execute(ctx, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		for _, exp := range report.TopExpenses {
			if exp.ID == "exp_default_account" {
				if exp.Account != "Cash" {
					t.Errorf("expected default account 'Cash', got '%s'", exp.Account)
				}
				return
			}
		}
		t.Error("expense exp_default_account not found in report")
	})
}
