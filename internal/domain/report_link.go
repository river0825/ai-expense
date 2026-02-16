package domain

import (
	"context"
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

// GenerateExpenseLinkUseCase defines the interface for generating expense edit links
type GenerateExpenseLinkUseCase interface {
	Execute(ctx context.Context, userID, expenseID string) (string, error)
}
