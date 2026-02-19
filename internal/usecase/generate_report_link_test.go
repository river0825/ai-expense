package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/riverlin/aiexpense/internal/domain"
	"github.com/riverlin/aiexpense/internal/pkg/jwtutil"
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

func (m *MockShortLinkRepository) DeprecateByUserID(ctx context.Context, userID string) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func TestGenerateReportLinkUseCase_Execute(t *testing.T) {
	baseURL := "http://localhost:3000"
	mockRepo := new(MockShortLinkRepository)
	tokenManager := jwtutil.NewTokenManager("test-secret")
	uc := NewGenerateReportLinkUseCase(baseURL, mockRepo, tokenManager)

	userID := "user123"

	mockRepo.On("DeprecateByUserID", mock.Anything, userID).Return(nil)
	mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(link *domain.ShortLink) bool {
		return len(link.ID) == 6 && link.ExpiresAt.After(time.Now()) && link.UserID == userID
	})).Return(nil)

	link, err := uc.Execute(userID)

	assert.NoError(t, err)
	assert.Contains(t, link, baseURL+"/r/")

	shortID := link[len(baseURL+"/r/"):]
	assert.Len(t, shortID, 6)

	mockRepo.AssertCalled(t, "DeprecateByUserID", mock.Anything, userID)
}
