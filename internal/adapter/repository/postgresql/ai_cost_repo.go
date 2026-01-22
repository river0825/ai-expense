package postgresql

import (
	"context"
	"database/sql"
	"time"

	"github.com/riverlin/aiexpense/internal/domain"
)

var _ domain.AICostRepository = (*AICostRepository)(nil)

type AICostRepository struct {
	db *sql.DB
}

func NewAICostRepository(db *sql.DB) *AICostRepository {
	return &AICostRepository{db: db}
}

func (r *AICostRepository) Create(ctx context.Context, log *domain.AICostLog) error {
	const query = `
		INSERT INTO ai_cost_logs (id, user_id, model, input_tokens, output_tokens, cost, operation, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := r.db.ExecContext(ctx, query,
		log.ID, log.UserID, log.Model,
		log.InputTokens, log.OutputTokens, log.Cost,
		log.Operation, log.CreatedAt,
	)
	return err
}

func (r *AICostRepository) GetByUserID(ctx context.Context, userID string, limit int) ([]*domain.AICostLog, error) {
	const query = `
		SELECT id, user_id, model, input_tokens, output_tokens, cost, operation, created_at
		FROM ai_cost_logs
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`

	rows, err := r.db.QueryContext(ctx, query, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []*domain.AICostLog
	for rows.Next() {
		log := &domain.AICostLog{}
		if err := rows.Scan(
			&log.ID, &log.UserID, &log.Model,
			&log.InputTokens, &log.OutputTokens, &log.Cost,
			&log.Operation, &log.CreatedAt,
		); err != nil {
			return nil, err
		}
		logs = append(logs, log)
	}
	return logs, rows.Err()
}

func (r *AICostRepository) GetSummary(ctx context.Context, from, to time.Time) (*domain.AICostSummary, error) {
	const query = `
		SELECT
			COUNT(*) as total_requests,
			SUM(input_tokens) as total_input_tokens,
			SUM(output_tokens) as total_output_tokens,
			SUM(cost) as total_cost
		FROM ai_cost_logs
		WHERE created_at >= $1 AND created_at <= $2
	`

	summary := &domain.AICostSummary{}
	err := r.db.QueryRowContext(ctx, query, from, to).Scan(
		&summary.TotalCalls,
		&summary.TotalInputTokens,
		&summary.TotalOutputTokens,
		&summary.TotalCost,
	)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	return summary, nil
}

func (r *AICostRepository) GetDailyStats(ctx context.Context, from, to time.Time) ([]*domain.AICostDailyStats, error) {
	const query = `
		SELECT
			DATE(created_at) as date,
			COUNT(*) as calls,
			SUM(input_tokens) as input_tokens,
			SUM(output_tokens) as output_tokens,
			SUM(cost) as cost
		FROM ai_cost_logs
		WHERE created_at >= $1 AND created_at <= $2
		GROUP BY DATE(created_at)
		ORDER BY date DESC
	`

	rows, err := r.db.QueryContext(ctx, query, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []*domain.AICostDailyStats
	for rows.Next() {
		stat := &domain.AICostDailyStats{}
		if err := rows.Scan(
			&stat.Date,
			&stat.Calls,
			&stat.InputTokens,
			&stat.OutputTokens,
			&stat.Cost,
		); err != nil {
			return nil, err
		}
		stats = append(stats, stat)
	}
	return stats, rows.Err()
}

func (r *AICostRepository) GetByOperation(ctx context.Context, from, to time.Time) ([]*domain.AICostByOperation, error) {
	const query = `
		SELECT
			operation,
			COUNT(*) as calls,
			SUM(input_tokens) as input_tokens,
			SUM(output_tokens) as output_tokens,
			SUM(cost) as cost
		FROM ai_cost_logs
		WHERE created_at >= $1 AND created_at <= $2
		GROUP BY operation
		ORDER BY cost DESC
	`

	rows, err := r.db.QueryContext(ctx, query, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*domain.AICostByOperation
	for rows.Next() {
		result := &domain.AICostByOperation{}
		if err := rows.Scan(
			&result.Operation,
			&result.Calls,
			&result.InputTokens,
			&result.OutputTokens,
			&result.Cost,
		); err != nil {
			return nil, err
		}
		results = append(results, result)
	}
	return results, rows.Err()
}

func (r *AICostRepository) GetByUserSummary(ctx context.Context, from, to time.Time, limit int) ([]*domain.AICostByUser, error) {
	const query = `
		SELECT
			user_id,
			COUNT(*) as calls,
			SUM(input_tokens) as input_tokens,
			SUM(output_tokens) as output_tokens,
			SUM(cost) as cost
		FROM ai_cost_logs
		WHERE created_at >= $1 AND created_at <= $2
		GROUP BY user_id
		ORDER BY cost DESC
		LIMIT $3
	`

	rows, err := r.db.QueryContext(ctx, query, from, to, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*domain.AICostByUser
	for rows.Next() {
		result := &domain.AICostByUser{}
		if err := rows.Scan(
			&result.UserID,
			&result.Calls,
			&result.InputTokens,
			&result.OutputTokens,
			&result.Cost,
		); err != nil {
			return nil, err
		}
		results = append(results, result)
	}
	return results, rows.Err()
}
