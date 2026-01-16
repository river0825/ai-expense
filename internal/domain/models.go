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
