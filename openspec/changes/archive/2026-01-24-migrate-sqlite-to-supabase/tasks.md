# Tasks: Migrate SQLite to Supabase with golang-migrate

This is a step-by-step implementation checklist. Each task is small and verifiable. Complete in order.

---

## Phase 1: golang-migrate Setup

### Task 1.1: Add golang-migrate Dependencies
- [ ] Add golang-migrate main package: `go get -u github.com/golang-migrate/migrate/v4`
- [ ] Add SQLite driver: `go get -u github.com/golang-migrate/migrate/v4/database/sqlite3`
- [ ] Add PostgreSQL driver: `go get -u github.com/golang-migrate/migrate/v4/database/postgres`
- [ ] Add file source: `go get -u github.com/golang-migrate/migrate/v4/source/file`
- [ ] Verify `go.mod` includes all four packages
- [ ] Run `go mod tidy` to clean up
- [ ] Verify no build errors: `go build ./cmd/server`

**Validation**: `go.mod` shows migrate v4 with all drivers

---

### Task 1.2: Create Migration Runner (Shared Utility)
- [ ] Create new file: `internal/adapter/repository/migrations/runner.go`
- [ ] Implement: `RunMigrations(db *sql.DB, dbType string) error`
  - Accept database connection and type ("sqlite3" or "postgres")
  - Initialize appropriate driver based on dbType
  - Call `m.Up()` with error handling
  - Return nil on success, error on failure
