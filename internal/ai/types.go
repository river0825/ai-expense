package ai

import "github.com/riverlin/aiexpense/internal/domain"

// TokenMetadata represents actual token usage from an AI API response
type TokenMetadata struct {
	InputTokens  int
	OutputTokens int
	TotalTokens  int
}

// ParseExpenseResponse wraps parsed expenses with token metadata
type ParseExpenseResponse struct {
	Expenses []*domain.ParsedExpense
	Tokens   *TokenMetadata
}

// SuggestCategoryResponse wraps suggested category with token metadata
type SuggestCategoryResponse struct {
	Category string
	Tokens   *TokenMetadata
}
