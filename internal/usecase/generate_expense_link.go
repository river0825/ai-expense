package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/riverlin/aiexpense/internal/domain"
)

type GenerateExpenseLinkUseCase struct {
	shortLinkRepo domain.ShortLinkRepository
	jwtSecret     string
	baseURL       string
}

func NewGenerateExpenseLinkUseCase(
	shortLinkRepo domain.ShortLinkRepository,
	jwtSecret string,
	baseURL string,
) *GenerateExpenseLinkUseCase {
	return &GenerateExpenseLinkUseCase{
		shortLinkRepo: shortLinkRepo,
		jwtSecret:     jwtSecret,
		baseURL:       baseURL,
	}
}

func (u *GenerateExpenseLinkUseCase) Execute(ctx context.Context, userID, expenseID string) (string, error) {
	// Generate JWT token for the user
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(7 * 24 * time.Hour).Unix(), // 7 days expiration
	})

	tokenString, err := token.SignedString([]byte(u.jwtSecret))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	// Create Short Link
	shortID, err := domain.GenerateShortID(6)
	if err != nil {
		return "", fmt.Errorf("failed to generate short ID: %w", err)
	}

	link := &domain.ShortLink{
		ID:           shortID,
		UserID:       userID,
		TargetToken:  tokenString,
		RedirectPath: fmt.Sprintf("/expenses?edit=%s", expenseID),
		ExpiresAt:    time.Now().Add(7 * 24 * time.Hour),
		CreatedAt:    time.Now(),
	}

	if err := u.shortLinkRepo.Create(ctx, link); err != nil {
		return "", fmt.Errorf("failed to save short link: %w", err)
	}

	return fmt.Sprintf("%s/r/%s", u.baseURL, shortID), nil
}
