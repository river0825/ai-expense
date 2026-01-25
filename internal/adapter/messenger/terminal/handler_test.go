package terminal

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/riverlin/aiexpense/internal/domain"
	"github.com/stretchr/testify/mock"
)

// Mock ProcessMessageUseCase
type mockProcessMessageUseCase struct {
	mock.Mock
}

func (m *mockProcessMessageUseCase) Execute(ctx context.Context, msg *domain.UserMessage) (*domain.MessageResponse, error) {
	args := m.Called(ctx, msg)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.MessageResponse), args.Error(1)
}

func TestTerminalHandler_HandleMessage_Success(t *testing.T) {
	// Setup
	mockUC := new(mockProcessMessageUseCase)
	handler := NewHandler(mockUC)

	// Expectations
	expectedResp := &domain.MessageResponse{
		Text: "Recorded 1 expense",
		Data: map[string]interface{}{"id": "123"},
	}
	mockUC.On("Execute", mock.Anything, mock.MatchedBy(func(msg *domain.UserMessage) bool {
		return msg.UserID == "test_user" && msg.Content == "breakfast $20" && msg.Source == "terminal"
	})).Return(expectedResp, nil)

	// Create request
	req := TerminalRequest{
		UserID:  "test_user",
		Message: "breakfast $20",
	}

	body, _ := json.Marshal(req)
	httpReq := httptest.NewRequest("POST", "/api/chat/terminal", bytes.NewReader(body))
	w := httptest.NewRecorder()

	// Execute
	handler.HandleMessage(w, httpReq)

	// Assert
	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp TerminalResponse
	json.NewDecoder(w.Body).Decode(&resp)

	if resp.Status != "success" {
		t.Errorf("expected status success, got %s", resp.Status)
	}

	if resp.Message != "Recorded 1 expense" {
		t.Errorf("expected message 'Recorded 1 expense', got %s", resp.Message)
	}
}

func TestTerminalHandler_HandleMessage_Error(t *testing.T) {
	// Setup
	mockUC := new(mockProcessMessageUseCase)
	handler := NewHandler(mockUC)

	// Expectations
	mockUC.On("Execute", mock.Anything, mock.Anything).Return(nil, fmt.Errorf("process error"))

	// Create request
	req := TerminalRequest{
		UserID:  "test_user",
		Message: "error",
	}

	body, _ := json.Marshal(req)
	httpReq := httptest.NewRequest("POST", "/api/chat/terminal", bytes.NewReader(body))
	w := httptest.NewRecorder()

	// Execute
	handler.HandleMessage(w, httpReq)

	// Assert
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", w.Code)
	}

	var resp TerminalResponse
	json.NewDecoder(w.Body).Decode(&resp)

	if resp.Status != "error" {
		t.Errorf("expected status error, got %s", resp.Status)
	}
}
