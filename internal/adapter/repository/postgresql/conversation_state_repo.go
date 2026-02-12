package postgresql

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/riverlin/aiexpense/internal/domain"
)

var _ domain.ConversationStateRepository = (*ConversationStateRepository)(nil)

type ConversationStateRepository struct {
	db *sql.DB
}

func NewConversationStateRepository(db *sql.DB) *ConversationStateRepository {
	return &ConversationStateRepository{db: db}
}

func (r *ConversationStateRepository) GetByUserID(ctx context.Context, userID string) (*domain.ConversationState, error) {
	const query = `
		SELECT user_id, active_intent, pending_slots, status, expires_at, created_at, updated_at
		FROM conversation_states
		WHERE user_id = $1
	`

	var state domain.ConversationState
	var pendingSlotsRaw []byte
	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&state.UserID,
		&state.ActiveIntent,
		&pendingSlotsRaw,
		&state.Status,
		&state.ExpiresAt,
		&state.CreatedAt,
		&state.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	state.PendingSlots = map[string]string{}
	if len(pendingSlotsRaw) > 0 {
		_ = json.Unmarshal(pendingSlotsRaw, &state.PendingSlots)
	}

	return &state, nil
}

func (r *ConversationStateRepository) Upsert(ctx context.Context, state *domain.ConversationState) error {
	pendingSlotsRaw, err := json.Marshal(state.PendingSlots)
	if err != nil {
		return err
	}

	const query = `
		INSERT INTO conversation_states (user_id, active_intent, pending_slots, status, expires_at, created_at, updated_at)
		VALUES ($1, $2, $3::jsonb, $4, $5, $6, $7)
		ON CONFLICT (user_id)
		DO UPDATE SET
			active_intent = EXCLUDED.active_intent,
			pending_slots = EXCLUDED.pending_slots,
			status = EXCLUDED.status,
			expires_at = EXCLUDED.expires_at,
			updated_at = EXCLUDED.updated_at
	`

	_, err = r.db.ExecContext(ctx, query,
		state.UserID,
		state.ActiveIntent,
		pendingSlotsRaw,
		state.Status,
		state.ExpiresAt,
		state.CreatedAt,
		state.UpdatedAt,
	)
	return err
}

func (r *ConversationStateRepository) DeleteByUserID(ctx context.Context, userID string) error {
	const query = `DELETE FROM conversation_states WHERE user_id = $1`
	_, err := r.db.ExecContext(ctx, query, userID)
	return err
}
