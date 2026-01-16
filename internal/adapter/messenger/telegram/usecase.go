package telegram

import (
	"context"
	"log"

	"github.com/riverlin/aiexpense/internal/usecase"
)

// TelegramUseCase handles Telegram-specific business logic
type TelegramUseCase struct {
	autoSignupUC        *usecase.AutoSignupUseCase
	parseConversationUC *usecase.ParseConversationUseCase
	createExpenseUC     *usecase.CreateExpenseUseCase
	client              *Client
}

// NewTelegramUseCase creates a new Telegram use case
func NewTelegramUseCase(
	autoSignupUC *usecase.AutoSignupUseCase,
	parseConversationUC *usecase.ParseConversationUseCase,
	createExpenseUC *usecase.CreateExpenseUseCase,
	client *Client,
) *TelegramUseCase {
	return &TelegramUseCase{
		autoSignupUC:        autoSignupUC,
		parseConversationUC: parseConversationUC,
		createExpenseUC:     createExpenseUC,
		client:              client,
	}
}

// HandleMessage handles an incoming Telegram message
func (u *TelegramUseCase) HandleMessage(ctx context.Context, userID string, chatID int64, text string) error {
	// Auto-signup user
	if err := u.autoSignupUC.Execute(ctx, userID, "telegram"); err != nil {
		log.Printf("Error auto-signing up user: %v", err)
	}

	// Parse conversation to extract expenses
	parsedExpenses, err := u.parseConversationUC.Execute(ctx, text, userID)
	if err != nil {
		log.Printf("Error parsing conversation: %v", err)
		return u.sendMessage(ctx, chatID, "處理失敗，請稍後重試")
	}

	if len(parsedExpenses) == 0 {
		return u.sendMessage(ctx, chatID, "找不到有效的消費項目。請提供金額和品項，例如：早餐$20")
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
			messages = append(messages, parsed.Description+" 儲存失敗")
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

	return u.sendMessage(ctx, chatID, consolidatedMessage)
}

// sendMessage sends a message via Telegram
func (u *TelegramUseCase) sendMessage(ctx context.Context, chatID int64, text string) error {
	if u.client == nil {
		// Mock sending for testing
		log.Printf("[Telegram] Reply to chat %d: %s", chatID, text)
		return nil
	}

	// TODO: Implement actual Telegram API call
	// return u.client.SendMessage(ctx, chatID, text)
	log.Printf("[Telegram] Reply to chat %d: %s", chatID, text)
	return nil
}
