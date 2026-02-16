package domain

import (
	"context"
	"crypto/rand"
	"time"
)

func GenerateShortID(length int) (string, error) {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	for i := 0; i < length; i++ {
		b[i] = charset[int(b[i])%len(charset)]
	}
	return string(b), nil
}

type ShortLink struct {
	ID           string     `json:"id" db:"id"`
	UserID       string     `json:"user_id" db:"user_id"`
	TargetToken  string     `json:"target_token" db:"target_token"`
	RedirectPath string     `json:"redirect_path" db:"redirect_path"`
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
