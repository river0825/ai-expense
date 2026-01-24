# Proposal: Migrate SQLite to Supabase for Production

**Change ID**: `migrate-sqlite-to-supabase`

**Status**: 🔄 In Review

**Author**: Claude Code

**Created**: 2026-01-24

---

## Executive Summary

Currently, the application supports **dual database backends**: SQLite (local development) and PostgreSQL (production via environment variable). However, SQLite migrations are ad-hoc (loaded at runtime), and there's no proper versioning system. This proposal establishes:

1. **Versioned Database Migrations** using `golang-migrate` for both SQLite (dev) and PostgreSQL/Supabase (prod)
2. **Local Supabase Testing** using Supabase CLI for realistic local development
3. **Automatic Migration Tracking** to prevent re-runs and maintain audit trail
4. **Seamless Deployment** to Supabase with safe schema versioning

### Why This Matters

- **Production Safety**: Tracked, versioned migrations prevent schema drift and enable safe rollbacks
- **Local Testing**: Developers can test against Supabase schema locally before deployment
- **Operational Excellence**: Automated migrations eliminate manual SQL execution risk
- **Scalability**: As the application grows, manual migrations become unsustainable

## Why (Required Context)

**Problem**: The application currently handles SQLite migrations ad-hoc (files loaded at runtime, re-run on every startup via `CREATE TABLE IF NOT EXISTS`). PostgreSQL has no migration system—schemas are assumed to exist. This creates:
- No audit trail of schema changes
- No versioning or idempotency guarantees
- Risky manual Supabase deployments
- Difficulty testing against production schema locally

**Solution**: Integrate golang-migrate for versioned, idempotent migrations + Supabase CLI for realistic local testing. This ensures:
- Every migration is tracked and versioned
- Migrations are guaranteed idempotent (run only once per version)
- Developers test locally against real PostgreSQL before cloud deployment
- Production deployments are safe and auditable
- Emergency rollbacks are possible via `.down.sql` files

**Impact**: Schema changes become predictable and safe. Development and production converge on the same migration system. Local testing mimics production exactly.

---

## Current State

### What Works Today

- ✅ SQLite supports runtime migrations (ad-hoc, idempotent)
- ✅ PostgreSQL can connect via `DATABASE_URL` environment variable
- ✅ Main application already has dual-DB support (SQLite vs PostgreSQL decision in main.go)
- ✅ Repository pattern abstracts database differences

### What's Missing

- ❌ No migration versioning system (migrations re-run on every startup)
- ❌ No schema migration audit trail
- ❌ PostgreSQL migrations are manual (schema assumed to exist)
- ❌ No local Supabase testing setup
- ❌ Migration failures are not tracked
- ❌ No rollback capability

---

## Proposed Solution

### Architecture Overview

```
App Startup Flow (Current):
  1. Config checks: if DATABASE_URL → PostgreSQL, else → SQLite
  2. SQLite: Runs migrations from disk (ad-hoc)
  3. PostgreSQL: Assumes schema exists (manual setup)

App Startup Flow (Proposed):
  1. Config checks: if DATABASE_URL → PostgreSQL, else → SQLite
  2. BOTH: Run golang-migrate (manages schema_migrations table)
  3. Migrations apply only if version not already recorded
  4. Application starts with guaranteed schema consistency
```

### Key Components

#### 1. golang-migrate Integration
- **Library**: `github.com/golang-migrate/migrate/v4`
- **Benefit**: Industry standard, supports both SQLite and PostgreSQL, versioned tracking
- **File Format**: SQL files with `.up.sql` and `.down.sql` variants
- **Version Tracking**: Automatic `schema_migrations` table in database

#### 2. Migration File Structure

```
migrations/
  ├── 001_init_schema.up.sql
  ├── 001_init_schema.down.sql
  ├── 002_optimize_indexes.up.sql
  ├── 002_optimize_indexes.down.sql
  ├── 003_create_ai_cost_logs.up.sql
  ├── 003_create_ai_cost_logs.down.sql
  ├── 004_create_policies_table.up.sql
  ├── 004_create_policies_table.down.sql
  └── ...
```

#### 3. Local Supabase Testing (New)

Using **Supabase CLI**, developers can:
- Run a local PostgreSQL instance matching Supabase schema
- Test migrations against real PostgreSQL schema before deployment
- Extract connection info via `supabase status`

