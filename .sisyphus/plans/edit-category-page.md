# Edit Category Page - Settings Integration

## TL;DR

> **Quick Summary**: Add a "Category Management" section to the existing Settings page that allows users to view, add, rename, and delete their expense categories. **Includes smart merge**: when renaming to an existing category name, offers to merge expenses instead of showing error. **Categories include a description field** for AI-powered expense categorization hints.
> 
> **Deliverables**:
> - Database migration for category `description` column
> - Category model and repository in frontend domain layer
> - HttpCategoryRepository implementation
> - "Category Management" section in Settings page with inline editing
> - **Backend merge endpoint** for combining categories (`/api/user/categories/merge`)
> - **Merge confirmation UI** when rename conflicts
> - Playwright E2E tests for all category operations
> 
> **Estimated Effort**: Medium-Large (12-15 files, ~800 lines)
> **Parallel Execution**: YES - 5 waves
> **Critical Path**: Task 0 (Migration) → Task 1-2 (Backend) → Task 3-4 (Frontend Domain) → Task 5 (UI) → Task 6-9 (Features) → Task 10 (Tests)

---

## Context

### Original Request
"Add a page to edit category"

### Interview Summary
**Key Discussions**:
- User wants User Aggregate pattern: `User -> Categories, HomeCurrency, Accounts`
- Integrate into existing Settings page, not a new route
- Backend already has full CRUD API for categories
- Backend has `UserContext` aggregate with User + Categories
- **Merge Feature**: When renaming to existing name, merge categories instead of error
- **API Prefix**: Use `/api/user/` prefix for user-facing endpoints (to distinguish from future admin APIs)
- **Description Field**: Add description to Category for AI categorization context

**Research Findings**:
- Backend Category API fully implemented (currently at `/api/categories`)
- Frontend has RepositoryFactory pattern, Settings page, Vitest+Playwright tests
- No category management UI exists yet
- `ExpenseList.tsx` has inline edit pattern to follow
- **Expense-Category link**: `Expense.CategoryID *string` (optional FK)
- **No merge functionality exists** - needs to be built
- **FK constraint on delete** - prevents deleting categories with expenses

### Metis Review
**Identified Gaps** (addressed):
- Delete confirmation component doesn't exist → Use inline confirmation
- FK violation on delete categories with expenses → Show human-readable error
- Keywords in backend but not exposed → Explicitly out of scope
- Empty state handling → Show default categories always

### Oracle Review (Merge Feature)
**Architecture Decision**: Backend-first approach with dedicated merge endpoint
- Keep `PUT /api/user/categories` as "rename only" (throws explicit error on duplicate)
- Add `POST /api/user/categories/merge` for combining categories
- Merge as atomic transaction: reassign expenses → delete source
- Return `{merged_count, message}` for user feedback
- Frontend: detect duplicate error → show confirm → call merge

---

## Work Objectives

### Core Objective
Add category management UI to Settings page allowing users to view, add, rename, and delete expense categories. **When renaming to an existing category name, offer to merge expenses into the existing category.** Categories include a description field for AI context.

### API Design
**User-facing endpoints use `/api/user/` prefix** to distinguish from future admin APIs:
- `GET /api/user/categories?user_id=X` - List categories
- `POST /api/user/categories` - Create category
- `PUT /api/user/categories` - Update category (throws error on duplicate name)
- `DELETE /api/user/categories` - Delete category
- `POST /api/user/categories/merge` - Merge two categories

### Concrete Deliverables
- `migrations/XXXXXX_add_category_description.up.sql` - Add description column
- `internal/usecase/manage_category.go` - Add MergeCategories method, explicit duplicate check (backend)
- `internal/adapter/http/handler.go` - Add `/api/user/categories/*` endpoints (backend)
- `frontend/dashboard/src/domain/models/Category.ts` - Category interface with description
- `frontend/dashboard/src/domain/repositories/CategoryRepository.ts` - Repository interface with merge
- `frontend/dashboard/src/infrastructure/repositories/http/HttpCategoryRepository.ts` - HTTP implementation
- `frontend/dashboard/src/infrastructure/RepositoryFactory.ts` - Updated with category repository
- `frontend/dashboard/src/app/[locale]/dashboard/settings/page.tsx` - Updated with Categories section + merge UI
- `frontend/dashboard/tests/settings-categories.spec.ts` - E2E tests

### Definition of Done
- [x] Settings page displays user's categories in a new section
- [x] Each category shows name and description
- [x] Users can add new categories with name and optional description
- [x] Users can rename existing custom categories
- [x] Users can edit category descriptions
- [x] **When renaming to existing name, merge dialog appears with expense count**
- [x] **User can confirm merge, expenses move, source category deleted**
- [x] Users can delete custom categories (with confirmation)
- [x] Default categories are shown but edit/delete disabled
- [x] All E2E tests pass: `bunx playwright test settings-categories`

### Must Have
- Category list display with default/custom distinction
- **Description field** visible and editable for each category
- Add category with validation (name: non-empty, max 50 chars; description: optional, max 200 chars)
- Inline rename functionality
- **Explicit duplicate name error** (not just DB constraint)
- **Merge confirmation when renaming to existing category**
- **Clear feedback: "Merged X expenses from 'A' into 'B'"**
- Delete with inline confirmation
- Error handling for API failures

