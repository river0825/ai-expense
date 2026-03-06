package ai

import (
	"context"

	"github.com/riverlin/aiexpense/internal/domain"
)

// Service defines the AI service interface for expense parsing and categorization
type Service interface {
	// ParseExpense extracts expenses from natural language text
	// Returns parsed expenses with actual token usage from API response
	// userCtx provides user preferences (currency, categories) for personalized prompts
	ParseExpense(ctx context.Context, text string, userCtx *domain.UserContext) (*ParseExpenseResponse, error)

	// SuggestCategory suggests a category based on description
	// Returns suggested category with actual token usage from API response
	// userCtx provides user preferences for personalized suggestions
	SuggestCategory(ctx context.Context, description string, userCtx *domain.UserContext) (*SuggestCategoryResponse, error)

	// ClassifyIntent classifies the intent of a user message
	// Returns classified intent with actual token usage from API response
	// userCtx provides user preferences for personalized classification
	ClassifyIntent(ctx context.Context, text string, userCtx *domain.UserContext) (*ClassifyIntentResponse, error)
}

// Factory creates an AI service based on the provider type
// Note: costRepo parameter is deprecated and kept only for backward compatibility during migration
func Factory(provider string, apiKey string, model string, costRepo interface{}) (Service, error) {
	switch provider {
	case "gemini":
		return NewGeminiAI(apiKey, model, nil)
	case "claude":
		// TODO: Implement Claude AI
		return nil, nil
	case "openai":
		// TODO: Implement OpenAI
		return nil, nil
	default:
		return NewGeminiAI(apiKey, model, nil)
	}
}
