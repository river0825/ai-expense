package teams

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

// Handler handles Microsoft Teams webhook events
type Handler struct {
	appID       string
	appPassword string
	useCase     MessageProcessor
	client      *Client
}

// NewHandler creates a new Teams webhook handler
func NewHandler(appID, appPassword string, useCase MessageProcessor, client *Client) *Handler {
	return &Handler{
		appID:       appID,
		appPassword: appPassword,
		useCase:     useCase,
		client:      client,
	}
}

// Activity represents a Teams activity/event
type Activity struct {
	Type           string       `json:"type"`
	ID             string       `json:"id"`
	Timestamp      string       `json:"timestamp"`
	LocalTimestamp string       `json:"localTimestamp"`
	ServiceURL     string       `json:"serviceUrl"`
	ChannelID      string       `json:"channelId"`
	ChannelData    ChannelData  `json:"channelData"`
	From           User         `json:"from"`
	Conversation   Conversation `json:"conversation"`
	Recipient      User         `json:"recipient"`
	Text           string       `json:"text"`
	ReplyToID      string       `json:"replyToId"`
	Mentions       []Mention    `json:"entities"`
}

// ChannelData contains Teams-specific channel data
type ChannelData struct {
	EventType   string `json:"eventType"`
	TeamID      string `json:"teamsTeamId"`
	ChannelID   string `json:"teamsChannelId"`
	ChannelName string `json:"teamsChannelName"`
}

// User represents a Teams user
type User struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Conversation represents Teams conversation context
type Conversation struct {
	ConversationType string `json:"conversationType"`
	ID               string `json:"id"`
	IsGroup          bool   `json:"isGroup"`
}

// Mention represents a mention entity
type Mention struct {
	Type      string `json:"type"`
	Mentioned User   `json:"mentioned"`
	Text      string `json:"text"`
}

// HandleWebhook handles incoming Teams webhook requests
func (h *Handler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	// Read the request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Teams: failed to read request body: %v", err)
		http.Error(w, "failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Verify request signature
	if !h.verifySignature(r, body) {
		log.Printf("Teams: signature verification failed")
		http.Error(w, "signature verification failed", http.StatusUnauthorized)
		return
	}

	// Parse the activity
	var activity Activity
	if err := json.Unmarshal(body, &activity); err != nil {
		log.Printf("Teams: failed to parse activity: %v", err)
		http.Error(w, "failed to parse activity", http.StatusBadRequest)
		return
	}

	// Set service URL for reply
	if h.client != nil {
		h.client.SetServiceURL(activity.ServiceURL)
	}

	// Handle different activity types
	switch activity.Type {
	case "message":
		// Process text messages
		if activity.Text != "" && activity.From.ID != "" {
			// Map to UserMessage
			userMsg := &domain.UserMessage{
				UserID:    activity.From.ID,
				Content:   activity.Text,
				Source:    "teams",
				Timestamp: time.Now(), // Should parse activity.Timestamp if precise time needed
				Metadata: map[string]interface{}{
					"conversation_id": activity.Conversation.ID,
					"service_url":     activity.ServiceURL,
				},
			}

			go func() {
				ctx := context.Background()
				resp, err := h.useCase.Execute(ctx, userMsg)
				if err != nil {
					log.Printf("Teams: processing failed: %v", err)
				} else {
					// Send reply
					if resp.Text != "" && h.client != nil {
						if err := h.client.SendMessage(activity.Conversation.ID, resp.Text); err != nil {
							log.Printf("Teams: failed to send reply: %v", err)
						}
					}
				}
			}()
		}

	case "conversationUpdate":
		// Handle bot added to conversation
		log.Printf("Teams: bot added to conversation: %s", activity.Conversation.ID)

	case "event":
		// Handle other events
		log.Printf("Teams: event received: %v", activity)
	}

	// Always respond with 200 OK
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"ok": "true"})
}

// verifySignature verifies the Teams request signature
func (h *Handler) verifySignature(r *http.Request, body []byte) bool {
	// Get signature from header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return false
	}

	// Extract the signature from "Bearer <signature>"
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return false
	}
	signature := parts[1]

	// Compute HMAC
	mac := hmac.New(sha256.New, []byte(h.appPassword))
	mac.Write(body)
	expectedSignature := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	// Compare signatures
	return hmac.Equal([]byte(signature), []byte(expectedSignature))
}
