package memory

import (
	"context"
	"testing"
	"time"

	"github.com/riverlin/aiexpense/internal/domain"
)

func TestAdminAuthRepository_CreateAndGetSession(t *testing.T) {
	repo := NewAdminAuthRepository()
	ctx := context.Background()
	now := time.Now().UTC()

	session := &domain.AdminSession{
		TokenHash: "token-hash",
		Username:  "admin",
		CreatedAt: now,
		ExpiresAt: now.Add(time.Hour),
	}

	if err := repo.CreateSession(ctx, session); err != nil {
		t.Fatalf("CreateSession returned error: %v", err)
	}

	got, err := repo.GetSessionByTokenHash(ctx, session.TokenHash)
	if err != nil {
		t.Fatalf("GetSessionByTokenHash returned error: %v", err)
	}

	if got == nil {
		t.Fatal("expected session, got nil")
	}

	if got.Username != session.Username {
		t.Fatalf("expected username %q, got %q", session.Username, got.Username)
	}

	if got.ExpiresAt != session.ExpiresAt {
		t.Fatalf("expected expires_at %v, got %v", session.ExpiresAt, got.ExpiresAt)
	}
}

func TestAdminAuthRepository_DeleteSession(t *testing.T) {
	repo := NewAdminAuthRepository()
	ctx := context.Background()

	session := &domain.AdminSession{
		TokenHash: "token-hash",
		Username:  "admin",
		CreatedAt: time.Now().UTC(),
		ExpiresAt: time.Now().UTC().Add(time.Hour),
	}

	if err := repo.CreateSession(ctx, session); err != nil {
		t.Fatalf("CreateSession returned error: %v", err)
	}

	if err := repo.DeleteSessionByTokenHash(ctx, session.TokenHash); err != nil {
		t.Fatalf("DeleteSessionByTokenHash returned error: %v", err)
	}

	got, err := repo.GetSessionByTokenHash(ctx, session.TokenHash)
	if err != nil {
		t.Fatalf("GetSessionByTokenHash returned error: %v", err)
	}

	if got != nil {
		t.Fatal("expected nil session after delete")
	}
}
