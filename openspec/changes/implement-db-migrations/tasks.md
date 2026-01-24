# Tasks: Implement Automated Database Migrations

**Change ID**: `implement-db-migrations`

---

## Dependency Map

```
PHASE 1: Infrastructure
  ├─ T1: Add golang-migrate dependency
  ├─ T2: Create .down.sql files for existing migrations
  └─ T3: Implement migration runner interface

PHASE 2: Integration (depends on Phase 1)
  ├─ T4: Implement PostgreSQL migration adapter
  ├─ T5: Update SQLite migration adapter
  └─ T6: Update main.go to call migration runner

PHASE 3: Testing (depends on Phase 2)
  ├─ T7: Set up local Supabase testing environment
  ├─ T8: Write integration tests (Local Supabase PostgreSQL)
  ├─ T9: Write integration tests (SQLite)
  ├─ T10: Test idempotency and rollback
  └─ T11: Test dirty state recovery

PHASE 4: Documentation (depends on Phase 3)
  ├─ T12: Create Supabase deployment script
  ├─ T13: Update DEPLOYMENT.md
  ├─ T14: Document migration procedures in README
  └─ T15: Archive this proposal
```

---

## Phase 1: Infrastructure Setup

### T1: Add golang-migrate Dependency
- **What**: Add golang-migrate to go.mod
- **Validation**:
  - [ ] `go get github.com/golang-migrate/migrate/v4`
  - [ ] `go get github.com/golang-migrate/migrate/v4/database/postgres`
  - [ ] `go get github.com/golang-migrate/migrate/v4/database/sqlite3`
  - [ ] `go get github.com/golang-migrate/migrate/v4/source/file`
  - [ ] `go mod tidy` succeeds
  - [ ] `go mod verify` succeeds
- **Effort**: Trivial (< 5 min)
- **Owner**: Backend

### T2: Create .down.sql Files for Existing Migrations
- **What**: Create rollback migrations for each existing .up.sql
- **Details**:
  - `001_init_schema.down.sql`: DROP TABLE statements in reverse order
  - `002_optimize_indexes.down.sql`: DROP INDEX statements
  - `003_create_ai_cost_logs.down.sql`: DROP TABLE statement
  - `004_create_policies_table.down.sql`: DROP TABLE statement
- **Validation**:
  - [ ] Each .down.sql exactly reverses its .up.sql
  - [ ] Tested locally: `migrate -database "sqlite:///test.db" down` executes without error
  - [ ] Schema is empty after rollback
- **Effort**: Small (< 20 min)
- **Owner**: Backend

### T3: Implement Migration Runner Interface
- **What**: Create reusable migration runner function
- **Details**:
  - Location: `internal/adapter/repository/migration.go` (new file)
  - Function signature: `RunMigrations(db *sql.DB, dbType string) error`
  - Supports both PostgreSQL and SQLite drivers
  - Returns error if dirty state detected
  - Logs applied migrations
- **Validation**:
  - [ ] Function compiles without errors
  - [ ] Handles missing `./migrations` directory gracefully
  - [ ] Returns proper error types for dirty/failed states
- **Effort**: Small (< 30 min)
- **Owner**: Backend

---

## Phase 2: Integration

### T4: Implement PostgreSQL Migration Adapter
- **What**: Update `internal/adapter/repository/postgresql/db.go` to use golang-migrate
- **Details**:
  - Replace current OpenDB() with migration runner call
  - Initialize postgres driver for migrate
  - Call RunMigrations() before returning db connection
  - Log migration results
- **Validation**:
  - [ ] App starts with PostgreSQL and runs migrations
  - [ ] Subsequent starts skip already-applied migrations
  - [ ] schema_migrations table exists with correct version entries
- **Effort**: Small (< 30 min)
- **Owner**: Backend

