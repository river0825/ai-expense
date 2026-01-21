package repository

import (
	"context"
	"database/sql"
	"github.com/lib/pq"
	"github.com/riverlin/aiexpense/internal/domain"
)

type DatabaseType string

const (
	DatabaseTypeSQLite     DatabaseType = "sqlite"
	DatabaseTypePostgreSQL DatabaseType = "postgresql"
)

type Database struct {
	Type     DatabaseType
	SQLite   *sql.DB
	Postgres *sql.DB
}

type DBConfig interface {
	GetDB() *Database
	Close() error
}
