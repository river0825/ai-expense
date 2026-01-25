package slack

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/riverlin/aiexpense/internal/domain"
	"github.com/stretchr/testify/mock"
)

// MockMessageProcessor for testing
type MockMessageProcessor struct {
	mock.Mock
}

func (m *MockMessageProcessor) Execute(ctx context.Context, msg *domain.UserMessage) (*domain.MessageResponse, error) {
	args := m.Called(ctx, msg)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.MessageResponse), args.Error(1)
}

func TestSlackHandler_HandleWebhook_Success(t *testing.T) {
	// Setup
	mockUC := new(MockMessageProcessor)

	// Create handler with empty signing secret to bypass signature verification in tests if implementation allows,
	// OR we need to generate valid signature.
	// Based on implementation: if h.signingSecret != "" -> verify.
	// So passing "" skips verification logic inside HandleWebhook (if check exists) or fails.
	// Let's check implementation again... `if h.signingSecret != "" { ... }` in HandleWebhook.
	// So passing "" is safe for testing without signatures.
	handler := NewHandler("", mockUC, nil)

	// Expectations
	mockUC.On("Execute", mock.Anything, mock.MatchedBy(func(msg *domain.UserMessage) bool {
		return msg.UserID == "U12345" && msg.Content == "breakfast $20" && msg.Source == "slack"
	})).Return(&domain.MessageResponse{Text: "Saved"}, nil)

	// Create valid webhook payload (simplified)
	event := SlackEvent{
		Token: "token",
		Type:  "event_callback",
		Event: &Event{
			Type:      "message",
			User:      "U12345",
			Text:      "breakfast $20",
			Channel:   "C12345",
			Timestamp: "1625097600.000000",
		},
	}
	body, _ := json.Marshal(event)

	req := httptest.NewRequest("POST", "/webhook/slack", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.HandleWebhook(w, req)

	// Verify response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	// Wait for async goroutine
	time.Sleep(100 * time.Millisecond)

	mockUC.AssertExpectations(t)
}
