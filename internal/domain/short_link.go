package domain

import (
	"context"
	"time"
)

type ShortLink struct {
	ID           string     `json:"id" db:"id"`
	UserID       string     `json:"user_id" db:"user_id"`
	TargetToken  string     `json:"target_token" db:"target_token"`
	ExpiresAt    time.Time  `json:"expires_at" db:"expires_at"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	DeprecatedAt *time.Time `json:"deprecated_at" db:"deprecated_at"`
}

type ShortLinkRepository interface {
	Create(ctx context.Context, link *ShortLink) error
	Get(ctx context.Context, id string) (*ShortLink, error)
	DeprecateByUserID(ctx context.Context, userID string) error
	DeleteExpired(ctx context.Context) error
}
