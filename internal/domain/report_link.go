package domain

import (
	"time"
)

type ReportTokenClaims struct {
	UserID    string    `json:"sub"`
	ExpiresAt time.Time `json:"exp"`
	Type      string    `json:"type"`
}

type GenerateReportLinkUseCase interface {
	Execute(userID string) (string, error)
}
