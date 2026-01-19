package usecase

import (
	"context"
	"time"

	"github.com/riverlin/aiexpense/internal/domain"
)

type AICostUseCase struct {
	aiCostRepo domain.AICostRepository
}

func NewAICostUseCase(aiCostRepo domain.AICostRepository) *AICostUseCase {
	return &AICostUseCase{aiCostRepo: aiCostRepo}
}

type AICostMetricsRequest struct {
	Days int
}

type AICostMetricsResponse struct {
	Summary     *domain.AICostSummary       `json:"summary"`
	DailyStats  []*domain.AICostDailyStats  `json:"daily_stats"`
	ByOperation []*domain.AICostByOperation `json:"by_operation"`
	TopUsers    []*domain.AICostByUser      `json:"top_users"`
}

func (u *AICostUseCase) GetAICostMetrics(ctx context.Context, req *AICostMetricsRequest) (*AICostMetricsResponse, error) {
	if req.Days == 0 {
		req.Days = 30
	}

	to := time.Now()
	from := to.AddDate(0, 0, -req.Days)

	summary, err := u.aiCostRepo.GetSummary(ctx, from, to)
	if err != nil {
		return nil, err
	}

	dailyStats, err := u.aiCostRepo.GetDailyStats(ctx, from, to)
	if err != nil {
		return nil, err
	}

	byOperation, err := u.aiCostRepo.GetByOperation(ctx, from, to)
	if err != nil {
		return nil, err
	}

	topUsers, err := u.aiCostRepo.GetByUserSummary(ctx, from, to, 10)
	if err != nil {
		return nil, err
	}

	return &AICostMetricsResponse{
		Summary:     summary,
		DailyStats:  dailyStats,
		ByOperation: byOperation,
		TopUsers:    topUsers,
	}, nil
}

type AICostSummaryRequest struct {
	Days int
}

func (u *AICostUseCase) GetSummary(ctx context.Context, req *AICostSummaryRequest) (*domain.AICostSummary, error) {
	if req.Days == 0 {
		req.Days = 30
	}

	to := time.Now()
	from := to.AddDate(0, 0, -req.Days)

	return u.aiCostRepo.GetSummary(ctx, from, to)
}

type AICostDailyRequest struct {
	Days int
}

func (u *AICostUseCase) GetDailyStats(ctx context.Context, req *AICostDailyRequest) ([]*domain.AICostDailyStats, error) {
	if req.Days == 0 {
		req.Days = 30
	}

	to := time.Now()
	from := to.AddDate(0, 0, -req.Days)

	return u.aiCostRepo.GetDailyStats(ctx, from, to)
}

type AICostByOperationRequest struct {
	Days int
}

func (u *AICostUseCase) GetByOperation(ctx context.Context, req *AICostByOperationRequest) ([]*domain.AICostByOperation, error) {
	if req.Days == 0 {
		req.Days = 30
	}

	to := time.Now()
	from := to.AddDate(0, 0, -req.Days)

	return u.aiCostRepo.GetByOperation(ctx, from, to)
}

type AICostByUserRequest struct {
	Days  int
	Limit int
}

func (u *AICostUseCase) GetTopUsers(ctx context.Context, req *AICostByUserRequest) ([]*domain.AICostByUser, error) {
	if req.Days == 0 {
		req.Days = 30
	}
	if req.Limit == 0 {
		req.Limit = 10
	}

	to := time.Now()
	from := to.AddDate(0, 0, -req.Days)

	return u.aiCostRepo.GetByUserSummary(ctx, from, to, req.Limit)
}
