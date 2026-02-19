package usecase

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/riverlin/aiexpense/internal/domain"
)

// ProcessMessageUseCase handles the core logic for processing messages from any source
type ProcessMessageUseCase struct {
	autoSignup         AutoSignup
	parseConversation  ParseConversation
	createExpense      CreateExpense
	getExpenses        GetExpenses
	generateReportLink domain.GenerateReportLinkUseCase
	interactionRepo    domain.InteractionLogRepository
	expenseRepo        domain.ExpenseRepository
	logger             *slog.Logger
}

// Interfaces to break dependency cycles (if needed) or mock easier
type AutoSignup interface {
	Execute(ctx context.Context, userID, sourceType string) error
}

type ParseConversation interface {
	Execute(ctx context.Context, text, userID string) (*domain.ParseResult, error)
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
	generateReportLink domain.GenerateReportLinkUseCase,
	interactionRepo domain.InteractionLogRepository,
	expenseRepo domain.ExpenseRepository,
	logger *slog.Logger,
) *ProcessMessageUseCase {
	return &ProcessMessageUseCase{
		autoSignup:         autoSignup,
		parseConversation:  parseConversation,
		createExpense:      createExpense,
		getExpenses:        getExpenses,
		generateReportLink: generateReportLink,
		interactionRepo:    interactionRepo,
		expenseRepo:        expenseRepo,
		logger:             logger,
	}
}

// Execute processes the incoming UserMessage
func (u *ProcessMessageUseCase) Execute(ctx context.Context, msg *domain.UserMessage) (*domain.MessageResponse, error) {
	start := time.Now()
	var botReply string
	var err error
	var systemPrompt, rawResponse string

	defer func() {
		// Log interaction asynchronously
		if u.interactionRepo != nil {
			go func() {
				// Use a background context for logging to ensure it completes even if request cancels
				logCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()

				errMsg := ""
				if err != nil {
					errMsg = err.Error()
				}

				interactionLog := &domain.InteractionLog{
					ID:            fmt.Sprintf("int_%d", start.UnixNano()),
					UserID:        msg.UserID,
					UserInput:     msg.Content,
					SystemPrompt:  systemPrompt,
					AIRawResponse: rawResponse,
					BotFinalReply: botReply,
					DurationMs:    time.Since(start).Milliseconds(),
					Error:         errMsg,
					Timestamp:     start,
				}
				_ = u.interactionRepo.Create(logCtx, interactionLog)
			}()
		}
	}()

	// 1. Auto-signup
	if err = u.autoSignup.Execute(ctx, msg.UserID, msg.Source); err != nil {
		botReply = fmt.Sprintf("Failed to signup user: %v", err)
		return &domain.MessageResponse{
			Type: domain.ResponseTypeError,
			Text: botReply,
		}, nil // We return success to the adapter so it can send the error message back to user
	}

	// 1.5. Check for "View Report" intent
	msgLower := strings.ToLower(strings.TrimSpace(msg.Content))
	if u.isReportIntent(msgLower) {
		link, err := u.generateReportLink.Execute(msg.UserID)
		if err != nil {
			// Log the error for debugging
			fmt.Printf("Error generating report link: %v\n", err)
			botReply = "Sorry, I couldn't generate the report link. Please try again later."
			return &domain.MessageResponse{
				Type: domain.ResponseTypeError,
				Text: botReply,
			}, nil
		}

		botReply = fmt.Sprintf("Here is your expense report:\n%s\n(Link valid for 5 minutes)", link)
		return &domain.MessageResponse{
			Type: domain.ResponseTypeReport,
			Text: botReply,
			Data: map[string]string{"link": link},
		}, nil
	}

	// 1.8 Check for duplicate message (Idempotency)
	if msg.MessageID != "" && u.expenseRepo != nil {
		existingExpenses, err := u.expenseRepo.GetBySourceMessageID(ctx, msg.MessageID)
		if err != nil {
			u.logger.Warn("Failed to check for duplicate message", "error", err)
		} else if len(existingExpenses) > 0 {
			u.logger.Info("Duplicate message detected", "message_id", msg.MessageID, "count", len(existingExpenses))
			systemPrompt = "IDEMPOTENCY_CHECK"
			rawResponse = fmt.Sprintf("Skipped AI processing due to duplicate message (found %d existing expenses)", len(existingExpenses))

			// Construct response from existing expenses
			createdExpenses := []map[string]interface{}{}
			totalAmount := 0.0

			for _, exp := range existingExpenses {
				totalAmount += exp.HomeAmount
				createdExpenses = append(createdExpenses, map[string]interface{}{
					"id":              exp.ID,
					"description":     exp.Description,
					"original_amount": exp.OriginalAmount,
					"currency":        exp.Currency,
					"home_amount":     exp.HomeAmount,
					"home_currency":   exp.HomeCurrency,
					"category":        "", // Category name not readily available, skipping
					"date":            exp.ExpenseDate,
					"account":         exp.Account,
				})
			}

			// Format response (similar to step 4)
			botReply = u.formatExpenseResponse(createdExpenses, totalAmount)
			return &domain.MessageResponse{
				Type: domain.ResponseTypeExpense,
				Text: botReply,
				Data: createdExpenses,
			}, nil
		}
	}

	// 2. Parse Message
	var parseResult *domain.ParseResult
	parseResult, err = u.parseConversation.Execute(ctx, msg.Content, msg.UserID)
	if err != nil {
		botReply = fmt.Sprintf("Failed to parse message: %v", err)
		return &domain.MessageResponse{
			Type: domain.ResponseTypeError,
			Text: botReply,
		}, nil
	}

	systemPrompt = parseResult.SystemPrompt
	rawResponse = parseResult.RawResponse
	expenses := parseResult.Expenses

	if len(expenses) == 0 {
		botReply = "No expenses detected in message"
		return &domain.MessageResponse{
			Type: domain.ResponseTypeInfo,
			Text: botReply,
		}, nil
	}

	// 3. Create Expenses
	createdExpenses := []map[string]interface{}{}
	totalAmount := 0.0

	for i, parsedExp := range expenses {
		var sourceMessageID *string
		if msg.MessageID != "" {
			id := fmt.Sprintf("%s_%d", msg.MessageID, i)
			sourceMessageID = &id
		}

		req := &CreateRequest{
			UserID:           msg.UserID,
			Description:      parsedExp.Description,
			Amount:           parsedExp.Amount,
			Currency:         parsedExp.Currency,
			CurrencyOriginal: parsedExp.CurrencyOriginal,
			Account:          parsedExp.Account,
			SourceMessageID:  sourceMessageID,
			Date:             parsedExp.Date,
		}

		resp, err := u.createExpense.Execute(ctx, req)
		if err != nil {
			u.logger.Error("Failed to create expense", "user_id", msg.UserID, "error", err)
			continue
		}

		totalAmount += resp.HomeAmount
		account := resp.Account
		if account == "" {
			account = parsedExp.Account
		}
		createdExpenses = append(createdExpenses, map[string]interface{}{
			"id":              resp.ID,
			"description":     parsedExp.Description,
			"original_amount": resp.OriginalAmount,
			"currency":        resp.Currency,
			"home_amount":     resp.HomeAmount,
			"home_currency":   resp.HomeCurrency,
			"category":        resp.Category,
			"date":            parsedExp.Date,
			"account":         account,
		})
	}

	// 4. Format Response
	botReply = u.formatExpenseResponse(createdExpenses, totalAmount)

	return &domain.MessageResponse{
		Type: domain.ResponseTypeExpense,
		Text: botReply,
		Data: createdExpenses,
	}, nil
}

