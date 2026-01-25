package slack

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
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

// Handler handles Slack webhook events
type Handler struct {
	signingSecret string
	useCase       MessageProcessor
	client        *Client
}

// NewHandler creates a new Slack webhook handler
func NewHandler(signingSecret string, useCase MessageProcessor, client *Client) *Handler {
	return &Handler{
		signingSecret: signingSecret,
		useCase:       useCase,
		client:        client,
	}
}

// SlackEvent represents a Slack event
type SlackEvent struct {
	Token     string `json:"token"`
	TeamID    string `json:"team_id"`
	ApiAppID  string `json:"api_app_id"`
	Event     *Event `json:"event"`
	Type      string `json:"type"`
	EventID   string `json:"event_id"`
	EventTime int64  `json:"event_time"`
	Challenge string `json:"challenge"`
}

// Event represents the event payload
type Event struct {
	Type            string `json:"type"`
	User            string `json:"user"`
	Text            string `json:"text"`
	Channel         string `json:"channel"`
	Timestamp       string `json:"ts"`
	BotID           string `json:"bot_id"`
	ThreadTimestamp string `json:"thread_ts"`
}

// HandleWebhook handles incoming Slack webhook requests
func (h *Handler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	// Read the request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Slack: failed to read request body: %v", err)
		http.Error(w, "failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Verify request signature
	if h.signingSecret != "" {
		if !h.verifySignature(r, body) {
			log.Printf("Slack: signature verification failed")
			http.Error(w, "signature verification failed", http.StatusUnauthorized)
			return
		}
	}

	// Parse the event
	var slackEvent SlackEvent
	if err := json.Unmarshal(body, &slackEvent); err != nil {
		log.Printf("Slack: failed to parse event: %v", err)
		http.Error(w, "failed to parse event", http.StatusBadRequest)
		return
	}

	// Handle URL verification challenge
	if slackEvent.Type == "url_verification" {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(slackEvent.Challenge))
		return
	}

	// Ignore bot messages and other non-user messages
	if slackEvent.Event == nil || slackEvent.Event.BotID != "" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Handle different event types
	if (slackEvent.Event.Type == "message" || slackEvent.Event.Type == "app_mention") && slackEvent.Event.Text != "" && slackEvent.Event.User != "" {
		// Map to UserMessage
		userMsg := &domain.UserMessage{
			UserID:    slackEvent.Event.User,
			Content:   slackEvent.Event.Text,
			Source:    "slack",
			Timestamp: time.Now(), // Slack timestamp is a string, using Now for simplicity or parse if needed
			Metadata: map[string]interface{}{
				"channel":   slackEvent.Event.Channel,
				"thread_ts": slackEvent.Event.ThreadTimestamp,
			},
		}

		// Handle asynchronously as Slack requires quick response
		go func(msg *domain.UserMessage, channelID string) {
			ctx := context.Background() // Create new context for async
			resp, err := h.useCase.Execute(ctx, msg)
			if err != nil {
				log.Printf("Slack: message processing failed: %v", err)
				// Optionally send error message
			} else {
				// Send reply
				if resp.Text != "" && h.client != nil {
					if err := h.client.PostMessage(ctx, channelID, resp.Text); err != nil {
						log.Printf("Slack: failed to send reply: %v", err)
					}
				}
			}
		}(userMsg, slackEvent.Event.Channel)
	}

	// Always respond with 200 OK to acknowledge receipt
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"ok": "true"})
}

// verifySignature verifies the Slack request signature
func (h *Handler) verifySignature(r *http.Request, body []byte) bool {
	// Get signature from headers
	signature := r.Header.Get("X-Slack-Request-Signature")
	timestamp := r.Header.Get("X-Slack-Request-Timestamp")

	if signature == "" || timestamp == "" {
		return false
	}

	// Check timestamp is recent (within 5 minutes)
	ts := time.Now().Unix()
	var requestTS int64
	fmt.Sscanf(timestamp, "%d", &requestTS)

	if ts-requestTS > 300 {
		// Request is too old
		return false
	}

	// Build the basestring
	basestring := fmt.Sprintf("v0:%s:%s", timestamp, string(body))

	// Create HMAC
	mac := hmac.New(sha256.New, []byte(h.signingSecret))
	mac.Write([]byte(basestring))
	expectedSignature := "v0=" + hex.EncodeToString(mac.Sum(nil))

	// Compare signatures
	return hmac.Equal([]byte(signature), []byte(expectedSignature))
}
