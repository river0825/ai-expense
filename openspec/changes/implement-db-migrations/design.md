# Design: Automated Database Migrations for PostgreSQL/Supabase

**Change ID**: `implement-db-migrations`

---

## System Architecture

### Migration Lifecycle

```
Developer Workflow:
  ┌─────────────────────────────────────────────────────┐
  │ 1. Create new schema change (.up.sql + .down.sql)  │
  │ 2. Test locally:                                    │
  │    a. SQLite (quick iteration)                      │
  │    b. Supabase local (matches production PG)       │
  │ 3. Run integration tests against both              │
  │ 4. Commit migration files                          │
  │ 5. Deploy to prod (Supabase)                       │
  └─────────────────────────────────────────────────────┘

App Startup Flow:
  ┌─────────────────────────────────────────────────────┐
  │ 1. Load config (detect DB type)                    │
  │ 2. Open DB connection                              │
  │ 3. Initialize migration runner (golang-migrate)    │
  │ 4. Read pending migrations from ./migrations       │
  │ 5. Check schema_migrations table for applied       │
  │ 6. Apply new migrations in order                   │
  │ 7. Continue to business logic                      │
  └─────────────────────────────────────────────────────┘

Local Supabase Testing Flow:
  ┌─────────────────────────────────────────────────────┐
  │ 1. Start local Supabase: supabase start             │
  │ 2. Export local DATABASE_URL                       │
  │ 3. Test migrations: migrate -path ./migrations \  │
  │                      -database $DATABASE_URL up   │
  │ 4. Run integration tests against local PG          │
  │ 5. Stop local Supabase: supabase stop              │
  └─────────────────────────────────────────────────────┘

Production Supabase Deployment Flow:
  ┌─────────────────────────────────────────────────────┐
  │ 1. Provision Supabase project + PostgreSQL DB     │
  │ 2. Export DATABASE_URL from Supabase console      │
  │ 3. Run: migrate -path ./migrations \              │
  │           -database $DATABASE_URL up              │
  │ 4. Deploy app (migrations already applied)        │
  └─────────────────────────────────────────────────────┘
```

### Component Interactions

```
cmd/server/main.go
  │
  ├─ config.Load()
  │   └─ Determine: SQLite vs PostgreSQL
  │
  └─ database.Initialize()
       │
       ├─ sqlite adapter
       │  ├─ OpenDB()
       │  └─ RunMigrations() ◄──── golang-migrate
       │
       └─ postgresql adapter
          ├─ OpenDB()
          └─ RunMigrations() ◄──── golang-migrate
```

---

## Implementation Strategy

### 1. Migration File Structure

**Naming Convention**: `NNN_description.{up,down}.sql`
- `NNN`: 3-digit version number (001–999)
- `description`: Human-readable summary
- `.up.sql`: Forward migration (apply schema change)
- `.down.sql`: Reverse migration (undo schema change)

**Example**:
```
migrations/
  ├── 001_init_schema.up.sql
  ├── 001_init_schema.down.sql
  ├── 002_optimize_indexes.up.sql
  ├── 002_optimize_indexes.down.sql
  └── ...
```

### 2. golang-migrate Integration

**Installation**:
```go
require github.com/golang-migrate/migrate/v4 v4.x.x
require github.com/golang-migrate/migrate/v4/database/postgres v4.x.x
require github.com/golang-migrate/migrate/v4/database/sqlite3 v4.x.x
require github.com/golang-migrate/migrate/v4/source/file v4.x.x
```

**Usage Pattern** (embedded in app):
```go
// In postgresql/db.go and sqlite/db.go
import "github.com/golang-migrate/migrate/v4"

func runMigrations(db *sql.DB, dbType string) error {
    driver := /* postgres or sqlite driver */
    m, err := migrate.NewWithDatabaseInstance(
        "file://./migrations",
        dbType,
        driver,
    )
    if err != nil {
        return err
    }

    // Auto-apply pending migrations
    err = m.Up()
    if err != nil && err != migrate.ErrNoChange {
        return err // Fail loudly
    }
    return nil
}
```

### 3. Supabase Deployment

**Manual Approach (shell script)**:
```bash
#!/bin/bash
# deploy-migrations.sh

DATABASE_URL=$1  # e.g., postgresql://user:pass@host/db

# Install migrate CLI if not present
if ! command -v migrate &> /dev/null; then
    echo "Installing golang-migrate..."
    go install -tags 'postgres' github.com/golang-migrate/migrate/cmd/migrate@latest
fi

# Apply migrations
migrate -path ./migrations \
        -database "$DATABASE_URL" \
        up

if [ $? -eq 0 ]; then
    echo "✓ Migrations applied successfully"
else
    echo "✗ Migration failed—check logs and roll back if needed"
    exit 1
fi
```

