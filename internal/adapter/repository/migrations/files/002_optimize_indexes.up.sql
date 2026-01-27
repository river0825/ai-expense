-- Phase 19: Performance Optimization - Additional Indexes and Optimizations

-- Add index on created_at for time-based queries (metrics, archiving)
CREATE INDEX IF NOT EXISTS idx_expenses_created_at ON expenses(created_at DESC);

-- Add index on user_id alone for faster user lookups
CREATE INDEX IF NOT EXISTS idx_expenses_user ON expenses(user_id);

-- Add composite index for user + created_at (recent expenses)
CREATE INDEX IF NOT EXISTS idx_expenses_user_created ON expenses(user_id, created_at DESC);

-- Add index on category_id for category filtering
CREATE INDEX IF NOT EXISTS idx_expenses_category ON expenses(category_id);

-- Add index on expense_date for date-range queries
CREATE INDEX IF NOT EXISTS idx_expenses_date ON expenses(expense_date DESC);

-- Add composite index for user + date for periodic reports
CREATE INDEX IF NOT EXISTS idx_expenses_user_period ON expenses(user_id, expense_date);

-- Add index on category keywords for keyword-based lookups
CREATE INDEX IF NOT EXISTS idx_keywords_priority ON category_keywords(priority DESC);

-- Add index on created_at for category operations
CREATE INDEX IF NOT EXISTS idx_categories_created ON categories(created_at DESC);

-- Optimize foreign key lookups
CREATE INDEX IF NOT EXISTS idx_users_created ON users(created_at DESC);

-- PRAGMA settings for SQLite optimization (note: these would be set per connection in code)
-- PRAGMA journal_mode = WAL;           -- Write-Ahead Logging for better concurrency
-- PRAGMA synchronous = NORMAL;         -- Balance between safety and speed
-- PRAGMA cache_size = 10000;           -- Increase cache size
-- PRAGMA temp_store = MEMORY;          -- Use memory for temporary operations
-- PRAGMA query_only = FALSE;           -- Allow writes (default)
