package sqlite

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
		INSERT INTO ai_cost_logs (
			id, user_id, operation, provider, model, 
			input_tokens, output_tokens, total_tokens, 
			cost, currency, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err := r.db.ExecContext(ctx, query,
		log.ID, log.UserID, log.Operation, log.Provider, log.Model,
		log.InputTokens, log.OutputTokens, log.TotalTokens,
		log.Cost, log.Currency, log.CreatedAt,
	)
	return err
}

func (r *AICostRepository) GetByUserID(ctx context.Context, userID string, limit int) ([]*domain.AICostLog, error) {
	const query = `
		SELECT 
			id, user_id, operation, provider, model, 
			input_tokens, output_tokens, total_tokens, 
			cost, currency, created_at
		FROM ai_cost_logs
		WHERE user_id = ?
		ORDER BY created_at DESC
		LIMIT ?
	`
	rows, err := r.db.QueryContext(ctx, query, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []*domain.AICostLog
	for rows.Next() {
		log := &domain.AICostLog{}
		err := rows.Scan(
			&log.ID, &log.UserID, &log.Operation, &log.Provider, &log.Model,
			&log.InputTokens, &log.OutputTokens, &log.TotalTokens,
			&log.Cost, &log.Currency, &log.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		logs = append(logs, log)
	}
	return logs, rows.Err()
}

func (r *AICostRepository) GetSummary(ctx context.Context, from, to time.Time) (*domain.AICostSummary, error) {
	const query = `
		SELECT 
			COUNT(*) as total_calls,
			COALESCE(SUM(input_tokens), 0) as total_input_tokens,
			COALESCE(SUM(output_tokens), 0) as total_output_tokens,
			COALESCE(SUM(total_tokens), 0) as total_tokens,
			COALESCE(SUM(cost), 0) as total_cost
		FROM ai_cost_logs
		WHERE created_at >= ? AND created_at <= ?
	`
	summary := &domain.AICostSummary{Currency: "USD"}
	err := r.db.QueryRowContext(ctx, query, from, to).Scan(
		&summary.TotalCalls,
		&summary.TotalInputTokens,
		&summary.TotalOutputTokens,
		&summary.TotalTokens,
		&summary.TotalCost,
	)
	if err != nil {
		return nil, err
	}
	return summary, nil
}

func (r *AICostRepository) GetDailyStats(ctx context.Context, from, to time.Time) ([]*domain.AICostDailyStats, error) {
	const query = `
		SELECT 
			DATE(created_at) as date,
			COUNT(*) as calls,
			COALESCE(SUM(input_tokens), 0) as input_tokens,
			COALESCE(SUM(output_tokens), 0) as output_tokens,
			COALESCE(SUM(total_tokens), 0) as total_tokens,
			COALESCE(SUM(cost), 0) as cost
		FROM ai_cost_logs
		WHERE created_at >= ? AND created_at <= ?
		GROUP BY DATE(created_at)
		ORDER BY date ASC
	`
	rows, err := r.db.QueryContext(ctx, query, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []*domain.AICostDailyStats
	for rows.Next() {
		s := &domain.AICostDailyStats{}
		var dateStr string
		err := rows.Scan(&dateStr, &s.Calls, &s.InputTokens, &s.OutputTokens, &s.TotalTokens, &s.Cost)
		if err != nil {
			return nil, err
		}
		s.Date, _ = time.Parse("2006-01-02", dateStr)
		stats = append(stats, s)
	}
	return stats, rows.Err()
}

func (r *AICostRepository) GetByOperation(ctx context.Context, from, to time.Time) ([]*domain.AICostByOperation, error) {
	const query = `
		SELECT 
			operation,
			COUNT(*) as calls,
			COALESCE(SUM(input_tokens), 0) as input_tokens,
			COALESCE(SUM(output_tokens), 0) as output_tokens,
			COALESCE(SUM(total_tokens), 0) as total_tokens,
			COALESCE(SUM(cost), 0) as cost
		FROM ai_cost_logs
		WHERE created_at >= ? AND created_at <= ?
		GROUP BY operation
		ORDER BY total_tokens DESC
	`
	rows, err := r.db.QueryContext(ctx, query, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*domain.AICostByOperation
	var totalTokensAll int
	for rows.Next() {
		r := &domain.AICostByOperation{}
		err := rows.Scan(&r.Operation, &r.Calls, &r.InputTokens, &r.OutputTokens, &r.TotalTokens, &r.Cost)
		if err != nil {
			return nil, err
		}
		totalTokensAll += r.TotalTokens
		results = append(results, r)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	for _, r := range results {
		if totalTokensAll > 0 {
			r.Percent = float64(r.TotalTokens) / float64(totalTokensAll) * 100
		}
	}
	return results, nil
}

func (r *AICostRepository) GetByUserSummary(ctx context.Context, from, to time.Time, limit int) ([]*domain.AICostByUser, error) {
	const query = `
		SELECT 
			user_id,
			COUNT(*) as calls,
			COALESCE(SUM(input_tokens), 0) as input_tokens,
			COALESCE(SUM(output_tokens), 0) as output_tokens,
			COALESCE(SUM(total_tokens), 0) as total_tokens,
			COALESCE(SUM(cost), 0) as cost
		FROM ai_cost_logs
		WHERE created_at >= ? AND created_at <= ?
		GROUP BY user_id
		ORDER BY total_tokens DESC
		LIMIT ?
	`
	rows, err := r.db.QueryContext(ctx, query, from, to, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*domain.AICostByUser
	for rows.Next() {
		u := &domain.AICostByUser{}
		err := rows.Scan(&u.UserID, &u.Calls, &u.InputTokens, &u.OutputTokens, &u.TotalTokens, &u.Cost)
		if err != nil {
			return nil, err
		}
		results = append(results, u)
	}
	return results, rows.Err()
}
