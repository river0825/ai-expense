package discord

import (
	"context"
	"encoding/json"
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

// Handler handles Discord webhook events
type Handler struct {
	botToken string
	useCase  MessageProcessor
	client   *Client
}

// NewHandler creates a new Discord webhook handler
func NewHandler(botToken string, useCase MessageProcessor, client *Client) *Handler {
	return &Handler{
		botToken: botToken,
		useCase:  useCase,
		client:   client,
	}
}

// DiscordInteraction represents an interaction from Discord
type DiscordInteraction struct {
	Type      int             `json:"type"`
	ID        string          `json:"id"`
	Token     string          `json:"token"`
	Data      InteractionData `json:"data,omitempty"`
	Message   DiscordMessage  `json:"message,omitempty"`
	Member    Member          `json:"member,omitempty"`
	User      User            `json:"user,omitempty"`
	ChannelID string          `json:"channel_id"`
	GuildID   string          `json:"guild_id,omitempty"`
}

// InteractionData represents the data payload in an interaction
type InteractionData struct {
	Content string `json:"content"`
}

// DiscordMessage represents a Discord message
type DiscordMessage struct {
	ID      string `json:"id"`
	Content string `json:"content"`
	Author  Author `json:"author"`
}

// Author represents the author of a Discord message
type Author struct {
	ID       string `json:"id"`
	Username string `json:"username"`
}

// Member represents a guild member
type Member struct {
	User User `json:"user"`
}

// User represents a Discord user
type User struct {
	ID       string `json:"id"`
	Username string `json:"username"`
}

// InteractionType values
const (
	InteractionTypePing = iota + 1
	InteractionTypeApplicationCommand
	InteractionTypeMessageComponent
	InteractionTypeApplicationCommandAutocomplete
	InteractionTypeModalSubmit
)

// HandleWebhook handles incoming Discord interactions
func (h *Handler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading request body: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var interaction DiscordInteraction
	if err := json.Unmarshal(body, &interaction); err != nil {
		log.Printf("Error unmarshaling interaction: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Handle ping (required by Discord for webhook validation)
	if interaction.Type == InteractionTypePing {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]int{"type": 1}) // PONG
		return
	}

	// Extract user info
	userID := interaction.User.ID
	if userID == "" && interaction.Member.User.ID != "" {
		userID = interaction.Member.User.ID
	}

	// Extract message content
	var messageText string
	if interaction.Data.Content != "" {
		messageText = interaction.Data.Content
	}

	if messageText == "" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"type": 4,
			"data": map[string]string{
				"content": "No message content provided",
			},
		})
		return
	}

	// Map to UserMessage
	userMsg := &domain.UserMessage{
		UserID:    userID,
		Content:   messageText,
		Source:    "discord",
		Timestamp: time.Now(), // Interaction doesn't provide easy timestamp, using Now
		Metadata: map[string]interface{}{
			"token":          interaction.Token,
			"interaction_id": interaction.ID,
		},
	}

	// Process message
	resp, err := h.useCase.Execute(r.Context(), userMsg)
	if err != nil {
		log.Printf("Error processing message: %v", err)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"type": 4,
			"data": map[string]string{
				"content": "Failed to process message",
			},
		})
		return
	}

	// Send reply
	// Discord allows initial response via HTTP response (type 4)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"type": 4,
		"data": map[string]string{
			"content": resp.Text,
		},
	})
}
