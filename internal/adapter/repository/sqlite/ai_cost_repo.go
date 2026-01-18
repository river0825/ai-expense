package sqlite

import (
	"context"
	"database/sql"

	"github.com/riverlin/aiexpense/internal/domain"
)

// AICostRepository implements domain.AICostRepository using SQLite
type AICostRepository struct {
	db *sql.DB
}

// NewAICostRepository creates a new AI cost repository
func NewAICostRepository(db *sql.DB) *AICostRepository {
	return &AICostRepository{db: db}
}

// Create creates a new cost log entry
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

// GetByUserID retrieves cost logs for a user
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
