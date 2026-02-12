# Redesign Admin Panel for Decision-Grade Business Stats

## TL;DR

> **Quick Summary**: Build a new admin panel from scratch focused on Revenue & Retention decisions, backed by new analytics endpoints and independent admin authentication.
>
> **Deliverables**:
> - New admin analytics API surface (definitions + endpoints + tests)
> - New admin auth flow isolated from current metrics API-key checks
> - New admin frontend in `frontend/admin` with KPI hierarchy and drill-downs
> - TDD-backed implementation + agent-executed QA + CI enforcement
>
> **Estimated Effort**: Large
> **Parallel Execution**: YES - 3 waves
> **Critical Path**: Task 1 -> Task 2 -> Task 3 -> Task 4 -> Task 6 -> Task 7

---

## Context

### Original Request
Redesign the admin panel from scratch so it shows stronger business stats to support better decisions.

### Interview Summary
**Key Discussions**:
- Primary lens: Revenue & Retention
- Decision cadence: Daily operations + weekly review
- Data source: new analytics endpoints (not only existing report endpoints)
- Auth model: independent admin authentication
- Testing: TDD, plus mandatory agent-executed QA

**Research Findings**:
- Existing frontend architecture/patterns live in `frontend/dashboard`
- Existing backend metrics/auth patterns live in `internal/adapter/http/handler.go`
- Existing dashboard test infra exists (Vitest + Playwright)
- `frontend/admin` exists and is currently empty

### Gap Review
**Identified Gaps (addressed in this plan)**:
- Missing explicit metric semantics for revenue/retention -> add metric dictionary and fixture-based correctness tests
- Scope creep risk into full BI platform -> set strict V1 exclusions
- Security ambiguity for admin auth -> isolate auth surface and require audit trail checks
- Edge-case risk (refunds, chargebacks, time boundaries) -> add explicit negative scenarios and acceptance criteria

---

## Work Objectives

### Core Objective
Deliver a production-ready V1 admin analytics experience that enables daily and weekly Revenue/Retention decisions using reliable definitions, fast drill-downs, and isolated admin access.

### Concrete Deliverables
- OpenSpec-approved change for admin analytics + auth
- New backend analytics contracts and endpoints for revenue/retention metrics
- New independent admin auth flow for admin APIs/panel
- New admin frontend app experience in `frontend/admin`
- Automated tests + E2E verification + CI gate updates

### Definition of Done
- [x] OpenSpec proposal validated with `openspec validate <change-id> --strict`
- [x] Revenue/retention endpoints return schema-validated responses for happy and edge cases
- [x] Admin auth protects new analytics routes with explicit unauthorized behavior
- [x] New admin panel renders KPI tree (L1/L2/L3), filters, and drill-down actions
- [x] `go test ./...` passes
- [x] `bun run test` and Playwright suite for admin pass in `frontend/admin`
- [x] CI runs dashboard/admin unit tests and e2e checks for changed surface

### Must Have
- Revenue/Retention-first information architecture (not generic vanity dashboard)
- Contextual metrics (trend vs prior period/target), not raw numbers without context
- Daily ops + weekly review workflows reflected directly in UI
- Independent admin authentication (no reliance on bare `X-API-Key` in UI)

### Must NOT Have (Guardrails)
- No V1 full BI features (custom query builder, scheduled report engine, forecasting)
- No unrelated domain expansion (support ops, product telemetry, broad IAM program)
- No human-only verification steps; all acceptance checks must be agent-executable

---

## Verification Strategy (MANDATORY)

> **UNIVERSAL RULE: ZERO HUMAN INTERVENTION**
>
> All acceptance criteria must be verifiable by an executing agent using commands/tools.

### Test Decision
- **Infrastructure exists**: YES
- **Automated tests**: TDD
- **Framework**: Go test + Vitest + Playwright

### TDD Task Structure
For each implementation task:
1. RED: create failing test
2. GREEN: implement minimum behavior
3. REFACTOR: cleanup while keeping tests green

