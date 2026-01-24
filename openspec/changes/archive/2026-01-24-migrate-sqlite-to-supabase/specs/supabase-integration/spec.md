# Specification: Local Supabase Integration for Testing

**Capability ID**: `supabase-integration`

**Status**: 🔄 In Review

**Related Capabilities**: `database-migrations`

---

## Purpose

Enable developers to test the application against a local Supabase instance (PostgreSQL) that mirrors production schema, ensuring application code and migrations work correctly before deployment to cloud Supabase.

---

## ADDED Requirements

### Requirement: Local Supabase CLI Setup
**ID**: `SUPABASE_LOCAL_CLI_SETUP`

Developers MUST be able to install and initialize the Supabase CLI to run a local PostgreSQL instance.

#### Scenario: Supabase CLI Installation

Given a developer on macOS, Linux, or Windows
When they follow the provided setup guide
Then they can successfully install Supabase CLI via:
  - macOS: `brew install supabase/tap/supabase`
  - Linux/Windows: Official Supabase CLI installer
And the CLI is available in their PATH
And `supabase --version` returns a version number

#### Scenario: Local Supabase Instance Startup

Given Supabase CLI is installed
When a developer runs `supabase start` in the project directory
Then:
  - A local PostgreSQL database starts
  - Local Redis, Auth services (if needed) start
  - CLI outputs connection details
  - The instance is ready for application connections
And running `supabase status` shows:
  - PostgreSQL connection string (DATABASE_URL)
  - Port (default 54322)
  - Credentials (default: postgres/postgres)

#### Scenario: Local Supabase Instance Shutdown

Given a local Supabase instance is running
When a developer runs `supabase stop`
Then:
  - All services gracefully shut down
  - Local data is preserved (can restart and resume)
  - Or data is cleared (configurable per team preference)
And the developer can restart later with `supabase start`

---

### Requirement: Environment Configuration
**ID**: `SUPABASE_ENV_CONFIG`

The application MUST support easy switching between SQLite (local) and Supabase PostgreSQL (local testing or cloud).

#### Scenario: .env Configuration File

Given a developer setting up local development
When they create or modify `.env` file
Then they can set ONE of:
  - Option A: `DATABASE_PATH=./aiexpense.db` (SQLite)
  - Option B: `DATABASE_URL=postgresql://postgres:postgres@localhost:54322/postgres` (Local Supabase)
  - Option C: `DATABASE_URL=postgresql://user:pass@db.supabase.co/postgres` (Cloud Supabase)
And they MUST NOT set both DatabasePath and DatabaseURL
And the application automatically selects the correct adapter

#### Scenario: Connection String Extraction

Given a local Supabase instance is running
When a developer runs `supabase status`
Then the output includes the PostgreSQL connection string (DATABASE_URL)
And they can copy it directly into `.env`
And the application connects successfully

#### Scenario: Environment Override

Given a `.env` file with SQLite configuration
When a developer temporarily wants to test with Supabase
Then they can override via shell:
  ```bash
  export DATABASE_URL="postgresql://..."
  go run ./cmd/server/main.go
  ```
And the application uses PostgreSQL (environment takes precedence)
And reverting the override restores SQLite behavior

---

### Requirement: Application Testing Against Local Supabase
**ID**: `SUPABASE_LOCAL_APP_TESTING`

The application MUST run successfully against a local Supabase instance with automatic schema migration.

#### Scenario: Cold Start Migration on Supabase

Given:
  - Local Supabase is running (`supabase start`)
  - DATABASE_URL points to local Supabase
  - Schema does not exist (fresh Supabase instance)
When the application starts
Then:
  - golang-migrate connects to PostgreSQL
  - All migration files execute (001-004)
  - Tables are created: users, categories, expenses, metrics, ai_cost_logs, policies
  - Indexes are optimized
  - Application starts successfully
And running the app again confirms idempotency

#### Scenario: Data Persistence Between Restarts

