package usecase

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/riverlin/aiexpense/internal/domain"
)

var _ domain.GenerateReportLinkUseCase = (*GenerateReportLinkUseCase)(nil)

type GenerateReportLinkUseCase struct {
	baseURL       string
	jwtSecret     []byte
	shortLinkRepo domain.ShortLinkRepository
}

func NewGenerateReportLinkUseCase(baseURL string, shortLinkRepo domain.ShortLinkRepository) *GenerateReportLinkUseCase {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "default-secret-do-not-use-in-prod"
	}

	return &GenerateReportLinkUseCase{
		baseURL:       baseURL,
		jwtSecret:     []byte(secret),
		shortLinkRepo: shortLinkRepo,
	}
}

func (u *GenerateReportLinkUseCase) Execute(userID string) (string, error) {
	// 1. Deprecate any existing short links for this user
	if err := u.shortLinkRepo.DeprecateByUserID(context.Background(), userID); err != nil {
		fmt.Printf("Failed to deprecate existing short links: %v\n", err)
	}

	// 2. Generate JWT (valid for 7 days)
	claims := jwt.MapClaims{
		"sub":  userID,
		"exp":  time.Now().Add(7 * 24 * time.Hour).Unix(),
		"type": "report_access",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(u.jwtSecret)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	// 3. Generate Short Link (no time limit - expires with JWT)
	shortID, err := domain.GenerateShortID(6)
	if err != nil {
		return "", fmt.Errorf("failed to generate short ID: %w", err)
	}
	expiresAt := time.Now().Add(7 * 24 * time.Hour)

	shortLink := &domain.ShortLink{
		ID:          shortID,
		UserID:      userID,
		TargetToken: tokenString,
		ExpiresAt:   expiresAt,
		CreatedAt:   time.Now(),
	}

	if err := u.shortLinkRepo.Create(context.Background(), shortLink); err != nil {
		fmt.Printf("Failed to save short link to repository: %v\n", err)
		return "", fmt.Errorf("failed to create short link: %w", err)
	}

	// 4. Return Short Link URL
	return fmt.Sprintf("%s/r/%s", u.baseURL, shortID), nil
}