### Must NOT Have (Guardrails)
- ❌ Keyword management (even though API supports it)
- ❌ Category icons, colors, or emojis
- ❌ Drag-and-drop reordering
- ❌ Full modal/dialog system (use inline only)
- ❌ Toast notification system (use inline feedback)
- ❌ InMemoryCategoryRepository (only HTTP needed)
- ❌ Separate `/categories` route
- ❌ Category search or filter
- ❌ Batch operations
- ❌ Merging into/from default/reserved categories
- ❌ Cross-user category operations
- ❌ Changing existing non-category endpoints (keep current paths)

---

## Verification Strategy

> **UNIVERSAL RULE: ZERO HUMAN INTERVENTION**
>
> ALL tasks in this plan are verifiable by the executing agent using Playwright, bash, or curl.
> No acceptance criteria require human action.

### Test Decision
- **Infrastructure exists**: YES (Vitest + Playwright)
- **Automated tests**: Tests-after
- **Framework**: Playwright for E2E

### Agent-Executed QA Scenarios (MANDATORY — ALL tasks)

**Verification Tool by Deliverable Type:**

| Type | Tool | How Agent Verifies |
|------|------|-------------------|
| **Database Migration** | Bash (psql/sqlite) | Run migration, verify column exists |
| **Domain/Repository code** | Bash (TypeScript compile) | `cd frontend/dashboard && bun run build` |
| **Settings UI** | Playwright | Navigate, interact, assert DOM, screenshot |
| **API Integration** | Playwright + DevTools | Verify network requests succeed |

---

## Execution Strategy

### Parallel Execution Waves

```
Wave 0 (Start Immediately - Database):
└── Task 0: Add description column migration

Wave 1 (After Wave 0 - Backend):
├── Task 1: Backend - Add MergeCategories usecase + explicit duplicate check
└── Task 2: Backend - Add /api/user/categories/* endpoints

Wave 2 (After Wave 1 - Frontend Domain):
├── Task 3: Frontend Domain Layer (Category model with description + repository interface with merge)
└── Task 4: Frontend Infrastructure (HttpCategoryRepository + RepositoryFactory)

Wave 3 (After Wave 2):
└── Task 5: UI - Category List Display in Settings (with description)

Wave 4 (After Wave 3, can be parallel):
├── Task 6: UI - Add Category (with description)
├── Task 7: UI - Inline Edit with Merge Detection (name + description)
└── Task 8: UI - Delete with Confirmation

Wave 5 (After Wave 4):
└── Task 9: E2E Tests (including merge + description scenarios)
```

### Dependency Matrix

| Task | Depends On | Blocks | Can Parallelize With |
|------|------------|--------|---------------------|
| 0 | None | 1, 2 | None |
| 1 | 0 | 2 | None |
| 2 | 0, 1 | 3, 4 | None |
| 3 | 2 | 4, 5 | 4 |
| 4 | 2, 3 (types) | 5 | 3 (partial) |
| 5 | 3, 4 | 6, 7, 8 | None |
| 6 | 5 | 9 | 7, 8 |
| 7 | 5 | 9 | 6, 8 |
| 8 | 5 | 9 | 6, 7 |
| 9 | 6, 7, 8 | None | None (final) |

---

## TODOs

