package ai

import "context"

// Service defines the AI service interface for expense parsing and categorization
type Service interface {
	// ParseExpense extracts expenses from natural language text
	// Returns parsed expenses with actual token usage from API response
	ParseExpense(ctx context.Context, text string, userID string) (*ParseExpenseResponse, error)

	// SuggestCategory suggests a category based on description
	// Returns suggested category with actual token usage from API response
	SuggestCategory(ctx context.Context, description string, userID string) (*SuggestCategoryResponse, error)
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
