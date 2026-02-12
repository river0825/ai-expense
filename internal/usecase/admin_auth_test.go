package usecase

import (
	"context"
	"testing"

	"github.com/riverlin/aiexpense/internal/domain"
)

type MockAdminAuthRepository struct {
	sessions map[string]*domain.AdminSession
}

func NewMockAdminAuthRepository() *MockAdminAuthRepository {
	return &MockAdminAuthRepository{sessions: make(map[string]*domain.AdminSession)}
}

func (m *MockAdminAuthRepository) GetAdminByUsername(ctx context.Context, username string) (*domain.AdminCredentials, error) {
	if username == "admin" {
		return &domain.AdminCredentials{Username: "admin"}, nil
	}
	return nil, nil
}

func (m *MockAdminAuthRepository) CreateSession(ctx context.Context, session *domain.AdminSession) error {
	m.sessions[session.ID] = session
	return nil
}

func (m *MockAdminAuthRepository) GetSessionByToken(ctx context.Context, token string) (*domain.AdminSession, error) {
	for _, s := range m.sessions {
		if s.Token == token {
			return s, nil
		}
	}
	return nil, nil
}

func (m *MockAdminAuthRepository) DeleteSession(ctx context.Context, sessionID string) error {
	delete(m.sessions, sessionID)
	return nil
}

func TestLoginSuccess(t *testing.T) {
	repo := NewMockAdminAuthRepository()
	uc := NewLoginUseCase(repo, "test-secret")

	resp, err := uc.Execute(context.Background(), LoginRequest{Username: "admin", Password: "admin123"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Token == "" {
		t.Error("expected token")
	}
}

func TestLoginInvalidCredentials(t *testing.T) {
	repo := NewMockAdminAuthRepository()
	uc := NewLoginUseCase(repo, "test-secret")

	_, err := uc.Execute(context.Background(), LoginRequest{Username: "admin", Password: "wrong"})
	if err == nil {
		t.Error("expected error")
	}
}

func TestVerifyValidToken(t *testing.T) {
	repo := NewMockAdminAuthRepository()
	loginUC := NewLoginUseCase(repo, "test-secret")
	loginResp, _ := loginUC.Execute(context.Background(), LoginRequest{Username: "admin", Password: "admin123"})

	verifyUC := NewVerifyTokenUseCase(repo, "test-secret")
	_, err := verifyUC.Execute(context.Background(), loginResp.Token)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestVerifyInvalidToken(t *testing.T) {
	repo := NewMockAdminAuthRepository()
	verifyUC := NewVerifyTokenUseCase(repo, "test-secret")

	_, err := verifyUC.Execute(context.Background(), "invalid-token")
	if err == nil {
		t.Error("expected error")
	}
}
