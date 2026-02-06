# Fix Category API Authentication

## TL;DR

> **Quick Summary**: Update `/api/user/categories` endpoints to validate JWT tokens instead of accepting raw user_id, matching the auth pattern of other `/api/user/*` endpoints.
> 
> **Deliverables**:
> - Backend: 5 category handlers updated with JWT token validation
> - Frontend: HttpCategoryRepository updated to pass token correctly
> 
> **Estimated Effort**: Quick
> **Parallel Execution**: NO - sequential (backend first, then frontend)
> **Critical Path**: Task 1 → Task 2

---

## Context

### Original Request
The frontend is calling `/api/user/categories?user_id=<JWT_TOKEN>` but the backend expects a raw `user_id`. The token should be validated as a JWT to extract the real user_id, matching the pattern used by:
- `/api/user` (GetUserSettings)
- `/api/user/settings` (UpdateUserSettings)  
- `/api/reports/summary` (GetReportSummary)

### Current Behavior (Bug)
```
Frontend: GET /api/user/categories?user_id=eyJhbGc...  (sends JWT as user_id)
Backend:  Uses "eyJhbGc..." literally as user_id → No categories found
```

### Expected Behavior (Fix)
```
Frontend: GET /api/user/categories?token=eyJhbGc...
Backend:  Validates JWT → Extracts real user_id → Returns categories
```

---

## Work Objectives

### Core Objective
Enable category endpoints to authenticate users via JWT token, consistent with other `/api/user/*` endpoints.

### Concrete Deliverables
- `internal/adapter/http/handler.go` - 5 category handlers updated
- `frontend/dashboard/src/infrastructure/repositories/http/HttpCategoryRepository.ts` - API calls updated

### Definition of Done
- [x] `curl /api/user/categories?token=<JWT>` returns categories
- [x] `curl -X POST /api/user/categories -d '{"token":"<JWT>","name":"Test"}'` creates category
- [x] Dashboard settings page loads and displays categories correctly
- [x] All existing tests pass

### Must Have
- JWT validation on all 5 category endpoints
- Backward compatibility with `user_id` param (for internal/testing use)
- Consistent error messages for invalid tokens

### Must NOT Have (Guardrails)
- ❌ Changes to other endpoints
- ❌ New authentication mechanisms
- ❌ Changes to JWT secret or validation logic (reuse existing `validateToken`)

---

## Verification Strategy

### Test Decision
- **Infrastructure exists**: YES (go test)
- **Automated tests**: Tests-after (verify existing tests pass)
- **Framework**: go test

### Agent-Executed QA Scenarios (MANDATORY)

```
Scenario: List categories with valid JWT token
  Tool: Bash (curl)
  Preconditions: Server running, test user exists with categories
  Steps:
    1. Get a valid JWT token (via /api/r/{shortcode} flow or existing test token)
    2. curl -s "http://localhost:8080/api/user/categories?token=${TOKEN}"
    3. Assert: HTTP 200
    4. Assert: response.status == "success"
    5. Assert: response.data is array of categories
  Expected Result: Categories returned for authenticated user
  Evidence: Response body captured

Scenario: List categories with invalid token returns 401
  Tool: Bash (curl)
  Preconditions: Server running
  Steps:
    1. curl -s "http://localhost:8080/api/user/categories?token=invalid-token"
    2. Assert: HTTP 401
    3. Assert: response.error contains "Invalid token"
  Expected Result: Unauthorized error
  Evidence: Response body captured

Scenario: Create category with valid JWT token
  Tool: Bash (curl)
  Preconditions: Server running, valid JWT token
  Steps:
    1. curl -s -X POST "http://localhost:8080/api/user/categories" \
         -H "Content-Type: application/json" \
         -d '{"token":"${TOKEN}","name":"Test Category","description":"Test"}'
    2. Assert: HTTP 200
    3. Assert: response.status == "success"
    4. Assert: response.data.name == "Test Category"
  Expected Result: Category created successfully
  Evidence: Response body captured

Scenario: Dashboard Settings page loads categories
  Tool: Playwright (playwright skill)
  Preconditions: Dev servers running (backend :8080, frontend :3000), user authenticated
  Steps:
    1. Navigate to dashboard with valid token
    2. Navigate to Settings page
    3. Wait for Category Management section to load
    4. Assert: Category list is visible (not empty or error)
  Expected Result: Categories display correctly
  Evidence: Screenshot captured
```

