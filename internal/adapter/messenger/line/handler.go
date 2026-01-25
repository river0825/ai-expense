package line

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/riverlin/aiexpense/internal/domain"
)

// MessageProcessor defines the interface for processing messages
type MessageProcessor interface {
	Execute(ctx context.Context, msg *domain.UserMessage) (*domain.MessageResponse, error)
}

// Handler handles LINE bot webhook events
type Handler struct {
	channelSecret string
	useCase       MessageProcessor
	client        *Client
}

// NewHandler creates a new LINE webhook handler
func NewHandler(channelSecret string, useCase MessageProcessor, client *Client) *Handler {
	return &Handler{
		channelSecret: channelSecret,
		useCase:       useCase,
		client:        client,
	}
}

// LineEvent represents a LINE messaging event
type LineEvent struct {
	Events []struct {
		Type    string `json:"type"`
		Message struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"message"`
		Source struct {
			Type   string `json:"type"`
			UserID string `json:"userId"`
		} `json:"source"`
		ReplyToken string `json:"replyToken"`
		Timestamp  int64  `json:"timestamp"`
	} `json:"events"`
}

// HandleWebhook processes incoming LINE webhook events
func (h *Handler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	// Verify signature
	signature := r.Header.Get("X-Line-Signature")
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read body", http.StatusBadRequest)
		return
	}

	// Debug log: Print webhook details
	log.Printf("[LINE Webhook] Signature: %s", signature)
	log.Printf("[LINE Webhook] Body: %s", string(body))

	if !h.verifySignature(signature, body) {
		log.Printf("[LINE Webhook] Invalid signature")
		http.Error(w, "Invalid signature", http.StatusUnauthorized)
		return
	}

	// Parse events
	var event LineEvent
	if err := json.Unmarshal(body, &event); err != nil {
		http.Error(w, "Failed to parse event", http.StatusBadRequest)
		return
	}

	ctx := context.Background()

	// Process each event
	for _, e := range event.Events {
		if e.Type != "message" || e.Message.Type != "text" {
			continue
		}

		log.Printf("[LINE Webhook] Processing message event from user %s: %s", e.Source.UserID, e.Message.Text)

		// Map to UserMessage
		userMsg := &domain.UserMessage{
			UserID:  e.Source.UserID,
			Content: e.Message.Text,
			Source:  "line",
			// Use event timestamp if available, otherwise Now
			Timestamp: time.Unix(e.Timestamp/1000, 0),
			Metadata: map[string]interface{}{
				"reply_token": e.ReplyToken,
			},
		}

		// Execute logic
		resp, err := h.useCase.Execute(ctx, userMsg)
		if err != nil {
			log.Printf("[LINE Webhook] Error handling message: %v", err)
			// Optionally send error message to user if appropriate
			continue
		}

		// Send reply
		if resp.Text != "" && h.client != nil {
			if err := h.client.SendReply(ctx, e.ReplyToken, resp.Text); err != nil {
				log.Printf("[LINE Webhook] Failed to send reply: %v", err)
			} else {
				log.Printf("[LINE Webhook] Reply sent successfully")
			}
		}
	}

	w.WriteHeader(http.StatusOK)
}

// verifySignature verifies the LINE webhook signature
func (h *Handler) verifySignature(signature string, body []byte) bool {
	hash := hmac.New(sha256.New, []byte(h.channelSecret))
	hash.Write(body)
	computed := base64.StdEncoding.EncodeToString(hash.Sum(nil))
	return strings.EqualFold(signature, computed)
}
