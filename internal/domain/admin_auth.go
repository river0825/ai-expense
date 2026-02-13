package domain

import "time"

type AdminSession struct {
	TokenHash string
	Username  string
	CreatedAt time.Time
	ExpiresAt time.Time
}
