# Specification: Database Migration System

**Capability ID**: `database-migrations`

**Status**: 🔄 In Review

**Related Capabilities**: `supabase-integration`

---

## Purpose

Establish a versioned, audited database migration system for both SQLite (development) and PostgreSQL/Supabase (production and local testing) that tracks applied migrations, prevents re-runs, and enables safe rollbacks.

---

## ADDED Requirements

### Requirement: Automatic Migration Execution
**ID**: `DB_MIGRATION_AUTO_EXEC`

The application MUST automatically execute pending database migrations on startup before serving requests.

#### Scenario: New Database Schema

Given a fresh database connection (no existing schema)
When the application starts
Then golang-migrate initializes the `schema_migrations` table
And all pending migrations are applied in version order
And the application logs each migration applied
And the application successfully starts after migrations complete

#### Scenario: Existing Schema with Pending Migrations

Given a database with schema_migrations table containing versions 1-3
When the application starts with new migrations 4-5 available
Then migrations 4-5 are applied
And migrations 1-3 are not re-applied
And the application logs "Applying migration 4..." then "Applying migration 5..."
And the application successfully starts

#### Scenario: Idempotent Re-runs

Given a database with all migrations applied (versions 1-4 recorded)
When the application starts again
Then no migrations are applied (identical schema)
And the application logs "No migrations to apply" or similar
And the application starts without errors
And repeated restarts produce identical behavior

---

### Requirement: Version Tracking
**ID**: `DB_MIGRATION_VERSION_TRACKING`

The application MUST track applied migration versions in a persistent table to prevent duplicate execution and enable audit trails.

#### Scenario: Schema Migrations Table Creation

Given any database connection (SQLite or PostgreSQL)
When the first migration runs
Then a `schema_migrations` table is automatically created
And the table contains at minimum: version (integer), dirty (boolean), applied_at (timestamp)
And subsequent migrations record their version in this table

#### Scenario: Version History Audit

Given a database with multiple applied migrations
When querying the `schema_migrations` table
Then each applied migration is listed with:
  - Version number
  - Applied timestamp
  - Dirty flag (false if clean)
And the history can be used to trace schema evolution

---

### Requirement: Rollback Support
**ID**: `DB_MIGRATION_ROLLBACK`

The application MUST support rollback migrations (`.down.sql` files) for every forward migration to enable emergency schema reversals.

#### Scenario: Rollback File Existence

Given a migration `003_create_payments.up.sql`
When migrations are present in the migrations directory
Then a corresponding `003_create_payments.down.sql` MUST exist
And the down migration reverses all schema changes from the up migration

#### Scenario: Rollback Capability

Given migrations 1-4 applied to a database
When a migration tool (e.g., `migrate` CLI) executes a rollback
Then the `.down.sql` for the latest migration is executed
And the `schema_migrations` table is updated (version removed)
And the schema returns to the previous state
And no data is lost during the rollback (down migration handles data carefully)

---

### Requirement: Error Handling
**ID**: `DB_MIGRATION_ERROR_HANDLING`

The application MUST handle migration failures gracefully and provide clear error messages.

#### Scenario: SQL Syntax Error in Migration

Given a migration with invalid SQL (e.g., missing semicolon, wrong table name)
When the application attempts to apply it
Then the migration fails with a clear error message
And the application logs the SQL error
And the `schema_migrations` table marks the migration as dirty (if applicable)
And the application does NOT start (fail-safe)

#### Scenario: Dirty State Detection

Given a database with a dirty flag set (interrupted previous migration)
When the application attempts to start
Then golang-migrate detects the dirty state
And the application logs a clear message: "Database is in dirty state, manual intervention required"
And the application does NOT start
And the operator must manually recover (rollback and re-apply)

#### Scenario: Missing Migration File

Given a database with recorded version 5
But migration file `005_*.sql` does not exist in the migrations directory
When the application starts
Then golang-migrate logs a warning or error
And the application either gracefully skips or fails with clear messaging
And recovery steps are documented

---

