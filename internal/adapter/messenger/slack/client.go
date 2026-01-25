package slack

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// Client handles Slack API communication
type Client struct {
	botToken   string
	httpClient *http.Client
}

// NewClient creates a new Slack client
func NewClient(botToken string) (*Client, error) {
	if botToken == "" {
		return nil, fmt.Errorf("slack bot token is required")
	}

	return &Client{
		botToken:   botToken,
		httpClient: &http.Client{},
	}, nil
}

// SendMessage sends a message to a Slack user or channel
func (c *Client) SendMessage(userID, text string) error {
	if userID == "" || text == "" {
		return fmt.Errorf("user_id and text are required")
	}

	payload := map[string]interface{}{
		"channel": userID,
		"text":    text,
		"type":    "mrkdwn",
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequest("POST", "https://slack.com/api/chat.postMessage", bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.botToken))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("slack API returned status %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	// Check if the API call was successful
	if ok, exists := result["ok"].(bool); exists && !ok {
		if errMsg, hasErr := result["error"].(string); hasErr {
			return fmt.Errorf("slack API error: %s", errMsg)
		}
	}

	return nil
}

// PostMessage sends a message to a Slack channel (alias for SendMessage)
func (c *Client) PostMessage(ctx context.Context, channelID, text string) error {
	// For now ignoring context as SendMessage doesn't use it, but keeping signature correct for future
	return c.SendMessage(channelID, text)
}

// GetBotInfo retrieves information about the bot
func (c *Client) GetBotInfo() (map[string]interface{}, error) {
	req, err := http.NewRequest("GET", "https://slack.com/api/auth.test", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.botToken))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get bot info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("slack API returned status %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return result, nil
}

// OpenConversation opens a direct message with a user
func (c *Client) OpenConversation(userID string) (string, error) {
	if userID == "" {
		return "", fmt.Errorf("user_id is required")
	}

	payload := map[string]interface{}{
		"users": userID,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequest("POST", "https://slack.com/api/conversations.open", bytes.NewBuffer(body))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.botToken))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to open conversation: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("slack API returned status %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	// Check if the API call was successful
	if ok, exists := result["ok"].(bool); exists && !ok {
		if errMsg, hasErr := result["error"].(string); hasErr {
			return "", fmt.Errorf("slack API error: %s", errMsg)
		}
	}

	// Extract channel ID from response
	if channel, exists := result["channel"].(map[string]interface{}); exists {
		if channelID, hasID := channel["id"].(string); hasID {
			return channelID, nil
		}
	}

	return "", fmt.Errorf("failed to extract channel ID from response")
}
