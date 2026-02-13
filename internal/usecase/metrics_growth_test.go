package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/riverlin/aiexpense/internal/domain"
)

type growthMetricsRepoStub struct {
	growth map[string]interface{}
}

func (s *growthMetricsRepoStub) GetDailyActiveUsers(ctx context.Context, from, to time.Time) ([]*domain.DailyMetrics, error) {
	return nil, nil
}

func (s *growthMetricsRepoStub) GetExpensesSummary(ctx context.Context, from, to time.Time) ([]*domain.DailyMetrics, error) {
	return nil, nil
}

func (s *growthMetricsRepoStub) GetCategoryTrends(ctx context.Context, userID string, from, to time.Time) ([]*domain.CategoryMetrics, error) {
	return nil, nil
}

func (s *growthMetricsRepoStub) GetGrowthMetrics(ctx context.Context, days int) (map[string]interface{}, error) {
	return s.growth, nil
}

func (s *growthMetricsRepoStub) GetNewUsersPerDay(ctx context.Context, from, to time.Time) ([]*domain.DailyMetrics, error) {
	return nil, nil
}

func TestGetGrowthMetricsHandlesMissingKeysWithoutPanic(t *testing.T) {
	uc := NewMetricsUseCase(&growthMetricsRepoStub{growth: map[string]interface{}{
		"total_users":    10,
		"total_expenses": 100,
		"new_users":      3,
	}})

	resp, err := uc.GetGrowthMetrics(context.Background(), &GrowthMetricsRequest{Days: 30})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.TotalUsers != 10 {
		t.Fatalf("expected total users 10, got %d", resp.TotalUsers)
	}
	if resp.NewUsersThisMonth != 3 {
		t.Fatalf("expected new users this month 3, got %d", resp.NewUsersThisMonth)
	}
	if resp.NewUsersToday != 0 {
		t.Fatalf("expected new users today 0, got %d", resp.NewUsersToday)
	}
	if resp.TotalExpenses != 100 {
		t.Fatalf("expected total expenses 100, got %f", resp.TotalExpenses)
	}
}

func TestGetGrowthMetricsAcceptsFloatValues(t *testing.T) {
	uc := NewMetricsUseCase(&growthMetricsRepoStub{growth: map[string]interface{}{
		"total_users":          float64(5),
		"new_users_today":      float64(1),
		"new_users_this_week":  float64(2),
		"new_users_this_month": float64(4),
		"total_expenses":       float64(50.5),
	}})

	resp, err := uc.GetGrowthMetrics(context.Background(), &GrowthMetricsRequest{Days: 30})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.TotalUsers != 5 || resp.NewUsersToday != 1 || resp.NewUsersThisWeek != 2 || resp.NewUsersThisMonth != 4 {
		t.Fatalf("unexpected integer conversions: %+v", resp)
	}
	if resp.TotalExpenses != 50.5 {
		t.Fatalf("expected total expenses 50.5, got %f", resp.TotalExpenses)
	}
}
