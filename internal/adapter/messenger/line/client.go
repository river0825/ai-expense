package line

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// Client represents the LINE Messaging API client
type Client struct {
	channelToken string
	apiURL       string
	profileURL   string
	httpClient   *http.Client
}

// NewClient creates a new LINE client
func NewClient(channelToken string) (*Client, error) {
	if channelToken == "" {
		return nil, fmt.Errorf("channel token is required")
	}

	return &Client{
		channelToken: channelToken,
		apiURL:       "https://api.line.me/v2/bot/message",
		profileURL:   "https://api.line.me/v2/bot/profile",
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

// FlexReplyRequest represents the request to send a Flex Message reply
type FlexReplyRequest struct {
	ReplyToken string        `json:"replyToken"`
	Messages   []FlexMessage `json:"messages"`
}

// FlexMessage represents a LINE Flex Message
type FlexMessage struct {
	Type     string      `json:"type"`
	AltText  string      `json:"altText"`
	Contents interface{} `json:"contents"`
}

// SendMessage sends a reply message to a user via LINE Messaging API
func (c *Client) SendMessage(ctx context.Context, replyToken, text string) error {
	// https://developers.line.biz/en/docs/messaging-api/message-types/#text-messages-v2
	// Ensure we're using the correct format. The current struct TextMessage matches standard text message format.
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
		lineLogger.ErrorContext(ctx, "failed to marshal LINE text reply payload", "error", err)
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
		lineLogger.ErrorContext(ctx, "failed to send LINE text reply request", "error", err, "reply_token", maskToken(replyToken))
		return fmt.Errorf("failed to send message: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	// Check HTTP status code
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		lineLogger.ErrorContext(
			ctx,
			"LINE API returned error for text reply",
			"status", resp.StatusCode,
			"body_preview", previewText(string(body), 400),
			"reply_token", maskToken(replyToken),
		)
		var apiResp LineAPIResponse
		if err := json.Unmarshal(body, &apiResp); err == nil && apiResp.Message != "" {
			return fmt.Errorf("line api error: %s (status: %d)", apiResp.Message, resp.StatusCode)
		}
		return fmt.Errorf("line api error: status %d, body: %s", resp.StatusCode, string(body))
	}

	lineLogger.InfoContext(ctx, "LINE text reply sent", "reply_token", maskToken(replyToken))
	return nil
}

// SendReply sends a reply message
func (c *Client) SendReply(ctx context.Context, replyToken, text string) error {
	return c.SendMessage(ctx, replyToken, text)
}

// SendFlexReply sends a Flex Message reply to a user via LINE Messaging API
func (c *Client) SendFlexReply(ctx context.Context, replyToken, altText string, contents interface{}) error {
	req := FlexReplyRequest{
		ReplyToken: replyToken,
		Messages: []FlexMessage{
			{
				Type:     "flex",
				AltText:  altText,
				Contents: contents,
			},
		},
	}

	payload, err := json.Marshal(req)
	if err != nil {
		lineLogger.ErrorContext(ctx, "failed to marshal LINE flex reply payload", "error", err)
		return fmt.Errorf("failed to marshal flex request: %w", err)
	}
	lineLogger.DebugContext(
		ctx,
		"sending LINE flex reply",
		"reply_token", maskToken(replyToken),
		"alt_text_len", len(altText),
		"contents_type", fmt.Sprintf("%T", contents),
		"payload_preview", previewText(string(payload), 400),
	)

	httpReq, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/reply", c.apiURL), bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.channelToken))

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		lineLogger.ErrorContext(ctx, "failed to send LINE flex reply request", "error", err, "reply_token", maskToken(replyToken))
		return fmt.Errorf("failed to send flex message: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		lineLogger.ErrorContext(
			ctx,
			"LINE API returned error for flex reply",
			"status", resp.StatusCode,
			"body_preview", previewText(string(body), 400),
			"reply_token", maskToken(replyToken),
		)
		var apiResp LineAPIResponse
		if err := json.Unmarshal(body, &apiResp); err == nil && apiResp.Message != "" {
			return fmt.Errorf("line api error: %s (status: %d)", apiResp.Message, resp.StatusCode)
		}
		return fmt.Errorf("line api error: status %d, body: %s", resp.StatusCode, string(body))
	}

	lineLogger.InfoContext(ctx, "LINE flex reply sent", "reply_token", maskToken(replyToken))
	return nil
}
