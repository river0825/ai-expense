package usecase

import (
	"context"
	"fmt"
	"log"
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
	userRepo           domain.UserRepository
	generateReportLink domain.GenerateReportLinkUseCase
	interactionRepo    domain.InteractionLogRepository
	conversationRepo   domain.ConversationStateRepository
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
	userRepo domain.UserRepository,
	generateReportLink domain.GenerateReportLinkUseCase,
	interactionRepo domain.InteractionLogRepository,
	conversationRepo domain.ConversationStateRepository,
) *ProcessMessageUseCase {
	return &ProcessMessageUseCase{
		autoSignup:         autoSignup,
		parseConversation:  parseConversation,
		createExpense:      createExpense,
		getExpenses:        getExpenses,
		userRepo:           userRepo,
		generateReportLink: generateReportLink,
		interactionRepo:    interactionRepo,
		conversationRepo:   conversationRepo,
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

	if pendingResp, handled := u.handlePendingConversation(ctx, msg, msgLower); handled {
		botReply = pendingResp.Text
		return pendingResp, nil
	}

	if u.isSetCurrencyIntent(msgLower) {
		currency := u.extractCurrencyCode(msg.Content)
		if currency == "" {
			locale := u.getUserLocale(ctx, msg.UserID)
			if err := u.savePendingCurrencyClarification(ctx, msg.UserID); err != nil {
				botReply = "Sorry, I couldn't save your request. Please try again."
				return &domain.MessageResponse{
					Type: domain.ResponseTypeError,
					Text: botReply,
				}, nil
			}

			botReply = u.currencyClarificationPrompt(locale)
			return &domain.MessageResponse{
				Type: domain.ResponseTypeInfo,
				Text: botReply,
			}, nil
		}

		locale := u.getUserLocale(ctx, msg.UserID)
		if err := u.updateUserHomeCurrency(ctx, msg.UserID, currency); err != nil {
			botReply = "Sorry, I couldn't update your currency. Please try again later."
			return &domain.MessageResponse{
				Type: domain.ResponseTypeError,
				Text: botReply,
			}, nil
		}

		botReply = u.currencyUpdatedReply(locale, currency)
		return &domain.MessageResponse{
			Type: domain.ResponseTypeInfo,
			Text: botReply,
		}, nil
	}

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
			log.Printf("ERROR: Failed to create expense for user %s: %v", msg.UserID, err)
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
	var sb strings.Builder
	primaryCurrency := getPrimaryCurrency(createdExpenses)
	sb.WriteString(fmt.Sprintf("✓ Recorded %d expense(s), total: %s %s", len(createdExpenses), formatAmount(totalAmount), primaryCurrency))
	for _, exp := range createdExpenses {
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
		line := fmt.Sprintf("\n• [%s] %s (%s)", dateStr, exp["description"], exp["category"])
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

	botReply = sb.String()

	return &domain.MessageResponse{
		Type: domain.ResponseTypeExpense,
		Text: botReply,
		Data: createdExpenses,
	}, nil
}

func (u *ProcessMessageUseCase) isReportIntent(text string) bool {
	keywords := []string{"report", "summary", "stats", "chart", "analysis", "expense report", "show report", "報表", "報告", "統計"}
	for _, k := range keywords {
		if strings.Contains(text, k) {
			return true
		}
	}
	return false
}

func (u *ProcessMessageUseCase) isSetCurrencyIntent(text string) bool {
	actionKeywords := []string{"set", "change", "switch", "default", "currency", "切換", "改成", "預設", "幣別"}
	hasAction := false
	for _, k := range actionKeywords {
		if strings.Contains(text, k) {
			hasAction = true
			break
		}
	}
	if !hasAction {
		return false
	}

	currencyKeywords := []string{"jpy", "usd", "twd", "eur", "cny", "krw", "thb", "日幣", "美元", "台幣", "歐元", "人民幣", "韓元", "泰銖", "currency", "幣別"}
	for _, k := range currencyKeywords {
		if strings.Contains(text, k) {
			return true
		}
	}

	return strings.Contains(text, "幣") || strings.Contains(text, "currency")
}

func (u *ProcessMessageUseCase) handlePendingConversation(ctx context.Context, msg *domain.UserMessage, msgLower string) (*domain.MessageResponse, bool) {
	if u.conversationRepo == nil {
		return nil, false
	}

	state, err := u.conversationRepo.GetByUserID(ctx, msg.UserID)
	if err != nil || state == nil {
		return nil, false
	}

	locale := u.getUserLocale(ctx, msg.UserID)

	if time.Now().After(state.ExpiresAt) {
		_ = u.conversationRepo.DeleteByUserID(ctx, msg.UserID)
		if u.extractCurrencyCode(msg.Content) != "" {
			return &domain.MessageResponse{Type: domain.ResponseTypeInfo, Text: u.pendingStateExpiredReply(locale)}, true
		}
		return nil, false
	}

	if msgLower == "cancel" || msgLower == "取消" {
		_ = u.conversationRepo.DeleteByUserID(ctx, msg.UserID)
		return &domain.MessageResponse{Type: domain.ResponseTypeInfo, Text: u.pendingStateCancelledReply(locale)}, true
	}

	if state.ActiveIntent != "settings.currency.set" {
		return nil, false
	}

	targetCurrency := u.extractCurrencyCode(msg.Content)
	if targetCurrency == "" {
		return &domain.MessageResponse{Type: domain.ResponseTypeInfo, Text: u.currencyClarificationPrompt(locale)}, true
	}

	if err := u.updateUserHomeCurrency(ctx, msg.UserID, targetCurrency); err != nil {
		return &domain.MessageResponse{Type: domain.ResponseTypeError, Text: "Sorry, I couldn't update your currency. Please try again later."}, true
	}
	_ = u.conversationRepo.DeleteByUserID(ctx, msg.UserID)

	return &domain.MessageResponse{Type: domain.ResponseTypeInfo, Text: u.currencyUpdatedReply(locale, targetCurrency)}, true
}

func (u *ProcessMessageUseCase) savePendingCurrencyClarification(ctx context.Context, userID string) error {
	if u.conversationRepo == nil {
		return nil
	}
	now := time.Now()
	return u.conversationRepo.Upsert(ctx, &domain.ConversationState{
		UserID:       userID,
		ActiveIntent: "settings.currency.set",
		PendingSlots: map[string]string{"target_currency": ""},
		Status:       "pending",
		ExpiresAt:    now.Add(10 * time.Minute),
		CreatedAt:    now,
		UpdatedAt:    now,
	})
}

func (u *ProcessMessageUseCase) updateUserHomeCurrency(ctx context.Context, userID, currency string) error {
	if u.userRepo == nil {
		return fmt.Errorf("user repository is not configured")
	}
	user, err := u.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}
	if user == nil {
		return fmt.Errorf("user not found")
	}
	user.HomeCurrency = currency
	return u.userRepo.Update(ctx, user)
}

func (u *ProcessMessageUseCase) getUserLocale(ctx context.Context, userID string) string {
	if u.userRepo == nil {
		return "zh-TW"
	}
	user, err := u.userRepo.GetByID(ctx, userID)
	if err != nil || user == nil || user.Locale == "" {
		return "zh-TW"
	}
	return user.Locale
}

func (u *ProcessMessageUseCase) extractCurrencyCode(content string) string {
	normalized := strings.ToLower(strings.TrimSpace(content))
	replacer := strings.NewReplacer(" ", "", "-", "", "_", "")
	normalized = replacer.Replace(normalized)

	aliasToCode := map[string]string{
		"twd": "TWD", "ntd": "TWD", "nt$": "TWD", "台幣": "TWD", "新台幣": "TWD",
		"usd": "USD", "美金": "USD", "美元": "USD",
		"jpy": "JPY", "yen": "JPY", "日幣": "JPY", "日元": "JPY", "円": "JPY",
		"eur": "EUR", "euro": "EUR", "歐元": "EUR",
		"cny": "CNY", "rmb": "CNY", "人民幣": "CNY", "人民币": "CNY",
		"krw": "KRW", "韓元": "KRW", "韩元": "KRW", "won": "KRW",
		"thb": "THB", "泰銖": "THB", "泰铢": "THB", "baht": "THB",
	}

	for alias, code := range aliasToCode {
		if strings.Contains(normalized, strings.ToLower(alias)) {
			return code
		}
	}

	return ""
}

func (u *ProcessMessageUseCase) currencyClarificationPrompt(locale string) string {
	if strings.HasPrefix(strings.ToLower(locale), "en") {
		return "Which currency would you like to switch to? (e.g. JPY, USD, TWD)"
	}
	return "你要切換成哪個幣別？（例如 JPY、USD、TWD）"
}

func (u *ProcessMessageUseCase) currencyUpdatedReply(locale, currency string) string {
	if strings.HasPrefix(strings.ToLower(locale), "en") {
		return fmt.Sprintf("Done. Your default currency is now %s.", currency)
	}
	return fmt.Sprintf("已更新，預設幣別現在是 %s。", currency)
}

func (u *ProcessMessageUseCase) pendingStateExpiredReply(locale string) string {
	if strings.HasPrefix(strings.ToLower(locale), "en") {
		return "Your previous currency-switch request expired. Please ask again, for example: switch default currency to JPY."
	}
	return "先前的幣別切換已逾時，請重新告訴我要切換成哪個幣別。"
}

func (u *ProcessMessageUseCase) pendingStateCancelledReply(locale string) string {
	if strings.HasPrefix(strings.ToLower(locale), "en") {
		return "Got it. I cancelled the pending request."
	}
	return "好的，已取消目前的待確認操作。"
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
