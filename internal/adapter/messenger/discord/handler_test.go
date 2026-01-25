package discord

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

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

func TestDiscordHandler_HandleWebhook_Success(t *testing.T) {
	// Setup
	mockUC := new(MockMessageProcessor)
	handler := NewHandler("test_bot_token", mockUC, nil)

	// Expectations
	mockUC.On("Execute", mock.Anything, mock.MatchedBy(func(msg *domain.UserMessage) bool {
		return msg.UserID == "user_123" && msg.Content == "breakfast $20" && msg.Source == "discord"
	})).Return(&domain.MessageResponse{Text: "Saved"}, nil)

	// Create valid webhook payload (simplified)
	payload := DiscordInteraction{
		Type:  2, // ApplicationCommand
		ID:    "interaction_123",
		Token: "interaction_token",
		User:  User{ID: "user_123", Username: "test_user"},
		Data:  InteractionData{Content: "breakfast $20"},
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest("POST", "/webhook/discord", bytes.NewReader(body))
	w := httptest.NewRecorder()
	handler.HandleWebhook(w, req)

	// Verify response (Discord uses type 4 for immediate response)
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	mockUC.AssertExpectations(t)
}
