package discord

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

// Client represents the Discord Bot API client
type Client struct {
	botToken   string
	apiURL     string
	httpClient *http.Client
}

// NewClient creates a new Discord client
func NewClient(botToken string) (*Client, error) {
	if botToken == "" {
		return nil, fmt.Errorf("bot token is required")
	}

	return &Client{
		botToken:   botToken,
		apiURL:     "https://discord.com/api/v10",
		httpClient: &http.Client{},
	}, nil
}

// InteractionResponse represents a response to a Discord interaction
type InteractionResponse struct {
	Type int                     `json:"type"`
	Data InteractionCallbackData `json:"data,omitempty"`
}

// InteractionCallbackData represents the data for an interaction response
type InteractionCallbackData struct {
	Content string `json:"content"`
	TTS     bool   `json:"tts,omitempty"`
}

// FollowupMessage represents a followup message to an interaction
type FollowupMessage struct {
	Content string `json:"content"`
	TTS     bool   `json:"tts,omitempty"`
}

// DiscordAPIError represents an error from Discord API
type DiscordAPIError struct {
	Code    int    `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
}

// SendMessage sends a message as a response to a Discord interaction
// Uses followup message API for deferred responses
func (c *Client) SendMessage(ctx context.Context, token, interactionID, text string) error {
	// First, acknowledge the interaction with a deferred response
	ackURL := fmt.Sprintf("%s/interactions/%s/%s/callback", c.apiURL, interactionID, token)
	ackReq := InteractionResponse{
		Type: 5, // DEFERRED_CHANNEL_MESSAGE_WITH_SOURCE
	}

	payload, err := json.Marshal(ackReq)
	if err != nil {
		log.Printf("Error marshaling ack request: %v", err)
		return fmt.Errorf("failed to marshal ack request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", ackURL, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("failed to create ack request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", fmt.Sprintf("Bot %s", c.botToken))

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		log.Printf("Error sending ack to Discord: %v", err)
		return fmt.Errorf("failed to send ack: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("Ack error response: %s", string(body))
		return fmt.Errorf("discord api error on ack: status %d", resp.StatusCode)
	}

	// Send the actual message as a followup
	followupURL := fmt.Sprintf("%s/webhooks/%s/%s", c.apiURL, interactionID, token)
	followupMsg := FollowupMessage{
		Content: text,
	}

	payload, err = json.Marshal(followupMsg)
	if err != nil {
		log.Printf("Error marshaling followup request: %v", err)
		return fmt.Errorf("failed to marshal followup request: %w", err)
	}

	httpReq, err = http.NewRequestWithContext(ctx, "POST", followupURL, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("failed to create followup request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", fmt.Sprintf("Bot %s", c.botToken))

	resp, err = c.httpClient.Do(httpReq)
	if err != nil {
		log.Printf("Error sending followup to Discord: %v", err)
		return fmt.Errorf("failed to send followup: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		var apiErr DiscordAPIError
		if err := json.Unmarshal(body, &apiErr); err == nil && apiErr.Message != "" {
			return fmt.Errorf("discord api error: %s (code: %d)", apiErr.Message, apiErr.Code)
		}
		return fmt.Errorf("discord api error: status %d", resp.StatusCode)
	}

	log.Printf("[Discord] Message sent for interaction %s", interactionID)
	return nil
}

// GetBotInfo retrieves bot information
func (c *Client) GetBotInfo(ctx context.Context) error {
	url := fmt.Sprintf("%s/users/@me", c.apiURL)

	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Authorization", fmt.Sprintf("Bot %s", c.botToken))

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to get bot info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("discord api error: status %d - %s", resp.StatusCode, string(body))
	}

	log.Printf("[Discord] Bot connected and ready")
	return nil
}
