package usecase

import (
	"context"
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
	if err != nil {
		// Fallback to regex parsing if AI fails
		expenses = u.parseWithRegex(text)
	}

	// Parse relative dates
	for _, expense := range expenses {
		expense.Date = u.parseDate(text)
	}

	return expenses, nil
}

// parseDate extracts relative dates from text (昨天, 上週, etc.)
func (u *ParseConversationUseCase) parseDate(text string) time.Time {
	text = strings.ToLower(text)

	// Check for yesterday
	if strings.Contains(text, "昨天") || strings.Contains(text, "昨日") {
		return time.Now().AddDate(0, 0, -1)
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
	var expenses []*domain.ParsedExpense

	// Pattern: description$amount
	re := regexp.MustCompile(`([^\d$]+)\$(\d+(?:\.\d{2})?)`)
	matches := re.FindAllStringSubmatch(text, -1)

	for _, match := range matches {
		if len(match) < 3 {
			continue
		}

		description := strings.TrimSpace(match[1])
		if description == "" {
			continue
		}

		amount := 0.0
		_, _ = parseFloat(match[2], &amount)

		expense := &domain.ParsedExpense{
			Description: description,
			Amount:      amount,
			Date:        time.Now(),
		}
		expenses = append(expenses, expense)
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
