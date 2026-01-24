# VSCode Debugging Setup

## Quick Start

The `.vscode/launch.json` has been configured with two debug configurations:

### 1. SQLite (Default)
**Name**: `Launch (SQLite)`
- Uses local SQLite database (`./aiexpense.db`)
- Simple setup, no external dependencies
- Migrations run automatically on startup
- Press `F5` to start debugging

**Environment**:
- `DATABASE_PATH=./aiexpense.db`
- `AI_PROVIDER=gemini`
- `SERVER_PORT=8080`

### 2. PostgreSQL/Supabase
**Name**: `Launch (PostgreSQL/Supabase)`
- Uses local Supabase instance (requires Supabase CLI)
- Tests against real PostgreSQL schema
- Automatically starts Supabase before debugging

**Setup Steps**:
1. Install Supabase CLI:
   ```bash
   brew install supabase/tap/supabase
   ```

2. Update `.env.local.example` to `.env` and set required keys:
   ```bash
   cp .env.local.example .env
   ```

3. In VSCode, select debug config: `Launch (PostgreSQL/Supabase)`

4. Press `F5` to start:
   - Automatically starts Supabase
   - Runs migrations
   - Debugger attaches to server

5. When done, manually stop Supabase:
   ```bash
   supabase stop --project-id aiexpense
   ```

## Troubleshooting

### "Failed to open source, file://migrations: open .: no such file or directory"
This means the working directory is wrong. The fix is already applied in `.vscode/launch.json` with `"cwd": "${workspaceFolder}"`.

### "Address already in use :8080"
Kill the existing process:
```bash
lsof -i :8080 | grep LISTEN | awk '{print $2}' | xargs kill -9
```

### Supabase won't start
Check if another instance is running:
```bash
supabase status
supabase stop --project-id aiexpense
```

## Manual Environment Setup

If you prefer to manually set environment variables instead of using `.env`:

### SQLite
```bash
export GEMINI_API_KEY=dev-key
export AI_PROVIDER=gemini
export SERVER_PORT=8080
export DATABASE_PATH=./aiexpense.db
```

### PostgreSQL/Supabase
```bash
# Start Supabase first
supabase start

# Get connection string
supabase status  # Copy the PostgreSQL URL

# Set environment
export DATABASE_URL="postgresql://postgres:postgres@127.0.0.1:54322/postgres?sslmode=disable"
export GEMINI_API_KEY=dev-key
export AI_PROVIDER=gemini
export SERVER_PORT=8080
```

## File Structure

- `.vscode/launch.json` - Debug configurations (SQLite and PostgreSQL)
- `.vscode/tasks.json` - VSCode tasks (Supabase start/stop)
- `.env.local.example` - Environment variable template
- `migrations/` - Database migration files (.up.sql and .down.sql)

## Debugging Features

- **Breakpoints**: Click on line numbers to set breakpoints
- **Step Over** (F10): Execute one line
- **Step Into** (F11): Enter function calls
- **Evaluate**: Hover over variables to see values
- **Debug Console**: Type expressions to evaluate at breakpoint
- **Call Stack**: View function call hierarchy

## Testing Migrations

### SQLite
Migrations run automatically on startup. Check logs for:
```
Migrations applied successfully
No new migrations to apply
```

### PostgreSQL
Same behavior, but uses Supabase PostgreSQL. Verify schema:
```bash
psql "postgresql://postgres:postgres@127.0.0.1:54322/postgres?sslmode=disable" -c "SELECT * FROM schema_migrations;"
```

## Notes

- Environment variables in `env` section override `.env` file
- `cwd` is set to workspace root to find `migrations/` directory
- Supabase task runs with `runOn: folderOpen` but can be triggered manually
- Debug mode includes full stack traces and symbol information
