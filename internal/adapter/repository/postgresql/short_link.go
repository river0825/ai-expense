package postgresql

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/riverlin/aiexpense/internal/domain"
)

var _ domain.ShortLinkRepository = (*ShortLinkRepository)(nil)

type ShortLinkRepository struct {
	db *sql.DB
}

func NewShortLinkRepository(db *sql.DB) *ShortLinkRepository {
	return &ShortLinkRepository{db: db}
}

func (r *ShortLinkRepository) Create(ctx context.Context, link *domain.ShortLink) error {
	query := `
		INSERT INTO short_links (id, user_id, target_token, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := r.db.ExecContext(ctx, query, link.ID, link.UserID, link.TargetToken, link.ExpiresAt, link.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to create short link: %w", err)
	}
	return nil
}

func (r *ShortLinkRepository) Get(ctx context.Context, id string) (*domain.ShortLink, error) {
	query := `
		SELECT id, user_id, target_token, expires_at, created_at
		FROM short_links
		WHERE id = $1 AND expires_at > $2 AND deprecated_at IS NULL
	`
	row := r.db.QueryRowContext(ctx, query, id, time.Now())

	var link domain.ShortLink
	err := row.Scan(&link.ID, &link.UserID, &link.TargetToken, &link.ExpiresAt, &link.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("short link not found or expired")
		}
		return nil, fmt.Errorf("failed to get short link: %w", err)
	}
	return &link, nil
}

func (r *ShortLinkRepository) DeprecateByUserID(ctx context.Context, userID string) error {
	query := `UPDATE short_links SET deprecated_at = $1 WHERE user_id = $2 AND deprecated_at IS NULL`
	_, err := r.db.ExecContext(ctx, query, time.Now(), userID)
	if err != nil {
		return fmt.Errorf("failed to deprecate short links: %w", err)
	}
	return nil
}

func (r *ShortLinkRepository) DeleteExpired(ctx context.Context) error {
	query := `DELETE FROM short_links WHERE expires_at <= $1`
	_, err := r.db.ExecContext(ctx, query, time.Now())
	if err != nil {
		return fmt.Errorf("failed to delete expired short links: %w", err)
	}
	return nil
}
