# Plan: Fix Report Redirect Link

## TL;DR

> **Quick Summary**: Fix the broken report redirect link by updating the backend to point to `/dashboard` and ensuring the frontend middleware correctly routes that path to a locale-specific URL (e.g., `/en/dashboard`).
> 
> **Deliverables**:
> - Updated `internal/adapter/http/short_link_handler.go` (backend)
> - Updated `internal/adapter/http/short_link_handler_test.go` (backend)
> - Updated `frontend/dashboard/src/middleware.ts` (frontend)
> 
> **Estimated Effort**: Quick
> **Parallel Execution**: Sequential

---

## Context

### Original Request
User reported: "I have report feature, but now it broke. fix it".
Backend returns short link `http://localhost:8080/r/UO85tf`.

### Problem
1. **Backend**: `ShortLinkHandler` redirects to `/user/reports`, which no longer exists.
2. **Frontend**: The new path is `/dashboard`. However, the frontend uses `next-intl` with strict middleware matching that currently *ignores* `/dashboard` (non-localized), causing a 404 instead of a redirect to `/en/dashboard`.

### Solution
1. **Backend**: Change redirect to `/dashboard?token=...`.
2. **Frontend**: Update middleware matcher to include `/dashboard/:path*`, enabling `next-intl` to detect the missing locale and redirect to `/{locale}/dashboard`.

### Metis/Momus Review
**Identified Gaps** (addressed):
- **Routing**: Momus correctly identified that `/dashboard` would 404 because `middleware.ts` excluded it.
- **Fix**: Explicitly added frontend middleware update to the plan.

---

## Work Objectives

### Core Objective
Restore functionality of the expense report short links.

### Concrete Deliverables
- [x] Backend: Redirects to `/dashboard`
- [x] Frontend: Middleware handles `/dashboard` and redirects to `/{locale}/dashboard`

### Definition of Done
- [x] `go test` passes
- [x] `curl -I http://localhost:8080/r/{id}` -> `Location: .../dashboard?token=...`
- [x] Accessing `http://localhost:3000/dashboard` redirects to `http://localhost:3000/en/dashboard` (or other locale)

---

## Verification Strategy (MANDATORY)

> **UNIVERSAL RULE: ZERO HUMAN INTERVENTION**
> ALL verification is executed by the agent using tools.

### Test Decision
- **Infrastructure exists**: YES
- **Automated tests**: YES (Backend unit test)
- **Framework**: `go test`, `curl`

### Agent-Executed QA Scenarios

```
Scenario: Verify Backend Redirect
  Tool: Bash
  Preconditions: Server running (or mocked)
  Steps:
    1. Generate short link via API
    2. curl -I {short_link}
    3. Assert Location header contains `/dashboard`
  Expected Result: HTTP 302, Location: .../dashboard...

Scenario: Verify Frontend Routing (Middleware)
  Tool: Bash (curl)
  Preconditions: Frontend running on port 3000
  Steps:
    1. curl -I http://localhost:3000/dashboard
  Expected Result: HTTP 307 (Temporary Redirect) or 308, Location: /en/dashboard
```

---

## TODOs

- [x] 1. Update Frontend Middleware

  **What to do**:
  - Edit `frontend/dashboard/src/middleware.ts`.
  - Update `matcher` array to include `/dashboard/:path*`.
  - Current: `['/', '/(en|es|zh-TW)/:path*']`
  - New: `['/', '/dashboard/:path*', '/(en|es|zh-TW)/:path*']`

  **Recommended Agent Profile**:
  - **Category**: `quick`

  **Acceptance Criteria**:
  - [ ] `matcher` includes `/dashboard/:path*`

- [x] 2. Update Backend Redirect

  **What to do**:
  - Edit `internal/adapter/http/short_link_handler.go`: Change `/user/reports` to `/dashboard`.
  - Edit `internal/adapter/http/short_link_handler_test.go`: Update expected location in test.

  **Recommended Agent Profile**:
  - **Category**: `quick`
  - **Skills**: [`git-master`]

  **Acceptance Criteria**:
  - [ ] `redirectURL` is `.../dashboard?token=...`
  - [ ] `go test` passes

  **Agent-Executed QA Scenarios**:
  ```
  Scenario: Unit Test Verification
    Tool: Bash
    Steps:
      1. go test -v ./internal/adapter/http/ -run TestShortLinkHandler_HandleRedirect
    Expected Result: PASS
  ```

- [x] 3. Verify End-to-End Fix

  **What to do**:
  - Use `curl` to verify backend redirect.
  - (Optional) Use `curl` to verify frontend middleware redirect if frontend is running.

  **Recommended Agent Profile**:
  - **Category**: `quick`

  **Acceptance Criteria**:
  - [ ] Backend redirects to `/dashboard`

---

## Success Criteria

### Final Checklist
- [x] Backend tests pass
- [x] Frontend middleware updated
- [x] Redirect chain is valid: `Backend` -> `/dashboard` -> `Frontend` -> `/en/dashboard`
