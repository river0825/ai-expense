package sqlite

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/riverlin/aiexpense/internal/adapter/repository/migrations"
)

// OpenDB opens a SQLite database connection and runs migrations
func OpenDB(dbPath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Configure connection pooling for performance
	// MaxOpenConns: Allow up to 25 concurrent database connections
	db.SetMaxOpenConns(25)
	// MaxIdleConns: Keep up to 5 idle connections for reuse
	db.SetMaxIdleConns(5)
	// ConnMaxLifetime: Recycle connections every 5 minutes to avoid stale connections
	db.SetConnMaxLifetime(5 * time.Minute)

	// Apply SQLite optimizations via pragmas
	if err := optimizeSQLite(db); err != nil {
		return nil, fmt.Errorf("failed to optimize SQLite: %w", err)
	}

	// Run migrations using golang-migrate
	if err := migrations.RunMigrations(db, "sqlite3"); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return db, nil
}

// optimizeSQLite applies performance optimization settings
func optimizeSQLite(db *sql.DB) error {
	pragmas := []string{
		"PRAGMA journal_mode = WAL",   // Write-Ahead Logging for better concurrency
		"PRAGMA synchronous = NORMAL", // Balance between safety and speed
		"PRAGMA cache_size = 10000",   // Increase cache size (pages)
		"PRAGMA temp_store = MEMORY",  // Use memory for temporary operations
		"PRAGMA foreign_keys = ON",    // Enable foreign key constraints
		"PRAGMA query_only = FALSE",   // Allow writes
		"PRAGMA busy_timeout = 5000",  // 5 second timeout for busy database
	}

	for _, pragma := range pragmas {
		if _, err := db.Exec(pragma); err != nil {
			return fmt.Errorf("failed to execute pragma %q: %w", pragma, err)
		}
	}

	return nil
}