---

## TODOs

- [x] 1. Update backend category handlers to validate JWT tokens

  **What to do**:
  1. Update `ListCategories` handler:
     - Add `token := r.URL.Query().Get("token")`
     - If `userID == ""` and `token != ""`, call `h.validateToken(token)` to get userID
     - Return 401 if token validation fails
     - Update error message: "user_id or token is required"
  
  2. Update `CreateCategory` handler:
     - Add `Token string` field to `CreateCategoryRequest` struct
     - If `req.UserID == ""` and `req.Token != ""`, validate token
     - Use extracted userID for the usecase call
  
  3. Update `UpdateCategory` handler (same pattern as CreateCategory)
  
  4. Update `DeleteCategory` handler (same pattern as CreateCategory)
  
  5. Update `MergeCategories` handler (same pattern as CreateCategory)

  **Must NOT do**:
  - Don't change the `validateToken` function (it already exists and works)
  - Don't remove `user_id` support (keep for backward compatibility)

  **Recommended Agent Profile**:
  - **Category**: `quick`
  - **Skills**: [`git-master`]

  **Parallelization**:
  - **Can Run In Parallel**: NO
  - **Parallel Group**: Sequential
  - **Blocks**: Task 2
  - **Blocked By**: None

  **References**:

  **Pattern References** (existing code to follow):
  - `internal/adapter/http/handler.go:1431-1463` - `GetUserSettings` handler shows exact pattern to follow
  - `internal/adapter/http/handler.go:1556-1583` - `validateToken` function to reuse

  **Implementation Details**:
  - `ListCategories` is at line 551-571
  - `CreateCategory` is at line 443-477
  - `UpdateCategory` is at line 479-516
  - `DeleteCategory` is at line 518-549
  - `MergeCategories` is at line 573-606

  **Acceptance Criteria**:
  - [ ] `ListCategories` validates JWT token from `?token=` query param
  - [ ] `CreateCategory` validates JWT token from request body `token` field
  - [ ] `UpdateCategory` validates JWT token from request body `token` field
  - [ ] `DeleteCategory` validates JWT token from request body `token` field
  - [ ] `MergeCategories` validates JWT token from request body `token` field
  - [ ] All handlers return 401 with "Invalid token" for bad tokens
  - [ ] All handlers still accept `user_id` for backward compatibility
  - [ ] `go build ./...` succeeds
  - [ ] `go test ./internal/adapter/http/...` passes

  **Commit**: YES
  - Message: `fix(backend): add JWT token validation to category endpoints`
  - Files: `internal/adapter/http/handler.go`
  - Pre-commit: `go build ./... && go test ./internal/adapter/http/... -v`

---

