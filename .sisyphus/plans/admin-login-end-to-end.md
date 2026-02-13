# Fix Admin Login End-to-End

## TL;DR

> **Quick Summary**: Wire the existing admin auth building blocks into a working end-to-end flow so `frontend/admin` users can sign in, access protected analytics, and be redirected on unauthorized access.
>
> **Deliverables**:
> - Registered backend admin auth + admin analytics routes under `/api/admin/*`
> - Working login form submission in `frontend/admin` with token persistence
> - Protected dashboard behavior and stable unauthorized handling
> - CI-verifiable tests (Go + Vitest + Playwright) for login and protected route behavior
>
> **Estimated Effort**: Medium
> **Parallel Execution**: YES - 3 waves
> **Critical Path**: Task 1 -> Task 2 -> Task 3 -> Task 5

---

## Context

### Original Request
"How to login the admin panel?" followed by request to create a focused work plan to make admin login work end-to-end.

### Interview Summary
**Key Discussions**:
- Admin frontend exists in `frontend/admin` and has `/login` + `/dashboard` pages.
- Current login page is static UI and does not submit to backend.
- Token helper exists (`admin_token` in localStorage) and API helper attaches Bearer token.
- Backend auth handler/usecase exists but is not fully wired into app routes.

**Research Findings**:
- `frontend/admin/src/app/login/page.tsx`: form-only, missing submit logic.
- `frontend/admin/src/lib/auth.ts`: token set/get/remove utilities.
- `frontend/admin/src/lib/api.ts`: redirects to `/login` on 401.
- `internal/adapter/http/admin_auth_handler.go`: login/logout/verify/middleware exist.
- `internal/usecase/admin_auth.go`: dev credentials `admin` / `admin123`, JWT issuance.
- `internal/adapter/http/handler.go`: admin routes are not registered.

### Gap Review (Oracle)
**Identified Gaps (addressed in this plan)**:
- Route-registration gap: admin handlers exist but are not attached to the server mux.
- Wiring gap: repository/usecase/handler bootstrap path for admin auth must be fully connected.
- Contract mismatch risk: frontend API base URL semantics can break CI.
- Stability gap: token-verify nil-session path needs explicit safe handling to avoid 500/panic.

---

## Work Objectives

### Core Objective
Deliver a reliable admin login experience where valid credentials unlock protected admin analytics and invalid/expired auth paths fail safely with deterministic redirects/errors.

### Concrete Deliverables
- Backend route registration for admin auth and analytics endpoints.
- Frontend login form submission + token persistence + redirect.
- Dashboard guard behavior for unauthenticated access.
- Passing automated verification for login happy/failure flows.

### Definition of Done
- [ ] `POST /api/admin/auth/login` returns token on valid credentials and 401 on invalid credentials.
- [ ] Protected admin analytics endpoint returns 401 when token is missing/invalid.
- [ ] `frontend/admin/login` signs in and redirects to `/dashboard`.
- [ ] Unauthenticated access to `/dashboard` lands on `/login`.
- [ ] `go test ./internal/usecase ./internal/adapter/http` passes for affected auth/route tests.
- [ ] `cd frontend/admin && bun run test` passes.
- [ ] `cd frontend/admin && bunx playwright test e2e/admin.spec.ts` passes.

### Must Have
- Keep authentication model JWT-based for admin panel.
- Keep current dev credentials behavior unchanged for this fix scope.
- Prevent unauthorized access to protected admin analytics routes.
- No panic/500 path from invalid token/session mismatch.

### Must NOT Have (Guardrails)
- No RBAC/IAM or admin user-management expansion in this plan.
- No full credential storage redesign (hashing/user table migration) in this plan.
- No unrelated refactor of legacy dashboard (`frontend/dashboard`) or metrics API-key flow.

---

## Verification Strategy (MANDATORY)

> **UNIVERSAL RULE: ZERO HUMAN INTERVENTION**
>
> Every acceptance criterion must be executable by commands/tools.

### Test Decision
- **Infrastructure exists**: YES
- **Automated tests**: TDD
- **Framework**: Go test + Vitest + Playwright

### Agent-Executed QA Scenarios (all tasks)
- Backend/API: `curl` for auth login/verify/protected-route status and payload assertions.
- Frontend/UI: Playwright for login submission, redirect, and protected route checks.
- Build/test: Bash commands with explicit pass/fail outcomes.

---

## Execution Strategy

### Dev Process Compliance (MANDATORY)