**Terraform Integration** (optional future enhancement):
```hcl
resource "null_resource" "database_migrations" {
  provisioner "local-exec" {
    command = "bash ./scripts/deploy-migrations.sh ${var.supabase_database_url}"
  }
  depends_on = [supabase_project.main]
}
```

### 4. Error Handling Strategy

| Scenario | Behavior | Rationale |
|----------|----------|-----------|
| Migration not found | Fail loudly | Prevents inconsistent state |
| Partial migration applied | Fail loudly (dirty state) | Operator must manually resolve |
| Rollback requested | Execute `.down.sql` | Safe recovery mechanism |
| No pending migrations | Success (no-op) | Idempotent behavior |

### 5. Migration Validation Checks

**Before Running**:
- [ ] All `.up.sql` files have corresponding `.down.sql`
- [ ] Migration numbers are sequential (no gaps)
- [ ] SQL syntax is valid for target database
- [ ] No conflicting schema changes detected

**After Running**:
- [ ] schema_migrations table exists and is populated
- [ ] All applied versions are present in table
- [ ] No dirty flag set (partial application)

---

## Rollout Plan

### Phase 1: Infrastructure Setup
1. Add golang-migrate dependency
2. Create `.down.sql` files for all existing migrations
3. Update database adapters to use golang-migrate
4. Add migration runner in both PostgreSQL and SQLite paths

### Phase 2: Testing
1. Integration tests (both databases)
2. Idempotency checks (run migrations twice)
3. Rollback validation
4. Dirty state recovery

### Phase 3: Documentation & Deployment
1. Create deployment script for Supabase
2. Document in README/DEPLOYMENT.md
3. Update CI/CD pipelines if needed
4. Run on staging first, then production

---

## Local Supabase Testing

### Setup

Developers can test migrations against a real PostgreSQL instance locally using Supabase CLI:

```bash
# Install Supabase CLI (one-time)
npm install -g supabase

# Start local Supabase stack (Docker required)
supabase start

# Print connection details
supabase status

# Output includes:
# DATABASE_URL=postgresql://postgres:postgres@localhost:54322/postgres
```

### Testing Workflow

```bash
# 1. Export local database URL
export DATABASE_URL="postgresql://postgres:postgres@localhost:54322/postgres"

# 2. Test migrations against local PostgreSQL
migrate -path ./migrations -database $DATABASE_URL up

# 3. Run integration tests
go test ./internal/adapter/repository/postgresql -v

# 4. Test rollback
migrate -path ./migrations -database $DATABASE_URL down

# 5. Clean up
supabase stop
```

### Benefits

- **Production Parity**: Test against actual PostgreSQL (not SQLite)
- **Feature Validation**: PostgreSQL-specific features work correctly
- **Rollback Testing**: Safe to test rollback procedures locally
- **CI/CD Ready**: Same tests run in local Supabase and production Supabase
- **Minimal Setup**: Docker handles all dependencies

---

## Trade-offs and Considerations

### Why golang-migrate over alternatives?

| Tool | Pros | Cons | Choice |
|------|------|------|--------|
| **golang-migrate** | Multi-DB, versioned, Go-native | Requires CLI for some uses | ✓ Selected |
| **Liquibase** | Very mature, XML/YAML | Heavyweight, Java dependency | ✗ |
| **Flyway** | Simple, popular | Java dependency | ✗ |
| **Custom solution** | Full control | Maintenance burden | ✗ |

### Why auto-run on startup?

- **Pro**: Schema always matches code; no manual steps during deployment
- **Con**: Increases startup time slightly; failure blocks app launch
- **Mitigation**: Pre-validation and rollback procedures

### Why fail loudly?

- **Pro**: Prevents silent data corruption; forces explicit resolution
- **Con**: Blocks production deployment on error
- **Mitigation**: Test migrations in staging first; maintain down migrations for emergency rollback

---

## Success Criteria

- [ ] golang-migrate integrated and tested locally
- [ ] All existing SQL files have `.down` counterparts
- [ ] Both SQLite and PostgreSQL support versioned migrations
- [ ] Migrations auto-run on startup (configurable)
- [ ] Supabase deployment script provided
- [ ] Integration tests cover both databases
- [ ] Documentation updated (README, DEPLOYMENT.md)
- [ ] Dirty state recovery procedure documented
- [ ] CI/CD validates migration syntax before merge

