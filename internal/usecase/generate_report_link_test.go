package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/riverlin/aiexpense/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockShortLinkRepository struct {
	mock.Mock
}

func (m *MockShortLinkRepository) Create(ctx context.Context, link *domain.ShortLink) error {
	args := m.Called(ctx, link)
	return args.Error(0)
}

func (m *MockShortLinkRepository) Get(ctx context.Context, id string) (*domain.ShortLink, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.ShortLink), args.Error(1)
}

func (m *MockShortLinkRepository) DeleteExpired(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func TestGenerateReportLinkUseCase_Execute(t *testing.T) {
	baseURL := "http://localhost:3000"
	mockRepo := new(MockShortLinkRepository)
	uc := NewGenerateReportLinkUseCase(baseURL, mockRepo)

	// Override secret for consistent testing
	uc.jwtSecret = []byte("test-secret")

	userID := "user123"

	// Expect short link creation
	mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(link *domain.ShortLink) bool {
		return len(link.ID) == 6 && link.ExpiresAt.After(time.Now())
	})).Return(nil)

	link, err := uc.Execute(userID)

	assert.NoError(t, err)
	// Link should now be a short link redirect
	assert.Contains(t, link, baseURL+"/r/")

	// Ensure ID is length 6
	shortID := link[len(baseURL+"/r/"):]
	assert.Len(t, shortID, 6)
}
