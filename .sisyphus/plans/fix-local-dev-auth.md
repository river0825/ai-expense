# Fix Local Development Authentication

## TL;DR

> **Quick Summary**: The frontend sends `token=test-user` in local development, but the backend strictly validates JWTs, causing 401 errors. We need to allow a bypass for "test-user" when running in development mode.
>
> **Deliverables**:
> - Updated `config.go` with Environment support
> - Updated `handler.go` with dev-mode bypass logic
> - Updated `main.go` to wire it up
>
> **Estimated Effort**: Quick
> **Parallel Execution**: NO
> **Critical Path**: Task 1 -> Task 2 -> Task 3

---

## Context

### Original Request
User reported `curl 'http://localhost:8080/api/user?token=test-user'` fails because of strict JWT validation introduced in `fix-category-auth`.

### Problem
- Frontend defaults to `test-user` string when no token is present (dev mode).
- Backend `validateToken` strictly parses JWTs.
- `test-user` is not a valid JWT.

### Solution
- Add `Environment` configuration (default: dev).
- In `validateToken`, if `Environment == "dev"` AND token is `test-user`, return `test-user` ID directly without JWT parsing.

---

## Verification Strategy

### Automated Tests
- None required for this specific hotfix, but manual verification is mandatory.

### Agent-Executed QA Scenarios (MANDATORY)

```
Scenario: Verify test-user bypass in dev mode
  Tool: Bash (curl)
  Preconditions: Server running in dev mode (default)
  Steps:
    1. curl -s "http://localhost:8080/api/user/categories?token=test-user"
    2. Assert: HTTP 200 (or at least not 401)
    3. Assert: Returns categories list
  Expected Result: Success
  Evidence: Response body captured
```

---

## TODOs

- [ ] 1. Add Environment to Config
  **What to do**:
  - Modify `internal/config/config.go`
  - Add `Environment` string field to `Config` struct
  - Load it from `APP_ENV` environment variable (default: "dev")

- [ ] 2. Update Handler to Support Dev Mode
  **What to do**:
  - Modify `internal/adapter/http/handler.go`
  - Add `isDev bool` field to `Handler` struct
  - Update `NewHandler` signature to accept `isDev bool`
  - Update `validateToken` method:
    ```go
    if h.isDev && tokenString == "test-user" {
        return "test-user", nil
    }
    ```

- [ ] 3. Update Main Entrypoint
  **What to do**:
  - Modify `cmd/server/main.go`
  - Determine `isDev` from `cfg.Environment` (e.g. `cfg.Environment == "dev"`)
  - Pass `isDev` to `httpAdapter.NewHandler`

- [ ] 4. Verify Fix
  **What to do**:
  - Rebuild server: `go build -o server ./cmd/server`
  - Start server
  - Run curl check with `token=test-user`

---

## Success Criteria
- `curl ... token=test-user` returns 200 OK.
- Production security is maintained (bypass only enabled if explicitly in dev mode/config).