func (u *ProcessMessageUseCase) formatExpenseResponse(expenses []map[string]interface{}, totalAmount float64) string {
	var sb strings.Builder
	primaryCurrency := getPrimaryCurrency(expenses)
	sb.WriteString(fmt.Sprintf("✓ Recorded %d expense(s), total: %s %s", len(expenses), formatAmount(totalAmount), primaryCurrency))
	for _, exp := range expenses {
		dateStr := ""
		if d, ok := exp["date"].(time.Time); ok {
			dateStr = d.Format("2006-01-02")
		}
		homeAmount := asFloat(exp["home_amount"])
		homeCurrency, _ := exp["home_currency"].(string)
		account, _ := exp["account"].(string)
		if homeCurrency == "" {
			homeCurrency = "TWD"
		}
		if homeAmount == 0 {
			homeAmount = asFloat(exp["original_amount"])
		}
		description, _ := exp["description"].(string)
		category, _ := exp["category"].(string)

		line := fmt.Sprintf("\n• [%s] %s", dateStr, description)
		if category != "" {
			line = fmt.Sprintf("%s (%s)", line, category)
		}
		if account != "" {
			line = fmt.Sprintf("%s [%s]", line, account)
		}
		line = fmt.Sprintf("%s: %s %s", line, formatAmount(homeAmount), homeCurrency)
		if orig := asFloat(exp["original_amount"]); orig > 0 {
			if curr, _ := exp["currency"].(string); curr != "" && curr != homeCurrency {
				line = fmt.Sprintf("%s (≈ %s %s)", line, formatAmount(orig), curr)
			}
		}
		sb.WriteString(line)
	}
	return sb.String()
}

func (u *ProcessMessageUseCase) isReportIntent(text string) bool {
	keywords := []string{"report", "summary", "stats", "chart", "analysis", "expense report", "show report"}
	for _, k := range keywords {
		if strings.Contains(text, k) {
			return true
		}
	}
	return false
}

func getPrimaryCurrency(expenses []map[string]interface{}) string {
	for _, exp := range expenses {
		if currency, ok := exp["home_currency"].(string); ok && currency != "" {
			return currency
		}
	}
	return "TWD"
}

func asFloat(v interface{}) float64 {
	switch val := v.(type) {
	case float64:
		return val
	case float32:
		return float64(val)
	default:
		return 0
	}
}
