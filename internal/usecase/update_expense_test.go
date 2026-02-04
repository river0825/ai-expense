package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/riverlin/aiexpense/internal/domain"
)

type mockExchangeRateService struct{}

func (m *mockExchangeRateService) Convert(ctx context.Context, amount float64, fromCurrency, toCurrency string, txTime time.Time) (float64, float64, error) {
	return amount * 31.5, 31.5, nil
}

func (m *mockExchangeRateService) RefreshRates(ctx context.Context) error {
	return nil
}

func (m *mockExchangeRateService) GetRate(ctx context.Context, fromCurrency, toCurrency string, txTime time.Time) (*domain.ExchangeRate, error) {
	return nil, nil
}

func TestUpdateExpenseUseCase_RecalculateHomeAmount(t *testing.T) {
	repo := NewMockExpenseRepository()
	categoryRepo := NewMockCategoryRepository()
	uc := NewUpdateExpenseUseCase(repo, categoryRepo, &mockExchangeRateService{})
	ctx := context.Background()

	expense := &domain.Expense{
		ID:             "exp_update_1",
		UserID:         "user_update_1",
		Description:    "Launch",
		OriginalAmount: 9482.39,
		Currency:       "TWD",
		HomeAmount:     9482.39,
		HomeCurrency:   "TWD",
		ExpenseDate:    time.Now(),
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if err := repo.Create(ctx, expense); err != nil {
		t.Fatalf("failed to seed expense: %v", err)
	}

	newAmount := 300.0
	newCurrency := "USD"
	if _, err := uc.Execute(ctx, &UpdateRequest{
		ID:             expense.ID,
		UserID:         expense.UserID,
		OriginalAmount: &newAmount,
		Currency:       &newCurrency,
	}); err != nil {
		t.Fatalf("update failed: %v", err)
	}

	updated, _ := repo.GetByID(ctx, expense.ID)
	if updated.OriginalAmount != newAmount {
		t.Fatalf("expected original amount %v, got %v", newAmount, updated.OriginalAmount)
	}
	if updated.Currency != newCurrency {
		t.Fatalf("expected currency %s, got %s", newCurrency, updated.Currency)
	}
	expectedHome := newAmount * 31.5
	if updated.HomeAmount != expectedHome {
		t.Fatalf("expected home amount %v, got %v", expectedHome, updated.HomeAmount)
	}
}
