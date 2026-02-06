# User Aggregate Settings API

## TL;DR

> **Quick Summary**: Create a unified API (`/api/user/aggregate`) to fetch and save all user settings (Profile, Categories, Accounts, Currencies) in a single call. Introduce a managed `accounts` table to allow editing/renaming accounts, while maintaining backward compatibility with the string-based expense API.
>
> **Deliverables**:
> - New `accounts` table (migration + seeding)
> - `GET /api/user/aggregate` endpoint
> - `PUT /api/user/aggregate` endpoint (Full replacement logic)
> - Frontend `HttpAggregateRepository` and refactored `SettingsPage`
>
> **Estimated Effort**: Medium
> **Parallel Execution**: Sequential (DB -> Backend -> Frontend)
> **Critical Path**: Migration -> Backend Logic -> Frontend Integration

---

## Context

### User Request
- **Goal**: "Get user aggregate settings at once... including accounts, home currency, category... client use this to show and edit... API endpoint to save... simplify process."
- **Constraint**: "Expense remain old API it used" (Account remains a string on Expense).
- **Requirement**: "Create a new worktree" (addressed via branch/worktree setup).
- **Clarification**:
  - **Accounts**: New managed table.
  - **Rename**: Updates historical expenses.
  - **Save**: Full replacement.

### Architecture Decisions
1.  **Managed Accounts Table**: We will create a `accounts` table (`id`, `user_id`, `name`).
2.  **String Compatibility**: Expenses still store `account` as a string.
3.  **Sync Logic**:
    - **Rename**: When `PUT` updates an account name, we update the `accounts` table AND perform a `UPDATE expenses SET account = new_name WHERE account = old_name`.
    - **Delete**: Removing an account from the list only removes it from the "Quick Pick" list. Historical expenses preserve the string name (preventing data loss).
4.  **Worktree**: The plan uses a feature branch `feature/aggregate-settings`. User can checkout this branch in a worktree if desired.

### Metis/Momus Review
**Identified Gaps** (addressed):
- **Gap**: Account deletion handling. **Resolution**: Only remove from definitions, keep history.
- **Gap**: Concurrency. **Resolution**: Last write wins (acceptable for single-user app).

---

## Work Objectives

### Core Objective
Simplify user settings management by aggregating all configuration into a single API surface and enabling account management.

### Concrete Deliverables
- [ ] DB Migration: `create_accounts_table`
- [ ] Backend: `AggregateSettings` struct & handlers
- [ ] Frontend: `SettingsPage` using new API

### Definition of Done
- [ ] `GET /api/user/aggregate` returns all data.
- [ ] `PUT /api/user/aggregate` updates profile, categories, and accounts.
- [ ] Renaming an account in Settings updates it in the Dashboard expenses list.
- [ ] Frontend loads settings in a single request.

### Must Have
- [ ] Transactional updates for Renames (Account + Expenses).
- [ ] Backward compatibility (Old API still works).

### Must NOT Have
- [ ] Breaking changes to `GET /api/expenses`.
- [ ] "Partial" patch support (API expects full object).

---

## Verification Strategy (MANDATORY)

> **UNIVERSAL RULE: ZERO HUMAN INTERVENTION**
> ALL verification is executed by the agent using tools.

### Test Decision
- **Infrastructure exists**: YES
- **Automated tests**: YES (Backend TDD + Integration)
- **Framework**: `go test`, `playwright` (Frontend)

### Agent-Executed QA Scenarios

