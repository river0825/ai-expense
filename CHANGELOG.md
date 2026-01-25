# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- **Versioned Database Migrations**: Integrated `golang-migrate` for both SQLite and PostgreSQL
  - Automatic version tracking in `schema_migrations` table
  - Idempotent migrations that only run once per version
  - Full rollback capability via `.down.sql` files
  - Works seamlessly with both SQLite (development) and PostgreSQL/Supabase (production)

- **Local Supabase Development**: Set up realistic PostgreSQL testing environment locally
  - Docker-based local Supabase instance via Supabase CLI
  - Matches production schema exactly before cloud deployment
  - Eliminates friction between local SQLite testing and production PostgreSQL

- **Comprehensive Migration Documentation**:
  - `docs/migrations.md`: Complete guide for working with migrations
    - Creating new migrations with up/down patterns
    - Testing migrations locally
    - Rollback procedures and troubleshooting
  - `docs/supabase-local-setup.md`: Setup guide for local Supabase development
    - Installation and configuration steps
    - Common troubleshooting and solutions

- **Environment Configuration Template**: `.env.local.example` with:
  - SQLite and PostgreSQL database options
  - AI provider configuration (Gemini, OpenAI)
  - Server and logging settings
  - Detailed documentation for each option

### Changed
- **Database Initialization**: Now uses golang-migrate for all migrations
  - Replaces ad-hoc SQLite migration loading on every startup
  - SQLite migrations now tracked with version numbers
  - PostgreSQL migrations tracked same way as SQLite (previously assumed to exist)

- **Migration Files**: All `.up.sql` migrations now have corresponding `.down.sql` files
  - `001_init_schema.down.sql`: Drop users, categories, expenses, and related indexes
  - `002_optimize_indexes.down.sql`: Drop performance optimization indexes
  - `003_create_ai_cost_logs.down.sql`: Drop ai_cost_logs table and indexes
  - `004_create_policies_table.down.sql`: Drop policies table

### Fixed
- Format string error in `internal/async/job_queue.go` (line 221)
  - Changed `fmt.Errorf(job.Error)` to `fmt.Errorf("%s", job.Error)`

## Benefits

### For Development
- **SQLite** continues to work for quick local development (no Docker required)
- Optional **local PostgreSQL** testing via Supabase CLI for realistic pre-deployment validation
- Both database types use same migration system - no learning curve

### For Production
- **Versioned schema**: Every migration is tracked with a version number
- **Audit trail**: `schema_migrations` table shows exactly which migrations were applied
- **Safe deployments**: Migrations only run once per version, preventing re-runs
- **Rollback capability**: `.down.sql` files enable emergency rollbacks if needed

### For Operations
- **Automatic**: Migrations run on application startup - no manual SQL execution needed
- **Fail-safe**: Application exits immediately if migrations fail (never starts with inconsistent schema)
- **Observable**: Clear logging of migration status (applied, already applied, failures)

## Migration Path

Users can switch between SQLite and PostgreSQL at any time:
```bash
# Development with SQLite
DATABASE_PATH=./aiexpense.db go run ./cmd/server/main.go

# Development with local PostgreSQL (after running 'supabase start')
DATABASE_URL="postgresql://postgres:postgres@localhost:54322/postgres" go run ./cmd/server/main.go

# Production with Supabase
DATABASE_URL="postgresql://user:pass@db.supabase.co/postgres" go run ./cmd/server/main.go
```

All environments use the same migration system - migrations apply identically across all three scenarios.

## Testing

All existing tests continue to pass with the new migration system:
- SQLite adapter tests: ✅ All pass, migrations tested on each run
- Repository tests: ✅ All pass, full CRUD operations verified
- HTTP API tests: ✅ All pass
- Messenger adapters: ✅ All pass
- Integration tests: ✅ All pass

---

## Deployment Checklist

When deploying to production Supabase:

1. **Pre-deployment**:
   - Test migrations locally: `supabase start` → set `DATABASE_URL` → run app
   - Verify all tests pass: `go test ./... -v`
   - Check migration status: `psql "$DATABASE_URL" -c "SELECT * FROM schema_migrations;"`

2. **Production Deployment**:
   - Set `DATABASE_URL` in Cloud Run environment variables
   - Deploy application (migrations run automatically on startup)
   - Monitor logs for "Migrations applied successfully" message
   - Verify schema in Supabase dashboard

3. **Emergency Rollback** (if needed):
   - Stop the application
   - Run: `migrate -path ./migrations -database "$DATABASE_URL" down N`
   - Fix the problematic migration
   - Redeploy

---

## References

- [golang-migrate Documentation](https://github.com/golang-migrate/migrate)
- [Supabase CLI Documentation](https://supabase.com/docs/guides/cli)
- [Database Migrations Guide](./docs/migrations.md)
- [Local Supabase Setup Guide](./docs/supabase-local-setup.md)