- Before implementation, follow `AGENTS.md` workflow exactly:
  1. Create GitHub issue (`gh issue create`) with scope + acceptance criteria.
  2. Create linked branch (`gh issue develop <issue-number> --checkout`).
  3. Create/use worktree under `.worktrees/<branch-name>`.
  4. Implement via TDD inside that worktree.
  5. Commit using conventional commits, push, and open PR (`gh pr create`).
  6. Address review feedback with follow-up commits (no force-push during review).
- This workflow is part of Definition of Done for this plan.

### Parallel Execution Waves

Wave 1 (Backend foundation)
- Task 1: Wire admin auth dependencies and route registration
- Task 2: Harden token/session verification path

Wave 2 (Frontend auth flow)
- Task 3: Implement login form submission and token persistence
- Task 4: Add dashboard auth guard behavior

Wave 3 (Integration and quality gates)
- Task 5: Align API URL contract and integration tests
- Task 6: E2E and CI verification for admin login flow

Critical Path: 1 -> 2 -> 3 -> 5

### Dependency Matrix

| Task | Depends On | Blocks | Can Parallelize With |
|------|------------|--------|----------------------|
| 1 | None | 3,5,6 | 2 |
| 2 | None | 5,6 | 1 |
| 3 | 1 | 4,5,6 | None |
| 4 | 3 | 6 | 5 |
| 5 | 1,2,3 | 6 | 4 |
| 6 | 4,5 | None | None |

---

## TODOs

- [ ] 1. Register admin auth and analytics routes in backend bootstrap

  **What to do**:
  - Wire `AdminAuthRepository` adapter and create login/verify/logout usecases in server bootstrap.
  - Register admin auth endpoints (`/api/admin/auth/login`, `/api/admin/auth/verify`, `/api/admin/auth/logout`).
  - Register protected analytics endpoint(s) under `/api/admin/analytics/*` with auth middleware.

  **Must NOT do**:
  - Reuse legacy `X-API-Key` checks for this new admin JWT flow.
  - Change unrelated public API routes.

  **Recommended Agent Profile**:
  - **Category**: `deep`
    - Reason: backend composition + route wiring + security boundary correctness.
  - **Skills**: `git-master`
    - `git-master`: keep route and bootstrap changes reviewable and atomic.

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1
  - **Blocks**: 3,5,6
  - **Blocked By**: None

  **References**:
  - `internal/adapter/http/admin_auth_handler.go:45` - login handler contract and status codes.
  - `internal/adapter/http/admin_auth_handler.go:122` - middleware contract for protected routes.
  - `internal/adapter/http/admin_analytics_handler.go:21` - analytics overview endpoint intent.
  - `internal/adapter/http/handler.go:1720` - existing route registration style via `mux.HandleFunc`.
  - `cmd/server/main.go` - current dependency wiring location for handlers/usecases.

  **Acceptance Criteria**:
  - [ ] `curl -s -o /dev/null -w "%{http_code}" -X POST http://localhost:8080/api/admin/auth/login` returns non-404.
  - [ ] Protected `/api/admin/analytics/overview` returns 401 without Bearer token.
  - [ ] Backend starts with registered admin routes and no wiring panic.

  **Agent-Executed QA Scenarios**:
  ```text
  Scenario: Admin auth routes are reachable
    Tool: Bash (curl)
    Preconditions: Backend running at localhost:8080
    Steps:
      1. POST /api/admin/auth/login with JSON credentials
      2. Assert: HTTP status is 200 or 401 (never 404)
      3. GET /api/admin/auth/verify without header
      4. Assert: HTTP status is 400 or 401
    Expected Result: Routes are registered and active
    Evidence: Terminal output capture

  Scenario: Protected analytics is actually protected
    Tool: Bash (curl)
    Preconditions: Backend running
    Steps:
      1. GET /api/admin/analytics/overview without Authorization
      2. Assert: HTTP status 401
      3. Assert: response JSON includes status=error
    Expected Result: No anonymous analytics access
    Evidence: Response body capture
  ```

