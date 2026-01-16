package teams

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Client handles Microsoft Teams Bot API communication
type Client struct {
	appID       string
	appPassword string
	serviceURL  string
	httpClient  *http.Client
}

// NewClient creates a new Teams bot client
func NewClient(appID, appPassword string) (*Client, error) {
	if appID == "" || appPassword == "" {
		return nil, fmt.Errorf("teams app_id and app_password are required")
	}

	return &Client{
		appID:       appID,
		appPassword: appPassword,
		httpClient:  &http.Client{Timeout: 10 * time.Second},
	}, nil
}

// SetServiceURL sets the service URL for the current conversation
func (c *Client) SetServiceURL(serviceURL string) {
	c.serviceURL = serviceURL
}

// SendMessage sends a message to a Teams channel or user
func (c *Client) SendMessage(conversationID, text string) error {
	if conversationID == "" || text == "" {
		return fmt.Errorf("conversation_id and text are required")
	}

	if c.serviceURL == "" {
		return fmt.Errorf("service_url not set; must be called within activity context")
	}

	// Get access token
	token, err := c.getAccessToken()
	if err != nil {
		return fmt.Errorf("failed to get access token: %w", err)
	}

	payload := map[string]interface{}{
		"type": "message",
		"from": map[string]interface{}{
			"id":   c.appID,
			"name": "AIExpense Bot",
		},
		"conversation": map[string]interface{}{
			"id": conversationID,
		},
		"text": text,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	// Construct API endpoint
	url := fmt.Sprintf("%s/v3/conversations/%s/activities", c.serviceURL, conversationID)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("teams API returned status %d", resp.StatusCode)
	}

	return nil
}

// getAccessToken retrieves a Bearer token for authentication
func (c *Client) getAccessToken() (string, error) {
	// In a production system, this would use Azure AD OAuth flow
	// For now, we'll return a placeholder that would be replaced
	// with actual OAuth token retrieval

	// This is a simplified approach - in production, implement proper OAuth2
	// using github.com/Azure/azure-sdk-for-go or similar

	// For now, return the app password as a simple authentication
	// In real implementation, exchange with Azure AD token endpoint
	return "token_" + c.appID, nil
}

// GetBotInfo retrieves information about the bot
func (c *Client) GetBotInfo() map[string]interface{} {
	return map[string]interface{}{
		"app_id":   c.appID,
		"app_name": "AIExpense Bot",
		"platform": "Microsoft Teams",
	}
}

// UpdateActivity updates an existing message in Teams
func (c *Client) UpdateActivity(conversationID, activityID, text string) error {
	if conversationID == "" || activityID == "" || text == "" {
		return fmt.Errorf("conversation_id, activity_id, and text are required")
	}

	if c.serviceURL == "" {
		return fmt.Errorf("service_url not set")
	}

	token, err := c.getAccessToken()
	if err != nil {
		return fmt.Errorf("failed to get access token: %w", err)
	}

	payload := map[string]interface{}{
		"type": "message",
		"id":   activityID,
		"text": text,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	url := fmt.Sprintf("%s/v3/conversations/%s/activities/%s", c.serviceURL, conversationID, activityID)

	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to update activity: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("teams API returned status %d", resp.StatusCode)
	}

	return nil
}