Given the application successfully created schema on Supabase
When the application is stopped and restarted
Then:
  - Migrations are skipped (already applied)
  - Database connection succeeds
  - Application starts without errors
  - Any test data inserted persists

#### Scenario: API Testing Against Supabase

Given the application is running with Supabase backend
When developers call API endpoints (e.g., POST /expenses, GET /expenses)
Then:
  - Requests are routed to PostgreSQL (not SQLite)
  - Data is persisted and retrieved correctly
  - Performance is acceptable (no excessive latency)
  - Schema queries execute without errors

#### Scenario: Rollback Testing on Supabase

Given a local Supabase with all migrations applied
When a developer manually rolls back (optional advanced testing)
  ```bash
  migrate -path ./migrations -database "$DATABASE_URL" down
  ```
Then:
  - The `.down.sql` files execute
  - Schema is reversed to previous state
  - No data corruption occurs
  - Applying migrations again restores the schema

---

### Requirement: Testing Both SQLite and PostgreSQL
**ID**: `SUPABASE_DUAL_DB_TESTING`

The test suite and CI/CD MUST validate the application works correctly with both SQLite and PostgreSQL backends.

#### Scenario: Unit Tests with SQLite

Given test suite runs locally
When tests execute using SQLite (in-memory or temp file)
Then:
  - Repository tests use SQLite adapter
  - In-memory `:memory:` database for isolation
  - Tests are fast and independent
  - All data access logic verified for SQLite compatibility

#### Scenario: Integration Tests with Local Supabase

Given CI/CD pipeline or developer explicitly runs integration tests
When tests execute with DATABASE_URL set to local Supabase
Then:
  - Integration tests use PostgreSQL adapter
  - Real database operations verified
  - Connection pooling tested
  - Query performance validated
  - Tests clean up and leave database in good state

#### Scenario: Test Isolation

Given multiple test runs (SQLite then PostgreSQL)
When tests execute sequentially
Then:
  - SQLite tests don't interfere with PostgreSQL tests
  - Database state is properly isolated
  - Each test starts with clean schema
  - No flaky tests due to data carryover

---

### Requirement: Documentation and Developer Workflow
**ID**: `SUPABASE_DEV_WORKFLOW_DOCS`

Developers MUST have clear documentation to set up and use local Supabase for testing.

#### Scenario: Setup Guide

Given a developer new to the project
When they read the setup documentation
Then they find clear step-by-step instructions:
  1. Install Supabase CLI
  2. Run `supabase start` in project directory
  3. Copy DATABASE_URL from `supabase status`
  4. Set in `.env` or export
  5. Run application
  6. Verify schema created (check Supabase dashboard or CLI)
  7. Test API endpoints
  8. Stop Supabase when done: `supabase stop`

#### Scenario: Troubleshooting Guide

Given a developer encounters an issue (e.g., connection refused, migrations fail)
When they check the troubleshooting documentation
Then they find:
  - Common errors and solutions
  - How to check Supabase status
  - How to view logs
  - How to reset Supabase state
  - Link to Supabase CLI documentation

#### Scenario: Migration Testing Workflow

Given a developer needs to create a new migration
When they follow the migration creation guide
Then they can:
  1. Create `NNN_feature.up.sql` and `NNN_feature.down.sql`
  2. Test against SQLite: `go run ./cmd/server/main.go`
  3. Test against Supabase: Set DATABASE_URL, run app
  4. Verify schema changes in both databases
  5. Commit files when satisfied

---

### Requirement: Performance and Resource Efficiency
**ID**: `SUPABASE_RESOURCE_EFFICIENCY`

Local Supabase setup MUST not impose excessive resource requirements on developer machines.

#### Scenario: Memory and CPU Usage

Given a local Supabase instance running
When monitoring resource usage
Then:
  - PostgreSQL uses <500MB memory (typical)
  - CPU usage is idle/minimal when not in use
  - Application can run alongside other development tools
  - Total setup is reasonable for typical dev machine (8GB RAM+)

