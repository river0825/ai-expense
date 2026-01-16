package teams

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/riverlin/aiexpense/internal/usecase"
)

// UseCase handles Microsoft Teams message processing
type UseCase struct {
	autoSignupUC        *usecase.AutoSignupUseCase
	parseConversationUC *usecase.ParseConversationUseCase
	createExpenseUC     *usecase.CreateExpenseUseCase
	client              *Client
}

// NewTeamsUseCase creates a new Teams use case
func NewTeamsUseCase(
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

// ProcessMessage processes an incoming Teams message
func (u *UseCase) ProcessMessage(ctx context.Context, userID, text string) error {
	if userID == "" || text == "" {
		return fmt.Errorf("user_id and text are required")
	}

	// Format user ID with Teams platform prefix
	platformUserID := fmt.Sprintf("teams_%s", userID)

	// Step 1: Auto-signup (idempotent)
	if err := u.autoSignupUC.Execute(ctx, platformUserID, "teams"); err != nil {
		log.Printf("Teams auto-signup failed: %v", err)
	}

	// Step 2: Parse the conversation
	parsedExpenses, err := u.parseConversationUC.Execute(ctx, text, platformUserID)
	if err != nil {
		log.Printf("Teams parsing failed: %v", err)
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
			log.Printf("Teams expense creation failed: %v", err)
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

// sendMessage sends a message to a Teams user
func (u *UseCase) sendMessage(userID, text string) error {
	if u.client == nil {
		log.Printf("[Teams] Reply to user %s: %s (client not configured)", userID, text)
		return nil
	}
	return u.client.SendMessage(userID, text)
}

// ProcessMention handles when the bot is mentioned in Teams
func (u *UseCase) ProcessMention(ctx context.Context, userID, text string) error {
	// Remove bot mention from text if present
	cleanText := strings.TrimSpace(text)
	// Remove mention pattern like <at>BotName</at>
	for strings.Contains(cleanText, "<at>") && strings.Contains(cleanText, "</at>") {
		start := strings.Index(cleanText, "<at>")
		end := strings.Index(cleanText, "</at>")
		if start >= 0 && end > start {
			cleanText = strings.TrimSpace(cleanText[:start] + cleanText[end+5:])
		} else {
			break
		}
	}

	if cleanText == "" {
		response := "Hi! I'm here to help you track expenses. Just send me expense information and I'll record it for you."
		return u.sendMessage(userID, response)
	}

	return u.ProcessMessage(ctx, userID, cleanText)
}