- [ ] Handle `migrate.ErrNoChange` gracefully (log, don't fail)
- [ ] Handle `migrate.ErrDirty` with clear error message
- [ ] Add logging for applied migrations
- [ ] Verify function signature is testable

**Validation**: Function compiles, handles all error cases

---

## Phase 2: SQLite Adapter Refactoring

### Task 2.1: Update SQLite db.go to Use golang-migrate
- [ ] Edit: `internal/adapter/repository/sqlite/db.go`
- [ ] Remove old `runMigrations()` function (reads files manually)
- [ ] Import new migration runner: `migrations "github.com/riverlin/aiexpense/internal/adapter/repository/migrations"`
- [ ] Call `migrations.RunMigrations(db, "sqlite3")` in `OpenDB()`
- [ ] Ensure error handling: return error if migrations fail
- [ ] Verify function still returns `*sql.DB` on success

**Validation**: SQLite adapter still compiles, migrations called on OpenDB

---

### Task 2.2: Test SQLite Migrations Locally
- [ ] Run app with SQLite: `go run ./cmd/server/main.go`
- [ ] Verify logs show: "Applying migration" or "No migrations to apply"
- [ ] Check SQLite file: `sqlite3 aiexpense.db ".tables"` shows all tables
- [ ] Verify `schema_migrations` table exists: `sqlite3 aiexpense.db "SELECT * FROM schema_migrations;"`
- [ ] Stop app and restart: Verify no re-migration happens (idempotent)
- [ ] Run tests: `go test ./internal/adapter/repository/sqlite/... -v`

**Validation**: SQLite migrations work idempotently, all tests pass

---

## Phase 3: PostgreSQL Adapter Updates

### Task 3.1: Add Migration Runner to PostgreSQL Adapter
- [ ] Edit: `internal/adapter/repository/postgresql/db.go`
- [ ] Import migration runner
- [ ] Add `runMigrations()` function (internal, calls shared runner)
- [ ] Modify `OpenDB()` to call `runMigrations()` after connection
- [ ] Handle migration errors: Close DB and return error
- [ ] Verify PostgreSQL connection still works

**Validation**: PostgreSQL adapter compiles, migrations called on OpenDB

---

### Task 3.2: Test PostgreSQL Migrations (Local Supabase)
- [ ] Install Supabase CLI: `brew install supabase/tap/supabase`
- [ ] Start local Supabase: `supabase start` (wait for "Local development started")
- [ ] Extract connection: `supabase status` → copy `DATABASE_URL`
- [ ] Set environment: `export DATABASE_URL="postgresql://..."`
- [ ] Run app: `go run ./cmd/server/main.go`
- [ ] Verify logs show migrations applied
- [ ] Check Postgres: `psql "$DATABASE_URL" -c "SELECT * FROM schema_migrations;"`
- [ ] Stop app and restart with same DATABASE_URL: Verify idempotent
- [ ] Stop Supabase: `supabase stop`

**Validation**: PostgreSQL migrations work against local Supabase

---

## Phase 4: Migration File Refactoring

### Task 4.1: Create .down.sql for Migration 001
- [ ] Create: `migrations/001_init_schema.down.sql`
- [ ] Drop all tables created in `001_init_schema.up.sql` (reverse order)
- [ ] Include: DROP users, categories, expenses, metrics tables
- [ ] Test locally: `supabase start` → Run app → Verify schema created
- [ ] Manually test rollback: `migrate -database "$DATABASE_URL" down 1`
- [ ] Verify tables dropped
- [ ] Verify `schema_migrations` updated

**Validation**: Rollback works, schema correctly reversed

---

### Task 4.2: Create .down.sql for Migration 002
- [ ] Create: `migrations/002_optimize_indexes.down.sql`
- [ ] Drop all indexes created in `002_optimize_indexes.up.sql`
- [ ] Test: Start Supabase, run app, check indexes exist, rollback to 001
- [ ] Verify indexes dropped

**Validation**: Index rollback works

---

### Task 4.3: Create .down.sql for Migration 003
- [ ] Create: `migrations/003_create_ai_cost_logs.down.sql`
- [ ] Drop `ai_cost_logs` table and related indexes
- [ ] Test rollback locally

**Validation**: ai_cost_logs table drops cleanly

---

### Task 4.4: Create .down.sql for Migration 004
- [ ] Create: `migrations/004_create_policies_table.down.sql`
- [ ] Drop `policies` table and related indexes
- [ ] Test rollback locally

**Validation**: policies table drops cleanly

---

## Phase 5: Environment and Documentation

### Task 5.1: Create .env.local.example
- [ ] Create: `.env.local.example`
- [ ] Add two options:
  - Option 1: SQLite (commented): `# DATABASE_PATH=./aiexpense.db`
  - Option 2: Supabase (commented): `# DATABASE_URL=postgresql://postgres:postgres@localhost:54322/postgres`
- [ ] Add other required env vars (GEMINI_API_KEY, AI_PROVIDER, etc.)
- [ ] Add comment: "For local Supabase: run 'supabase start' and copy DATABASE_URL from 'supabase status'"

**Validation**: File exists, examples are clear and commented

---

### Task 5.2: Create Migration Testing Documentation
- [ ] Create: `docs/migrations.md`
- [ ] Document:
  - How to write new migrations (naming, up/down)
  - How to test migrations locally (SQLite and Supabase)
  - How to rollback if needed
  - How to check migration status
- [ ] Include examples: Creating a new migration, applying to both DBs
- [ ] Link to golang-migrate docs

**Validation**: Documentation is clear and complete

---

### Task 5.3: Create Local Supabase Setup Guide
- [ ] Create: `docs/supabase-local-setup.md`
- [ ] Steps:
  1. Install Supabase CLI
  2. Start local instance (`supabase start`)
  3. Get connection string (`supabase status`)
  4. Set DATABASE_URL
  5. Run app
  6. Stop Supabase (`supabase stop`)
- [ ] Include troubleshooting section
- [ ] Link to Supabase CLI docs

**Validation**: Developer can follow guide and get Supabase working locally

---

## Phase 6: Testing and Validation

### Task 6.1: Run All Tests
- [ ] Run: `go test ./... -v`
- [ ] Verify all tests pass (both SQLite and PostgreSQL if enabled)
- [ ] Check for any migration-related test failures
- [ ] Address any failures before proceeding

**Validation**: All tests pass

---

### Task 6.2: Manual Migration Test Scenario
- [ ] Scenario: Cold start, no database
  - [ ] Delete `aiexpense.db` if it exists
  - [ ] Run app: `go run ./cmd/server/main.go`
  - [ ] Verify schema created, app starts successfully

- [ ] Scenario: Restart with existing schema
  - [ ] Stop app
  - [ ] Restart: `go run ./cmd/server/main.go`
  - [ ] Verify migrations skipped (logged as "No migrations"), app starts

- [ ] Scenario: Switch from SQLite to Supabase
  - [ ] Run Supabase: `supabase start`
  - [ ] Stop app if running
  - [ ] Set: `export DATABASE_URL="postgresql://postgres:postgres@localhost:54322/postgres"`
  - [ ] Run app: `go run ./cmd/server/main.go`
  - [ ] Verify schema created in Supabase, app starts
  - [ ] Stop app, unset DATABASE_URL, restart with SQLite
  - [ ] Verify SQLite schema still works (independent from Supabase)

**Validation**: All scenarios work as expected

---

### Task 6.3: Verify Idempotency
- [ ] SQLite: Run app 3 times, verify schema_migrations unchanged (same version rows)
- [ ] PostgreSQL: Run app 3 times against Supabase, verify same result
- [ ] Check logs: No duplicate migration messages

**Validation**: Migrations are truly idempotent

---

### Task 6.4: Verify Dirty State Handling
- [ ] Simulate dirty state (optional, advanced):
  - [ ] Manually set dirty flag in schema_migrations (if testable)
  - [ ] Try to run app, verify it fails with clear error
- [ ] Document recovery procedure in docs

**Validation**: Dirty state is handled gracefully

---

## Phase 7: Integration and Documentation Updates

### Task 7.1: Update README
- [ ] Update database section to mention golang-migrate
- [ ] Add local Supabase setup as alternative to SQLite for dev
- [ ] Link to new migration docs
- [ ] Update quick-start to show both options

**Validation**: README is accurate and helpful

---

### Task 7.2: Create CHANGELOG Entry
- [ ] Add entry: "Migrated from ad-hoc SQLite migrations to golang-migrate versioning"
- [ ] Include: Benefits (audit trail, idempotency, rollback)
- [ ] Note: Both SQLite and PostgreSQL now use same migration system

**Validation**: CHANGELOG updated

---

## Final Verification

### Checklist Before Marking Complete

- [ ] All golang-migrate dependencies added
- [ ] SQLite adapter uses golang-migrate
- [ ] PostgreSQL adapter uses golang-migrate
- [ ] All .down.sql files created for rollback
- [ ] Local Supabase testing works
- [ ] All tests pass
- [ ] Documentation complete and clear
- [ ] No breaking changes to existing API
- [ ] README and CHANGELOG updated

### Sign-Off

Once all tasks complete:
1. Run full test suite one more time: `go test ./... -v`
2. Test both SQLite and PostgreSQL locally
3. Verify no errors or warnings
4. Update this checklist to mark all as `[x]`
5. Ready for code review and merge
