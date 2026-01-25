package usecase

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/riverlin/aiexpense/internal/ai"
	"github.com/riverlin/aiexpense/internal/domain"
)

// ParseConversationUseCase handles parsing of conversation text to extract expenses
type ParseConversationUseCase struct {
	aiService   ai.Service
	pricingRepo domain.PricingRepository
	costRepo    domain.AICostRepository
	provider    string // e.g., "gemini"
	model       string // e.g., "gemini-2.5-lite"
}

// NewParseConversationUseCase creates a new parse conversation use case
func NewParseConversationUseCase(
	aiService ai.Service,
	pricingRepo domain.PricingRepository,
	costRepo domain.AICostRepository,
	provider string,
	model string,
) *ParseConversationUseCase {
	return &ParseConversationUseCase{
		aiService:   aiService,
		pricingRepo: pricingRepo,
		costRepo:    costRepo,
		provider:    provider,
		model:       model,
	}
}

// Execute parses conversation text and extracts expenses with cost tracking
func (u *ParseConversationUseCase) Execute(ctx context.Context, text, userID string) (*domain.ParseResult, error) {
	// Call AI service to parse expenses (returns token metadata)
	resp, err := u.aiService.ParseExpense(ctx, text, userID)
	var expenses []*domain.ParsedExpense
	var tokens *ai.TokenMetadata
	var systemPrompt, rawResponse string

	if err != nil || resp == nil || len(resp.Expenses) == 0 {
		// Fallback to regex parsing if AI fails or returns no expenses
		expenses = u.parseWithRegex(text)
		tokens = &ai.TokenMetadata{InputTokens: 0, OutputTokens: 0, TotalTokens: 0}
	} else {
		expenses = resp.Expenses
		tokens = resp.Tokens
		systemPrompt = resp.SystemPrompt
		rawResponse = resp.RawResponse
	}

	// Parse relative dates ONLY if date is zero (not set by AI)
	for _, expense := range expenses {
		if expense.Date.IsZero() {
			log.Printf("DEBUG: Expense date is zero, parsing relative date from text: %s", text)
			expense.Date = u.parseDate(text)
		} else {
			log.Printf("DEBUG: Expense date already set (by AI?): %v", expense.Date)
		}
	}

	// Log cost asynchronously (if pricing available)
	go u.logCost(context.Background(), userID, tokens)

	return &domain.ParseResult{
		Expenses:     expenses,
		SystemPrompt: systemPrompt,
		RawResponse:  rawResponse,
	}, nil
}

// logCost calculates and logs the cost of the AI API call
func (u *ParseConversationUseCase) logCost(ctx context.Context, userID string, tokens *ai.TokenMetadata) {
	if tokens == nil || u.costRepo == nil || u.pricingRepo == nil {
		return
	}

	// Skip logging if no tokens were used (fallback parsing or zero input)
	if tokens.TotalTokens == 0 {
		return
	}

	// Look up pricing for provider/model
	pricing, err := u.pricingRepo.GetByProviderAndModel(ctx, u.provider, u.model)
	if err != nil {
		log.Printf("ERROR: Failed to lookup pricing for %s/%s: %v", u.provider, u.model, err)
		return
	}

	var cost float64
	var costNote *string
	if pricing == nil {
		// Pricing not configured
		cost = 0
		msg := "pricing_not_configured"
		costNote = &msg
		log.Printf("WARN: Pricing not configured for %s/%s", u.provider, u.model)
	} else {
		// Calculate cost
		cost = pricing.GetCost(tokens.InputTokens, tokens.OutputTokens)
	}

	// Create and persist cost log
	costLog := &domain.AICostLog{
		ID:           fmt.Sprintf("log_%d", time.Now().UnixNano()),
		UserID:       userID,
		Operation:    "parse_conversation",
		Provider:     u.provider,
		Model:        u.model,
		InputTokens:  tokens.InputTokens,
		OutputTokens: tokens.OutputTokens,
		TotalTokens:  tokens.TotalTokens,
		Cost:         cost,
		Currency:     "USD",
		CostNote:     costNote,
		CreatedAt:    time.Now().UTC(),
	}

	if err := u.costRepo.Create(ctx, costLog); err != nil {
		log.Printf("ERROR: Failed to log cost: %v", err)
	}
}