### T5: Update SQLite Migration Adapter
- **What**: Update `internal/adapter/repository/sqlite/db.go` to use golang-migrate instead of runMigrations()
- **Details**:
  - Replace current runMigrations() implementation
  - Initialize sqlite3 driver for migrate
  - Call shared RunMigrations() function
  - Maintain backward compatibility with existing schema
- **Validation**:
  - [ ] SQLite migrations run on startup
  - [ ] Existing tests still pass
  - [ ] schema_migrations table created in SQLite
- **Effort**: Small (< 30 min)
- **Owner**: Backend

### T6: Update Main Application Entry Point
- **What**: Verify cmd/server/main.go calls migration automatically
- **Details**:
  - Migrations already run in OpenDB() (from T4, T5)
  - Verify no additional calls needed
  - Test app startup logs migration status
- **Validation**:
  - [ ] `go run cmd/server/main.go` logs migration execution
  - [ ] App starts successfully with clean database
  - [ ] App starts successfully with migrated database (idempotent)
- **Effort**: Trivial (< 5 min)
- **Owner**: Backend

---

## Phase 3: Testing

### T7: Set Up Local Supabase Testing Environment
- **What**: Configure local Supabase for migration testing
- **Details**:
  - Document Supabase CLI installation
  - Create `docker-compose` reference or document `supabase start` workflow
  - Create helper script: `scripts/test-with-local-supabase.sh`
  - Script should: start Supabase → export DATABASE_URL → run tests → stop Supabase
  - Document in README/TESTING.md
- **Validation**:
  - [ ] Supabase CLI installed locally
  - [ ] `supabase start` brings up PostgreSQL on localhost:54322
  - [ ] Can export DATABASE_URL from `supabase status`
  - [ ] Helper script is executable and documented
- **Effort**: Small (< 30 min)
- **Owner**: Backend

### T8: Write PostgreSQL Integration Tests (Local Supabase)
- **What**: Create test file for PostgreSQL migrations using local Supabase
- **Location**: `internal/adapter/repository/postgresql/migration_test.go` (new file)
- **Test Cases**:
  - [ ] Fresh database: all migrations applied
  - [ ] Existing database: pending migrations only
  - [ ] Dirty state detection and error
  - [ ] Rollback capability
- **Validation**:
  - [ ] Tests run against local Supabase PostgreSQL (via helper script from T7)
  - [ ] `go test ./internal/adapter/repository/postgresql -v -run Migration` passes
  - [ ] Coverage > 80%
  - [ ] Same test works against production Supabase (no breaking changes)
- **Effort**: Medium (< 1 hour)
- **Owner**: Backend

### T9: Write SQLite Integration Tests
- **What**: Create test file for SQLite migrations
- **Location**: `internal/adapter/repository/sqlite/migration_test.go` (new file)
- **Test Cases**:
  - [ ] Fresh database: all migrations applied
  - [ ] Existing database: pending migrations only
  - [ ] schema_migrations table exists and is populated
- **Validation**:
  - [ ] `go test ./internal/adapter/repository/sqlite -v -run Migration` passes
  - [ ] Coverage > 80%
  - [ ] Tests run on CI/CD
- **Effort**: Medium (< 1 hour)
- **Owner**: Backend

### T10: Test Idempotency and Rollback
- **What**: Validate that migrations are safe to run multiple times and can be rolled back
- **Details**:
  - Run migrations twice; second run should be no-op
  - Run down migrations; schema should match pre-migration state
  - Run up again; schema should match post-migration state
- **Location**: New test in `internal/adapter/repository/postgresql/migration_test.go`
- **Validation**:
  - [ ] Idempotency test passes (no errors on duplicate run)
  - [ ] Rollback test passes (schema reverted correctly)
  - [ ] Forward-backward-forward cycle succeeds
- **Effort**: Small (< 30 min)
- **Owner**: Backend

### T11: Test Dirty State Recovery
- **What**: Validate app behavior when schema_migrations is in a dirty state
- **Details**:
  - Manually mark a migration as incomplete in schema_migrations
  - Start app and verify it detects dirty state
  - Document manual recovery steps
