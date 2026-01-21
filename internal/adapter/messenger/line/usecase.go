package line

import (
	"context"
	"log"

	"github.com/riverlin/aiexpense/internal/usecase"
)

// LineUseCase handles LINE-specific business logic
type LineUseCase struct {
	autoSignupUC        *usecase.AutoSignupUseCase
	parseConversationUC *usecase.ParseConversationUseCase
	createExpenseUC     *usecase.CreateExpenseUseCase
	client              *Client
}

// NewLineUseCase creates a new LINE use case
func NewLineUseCase(
	autoSignupUC *usecase.AutoSignupUseCase,
	parseConversationUC *usecase.ParseConversationUseCase,
	createExpenseUC *usecase.CreateExpenseUseCase,
	client *Client,
) *LineUseCase {
	return &LineUseCase{
		autoSignupUC:        autoSignupUC,
		parseConversationUC: parseConversationUC,
		createExpenseUC:     createExpenseUC,
		client:              client,
	}
}

// HandleMessage handles an incoming LINE message
func (u *LineUseCase) HandleMessage(ctx context.Context, userID, text, replyToken string) error {
	// Auto-signup user
	if err := u.autoSignupUC.Execute(ctx, userID, "line"); err != nil {
		log.Printf("Error auto-signing up user: %v", err)
	}

	// Parse conversation to extract expenses
	parsedExpenses, err := u.parseConversationUC.Execute(ctx, text, userID)
	if err != nil {
		log.Printf("Error parsing conversation: %v", err)
		return u.sendMessage(ctx, replyToken, "處理失敗，請稍後重試")
	}

	if len(parsedExpenses) == 0 {
		return u.sendMessage(ctx, replyToken, "找不到有效的消費項目。請提供金額和品項，例如：早餐$20")
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

	return u.sendMessage(ctx, replyToken, consolidatedMessage)
}

// sendMessage sends a message via LINE
func (u *LineUseCase) sendMessage(ctx context.Context, replyToken, text string) error {
	if u.client == nil {
		// No client configured, just log
		log.Printf("[LINE] Reply to %s: %s (client not configured)", replyToken, text)
		return nil
	}

	log.Printf("[LINE] Sending reply to %s: %s", replyToken, text)
	// Send actual message via LINE Messaging API
	if err := u.client.SendMessage(ctx, replyToken, text); err != nil {
		log.Printf("[LINE] Failed to send message: %v", err)
		return err
	}
	log.Printf("[LINE] Reply sent successfully")
	return nil
}
