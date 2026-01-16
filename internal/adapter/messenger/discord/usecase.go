package discord

import (
	"context"
	"log"

	"github.com/riverlin/aiexpense/internal/usecase"
)

// DiscordUseCase handles Discord-specific business logic
type DiscordUseCase struct {
	autoSignupUC        *usecase.AutoSignupUseCase
	parseConversationUC *usecase.ParseConversationUseCase
	createExpenseUC     *usecase.CreateExpenseUseCase
	client              *Client
}

// NewDiscordUseCase creates a new Discord use case
func NewDiscordUseCase(
	autoSignupUC *usecase.AutoSignupUseCase,
	parseConversationUC *usecase.ParseConversationUseCase,
	createExpenseUC *usecase.CreateExpenseUseCase,
	client *Client,
) *DiscordUseCase {
	return &DiscordUseCase{
		autoSignupUC:        autoSignupUC,
		parseConversationUC: parseConversationUC,
		createExpenseUC:     createExpenseUC,
		client:              client,
	}
}

// HandleMessage handles an incoming Discord message
func (u *DiscordUseCase) HandleMessage(ctx context.Context, userID, text, token, interactionID string) error {
	// Auto-signup user
	if err := u.autoSignupUC.Execute(ctx, userID, "discord"); err != nil {
		log.Printf("Error auto-signing up user: %v", err)
	}

	// Parse conversation to extract expenses
	parsedExpenses, err := u.parseConversationUC.Execute(ctx, text, userID)
	if err != nil {
		log.Printf("Error parsing conversation: %v", err)
		return u.sendMessage(ctx, token, interactionID, "Processing failed, please try again later")
	}

	if len(parsedExpenses) == 0 {
		return u.sendMessage(ctx, token, interactionID, "No valid expense items found. Please provide an amount and item, for example: breakfast $20")
	}

	// Create expenses
	var messages []string
	for _, parsed := range parsedExpenses {
		createReq := &usecase.CreateRequest{
			UserID:      userID,
			Description: parsed.Description,
			Amount:      parsed.Amount,
			Date:        parsed.Date,
		}

		resp, err := u.createExpenseUC.Execute(ctx, createReq)
		if err != nil {
			log.Printf("Error creating expense: %v", err)
			messages = append(messages, parsed.Description+" (save failed)")
			continue
		}

		messages = append(messages, resp.Message)
	}

	// Send consolidated response
	consolidatedMessage := ""
	for i, msg := range messages {
		if i > 0 {
			consolidatedMessage += "\n"
		}
		consolidatedMessage += msg
	}

	return u.sendMessage(ctx, token, interactionID, consolidatedMessage)
}

// sendMessage sends a message via Discord
func (u *DiscordUseCase) sendMessage(ctx context.Context, token, interactionID, text string) error {
	if u.client == nil {
		// No client configured, just log
		log.Printf("[Discord] Reply to interaction %s: %s (client not configured)", interactionID, text)
		return nil
	}

	// Send actual message via Discord API
	return u.client.SendMessage(ctx, token, interactionID, text)
}
