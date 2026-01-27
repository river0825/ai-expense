package migrations

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// RunMigrations runs pending migrations for the given database
// dbType should be "sqlite3" or "postgres"
func RunMigrations(db *sql.DB, dbType string) error {
	var m *migrate.Migrate
	var err error

	path := "file://internal/adapter/repository/migrations/files"
	if _, err := os.Stat("/app/migrations"); err == nil {
		path = "file:///app/migrations"
	}

	// Initialize migrator with appropriate driver based on database type
	switch dbType {
	case "sqlite3":
		driver, err := sqlite3.WithInstance(db, &sqlite3.Config{})
		if err != nil {
			return fmt.Errorf("failed to create SQLite migrate driver: %w", err)
		}
		m, err = migrate.NewWithDatabaseInstance(
			path,
			"sqlite3",
			driver,
		)
		if err != nil {
			return fmt.Errorf("failed to initialize migrator for SQLite: %w", err)
		}

	case "postgres":
		driver, err := postgres.WithInstance(db, &postgres.Config{})
		if err != nil {
			return fmt.Errorf("failed to create PostgreSQL migrate driver: %w", err)
		}
		m, err = migrate.NewWithDatabaseInstance(
			path,
			"postgres",
			driver,
		)
		if err != nil {
			return fmt.Errorf("failed to initialize migrator for PostgreSQL: %w", err)
		}

	default:
		return fmt.Errorf("unsupported database type: %s", dbType)
	}

	// Run migrations (idempotent - only applies new versions)
	err = m.Up()
	if err != nil {
		if err.Error() == "no change" {
			log.Println("No new migrations to apply")
			return nil
		}
		if err.Error() == "dirty" {
			return fmt.Errorf("database is in dirty state (interrupted migration). Manual intervention required: check schema_migrations table and rollback if needed")
		}
		return fmt.Errorf("migration failed: %w", err)
	}

	log.Println("Migrations applied successfully")
	return nil
}
