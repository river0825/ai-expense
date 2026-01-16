package line

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

// Client represents the LINE Messaging API client
type Client struct {
	channelToken string
	apiURL       string
	httpClient   *http.Client
}

// NewClient creates a new LINE client
func NewClient(channelToken string) (*Client, error) {
	if channelToken == "" {
		return nil, fmt.Errorf("channel token is required")
	}

	return &Client{
		channelToken: channelToken,
		apiURL:       "https://api.line.biz/v2/bot/message",
		httpClient:   &http.Client{},
	}, nil
}

// ReplyMessageRequest represents the request to send a reply message
type ReplyMessageRequest struct {
	ReplyToken string        `json:"replyToken"`
	Messages   []TextMessage `json:"messages"`
}

// TextMessage represents a text message
type TextMessage struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// LineAPIResponse represents the response from LINE API
type LineAPIResponse struct {
	Message string `json:"message,omitempty"`
}

// SendMessage sends a reply message to a user via LINE Messaging API
func (c *Client) SendMessage(ctx context.Context, replyToken, text string) error {
	req := ReplyMessageRequest{
		ReplyToken: replyToken,
		Messages: []TextMessage{
			{
				Type: "text",
				Text: text,
			},
		},
	}

	payload, err := json.Marshal(req)
	if err != nil {
		log.Printf("Error marshaling request: %v", err)
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/reply", c.apiURL), bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.channelToken))

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		log.Printf("Error sending message to LINE: %v", err)
		return fmt.Errorf("failed to send message: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	// Check HTTP status code
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		var apiResp LineAPIResponse
		if err := json.Unmarshal(body, &apiResp); err == nil && apiResp.Message != "" {
			return fmt.Errorf("line api error: %s (status: %d)", apiResp.Message, resp.StatusCode)
		}
		return fmt.Errorf("line api error: status %d", resp.StatusCode)
	}

	log.Printf("[LINE] Message sent to reply token %s", replyToken)
	return nil
}

// SendReply sends a reply message
func (c *Client) SendReply(ctx context.Context, replyToken, text string) error {
	return c.SendMessage(ctx, replyToken, text)
}
