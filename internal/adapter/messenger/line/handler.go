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

	"github.com/riverlin/aiexpense/internal/usecase"
)

// Handler handles LINE bot webhook events
type Handler struct {
	channelSecret string
	useCase       *LineUseCase
}

// NewHandler creates a new LINE webhook handler
func NewHandler(channelSecret string, useCase *LineUseCase) *Handler {
	return &Handler{
		channelSecret: channelSecret,
		useCase:       useCase,
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

	if !h.verifySignature(signature, body) {
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

		// Handle the message
		if err := h.useCase.HandleMessage(ctx, e.Source.UserID, e.Message.Text, e.ReplyToken); err != nil {
			log.Printf("Error handling message: %v", err)
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
