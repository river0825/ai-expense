package sqlite

import (
	"context"
	"database/sql"
	"time"

	"github.com/riverlin/aiexpense/internal/domain"
)

var _ domain.MetricsRepository = (*MetricsRepository)(nil)

type MetricsRepository struct {
	db *sql.DB
}

// NewMetricsRepository creates a new metrics repository
func NewMetricsRepository(db *sql.DB) *MetricsRepository {
	return &MetricsRepository{db: db}
}

// GetDailyActiveUsers retrieves DAU for a date range
func (r *MetricsRepository) GetDailyActiveUsers(ctx context.Context, from, to time.Time) ([]*domain.DailyMetrics, error) {
	const query = `
		SELECT
			DATE(created_at) as date,
			COUNT(DISTINCT user_id) as active_users,
			0 as total_expense,
			0 as expense_count,
			0.0 as average_expense
		FROM users
		WHERE created_at >= ? AND created_at <= ?
		GROUP BY DATE(created_at)
		ORDER BY date DESC
	`
	rows, err := r.db.QueryContext(ctx, query, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var metrics []*domain.DailyMetrics
	for rows.Next() {
		m := &domain.DailyMetrics{}
		var dateStr string
		if err := rows.Scan(&dateStr, &m.ActiveUsers, &m.TotalExpense, &m.ExpenseCount, &m.AverageExpense); err != nil {
			return nil, err
		}
		m.Date, _ = time.Parse("2006-01-02", dateStr)
		metrics = append(metrics, m)
	}
	return metrics, rows.Err()
}

// GetExpensesSummary retrieves expense totals by date
func (r *MetricsRepository) GetExpensesSummary(ctx context.Context, from, to time.Time) ([]*domain.DailyMetrics, error) {
	const query = `
		SELECT
			expense_date,
			0 as active_users,
			SUM(home_amount) as total_expense,
			COUNT(*) as expense_count,
			AVG(home_amount) as average_expense
		FROM expenses
		WHERE expense_date >= ? AND expense_date <= ?
		GROUP BY expense_date
		ORDER BY expense_date DESC
	`
	rows, err := r.db.QueryContext(ctx, query, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var metrics []*domain.DailyMetrics
	for rows.Next() {
		m := &domain.DailyMetrics{}
		if err := rows.Scan(&m.Date, &m.ActiveUsers, &m.TotalExpense, &m.ExpenseCount, &m.AverageExpense); err != nil {
			return nil, err
		}
		metrics = append(metrics, m)
	}
	return metrics, rows.Err()
}

// GetCategoryTrends retrieves expense breakdown by category
func (r *MetricsRepository) GetCategoryTrends(ctx context.Context, userID string, from, to time.Time) ([]*domain.CategoryMetrics, error) {
	const query = `
		SELECT
			c.id,
			c.name,
			COALESCE(SUM(e.home_amount), 0) as total,
			COUNT(e.id) as count
		FROM categories c
		LEFT JOIN expenses e ON c.id = e.category_id AND e.user_id = ? AND e.expense_date >= ? AND e.expense_date <= ?
		WHERE c.user_id = ?
		GROUP BY c.id, c.name
		ORDER BY total DESC
	`
	rows, err := r.db.QueryContext(ctx, query, userID, from, to, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var metrics []*domain.CategoryMetrics
	var totalAll float64

	// First pass: get all totals
	tempMetrics := []*domain.CategoryMetrics{}
	for rows.Next() {
		m := &domain.CategoryMetrics{}
		if err := rows.Scan(&m.CategoryID, &m.Category, &m.Total, &m.Count); err != nil {
			return nil, err
		}
		totalAll += m.Total
		tempMetrics = append(tempMetrics, m)
	}

	// Second pass: calculate percentages
	for _, m := range tempMetrics {
		if totalAll > 0 {
			m.Percent = (m.Total / totalAll) * 100
		}
		metrics = append(metrics, m)
	}

	return metrics, rows.Err()
}

// GetGrowthMetrics retrieves user growth metrics
func (r *MetricsRepository) GetGrowthMetrics(ctx context.Context, days int) (map[string]interface{}, error) {
	// Get total users
	var totalUsers int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM users").Scan(&totalUsers)
	if err != nil {
		return nil, err
	}

	// Get new users today
	var newUsersToday int
	today := time.Now().Format("2006-01-02")
	err = r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM users WHERE DATE(created_at) = ?", today).Scan(&newUsersToday)
	if err != nil {
		return nil, err
	}

	// Get new users this week
	var newUsersWeek int
	weekAgo := time.Now().AddDate(0, 0, -7).Format("2006-01-02")
	err = r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM users WHERE created_at >= ?", weekAgo).Scan(&newUsersWeek)
	if err != nil {
		return nil, err
	}

	// Get new users this month
	var newUsersMonth int
	monthAgo := time.Now().AddDate(0, -1, 0).Format("2006-01-02")
	err = r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM users WHERE created_at >= ?", monthAgo).Scan(&newUsersMonth)
	if err != nil {
		return nil, err
	}

	// Get total expenses
	var totalExpenses float64
	err = r.db.QueryRowContext(ctx, "SELECT COALESCE(SUM(home_amount), 0) FROM expenses").Scan(&totalExpenses)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"total_users":          totalUsers,
		"new_users_today":      newUsersToday,
		"new_users_this_week":  newUsersWeek,
		"new_users_this_month": newUsersMonth,
		"total_expenses":       totalExpenses,
	}, nil
}

// GetNewUsersPerDay retrieves new users created per day
func (r *MetricsRepository) GetNewUsersPerDay(ctx context.Context, from, to time.Time) ([]*domain.DailyMetrics, error) {
	const query = `
		SELECT
			DATE(created_at) as date,
			COUNT(*) as active_users,
			0 as total_expense,
			0 as expense_count,
			0.0 as average_expense
		FROM users
		WHERE created_at >= ? AND created_at <= ?
		GROUP BY DATE(created_at)
		ORDER BY date DESC
	`
	rows, err := r.db.QueryContext(ctx, query, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var metrics []*domain.DailyMetrics
	for rows.Next() {
		m := &domain.DailyMetrics{}
		var dateStr string
		if err := rows.Scan(&dateStr, &m.ActiveUsers, &m.TotalExpense, &m.ExpenseCount, &m.AverageExpense); err != nil {
			return nil, err
		}
		m.Date, _ = time.Parse("2006-01-02", dateStr)
		metrics = append(metrics, m)
	}
	return metrics, rows.Err()
}