- [x] 2. Update frontend HttpCategoryRepository to pass token correctly

  **What to do**:
  1. Update `list()` method:
     - Change from `params: { user_id: token }` to `params: { token: token }`
  
  2. Update `create()` method:
     - Change from `user_id: token` to `token: token` in request body
  
  3. Update `update()` method:
     - Change from `user_id: token` to `token: token` in request body
  
  4. Update `delete()` method:
     - Change from `user_id: token` to `token: token` in request body
  
  5. Update `merge()` method:
     - Change from `user_id: token` to `token: token` in request body

  **Must NOT do**:
  - Don't change the method signatures (still accept `token: string` parameter)
  - Don't change how the repository is used in components

  **Recommended Agent Profile**:
  - **Category**: `quick`
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: NO
  - **Parallel Group**: Sequential (after Task 1)
  - **Blocks**: None
  - **Blocked By**: Task 1

  **References**:

  **File to modify**:
  - `frontend/dashboard/src/infrastructure/repositories/http/HttpCategoryRepository.ts`

  **Acceptance Criteria**:
  - [ ] `list()` sends `?token=<jwt>` instead of `?user_id=<jwt>`
  - [ ] `create()` sends `{"token":"<jwt>",...}` instead of `{"user_id":"<jwt>",...}`
  - [ ] `update()` sends `{"token":"<jwt>",...}` instead of `{"user_id":"<jwt>",...}`
  - [ ] `delete()` sends `{"token":"<jwt>",...}` instead of `{"user_id":"<jwt>",...}`
  - [ ] `merge()` sends `{"token":"<jwt>",...}` instead of `{"user_id":"<jwt>",...}`
  - [ ] `cd frontend/dashboard && bun run build` succeeds

  **Commit**: YES
  - Message: `fix(dashboard): pass JWT token correctly to category API`
  - Files: `frontend/dashboard/src/infrastructure/repositories/http/HttpCategoryRepository.ts`
  - Pre-commit: `cd frontend/dashboard && bun run build`

---

## Commit Strategy

| After Task | Message | Files | Verification |
|------------|---------|-------|--------------|
| 1 | `fix(backend): add JWT token validation to category endpoints` | handler.go | go build && go test |
| 2 | `fix(dashboard): pass JWT token correctly to category API` | HttpCategoryRepository.ts | bun run build |

---

- [x] 3. Fix missing ReassignExpenses method in test repositories

  **What to do**:
  The `ExpenseRepository` interface now includes `ReassignExpenses` method (added in Wave 1 of edit-category-page plan), but test repositories don't implement it.
  
  Add stub implementations to:
  1. `test/load/load_test.go` - `LoadTestExpenseRepository`
  2. `test/bench/usecase_bench_test.go` - `BenchExpenseRepository`
  3. `test/e2e/webhook_flow_test.go` - `E2EExpenseRepository`

  **Implementation** (same for all):
  ```go
  func (r *XxxExpenseRepository) ReassignExpenses(ctx context.Context, userID, fromCategoryID, toCategoryID string) (int, error) {
      return 0, nil // Stub - not used in these tests
  }
  ```

  **Must NOT do**:
  - Don't implement actual logic (not needed for these tests)

  **Recommended Agent Profile**:
  - **Category**: `quick`
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES (with Task 1)
  - **Parallel Group**: Wave 1 (with Task 1)
  - **Blocks**: None
  - **Blocked By**: None

  **References**:
  - `internal/domain/repositories.go` - Interface definition with `ReassignExpenses`
  - `internal/adapter/repository/postgresql/expense_repo.go` - Real implementation to reference

  **Acceptance Criteria**:
  - [ ] `LoadTestExpenseRepository` implements `ReassignExpenses`
  - [ ] `BenchExpenseRepository` implements `ReassignExpenses`
  - [ ] `E2EExpenseRepository` implements `ReassignExpenses`
  - [ ] `go build ./...` succeeds (no interface errors)

  **Commit**: YES
  - Message: `fix(test): add ReassignExpenses stub to test repositories`
  - Files: `test/load/load_test.go`, `test/bench/usecase_bench_test.go`, `test/e2e/webhook_flow_test.go`
  - Pre-commit: `go build ./...`

---

## Success Criteria

### Verification Commands
```bash
# Backend builds
go build ./...

# Backend tests pass
go test ./internal/adapter/http/... -v

# Frontend builds
cd frontend/dashboard && bun run build

# Manual verification with curl
TOKEN="<get-valid-jwt>"
curl "http://localhost:8080/api/user/categories?token=${TOKEN}"
```

### Final Checklist
- [x] All category endpoints validate JWT tokens
- [x] Frontend passes token in correct field
- [x] Dashboard Settings page displays categories
- [x] No regression in existing functionality