### Requirement: Multi-Database Support
**ID**: `DB_MIGRATION_MULTI_DB_SUPPORT`

The application MUST support migration execution for both SQLite and PostgreSQL/Supabase with the same migration files and versioning system.

#### Scenario: SQLite Migration Execution

Given a SQLite database connection
When the application starts
Then golang-migrate executes migrations using the SQLite driver
And the `schema_migrations` table is created in SQLite with appropriate data types
And all migrations apply successfully to SQLite schema

#### Scenario: PostgreSQL Migration Execution

Given a PostgreSQL database connection (Supabase or local)
When the application starts
Then golang-migrate executes migrations using the PostgreSQL driver
And the `schema_migrations` table is created in PostgreSQL with appropriate data types
And all migrations apply successfully to PostgreSQL schema

#### Scenario: Migration Consistency Across Databases

Given identical migration files
When applied to both SQLite and PostgreSQL
Then the resulting schema is functionally equivalent
And both databases contain the same tables, columns, and indexes
And both pass the same data access tests

---

### Requirement: Logging and Monitoring
**ID**: `DB_MIGRATION_LOGGING`

The application MUST log migration events for operational visibility and debugging.

#### Scenario: Migration Logs

When the application starts and migrations execute
Then logs include:
  - "Initializing migrations for [database-type]"
  - "Applying migration NNN: [description]" (for each new migration)
  - "No migrations to apply" (if already up-to-date)
  - "Migrations completed successfully" (on success)
  - Full error details (on failure)

#### Scenario: Log Levels

Given various migration scenarios
When logs are generated
Then successful migrations log at INFO level
And skipped migrations (already applied) log at INFO level
And errors log at ERROR level with full stack trace

---

## MODIFIED Requirements

### Requirement: Application Startup Flow
**ID**: `APP_STARTUP_DB_INIT` (MODIFIED)

The application startup MUST include database migration as a critical initialization step.

#### Scenario: Startup Sequence (Updated)

Given the application is starting
Then the sequence MUST be:
  1. Load configuration (DatabaseURL or DatabasePath)
  2. Open database connection
  3. Run migrations (NEW - critical step)
  4. If migrations fail: exit with error (fail-safe)
  5. If migrations succeed: continue to initialize use cases
  6. Start HTTP server

#### Scenario: Startup Failure Recovery

Given a migration failure during startup
When logs show the error
Then the operator can:
  1. Fix the underlying issue (SQL syntax, schema conflict, etc.)
  2. Manually recover if needed (rollback, resolve conflicts)
  3. Restart the application
And the application will retry migrations on next startup

---

## REMOVED Requirements

None. This is an additive change; no existing requirements are removed.

---

## Cross-References

- **Specification**: `supabase-integration` - Handles local Supabase setup for testing
- **File**: `internal/adapter/repository/migrations/runner.go` - Migration execution logic
- **File**: `internal/adapter/repository/sqlite/db.go` - SQLite adapter integration
- **File**: `internal/adapter/repository/postgresql/db.go` - PostgreSQL adapter integration
- **Directory**: `migrations/` - Migration SQL files (001-NNN_*.{up,down}.sql)

---

## Implementation Notes

- **Tool**: golang-migrate (https://github.com/golang-migrate/migrate)
- **Version Format**: `NNN_description.{up,down}.sql` (e.g., `001_init_schema.up.sql`)
- **Idempotency**: Achieved via version tracking; migrations are applied only once per version
- **Rollback Safety**: Down migrations must carefully handle data (no irreversible deletes unless intentional)

---

## Validation Checklist

After implementation, verify:

- [ ] `go test ./... -v` - All tests pass
- [ ] SQLite migrations work locally with cold start (no prior schema)
- [ ] PostgreSQL migrations work with Supabase (local and cloud)
- [ ] Re-running app doesn't re-apply migrations (idempotent)
- [ ] Rollback `.down.sql` files work correctly
- [ ] `schema_migrations` table accurately tracks applied versions
- [ ] Error messages are clear and actionable
- [ ] Logs show migration events appropriately