- **Validation**:
  - [ ] App exits with error when dirty state is detected
  - [ ] Log message is clear and actionable
  - [ ] Recovery procedure is documented in logs
- **Effort**: Small (< 30 min)
- **Owner**: Backend

---

## Phase 4: Documentation and Deployment

### T12: Create Supabase Deployment Script
- **What**: Bash script to deploy migrations to Supabase
- **Location**: `scripts/deploy-migrations.sh` (new file)
- **Details**:
  - Accept DATABASE_URL as argument
  - Check/install golang-migrate CLI
  - Apply migrations: `migrate -path ./migrations -database "$DATABASE_URL" up`
  - Exit with error code if migration fails
- **Validation**:
  - [ ] Script is executable (`chmod +x`)
  - [ ] Script accepts DATABASE_URL parameter
  - [ ] Test run: `./scripts/deploy-migrations.sh "postgresql://localhost/test"`
- **Effort**: Trivial (< 15 min)
- **Owner**: DevOps

### T13: Update DEPLOYMENT.md
- **What**: Document migration deployment procedures for Supabase and local testing
- **Location**: Create/update `DEPLOYMENT.md`
- **Content**:
  - How to run migrations locally (SQLite)
  - How to test migrations against local Supabase (using helper script from T7)
  - How to run migrations on production Supabase (using script from T12)
  - Rollback procedure (local and production)
  - Dirty state recovery
  - Troubleshooting common issues
- **Validation**:
  - [ ] Documentation is clear and step-by-step
  - [ ] Includes examples for both local and production workflows
  - [ ] Links to relevant files (scripts, specs)
- **Effort**: Small (< 30 min)
- **Owner**: Documentation

### T14: Update README.md
- **What**: Add migration overview to README
- **Location**: `README.md` → Database section
- **Content**:
  - Brief explanation of golang-migrate
  - Link to DEPLOYMENT.md for detailed procedures
  - Quick start: `go run cmd/server/main.go` (mentions migrations run automatically)
  - Testing locally with Supabase: reference to T7 helper script
- **Validation**:
  - [ ] README is still readable and not too long
  - [ ] Links are correct and functional
- **Effort**: Trivial (< 10 min)
- **Owner**: Documentation

### T15: Archive Proposal and Close Change
- **What**: Mark this proposal as complete and archive it
- **Details**:
  - Update proposal.md status to "✓ Complete"
  - Run `openspec archive implement-db-migrations`
  - Verify no warnings from validation
- **Validation**:
  - [ ] Proposal archived successfully
  - [ ] All deliverables in place (design, spec, tasks)
  - [ ] At least one commit references this change
- **Effort**: Trivial (< 5 min)
- **Owner**: Project Lead

---

## Validation Checklist (Pre-Deployment)

Before merging to main:
- [ ] All Phase 1–4 tasks completed
- [ ] All tests pass: `go test ./... -v`
- [ ] Migrations tested locally (SQLite + PostgreSQL)
- [ ] Deployment script tested with real Supabase URL
- [ ] Documentation reviewed and accurate
- [ ] Code review approved
- [ ] No breaking changes to existing functionality

---

## Parallelizable Work

These tasks can run in parallel:
- **T8 & T9**: PostgreSQL (Supabase) and SQLite tests (independent)
- **T12, T13, T14**: Documentation and scripts (independent)
- **T10 & T11**: Additional tests (can start after T9)
- **T7**: Set up local Supabase (blocking for T8, but can happen in parallel with T1–T6)

---

## Success Metrics

✅ **Outcomes**:
1. All migrations tracked in `schema_migrations` table
2. Zero manual SQL execution needed for Supabase deployments
3. App startup time increases by < 50ms
4. Tests cover > 80% of migration logic
5. Documentation enables on-call engineers to handle rollbacks
6. Zero production incidents due to schema mismatches