// parseDate extracts relative dates from text (昨天, 上週, etc.)
func (u *ParseConversationUseCase) parseDate(text string) time.Time {
	text = strings.ToLower(text)
	log.Printf("DEBUG: parseDate called with: %s", text)

	// Check for day before yesterday (前天) - MUST check before yesterday
	if strings.Contains(text, "前天") || strings.Contains(text, "前日") {
		d := time.Now().AddDate(0, 0, -2)
		log.Printf("DEBUG: Detect '前天', returning %v", d)
		return d
	}

	// Check for yesterday (昨天)
	if strings.Contains(text, "昨天") || strings.Contains(text, "昨日") {
		return time.Now().AddDate(0, 0, -1)
	}

	// Check for tomorrow (明天)
	if strings.Contains(text, "明天") || strings.Contains(text, "明日") {
		return time.Now().AddDate(0, 0, 1)
	}

	// Check for day after tomorrow (後天)
	if strings.Contains(text, "後天") || strings.Contains(text, "后天") {
		return time.Now().AddDate(0, 0, 2)
	}

	// Check for last week
	if strings.Contains(text, "上週") || strings.Contains(text, "上周") {
		return time.Now().AddDate(0, 0, -7)
	}

	// Check for last month
	if strings.Contains(text, "上個月") || strings.Contains(text, "上月") {
		return time.Now().AddDate(0, -1, 0)
	}

	// Default to today
	return time.Now()
}

// parseWithRegex uses regex to extract expenses (fallback)
func (u *ParseConversationUseCase) parseWithRegex(text string) []*domain.ParsedExpense {
	// Debug log
	log.Printf("DEBUG: parseWithRegex called with: %s\n", text)
	var expenses []*domain.ParsedExpense

	// Helper
	addExpense := func(desc, amtStr string) {
		description := strings.TrimSpace(desc)
		if description == "" {
			return
		}
		description = strings.TrimSuffix(description, " ")
		amount := 0.0
		// Use simple ParseFloat since helper parseFloat isn't exported/shared easily
		// or use the one defined in this file.
		// The existing file has parseFloat helper at bottom.
		val, err := parseFloat(amtStr, &amount)
		if err != nil {
			return
		}
		amount = val

		expense := &domain.ParsedExpense{
			Description: description,
			Amount:      amount,
			// Date:        time.Now(), // DON'T SET DATE HERE, let Execute() handle it
		}
		expenses = append(expenses, expense)
	}

	// Pattern 1: description$amount
	reDollar := regexp.MustCompile(`([^\d$]+?)\s*\$(\d+(?:\.\d{2})?)`)
	dollarMatches := reDollar.FindAllStringSubmatch(text, -1)
	fmt.Printf("DEBUG: dollarMatches: %v\n", dollarMatches)

	// Pattern 2: description amount 元
	reYuan := regexp.MustCompile(`(.*?)\s+(\d+(?:\.\d{2})?)\s*元`)
	yuanMatches := reYuan.FindAllStringSubmatch(text, -1)
	fmt.Printf("DEBUG: yuanMatches: %v\n", yuanMatches)

	if len(dollarMatches) > 0 || len(yuanMatches) > 0 {
		for _, match := range dollarMatches {
			addExpense(match[1], match[2])
		}
		for _, match := range yuanMatches {
			addExpense(match[1], match[2])
		}
	} else {
		// Pattern 3: Loose space
		reSpace := regexp.MustCompile(`([^\d]+?)\s+(\d+(?:\.\d{2})?)(?:\s|$)`)
		matches := reSpace.FindAllStringSubmatch(text, -1)
		fmt.Printf("DEBUG: reSpace matches: %v\n", matches)
		for _, match := range matches {
			addExpense(match[1], match[2])
		}
	}

	return expenses
}

// Helper function for parsing float
func parseFloat(s string, f *float64) (float64, error) {
	result := 0.0
	for i := 0; i < len(s); i++ {
		if s[i] >= '0' && s[i] <= '9' {
			result = result*10 + float64(s[i]-'0')
		} else if s[i] == '.' {
			// Handle decimal part
			decimal := 0.1
			for i++; i < len(s); i++ {
				if s[i] >= '0' && s[i] <= '9' {
					result += decimal * float64(s[i]-'0')
					decimal *= 0.1
				}
			}
			break
		}
	}
	*f = result
	return result, nil
}
