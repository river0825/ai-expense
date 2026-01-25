package whatsapp

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
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

func TestWhatsAppHandler_HandleWebhook_Success(t *testing.T) {
	// Setup
	mockUC := new(MockMessageProcessor)
	handler := NewHandler("test_app_secret", "1234567890", mockUC)

	// Expectations
	mockUC.On("Execute", mock.Anything, mock.MatchedBy(func(msg *domain.UserMessage) bool {
		return msg.UserID == "1234567890" && msg.Content == "breakfast $20" && msg.Source == "whatsapp"
	})).Return(&domain.MessageResponse{Text: "Saved"}, nil)

	// Create valid webhook payload
	payload, signature := createWhatsAppWebhookPayload("1234567890", "breakfast $20")

	req := httptest.NewRequest("POST", "/webhook/whatsapp", bytes.NewReader(payload))
	req.Header.Set("X-Hub-Signature-256", signature)
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

// Helper to create valid WhatsApp webhook payload
func createWhatsAppWebhookPayload(from, text string) ([]byte, string) {
	// Create payload structure matching WhatsApp schema
	payload := WebhookPayload{
		Object: "whatsapp_business_account",
		Entry: []WebhookEntry{
			{
				ID: "entry_123",
				Changes: []WebhookChange{
					{
						Field: "messages",
						Value: WebhookChangeValue{
							MessagingProduct: "whatsapp",
							Metadata: MessageMetadata{
								DisplayPhoneNumber: "1234567890",
								PhoneNumberID:      "phone_id_123",
							},
							Messages: []IncomingMessage{
								{
									From:      from,
									ID:        "msg_123",
									Timestamp: "1625097600",
									Type:      "text",
									Text: TextContent{
										Body: text,
									},
								},
							},
						},
					},
				},
			},
		},
	}

	body, _ := json.Marshal(payload)

	// Compute signature
	hash := hmac.New(sha256.New, []byte("test_app_secret"))
	hash.Write(body)
	signature := "sha256=" + hex.EncodeToString(hash.Sum(nil))

	return body, signature
}