- [ ] 2. Harden token verification and invalid-session handling

  **What to do**:
  - Ensure verify path returns explicit auth error when JWT is valid but session lookup fails.
  - Add tests covering malformed token, expired token, and missing session cases.

  **Must NOT do**:
  - Leave nil-session path that can cause panic/500.

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
    - Reason: auth edge-case correctness and failure-safety.
  - **Skills**: `git-master`

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1
  - **Blocks**: 5,6
  - **Blocked By**: None

  **References**:
  - `internal/usecase/admin_auth.go:114` - token verification flow.
  - `internal/adapter/http/admin_auth_handler.go:106` - verify handler expects safe error propagation.
  - `internal/usecase/admin_auth_test.go` - existing auth usecase test patterns.

  **Acceptance Criteria**:
  - [ ] Invalid/malformed token returns 401 from verify/auth middleware.
  - [ ] JWT-valid-but-no-session path returns 401, not 500.
  - [ ] `go test ./internal/usecase -run AdminAuth -v` passes.

  **Agent-Executed QA Scenarios**:
  ```text
  Scenario: Valid format token without session is rejected safely
    Tool: Bash (curl)
    Preconditions: Backend running with empty/no matching admin session
    Steps:
      1. Call protected endpoint with forged or stale Bearer token
      2. Assert: HTTP status 401
      3. Assert: response does not contain stack trace/panic text
    Expected Result: Safe auth failure path
    Evidence: Response body capture
  ```

- [ ] 3. Wire login page submit flow to backend and persist token

  **What to do**:
  - Convert login page to client-side submit handler.
  - POST credentials to admin auth login endpoint.
  - On success, store token via auth helper and navigate to `/dashboard`.
  - Render clear error message on failed login.

  **Must NOT do**:
  - Store credentials in localStorage.
  - Continue rendering static no-op form submit.

  **Recommended Agent Profile**:
  - **Category**: `visual-engineering`
    - Reason: UI state transitions + API interaction.
  - **Skills**: `frontend-ui-ux`

  **Parallelization**:
  - **Can Run In Parallel**: NO
  - **Parallel Group**: Wave 2
  - **Blocks**: 4,5,6
  - **Blocked By**: 1

  **References**:
  - `frontend/admin/src/app/login/page.tsx` - current login form structure.
  - `frontend/admin/src/lib/auth.ts:3` - token persistence helper (`admin_token`).
  - `frontend/admin/e2e/admin.spec.ts:17` - expected login behavior and redirect.
  - `internal/adapter/http/admin_auth_handler.go:61` - login success response shape.

  **Acceptance Criteria**:
  - [ ] Valid credentials (`admin/admin123`) redirect to `/dashboard`.
  - [ ] Invalid credentials show visible error and remain on `/login`.
  - [ ] Token is written under `admin_token` after success.
  - [ ] `cd frontend/admin && bun run test` passes.

  **Agent-Executed QA Scenarios**:
  ```text
  Scenario: Successful login stores token and navigates
    Tool: Playwright
    Preconditions: Frontend at localhost:3000, backend at localhost:8080
    Steps:
      1. Navigate to /login
      2. Fill input[name="username"] with "admin"
      3. Fill input[name="password"] with "admin123"
      4. Click submit/sign-in button
      5. Wait for URL /dashboard
      6. Evaluate localStorage key admin_token exists
    Expected Result: Logged-in state established
    Evidence: Screenshot + browser storage assertion output

  Scenario: Invalid login remains on login page
    Tool: Playwright
    Preconditions: Frontend and backend running
    Steps:
      1. Navigate to /login
      2. Submit username "admin" and wrong password
      3. Assert: URL remains /login
      4. Assert: error text visible
      5. Assert: localStorage admin_token absent
    Expected Result: Invalid auth is blocked cleanly
    Evidence: Screenshot path under .sisyphus/evidence/
  ```

- [ ] 4. Enforce dashboard access guard for unauthenticated users

  **What to do**:
  - Add guard to redirect to `/login` when no valid token exists.
  - Ensure 401 from analytics fetch also routes to login without infinite loops.

  **Must NOT do**:
  - Allow dashboard rendering without auth check.
  - Create redirect loop between `/login` and `/dashboard`.

  **Recommended Agent Profile**:
  - **Category**: `visual-engineering`
  - **Skills**: `frontend-ui-ux`, `playwright`

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 2
  - **Blocks**: 6
  - **Blocked By**: 3

  **References**:
  - `frontend/admin/src/app/dashboard/page.tsx` - current dashboard render flow.
  - `frontend/admin/src/lib/api.ts:27` - existing 401 redirect behavior.
  - `frontend/admin/e2e/admin.spec.ts:5` - expected redirect when unauthenticated.

  **Acceptance Criteria**:
  - [ ] Visiting `/dashboard` without token redirects to `/login`.
  - [ ] Expired/invalid token causes redirect to `/login` from dashboard flow.
  - [ ] No repeated redirect loop in browser.

  **Agent-Executed QA Scenarios**:
  ```text
  Scenario: Anonymous user is redirected off dashboard
    Tool: Playwright
    Preconditions: Browser storage cleared
    Steps:
      1. Navigate to /dashboard
      2. Wait for URL change
      3. Assert: URL matches /login
    Expected Result: Dashboard protected from anonymous access
    Evidence: Screenshot capture
  ```

