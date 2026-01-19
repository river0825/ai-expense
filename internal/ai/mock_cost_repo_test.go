package ai

import (
	"context"
	"time"

	"github.com/riverlin/aiexpense/internal/domain"
)

type MockAICostRepository struct {
	CreatedLogs []*domain.AICostLog
}

func (m *MockAICostRepository) Create(ctx context.Context, log *domain.AICostLog) error {
	m.CreatedLogs = append(m.CreatedLogs, log)
	return nil
}

func (m *MockAICostRepository) GetByUserID(ctx context.Context, userID string, limit int) ([]*domain.AICostLog, error) {
	return nil, nil
}

func (m *MockAICostRepository) GetSummary(ctx context.Context, from, to time.Time) (*domain.AICostSummary, error) {
	return &domain.AICostSummary{Currency: "USD"}, nil
}

func (m *MockAICostRepository) GetDailyStats(ctx context.Context, from, to time.Time) ([]*domain.AICostDailyStats, error) {
	return nil, nil
}

func (m *MockAICostRepository) GetByOperation(ctx context.Context, from, to time.Time) ([]*domain.AICostByOperation, error) {
	return nil, nil
}

func (m *MockAICostRepository) GetByUserSummary(ctx context.Context, from, to time.Time, limit int) ([]*domain.AICostByUser, error) {
	return nil, nil
}
