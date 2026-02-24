package usecase

import (
	"context"
	"testing"

	"github.com/riverlin/aiexpense/internal/domain"
	"github.com/riverlin/aiexpense/internal/pkg/jwtutil"
)

type MockAdminAuthRepository struct {
	sessions map[string]*domain.AdminSession
}

func NewMockAdminAuthRepository() *MockAdminAuthRepository {
	return &MockAdminAuthRepository{sessions: make(map[string]*domain.AdminSession)}
}

func (m *MockAdminAuthRepository) CreateSession(ctx context.Context, session *domain.AdminSession) error {
	m.sessions[session.TokenHash] = session
	return nil
}

func (m *MockAdminAuthRepository) GetSessionByTokenHash(ctx context.Context, tokenHash string) (*domain.AdminSession, error) {
	s, ok := m.sessions[tokenHash]
	if !ok {
		return nil, nil
	}
	return s, nil
}

func (m *MockAdminAuthRepository) DeleteSessionByTokenHash(ctx context.Context, tokenHash string) error {
	delete(m.sessions, tokenHash)
	return nil
}

func TestLoginSuccess(t *testing.T) {
	repo := NewMockAdminAuthRepository()
	tm := jwtutil.NewTokenManager("test-secret")
	uc := NewAdminLoginUseCase(repo, tm)

	resp, err := uc.Execute(context.Background(), "admin", "admin123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Token == "" {
		t.Error("expected token")
	}
}

func TestLoginInvalidCredentials(t *testing.T) {
	repo := NewMockAdminAuthRepository()
	tm := jwtutil.NewTokenManager("test-secret")
	uc := NewAdminLoginUseCase(repo, tm)

	_, err := uc.Execute(context.Background(), "admin", "wrong")
	if err == nil {
		t.Error("expected error")
	}
}

func TestVerifyValidToken(t *testing.T) {
	repo := NewMockAdminAuthRepository()
	tm := jwtutil.NewTokenManager("test-secret")
	loginUC := NewAdminLoginUseCase(repo, tm)
	loginResp, _ := loginUC.Execute(context.Background(), "admin", "admin123")

	verifyUC := NewAdminVerifyTokenUseCase(repo, tm)
	_, err := verifyUC.Execute(context.Background(), loginResp.Token)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestVerifyInvalidToken(t *testing.T) {
	repo := NewMockAdminAuthRepository()
	tm := jwtutil.NewTokenManager("test-secret")
	verifyUC := NewAdminVerifyTokenUseCase(repo, tm)

	_, err := verifyUC.Execute(context.Background(), "invalid-token")
	if err == nil {
		t.Error("expected error")
	}
}
