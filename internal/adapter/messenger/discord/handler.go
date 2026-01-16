package discord

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
)

// Handler handles Discord webhook events
type Handler struct {
	botToken string
	useCase  *DiscordUseCase
}

// NewHandler creates a new Discord webhook handler
func NewHandler(botToken string, useCase *DiscordUseCase) *Handler {
	return &Handler{
		botToken: botToken,
		useCase:  useCase,
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

	// Handle the message
	if err := h.useCase.HandleMessage(r.Context(), userID, messageText, interaction.Token, interaction.ID); err != nil {
		log.Printf("Error handling message: %v", err)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"type": 4,
			"data": map[string]string{
				"content": "Failed to process message",
			},
		})
		return
	}

	// Return success response (message handled asynchronously)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"type": 5, // Defer the response, will send later
	})
}
