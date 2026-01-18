package ai

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/riverlin/aiexpense/internal/domain"
)

// GeminiAI implements the AI Service using Google Gemini API
type GeminiAI struct {
	apiKey   string
	costRepo domain.AICostRepository
	// client *genai.Client // TODO: Initialize when Gemini SDK is available
}

// NewGeminiAI creates a new Gemini AI service
func NewGeminiAI(apiKey string, costRepo domain.AICostRepository) (*GeminiAI, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("Gemini API key is required")
	}

	// TODO: Initialize Gemini client
	// client, err := genai.NewClient(context.Background(), option.WithAPIKey(apiKey))
	// if err != nil {
	//     return nil, fmt.Errorf("failed to create Gemini client: %w", err)
	// }

	return &GeminiAI{
		apiKey:   apiKey,
		costRepo: costRepo,
		// client: client,
	}, nil
}

// ParseExpense extracts expenses from natural language text
// For now, uses regex-based fallback parsing until Gemini SDK is fully integrated
func (g *GeminiAI) ParseExpense(ctx context.Context, text string, userID string) ([]*domain.ParsedExpense, error) {
	// TODO: Call Gemini API with prompt engineered for expense parsing
	// For now, use regex-based fallback parsing
	expenses, err := g.parseExpenseRegex(text)
	if err != nil {
		return nil, err
	}

	// Log cost (using hardcoded estimates for now since it's regex fallback)
	// In real AI implementation, we would get usage from response
	if g.costRepo != nil && userID != "" {
		costLog := &domain.AICostLog{
			ID:           fmt.Sprintf("log_%d", time.Now().UnixNano()),
			UserID:       userID,
			Operation:    "parse_expense",
			Provider:     "gemini",
			Model:        "regex-fallback",
			InputTokens:  len(text), // Rough estimate
			OutputTokens: len(expenses) * 10,
			TotalTokens:  len(text) + len(expenses)*10,
			Cost:         0, // Regex is free
			Currency:     "USD",
			CreatedAt:    time.Now(),
		}
		// Fire and forget cost logging to not block response
		go func() {
			_ = g.costRepo.Create(context.Background(), costLog)
		}()
	}

	return expenses, nil
}

// parseExpenseRegex uses regex to extract expenses (fallback when AI unavailable)
func (g *GeminiAI) parseExpenseRegex(text string) ([]*domain.ParsedExpense, error) {
	// Pattern: description$amount or description amount
	// Examples: "早餐$20", "午餐 30", "加油$200"

	var expenses []*domain.ParsedExpense

	// Try to extract items with format: text$amount
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

		amount, err := strconv.ParseFloat(match[2], 64)
		if err != nil {
			continue
		}

		expense := &domain.ParsedExpense{
			Description:       description,
			Amount:            amount,
			SuggestedCategory: "Other", // Default category
			Date:              time.Now(),
		}
		expenses = append(expenses, expense)
	}

	return expenses, nil
}

// SuggestCategory suggests a category based on description
// For now, uses keyword matching until Gemini SDK is fully integrated
func (g *GeminiAI) SuggestCategory(ctx context.Context, description string, userID string) (string, error) {
	// TODO: Call Gemini API with category suggestion prompt
	// For now, use keyword-based matching
	category := g.suggestCategoryKeywords(description)

	// Log cost
	if g.costRepo != nil && userID != "" {
		costLog := &domain.AICostLog{
			ID:           fmt.Sprintf("log_%d", time.Now().UnixNano()),
			UserID:       userID,
			Operation:    "suggest_category",
			Provider:     "gemini",
			Model:        "keyword-fallback",
			InputTokens:  len(description), // Rough estimate
			OutputTokens: 10,               // Result is short string
			TotalTokens:  len(description) + 10,
			Cost:         0, // Keyword match is free
			Currency:     "USD",
			CreatedAt:    time.Now(),
		}
		// Fire and forget cost logging to not block response
		go func() {
			_ = g.costRepo.Create(context.Background(), costLog)
		}()
	}

	return category, nil
}

// suggestCategoryKeywords uses keyword matching for category suggestion (fallback)
func (g *GeminiAI) suggestCategoryKeywords(description string) string {
	description = strings.ToLower(description)

	foodKeywords := []string{"早餐", "午餐", "晚餐", "咖啡", "吃", "食物", "餐", "飯", "菜", "麵"}
	transportKeywords := []string{"加油", "公交", "計程車", "uber", "高鐵", "火車", "飛機", "停車", "油"}
	shoppingKeywords := []string{"買", "衣服", "鞋", "包", "購物", "店", "超市"}
	entertainmentKeywords := []string{"電影", "遊戲", "演唱會", "娛樂", "門票", "樂園"}

	for _, kw := range foodKeywords {
		if strings.Contains(description, kw) {
			return "Food"
		}
	}
	for _, kw := range transportKeywords {
		if strings.Contains(description, kw) {
			return "Transport"
		}
	}
	for _, kw := range shoppingKeywords {
		if strings.Contains(description, kw) {
			return "Shopping"
		}
	}
	for _, kw := range entertainmentKeywords {
		if strings.Contains(description, kw) {
			return "Entertainment"
		}
	}

	return "Other"
}
