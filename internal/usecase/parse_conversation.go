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
	aiService ai.Service
}

// NewParseConversationUseCase creates a new parse conversation use case
func NewParseConversationUseCase(aiService ai.Service) *ParseConversationUseCase {
	return &ParseConversationUseCase{
		aiService: aiService,
	}
}

// Execute parses conversation text and extracts expenses
func (u *ParseConversationUseCase) Execute(ctx context.Context, text string, userID string) ([]*domain.ParsedExpense, error) {
	// Call AI service to parse expenses
	expenses, err := u.aiService.ParseExpense(ctx, text, userID)
	if err != nil || len(expenses) == 0 {
		// Fallback to regex parsing if AI fails or returns no expenses
		expenses = u.parseWithRegex(text)
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

	return expenses, nil
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
