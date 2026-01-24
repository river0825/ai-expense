# Design: Database Migration and Supabase Integration

## Overview

This document outlines the technical architecture for integrating `golang-migrate` and setting up local Supabase testing for the aiexpense application.

---

## Architecture

### Component Interaction Diagram

```
┌─────────────────────────────────────────────────────────────┐
│                  Application Startup                        │
└─────────────────────┬───────────────────────────────────────┘
                      │
                      ▼
        ┌─────────────────────────────┐
        │  Load Configuration         │
        │  (DATABASE_URL or PATH)     │
        └─────────┬───────────────────┘
                  │
        ┌─────────▼─────────┐
        │ Detect DB Type    │
        └────┬──────────┬───┘
             │          │
    ┌────────▼─┐   ┌────▼──────┐
    │ SQLite   │   │ PostgreSQL │
    │ Adapter  │   │  Adapter   │
    └────┬─────┘   └────┬───────┘
         │              │
         └──────┬───────┘
                │
         ┌──────▼────────────────┐
         │ golang-migrate        │
         │ (Version Tracking)    │
         └──────┬────────────────┘
                │
         ┌──────▼────────────────┐
         │ Run Pending           │
         │ Migrations            │
         └──────┬────────────────┘
                │
                ├─ Success: Continue startup
                └─ Failure: Exit with error
```

---

## Implementation Strategy

### 1. golang-migrate Integration

#### Dependency Addition

```bash
go get -u github.com/golang-migrate/migrate/v4
go get -u github.com/golang-migrate/migrate/v4/database/sqlite3
go get -u github.com/golang-migrate/migrate/v4/database/postgres
go get -u github.com/golang-migrate/migrate/v4/source/file
```

#### Database Adapter Changes

**File**: `internal/adapter/repository/sqlite/db.go`

Replace manual migration logic:
```go
// OLD (current)
func runMigrations(db *sql.DB) error {
  migrationFiles := []string{...}
  for _, filepath := range migrationFiles {
    schemaSQL, err := os.ReadFile(filepath)
    if err != nil { return err }
    if _, err := db.Exec(string(schemaSQL)); err != nil {
      return err
    }
  }
  return nil
}

// NEW (with golang-migrate)
func runMigrations(db *sql.DB) error {
  driver, err := sqlite3.WithInstance(db, &sqlite3.Config{})
  if err != nil { return err }

  m, err := migrate.NewWithDatabaseInstance(
    "file://migrations",
    "sqlite3",
    driver,
  )
  if err != nil { return err }

  if err := m.Up(); err != nil && err != migrate.ErrNoChange {
    return err
  }
  return nil
}
```

**File**: `internal/adapter/repository/postgresql/db.go`

Add migration runner (new):
```go
func runMigrations(db *sql.DB) error {
  driver, err := postgres.WithInstance(db, &postgres.Config{})
  if err != nil { return err }

  m, err := migrate.NewWithDatabaseInstance(
    "file://migrations",
    "postgres",
    driver,
  )
  if err != nil { return err }

  if err := m.Up(); err != nil && err != migrate.ErrNoChange {
    return err
  }
  return nil
}

func OpenDB(databaseURL string) (*sql.DB, error) {
  db, err := sql.Open("postgres", databaseURL)
  if err != nil { return nil, err }

  ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
  defer cancel()
  if err := db.PingContext(ctx); err != nil {
    db.Close()
    return nil, err
  }

  // RUN MIGRATIONS (new)
  if err := runMigrations(db); err != nil {
    db.Close()
    return nil, fmt.Errorf("migration failed: %w", err)
  }

  return db, nil
}
```

---

### 2. Migration File Refactoring

#### Current Files
```
migrations/001_init_schema.up.sql
migrations/002_optimize_indexes.up.sql
migrations/003_create_ai_cost_logs.up.sql
migrations/004_create_policies_table.up.sql
```

#### Additions Needed

Each `.up.sql` needs a corresponding `.down.sql`:

```
migrations/001_init_schema.down.sql
migrations/002_optimize_indexes.down.sql
migrations/003_create_ai_cost_logs.down.sql
migrations/004_create_policies_table.down.sql
```

#### golang-migrate Expectations

Files must be:
- Named: `NNN_description.up.sql` or `NNN_description.down.sql`
- Version format: `NNN` (version number, can be any integer)
- Located in: `file://migrations/` (relative to app binary or config)

#### Migration Tracking Table

golang-migrate creates this automatically:

```sql
CREATE TABLE schema_migrations (
  version BIGINT PRIMARY KEY,
  dirty BOOLEAN NOT NULL DEFAULT false,
  applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

---

### 3. Local Supabase Setup

#### Supabase CLI Commands

```bash
# Install Supabase CLI (macOS)
brew install supabase/tap/supabase

# Link to your Supabase project
supabase link --project-ref <project-ref>

# Start local Supabase instance
supabase start

# Get connection details
supabase status

# Example output:
# DATABASE_URL: postgresql://postgres:postgres@localhost:54322/postgres

# Stop local instance
supabase stop
```

#### Environment Configuration

**File**: `.env.local.example`

```bash
# Local Supabase (for development)
# DATABASE_URL=postgresql://postgres:postgres@localhost:54322/postgres

# Or use SQLite for simpler local dev
# DATABASE_PATH=./aiexpense.db