#### Scenario: Startup and Shutdown Time

Given a developer starts/stops Supabase multiple times daily
When they run `supabase start` and `supabase stop`
Then:
  - Startup completes in <30 seconds
  - Shutdown is immediate
  - Developer workflow is not blocked by slow startup

#### Scenario: Disk Space

Given a developer has limited disk space
When they use local Supabase
Then:
  - Database footprint is reasonable (<1GB for typical test data)
  - Can be cleared with `supabase stop --no-backup` (if supported)
  - Documentation explains cleanup options

---

### Requirement: Error Recovery
**ID**: `SUPABASE_ERROR_RECOVERY`

Developers MUST be able to recover from common Supabase errors and reset the local instance.

#### Scenario: Schema Corruption Recovery

Given a local Supabase instance in a bad state (e.g., dirty migration, schema conflict)
When a developer needs to recover
Then they can:
  1. Stop Supabase: `supabase stop`
  2. Reset to clean state (clear data, or full reset)
  3. Start fresh: `supabase start`
  4. Restart application (migrations re-apply)
And the instance is fully usable again

#### Scenario: Connection Timeout Recovery

Given the application fails to connect to Supabase (timeout or connection refused)
When a developer checks logs and troubleshooting guide
Then they can:
  1. Verify Supabase is running: `supabase status`
  2. Restart if needed: `supabase stop` then `supabase start`
  3. Verify connectivity: `psql "$DATABASE_URL" -c "SELECT 1;"`
  4. Restart application
And connection is restored

---

## MODIFIED Requirements

### Requirement: Configuration Loading
**ID**: `CONFIG_DATABASE_SELECTION` (MODIFIED)

The configuration loading MUST support three database options with clear precedence.

#### Scenario: Database Selection Logic (Updated)

Given application startup
When loading configuration
Then:
  1. Check environment variable `DATABASE_URL`
     - If set: Use PostgreSQL adapter (for Supabase local or cloud)
     - If not set: Continue to step 2
  2. Check environment variable `DATABASE_PATH`
     - If set: Use SQLite adapter (path to .db file)
     - If not set: Continue to step 3
  3. Use default: `DATABASE_PATH=./aiexpense.db` (SQLite)

#### Scenario: Validation

Given any configuration
When validating
Then:
  - Exactly one of DATABASE_URL or DATABASE_PATH is effective
  - Both can coexist in `.env` but only one is used (precedence honored)
  - Clear error message if neither is valid

---

## REMOVED Requirements

None. This is an additive change; no existing requirements are removed.

---

## Cross-References

- **Specification**: `database-migrations` - Migration system that runs on any database
- **Documentation**: `docs/supabase-local-setup.md` (to be created)
- **Documentation**: `docs/migrations.md` (to be created)
- **Environment**: `.env.local.example` (to be created)
- **Tool**: Supabase CLI - https://supabase.com/docs/guides/cli

---

## Implementation Notes

- **Supabase CLI**: Used only for local development, not production
- **Database Selection**: Precedence is DATABASE_URL > DATABASE_PATH > default
- **Local vs Cloud**: Same codebase works with both; only DATABASE_URL differs
- **Data Isolation**: Local Supabase data is separate from cloud Supabase (no cross-contamination)

---

## Validation Checklist

After implementation, verify:

- [ ] Supabase CLI installation documented and tested
- [ ] Developer can start/stop local Supabase
- [ ] DATABASE_URL extraction works
- [ ] Application connects and runs migrations successfully on local Supabase
- [ ] API endpoints work against Supabase backend
- [ ] Tests pass with both SQLite and PostgreSQL
- [ ] Troubleshooting guide covers common issues
- [ ] Setup time is <5 minutes for experienced developer
- [ ] Resource usage is reasonable
- [ ] Error recovery procedure works
