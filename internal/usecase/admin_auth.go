package usecase

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/riverlin/aiexpense/internal/domain"
)

var (
	ErrInvalidAdminCredentials = errors.New("invalid admin credentials")
	ErrInvalidAdminToken       = errors.New("invalid admin token")
)

type AdminLoginUseCase struct {
	repo      domain.AdminAuthRepository
	jwtSecret []byte
	ttl       time.Duration
	now       func() time.Time
}

type AdminVerifyTokenUseCase struct {
	repo      domain.AdminAuthRepository
	jwtSecret []byte
	now       func() time.Time
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

func NewAdminLoginUseCase(repo domain.AdminAuthRepository, jwtSecret string) *AdminLoginUseCase {
	return &AdminLoginUseCase{
		repo:      repo,
		jwtSecret: []byte(jwtSecret),
		ttl:       24 * time.Hour,
		now:       time.Now,
	}
}

func NewAdminVerifyTokenUseCase(repo domain.AdminAuthRepository, jwtSecret string) *AdminVerifyTokenUseCase {
	return &AdminVerifyTokenUseCase{
		repo:      repo,
		jwtSecret: []byte(jwtSecret),
		now:       time.Now,
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

	claims := jwt.RegisteredClaims{
		Subject:   username,
		ExpiresAt: jwt.NewNumericDate(expiresAt),
		IssuedAt:  jwt.NewNumericDate(now),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(u.jwtSecret)
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
	claims := &jwt.RegisteredClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidAdminToken
		}
		return u.jwtSecret, nil
	})
	if err != nil || !token.Valid {
		return nil, ErrInvalidAdminToken
	}

	if claims.Subject == "" || claims.ExpiresAt == nil {
		return nil, ErrInvalidAdminToken
	}

	now := u.now().UTC()
	if claims.ExpiresAt.Time.Before(now) {
		return nil, ErrInvalidAdminToken
	}

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
