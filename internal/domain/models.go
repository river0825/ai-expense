package domain

import (
	"context"
	"time"
)

// User represents a user in the system
type User struct {
	UserID        string    `db:"user_id"`
	MessengerType string    `db:"messenger_type"`
	CreatedAt     time.Time `db:"created_at"`
	HomeCurrency  string    `db:"home_currency"`
	Locale        string    `db:"locale"`
}

// Expense represents a single expense record
type Expense struct {
	ID             string    `db:"id"`
	UserID         string    `db:"user_id"`
	Description    string    `db:"description"`
	OriginalAmount float64   `db:"original_amount"`
	Currency       string    `db:"currency"`
	HomeAmount     float64   `db:"home_amount"`
	HomeCurrency   string    `db:"home_currency"`
	ExchangeRate   float64   `db:"exchange_rate"`
	CategoryID     *string   `db:"category_id"`
	ExpenseDate    time.Time `db:"expense_date"`
	CreatedAt      time.Time `db:"created_at"`
	UpdatedAt      time.Time `db:"updated_at"`
	Amount         float64   `db:"-"` // Deprecated: kept for backward compatibility until callers migrate to HomeAmount
}

// Currency represents a supported currency definition
type Currency struct {
	Code      string    `db:"code"`
	Symbol    string    `db:"symbol"`
	Aliases   []string  `db:"aliases"`
	IsActive  bool      `db:"is_active"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

// CurrencyTranslation stores localized currency names
type CurrencyTranslation struct {
	ID           int       `db:"id"`
	CurrencyCode string    `db:"currency_code"`
	Locale       string    `db:"locale"`
	Name         string    `db:"name"`
	CreatedAt    time.Time `db:"created_at"`
}

// ExchangeRate stores a cached conversion rate for a given day
type ExchangeRate struct {
	ID             int64     `db:"id"`
	Provider       string    `db:"provider"`
	BaseCurrency   string    `db:"base_currency"`
	TargetCurrency string    `db:"target_currency"`
	Rate           float64   `db:"rate"`
	RateDate       time.Time `db:"rate_date"`
	FetchedAt      time.Time `db:"fetched_at"`
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
	Currency          string
	CurrencyOriginal  string
	SuggestedCategory string
	Date              time.Time
}

// ParseResult represents the result of parsing a conversation
type ParseResult struct {
	Expenses     []*ParsedExpense
	SystemPrompt string
	RawResponse  string
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
	Currency     string    `db:"currency"`  // e.g., "USD"
	CostNote     *string   `db:"cost_note"` // Optional: reason for special cost (e.g., "pricing_not_configured")
	CreatedAt    time.Time `db:"created_at"`
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

// PricingConfig represents pricing configuration for an AI model
type PricingConfig struct {
	ID               string    `db:"id" json:"id"`
	Provider         string    `db:"provider" json:"provider"`
	Model            string    `db:"model" json:"model"`
	InputTokenPrice  float64   `db:"input_token_price" json:"input_token_price"`
	OutputTokenPrice float64   `db:"output_token_price" json:"output_token_price"`
	Currency         string    `db:"currency" json:"currency"`
	EffectiveDate    time.Time `db:"effective_date" json:"effective_date"`
	IsActive         bool      `db:"is_active" json:"is_active"`
	CreatedAt        time.Time `db:"created_at" json:"created_at"`
	UpdatedAt        time.Time `db:"updated_at" json:"updated_at"`
}

// InteractionLog represents a complete interaction trace for prompt engineering and debugging
type InteractionLog struct {
	ID            string    `db:"id" json:"id"`
	UserID        string    `db:"user_id" json:"user_id"`
	UserInput     string    `db:"user_input" json:"user_input"`
	SystemPrompt  string    `db:"system_prompt" json:"system_prompt"`
	AIRawResponse string    `db:"ai_raw_response" json:"ai_raw_response"`
	BotFinalReply string    `db:"bot_final_reply" json:"bot_final_reply"`
	DurationMs    int64     `db:"duration_ms" json:"duration_ms"` // processing time in milliseconds
	Error         string    `db:"error" json:"error"`             // any error message occurred
	Timestamp     time.Time `db:"timestamp" json:"timestamp"`
}

// PricingProvider defines the contract for fetching pricing from an AI provider
type PricingProvider interface {
	// Fetch retrieves current pricing from the provider
	Fetch(ctx context.Context) ([]*PricingConfig, error)

	// Provider returns the provider name (e.g., "gemini", "claude")
	Provider() string
}
