package usecase

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/riverlin/aiexpense/internal/domain"
	"github.com/riverlin/aiexpense/internal/pkg/jwtutil"
)

var (
	ErrInvalidAdminCredentials = errors.New("invalid admin credentials")
	ErrInvalidAdminToken       = errors.New("invalid admin token")
)

type AdminLoginUseCase struct {
	repo         domain.AdminAuthRepository
	tokenManager *jwtutil.TokenManager
	ttl          time.Duration
	now          func() time.Time
}

type AdminVerifyTokenUseCase struct {
	repo         domain.AdminAuthRepository
	tokenManager *jwtutil.TokenManager
	now          func() time.Time
}

type AdminLogoutUseCase struct {
	repo domain.AdminAuthRepository
}

type AdminLoginResult struct {
	Token     string
	ExpiresAt time.Time
}

type AdminTokenClaims struct {
	Username  string
	ExpiresAt time.Time
}

func NewAdminLoginUseCase(repo domain.AdminAuthRepository, tokenManager *jwtutil.TokenManager) *AdminLoginUseCase {
	return &AdminLoginUseCase{
		repo:         repo,
		tokenManager: tokenManager,
		ttl:          24 * time.Hour,
		now:          time.Now,
	}
}

func NewAdminVerifyTokenUseCase(repo domain.AdminAuthRepository, tokenManager *jwtutil.TokenManager) *AdminVerifyTokenUseCase {
	return &AdminVerifyTokenUseCase{
		repo:         repo,
		tokenManager: tokenManager,
		now:          time.Now,
	}
}

func NewAdminLogoutUseCase(repo domain.AdminAuthRepository) *AdminLogoutUseCase {
	return &AdminLogoutUseCase{repo: repo}
}

func (u *AdminLoginUseCase) Execute(ctx context.Context, username, password string) (*AdminLoginResult, error) {
	if username != "admin" || password != "admin123" {
		return nil, ErrInvalidAdminCredentials
	}

	now := u.now().UTC()
	expiresAt := now.Add(u.ttl)

	tokenString, err := u.tokenManager.GenerateUserToken(username, u.ttl)
	if err != nil {
		return nil, fmt.Errorf("sign token: %w", err)
	}

	if err := u.repo.CreateSession(ctx, &domain.AdminSession{
		TokenHash: hashToken(tokenString),
		Username:  username,
		CreatedAt: now,
		ExpiresAt: expiresAt,
	}); err != nil {
		return nil, fmt.Errorf("create session: %w", err)
	}

	return &AdminLoginResult{Token: tokenString, ExpiresAt: expiresAt}, nil
}

func (u *AdminVerifyTokenUseCase) Execute(ctx context.Context, tokenString string) (*AdminTokenClaims, error) {
	claims, err := u.tokenManager.ValidateToken(tokenString)
	if err != nil {
		return nil, ErrInvalidAdminToken
	}

	// Verify structure but we rely on session lookup
	username, err := u.tokenManager.GetUserIDFromClaims(claims)
	if err != nil {
		return nil, ErrInvalidAdminToken
	}
	_ = username // Suppress unused error, or we could validating against session later


	// ValidateToken checks expiration, but we need to check session in DB
	now := u.now().UTC()

	session, err := u.repo.GetSessionByTokenHash(ctx, hashToken(tokenString))
	if err != nil {
		return nil, fmt.Errorf("get session: %w", err)
	}
	if session == nil || session.ExpiresAt.Before(now) {
		return nil, ErrInvalidAdminToken
	}

	return &AdminTokenClaims{Username: session.Username, ExpiresAt: session.ExpiresAt}, nil
}

func (u *AdminLogoutUseCase) Execute(ctx context.Context, tokenString string) error {
	return u.repo.DeleteSessionByTokenHash(ctx, hashToken(tokenString))
}

func hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}
