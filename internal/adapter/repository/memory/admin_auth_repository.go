package memory

import (
	"context"
	"sync"

	"github.com/riverlin/aiexpense/internal/domain"
)

var _ domain.AdminAuthRepository = (*AdminAuthRepository)(nil)

type AdminAuthRepository struct {
	mu       sync.RWMutex
	sessions map[string]*domain.AdminSession
}

func NewAdminAuthRepository() *AdminAuthRepository {
	return &AdminAuthRepository{
		sessions: make(map[string]*domain.AdminSession),
	}
}

func (r *AdminAuthRepository) CreateSession(_ context.Context, session *domain.AdminSession) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	copySession := *session
	r.sessions[session.TokenHash] = &copySession
	return nil
}

func (r *AdminAuthRepository) GetSessionByTokenHash(_ context.Context, tokenHash string) (*domain.AdminSession, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	session, ok := r.sessions[tokenHash]
	if !ok {
		return nil, nil
	}

	copySession := *session
	return &copySession, nil
}

func (r *AdminAuthRepository) DeleteSessionByTokenHash(_ context.Context, tokenHash string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.sessions, tokenHash)
	return nil
}
