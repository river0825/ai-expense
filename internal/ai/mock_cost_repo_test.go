package ai

import (
	"context"
	"github.com/riverlin/aiexpense/internal/domain"
)

// MockAICostRepository for testing
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
