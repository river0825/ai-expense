package domain

import "context"

// AIService defines operations for AI-powered services
type AIService interface {
	// ParseExpense extracts expenses from natural language text
	ParseExpense(ctx context.Context, text string, userID string) ([]*ParsedExpense, error)

	// SuggestCategory suggests a category based on description
	SuggestCategory(ctx context.Context, description string) (string, error)
}

// ConversationParser defines operations for parsing user messages
type ConversationParser interface {
	// Parse extracts expenses from conversation text
	Parse(ctx context.Context, text string, userID string) ([]*ParsedExpense, error)
}

// ReportGenerator defines operations for generating reports
type ReportGenerator interface {
	// GenerateSummary generates a summary report
	GenerateSummary(ctx context.Context, userID string) (string, error)

	// GenerateCategoryBreakdown generates category breakdown
	GenerateCategoryBreakdown(ctx context.Context, userID string) (string, error)
}

// MessengerService defines operations for sending messages to users
type MessengerService interface {
	// SendMessage sends a message to a user
	SendMessage(ctx context.Context, userID, message string) error

	// HandleWebhook handles incoming webhook events
	HandleWebhook(ctx context.Context, body []byte) error
}
