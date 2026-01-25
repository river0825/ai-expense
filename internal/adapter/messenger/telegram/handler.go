package telegram

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/riverlin/aiexpense/internal/domain"
)

// MessageProcessor defines the interface for processing messages
type MessageProcessor interface {
	Execute(ctx context.Context, msg *domain.UserMessage) (*domain.MessageResponse, error)
}

// Handler handles Telegram bot webhook events
type Handler struct {
	botToken string
	useCase  MessageProcessor
	client   *Client
}

// NewHandler creates a new Telegram webhook handler
func NewHandler(botToken string, useCase MessageProcessor, client *Client) *Handler {
	return &Handler{
		botToken: botToken,
		useCase:  useCase,
		client:   client,
	}
}

// TelegramUpdate represents a Telegram incoming update (webhook event)
type TelegramUpdate struct {
	UpdateID int64 `json:"update_id"`
	Message  *struct {
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
	} `json:"message"`
}

// TelegramResponse represents a Telegram API response
type TelegramResponse struct {
	OK     bool        `json:"ok"`
	Result interface{} `json:"result,omitempty"`
	Error  string      `json:"description,omitempty"`
}

// HandleWebhook processes incoming Telegram webhook events
func (h *Handler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read body", http.StatusBadRequest)
		return
	}

	var update TelegramUpdate
	if err := json.Unmarshal(body, &update); err != nil {
		http.Error(w, "Failed to parse update", http.StatusBadRequest)
		return
	}

	// Process message if present
	if update.Message != nil && update.Message.Text != "" {
		if update.Message.From != nil && update.Message.Chat != nil {
			userID := fmt.Sprintf("telegram_%d", update.Message.From.ID)
			chatID := update.Message.Chat.ID

			// Map to UserMessage
			userMsg := &domain.UserMessage{
				UserID:    userID,
				Content:   update.Message.Text,
				Source:    "telegram",
				Timestamp: time.Unix(update.Message.Date, 0),
				Metadata: map[string]interface{}{
					"chat_id": chatID,
				},
			}

			// Execute logic
			resp, err := h.useCase.Execute(r.Context(), userMsg)
			if err != nil {
				log.Printf("Error handling message: %v", err)
				// Optionally send error to user
			} else {
				// Send reply
				if resp.Text != "" && h.client != nil {
					if err := h.client.SendMessage(r.Context(), chatID, resp.Text); err != nil {
						log.Printf("Error sending reply: %v", err)
					}
				}
			}
		}
	}

	// Always respond 200 OK to Telegram
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{"ok": true})
}

// VerifySecret verifies the Telegram webhook secret (optional)
// Telegram doesn't require signature verification like LINE does,
// but you can implement custom secret verification if needed
func (h *Handler) verifySecret(secret string) bool {
	// Simple comparison for now
	// In production, use constant-time comparison
	return secret == h.botToken
}
