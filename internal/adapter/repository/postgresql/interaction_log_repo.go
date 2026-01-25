package postgresql

import (
	"context"
	"database/sql"

	"github.com/riverlin/aiexpense/internal/domain"
)

// InteractionLogRepository implements domain.InteractionLogRepository for PostgreSQL
type InteractionLogRepository struct {
	db *sql.DB
}

// NewInteractionLogRepository creates a new PostgreSQL interaction log repository
func NewInteractionLogRepository(db *sql.DB) *InteractionLogRepository {
	return &InteractionLogRepository{db: db}
}

// Create creates a new interaction log entry
func (r *InteractionLogRepository) Create(ctx context.Context, log *domain.InteractionLog) error {
	query := `
		INSERT INTO interaction_logs (
			id, user_id, user_input, system_prompt, 
			ai_raw_response, bot_final_reply, duration_ms, 
			error, timestamp
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err := r.db.ExecContext(ctx, query,
		log.ID,
		log.UserID,
		log.UserInput,
		log.SystemPrompt,
		log.AIRawResponse,
		log.BotFinalReply,
		log.DurationMs,
		log.Error,
		log.Timestamp,
	)
	return err
}