- [x] 0. Database Migration - Add Description Column to Categories

  **What to do**:
  - Create migration file: `migrations/XXXXXX_add_category_description.up.sql`
  - Add `description TEXT` column to `categories` table (nullable, default empty string)
  - Create corresponding `.down.sql` to drop the column
  - Update `internal/domain/models.go:Category` struct to include `Description string`
  - Update category repository to handle description field in CRUD operations

  **Must NOT do**:
  - Make description required (it's optional)
  - Add validation constraints in database (handle in application layer)

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: Simple migration, following existing patterns
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: NO (foundational)
  - **Parallel Group**: Wave 0
  - **Blocks**: Tasks 1, 2
  - **Blocked By**: None

  **References**:

  **Pattern References**:
  - `migrations/` - Existing migration files for pattern reference
  - `internal/domain/models.go:Category` - Current Category struct
  - `internal/adapter/repository/postgresql/category_repo.go` - Repository SQL patterns

  **Acceptance Criteria**:

  **Build Verification:**
  - [x] Migration runs: `migrate -path migrations -database $DATABASE_URL up` → SUCCESS
  - [x] Go compiles: `go build ./...` → SUCCESS
  - [x] Category struct has Description field

  **Agent-Executed QA Scenarios:**

  ```
  Scenario: Migration adds description column
    Tool: Bash
    Steps:
      1. Run migration up
      2. Query: SELECT column_name FROM information_schema.columns WHERE table_name='categories' AND column_name='description'
      3. Assert: Column exists
    Expected Result: description column present
    Evidence: SQL query output

  Scenario: Migration down removes column
    Tool: Bash
    Steps:
      1. Run migration down
      2. Query for description column
      3. Assert: Column does not exist
    Expected Result: Column removed
    Evidence: SQL query output
  ```

  **Commit**: YES
  - Message: `feat(backend): add description column to categories table`
  - Files: `migrations/*_add_category_description.*.sql`, `internal/domain/models.go`, `internal/adapter/repository/postgresql/category_repo.go`
  - Pre-commit: `go build ./...`

---

- [x] 1. Backend - Add MergeCategories Usecase Method + Explicit Duplicate Check

  **What to do**:
  - Add `MergeCategories(userID, sourceCategoryID, targetCategoryID)` method to `ManageCategoryUseCase`
  - Implement as atomic transaction:
    1. Validate both categories belong to user
    2. Validate source != target
    3. Validate neither is a default/reserved category
    4. Update all expenses: `SET category_id = target WHERE category_id = source`
    5. Delete source category (and its keywords)
    6. Return `{merged_count, message}`
  - Add `ReassignExpenses(fromCategoryID, toCategoryID)` method to `ExpenseRepository`
  - **Update `UpdateCategory` to explicitly check for duplicate names BEFORE attempting DB update**:
    - Query: Does another category with this name exist for this user?
    - If yes: Return explicit error `"category with name 'X' already exists"`
    - This provides cleaner error messages than relying on DB constraint

  **Must NOT do**:
  - Allow merging into/from default categories
  - Allow cross-user merges
  - Create separate usecase file (add to existing manage_category.go)
  - Rely only on DB constraint for duplicate detection (add explicit check)

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
    - Reason: Backend Go code with transaction handling, needs careful implementation
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: NO
  - **Parallel Group**: Wave 1
  - **Blocks**: Task 2
  - **Blocked By**: Task 0

  **References**:

  **Pattern References**:
  - `internal/usecase/manage_category.go` - Existing usecase structure, DeleteCategory for reference
  - `internal/adapter/repository/postgresql/expense_repo.go` - ExpenseRepository pattern
  - `internal/adapter/repository/postgresql/category_repo.go` - Transaction patterns

  **Domain References**:
  - `internal/domain/models.go:Expense` - Has `CategoryID *string`
  - `internal/domain/repositories.go:ExpenseRepository` - Add ReassignExpenses method

  **Acceptance Criteria**:

  **Build Verification:**
  - [x] Go compiles: `go build ./...` → SUCCESS
  - [x] Tests pass: `go test ./internal/usecase/... -v` → PASS

  **Agent-Executed QA Scenarios:**

  ```
  Scenario: Merge categories reassigns expenses
    Tool: Bash (curl + go test)
    Preconditions: Test categories and expenses exist
    Steps:
      1. Create test: source category with 5 expenses, target category with 2 expenses
      2. Call MergeCategories(user, source, target)
      3. Assert: Returns merged_count = 5
      4. Assert: Target category now has 7 expenses
      5. Assert: Source category is deleted
    Expected Result: Expenses reassigned, source deleted
    Evidence: Test output

  Scenario: Cannot merge default categories
    Tool: Bash (go test)
    Steps:
      1. Attempt to merge default category (is_default=true) as source
      2. Assert: Returns error "cannot merge default categories"
    Expected Result: Operation rejected
    Evidence: Test output

  Scenario: Cannot merge same category
    Tool: Bash (go test)
    Steps:
      1. Attempt to merge category into itself
      2. Assert: Returns error "source and target must be different"
    Expected Result: Operation rejected
    Evidence: Test output

  Scenario: UpdateCategory throws explicit error on duplicate name
    Tool: Bash (go test)
    Preconditions: Categories "Food" and "Groceries" exist for user
    Steps:
      1. Call UpdateCategory to rename "Groceries" to "Food"
      2. Assert: Returns error containing "already exists"
      3. Assert: Error message is clear and user-friendly (not DB constraint error)
    Expected Result: Explicit duplicate error before hitting DB
    Evidence: Test output
  ```

  **Commit**: YES
  - Message: `feat(backend): add MergeCategories usecase and explicit duplicate check`
  - Files: `internal/usecase/manage_category.go`, `internal/domain/repositories.go`, `internal/adapter/repository/postgresql/expense_repo.go`
  - Pre-commit: `go test ./internal/usecase/... -v`

---

- [x] 2. Backend - Add /api/user/categories/* Endpoints

  **What to do**:
  - Add new route group: `/api/user/categories`
  - Implement handlers (can reuse existing logic, just new routes):
    - `GET /api/user/categories` - List categories for user
    - `POST /api/user/categories` - Create category (with description)
    - `PUT /api/user/categories` - Update category (name + description)
    - `DELETE /api/user/categories` - Delete category
    - `POST /api/user/categories/merge` - Merge two categories
  - Request/Response includes `description` field
  - Merge endpoint:
    - Request body: `{source_id: string, target_id: string, user_id: string}`
    - Response: `{status: "success", data: {merged_count: number, message: string}}`
  - Error responses:
    - 400: Invalid request, same source/target, default category, duplicate name
    - 404: Category not found
    - 403: Category doesn't belong to user
  - Register routes in handler setup

  **Must NOT do**:
  - Modify existing `/api/categories` endpoints (keep for backwards compatibility if needed)
  - Add preview/dry-run endpoint (keep simple)
  - Add batch merge

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: Following existing handler pattern closely, mostly copying with modifications
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: NO (needs usecase from Task 1)
  - **Parallel Group**: Wave 1
  - **Blocks**: Tasks 3, 4
  - **Blocked By**: Tasks 0, 1

  **References**:

  **Pattern References**:
  - `internal/adapter/http/handler.go:UpdateCategory` (line 478) - Similar handler pattern
  - `internal/adapter/http/handler.go:DeleteCategory` - Error handling pattern

  **API Contract**:
  ```
  GET /api/user/categories?user_id=X
  Response: {status: "success", data: [{id, user_id, name, description, is_default}, ...]}

  POST /api/user/categories
  Request: {user_id, name, description?, keywords:[]}
  Response: {status: "success", data: {id, user_id, name, description, is_default}}

  PUT /api/user/categories
  Request: {id, user_id, name, description?, keywords:[]}
  Response: {status: "success", data: {id, user_id, name, description, is_default}}
  Error (duplicate): {status: "error", error: "category with name 'X' already exists"}

  DELETE /api/user/categories
  Request: {id, user_id}
  Response: {status: "success", message: "Category deleted"}

  POST /api/user/categories/merge
  Request: {source_id, target_id, user_id}
  Response: {status: "success", data: {merged_count: 15, message: "Merged 15 expenses from 'Groceries' into 'Food'"}}
  Error: {status: "error", error: "cannot merge default categories"}
  ```

  **Acceptance Criteria**:

  **Build Verification:**
  - [x] Go compiles: `go build ./...` → SUCCESS

  **Agent-Executed QA Scenarios:**

  ```
  Scenario: List categories via new endpoint
    Tool: Bash (curl)
    Preconditions: Backend running, test user with categories
    Steps:
      1. curl -X GET "http://localhost:8080/api/user/categories?user_id=test_user"
      2. Assert: HTTP 200
      3. Assert: response.data is array with categories
      4. Assert: Each category has id, name, description, is_default fields
    Expected Result: Categories listed with description field
    Evidence: curl output

  Scenario: Create category with description
    Tool: Bash (curl)
    Steps:
      1. curl -X POST http://localhost:8080/api/user/categories \
           -H "Content-Type: application/json" \
           -d '{"user_id":"test_user","name":"Groceries","description":"Weekly grocery shopping","keywords":[]}'
      2. Assert: HTTP 200
      3. Assert: response.data.description = "Weekly grocery shopping"
    Expected Result: Category created with description
    Evidence: curl output

  Scenario: Update returns explicit duplicate error
    Tool: Bash (curl)
    Preconditions: Categories "Food" and "Groceries" exist
    Steps:
      1. curl -X PUT http://localhost:8080/api/user/categories \
           -d '{"id":"<groceries-id>","user_id":"test_user","name":"Food"}'
      2. Assert: HTTP 400
      3. Assert: response.error contains "already exists"
    Expected Result: Clear duplicate error message
    Evidence: curl output

  Scenario: Merge API returns success with count
    Tool: Bash (curl)
    Preconditions: Backend running, test user with categories
    Steps:
      1. Create source category "Groceries" with 3 expenses
      2. Create target category "Food"
      3. curl -X POST http://localhost:8080/api/user/categories/merge \
           -H "Content-Type: application/json" \
           -d '{"source_id":"<source>","target_id":"<target>","user_id":"test_user"}'
      4. Assert: HTTP 200
      5. Assert: response.data.merged_count = 3
      6. Assert: response.data.message contains "Merged 3 expenses"
    Expected Result: Merge succeeds with count
    Evidence: curl output
  ```

  **Commit**: YES
  - Message: `feat(backend): add /api/user/categories endpoints with description and merge`
  - Files: `internal/adapter/http/handler.go`
  - Pre-commit: `go build ./...`

---

- [x] 3. Frontend - Create Category Domain Layer

  **What to do**:
  - Create `Category` interface in `frontend/dashboard/src/domain/models/Category.ts`
  - Create `CategoryRepository` interface in `frontend/dashboard/src/domain/repositories/CategoryRepository.ts`
  - Category fields: `id`, `user_id`, `name`, `description`, `is_default`
  - Repository methods: `list()`, `create(name, description?)`, `update(id, name, description?)`, `delete(id)`, **`merge(sourceId, targetId)`**
  - Add `MergeResult` type: `{merged_count: number, message: string}`

  **Must NOT do**:
  - Add keyword-related fields or methods
  - Add icon/color fields

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: Simple interface definitions, low complexity
  - **Skills**: []
    - No special skills needed for TypeScript interfaces

  **Parallelization**:
  - **Can Run In Parallel**: YES (with Task 4 partially)
  - **Parallel Group**: Wave 2
  - **Blocks**: Tasks 4, 5
  - **Blocked By**: Task 2 (backend API must exist)

  **References**:

  **Pattern References**:
  - `frontend/dashboard/src/domain/models/Expense.ts` - Interface pattern to follow
  - `frontend/dashboard/src/domain/models/User.ts` - User model structure
  - `frontend/dashboard/src/domain/repositories/ExpenseRepository.ts` - Repository interface pattern

  **API References**:
  - `internal/domain/models.go:Category` - Backend Category struct with exact fields
  - `POST /api/user/categories/merge` response shape for MergeResult type

  **Acceptance Criteria**:

  **Build Verification:**
  - [x] TypeScript compiles: `cd frontend/dashboard && bun run build` → SUCCESS

  **Agent-Executed QA Scenarios:**

  ```
  Scenario: Category type matches backend model
    Tool: Bash (grep + TypeScript)
    Preconditions: Files created
    Steps:
      1. Verify Category interface has: id, user_id, name, description, is_default
      2. Verify CategoryRepository has: list, create, update, delete, merge methods
      3. Verify MergeResult type has: merged_count, message
      4. Verify create/update methods accept optional description parameter
      5. Run: cd frontend/dashboard && bun run build
    Expected Result: Build succeeds, no type errors
    Evidence: Build output captured
  ```

  **Commit**: YES
  - Message: `feat(dashboard): add Category domain model with description and repository interface`
  - Files: `src/domain/models/Category.ts`, `src/domain/repositories/CategoryRepository.ts`
  - Pre-commit: `bun run build`

---

- [x] 4. Frontend - Create HttpCategoryRepository and Update RepositoryFactory

  **What to do**:
  - Create `HttpCategoryRepository` in `frontend/dashboard/src/infrastructure/repositories/http/HttpCategoryRepository.ts`
  - **Use new endpoints**: `/api/user/categories` (not `/api/categories`)
  - Implement all repository methods using axios including **`merge(sourceId, targetId)`**
  - Handle API response envelope: `{status, data, message, error}`
  - Pass `description` field in create/update requests
  - Update `RepositoryFactory.ts` to add `getCategoryRepository()` method
  - Use token-based auth like other repositories

  **Must NOT do**:
  - Create InMemoryCategoryRepository
  - Add caching layer
  - Use old `/api/categories` endpoints

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: Following existing HttpExpenseRepository pattern closely
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: Partial (needs Category type from Task 3)
  - **Parallel Group**: Wave 2
  - **Blocks**: Task 5
  - **Blocked By**: Tasks 2 (backend), 3 (types)

  **References**:

  **Pattern References**:
  - `frontend/dashboard/src/infrastructure/repositories/http/HttpExpenseRepository.ts:1-50` - HTTP repository pattern
  - `frontend/dashboard/src/infrastructure/repositories/http/HttpUserRepository.ts` - Another HTTP repo example
  - `frontend/dashboard/src/infrastructure/RepositoryFactory.ts` - Factory pattern to extend

  **API References**:
  - `GET /api/user/categories?user_id=X` → `{data: Category[]}`
  - `POST /api/user/categories` body: `{user_id, name, description?, keywords:[]}` → Created category
  - `PUT /api/user/categories` body: `{id, user_id, name, description?, keywords:[]}` → Updated category
  - `DELETE /api/user/categories` body: `{id, user_id}` → Deleted category
  - **`POST /api/user/categories/merge`** body: `{source_id, target_id, user_id}` → `{merged_count, message}`

  **Acceptance Criteria**:

  **Build Verification:**
  - [x] TypeScript compiles: `cd frontend/dashboard && bun run build` → SUCCESS
  - [x] RepositoryFactory exports `getCategoryRepository`

  **Agent-Executed QA Scenarios:**

  ```
  Scenario: Repository factory returns CategoryRepository with merge
    Tool: Bash (TypeScript)
    Preconditions: Files created and compiled
    Steps:
      1. Verify HttpCategoryRepository implements CategoryRepository interface
      2. Verify all API calls use /api/user/categories path
      3. Verify merge() method exists and calls POST /api/user/categories/merge
      4. Verify RepositoryFactory.getCategoryRepository() method exists
      5. Run: cd frontend/dashboard && bun run build
    Expected Result: Build succeeds
    Evidence: Build output
  ```

  **Commit**: YES
  - Message: `feat(dashboard): add HttpCategoryRepository with /api/user endpoints and merge`
  - Files: `src/infrastructure/repositories/http/HttpCategoryRepository.ts`, `src/infrastructure/RepositoryFactory.ts`
  - Pre-commit: `bun run build`

---

- [x] 5. UI - Add Category List Display to Settings Page

  **What to do**:
  - Add "Category Management" section to Settings page after currency section
  - Fetch categories on page load using CategoryRepository
  - Display categories in a list with glass-panel styling
  - **Show both name and description for each category**
  - Show visual distinction for default vs custom categories (default = muted/disabled look)
  - Add loading state while fetching
  - Handle empty state (will always have defaults, but handle gracefully)
  - Description displayed as secondary text below category name (muted color)

  **Must NOT do**:
  - Add edit/delete buttons yet (next tasks)
  - Create a separate route
  - Add keyword display

  **Recommended Agent Profile**:
  - **Category**: `visual-engineering`
    - Reason: UI component with styling, needs visual consistency
  - **Skills**: [`frontend-ui-ux`]
    - `frontend-ui-ux`: Glass-panel styling, visual hierarchy

  **Parallelization**:
  - **Can Run In Parallel**: NO
  - **Parallel Group**: Wave 3 (sequential)
  - **Blocks**: Tasks 6, 7, 8
  - **Blocked By**: Tasks 3, 4

  **References**:

  **Pattern References**:
  - `frontend/dashboard/src/app/[locale]/dashboard/settings/page.tsx:95-142` - Glass-panel section structure
  - `frontend/dashboard/src/components/ExpenseList.tsx:1-50` - List rendering pattern

  **Styling References**:
  - Use existing Tailwind classes: `glass-panel`, `p-6`, `rounded-2xl`, `border border-white/5`
  - Icons from `@heroicons/react/24/outline`
  - Description text: `text-sm text-white/60` or similar muted style

  **Acceptance Criteria**:

  **Agent-Executed QA Scenarios:**

  ```
  Scenario: Settings page shows Category Management section with descriptions
    Tool: Playwright (playwright skill)
    Preconditions: Dev server running, test user exists with categories (some with descriptions)
    Steps:
      1. Navigate to: http://localhost:3000/en/dashboard/settings?token=test_user
      2. Wait for: text "Category Management" visible (timeout: 10s)
      3. Assert: At least one category name visible (e.g., "Food")
      4. Assert: Description text visible for categories that have one
      5. Assert: Section appears after currency section
      6. Screenshot: .sisyphus/evidence/task-5-category-list.png
    Expected Result: Category Management section visible with names and descriptions
    Evidence: .sisyphus/evidence/task-5-category-list.png

  Scenario: Default categories have distinct visual style
    Tool: Playwright
    Preconditions: Dev server running
    Steps:
      1. Navigate to settings page with token
      2. Locate default category row (e.g., "Food")
      3. Assert: Has visual indicator (opacity, badge, or text like "Default")
    Expected Result: Default categories visually distinguished
    Evidence: Screenshot captured
  ```

  **Commit**: YES
  - Message: `feat(dashboard): add Category Management section to Settings with description display`
  - Files: `src/app/[locale]/dashboard/settings/page.tsx`
  - Pre-commit: `bun run build`

---

- [x] 6. UI - Add "Add Category" Functionality

  **What to do**:
  - Add "Add Category" button/input at bottom of category list
  - Show inline input fields when clicking add: name (required) + description (optional)
  - Validate: name non-empty, max 50 characters; description max 200 characters
  - Call repository.create(name, description) on submit
  - Update list after successful creation
  - Show error message if API fails (e.g., duplicate name)
  - Disable button during save operation

  **Must NOT do**:
  - Add modal dialog
  - Add keyword input
  - Pre-validate uniqueness on client

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: Single feature addition, following existing patterns
  - **Skills**: [`frontend-ui-ux`]

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 4 (with Tasks 7, 8)
  - **Blocks**: Task 9
  - **Blocked By**: Task 5

  **References**:

  **Pattern References**:
  - `frontend/dashboard/src/components/ExpenseList.tsx` - Inline form pattern
  - Existing Settings page button styling

  **Acceptance Criteria**:

  **Agent-Executed QA Scenarios:**

  ```
  Scenario: User can add a new category with description
    Tool: Playwright
    Preconditions: Dev server running, user on settings page
    Steps:
      1. Navigate to: http://localhost:3000/en/dashboard/settings?token=test_user
      2. Click: button or link with text "Add Category" or "+"
      3. Wait for: input fields visible
      4. Fill: name input with "Groceries"
      5. Fill: description input with "Weekly grocery shopping"
      6. Click: Save/Submit button
      7. Wait for: "Groceries" appears in category list (timeout: 5s)
      8. Assert: Description "Weekly grocery shopping" visible
      9. Screenshot: .sisyphus/evidence/task-6-add-category.png
    Expected Result: New category "Groceries" visible with description
    Evidence: .sisyphus/evidence/task-6-add-category.png

  Scenario: Empty category name shows validation error
    Tool: Playwright
    Preconditions: Dev server running
    Steps:
      1. Navigate to settings, click Add Category
      2. Leave name input empty, click Save
      3. Assert: Error message visible OR button disabled
    Expected Result: Cannot submit empty name
    Evidence: Screenshot captured

  Scenario: Duplicate category name shows error
    Tool: Playwright
    Preconditions: Category "Food" already exists
    Steps:
      1. Navigate to settings, click Add Category
      2. Type "Food" in name input, click Save
      3. Wait for: Error message visible (timeout: 5s)
      4. Assert: Error text contains "already exists" or similar
    Expected Result: API error shown to user
    Evidence: Screenshot with error message
  ```

  **Commit**: YES
  - Message: `feat(dashboard): add create category functionality with description`
  - Files: `src/app/[locale]/dashboard/settings/page.tsx`
  - Pre-commit: `bun run build`

---

- [x] 7. UI - Add Inline Edit with Merge Detection (KEY FEATURE)

  **What to do**:
  - Add Edit button/icon to each category row (disabled for default categories)
  - Track `editingId` state (only one category editable at a time)
  - Replace category display with input fields when editing (name + description)
  - Show Save/Cancel buttons inline
  - Validate: name non-empty, max 50 characters; description max 200 characters
  - Call repository.update() on save
  - **Handle duplicate name error**:
    1. Catch "already exists" error from API
    2. Show merge confirmation: "Category 'X' already exists. Merge expenses into 'X'?"
    3. If user confirms, call repository.merge(editingId, existingCategoryId)
    4. Show result: "Merged N expenses from 'A' into 'B'"
  - Handle other API errors gracefully

  **Must NOT do**:
  - Allow editing default categories
  - Open modal for editing
  - Add keyword editing

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
    - Reason: Complex state management with merge flow
  - **Skills**: [`frontend-ui-ux`]

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 4 (with Tasks 6, 8)
  - **Blocks**: Task 9
  - **Blocked By**: Task 5

  **References**:

  **Pattern References**:
  - `frontend/dashboard/src/components/ExpenseList.tsx:100-180` - Inline edit pattern with editingId state

  **Acceptance Criteria**:

  **Agent-Executed QA Scenarios:**

  ```
  Scenario: User can rename a custom category and edit description
    Tool: Playwright
    Preconditions: Dev server running, custom category "Groceries" exists
    Steps:
      1. Navigate to settings page with token
      2. Locate row for "Groceries"
      3. Click: Edit button on that row
      4. Wait for: Input fields visible with current values
      5. Clear name input, type "Supermarket"
      6. Clear description, type "Big box stores"
      7. Click: Save button
      8. Wait for: "Supermarket" appears in list, "Groceries" gone
      9. Assert: Description shows "Big box stores"
      10. Screenshot: .sisyphus/evidence/task-7-edit-category.png
    Expected Result: Category renamed with new description
    Evidence: .sisyphus/evidence/task-7-edit-category.png

  Scenario: Rename to existing name shows merge confirmation
    Tool: Playwright
    Preconditions: Categories "Food" and "Groceries" exist, Groceries has 5 expenses
    Steps:
      1. Navigate to settings page
      2. Click Edit on "Groceries"
      3. Change name to "Food"
      4. Click Save
      5. Wait for: Merge confirmation visible (timeout: 5s)
      6. Assert: Confirmation text mentions merging into "Food"
      7. Screenshot: .sisyphus/evidence/task-7-merge-confirm.png
    Expected Result: Merge confirmation shown
    Evidence: .sisyphus/evidence/task-7-merge-confirm.png

  Scenario: User confirms merge and sees result
    Tool: Playwright
    Preconditions: Merge confirmation visible from previous scenario
    Steps:
      1. Click Confirm on merge dialog
      2. Wait for: Success message visible (timeout: 5s)
      3. Assert: Message contains "Merged" and expense count
      4. Assert: "Groceries" no longer in list
      5. Screenshot: .sisyphus/evidence/task-7-merge-success.png
    Expected Result: Categories merged, source deleted
    Evidence: .sisyphus/evidence/task-7-merge-success.png

  Scenario: Default categories cannot be edited
    Tool: Playwright
    Preconditions: Dev server running
    Steps:
      1. Navigate to settings page
      2. Locate default category row (e.g., "Food" with is_default=true)
      3. Assert: Edit button is disabled OR not present
    Expected Result: No edit option for default categories
    Evidence: Screenshot showing disabled state

  Scenario: Cancel edit restores original values
    Tool: Playwright
    Steps:
      1. Click Edit on a category
      2. Change the name and description in inputs
      3. Click Cancel
      4. Assert: Original name and description displayed
    Expected Result: Edit cancelled, no changes persisted
    Evidence: Screenshot
  ```

  **Commit**: YES
  - Message: `feat(dashboard): add inline edit for categories with merge detection`
  - Files: `src/app/[locale]/dashboard/settings/page.tsx`
  - Pre-commit: `bun run build`

---

- [x] 8. UI - Add Delete with Inline Confirmation

  **What to do**:
  - Add Delete button/icon to each category row (disabled for default categories)
  - On click, show inline confirmation: "Delete [Name]? [Confirm] [Cancel]"
  - Call repository.delete() on confirm
  - Handle FK violation error (category in use) with user-friendly message
  - Update list after successful deletion

  **Must NOT do**:
  - Create modal/dialog component
  - Allow deleting default categories
  - Add batch delete

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: Inline confirmation, simple state management
  - **Skills**: [`frontend-ui-ux`]

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 4 (with Tasks 6, 7)
  - **Blocks**: Task 9
  - **Blocked By**: Task 5

  **References**:

  **Pattern References**:
  - Inline confirmation can use same row layout with conditional rendering
  - No existing delete confirmation pattern, use simple inline approach

  **Acceptance Criteria**:

  **Agent-Executed QA Scenarios:**

  ```
  Scenario: User can delete a custom category
    Tool: Playwright
    Preconditions: Custom category "Groceries" exists, not used by any expenses
    Steps:
      1. Navigate to settings page with token
      2. Locate row for "Groceries"
      3. Click: Delete button
      4. Wait for: Confirmation text visible "Delete Groceries?"
      5. Click: Confirm button
      6. Wait for: "Groceries" disappears from list (timeout: 5s)
      7. Screenshot: .sisyphus/evidence/task-8-delete-category.png
    Expected Result: Category deleted, no longer in list
    Evidence: .sisyphus/evidence/task-8-delete-category.png

  Scenario: Cancel delete keeps category
    Tool: Playwright
    Steps:
      1. Click Delete on a category
      2. Wait for confirmation UI
      3. Click Cancel
      4. Assert: Category still visible in list
    Expected Result: Deletion cancelled
    Evidence: Screenshot

  Scenario: Default categories cannot be deleted
    Tool: Playwright
    Steps:
      1. Navigate to settings page
      2. Locate default category row (e.g., "Food")
      3. Assert: Delete button is disabled OR not present
    Expected Result: No delete option for defaults
    Evidence: Screenshot

  Scenario: Deleting category in use shows error
    Tool: Playwright
    Preconditions: Category has expenses attached
    Steps:
      1. Try to delete category that has expenses
      2. Confirm deletion
      3. Wait for: Error message visible
      4. Assert: Error mentions "in use" or "cannot delete"
    Expected Result: User sees helpful error, category not deleted
    Evidence: Screenshot with error
  ```

  **Commit**: YES
  - Message: `feat(dashboard): add delete category with inline confirmation`
  - Files: `src/app/[locale]/dashboard/settings/page.tsx`
  - Pre-commit: `bun run build`

---

- [x] 9. Create Playwright E2E Tests for Category Management

  **What to do**:
  - Create `frontend/dashboard/tests/settings-categories.spec.ts`
  - Test all category operations: list, add (with description), edit (with description), delete
  - Test default category protection (no edit/delete)
  - Test error handling (duplicate name, category in use)
  - **Test merge flow** (rename to existing → confirm merge → verify result)
  - Follow existing test patterns from `dashboard.spec.ts`
  - Use `data-testid` attributes for reliable selection

  **Must NOT do**:
  - Create unit tests for this task (optional separate task)
  - Test keyword functionality

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: Following existing test patterns
  - **Skills**: [`playwright`]
    - `playwright`: E2E test writing and execution

  **Parallelization**:
  - **Can Run In Parallel**: NO (depends on all features complete)
  - **Parallel Group**: Wave 5 (final)
  - **Blocks**: None (final task)
  - **Blocked By**: Tasks 6, 7, 8

  **References**:

  **Pattern References**:
  - `frontend/dashboard/tests/dashboard.spec.ts` - Test structure and patterns
  - `frontend/dashboard/tests/report-flow.spec.ts` - Another E2E test example
  - `frontend/dashboard/playwright.config.ts` - Playwright configuration

  **Acceptance Criteria**:

  **Agent-Executed QA Scenarios:**

  ```
  Scenario: All E2E tests pass
    Tool: Bash
    Preconditions: Dev server running, all features implemented
    Steps:
      1. cd frontend/dashboard
      2. bunx playwright test settings-categories --reporter=list
      3. Assert: Exit code 0
      4. Assert: All tests pass
    Expected Result: All category tests pass
    Evidence: Terminal output captured

  Scenario: Tests cover all operations including merge and description
    Tool: Bash (grep)
    Steps:
      1. Verify test file contains tests for:
         - "displays categories with descriptions"
         - "can add category with description"  
         - "can edit category name and description"
         - "can delete category"
         - "default categories protected"
         - "merge flow when renaming to existing"
    Expected Result: All key scenarios have tests
    Evidence: grep output
  ```

  **Commit**: YES
  - Message: `test(dashboard): add Playwright E2E tests for category management with merge and description`
  - Files: `tests/settings-categories.spec.ts`
  - Pre-commit: `bunx playwright test settings-categories`

---

## Commit Strategy

| After Task | Message | Files | Verification |
|------------|---------|-------|--------------|
| 0 | `feat(backend): add description column to categories table` | migrations/*.sql, models.go, category_repo.go | `go build ./...` |
| 1 | `feat(backend): add MergeCategories usecase and explicit duplicate check` | manage_category.go, repositories.go, expense_repo.go | `go test ./internal/usecase/... -v` |
| 2 | `feat(backend): add /api/user/categories endpoints with description and merge` | handler.go | `go build ./...` |
| 3 | `feat(dashboard): add Category domain model with description and repository interface` | Category.ts, CategoryRepository.ts | `bun run build` |
| 4 | `feat(dashboard): add HttpCategoryRepository with /api/user endpoints and merge` | HttpCategoryRepository.ts, RepositoryFactory.ts | `bun run build` |
| 5 | `feat(dashboard): add Category Management section to Settings with description display` | settings/page.tsx | `bun run build` |
| 6 | `feat(dashboard): add create category functionality with description` | settings/page.tsx | `bun run build` |
| 7 | `feat(dashboard): add inline edit for categories with merge detection` | settings/page.tsx | `bun run build` |
| 8 | `feat(dashboard): add delete category with inline confirmation` | settings/page.tsx | `bun run build` |
| 9 | `test(dashboard): add Playwright E2E tests for category management` | settings-categories.spec.ts | `bunx playwright test settings-categories` |

---

## Success Criteria

### Verification Commands
```bash
# Backend build and tests
go build ./...  # Expected: SUCCESS
go test ./internal/usecase/... -v  # Expected: All tests pass

# Frontend build verification
cd frontend/dashboard && bun run build  # Expected: SUCCESS

# E2E tests
cd frontend/dashboard && bunx playwright test settings-categories  # Expected: All tests pass

# Full test suite
cd frontend/dashboard && bunx playwright test  # Expected: All tests pass
```

### Final Checklist
- [x] Migration adds description column to categories
- [x] Settings page has "Category Management" section
- [x] Categories list displays with name, description, and default/custom distinction
- [x] Users can add new categories with optional description
- [x] Users can rename custom categories and edit descriptions (inline)
- [x] **Explicit error shown when renaming to existing name**
- [x] **Merge confirmation flow works when renaming to existing name**
- [x] Users can delete custom categories (with confirmation)
- [x] Default categories are protected (no edit/delete)
- [x] Error messages shown for failures
- [x] All E2E tests pass
- [x] **All category endpoints use /api/user/categories path**
- [x] No keyword management (guardrail verified)
- [x] No separate route created (guardrail verified)
