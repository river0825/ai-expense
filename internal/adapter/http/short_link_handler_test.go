package http

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/riverlin/aiexpense/internal/domain"
	"github.com/stretchr/testify/mock"
)

// MockShortLinkRepository
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

func TestShortLinkHandler_HandleRedirect(t *testing.T) {
	// Setup
	mockRepo := new(MockShortLinkRepository)
	dashboardURL := "http://localhost:3000"
	handler := NewShortLinkHandler(mockRepo, dashboardURL)

	t.Run("Success - Valid ID", func(t *testing.T) {
		shortID := "abc1234"
		targetToken := "valid.jwt.token"

		mockRepo.On("Get", mock.Anything, shortID).Return(&domain.ShortLink{
			ID:          shortID,
			TargetToken: targetToken,
			ExpiresAt:   time.Now().Add(5 * time.Minute),
		}, nil)

		req := httptest.NewRequest("GET", "/r/"+shortID, nil)
		// Manually set path value for Go 1.22+ mux behavior simulation if needed,
		// but since we rely on `r.PathValue` which is 1.22 feature, we need to ensure test env supports it.
		// For standard `httptest` with `http.NewServeMux` in 1.22:
		req.SetPathValue("id", shortID)

		w := httptest.NewRecorder()

		handler.HandleRedirect(w, req)

		// Check Status Code (Found/Redirect)
		if w.Code != http.StatusFound {
			t.Errorf("Expected status 302 Found, got %d", w.Code)
		}

		// Check Location Header
		expectedLocation := fmt.Sprintf("%s/reports?token=%s", dashboardURL, targetToken)
		if loc := w.Header().Get("Location"); loc != expectedLocation {
			t.Errorf("Expected Location %s, got %s", expectedLocation, loc)
		}

		// Check Cookie
		cookies := w.Result().Cookies()
		foundCookie := false
		for _, c := range cookies {
			if c.Name == "report_token" && c.Value == targetToken {
				foundCookie = true
				break
			}
		}
		if !foundCookie {
			t.Error("Expected report_token cookie to be set")
		}
	})

	t.Run("Failure - Invalid ID", func(t *testing.T) {
		shortID := "invalid"
		mockRepo.On("Get", mock.Anything, shortID).Return(nil, fmt.Errorf("not found"))

		req := httptest.NewRequest("GET", "/r/"+shortID, nil)
		req.SetPathValue("id", shortID)
		w := httptest.NewRecorder()

		handler.HandleRedirect(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status 404 Not Found, got %d", w.Code)
		}
	})
}