```
Scenario: Get Aggregate Settings
  Tool: Bash (curl)
  Preconditions: User exists, has expenses
  Steps:
    1. curl GET /api/user/aggregate
    2. Assert: JSON contains "profile", "categories", "accounts", "currencies"
    3. Assert: "accounts" list contains values from existing expenses
  Expected Result: 200 OK with full data structure

Scenario: Rename Account (History Update)
  Tool: Bash (curl)
  Preconditions: Expense exists with account "Cash"
  Steps:
    1. Create/Seed Account "Cash"
    2. PUT /api/user/aggregate with accounts: [{name: "Gold", old_name: "Cash"}]
    3. Assert: 200 OK
    4. GET /api/expenses
    5. Assert: Expense that was "Cash" is now "Gold"
  Expected Result: History updated

---

## TODOs


- [x] 0. Setup Feature Branch
  **What to do**:
  - Create branch `feature/aggregate-settings`.
  - (Optional) User can run `git worktree add ../aggregate-settings feature/aggregate-settings`.
  - **Agent Action**: Just `git checkout -b feature/aggregate-settings`.

  **Recommended Agent Profile**:
  - **Category**: `quick`
  - **Skills**: [`git-master`]

  **Acceptance Criteria**:
  - [ ] On branch `feature/aggregate-settings`

- [x] 1. DB Migration: Accounts Table
  **What to do**:
  - Create migration `migrations/XXX_create_accounts_table.up.sql`.
  - Table `accounts`:
    - `user_id` (TEXT NOT NULL REFERENCES users(id))
    - `name` (TEXT NOT NULL)
    - `created_at` (TIMESTAMP DEFAULT CURRENT_TIMESTAMP)
    - `PRIMARY KEY (user_id, name)` (Composite Key - acts as ID)
  - Seed: `INSERT INTO accounts (user_id, name) SELECT DISTINCT user_id, account FROM expenses WHERE account IS NOT NULL AND account != ''`.

  **Recommended Agent Profile**:
  - **Category**: `quick`
  - **Skills**: [`git-master`]

  **Acceptance Criteria**:
  - [ ] Table exists with Composite Key
  - [ ] Populated with existing accounts

- [x] 2. Backend: Define Aggregate Models
  **What to do**:
  - Create `internal/domain/aggregate.go`:
    ```go
    type AggregateSettings struct {
        Profile    *User
        Categories []*Category
        Accounts   []*Account
        Currencies []Currency
    }
    ```
  - Define `Account` model (UserID and Name).

  **Recommended Agent Profile**:
  - **Category**: `quick`

  **Acceptance Criteria**:
  - [ ] Structs defined

- [ ] 3. Backend: Implement GET Handler
  **What to do**:
  - Create `internal/usecase/get_user_aggregate.go` (UseCase).
  - Create handler logic in `internal/adapter/http/handler.go` (or new file `aggregate_handler.go` linked in `handler.go`).
  - Register route `GET /api/user/aggregate` in `internal/adapter/http/handler.go:NewHandler`.
  - Fetch User, Categories, Accounts (from new table), Currencies.

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`

  **Acceptance Criteria**:
  - [ ] `curl /api/user/aggregate` returns JSON with all fields.

- [ ] 4. Backend: Implement PUT Handler (Logic Core)
  **What to do**:
  - Create `internal/usecase/update_user_aggregate.go` (UseCase).
  - Register route `PUT /api/user/aggregate` in `internal/adapter/http/handler.go`.
  - **Profile**: Update fields.
  - **Categories**: Full sync (Delete missing, Update existing, Create new).
  - **Accounts**:
    - **Renaming**: Frontend must send `old_name` and `new_name`.
    - Logic:
      - If `old_name` != `new_name`:
        - Update `accounts` table: `UPDATE accounts SET name = new WHERE user_id = X AND name = old`.
        - Update `expenses` table: `UPDATE expenses SET account = new WHERE user_id = X AND account = old`.
      - If `new_name` (no old): Insert `INSERT INTO accounts ...`.
      - If missing from input list: `DELETE FROM accounts WHERE user_id = X AND name = missing_name`.

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`

  **Acceptance Criteria**:
  - [ ] `PUT` updates profile.
  - [ ] `PUT` renames account and updates expense history.
  - [ ] `PUT` deletes account from definition list.

- [ ] 5. Frontend: Add Aggregate Repository
  **What to do**:
  - Create `frontend/dashboard/src/infrastructure/repositories/http/HttpAggregateRepository.ts`.
  - Methods: `getAggregate()`, `saveAggregate(data)`.
  - Define TypeScript interfaces.

  **Recommended Agent Profile**:
  - **Category**: `quick`

  **Acceptance Criteria**:
  - [ ] Repository compiles and is usable.

- [ ] 6. Frontend: Refactor Settings Page
  **What to do**:
  - Update `frontend/dashboard/src/app/[locale]/dashboard/settings/page.tsx` to call `getAggregate` on mount (replace parallel calls).
  - Update "Save" button to call `saveAggregate`.
  - Add "Accounts" tab/section to edit account list.

  **Recommended Agent Profile**:
  - **Category**: `visual-engineering`
  - **Skills**: [`frontend-ui-ux`, `playwright`]

  **Acceptance Criteria**:
  - [ ] Settings page loads faster (1 request).
  - [ ] Can add/rename/delete accounts.
  - [ ] Renaming "Cash" to "Gold" -> Dashboard shows "Gold".

---

## Success Criteria

### Final Checklist
- [ ] `GET /api/user/aggregate` works.
- [ ] `PUT /api/user/aggregate` works and handles renames.
- [ ] `expenses` table data is consistent with renames.
- [ ] Frontend Settings page is fully functional.
- [ ] No regression in `GET /api/expenses`.
