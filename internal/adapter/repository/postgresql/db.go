package postgresql

import (
	"database/sql"
	_ "github.com/lib/pq"
)

// OpenDB opens a PostgreSQL database connection
func OpenDB(databaseURL string) (*sql.DB, error) {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, err
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}
