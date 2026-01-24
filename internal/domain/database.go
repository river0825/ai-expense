package domain

import (
	"database/sql"
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
