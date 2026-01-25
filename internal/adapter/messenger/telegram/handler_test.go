package telegram

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

func TestTelegramHandler_HandleWebhook_Success(t *testing.T) {
	// Setup
	mockUC := new(MockMessageProcessor)
	handler := NewHandler("test_bot_token", mockUC, nil)

	// Expectations
	mockUC.On("Execute", mock.Anything, mock.MatchedBy(func(msg *domain.UserMessage) bool {
		return msg.UserID == "telegram_12345" && msg.Content == "breakfast $20" && msg.Source == "telegram"
	})).Return(&domain.MessageResponse{Text: "Saved"}, nil)

	// Create valid webhook payload (simplified)
	update := TelegramUpdate{
		UpdateID: 123,
		Message: &struct {
			MessageID int64 `json:"message_id"`
			From      *struct {
				ID        int64  `json:"id"`
				IsBot     bool   `json:"is_bot"`
				FirstName string `json:"first_name"`
				Username  string `json:"username"`
			} `json:"from"`
			Chat *struct {
				ID   int64  `json:"id"`
				Type string `json:"type"`
			} `json:"chat"`
			Date int64  `json:"date"`
			Text string `json:"text"`
		}{
			MessageID: 1,
			From: &struct {
				ID        int64  `json:"id"`
				IsBot     bool   `json:"is_bot"`
				FirstName string `json:"first_name"`
				Username  string `json:"username"`
			}{ID: 12345, FirstName: "Test"},
			Chat: &struct {
				ID   int64  `json:"id"`
				Type string `json:"type"`
			}{ID: 67890},
			Text: "breakfast $20",
		},
	}
	body, _ := json.Marshal(update)

	req := httptest.NewRequest("POST", "/webhook/telegram", bytes.NewReader(body))
	w := httptest.NewRecorder()
	handler.HandleWebhook(w, req)

	// Verify response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	mockUC.AssertExpectations(t)
}
