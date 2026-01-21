package terminal

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/riverlin/aiexpense/internal/domain"
	"github.com/riverlin/aiexpense/internal/usecase"
)

// TerminalUseCase handles Terminal Chat message processing for local testing
type TerminalUseCase struct {
	autoSignup        *usecase.AutoSignupUseCase
	parseConversation *usecase.ParseConversationUseCase
	createExpense     *usecase.CreateExpenseUseCase
	getExpenses       *usecase.GetExpensesUseCase
	userRepo          domain.UserRepository
}

// NewTerminalUseCase creates a new Terminal Chat use case
func NewTerminalUseCase(
	autoSignup *usecase.AutoSignupUseCase,
	parseConversation *usecase.ParseConversationUseCase,
	createExpense *usecase.CreateExpenseUseCase,
	getExpenses *usecase.GetExpensesUseCase,
	userRepo domain.UserRepository,
) *TerminalUseCase {
	return &TerminalUseCase{
		autoSignup:        autoSignup,
		parseConversation: parseConversation,
		createExpense:     createExpense,
		getExpenses:       getExpenses,
		userRepo:          userRepo,
	}
}

// HandleMessage processes a terminal chat message
// It auto-signs up the user, parses the message, creates expenses, and returns a response
func (u *TerminalUseCase) HandleMessage(ctx context.Context, userID, message string) (*TerminalResponse, error) {
	// Auto-signup user (idempotent - does nothing if user exists)
	if err := u.autoSignup.Execute(ctx, userID, "terminal"); err != nil {
		return &TerminalResponse{
			Status:  "error",
			Message: "Failed to signup user: " + err.Error(),
		}, nil
	}

	// Parse the message to extract expenses
	expenses, err := u.parseConversation.Execute(ctx, message, userID)
	if err != nil {
		return &TerminalResponse{
			Status:  "error",
			Message: "Failed to parse message: " + err.Error(),
		}, nil
	}

	if len(expenses) == 0 {
		return &TerminalResponse{
			Status:  "success",
			Message: "No expenses detected in message",
			Data: map[string]interface{}{
				"user_id":         userID,
				"message":         message,
				"expenses_parsed": 0,
			},
		}, nil
	}

	// Create expenses
	createdExpenses := []map[string]interface{}{}
	totalAmount := 0.0

	for _, parsedExp := range expenses {
		req := &usecase.CreateRequest{
			UserID:      userID,
			Description: parsedExp.Description,
			Amount:      parsedExp.Amount,
			CategoryID:  nil,
			Date:        parsedExp.Date,
		}

		resp, err := u.createExpense.Execute(ctx, req)
		if err != nil {
			// Log the error but continue with other expenses
			continue
		}

		totalAmount += parsedExp.Amount
		createdExpenses = append(createdExpenses, map[string]interface{}{
			"id":          resp.ID,
			"description": parsedExp.Description,
			"amount":      parsedExp.Amount,
			"category":    resp.Category,
			"message":     resp.Message,
			"date":        parsedExp.Date,
		})
	}

	// Build response message
	var msgBuilder strings.Builder
	msgBuilder.WriteString(fmt.Sprintf("✓ Recorded %d expense(s), total: $%.2f", len(createdExpenses), totalAmount))
	for _, exp := range createdExpenses {
		dateStr := ""
		if d, ok := exp["date"].(time.Time); ok {
			dateStr = d.Format("2006-01-02")
		}
		msgBuilder.WriteString(fmt.Sprintf("\n• [%s] %s (%s): $%.2f",
			dateStr,
			exp["description"],
			exp["category"],
			exp["amount"]))
	}
	responseMsg := msgBuilder.String()

	return &TerminalResponse{
		Status:  "success",
		Message: responseMsg,
		Data: map[string]interface{}{
			"user_id":          userID,
			"expenses_created": len(createdExpenses),
			"total_amount":     totalAmount,
			"expenses":         createdExpenses,
			"original_message": message,
		},
	}, nil
}

// GetUserInfo retrieves user information including stats
func (u *TerminalUseCase) GetUserInfo(ctx context.Context, userID string) (map[string]interface{}, error) {
	// Get user
	user, err := u.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, fmt.Errorf("user not found")
	}

	// Get user's expenses
	getReq := &usecase.GetAllRequest{
		UserID: userID,
	}

	resp, err := u.getExpenses.ExecuteGetAll(ctx, getReq)
	if err != nil {
		resp = &usecase.GetAllResponse{Expenses: []*usecase.ExpenseDTO{}}
	}

	// Convert DTOs back to domain Expenses for stats calculation
	expenses := convertDTOsToDomainExpenses(resp.Expenses)

	// Calculate stats
	totalExpense := 0.0
	expenseCount := len(expenses)

	for _, exp := range expenses {
		totalExpense += exp.Amount
	}

	averageExpense := 0.0
	if expenseCount > 0 {
		averageExpense = totalExpense / float64(expenseCount)
	}

	return map[string]interface{}{
		"user_id":           user.UserID,
		"messenger_type":    user.MessengerType,
		"created_at":        user.CreatedAt,
		"total_expenses":    totalExpense,
		"expense_count":     expenseCount,
		"average_expense":   averageExpense,
		"last_expense_date": getLastExpenseDateFromDTOs(resp.Expenses),
	}, nil
}

// convertDTOsToDomainExpenses converts ExpenseDTO list to domain Expense list
func convertDTOsToDomainExpenses(dtos []*usecase.ExpenseDTO) []*domain.Expense {
	expenses := make([]*domain.Expense, len(dtos))
	for i, dto := range dtos {
		expenses[i] = &domain.Expense{
			ID:          dto.ID,
			Description: dto.Description,
			Amount:      dto.Amount,
			CategoryID:  dto.CategoryID,
			ExpenseDate: dto.Date,
		}
	}
	return expenses
}

// getLastExpenseDateFromDTOs returns the most recent expense date from DTOs
func getLastExpenseDateFromDTOs(dtos []*usecase.ExpenseDTO) interface{} {
	if len(dtos) == 0 {
		return nil
	}

	latest := dtos[0]
	for _, dto := range dtos[1:] {
		if dto.Date.After(latest.Date) {
			latest = dto
		}
	}

	return latest.Date
}

// FormatResponse formats the response for display
func (r *TerminalResponse) FormatResponse() string {
	var sb strings.Builder

	sb.WriteString("=== Terminal Chat Response ===\n")
	sb.WriteString(fmt.Sprintf("Status: %s\n", r.Status))
	sb.WriteString(fmt.Sprintf("Message: %s\n", r.Message))

	if r.Data != nil {
		sb.WriteString("Data:\n")
		if dataMap, ok := r.Data.(map[string]interface{}); ok {
			for key, value := range dataMap {
				sb.WriteString(fmt.Sprintf("  %s: %v\n", key, value))
			}
		}
	}

	return sb.String()
}
