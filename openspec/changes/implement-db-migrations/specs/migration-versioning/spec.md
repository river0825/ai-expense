# Specification: Database Migration Versioning and Execution

**Capability**: migration-versioning

**Status**: Active

---

## ADDED Requirements

### Requirement: Versioned Migration Tracking

The system SHALL load all migration files from the `./migrations/` directory and execute only those not previously applied, as tracked in the `schema_migrations` table.

**Details:**
1. The system SHALL load all migration files from `./migrations/` directory
2. The system SHALL check the `schema_migrations` table for previously applied versions
3. The system SHALL apply only new migrations (versions not in the table)
4. The system SHALL record each applied migration with timestamp

#### Scenario: First Application Startup
```gherkin
Given a fresh PostgreSQL database
When the application starts
Then the app creates schema_migrations table
And executes all .up.sql files in order (001, 002, 003, 004)
And records each migration in schema_migrations
```

#### Scenario: Subsequent Startup (Idempotent)
```gherkin
Given a PostgreSQL database with 3 migrations already applied
When the application starts
Then the app reads schema_migrations table
And skips migrations 001, 002, 003
And executes only the new migration (004)
And records migration 004
```

#### Scenario: No Pending Migrations
```gherkin
Given a PostgreSQL database with all migrations applied
When the application starts
Then the app reads schema_migrations
And finds no pending migrations
And logs "No new migrations to apply"
And continues to business logic
```

---

### Requirement: Automatic Migration on Startup

The system SHALL automatically execute pending migrations during initialization and block startup if migrations fail.

#### Scenario: Successful Migration
```gherkin
Given a database with pending migrations
When the app initializes the database adapter
Then migrations are applied automatically
And app continues to business logic
```

#### Scenario: Migration Failure (Fail Loudly)
```gherkin
Given a database with a corrupt or invalid migration
When the app initializes the database adapter
Then the migration fails
And the app logs the error
And the app exits with a non-zero status
And NO partial migration is left in schema_migrations
```

#### Scenario: Dirty State Detection
```gherkin
Given a database with a failed migration (partial application)
When the app starts
Then the app detects the dirty flag
And logs a critical error with rollback instructions
And refuses to proceed
```

---

### Requirement: Down Migrations for Rollback

Each up migration SHALL have a corresponding down migration to enable rollbacks.

#### Scenario: Down Migration Exists
```gherkin
Given all .up.sql files have .down.sql counterparts
When a developer runs migrate -path ./migrations -database $DB down
Then each down migration undoes the corresponding up migration
And schema_migrations table is updated
```

#### Scenario: Missing Down Migration
```gherkin
Given a migration with no .down.sql file
When the app attempts to roll back
Then the app logs a warning
And does NOT apply an incomplete rollback
```

---

### Requirement: Support Both SQLite and PostgreSQL

The system SHALL support migrations for both SQLite (local development) and PostgreSQL (production/Supabase) using the same migration files.

#### Scenario: SQLite Migration
```gherkin
Given DATABASE_URL is not set (SQLite mode)
When the app starts
Then it initializes golang-migrate with sqlite3 driver
And applies migrations to ./data/app.db
And creates schema_migrations table in SQLite
```

#### Scenario: PostgreSQL Migration (Production)
```gherkin
Given DATABASE_URL is set to postgresql://user:pass@supabase-host/db
When the app starts
Then it initializes golang-migrate with postgres driver
And applies migrations to the remote PostgreSQL instance
And creates schema_migrations table in PostgreSQL
```

#### Scenario: PostgreSQL Migration (Local Supabase)
```gherkin
Given local Supabase is running (supabase start)
And DATABASE_URL is set to postgresql://postgres:postgres@localhost:54322/postgres
When the app starts
Then it initializes golang-migrate with postgres driver
And applies migrations to the local PostgreSQL instance
And matches schema applied to production Supabase
```

---

## MODIFIED Requirements

### Requirement: SQL File Naming Convention

Migration files SHALL follow the naming pattern: `NNN_description.{up,down}.sql` where NNN is a 3-digit version number and description is a human-readable summary.

#### Scenario: Valid Migration File Names
```gherkin
Given migration files named:
  - 001_init_schema.up.sql
  - 001_init_schema.down.sql
  - 002_optimize_indexes.up.sql
  - 002_optimize_indexes.down.sql
When the app scans ./migrations
Then it recognizes all four files correctly
And applies them in numeric order
```

#### Scenario: Invalid File Names (Ignored)
```gherkin
Given files named:
  - migrations.sql (no version)
  - 001_init (no .up or .down)
  - 001.up.sql (no description)
When the app scans ./migrations
Then it logs warnings for invalid names
And does NOT execute these files
```

---

## REMOVED Requirements

### Previous Requirement: SQLite-Only Inline Migrations

The old pattern of embedding `CREATE TABLE IF NOT EXISTS` directly in SQLite OpenDB() initialization is replaced by golang-migrate versioning.

#### Impact
- ❌ `runMigrations()` in sqlite/db.go is replaced
- ❌ Hardcoded migration file list is replaced by directory scanning
- ✅ New golang-migrate integration provides the same functionality + versioning

---

## Cross-Capability References

### Related Capabilities
- **database-initialization**: Depends on this spec; migrations run during DB initialization
- **supabase-deployment**: Depends on this spec; migrations are pre-applied before app deployment
- **error-handling**: Defines error scenarios (dirty state, missing down migrations, rollback failure)

### Dependencies
- golang-migrate library (github.com/golang-migrate/migrate/v4)
- Database drivers: postgres, sqlite3
- File I/O for migration discovery

