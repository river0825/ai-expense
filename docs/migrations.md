# Database Migrations Guide

This project uses [golang-migrate](https://github.com/golang-migrate/migrate) for versioned, idempotent database migrations. Migrations are applied automatically on application startup for both SQLite and PostgreSQL databases.

## Table of Contents

- [Overview](#overview)
- [How Migrations Work](#how-migrations-work)
- [Creating New Migrations](#creating-new-migrations)
- [Testing Migrations Locally](#testing-migrations-locally)
- [Migration Status](#migration-status)
- [Rollback Procedures](#rollback-procedures)
- [Troubleshooting](#troubleshooting)

---

## Overview

Migrations are versioned SQL files that define schema changes. Each migration has:
- **Up migration** (`.up.sql`): Applies the schema change
- **Down migration** (`.down.sql`): Reverts the schema change

**Key Features:**
- ✅ Automatic version tracking in `schema_migrations` table
- ✅ Idempotent: Running migrations multiple times is safe
- ✅ Works with both SQLite (development) and PostgreSQL (production)
- ✅ Rollback capability via down migrations
- ✅ Runs automatically on application startup

---

## How Migrations Work

### Migration File Structure

```
migrations/
├── 001_init_schema.up.sql
├── 001_init_schema.down.sql
├── 002_optimize_indexes.up.sql
├── 002_optimize_indexes.down.sql
├── 003_create_ai_cost_logs.up.sql
├── 003_create_ai_cost_logs.down.sql
├── 004_create_policies_table.up.sql
└── 004_create_policies_table.down.sql
```

### Version Tracking

When migrations run, golang-migrate creates a `schema_migrations` table to track applied versions:

```sql
CREATE TABLE schema_migrations (
  version BIGINT PRIMARY KEY,
  dirty BOOLEAN NOT NULL DEFAULT false
);
```

Each row represents an applied migration. This ensures migrations only run once, even if the application restarts.

### Application Startup Flow

1. Application starts
2. Detects database type (SQLite or PostgreSQL)
3. Calls `migrations.RunMigrations(db, dbType)`
4. golang-migrate checks `schema_migrations` table
5. Applies only new migrations (those not in the table)
6. Updates `schema_migrations` with new version numbers
7. Application continues with guaranteed schema consistency

---

## Creating New Migrations

### Step 1: Create Migration Files

Create matching `.up.sql` and `.down.sql` files with the next version number:

```bash
# Naming convention: NNN_description.{up,down}.sql
touch migrations/005_add_user_preferences.up.sql
touch migrations/005_add_user_preferences.down.sql
```

### Step 2: Write the Up Migration

**File**: `migrations/005_add_user_preferences.up.sql`

```sql
-- Create preferences table for user settings
CREATE TABLE IF NOT EXISTS user_preferences (
  id TEXT PRIMARY KEY,
  user_id TEXT NOT NULL,
  theme TEXT DEFAULT 'light',
  language TEXT DEFAULT 'en',
  notifications_enabled BOOLEAN DEFAULT TRUE,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (user_id) REFERENCES users(user_id)
);

-- Create index for user lookups
CREATE INDEX IF NOT EXISTS idx_user_preferences_user ON user_preferences(user_id);
```

### Step 3: Write the Down Migration

**File**: `migrations/005_add_user_preferences.down.sql`

The down migration should reverse the up migration exactly:

```sql
-- Drop index
DROP INDEX IF EXISTS idx_user_preferences_user;

-- Drop table
DROP TABLE IF EXISTS user_preferences;
```

### Step 4: Test the Migration

See [Testing Migrations Locally](#testing-migrations-locally) section below.

### Best Practices

1. **Keep migrations small**: One feature per migration
2. **Make them idempotent**: Use `IF NOT EXISTS` (SQLite) and `IF NOT EXISTS` (PostgreSQL)
3. **Reverse exactly**: Down migrations must undo the up migration perfectly
4. **Test both directions**: Verify up and down migrations work
5. **Consider indexes**: Add indexes in separate migrations after tables exist
6. **Foreign keys**: Create referenced tables before tables that reference them
7. **No breaking changes**: Migrations should be additive when possible

---

## Testing Migrations Locally

### Option 1: SQLite (Simplest)

SQLite migrations are tested automatically when you run the application:

```bash
# Delete existing database to test from scratch
rm -f aiexpense.db

# Run application (migrations apply automatically)
go run ./cmd/server/main.go

# Verify schema_migrations table
sqlite3 aiexpense.db "SELECT * FROM schema_migrations;"

# Verify tables exist
sqlite3 aiexpense.db ".tables"
```

### Option 2: PostgreSQL with Local Supabase

For realistic PostgreSQL testing:

```bash
# 1. Install Supabase CLI
brew install supabase/tap/supabase

# 2. Start local Supabase instance
supabase start
# Wait for: "Local development started"

# 3. Get connection details
supabase status
# Copy the DATABASE_URL from output

# 4. Set environment and run app
export DATABASE_URL="postgresql://postgres:postgres@localhost:54322/postgres"
go run ./cmd/server/main.go

# 5. Verify migrations
psql "$DATABASE_URL" -c "SELECT * FROM schema_migrations;"
psql "$DATABASE_URL" -c "\dt"  # List all tables

# 6. Stop Supabase when done
supabase stop
```

### Testing a New Migration

```bash
# 1. Create your migration files (005_example.up.sql and .down.sql)

# 2. Test the up migration
go run ./cmd/server/main.go

# 3. Verify the schema changed
sqlite3 aiexpense.db ".schema table_name"

# 4. Test the down migration
# Using migrate CLI directly (if installed):
brew install migrate
migrate -path ./migrations -database "sqlite3://aiexpense.db" down 1

# 5. Verify the table was dropped
sqlite3 aiexpense.db ".tables"

# 6. Test the up migration again
migrate -path ./migrations -database "sqlite3://aiexpense.db" up 1

# 7. Verify schema consistency
go run ./cmd/server/main.go
```

---

## Migration Status

### Check Applied Migrations

**SQLite:**
```bash
sqlite3 aiexpense.db "SELECT version, dirty FROM schema_migrations ORDER BY version;"
```

**PostgreSQL:**
```bash
psql "$DATABASE_URL" -c "SELECT version, dirty FROM schema_migrations ORDER BY version;"
```

### Get Migration Version Number

**Using migrate CLI:**
```bash
migrate -path ./migrations -database "sqlite3://aiexpense.db" version
# Output: 4 (current version)
```

### List All Tables

**SQLite:**
```bash
sqlite3 aiexpense.db ".tables"
```

**PostgreSQL:**
```bash
psql "$DATABASE_URL" -c "\dt public.*"
```

---

## Rollback Procedures

### Automatic Rollback (Application Failure)

If a migration fails during application startup:
1. Application logs error and exits
2. Database remains in its previous state
3. Fix the migration `.up.sql` file
4. Restart application to retry

### Manual Rollback (Using migrate CLI)

If you need to rollback after migrations have been applied:

```bash
# Install migrate CLI
brew install migrate

# Rollback one migration
migrate -path ./migrations -database "sqlite3://aiexpense.db" down 1

# Rollback to specific version
migrate -path ./migrations -database "sqlite3://aiexpense.db" goto 3

# Rollback all migrations (dangerous!)
migrate -path ./migrations -database "sqlite3://aiexpense.db" down
```

### Emergency Rollback (Dirty State)

If a migration leaves the database in "dirty" state (interrupted mid-migration):

```bash
# Check dirty flag
sqlite3 aiexpense.db "SELECT version, dirty FROM schema_migrations WHERE dirty = 1;"

# Force mark as clean (caution: only if you've manually verified schema)
sqlite3 aiexpense.db "UPDATE schema_migrations SET dirty = 0 WHERE version = 4;"

# Then rollback the dirty migration
migrate -path ./migrations -database "sqlite3://aiexpense.db" down 1
```

---

## Troubleshooting

### Problem: "Database is in dirty state"

**Cause**: A migration was interrupted (e.g., app crash during migration).

**Solution**:
1. Check which migration failed: `sqlite3 aiexpense.db "SELECT * FROM schema_migrations WHERE dirty = 1;"`
2. Manually verify the schema to see what was partially applied
3. Either:
   - Fix the `.down.sql` and rollback: `migrate down 1`
   - Or manually clean up the schema and mark clean: `UPDATE schema_migrations SET dirty = 0 WHERE version = X;`

### Problem: "Migration step X already applied"

**Cause**: You're trying to run a migration that's already in `schema_migrations`.

**Solution**: This is expected behavior! Migrations are idempotent. Just restart the application.

### Problem: "File not found: migrations/..."

**Cause**: golang-migrate can't find the migrations directory.

**Solution**: Ensure:
1. Migrations directory exists: `ls -la migrations/`
2. You're running app from project root: `pwd` should show `/path/to/aiexpense`
3. Migration files have correct naming: `NNN_description.{up,down}.sql`

### Problem: "Foreign key constraint violation on down migration"

**Cause**: Down migration is dropping a table that other tables depend on.

**Solution**: Drop dependent tables first, then the referenced table:

```sql
-- Correct order (down migration)
DROP TABLE IF EXISTS expenses;        -- depends on categories
DROP TABLE IF EXISTS categories;      -- depends on users
DROP TABLE IF EXISTS users;           -- no dependencies
```

### Problem: "No such table: schema_migrations"

**Cause**: First time running migrations on a new database.

**Solution**: This is expected! golang-migrate creates the table automatically on first run. Just run the application again.

---

## References

- [golang-migrate Documentation](https://github.com/golang-migrate/migrate)
- [golang-migrate CLI Reference](https://github.com/golang-migrate/migrate/tree/master/cmd/migrate)
- [SQLite Documentation](https://www.sqlite.org/docs.html)
- [PostgreSQL Documentation](https://www.postgresql.org/docs/)
- [Supabase CLI Documentation](https://supabase.com/docs/guides/cli)