# Other configuration...
GEMINI_API_KEY=your-key
AI_PROVIDER=gemini
SERVER_PORT=8080
```

#### Developer Workflow

1. Start Supabase: `supabase start`
2. Extract connection: `supabase status`
3. Copy connection string: `postgresql://...`
4. Set in `.env`: `DATABASE_URL=<connection-string>`
5. Run app: `go run ./cmd/server/main.go`
6. Migrations run automatically on startup
7. Stop Supabase: `supabase stop`

---

### 4. Testing Strategy

#### Integration Tests

**File**: `internal/adapter/repository/repository_test.go` (exists)

Extend with migration tests:

```go
func TestSQLiteMigrations(t *testing.T) {
  // Setup temp SQLite DB
  db, err := sqliteRepo.OpenDB(":memory:")
  require.NoError(t, err)
  defer db.Close()

  // Verify schema_migrations table exists
  var exists bool
  err = db.QueryRow("SELECT 1 FROM sqlite_master WHERE type='table' AND name='schema_migrations'").Scan(&exists)
  require.NoError(t, err)
  require.True(t, exists)
}

func TestPostgresMigrations(t *testing.T) {
  // Setup temp Postgres DB (via test container or local)
  // Similar validation...
}
```

#### Manual Testing

```bash
# Test SQLite
go run ./cmd/server/main.go
# Logs should show: "Running migrations..." and "Migration complete"

# Test PostgreSQL (if Supabase running)
export DATABASE_URL="postgresql://postgres:postgres@localhost:54322/postgres"
go run ./cmd/server/main.go
# Logs should show same migration messages
```

---

### 5. Error Handling

#### Scenarios

| Scenario | Behavior | Recovery |
|----------|----------|----------|
| Schema already up-to-date | Log "No migrations to apply", continue | N/A |
| New migration available | Log "Applying migration NNN_desc", apply, continue | N/A |
| Migration fails (SQL error) | Log error, exit with code 1 | Manual intervention + rollback |
| Dirty schema (interrupted migration) | Log "Dirty state detected", exit with code 1 | Run `.down.sql`, then `.up.sql` again |

#### Implementation

```go
func runMigrations(db *sql.DB) error {
  m, err := migrate.NewWithDatabaseInstance(...)
  if err != nil {
    return fmt.Errorf("failed to initialize migrator: %w", err)
  }

  if err := m.Up(); err != nil {
    if err == migrate.ErrNoChange {
      log.Println("No new migrations to apply")
      return nil
    }
    if err == migrate.ErrDirty {
      return fmt.Errorf("database is in dirty state (interrupted migration). Manually rollback and retry")
    }
    return fmt.Errorf("migration failed: %w", err)
  }

  log.Println("Migrations applied successfully")
  return nil
}
```

---

### 6. Deployment to Supabase (Cloud)

#### Option A: Automated (via app startup)

App connects to Supabase PostgreSQL and runs migrations automatically.

```bash
# In Supabase project
export DATABASE_URL="postgresql://user:pass@db.supabase.co/postgres"
# Deploy app
docker run -e DATABASE_URL aiexpense:latest
# Migrations run on startup
```

#### Option B: Manual (via CLI)

For extra safety or debugging:

```bash
# Install migrate CLI
brew install migrate

# Run migrations against Supabase
migrate -path ./migrations -database "$DATABASE_URL" up
```

---

## File Structure After Implementation

```
project-root/
├── migrations/
│   ├── 001_init_schema.up.sql       (exists, unchanged)
│   ├── 001_init_schema.down.sql     (NEW)
│   ├── 002_optimize_indexes.up.sql  (exists, unchanged)
│   ├── 002_optimize_indexes.down.sql (NEW)
│   ├── 003_create_ai_cost_logs.up.sql
│   ├── 003_create_ai_cost_logs.down.sql (NEW)
│   ├── 004_create_policies_table.up.sql
│   └── 004_create_policies_table.down.sql (NEW)
│
├── internal/adapter/repository/
│   ├── sqlite/
│   │   ├── db.go                    (MODIFIED - use golang-migrate)
│   │   └── ...
│   ├── postgresql/
│   │   ├── db.go                    (MODIFIED - add runMigrations)
│   │   └── ...
│   └── ...
│
├── .env.local.example               (NEW - Supabase config template)
├── go.mod                           (MODIFIED - add golang-migrate)
├── go.sum                           (auto-updated)
└── ...
```

---

## Security Considerations

1. **Connection Strings**: Store in environment variables (never commit)
2. **Schema Migrations**: Version controlled, reviewed before apply
3. **Rollback Safety**: Down migrations tested before production use
4. **Audit Trail**: `schema_migrations` table provides version history
5. **Production Access**: Supabase CLI should only run in CI/CD with restricted credentials

---

## Rollback Procedure

If a migration causes issues:

```bash
# Check current migration status
migrate -path ./migrations -database "$DATABASE_URL" version

# Rollback one migration
migrate -path ./migrations -database "$DATABASE_URL" down 1

# Or rollback to specific version
migrate -path ./migrations -database "$DATABASE_URL" goto 3
```

---

## Monitoring and Validation

After each deployment, verify:

```sql
-- Check applied migrations
SELECT * FROM schema_migrations ORDER BY version;

-- Verify schema exists
SELECT table_name FROM information_schema.tables
WHERE table_schema = 'public' ORDER BY table_name;

-- Test data access
SELECT COUNT(*) FROM expenses;
SELECT COUNT(*) FROM categories;
SELECT COUNT(*) FROM ai_cost_logs;
```

---

## References

- golang-migrate: https://github.com/golang-migrate/migrate
- Supabase CLI: https://supabase.com/docs/guides/cli
- PostgreSQL Schema: https://www.postgresql.org/docs/current/ddl.html
