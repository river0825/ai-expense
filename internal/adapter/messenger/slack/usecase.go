package slack

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/riverlin/aiexpense/internal/usecase"
)

// UseCase handles Slack message processing
type UseCase struct {
	autoSignupUC        *usecase.AutoSignupUseCase
	parseConversationUC *usecase.ParseConversationUseCase
	createExpenseUC     *usecase.CreateExpenseUseCase
	client              *Client
}

// NewSlackUseCase creates a new Slack use case
func NewSlackUseCase(
	autoSignupUC *usecase.AutoSignupUseCase,
	parseConversationUC *usecase.ParseConversationUseCase,
	createExpenseUC *usecase.CreateExpenseUseCase,
	client *Client,
) *UseCase {
	return &UseCase{
		autoSignupUC:        autoSignupUC,
		parseConversationUC: parseConversationUC,
		createExpenseUC:     createExpenseUC,
		client:              client,
	}
}

// ProcessMessage processes an incoming Slack message
func (u *UseCase) ProcessMessage(ctx context.Context, userID, text string) error {
	if userID == "" || text == "" {
		return fmt.Errorf("user_id and text are required")
	}

	// Format user ID with Slack platform prefix
	platformUserID := fmt.Sprintf("slack_%s", userID)

	// Step 1: Auto-signup (idempotent)
	if err := u.autoSignupUC.Execute(ctx, platformUserID, "slack"); err != nil {
		log.Printf("Slack auto-signup failed: %v", err)
	}

	// Step 2: Parse the conversation
	parsedExpenses, err := u.parseConversationUC.Execute(ctx, text, platformUserID)
	if err != nil {
		log.Printf("Slack parsing failed: %v", err)
		return u.sendMessage(userID, "處理失敗，請稍後重試")
	}

	if len(parsedExpenses) == 0 {
		// No expenses parsed
		response := "I didn't find any expenses to record. Try saying something like:\n• 'breakfast $8'\n• 'lunch 12 coffee 5'\n• 'spent $50 on groceries'"
		return u.sendMessage(userID, response)
	}

	// Step 3: Create expenses
	var messages []string

	for _, parsed := range parsedExpenses {
		createReq := &usecase.CreateRequest{
			UserID:      platformUserID,
			Description: parsed.Description,
			Amount:      parsed.Amount,
			Date:        parsed.Date,
		}

		resp, err := u.createExpenseUC.Execute(ctx, createReq)
		if err != nil {
			log.Printf("Slack expense creation failed: %v", err)
			messages = append(messages, parsed.Description+" 儲存失敗")
			continue
		}

		messages = append(messages, resp.Message)
	}

	// Build consolidated response
	consolidatedMessage := ""
	for i, msg := range messages {
		if i > 0 {
			consolidatedMessage += "\n"
		}
		consolidatedMessage += msg
	}

	// Step 4: Send consolidated response
	return u.sendMessage(userID, consolidatedMessage)
}

// sendMessage sends a message via Slack
func (u *UseCase) sendMessage(userID, text string) error {
	if u.client == nil {
		log.Printf("[Slack] Reply to user %s: %s (client not configured)", userID, text)
		return nil
	}
	return u.client.SendMessage(userID, text)
}

// ProcessAppMention processes when the bot is mentioned in a channel
func (u *UseCase) ProcessAppMention(ctx context.Context, userID, text string) error {
	// Remove bot mention from text if present (Slack includes <@BOTID> in the text)
	cleanText := strings.TrimSpace(text)
	// Remove mention pattern like <@U12345>
	if idx := strings.Index(cleanText, ">"); idx != -1 {
		cleanText = strings.TrimSpace(cleanText[idx+1:])
	}

	if cleanText == "" {
		response := "Hi! I'm here to help you track expenses. Just send me expense information and I'll record it for you."
		return u.sendMessage(userID, response)
	}

	return u.ProcessMessage(ctx, userID, cleanText)
}
