package ai

import (
	"context"

	"github.com/riverlin/aiexpense/internal/domain"
)

// Service defines the AI service interface for expense parsing and categorization
type Service interface {
	// ParseExpense extracts expenses from natural language text
	ParseExpense(ctx context.Context, text string, userID string) ([]*domain.ParsedExpense, error)

	// SuggestCategory suggests a category based on description
	SuggestCategory(ctx context.Context, description string, userID string) (string, error)
}

// Factory creates an AI service based on the provider type
func Factory(provider string, apiKey string, costRepo domain.AICostRepository) (Service, error) {
	switch provider {
	case "gemini":
		return NewGeminiAI(apiKey, costRepo)
	case "claude":
		// TODO: Implement Claude AI
		return nil, nil
	case "openai":
		// TODO: Implement OpenAI
		return nil, nil
	default:
		return NewGeminiAI(apiKey, costRepo)
	}
}
