package telegram

import (
	"context"
	"log"
)

// Client represents the Telegram Bot API client
type Client struct {
	botToken string
	// TODO: Add telegram-bot-api client when needed
	// client *tgbotapi.BotAPI
}

// NewClient creates a new Telegram client
func NewClient(botToken string) (*Client, error) {
	// TODO: Initialize Telegram Bot API client
	// client, err := tgbotapi.NewBotAPI(botToken)
	// if err != nil {
	//     return nil, err
	// }

	return &Client{
		botToken: botToken,
		// client: client,
	}, nil
}

// SendMessage sends a message to a chat
func (c *Client) SendMessage(ctx context.Context, chatID int64, text string) error {
	// TODO: Implement actual Telegram API call using SendMessage method
	// msg := tgbotapi.NewMessage(chatID, text)
	// _, err := c.client.Send(msg)
	// return err

	log.Printf("[Telegram Client] Sending message to chat %d: %s", chatID, text)
	return nil
}

// SendReply sends a reply message
func (c *Client) SendReply(ctx context.Context, chatID int64, text string) error {
	return c.SendMessage(ctx, chatID, text)
}
