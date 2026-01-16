package whatsapp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

// Client represents the WhatsApp Business API client
type Client struct {
	phoneNumberID string
	accessToken   string
	apiURL        string
	httpClient    *http.Client
}

// NewClient creates a new WhatsApp client
func NewClient(phoneNumberID, accessToken string) (*Client, error) {
	if phoneNumberID == "" || accessToken == "" {
		return nil, fmt.Errorf("phone number ID and access token are required")
	}

	return &Client{
		phoneNumberID: phoneNumberID,
		accessToken:   accessToken,
		apiURL:        "https://graph.instagram.com/v18.0",
		httpClient:    &http.Client{},
	}, nil
}

// SendMessageRequest represents a request to send a message
type SendMessageRequest struct {
	MessagingProduct string      `json:"messaging_product"`
	To               string      `json:"to"`
	Type             string      `json:"type"`
	Text             TextMessage `json:"text,omitempty"`
}

// TextMessage represents a text message
type TextMessage struct {
	PreviewURL bool   `json:"preview_url,omitempty"`
	Body       string `json:"body"`
}

// WhatsAppAPIResponse represents a response from WhatsApp API
type WhatsAppAPIResponse struct {
	Messages []struct {
		ID string `json:"id"`
	} `json:"messages,omitempty"`
	Error struct {
		Message string `json:"message,omitempty"`
		Type    string `json:"type,omitempty"`
		Code    int    `json:"code,omitempty"`
	} `json:"error,omitempty"`
}

// SendMessage sends a message via WhatsApp Business API
func (c *Client) SendMessage(ctx context.Context, phoneNumber, text string) error {
	// Ensure phone number format (without +)
	if len(phoneNumber) > 0 && phoneNumber[0] == '+' {
		phoneNumber = phoneNumber[1:]
	}

	req := SendMessageRequest{
		MessagingProduct: "whatsapp",
		To:               phoneNumber,
		Type:             "text",
		Text: TextMessage{
			PreviewURL: false,
			Body:       text,
		},
	}

	payload, err := json.Marshal(req)
	if err != nil {
		log.Printf("Error marshaling request: %v", err)
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/%s/messages", c.apiURL, c.phoneNumberID)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.accessToken))

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		log.Printf("Error sending message to WhatsApp: %v", err)
		return fmt.Errorf("failed to send message: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	var apiResp WhatsAppAPIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("whatsapp api error: %s (code: %d)", apiResp.Error.Message, apiResp.Error.Code)
	}

	if len(apiResp.Messages) > 0 {
		log.Printf("[WhatsApp] Message sent to %s (ID: %s)", phoneNumber, apiResp.Messages[0].ID)
	}

	return nil
}

// UploadMedia uploads media to WhatsApp
func (c *Client) UploadMedia(ctx context.Context, mediaURL, mediaType string) (string, error) {
	// This is a placeholder for media upload functionality
	// Implementation would depend on specific use case
	return "", fmt.Errorf("media upload not yet implemented")
}

// GetPhoneInfo retrieves phone number information
func (c *Client) GetPhoneInfo(ctx context.Context) error {
	url := fmt.Sprintf("%s/%s", c.apiURL, c.phoneNumberID)

	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.accessToken))

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to get phone info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("whatsapp api error: status %d - %s", resp.StatusCode, string(body))
	}

	log.Printf("[WhatsApp] Phone number verified and connected")
	return nil
}
