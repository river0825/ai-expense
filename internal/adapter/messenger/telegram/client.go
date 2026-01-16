package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

// Client represents the Telegram Bot API client
type Client struct {
	botToken   string
	apiURL     string
	httpClient *http.Client
}

// NewClient creates a new Telegram client
func NewClient(botToken string) (*Client, error) {
	if botToken == "" {
		return nil, fmt.Errorf("bot token is required")
	}

	return &Client{
		botToken:   botToken,
		apiURL:     fmt.Sprintf("https://api.telegram.org/bot%s", botToken),
		httpClient: &http.Client{},
	}, nil
}

// SendMessageRequest represents the request to send a message
type SendMessageRequest struct {
	ChatID                int64  `json:"chat_id"`
	Text                  string `json:"text"`
	ParseMode             string `json:"parse_mode,omitempty"`
	DisableWebPagePreview bool   `json:"disable_web_page_preview,omitempty"`
}

// TelegramAPIResponse represents the response from Telegram API
type TelegramAPIResponse struct {
	OK        bool        `json:"ok"`
	Result    interface{} `json:"result,omitempty"`
	Error     string      `json:"description,omitempty"`
	ErrorCode int         `json:"error_code,omitempty"`
}

// SendMessage sends a message to a chat via Telegram Bot API
func (c *Client) SendMessage(ctx context.Context, chatID int64, text string) error {
	req := SendMessageRequest{
		ChatID:    chatID,
		Text:      text,
		ParseMode: "HTML",
	}

	payload, err := json.Marshal(req)
	if err != nil {
		log.Printf("Error marshaling request: %v", err)
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/sendMessage", c.apiURL), bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		log.Printf("Error sending message to Telegram: %v", err)
		return fmt.Errorf("failed to send message: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	var apiResp TelegramAPIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	if !apiResp.OK {
		return fmt.Errorf("telegram api error: %s (code: %d)", apiResp.Error, apiResp.ErrorCode)
	}

	log.Printf("[Telegram] Message sent to chat %d", chatID)
	return nil
}

// SendReply sends a reply message
func (c *Client) SendReply(ctx context.Context, chatID int64, text string) error {
	return c.SendMessage(ctx, chatID, text)
}

// GetMe retrieves bot information
func (c *Client) GetMe(ctx context.Context) error {
	httpReq, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/getMe", c.apiURL), nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to get bot info: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	var apiResp TelegramAPIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	if !apiResp.OK {
		return fmt.Errorf("telegram api error: %s", apiResp.Error)
	}

	log.Printf("[Telegram] Bot connected and ready")
	return nil
}