- [ ] 5. Align API base URL contract and admin endpoint response compatibility

  **What to do**:
  - Standardize `NEXT_PUBLIC_API_URL` semantics so local and CI both resolve admin endpoints correctly.
  - Ensure admin analytics endpoint response matches frontend expected shape (or adapt frontend parser).

  **Must NOT do**:
  - Hardcode environment-specific URLs in component code.
  - Leave CI/local path mismatch unresolved.

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
  - **Skills**: `playwright`, `git-master`

  **Parallelization**:
  - **Can Run In Parallel**: NO
  - **Parallel Group**: Wave 3
  - **Blocks**: 6
  - **Blocked By**: 1,2,3

  **References**:
  - `frontend/admin/src/lib/api.ts:4` - API base URL construction.
  - `.github/workflows/ci.yml` - current env wiring for frontend admin tests.
  - `frontend/admin/src/lib/types.ts` - expected analytics data contract.
  - `internal/adapter/http/admin_analytics_handler.go` - provided analytics response shape.

  **Acceptance Criteria**:
  - [ ] Local and CI both resolve auth/analytics API URLs correctly.
  - [ ] Dashboard renders without contract-shape runtime errors.
  - [ ] `cd frontend/admin && bun run build` passes.

  **Agent-Executed QA Scenarios**:
  ```text
  Scenario: CI-equivalent API URL setup works
    Tool: Bash
    Preconditions: Env set to same NEXT_PUBLIC_API_URL style as CI
    Steps:
      1. Run frontend admin tests/build with CI-like env
      2. Assert: no API path mismatch errors
      3. Assert: login + dashboard E2E can reach backend endpoints
    Expected Result: No environment-specific breakage
    Evidence: Terminal output capture
  ```

- [ ] 6. Final integration verification and evidence capture

  **What to do**:
  - Execute backend tests, frontend tests, and admin Playwright suite.
  - Capture evidence for happy path and negative auth path.

  **Must NOT do**:
  - Mark done without running the command suite.

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
  - **Skills**: `playwright`, `git-master`

  **Parallelization**:
  - **Can Run In Parallel**: NO
  - **Parallel Group**: Wave 3 final gate
  - **Blocks**: completion
  - **Blocked By**: 4,5

  **References**:
  - `frontend/admin/e2e/admin.spec.ts` - primary login/dashboard E2E coverage.
  - `internal/usecase/admin_auth_test.go` - backend auth tests.

  **Acceptance Criteria**:
  - [ ] `go test ./internal/usecase ./internal/adapter/http` passes for touched areas.
  - [ ] `cd frontend/admin && bun run test` passes.
  - [ ] `cd frontend/admin && bunx playwright test e2e/admin.spec.ts` passes.
  - [ ] Evidence files stored in `.sisyphus/evidence/` for login success/failure scenarios.

---

## Commit Strategy

| After Task | Message | Files | Verification |
|------------|---------|-------|--------------|
| 1-2 | `fix(admin-auth): wire and harden backend auth routes` | `internal/adapter/http/*`, `internal/usecase/*`, `cmd/server/main.go` | `go test ./internal/usecase ./internal/adapter/http` |
| 3-4 | `feat(admin-ui): implement login flow and dashboard guard` | `frontend/admin/src/app/*`, `frontend/admin/src/lib/*` | `cd frontend/admin && bun run test` |
| 5-6 | `chore(admin-ci): align api contract and verify e2e auth flow` | `frontend/admin/*`, `.github/workflows/ci.yml` | `cd frontend/admin && bunx playwright test e2e/admin.spec.ts` |

---

## Success Criteria

### Verification Commands
```bash
go test ./internal/usecase ./internal/adapter/http
cd frontend/admin && bun run test
cd frontend/admin && bun run build
cd frontend/admin && bunx playwright test e2e/admin.spec.ts
```

### Final Checklist
- [ ] Admin login works with `admin/admin123` and redirects to dashboard.
- [ ] Unauthorized users cannot access `/dashboard` or protected admin analytics endpoints.
- [ ] Invalid/stale tokens fail safely with 401 and no panic.
- [ ] Local and CI API URL configuration both work for admin auth + analytics.
- [ ] Automated test suite for touched surface passes.
