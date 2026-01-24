# Proposal: Implement Automated Database Migrations for PostgreSQL/Supabase

**Change ID**: `implement-db-migrations`

**Status**: 🔄 In Review

**Author**: Claude Code

**Created**: 2026-01-22

---

## Executive Summary

Currently, the application handles SQLite migrations through an ad-hoc approach (reading `.sql` files from disk at runtime). With PostgreSQL/Supabase now in production, we need a **proper schema versioning and migration system** that:

1. Tracks which migrations have been applied (preventing re-runs and drift)
2. Supports both SQLite (dev) and PostgreSQL/Supabase (prod)
3. Detects schema changes automatically
4. Can be deployed safely to Supabase without manual SQL execution
5. Rolls back gracefully on failure

---

## Problem Statement

**Current State:**
- SQLite migrations are loaded from SQL files and executed every startup (idempotent via `CREATE TABLE IF NOT EXISTS`)
- PostgreSQL has no migration layer—schemas are assumed to already exist
- There is no versioning or audit trail of what has been applied
- Supabase deployments require manual SQL script execution or manual migrations via the Supabase dashboard
- If the database schema changes (new fields, tables), deployment becomes error-prone

**Why This Matters:**
- **Production Safety**: Without versioning, you can't safely track schema evolution or roll back changes
- **DevOps Compliance**: Automated migrations are industry standard for any data-driven application
- **Future Growth**: As the application scales, manual migrations become unsustainable and risky

---

## Proposed Solution

Implement **`golang-migrate`** (github.com/golang-migrate/migrate) as the schema management tool. This is:
- **Popular**: Industry standard for Go applications
- **DB-Agnostic**: Supports PostgreSQL and SQLite
- **Versioned**: Maintains a `schema_migrations` table tracking applied migrations
- **Flexible**: CLI for manual runs or embedded in application startup
- **Safe**: Supports up/down migrations and detects dirty states

### High-Level Architecture

```
App Startup Flow:
  1. Check database connection
  2. Run migrations (via golang-migrate)
     - Migrate reads schema_migrations table
     - Applies only new versions
     - Logs each applied migration
  3. Start application

Supabase Deployment Flow:
  1. Developer runs: migrate -path ./migrations -database $DATABASE_URL up
  2. Migrate validates and applies pending migrations
  3. Deployment safe-guards against partial/failed migrations
```

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

---

## Key Design Decisions

| Decision | Rationale |
|----------|-----------|
| **Use golang-migrate** | Battle-tested, supports both SQLite & PostgreSQL, versioned tracking, safe rollbacks |
| **Embed migrations in app** | Simplifies deployment; no external tooling required for common scenarios |
| **Provide CLI option** | Allows Supabase deployments or manual intervention when needed |
| **Track migrations** | `schema_migrations` table prevents re-runs and provides audit trail |
| **Create down migrations** | Enables rollback capability for emergencies |
| **Version format: `NNN_description`** | Matches existing convention; human-readable |

---

## Scope and Sequencing

### Phase 1: Core Migration Infrastructure
- Add `golang-migrate` dependency
- Refactor SQL files to include `.down` variants for rollback
- Create migration runner in PostgreSQL adapter (embedded)
- Create CLI command for manual migration runs

### Phase 2: Integration and Testing
- Update SQLite adapter to use golang-migrate (for consistency)
- Write integration tests for both databases
- Validate migration idempotency
- Document deployment procedures

### Phase 3: Deployment and Documentation
- Update Supabase deployment docs
- Create runbook for manual migration on Supabase
- Add CI/CD hooks to validate migrations
- Archive this proposal

---

## Open Questions

Before finalizing the spec, please clarify:

1. **CLI vs Embedded Migration**: Should migrations run automatically on app startup, or should they be manual via CLI?
   - **Option A (Recommended)**: Auto-run on startup (simpler, less manual work)
   - **Option B**: Manual via CLI (more control, requires deployment script)

2. **Rollback Strategy**: Should the application support automatic rollback on migration failure, or fail loudly?
   - **Option A (Recommended)**: Fail loudly (safer; prevents data corruption)
   - **Option B**: Auto-rollback (simpler recovery)

3. **Supabase-Specific Tooling**: Should we wrap migrate in a Terraform module or shell script for Supabase?
   - **Option A (Recommended)**: Shell script in deployment docs (lightweight)
   - **Option B**: Terraform module (infrastructure-as-code)
   - **Option C**: Neither (manual runs for now)

---

## Artifacts Pending User Input

Once you clarify the open questions, I will:
1. Finalize `spec.md` with concrete requirements
2. Break down into `tasks.md` with validation steps
3. Create `design.md` with implementation architecture
4. Run `openspec validate` to ensure completeness

---

## Resources and References

- **golang-migrate**: https://github.com/golang-migrate/migrate
- **Supabase Docs**: https://supabase.com/docs/guides/database/connecting-to-postgres
- **Migration Best Practices**: https://wiki.postgresql.org/wiki/Migration_tools

