package domain

import "time"

// User represents a user in the system
type User struct {
	UserID        string    `db:"user_id"`
	MessengerType string    `db:"messenger_type"`
	CreatedAt     time.Time `db:"created_at"`
}

// Expense represents a single expense record
type Expense struct {
	ID          string    `db:"id"`
	UserID      string    `db:"user_id"`
	Description string    `db:"description"`
	Amount      float64   `db:"amount"`
	CategoryID  *string   `db:"category_id"`
	ExpenseDate time.Time `db:"expense_date"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

// Category represents an expense category
type Category struct {
	ID        string    `db:"id"`
	UserID    string    `db:"user_id"`
	Name      string    `db:"name"`
	IsDefault bool      `db:"is_default"`
	CreatedAt time.Time `db:"created_at"`
}

// CategoryKeyword maps keywords to categories
type CategoryKeyword struct {
	ID         string    `db:"id"`
	CategoryID string    `db:"category_id"`
	Keyword    string    `db:"keyword"`
	Priority   int       `db:"priority"`
	CreatedAt  time.Time `db:"created_at"`
}

// ParsedExpense represents an expense extracted from conversation
type ParsedExpense struct {
	Description       string
	Amount            float64
	SuggestedCategory string
	Date              time.Time
}

// DailyMetrics represents metrics for a single day
type DailyMetrics struct {
	Date           time.Time
	ActiveUsers    int
	TotalExpense   float64
	ExpenseCount   int
	AverageExpense float64
}

// CategoryMetrics represents metrics for a category
type CategoryMetrics struct {
	CategoryID string
	Category   string
	Total      float64
	Count      int
	Percent    float64
}

// AICostLog represents a record of AI API usage and cost
type AICostLog struct {
	ID           string    `db:"id"`
	UserID       string    `db:"user_id"`
	Operation    string    `db:"operation"` // e.g., "parse_expense", "suggest_category"
	Provider     string    `db:"provider"`  // e.g., "gemini", "openai"
	Model        string    `db:"model"`     // e.g., "gemini-2.5-lite"
	InputTokens  int       `db:"input_tokens"`
	OutputTokens int       `db:"output_tokens"`
	TotalTokens  int       `db:"total_tokens"`
	Cost         float64   `db:"cost"`
	Currency     string    `db:"currency"`   // e.g., "USD"
	CostNote     *string   `db:"cost_note"`  // Optional: reason for special cost (e.g., "pricing_not_configured")
	CreatedAt    time.Time `db:"created_at"`
}

// PricingConfig represents AI provider and model pricing information
type PricingConfig struct {
	ID               string    `db:"id"`
	Provider         string    `db:"provider"`         // e.g., "gemini", "claude", "openai"
	Model            string    `db:"model"`            // e.g., "gemini-2.5-lite"
	InputTokenPrice  float64   `db:"input_token_price"`  // USD per 1M tokens
	OutputTokenPrice float64   `db:"output_token_price"` // USD per 1M tokens
	Currency         string    `db:"currency"`         // e.g., "USD"
	EffectiveDate    time.Time `db:"effective_date"`   // When pricing becomes active
	IsActive         bool      `db:"is_active"`        // Whether pricing is currently used
	CreatedAt        time.Time `db:"created_at"`
	UpdatedAt        time.Time `db:"updated_at"`
}

// GetCost calculates the cost based on token usage and this pricing configuration
// Returns cost in USD (same as currency field)
func (p *PricingConfig) GetCost(inputTokens, outputTokens int) float64 {
	inputCost := float64(inputTokens) * p.InputTokenPrice / 1_000_000
	outputCost := float64(outputTokens) * p.OutputTokenPrice / 1_000_000
	return inputCost + outputCost
}

// Policy represents a legal document (Privacy Policy, Terms of Use)
type Policy struct {
	ID        string    `db:"id" json:"id"`
	Key       string    `db:"key" json:"key"` // e.g., "privacy_policy", "terms_of_use"
	Title     string    `db:"title" json:"title"`
	Content   string    `db:"content" json:"content"`
	Version   string    `db:"version" json:"version"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

// AICostSummary represents aggregated AI cost metrics
type AICostSummary struct {
	TotalCalls        int     `json:"total_calls"`
	TotalInputTokens  int     `json:"total_input_tokens"`
	TotalOutputTokens int     `json:"total_output_tokens"`
	TotalTokens       int     `json:"total_tokens"`
	TotalCost         float64 `json:"total_cost"`
	Currency          string  `json:"currency"`
}

// AICostDailyStats represents daily AI usage statistics
type AICostDailyStats struct {
	Date         time.Time `json:"date"`
	Calls        int       `json:"calls"`
	InputTokens  int       `json:"input_tokens"`
	OutputTokens int       `json:"output_tokens"`
	TotalTokens  int       `json:"total_tokens"`
	Cost         float64   `json:"cost"`
}

// AICostByOperation represents AI cost breakdown by operation type
type AICostByOperation struct {
	Operation    string  `json:"operation"`
	Calls        int     `json:"calls"`
	InputTokens  int     `json:"input_tokens"`
	OutputTokens int     `json:"output_tokens"`
	TotalTokens  int     `json:"total_tokens"`
	Cost         float64 `json:"cost"`
	Percent      float64 `json:"percent"`
}

// AICostByUser represents AI cost breakdown by user
type AICostByUser struct {
	UserID       string  `json:"user_id"`
	Calls        int     `json:"calls"`
	InputTokens  int     `json:"input_tokens"`
	OutputTokens int     `json:"output_tokens"`
	TotalTokens  int     `json:"total_tokens"`
	Cost         float64 `json:"cost"`
}