---

## Design Decisions

| Decision | Rationale |
|----------|-----------|
| **Use golang-migrate** | Standard Go migration tool, supports SQLite+PostgreSQL, versioned, safe |
| **Embed migrations in app** | Simplifies deployment; app manages schema consistency on startup |
| **Create .down migrations** | Enables rollback capability for emergencies |
| **Version format: NNN_** | Matches current convention; human-readable ordering |
| **Supabase CLI for local dev** | Realistic PostgreSQL testing before Supabase push |
| **Auto-run on startup** | Simpler; ensures schema always matches code |

---

## Scope and Sequencing

### Phase 1: Migration Infrastructure Setup
- [ ] Add `golang-migrate` dependency to Go modules
- [ ] Create migration runner in PostgreSQL adapter
- [ ] Create migration runner in SQLite adapter
- [ ] Ensure both runners use same version tracking

### Phase 2: Existing Migration Refactoring
- [ ] Convert existing `.up.sql` files to golang-migrate format
- [ ] Create corresponding `.down.sql` files for rollback
- [ ] Validate migration idempotency
- [ ] Test both SQLite and PostgreSQL migrations

### Phase 3: Local Supabase Setup
- [ ] Document Supabase CLI installation and setup
- [ ] Create `.env.local.example` with Supabase connection string
- [ ] Add `supabase status` extraction script for developers
- [ ] Test local development flow against real PostgreSQL

### Phase 4: Testing and Documentation
- [ ] Write integration tests for both databases
- [ ] Create deployment runbook for Supabase
- [ ] Document local Supabase development workflow
- [ ] Update GitHub CI/CD if needed

---

## Implementation Notes

### Migration Runner Logic

```go
// Pseudocode
func RunMigrations(db *sql.DB, migrationType string) error {
  // Initialize migration driver
  driver := &sql.Driver{conn: db}
  m, err := migrate.NewWithDatabaseInstance(
    "file://./migrations",
    migrationType,
    driver,
  )

  // Run migrations (idempotent, only applies new versions)
  err := m.Up()
  if err != nil && err != migrate.ErrNoChange {
    return err
  }
  return nil
}
```

### Configuration Flow

```
1. Load DATABASE_URL or DATABASE_PATH
2. Open appropriate database
3. Detect DB type (sqlite3 vs postgres)
4. Run golang-migrate with correct driver
5. On success: Start application
6. On failure: Exit with error (fail safe)
```

---

## Open Questions

Before finalizing specs, clarify:

1. **Rollback Strategy**: How should the app handle migration failures?
   - **Option A (Recommended)**: Fail loudly, log error, exit (safest)
   - **Option B**: Auto-rollback last migration, retry (complex)

2. **Supabase Dashboard vs CLI**: Should developers push migrations via Supabase CLI or dashboard?
   - **Option A (Recommended)**: CLI in deployment script (reproducible, version-controlled)
   - **Option B**: Manual Supabase dashboard (simpler for one-off, risky for multi-env)

3. **Dry-Run Support**: Should the app support a `--dry-run` flag to preview pending migrations?
   - **Option A (Recommended)**: Yes, useful for pre-deployment validation
   - **Option B**: No, trust tests and CI (simpler)

---

## Validation Checklist

After implementation, verify:

- [ ] SQLite migrations run automatically on app startup
- [ ] PostgreSQL migrations run automatically on app startup
- [ ] schema_migrations table tracks applied versions
- [ ] Re-running app doesn't re-apply migrations (idempotent)
- [ ] Rollback (.down.sql) files work correctly
- [ ] Local Supabase setup matches production PostgreSQL schema
- [ ] Tests pass for both SQLite and PostgreSQL
- [ ] Documentation is complete and clear

---

## Artifacts

This proposal includes:
- **proposal.md** (this file) - Executive summary and context
- **design.md** - Technical architecture and implementation details
- **tasks.md** - Step-by-step implementation checklist
- **specs/database-migrations/spec.md** - Migration system requirements
- **specs/supabase-integration/spec.md** - Supabase testing requirements

---

## Next Steps

1. Review this proposal
2. Clarify the three open questions (if needed)
3. Review design.md and tasks.md
4. Approve and proceed with implementation