### Agent-Executed QA Scenarios (all tasks)
- Frontend/UI: Playwright (route, selectors, assertions, screenshots)
- Backend/API: curl + response assertions (status, body fields, edge cases)
- CLI/build/test: Bash commands with explicit expected outcomes

---

## Execution Strategy

### Dev Process Compliance (MANDATORY)

- Before implementation starts, follow project workflow from `AGENTS.md`:
  1. Create GitHub issue for this work (`gh issue create`)
  2. Create linked branch from issue (`gh issue develop <issue-number> --checkout`)
  3. Create/use git worktree under `.worktrees/<branch-name>`
  4. Implement via TDD in the worktree
  5. Commit with conventional commits, push, and open PR (`gh pr create`)
  6. Address review comments with follow-up commits (no force push during review)
- This workflow is part of Definition of Done for the execution phase.

### Parallel Execution Waves

Wave 1 (Foundation)
- Task 1: OpenSpec change proposal and acceptance framing
- Task 2: Metric dictionary + backend contracts

Wave 2 (Core Build)
- Task 3: Independent admin auth
- Task 4: Revenue/retention analytics endpoints
- Task 5: Frontend admin shell + layout system

Wave 3 (Integration + Quality)
- Task 6: Frontend KPI pages + endpoint wiring + drill-down UX
- Task 7: CI/test hardening and release verification

Critical Path: 1 -> 2 -> 3 -> 4 -> 6 -> 7

### Dependency Matrix

| Task | Depends On | Blocks | Can Parallelize With |
|------|------------|--------|----------------------|
| 1 | None | 2,3,4,5,6,7 | 2 |
| 2 | 1 | 4,6 | None |
| 3 | 1 | 4,6 | 5 |
| 4 | 2,3 | 6,7 | 5 |
| 5 | 1 | 6 | 3,4 |
| 6 | 4,5 | 7 | None |
| 7 | 6 | None | None |

### Agent Dispatch Summary

| Wave | Tasks | Recommended Agents |
|------|-------|--------------------|
| 1 | 1,2 | `task(category="unspecified-high", load_skills=["notion-spec-to-implementation"])` + `task(category="deep", load_skills=[])` |
| 2 | 3,4,5 | backend/auth: `task(category="deep", load_skills=[])`; frontend shell: `task(category="visual-engineering", load_skills=["frontend-ui-ux"])` |
| 3 | 6,7 | integration: `task(category="unspecified-high", load_skills=["playwright"])` |

---

## TODOs

