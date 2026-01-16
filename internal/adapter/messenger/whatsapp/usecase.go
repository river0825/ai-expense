package whatsapp

import (
	"context"
	"log"

	"github.com/riverlin/aiexpense/internal/usecase"
)

// WhatsAppUseCase handles WhatsApp-specific business logic
type WhatsAppUseCase struct {
	autoSignupUC        *usecase.AutoSignupUseCase
	parseConversationUC *usecase.ParseConversationUseCase
	createExpenseUC     *usecase.CreateExpenseUseCase
	client              *Client
}

// NewWhatsAppUseCase creates a new WhatsApp use case
func NewWhatsAppUseCase(
	autoSignupUC *usecase.AutoSignupUseCase,
	parseConversationUC *usecase.ParseConversationUseCase,
	createExpenseUC *usecase.CreateExpenseUseCase,
	client *Client,
) *WhatsAppUseCase {
	return &WhatsAppUseCase{
		autoSignupUC:        autoSignupUC,
		parseConversationUC: parseConversationUC,
		createExpenseUC:     createExpenseUC,
		client:              client,
	}
}

// HandleMessage handles an incoming WhatsApp message
func (u *WhatsAppUseCase) HandleMessage(ctx context.Context, userID, text string) error {
	// Auto-signup user
	if err := u.autoSignupUC.Execute(ctx, userID, "whatsapp"); err != nil {
		log.Printf("Error auto-signing up user: %v", err)
	}

	// Parse conversation to extract expenses
	parsedExpenses, err := u.parseConversationUC.Execute(ctx, text, userID)
	if err != nil {
		log.Printf("Error parsing conversation: %v", err)
		return u.sendMessage(ctx, userID, "Processing failed, please try again later")
	}

	if len(parsedExpenses) == 0 {
		return u.sendMessage(ctx, userID, "No valid expense items found. Please provide an amount and item, for example: breakfast $20")
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

	return u.sendMessage(ctx, userID, consolidatedMessage)
}

// sendMessage sends a message via WhatsApp
func (u *WhatsAppUseCase) sendMessage(ctx context.Context, userID, text string) error {
	if u.client == nil {
		// No client configured, just log
		log.Printf("[WhatsApp] Reply to %s: %s (client not configured)", userID, text)
		return nil
	}

	// Send actual message via WhatsApp API
	return u.client.SendMessage(ctx, userID, text)
}
