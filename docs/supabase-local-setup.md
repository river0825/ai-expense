# Local Supabase Setup Guide

This guide explains how to set up and use a local Supabase instance for development. This allows you to test the application against real PostgreSQL before deploying to production Supabase.

## Table of Contents

- [What is Local Supabase?](#what-is-local-supabase)
- [Prerequisites](#prerequisites)
- [Installation](#installation)
- [Starting Local Supabase](#starting-local-supabase)
- [Running the Application](#running-the-application)
- [Connecting to Local Supabase](#connecting-to-local-supabase)
- [Managing Local Data](#managing-local-data)
- [Stopping Local Supabase](#stopping-local-supabase)
- [Troubleshooting](#troubleshooting)

---

## What is Local Supabase?

Local Supabase runs a complete PostgreSQL database and supporting services on your machine, mimicking the cloud Supabase experience. This is useful for:

- **Development**: Test against real PostgreSQL before cloud deployment
- **Testing**: Run integration tests with actual database behavior
- **Offline work**: Develop without internet connection
- **Cost**: No cloud charges during development
- **Learning**: Understand Supabase architecture locally

---

## Prerequisites

### System Requirements

- **macOS**: 10.15+ (Catalina or newer)
- **Linux**: Ubuntu 18.04+ or equivalent
- **Windows**: WSL2 (Windows Subsystem for Linux)
- **Disk Space**: ~2-3 GB for Supabase containers
- **RAM**: 4GB+ recommended for smooth operation
- **Docker**: Required to run Supabase locally

### Install Docker

**macOS:**
```bash
# Using Homebrew
brew install docker

# Or download Docker Desktop from: https://www.docker.com/products/docker-desktop
```

**Linux (Ubuntu):**
```bash
sudo apt-get update
sudo apt-get install -y docker.io docker-compose
sudo usermod -aG docker $USER
# Log out and log back in for group changes to take effect
```

**Windows (WSL2):**
```bash
# Install WSL2 first, then Docker Desktop with WSL2 backend
# See: https://docs.docker.com/desktop/install/windows-install/
```

---

## Installation

### Step 1: Install Supabase CLI

**macOS:**
```bash
brew install supabase/tap/supabase
```

**Linux:**
```bash
# Using Homebrew on Linux
brew install supabase/tap/supabase

# Or download from GitHub releases
curl -fsSL https://github.com/supabase/cli/releases/download/v1.160.1/supabase_1.160.1_linux_arm64.tar.gz | tar xz
sudo mv supabase /usr/local/bin/
```

**Windows (WSL2):**
```bash
# From within WSL2 terminal
brew install supabase/tap/supabase
```

### Step 2: Verify Installation

```bash
supabase --version
# Should output: supabase version X.X.X
```

---

## Starting Local Supabase

### First Time Setup

Initialize Supabase in your project:

```bash
# From project root (where this file is located)
cd /path/to/aiexpense

# Link to your Supabase project (optional, for later cloud sync)
# supabase link --project-ref your-project-ref

# Start local Supabase instance
supabase start

# Wait for output like:
# Local development started successfully.
# ...
# API URL: http://localhost:54321
# DB URL: postgresql://postgres:postgres@localhost:54322/postgres
```

### Subsequent Starts

```bash
# Just start it (uses existing containers)
supabase start

# Or restart from scratch (delete all local data)
supabase stop
supabase reset
supabase start
```

---

## Running the Application

### Step 1: Get Connection String

```bash
# Get Supabase status and connection details
supabase status

# Look for this in the output:
# DATABASE_URL: postgresql://postgres:postgres@localhost:54322/postgres
```

### Step 2: Set Environment Variable

**Option A: Temporary (current terminal session)**
```bash
export DATABASE_URL="postgresql://postgres:postgres@localhost:54322/postgres"
```

**Option B: Permanent (add to .env.local)**
```bash
# Copy example environment file
cp .env.local.example .env.local

# Edit .env.local and uncomment/update DATABASE_URL:
# DATABASE_URL=postgresql://postgres:postgres@localhost:54322/postgres
```

### Step 3: Run Application

```bash
# If using temporary export
go run ./cmd/server/main.go

# Or if using .env.local, set other required variables too:
export GEMINI_API_KEY="your-api-key"
go run ./cmd/server/main.go
```

### Step 4: Verify Migrations

In another terminal, verify migrations ran:

```bash
# Get the connection string again
supabase_db="postgresql://postgres:postgres@localhost:54322/postgres"

# Check schema_migrations table
psql "$supabase_db" -c "SELECT version, dirty FROM schema_migrations ORDER BY version;"

# List all tables
psql "$supabase_db" -c "\dt public.*"

# Sample data queries
psql "$supabase_db" -c "SELECT COUNT(*) as user_count FROM users;"
psql "$supabase_db" -c "SELECT COUNT(*) as expense_count FROM expenses;"
```

---

## Connecting to Local Supabase

### Using psql (PostgreSQL CLI)

```bash
# Get connection string
PGPASSWORD=postgres psql -h localhost -U postgres -d postgres -p 54322

# Or with DATABASE_URL
psql "postgresql://postgres:postgres@localhost:54322/postgres"

# Common queries once connected
\dt public.*            -- List all tables
SELECT * FROM schema_migrations;  -- View migration history
SELECT COUNT(*) FROM users;       -- Count users
\q                      -- Exit psql
```

### Using pgAdmin (Web UI)

```bash
# Supabase starts pgAdmin automatically
# Access at: http://localhost:54323

# Login credentials:
# Email: supabase
# Password: (check 'supabase status' output)
```

### Using Supabase Studio (Dashboard)

```bash
# Supabase Studio is started automatically
# Access at: http://localhost:54323

# Navigate to:
# - SQL Editor: Run custom queries
# - Table Editor: Browse and edit data
# - Authentication: Manage users
```

---

## Managing Local Data

### Insert Test Data

```bash
psql "postgresql://postgres:postgres@localhost:54322/postgres" << EOF
-- Create a test user
INSERT INTO users (user_id, messenger_type, created_at)
VALUES ('user_001', 'telegram', CURRENT_TIMESTAMP);

-- Create a test category
INSERT INTO categories (id, user_id, name, is_default, created_at)
VALUES ('cat_001', 'user_001', 'Food', FALSE, CURRENT_TIMESTAMP);

-- Create a test expense
INSERT INTO expenses (id, user_id, description, amount, category_id, expense_date, created_at, updated_at)
VALUES ('exp_001', 'user_001', 'Lunch', 15.50, 'cat_001', CURRENT_DATE, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP);
EOF
```

### Export Data

```bash
# Backup entire database
pg_dump "postgresql://postgres:postgres@localhost:54322/postgres" > backup.sql

# Backup specific table
pg_dump -t users "postgresql://postgres:postgres@localhost:54322/postgres" > users_backup.sql
```

### Clear All Data

```bash
# Warning: This deletes everything!
supabase reset

# Confirm with 'y' when prompted
```

---

## Stopping Local Supabase

### Pause (Keep Data)

```bash
# Stop containers but preserve data
supabase stop

# Resume later with 'supabase start'
```

### Reset (Delete Data)

```bash
# Stop and delete all data (including Docker volumes)
supabase stop
supabase reset

# Or just delete one component:
# docker compose down -v  # Delete all volumes
```

---

## Troubleshooting

### Problem: "Port 54322 already in use"

**Cause**: Supabase or another PostgreSQL instance already using the port.

**Solution**:
```bash
# Option 1: Find and stop existing Supabase
supabase stop

# Option 2: Find process using port 54322 (macOS)
lsof -i :54322
kill -9 <PID>

# Option 3: Use different port (advanced)
# Edit supabase/config.toml and change port numbers
```

### Problem: "Cannot connect to docker daemon"

**Cause**: Docker Desktop not running or not installed.

**Solution**:
```bash
# Start Docker Desktop (macOS)
open /Applications/Docker.app

# Or restart Docker daemon (Linux)
sudo systemctl start docker

# Verify Docker is running
docker ps
```

### Problem: "Connection refused on localhost:54322"

**Cause**: Supabase containers not fully started.

**Solution**:
```bash
# Check status
supabase status

# Wait 30 seconds for containers to be ready
# Then try again

# Or restart fresh
supabase stop
supabase start
```

### Problem: "Database migrations failed"

**Cause**: Migrations couldn't apply to local Supabase PostgreSQL.

**Solution**:
1. Check Supabase is running: `supabase status`
2. Verify connection: `psql "postgresql://postgres:postgres@localhost:54322/postgres" -c "SELECT 1;"`
3. Check migration files exist: `ls migrations/*.up.sql`
4. Try resetting: `supabase reset`

### Problem: "No schema_migrations table"

**Cause**: Migrations haven't run yet.

**Solution**:
```bash
# Ensure app was started and migrations ran
go run ./cmd/server/main.go

# Check the logs for "Migrations applied successfully"
# Verify table was created:
psql "postgresql://postgres:postgres@localhost:54322/postgres" -c "SELECT * FROM schema_migrations;"
```

### Problem: "pgAdmin login fails"

**Cause**: Credentials incorrect or pgAdmin not started.

**Solution**:
```bash
# Check Supabase status for correct credentials
supabase status

# Restart Supabase
supabase stop
supabase start

# Default credentials:
# Email: supabase
# Password: check 'supabase status' output
```

### Problem: "Out of disk space"

**Cause**: Docker images and containers consuming too much space.

**Solution**:
```bash
# Clean up Docker (caution: removes unused images/containers)
docker system prune -a

# Or manually remove Supabase
supabase stop
docker system prune
```

---

## Performance Tips

1. **Use SQLite for quick testing**: Local Supabase is more realistic but slower than SQLite
2. **Close unused applications**: Docker benefits from available RAM
3. **Limit connections**: Connection pools can consume resources
4. **Clear old data**: Periodically `supabase reset` to start fresh
5. **Monitor resources**: `docker stats` to see container resource usage

---

## Next Steps

- Read [migrations.md](./migrations.md) for database schema management
- Check `.env.local.example` for all configuration options
- Review [README.md](../README.md) for project overview

---

## References

- [Supabase CLI Documentation](https://supabase.com/docs/guides/cli)
- [Supabase Local Development](https://supabase.com/docs/guides/cli/local-development)
- [PostgreSQL Documentation](https://www.postgresql.org/docs/)
- [Docker Documentation](https://docs.docker.com/)
- [pgAdmin Documentation](https://www.pgadmin.org/docs/)
