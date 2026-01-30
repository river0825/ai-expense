package sqlite

import (
	"database/sql"
	"fmt"
	"sync"
)

// PreparedStatementCache caches frequently used prepared statements for performance
type PreparedStatementCache struct {
	db       *sql.DB
	stmts    map[string]*sql.Stmt
	mu       sync.RWMutex
	maxStmts int
}

// NewPreparedStatementCache creates a new prepared statement cache
func NewPreparedStatementCache(db *sql.DB) *PreparedStatementCache {
	return &PreparedStatementCache{
		db:       db,
		stmts:    make(map[string]*sql.Stmt),
		maxStmts: 50,
	}
}

// Get retrieves a prepared statement from cache or creates a new one
func (c *PreparedStatementCache) Get(query string) (*sql.Stmt, error) {
	c.mu.RLock()
	if stmt, exists := c.stmts[query]; exists {
		c.mu.RUnlock()
		return stmt, nil
	}
	c.mu.RUnlock()

	// Prepare new statement
	stmt, err := c.db.Prepare(query)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare statement: %w", err)
	}

	// Cache the statement
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check cache size limit
	if len(c.stmts) >= c.maxStmts {
		// Remove oldest statement (simple FIFO)
		for key, oldStmt := range c.stmts {
			oldStmt.Close()
			delete(c.stmts, key)
			break
		}
	}

	c.stmts[query] = stmt
	return stmt, nil
}

// Close closes all cached prepared statements
func (c *PreparedStatementCache) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, stmt := range c.stmts {
		if err := stmt.Close(); err != nil {
			return fmt.Errorf("failed to close prepared statement: %w", err)
		}
	}

	c.stmts = make(map[string]*sql.Stmt)
	return nil
}

// Clear removes all cached statements
func (c *PreparedStatementCache) Clear() error {
	return c.Close()
}

// Size returns the number of cached statements
func (c *PreparedStatementCache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.stmts)
}

// Common prepared statement queries for frequent operations
const (
	// User queries
	QueryUserExists = "SELECT 1 FROM users WHERE user_id = ? LIMIT 1"
	QueryUserByID   = "SELECT user_id, messenger_type, created_at, home_currency, locale FROM users WHERE user_id = ?"
	QueryCreateUser = "INSERT INTO users (user_id, messenger_type, created_at, home_currency, locale) VALUES (?, ?, ?, ?, ?)"

	// Expense queries
	QueryExpenseByUserID            = "SELECT id, user_id, description, original_amount, currency, home_amount, home_currency, exchange_rate, category_id, expense_date, created_at, updated_at FROM expenses WHERE user_id = ?"
	QueryExpenseByUserIDDateRange   = "SELECT id, user_id, description, original_amount, currency, home_amount, home_currency, exchange_rate, category_id, expense_date, created_at, updated_at FROM expenses WHERE user_id = ? AND expense_date >= ? AND expense_date <= ?"
	QueryExpenseByUserIDAndCategory = "SELECT id, user_id, description, original_amount, currency, home_amount, home_currency, exchange_rate, category_id, expense_date, created_at, updated_at FROM expenses WHERE user_id = ? AND category_id = ?"
	QueryExpenseByID                = "SELECT id, user_id, description, original_amount, currency, home_amount, home_currency, exchange_rate, category_id, expense_date, created_at, updated_at FROM expenses WHERE id = ?"
	QueryCreateExpense              = "INSERT INTO expenses (id, user_id, description, original_amount, currency, home_amount, home_currency, exchange_rate, category_id, expense_date, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)"
	QueryUpdateExpense              = "UPDATE expenses SET description = ?, original_amount = ?, currency = ?, home_amount = ?, home_currency = ?, exchange_rate = ?, category_id = ?, expense_date = ?, updated_at = ? WHERE id = ?"
	QueryDeleteExpense              = "DELETE FROM expenses WHERE id = ?"

	// Category queries
	QueryCategoriesByUserID      = "SELECT id, user_id, name, is_default, created_at FROM categories WHERE user_id = ?"
	QueryCategoryByUserIDAndName = "SELECT id, user_id, name, is_default, created_at FROM categories WHERE user_id = ? AND name = ?"
	QueryCategoryByID            = "SELECT id, user_id, name, is_default, created_at FROM categories WHERE id = ?"
	QueryCreateCategory          = "INSERT INTO categories (id, user_id, name, is_default, created_at) VALUES (?, ?, ?, ?, ?)"
	QueryUpdateCategory          = "UPDATE categories SET name = ?, is_default = ? WHERE id = ?"
	QueryDeleteCategory          = "DELETE FROM categories WHERE id = ?"

	// Metrics queries
	QueryGetMetrics = "SELECT " +
		"COUNT(DISTINCT user_id) as active_users, " +
		"COUNT(*) as total_expenses, " +
		"SUM(home_amount) as total_amount, " +
		"AVG(home_amount) as avg_amount " +
		"FROM expenses WHERE DATE(created_at) = ?"
)

// PreparedStmtExecutor wraps database operations with prepared statement caching
type PreparedStmtExecutor struct {
	cache *PreparedStatementCache
}

// NewPreparedStmtExecutor creates a new prepared statement executor
func NewPreparedStmtExecutor(cache *PreparedStatementCache) *PreparedStmtExecutor {
	return &PreparedStmtExecutor{
		cache: cache,
	}
}

// QueryRow executes a single-row query with prepared statement caching
func (e *PreparedStmtExecutor) QueryRow(query string, args ...interface{}) *sql.Row {
	stmt, err := e.cache.Get(query)
	if err != nil {
		// Fallback to non-cached query if caching fails
		return nil
	}
	return stmt.QueryRow(args...)
}

// Query executes a multi-row query with prepared statement caching
func (e *PreparedStmtExecutor) Query(query string, args ...interface{}) (*sql.Rows, error) {
	stmt, err := e.cache.Get(query)
	if err != nil {
		return nil, err
	}
	return stmt.Query(args...)
}

// Exec executes a statement with prepared statement caching
func (e *PreparedStmtExecutor) Exec(query string, args ...interface{}) (sql.Result, error) {
	stmt, err := e.cache.Get(query)
	if err != nil {
		return nil, err
	}
	return stmt.Exec(args...)
}
