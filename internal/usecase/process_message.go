package usecase

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/riverlin/aiexpense/internal/domain"
)

// ProcessMessageUseCase handles the core logic for processing messages from any source
type ProcessMessageUseCase struct {
	autoSignup        AutoSignup
	parseConversation ParseConversation
	createExpense     CreateExpense
	getExpenses       GetExpenses
}

// Interfaces to break dependency cycles (if needed) or mock easier
type AutoSignup interface {
	Execute(ctx context.Context, userID, sourceType string) error
}

type ParseConversation interface {
	Execute(ctx context.Context, text, userID string) ([]*domain.ParsedExpense, error)
}

type CreateExpense interface {
	Execute(ctx context.Context, req *CreateRequest) (*CreateResponse, error)
}

type GetExpenses interface {
	ExecuteGetAll(ctx context.Context, req *GetAllRequest) (*GetAllResponse, error)
}

// NewProcessMessageUseCase creates a new use case
func NewProcessMessageUseCase(
	autoSignup AutoSignup,
	parseConversation ParseConversation,
	createExpense CreateExpense,
	getExpenses GetExpenses,
) *ProcessMessageUseCase {
	return &ProcessMessageUseCase{
		autoSignup:        autoSignup,
		parseConversation: parseConversation,
		createExpense:     createExpense,
		getExpenses:       getExpenses,
	}
}

// Execute processes the incoming UserMessage
func (u *ProcessMessageUseCase) Execute(ctx context.Context, msg *domain.UserMessage) (*domain.MessageResponse, error) {
	// 1. Auto-signup
	if err := u.autoSignup.Execute(ctx, msg.UserID, msg.Source); err != nil {
		return &domain.MessageResponse{
			Text: fmt.Sprintf("Failed to signup user: %v", err),
		}, nil // We return success to the adapter so it can send the error message back to user
	}

	// 2. Parse Message
	expenses, err := u.parseConversation.Execute(ctx, msg.Content, msg.UserID)
	if err != nil {
		return &domain.MessageResponse{
			Text: fmt.Sprintf("Failed to parse message: %v", err),
		}, nil
	}

	if len(expenses) == 0 {
		return &domain.MessageResponse{
			Text: "No expenses detected in message",
		}, nil
	}

	// 3. Create Expenses
	createdExpenses := []map[string]interface{}{}
	totalAmount := 0.0

	for _, parsedExp := range expenses {
		req := &CreateRequest{
			UserID:      msg.UserID,
			Description: parsedExp.Description,
			Amount:      parsedExp.Amount,
			Date:        parsedExp.Date,
		}

		resp, err := u.createExpense.Execute(ctx, req)
		if err != nil {
			// Log error but continue
			continue
		}

		totalAmount += parsedExp.Amount
		createdExpenses = append(createdExpenses, map[string]interface{}{
			"id":          resp.ID,
			"description": parsedExp.Description,
			"amount":      parsedExp.Amount,
			"category":    resp.Category,
			"date":        parsedExp.Date,
		})
	}

	// 4. Format Response
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("✓ Recorded %d expense(s), total: $%.2f", len(createdExpenses), totalAmount))
	for _, exp := range createdExpenses {
		dateStr := ""
		if d, ok := exp["date"].(time.Time); ok {
			dateStr = d.Format("2006-01-02")
		}
		sb.WriteString(fmt.Sprintf("\n• [%s] %s (%s): $%.2f",
			dateStr,
			exp["description"],
			exp["category"],
			exp["amount"]))
	}

	return &domain.MessageResponse{
		Text: sb.String(),
		Data: createdExpenses,
	}, nil
}
