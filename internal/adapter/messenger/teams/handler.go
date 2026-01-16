package teams

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

// Handler handles Microsoft Teams webhook events
type Handler struct {
	appID       string
	appPassword string
	useCase     *UseCase
}

// NewHandler creates a new Teams webhook handler
func NewHandler(appID, appPassword string, useCase *UseCase) *Handler {
	return &Handler{
		appID:       appID,
		appPassword: appPassword,
		useCase:     useCase,
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
	if h.useCase.client != nil {
		h.useCase.client.SetServiceURL(activity.ServiceURL)
	}

	// Handle different activity types
	switch activity.Type {
	case "message":
		// Process text messages
		if activity.Text != "" && activity.From.ID != "" {
			go func() {
				ctx := r.Context()
				isMention := h.containsBotMention(activity)
				if isMention {
					if err := h.useCase.ProcessMention(ctx, activity.From.ID, activity.Text); err != nil {
						log.Printf("Teams: mention processing failed: %v", err)
					}
				} else if activity.ChannelID == "" || activity.Conversation.ConversationType == "personal" {
					// Direct message only
					if err := h.useCase.ProcessMessage(ctx, activity.From.ID, activity.Text); err != nil {
						log.Printf("Teams: message processing failed: %v", err)
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

// containsBotMention checks if the activity contains a mention of the bot
func (h *Handler) containsBotMention(activity Activity) bool {
	if activity.Text == "" {
		return false
	}

	// Check for <at>BotName</at> pattern
	return strings.Contains(activity.Text, "<at>") && strings.Contains(activity.Text, "</at>")
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

// SendMessage sends a message to a Teams conversation
func (h *Handler) SendMessage(conversationID, text string) error {
	if h.useCase.client == nil {
		return fmt.Errorf("teams client not configured")
	}
	return h.useCase.client.SendMessage(conversationID, text)
}
