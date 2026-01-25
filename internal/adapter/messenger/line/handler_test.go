package line

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
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

// MockClient is tough because Client struct is concrete.
// But Handler accepts *Client. We can pass nil if we don't test replying (or test nil client).
// Or we can construct a Client with a mock HTTP client?
// Given `NewClient` returns a `*Client` with unexported fields, modifying it to be mockable is hard without refactoring `Client` too.
// For now, let's use a real Client with a mock HTTP server if we need to test replying,
// OR just verify `Execute` is called and don't verify reply if we pass nil client.
// Or we can modify Handler to accept an interface for Client too?
// Let's modify Handler to accept an interface `ReplySender` instead of `*Client`.

// Helper to create valid LINE webhook payload
func createLineWebhookPayload(userID, text string) ([]byte, string) {
	payload := map[string]interface{}{
		"events": []map[string]interface{}{
			{
				"type": "message",
				"source": map[string]string{
					"type":   "user",
					"userId": userID,
				},
				"message": map[string]string{
					"type": "text",
					"text": text,
				},
				"replyToken": "test_reply_token_123",
				"timestamp":  1625097600000,
			},
		},
	}
	body, _ := json.Marshal(payload)

	// Compute LINE signature
	hash := hmac.New(sha256.New, []byte("test_channel_secret"))
	hash.Write(body)
	signature := base64.StdEncoding.EncodeToString(hash.Sum(nil))

	return body, signature
}

func TestLineHandler_HandleWebhook_Success(t *testing.T) {
	// Setup
	mockUC := new(MockMessageProcessor)
	handler := NewHandler("test_channel_secret", mockUC, nil) // Client nil for now

	// Expectations
	mockUC.On("Execute", mock.Anything, mock.MatchedBy(func(msg *domain.UserMessage) bool {
		return msg.UserID == "line_test_user" && msg.Content == "breakfast $20" && msg.Source == "line"
	})).Return(&domain.MessageResponse{Text: "Saved"}, nil)

	// Create valid webhook
	payload, signature := createLineWebhookPayload("line_test_user", "breakfast $20")

	req := httptest.NewRequest("POST", "/webhook/line", bytes.NewReader(payload))
	req.Header.Set("X-Line-Signature", signature)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	handler.HandleWebhook(w, req)

	// Verify response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	mockUC.AssertExpectations(t)
}

func TestLineHandler_HandleWebhook_InvalidSignature(t *testing.T) {
	// Setup
	mockUC := new(MockMessageProcessor)
	handler := NewHandler("test_channel_secret", mockUC, nil)

	// Create webhook with invalid signature
	payload, _ := createLineWebhookPayload("line_test_user", "breakfast $20")

	req := httptest.NewRequest("POST", "/webhook/line", bytes.NewReader(payload))
	req.Header.Set("X-Line-Signature", "invalid_signature_value")
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	handler.HandleWebhook(w, req)

	// Verify rejection
	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}

	mockUC.AssertNotCalled(t, "Execute")
}

func TestLineHandler_HandleWebhook_ExecuteError(t *testing.T) {
	// Setup
	mockUC := new(MockMessageProcessor)
	handler := NewHandler("test_channel_secret", mockUC, nil)

	// Expectations
	mockUC.On("Execute", mock.Anything, mock.Anything).Return(nil, fmt.Errorf("processing failed"))

	// Create valid webhook
	payload, signature := createLineWebhookPayload("line_test_user", "error message")

	req := httptest.NewRequest("POST", "/webhook/line", bytes.NewReader(payload))
	req.Header.Set("X-Line-Signature", signature)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	handler.HandleWebhook(w, req)

	// Verify response (should still be 200 OK to LINE, but logged error)
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	mockUC.AssertExpectations(t)
}