- [x] 1. Create OpenSpec change for admin analytics redesign

  **What to do**:
  - Create `openspec/changes/add-admin-analytics-panel/` with `proposal.md`, `tasks.md`, and spec deltas
  - Define requirement-level scenarios for metric correctness, auth isolation, and dashboard behavior
  - Validate strictly before execution work starts

  **Must NOT do**:
  - Start implementation before OpenSpec validation
  - Write vague requirement text without executable scenarios

  **Recommended Agent Profile**:
  - **Category**: `writing`
    - Reason: spec/proposal quality and requirement precision are the core objective
  - **Skills**: `notion-spec-to-implementation`
    - `notion-spec-to-implementation`: strong structure for converting requirements into executable tasks
  - **Skills Evaluated but Omitted**:
    - `copywriting`: marketing text focus, not technical spec authoring

  **Parallelization**:
  - **Can Run In Parallel**: NO
  - **Parallel Group**: Wave 1 (starts first)
  - **Blocks**: 2,3,4,5,6,7
  - **Blocked By**: None

  **References**:
  - `openspec/AGENTS.md` - OpenSpec authoring and validation rules
  - `openspec/project.md` - project conventions for capabilities and deltas

  **Acceptance Criteria**:
  - [ ] `openspec list` shows `add-admin-analytics-panel`
  - [ ] `openspec validate add-admin-analytics-panel --strict` passes
  - [ ] Spec deltas include `#### Scenario:` blocks for each requirement

  **Agent-Executed QA Scenarios**:
  ```text
  Scenario: OpenSpec validation passes for new proposal
    Tool: Bash
    Preconditions: Proposal files exist under openspec/changes/add-admin-analytics-panel
    Steps:
      1. Run: openspec validate add-admin-analytics-panel --strict
      2. Assert: command exit code is 0
      3. Assert: stdout contains no validation errors
    Expected Result: Proposal is valid and executable by implementation agents
    Evidence: Terminal output capture

  Scenario: Missing scenario block fails validation
    Tool: Bash
    Preconditions: Temporarily remove one `#### Scenario:` from a draft delta in local working copy
    Steps:
      1. Run: openspec validate add-admin-analytics-panel --strict
      2. Assert: command exit code is non-zero
      3. Assert: output reports missing scenario formatting requirement
    Expected Result: Invalid spec changes are blocked before implementation
    Evidence: Validation error output capture
  ```

- [x] 2. Define revenue/retention metric contracts and test fixtures (TDD)

  **What to do**:
  - Add/extend domain contracts for revenue-retention metrics (MRR, NRR, GRR, churn, expansion/contraction)
  - Create table-driven backend tests for edge semantics (refunds, chargebacks, plan changes, cancel-at-period-end)
  - Ensure response schema is explicit and stable for frontend consumption

  **Must NOT do**:
  - Introduce metric formulas without explicit definitions in tests
  - Return untyped/ambiguous maps to frontend for core KPI contracts

  **Recommended Agent Profile**:
  - **Category**: `deep`
    - Reason: business semantics + correctness-critical domain logic
  - **Skills**: `git-master`
    - `git-master`: helps keep atomic commits and traceability during model/contract reshaping
  - **Skills Evaluated but Omitted**:
    - `frontend-ui-ux`: not relevant to backend metric semantics

  **Parallelization**:
  - **Can Run In Parallel**: NO
  - **Parallel Group**: Wave 1
  - **Blocks**: 4,6
  - **Blocked By**: 1

  **References**:
  - `internal/usecase/metrics.go` - current metrics request/response and aggregate style
  - `internal/domain/repositories.go` - existing `MetricsRepository` contract location
  - `internal/domain/models.go` - current daily/category metrics models
  - `internal/adapter/repository/postgresql/metrics_repo.go` - existing query-layer metric retrieval

  **Acceptance Criteria**:
  - [ ] RED: new metric-definition tests fail initially (`go test ./internal/usecase -run Metrics -v`)
  - [ ] GREEN: metric tests pass with explicit formulas and edge-case behavior
  - [ ] REFACTOR: no duplicate formula logic across usecase/repository layers

  **Agent-Executed QA Scenarios**:
  ```text
  Scenario: Metric formula fixtures validate NRR/GRR semantics
    Tool: Bash
    Preconditions: New metrics tests exist in internal/usecase
    Steps:
      1. Run: go test ./internal/usecase -run Metrics -v
      2. Assert: all metric contract tests PASS
      3. Assert: test output includes edge case names (refund, chargeback, downgrade)
    Expected Result: Metric definitions are unambiguous and executable
    Evidence: Go test verbose output

  Scenario: Invalid metric query returns typed error payload
    Tool: Bash (curl)
    Preconditions: Server running on localhost:8080
    Steps:
      1. Call analytics endpoint with unsupported period/filter
      2. Assert: HTTP status is 400
      3. Assert: JSON shape matches {"status":"error","error":"..."}
    Expected Result: Bad requests fail predictably
    Evidence: Response body capture
  ```

- [x] 3. Implement independent admin authentication for analytics surface (TDD)

  **What to do**:
  - Introduce admin auth flow independent from current simple API key checks
  - Add auth middleware/guard for new analytics routes
  - Implement login/session lifecycle and unauthorized/forbidden responses

  **Must NOT do**:
  - Reuse end-user report token flow for admin auth without explicit boundary
  - Leave fallback that silently bypasses auth when key is absent

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
    - Reason: security-sensitive behavior requiring disciplined handling
  - **Skills**: `git-master`
    - `git-master`: aids safe incremental commits around security logic
  - **Skills Evaluated but Omitted**:
    - `playwright`: useful later for E2E, not primary for backend auth implementation

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 2 (parallel with 5)
  - **Blocks**: 4,6
  - **Blocked By**: 1

  **References**:
  - `internal/adapter/http/handler.go` - existing admin auth check (`authenticateAdmin`) and protected metrics routes
  - `internal/config/config.go` - current admin key config and env loading
  - `internal/adapter/http/pricing_handler.go` - another admin-gated handler pattern
  - `internal/adapter/http/ai_cost_handler.go` - repeated admin auth pattern to consolidate/harden

  **Acceptance Criteria**:
  - [ ] RED: auth tests fail for missing/invalid admin credentials
  - [ ] GREEN: protected analytics endpoints return 401/403 correctly
  - [ ] REFACTOR: auth logic centralized to avoid drift across handlers

  **Agent-Executed QA Scenarios**:
  ```text
  Scenario: Admin login issues usable authenticated session/token
    Tool: Bash (curl)
    Preconditions: Server running with admin auth configuration
    Steps:
      1. POST /api/admin/auth/login with valid credentials
      2. Assert: HTTP 200
      3. Assert: response includes token/session metadata
      4. GET protected analytics endpoint with auth -> HTTP 200
    Expected Result: Authenticated admin can access analytics endpoints
    Evidence: Response bodies captured

  Scenario: Unauthorized analytics access is blocked
    Tool: Bash (curl)
    Preconditions: Server running
    Steps:
      1. GET protected analytics endpoint without auth
      2. Assert: HTTP 401
      3. GET with malformed token
      4. Assert: HTTP 401 or 403 with error payload
    Expected Result: No unauthenticated analytics exposure
    Evidence: Response bodies captured
  ```

- [x] 4. Build new analytics endpoints for revenue/retention drill-downs (TDD)

  **What to do**:
  - Add endpoints needed by admin panel (KPI summary, trend series, cohort retention, at-risk accounts list)
  - Support decision filters (period, comparison window, segment dimensions required for V1)
  - Add robust error handling for invalid ranges, empty cohorts, and partial data windows

  **Must NOT do**:
  - Build custom query engine / arbitrary SQL endpoint
  - Return slow unbounded payloads without pagination/limits

  **Recommended Agent Profile**:
  - **Category**: `deep`
    - Reason: endpoint design couples domain logic, data access, and performance constraints
  - **Skills**: `git-master`
    - `git-master`: keeps API evolution reviewable and reversible
  - **Skills Evaluated but Omitted**:
    - `frontend-ui-ux`: not applicable for backend endpoint implementation

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 2 (parallel with 5 after 2+3 are ready)
  - **Blocks**: 6,7
  - **Blocked By**: 2,3

  **References**:
  - `internal/adapter/http/metrics_handler.go` - existing metrics endpoint style and parameter parsing
  - `internal/adapter/http/handler.go` - route registration pattern and JSON response wrapper
  - `internal/usecase/metrics.go` - usecase boundary for metric calculations
  - `internal/adapter/repository/postgresql/metrics_repo.go` - SQL aggregation pattern baseline
  - `internal/domain/repositories.go` - canonical repository interface definitions

  **Acceptance Criteria**:
  - [ ] RED: endpoint tests fail for missing handlers/contracts
  - [ ] GREEN: all new analytics endpoints return expected schema and status codes
  - [ ] REFACTOR: duplicated metric assembly logic reduced behind clear usecase boundaries

  **Agent-Executed QA Scenarios**:
  ```text
  Scenario: KPI summary endpoint returns decision-grade context
    Tool: Bash (curl)
    Preconditions: Server running with seeded representative data
    Steps:
      1. GET /api/admin/analytics/overview?period=30d&compare=prev_30d
      2. Assert: HTTP 200
      3. Assert: payload contains mrr, nrr, grr, churn and comparison deltas
      4. Assert: each KPI includes current_value and delta_percent
    Expected Result: Overview supports immediate daily ops interpretation
    Evidence: JSON response capture

  Scenario: Invalid filter combination returns actionable 400
    Tool: Bash (curl)
    Preconditions: Server running
    Steps:
      1. GET /api/admin/analytics/overview?period=invalid
      2. Assert: HTTP 400
      3. Assert: error describes accepted period values
    Expected Result: API contract fails fast with clear guidance
    Evidence: Error payload capture
  ```

- [x] 5. Scaffold new admin frontend shell in `frontend/admin` (TDD)

  **What to do**:
  - Initialize admin app structure using current Next.js conventions (App Router, i18n route segment where applicable)
  - Create panel shell (navigation, top bar, responsive layout, loading/error states)
  - Define reusable KPI card/chart containers for the new style direction

  **Must NOT do**:
  - Copy old dashboard UI 1:1 without redesign intent
  - Use generic default typography/colors that ignore project visual direction

  **Recommended Agent Profile**:
  - **Category**: `visual-engineering`
    - Reason: this is foundational UX architecture with responsive interaction requirements
  - **Skills**: `frontend-ui-ux`
    - `frontend-ui-ux`: strong visual-system and interaction design execution
  - **Skills Evaluated but Omitted**:
    - `playwright`: used for verification, not for initial design implementation

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 2 (parallel with 3/4)
  - **Blocks**: 6
  - **Blocked By**: 1

  **References**:
  - `frontend/dashboard/src/app/[locale]/dashboard/page.tsx` - page orchestration, loading/error handling pattern
  - `frontend/dashboard/src/components/DashboardLayout.tsx` - responsive shell behavior pattern
  - `frontend/dashboard/src/components/DashboardCard.tsx` - card container baseline for KPI modules
  - `frontend/dashboard/src/app/[locale]/layout.tsx` - locale-aware routing/layout structure
  - `frontend/dashboard/src/app/globals.css` - tokenized design variables baseline

  **Acceptance Criteria**:
  - [ ] RED: initial frontend unit tests fail for missing shell components
  - [ ] GREEN: shell renders mobile/desktop navigation + empty/loading/error states
  - [ ] REFACTOR: shared shell components avoid duplicated markup across pages

  **Agent-Executed QA Scenarios**:
  ```text
  Scenario: Admin shell loads and adapts to mobile/desktop
    Tool: Playwright (playwright skill)
    Preconditions: frontend/admin dev server running on localhost:3001
    Steps:
      1. Navigate to http://localhost:3001/admin
      2. Assert: main heading for admin analytics is visible
      3. Set viewport 390x844
      4. Click menu toggle button
      5. Assert: mobile sidebar panel appears
      6. Set viewport 1440x900
      7. Assert: desktop sidebar remains visible and overlay is absent
      8. Screenshot: .sisyphus/evidence/task-5-admin-shell-responsive.png
    Expected Result: Shell is responsive and navigable across breakpoints
    Evidence: .sisyphus/evidence/task-5-admin-shell-responsive.png

  Scenario: Unauthorized state routes to admin login
    Tool: Playwright (playwright skill)
    Preconditions: no valid admin session present
    Steps:
      1. Clear storage/cookies
      2. Navigate to /admin
      3. Assert: redirect lands on /admin/login
      4. Assert: login form fields are visible
      5. Screenshot: .sisyphus/evidence/task-5-admin-login-redirect.png
    Expected Result: Protected shell cannot be accessed anonymously
    Evidence: .sisyphus/evidence/task-5-admin-login-redirect.png
  ```

- [x] 6. Implement revenue/retention decision views + API wiring in admin panel (TDD)

  **What to do**:
  - Build L1/L2/L3 metric hierarchy UI:
    - L1: headline business health KPIs
    - L2: driver views (funnel/cohort/trend)
    - L3: action queues (at-risk accounts, retention risk)
  - Wire filters and drill-down interactions to new analytics endpoints
  - Enforce context-rich KPI display (delta vs previous period/goal)

  **Must NOT do**:
  - Show context-free vanity metrics
  - Leave charts without drill-down or actionable next step

  **Recommended Agent Profile**:
  - **Category**: `visual-engineering`
    - Reason: heavy UI composition + data interaction logic
  - **Skills**: `frontend-ui-ux`, `playwright`
    - `frontend-ui-ux`: visual hierarchy and intentional dashboard composition
    - `playwright`: robust E2E validation for interaction-heavy UI
  - **Skills Evaluated but Omitted**:
    - `copywriting`: not core to data interaction correctness

  **Parallelization**:
  - **Can Run In Parallel**: NO
  - **Parallel Group**: Wave 3
  - **Blocks**: 7
  - **Blocked By**: 4,5

  **References**:
  - `frontend/dashboard/src/infrastructure/RepositoryFactory.ts` - existing repository factory pattern to mirror in admin app
  - `frontend/dashboard/src/infrastructure/repositories/http/HttpReportRepository.ts` - HTTP repository implementation style
  - `frontend/dashboard/src/components/ChartSection.tsx` - chart composition and tooltip conventions
  - `frontend/dashboard/src/components/MetricsGrid.tsx` - card grid and icon/label/value pattern baseline
  - `frontend/dashboard/src/components/ExpenseListTable.tsx` - tabular drill-down interaction pattern

  **Acceptance Criteria**:
  - [x] RED: UI tests fail before data wiring and interactions are implemented
  - [x] GREEN: KPI cards, trend charts, and drill-down tables render from live analytics endpoints
  - [x] REFACTOR: data fetching and transformation centralized to avoid per-component duplication

  **Agent-Executed QA Scenarios**:
  ```text
  Scenario: Revenue overview updates with period filter and comparison toggle
    Tool: Playwright (playwright skill)
    Preconditions: frontend/admin and backend running, authenticated admin session
    Steps:
      1. Navigate to /admin
      2. Wait for: [data-testid="kpi-mrr"] visible
      3. Select period filter: "Last 30 Days"
      4. Toggle comparison: "vs previous period"
      5. Assert: [data-testid="kpi-mrr-delta"] contains '%' value
      6. Assert: [data-testid="kpi-nrr"] displays numeric value
      7. Screenshot: .sisyphus/evidence/task-6-kpi-filter-comparison.png
    Expected Result: KPI context changes correctly with filters
    Evidence: .sisyphus/evidence/task-6-kpi-filter-comparison.png

  Scenario: Cohort drill-down opens actionable account list
    Tool: Playwright (playwright skill)
    Preconditions: cohort chart has data
    Steps:
      1. Click chart element: [data-testid="cohort-cell-2025-12"]
      2. Wait for panel: [data-testid="at-risk-accounts-table"]
      3. Assert: table row count > 0
      4. Assert: each row shows account, retention risk indicator, and last activity
      5. Screenshot: .sisyphus/evidence/task-6-cohort-drilldown.png
    Expected Result: Every major chart supports direct drill-down action
    Evidence: .sisyphus/evidence/task-6-cohort-drilldown.png

  Scenario: Empty-data range renders safe empty state
    Tool: Playwright (playwright skill)
    Preconditions: Choose date range with no data
    Steps:
      1. Set period to custom range without records
      2. Assert: [data-testid="empty-state"] is visible
      3. Assert: no runtime error toast appears
      4. Screenshot: .sisyphus/evidence/task-6-empty-state.png
    Expected Result: Zero-data scenarios are handled gracefully
    Evidence: .sisyphus/evidence/task-6-empty-state.png
  ```

- [x] 7. Harden tests, CI gates, and release verification for admin analytics

  **What to do**:
  - Ensure backend and admin/frontend test suites run in CI for affected scope
  - Add/adjust CI steps so unit tests are not skipped for dashboard/admin UI surfaces
  - Final integration pass with evidence capture and rollout checklist

  **Must NOT do**:
  - Merge with locally-passing-only test evidence
  - Leave CI without unit-test execution for changed frontend app

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
    - Reason: cross-layer integration and pipeline reliability work
  - **Skills**: `playwright`, `git-master`
    - `playwright`: deterministic E2E validation in CI/local
    - `git-master`: clean checkpointing for CI config updates
  - **Skills Evaluated but Omitted**:
    - `gh-fix-ci`: good for after-the-fact failures; this task is preemptive hardening

  **Parallelization**:
  - **Can Run In Parallel**: NO
  - **Parallel Group**: Wave 3 (final gate)
  - **Blocks**: completion
  - **Blocked By**: 6

  **References**:
  - `.github/workflows/ci.yml` - current CI jobs and missing dashboard unit test execution
  - `frontend/dashboard/package.json` - unit test script (`vitest`) pattern
  - `frontend/dashboard/playwright.config.ts` - e2e runner configuration baseline
  - `frontend/dashboard/tests/dashboard.spec.ts` - Playwright assertion style

  **Acceptance Criteria**:
  - [x] CI config includes required unit + e2e checks for admin analytics surface
  - [x] `go test ./...` passes
  - [x] frontend unit tests pass (admin app)
  - [x] Playwright admin smoke and negative flows pass

  **Agent-Executed QA Scenarios**:
  ```text
  Scenario: Full local verification command suite passes
    Tool: Bash
    Preconditions: dependencies installed
    Steps:
      1. Run: go test ./...
      2. Assert: exit code 0
      3. Run: bun run test (within frontend/admin)
      4. Assert: exit code 0
      5. Run: bunx playwright test (within frontend/admin)
      6. Assert: exit code 0
    Expected Result: all required quality gates pass before merge
    Evidence: terminal output capture

  Scenario: CI workflow includes frontend unit-test step
    Tool: Bash
    Preconditions: CI config updated
    Steps:
      1. Inspect `.github/workflows/ci.yml`
      2. Assert: dashboard/admin job contains unit test command (`bun run test`)
      3. Assert: e2e job still executes Playwright with backend startup
    Expected Result: regression-proof CI coverage for admin redesign
    Evidence: diff snippet and file contents capture

  Scenario: CI catches intentionally broken frontend unit test
    Tool: Bash
    Preconditions: Introduce a temporary failing test in frontend/admin test suite
    Steps:
      1. Run: bun run test (within frontend/admin)
      2. Assert: exit code is non-zero
      3. Revert temporary failing test
      4. Re-run: bun run test
      5. Assert: exit code returns to 0
    Expected Result: CI-relevant unit test failures are reliably detected
    Evidence: before/after terminal output capture
  ```

---

## Commit Strategy

| After Task | Message | Files | Verification |
|------------|---------|-------|--------------|
| 1 | `docs(openspec): propose admin analytics redesign` | `openspec/changes/add-admin-analytics-panel/*` | `openspec validate add-admin-analytics-panel --strict` |
| 2-4 | `feat(analytics): add revenue-retention contracts and admin auth` | backend domain/usecase/http/repository files | `go test ./...` |
| 5-6 | `feat(admin-ui): build revenue-retention decision dashboard` | `frontend/admin/**/*` | `bun run test && bunx playwright test` |
| 7 | `chore(ci): enforce admin analytics quality gates` | `.github/workflows/ci.yml` | local command suite matches CI commands |

---

## Success Criteria

### Verification Commands
```bash
openspec validate add-admin-analytics-panel --strict
go test ./...
bun run test
bunx playwright test
```

### Final Checklist
- [ ] Revenue/Retention KPI hierarchy is implemented and decision-focused
- [ ] Independent admin auth protects analytics routes
- [ ] Edge-case metric semantics are tested and documented
- [ ] Admin panel supports daily ops and weekly review workflows
- [ ] CI reliably blocks regressions across backend and admin frontend
