package usecase

import (
	"context"
	"time"

	"github.com/riverlin/aiexpense/internal/domain"
)

// OverviewRequest represents a request for analytics overview
type OverviewRequest struct {
	Period           string
	ComparisonPeriod string
}

// RevenueDelta represents the change between current and previous periods
type RevenueDelta struct {
	NRRDelta float64 `json:"nrr_delta"`
	GRRDelta float64 `json:"grr_delta"`
	MRRDelta float64 `json:"mrr_delta"`
}

// OverviewResponse represents the analytics overview response
type OverviewResponse struct {
	Current  *domain.RevenueRetentionReport `json:"current"`
	Previous *domain.RevenueRetentionReport `json:"previous,omitempty"`
	Delta    *RevenueDelta                  `json:"delta,omitempty"`
}

// AdminAnalyticsUseCase handles admin analytics operations
type AdminAnalyticsUseCase struct {
	revenueRepo domain.RevenueRetentionRepository
}

// NewAdminAnalyticsUseCase creates a new AdminAnalyticsUseCase
func NewAdminAnalyticsUseCase(revenueRepo domain.RevenueRetentionRepository) *AdminAnalyticsUseCase {
	return &AdminAnalyticsUseCase{
		revenueRepo: revenueRepo,
	}
}

// Overview returns revenue and retention overview for the specified period
func (uc *AdminAnalyticsUseCase) Overview(ctx context.Context, req OverviewRequest) (*OverviewResponse, error) {
	// Calculate date range from period
	now := time.Now()
	periodStart, periodEnd := parsePeriod(now, req.Period)

	// Get current period report
	current, err := uc.revenueRepo.GetReport(ctx, periodStart, periodEnd)
	if err != nil {
		return nil, err
	}

	response := &OverviewResponse{
		Current: current,
	}

	// Get comparison period if requested
	if req.ComparisonPeriod != "" {
		compStart, compEnd := parsePeriod(periodStart, req.ComparisonPeriod)
		previous, err := uc.revenueRepo.GetReport(ctx, compStart, compEnd)
		if err == nil && previous != nil {
			response.Previous = previous
			response.Delta = calculateDelta(current, previous)
		}
	}

	return response, nil
}

// parsePeriod calculates start and end dates from a period string
func parsePeriod(reference time.Time, period string) (time.Time, time.Time) {
	end := reference
	var start time.Time

	switch period {
	case "7d":
		start = end.AddDate(0, 0, -7)
	case "30d":
		start = end.AddDate(0, 0, -30)
	case "90d":
		start = end.AddDate(0, 0, -90)
	case "prev_7d":
		end = end.AddDate(0, 0, -7)
		start = end.AddDate(0, 0, -7)
	case "prev_30d":
		end = end.AddDate(0, 0, -30)
		start = end.AddDate(0, 0, -30)
	case "prev_90d":
		end = end.AddDate(0, 0, -90)
		start = end.AddDate(0, 0, -90)
	default:
		// Default to 30 days
		start = end.AddDate(0, 0, -30)
	}

	return start, end
}

// calculateDelta computes the difference between current and previous periods
func calculateDelta(current, previous *domain.RevenueRetentionReport) *RevenueDelta {
	if previous == nil || previous.StartingMRR == 0 {
		return &RevenueDelta{}
	}

	return &RevenueDelta{
		NRRDelta: current.NRR - previous.NRR,
		GRRDelta: current.GRR - previous.GRR,
		MRRDelta: ((current.EndingMRR - previous.EndingMRR) / previous.EndingMRR) * 100,
	}
}
