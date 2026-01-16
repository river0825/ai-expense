package usecase

import (
	"context"
	"time"

	"github.com/riverlin/aiexpense/internal/domain"
)

// MetricsUseCase handles metrics aggregation and reporting
type MetricsUseCase struct {
	metricsRepo domain.MetricsRepository
}

// NewMetricsUseCase creates a new metrics use case
func NewMetricsUseCase(metricsRepo domain.MetricsRepository) *MetricsUseCase {
	return &MetricsUseCase{
		metricsRepo: metricsRepo,
	}
}

// DailyActiveUsersRequest represents a request for DAU metrics
type DailyActiveUsersRequest struct {
	Days int // Number of days to retrieve (default: 30)
}

// DailyActiveUsersResponse represents DAU metrics
type DailyActiveUsersResponse struct {
	Data              []*domain.DailyMetrics `json:"data"`
	TotalActiveUsers  int                    `json:"total_active_users"`
	AverageDailyUsers float64                `json:"average_daily_users"`
}

// GetDailyActiveUsers retrieves DAU metrics
func (u *MetricsUseCase) GetDailyActiveUsers(ctx context.Context, req *DailyActiveUsersRequest) (*DailyActiveUsersResponse, error) {
	if req.Days == 0 {
		req.Days = 30
	}

	to := time.Now()
	from := to.AddDate(0, 0, -req.Days)

	metrics, err := u.metricsRepo.GetDailyActiveUsers(ctx, from, to)
	if err != nil {
		return nil, err
	}

	// Calculate aggregates
	totalActive := 0
	for _, m := range metrics {
		totalActive += m.ActiveUsers
	}

	averageDaily := 0.0
	if len(metrics) > 0 {
		averageDaily = float64(totalActive) / float64(len(metrics))
	}

	return &DailyActiveUsersResponse{
		Data:              metrics,
		TotalActiveUsers:  totalActive,
		AverageDailyUsers: averageDaily,
	}, nil
}

// ExpensesSummaryRequest represents a request for expense summary
type ExpensesSummaryRequest struct {
	Days int // Number of days to retrieve (default: 30)
}

// ExpensesSummaryResponse represents expense summary metrics
type ExpensesSummaryResponse struct {
	Data                   []*domain.DailyMetrics `json:"data"`
	TotalExpenses          float64                `json:"total_expenses"`
	AverageDailyExpenses   float64                `json:"average_daily_expenses"`
	TotalTransactions      int                    `json:"total_transactions"`
	AverageTransactionSize float64                `json:"average_transaction_size"`
}

// GetExpensesSummary retrieves expense summary metrics
func (u *MetricsUseCase) GetExpensesSummary(ctx context.Context, req *ExpensesSummaryRequest) (*ExpensesSummaryResponse, error) {
	if req.Days == 0 {
		req.Days = 30
	}

	to := time.Now()
	from := to.AddDate(0, 0, -req.Days)

	metrics, err := u.metricsRepo.GetExpensesSummary(ctx, from, to)
	if err != nil {
		return nil, err
	}

	// Calculate aggregates
	totalExpenses := 0.0
	totalTransactions := 0
	for _, m := range metrics {
		totalExpenses += m.TotalExpense
		totalTransactions += m.ExpenseCount
	}

	averageDaily := 0.0
	if len(metrics) > 0 {
		averageDaily = totalExpenses / float64(len(metrics))
	}

	averageTransaction := 0.0
	if totalTransactions > 0 {
		averageTransaction = totalExpenses / float64(totalTransactions)
	}

	return &ExpensesSummaryResponse{
		Data:                   metrics,
		TotalExpenses:          totalExpenses,
		AverageDailyExpenses:   averageDaily,
		TotalTransactions:      totalTransactions,
		AverageTransactionSize: averageTransaction,
	}, nil
}

// CategoryTrendsRequest represents a request for category trends
type CategoryTrendsRequest struct {
	UserID string
	Days   int // Number of days to retrieve (default: 30)
}

// CategoryTrendsResponse represents category trend metrics
type CategoryTrendsResponse struct {
	Data              []*domain.CategoryMetrics `json:"data"`
	TotalExpenses     float64                   `json:"total_expenses"`
	TopCategory       string                    `json:"top_category"`
	TopCategoryAmount float64                   `json:"top_category_amount"`
}

// GetCategoryTrends retrieves category trend metrics
func (u *MetricsUseCase) GetCategoryTrends(ctx context.Context, req *CategoryTrendsRequest) (*CategoryTrendsResponse, error) {
	if req.Days == 0 {
		req.Days = 30
	}

	to := time.Now()
	from := to.AddDate(0, 0, -req.Days)

	metrics, err := u.metricsRepo.GetCategoryTrends(ctx, req.UserID, from, to)
	if err != nil {
		return nil, err
	}

	// Calculate aggregates
	totalExpenses := 0.0
	topCategory := ""
	topAmount := 0.0

	for _, m := range metrics {
		totalExpenses += m.Total
		if m.Total > topAmount {
			topAmount = m.Total
			topCategory = m.Category
		}
	}

	return &CategoryTrendsResponse{
		Data:              metrics,
		TotalExpenses:     totalExpenses,
		TopCategory:       topCategory,
		TopCategoryAmount: topAmount,
	}, nil
}

// GrowthMetricsRequest represents a request for growth metrics
type GrowthMetricsRequest struct {
	Days int // Number of days to retrieve (default: 30)
}

// GrowthMetricsResponse represents growth metrics
type GrowthMetricsResponse struct {
	TotalUsers          int     `json:"total_users"`
	NewUsersToday       int     `json:"new_users_today"`
	NewUsersThisWeek    int     `json:"new_users_this_week"`
	NewUsersThisMonth   int     `json:"new_users_this_month"`
	TotalExpenses       float64 `json:"total_expenses"`
	AverageExpenseUser  float64 `json:"average_expense_per_user"`
	DailyGrowthPercent  float64 `json:"daily_growth_percent"`
	WeeklyGrowthPercent float64 `json:"weekly_growth_percent"`
}

// GetGrowthMetrics retrieves growth metrics
func (u *MetricsUseCase) GetGrowthMetrics(ctx context.Context, req *GrowthMetricsRequest) (*GrowthMetricsResponse, error) {
	if req.Days == 0 {
		req.Days = 30
	}

	metricsData, err := u.metricsRepo.GetGrowthMetrics(ctx, req.Days)
	if err != nil {
		return nil, err
	}

	// Convert map to typed response
	resp := &GrowthMetricsResponse{
		TotalUsers:        metricsData["total_users"].(int),
		NewUsersToday:     metricsData["new_users_today"].(int),
		NewUsersThisWeek:  metricsData["new_users_this_week"].(int),
		NewUsersThisMonth: metricsData["new_users_this_month"].(int),
		TotalExpenses:     metricsData["total_expenses"].(float64),
	}

	// Calculate derived metrics
	if resp.TotalUsers > 0 {
		resp.AverageExpenseUser = resp.TotalExpenses / float64(resp.TotalUsers)
	}

	// Calculate growth rates
	if resp.NewUsersThisMonth > 0 && resp.TotalUsers > 0 {
		resp.DailyGrowthPercent = (float64(resp.NewUsersToday) / float64(resp.TotalUsers)) * 100
		resp.WeeklyGrowthPercent = (float64(resp.NewUsersThisWeek) / float64(resp.TotalUsers)) * 100
	}

	return resp, nil
}
