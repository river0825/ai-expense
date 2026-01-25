package domain

import (
	"context"
	"time"
)

// UserRepository defines operations for user data
type UserRepository interface {
	// Create creates a new user
	Create(ctx context.Context, user *User) error

	// GetByID retrieves a user by ID
	GetByID(ctx context.Context, userID string) (*User, error)

	// Exists checks if a user exists
	Exists(ctx context.Context, userID string) (bool, error)
}

// ExpenseRepository defines operations for expense data
type ExpenseRepository interface {
	// Create creates a new expense
	Create(ctx context.Context, expense *Expense) error

	// GetByID retrieves an expense by ID
	GetByID(ctx context.Context, id string) (*Expense, error)

	// GetByUserID retrieves all expenses for a user
	GetByUserID(ctx context.Context, userID string) ([]*Expense, error)

	// GetByUserIDAndDateRange retrieves expenses for a user within a date range
	GetByUserIDAndDateRange(ctx context.Context, userID string, from, to time.Time) ([]*Expense, error)

	// GetByUserIDAndCategory retrieves expenses for a user in a category
	GetByUserIDAndCategory(ctx context.Context, userID, categoryID string) ([]*Expense, error)

	// Update updates an existing expense
	Update(ctx context.Context, expense *Expense) error

	// Delete deletes an expense
	Delete(ctx context.Context, id string) error
}

// CategoryRepository defines operations for category data
type CategoryRepository interface {
	// Create creates a new category
	Create(ctx context.Context, category *Category) error

	// GetByID retrieves a category by ID
	GetByID(ctx context.Context, id string) (*Category, error)

	// GetByUserID retrieves all categories for a user
	GetByUserID(ctx context.Context, userID string) ([]*Category, error)

	// GetByUserIDAndName retrieves a category by user and name
	GetByUserIDAndName(ctx context.Context, userID, name string) (*Category, error)

	// Update updates a category
	Update(ctx context.Context, category *Category) error

	// Delete deletes a category
	Delete(ctx context.Context, id string) error

	// CreateKeyword creates a keyword mapping
	CreateKeyword(ctx context.Context, keyword *CategoryKeyword) error

	// GetKeywordsByCategory retrieves keywords for a category
	GetKeywordsByCategory(ctx context.Context, categoryID string) ([]*CategoryKeyword, error)

	// DeleteKeyword deletes a keyword mapping
	DeleteKeyword(ctx context.Context, id string) error
}

// MetricsRepository defines operations for metrics queries
type MetricsRepository interface {
	// GetDailyActiveUsers retrieves DAU for a date range
	GetDailyActiveUsers(ctx context.Context, from, to time.Time) ([]*DailyMetrics, error)

	// GetExpensesSummary retrieves expense totals by date
	GetExpensesSummary(ctx context.Context, from, to time.Time) ([]*DailyMetrics, error)

	// GetCategoryTrends retrieves expense breakdown by category
	GetCategoryTrends(ctx context.Context, userID string, from, to time.Time) ([]*CategoryMetrics, error)

	// GetGrowthMetrics retrieves user growth metrics
	GetGrowthMetrics(ctx context.Context, days int) (map[string]interface{}, error)

	// GetNewUsersPerDay retrieves new users created per day
	GetNewUsersPerDay(ctx context.Context, from, to time.Time) ([]*DailyMetrics, error)
}

// AICostRepository defines operations for AI cost logging
type AICostRepository interface {
	// Create creates a new cost log entry
	Create(ctx context.Context, log *AICostLog) error

	// GetByUserID retrieves cost logs for a user
	GetByUserID(ctx context.Context, userID string, limit int) ([]*AICostLog, error)

	// GetSummary retrieves aggregated AI cost metrics for a date range
	GetSummary(ctx context.Context, from, to time.Time) (*AICostSummary, error)

	// GetDailyStats retrieves daily AI usage statistics
	GetDailyStats(ctx context.Context, from, to time.Time) ([]*AICostDailyStats, error)

	// GetByOperation retrieves AI cost breakdown by operation type
	GetByOperation(ctx context.Context, from, to time.Time) ([]*AICostByOperation, error)

	// GetByUserSummary retrieves AI cost breakdown by user
	GetByUserSummary(ctx context.Context, from, to time.Time, limit int) ([]*AICostByUser, error)
}

// PricingRepository defines operations for pricing configuration
type PricingRepository interface {
	// Create creates a new pricing config
	Create(ctx context.Context, config *PricingConfig) error

	// Update updates an existing pricing config
	Update(ctx context.Context, config *PricingConfig) error

	// GetByProviderAndModel retrieves active pricing for a specific model
	GetByProviderAndModel(ctx context.Context, provider, model string) (*PricingConfig, error)

	// GetAll retrieves all pricing configs
	GetAll(ctx context.Context) ([]*PricingConfig, error)

	// Deactivate deactivates a pricing config
	Deactivate(ctx context.Context, provider, model string) error
}

// PolicyRepository defines operations for policy documents
type PolicyRepository interface {
	// GetByKey retrieves a policy by its unique key
	GetByKey(ctx context.Context, key string) (*Policy, error)
}

// InteractionLogRepository defines operations for logging user-AI interactions
type InteractionLogRepository interface {
	// Create creates a new interaction log entry
	Create(ctx context.Context, log *InteractionLog) error
}
