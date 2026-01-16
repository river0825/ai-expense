package line

import (
	"context"
	"log"
)

// Client represents the LINE Messaging API client
type Client struct {
	channelToken string
	// TODO: Add line-bot-sdk client when available
	// client *line.Client
}

// NewClient creates a new LINE client
func NewClient(channelToken string) (*Client, error) {
	// TODO: Initialize LINE bot SDK client
	// client, err := line.NewClient(channelToken)
	// if err != nil {
	//     return nil, err
	// }

	return &Client{
		channelToken: channelToken,
		// client: client,
	}, nil
}

// SendMessage sends a message to a user
func (c *Client) SendMessage(ctx context.Context, replyToken, text string) error {
	// TODO: Implement actual LINE API call
	// For now, just log it
	log.Printf("[LINE Client] Sending message to %s: %s", replyToken, text)
	return nil
}

// SendReply sends a reply message
func (c *Client) SendReply(ctx context.Context, replyToken, text string) error {
	return c.SendMessage(ctx, replyToken, text)
}
