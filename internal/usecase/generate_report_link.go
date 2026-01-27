package usecase

import (
	"context"
	"crypto/rand"
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
	// 1. Generate JWT (valid for 15 min)
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

	// 2. Generate Short Link (valid for 5 min)
	shortID := generateShortID()
	expiresAt := time.Now().Add(5 * time.Minute)

	shortLink := &domain.ShortLink{
		ID:          shortID,
		TargetToken: tokenString,
		ExpiresAt:   expiresAt,
		CreatedAt:   time.Now(),
	}

	if err := u.shortLinkRepo.Create(context.Background(), shortLink); err != nil {
		fmt.Printf("Failed to save short link to repository: %v\n", err)
		return "", fmt.Errorf("failed to create short link: %w", err)
	}

	// 3. Return Short Link URL
	return fmt.Sprintf("%s/r/%s", u.baseURL, shortID), nil
}

func generateShortID() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const length = 6
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		// Fallback to less secure random if crypto/rand fails (unlikely)
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	for i := 0; i < length; i++ {
		b[i] = charset[int(b[i])%len(charset)]
	}
	return string(b)
}
