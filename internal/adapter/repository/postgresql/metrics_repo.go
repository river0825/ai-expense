package postgresql

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

func NewMetricsRepository(db *sql.DB) *MetricsRepository {
	return &MetricsRepository{db: db}
}

func (r *MetricsRepository) GetDailyActiveUsers(ctx context.Context, from, to time.Time) ([]*domain.DailyMetrics, error) {
	const query = `
		SELECT DATE(created_at) as date, COUNT(DISTINCT user_id) as count
		FROM users
		WHERE created_at >= $1 AND created_at <= $2
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
		metric := &domain.DailyMetrics{}
		if err := rows.Scan(&metric.Date, &metric.ActiveUsers); err != nil {
			return nil, err
		}
		metrics = append(metrics, metric)
	}
	return metrics, rows.Err()
}

func (r *MetricsRepository) GetExpensesSummary(ctx context.Context, from, to time.Time) ([]*domain.DailyMetrics, error) {
	const query = `
		SELECT DATE(expense_date) as date, COUNT(*) as count
		FROM expenses
		WHERE expense_date >= $1 AND expense_date <= $2
		GROUP BY DATE(expense_date)
		ORDER BY date DESC
	`

	rows, err := r.db.QueryContext(ctx, query, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var metrics []*domain.DailyMetrics
	for rows.Next() {
		metric := &domain.DailyMetrics{}
		if err := rows.Scan(&metric.Date, &metric.ExpenseCount); err != nil {
			return nil, err
		}
		metrics = append(metrics, metric)
	}
	return metrics, rows.Err()
}

func (r *MetricsRepository) GetCategoryTrends(ctx context.Context, userID string, from, to time.Time) ([]*domain.CategoryMetrics, error) {
	const query = `
		SELECT
			COALESCE(c.name, 'Uncategorized') as category_name,
			COUNT(e.id) as expense_count,
			SUM(e.amount) as total_amount
		FROM expenses e
		LEFT JOIN categories c ON e.category_id = c.id
		WHERE e.user_id = $1 AND e.expense_date >= $2 AND e.expense_date <= $3
		GROUP BY c.name
		ORDER BY total_amount DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var metrics []*domain.CategoryMetrics
	for rows.Next() {
		metric := &domain.CategoryMetrics{}
		if err := rows.Scan(&metric.Category, &metric.Count, &metric.Total); err != nil {
			return nil, err
		}
		metrics = append(metrics, metric)
	}
	return metrics, rows.Err()
}

func (r *MetricsRepository) GetGrowthMetrics(ctx context.Context, days int) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	// Total users
	var totalUsers int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM users").Scan(&totalUsers)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	result["total_users"] = totalUsers

	// Total expenses
	var totalExpenses int
	err = r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM expenses").Scan(&totalExpenses)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	result["total_expenses"] = totalExpenses

	// New users in period
	var newUsers int
	fromDate := time.Now().AddDate(0, 0, -days)
	err = r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM users WHERE created_at >= $1", fromDate).Scan(&newUsers)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	result["new_users"] = newUsers

	return result, nil
}

func (r *MetricsRepository) GetNewUsersPerDay(ctx context.Context, from, to time.Time) ([]*domain.DailyMetrics, error) {
	const query = `
		SELECT DATE(created_at) as date, COUNT(*) as count
		FROM users
		WHERE created_at >= $1 AND created_at <= $2
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
		metric := &domain.DailyMetrics{}
		if err := rows.Scan(&metric.Date, &metric.ActiveUsers); err != nil {
			return nil, err
		}
		metrics = append(metrics, metric)
	}
	return metrics, rows.Err()
}
