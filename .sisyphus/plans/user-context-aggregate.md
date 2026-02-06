# User Context Aggregate for Prompt Personalization

## TL;DR

> **Quick Summary**: Add a `UserContext` aggregate to load user settings (HomeCurrency, Locale) and categories, then inject this context into AI prompts for personalized expense parsing.
> 
> **Deliverables**:
> - `UserContext` struct in `internal/domain/models.go`
> - `GetUserContextUseCase` in `internal/usecase/get_user_context.go`
> - Updated `ai.Service` interface to accept `*domain.UserContext`
> - Modified prompts in `internal/ai/gemini.go` with user-specific data
> - Unit tests for new code
> 
> **Estimated Effort**: Medium
> **Parallel Execution**: YES - 2 waves
> **Critical Path**: Task 1 (types) -> Task 2 (use case) -> Task 4 (interface) -> Task 5 (wiring)

---

## Context

### Original Request
Enable personalized AI prompts by injecting user-specific context (home currency, locale, category names) into expense parsing prompts.

### Interview Summary
**Key Discussions**:
- User model has HomeCurrency and Locale but prompts use hardcoded "TWD" and static category list
- Categories are separate entities loaded via `CategoryRepository.GetByUserID()`
- OpenSpec proposal approved at `openspec/changes/add-user-context-aggregate/`

**Research Findings**:
- Prompt injection points: `gemini.go:199` (ParseExpense) and `gemini.go:297` (SuggestCategory)
- Existing repositories: `UserRepository.GetByID()`, `CategoryRepository.GetByUserID()`
- Test infrastructure exists: `go test ./...` with mocks in `internal/usecase/mocks.go`

### Metis Review
**Identified Gaps** (addressed):
- New user without categories: Use hardcoded defaults as fallback
- Interface change strategy: Option A (breaking) - replace `userID` param with `UserContext`
- UserContext loading location: In `ParseConversationUseCase` - cleanest ownership
- `SuggestCategory` inclusion: Also receives UserContext for consistent behavior

---

## Work Objectives

### Core Objective
Enable AI prompts to include user-specific settings and category names for better expense parsing accuracy.

### Concrete Deliverables
- `internal/domain/models.go`: Add `UserContext` struct
- `internal/usecase/get_user_context.go`: New file with use case
- `internal/usecase/get_user_context_test.go`: New test file
- `internal/ai/service.go`: Updated interface
- `internal/ai/gemini.go`: Updated prompt construction
- `internal/ai/gemini_test.go`: Updated tests
- `internal/usecase/parse_conversation.go`: Wire context loading

### Definition of Done
- [x] `go build ./...` succeeds with no errors
- [x] `go test ./... -v` all tests pass
- [x] `openspec validate add-user-context-aggregate --strict` passes
- [x] Prompts include user's HomeCurrency (not hardcoded TWD)
- [x] Prompts include user's category names (not hardcoded list)
- [x] When user has no custom categories, falls back to defaults

### Must Have
- `UserContext` struct with `User *User` and `Categories []*Category`
- Use case to load UserContext by userID
- Updated AI service interface accepting `*domain.UserContext`
- Graceful fallback when UserContext is nil

### Must NOT Have (Guardrails)
- NO caching for UserContext (premature optimization)
- NO lazy loading / partial loading
- NO database schema changes or migrations
- NO new API endpoints
- NO modification to `ProcessMessageUseCase` beyond adding context loading call
- NO more than 2 new files
- NO category keywords in UserContext (only names needed)
- NO locale-based prompt translation

---

## Verification Strategy (MANDATORY)

> **UNIVERSAL RULE: ZERO HUMAN INTERVENTION**
>
> ALL tasks in this plan MUST be verifiable WITHOUT any human action.

### Test Decision
- **Infrastructure exists**: YES (`go test ./...`)
- **Automated tests**: YES (TDD approach for new use case)
- **Framework**: Go standard testing + existing mocks

### Agent-Executed QA Scenarios (MANDATORY - ALL tasks)

Verification tools by deliverable type:
| Type | Tool | How Agent Verifies |
|------|------|-------------------|
| Go code compilation | Bash | `go build ./...` exits 0 |
| Unit tests | Bash | `go test ./... -v` exits 0 |
| Type existence | Bash | `go doc` command shows type |
| Integration | Bash | curl to local server, parse JSON response |

---

## Execution Strategy

### Parallel Execution Waves

```
Wave 1 (Start Immediately):
├── Task 1: Define UserContext struct (no dependencies)
└── Task 3: Create GetUserContext use case tests (can write tests first)

Wave 2 (After Wave 1):
├── Task 2: Implement GetUserContext use case (depends: 1)
├── Task 4: Update AI service interface (depends: 1)
└── Task 5: Update prompt construction (depends: 4)

Wave 3 (After Wave 2):
└── Task 6: Wire into ParseConversationUseCase (depends: 2, 4, 5)

Wave 4 (After Wave 3):
└── Task 7: Final integration test and OpenSpec update (depends: 6)

Critical Path: Task 1 → Task 2 → Task 4 → Task 5 → Task 6
```

### Dependency Matrix

| Task | Depends On | Blocks | Can Parallelize With |
|------|------------|--------|---------------------|
| 1 | None | 2, 3, 4 | 3 |
| 2 | 1 | 6 | 4, 5 |
| 3 | 1 | 2 | 1 (start test first) |
| 4 | 1 | 5, 6 | 2 |
| 5 | 4 | 6 | 2 |
| 6 | 2, 4, 5 | 7 | None |
| 7 | 6 | None | None |

---

## TODOs

- [x] 1. Define UserContext struct in domain

  **What to do**:
  - Add `UserContext` struct to `internal/domain/models.go` after the `User` struct
  - Fields: `User *User` and `Categories []*Category`
  - Add interface compliance check if needed

  **Must NOT do**:
  - Create a new file for this single struct
  - Add any methods to UserContext (keep it as a simple data struct)
  - Add fields beyond User and Categories

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: Single struct addition, trivial change
  - **Skills**: `[]`
    - No special skills needed for simple Go struct

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1 (with Task 3)
  - **Blocks**: Tasks 2, 3, 4
  - **Blocked By**: None

  **References**:
  - `internal/domain/models.go:8-15` - Existing User struct pattern to follow
  - `internal/domain/models.go:67-73` - Category struct definition

  **Acceptance Criteria**:

  **Build verification:**
  - [ ] `go build ./...` exits 0

  **Type verification:**
  - [ ] `go doc github.com/riverlin/aiexpense/internal/domain.UserContext` shows struct definition

  **Agent-Executed QA Scenarios:**

  ```
  Scenario: UserContext struct compiles correctly
    Tool: Bash
    Preconditions: None
    Steps:
      1. go build ./...
      2. Assert: exit code is 0
      3. go doc github.com/riverlin/aiexpense/internal/domain.UserContext
      4. Assert: output contains "type UserContext struct"
      5. Assert: output contains "User *User"
      6. Assert: output contains "Categories []*Category"
    Expected Result: Build succeeds, struct documented
    Evidence: Command outputs captured
  ```

  **Commit**: NO (groups with 2)

---

- [x] 2. Implement GetUserContext use case

  **What to do**:
  - Create new file `internal/usecase/get_user_context.go`
  - Define `GetUserContextUseCase` struct with `userRepo` and `categoryRepo` fields
  - Implement `Execute(ctx context.Context, userID string) (*domain.UserContext, error)` method
  - Load user via `userRepo.GetByID()`, return error if not found
  - Load categories via `categoryRepo.GetByUserID()`, return empty slice if none
  - Return `&domain.UserContext{User: user, Categories: categories}`

  **Must NOT do**:
  - Add caching logic
  - Add lazy loading
  - Add more than one public method

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: Single file use case following established pattern
  - **Skills**: `[]`
    - Standard Go use case, no special tools needed

  **Parallelization**:
  - **Can Run In Parallel**: NO (depends on Task 1)
  - **Parallel Group**: Wave 2
  - **Blocks**: Task 6
  - **Blocked By**: Task 1

  **References**:
  - `internal/usecase/parse_conversation.go:15-39` - Use case struct and constructor pattern to follow exactly
  - `internal/domain/repositories.go:8-21` - UserRepository interface
  - `internal/domain/repositories.go:62-90` - CategoryRepository interface

  **Acceptance Criteria**:

  **Build verification:**
  - [ ] `go build ./...` exits 0

  **Test verification:**
  - [ ] `go test ./internal/usecase -run TestGetUserContext -v` passes

  **Agent-Executed QA Scenarios:**

  ```
  Scenario: GetUserContext returns user with categories
    Tool: Bash
    Preconditions: Task 1 and Task 3 completed
    Steps:
      1. go test ./internal/usecase -run TestGetUserContext -v
      2. Assert: exit code is 0
      3. Assert: output contains "PASS"
    Expected Result: Use case tests pass
    Evidence: Test output captured

  Scenario: GetUserContext handles missing user
    Tool: Bash
    Preconditions: Test includes error case
    Steps:
      1. go test ./internal/usecase -run TestGetUserContext_UserNotFound -v
      2. Assert: output contains "PASS" or test exists
    Expected Result: Error handling tested
    Evidence: Test output captured
  ```

  **Commit**: YES
  - Message: `feat(usecase): add GetUserContext use case for prompt personalization`
  - Files: `internal/usecase/get_user_context.go`, `internal/usecase/get_user_context_test.go`, `internal/domain/models.go`
  - Pre-commit: `go test ./internal/usecase -v`

---

- [x] 3. Write GetUserContext use case tests (TDD)

  **What to do**:
  - Create new file `internal/usecase/get_user_context_test.go`
  - Write table-driven tests covering:
    - Happy path: user exists with categories
    - User exists with no categories (returns empty slice)
    - User not found (returns error)
  - Use existing mocks from `internal/usecase/mocks.go` or create minimal mocks

  **Must NOT do**:
  - Create integration tests (unit tests only)
  - Test with real database

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: Standard Go table-driven test pattern
  - **Skills**: `[]`
    - Standard Go testing, no special skills needed

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1 (with Task 1)
  - **Blocks**: Task 2 (tests written first)
  - **Blocked By**: Task 1 (needs UserContext type)

  **References**:
  - `internal/usecase/create_expense_test.go` - Existing test pattern with mocks
  - `internal/usecase/parse_conversation_test.go` - If exists, follow its mock patterns
  - `internal/usecase/auto_signup_test.go` - Test structure reference

  **Acceptance Criteria**:

  **Test file exists:**
  - [ ] File `internal/usecase/get_user_context_test.go` created
  - [ ] Tests compile: `go build ./internal/usecase/...` exits 0

  **Agent-Executed QA Scenarios:**

  ```
  Scenario: Test file compiles and has correct structure
    Tool: Bash
    Preconditions: Task 1 completed
    Steps:
      1. ls internal/usecase/get_user_context_test.go
      2. Assert: file exists
      3. go build ./internal/usecase/...
      4. Assert: exit code is 0
      5. grep -c "func Test" internal/usecase/get_user_context_test.go
      6. Assert: output is at least 2 (multiple test cases)
    Expected Result: Test file created with multiple test functions
    Evidence: File exists, builds, has test functions
  ```

  **Commit**: NO (groups with Task 2)

---

- [x] 4. Update AI service interface to accept UserContext

  **What to do**:
  - Modify `internal/ai/service.go` interface
  - Change `ParseExpense(ctx, text, userID)` to `ParseExpense(ctx, text string, userCtx *domain.UserContext)`
  - Change `SuggestCategory(ctx, description, userID)` to `SuggestCategory(ctx, description string, userCtx *domain.UserContext)`
  - Update `internal/ai/gemini.go` method signatures to match
  - Add import for `domain` package if not present

  **Must NOT do**:
  - Add new methods (modify existing)
  - Change return types
  - Add optional parameters

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: Interface signature change, straightforward
  - **Skills**: `[]`
    - Standard Go interface modification

  **Parallelization**:
  - **Can Run In Parallel**: YES (after Task 1)
  - **Parallel Group**: Wave 2 (with Tasks 2, 5)
  - **Blocks**: Tasks 5, 6
  - **Blocked By**: Task 1

  **References**:
  - `internal/ai/service.go:6-14` - Current interface definition to modify
  - `internal/ai/gemini.go:19` - Interface compliance check pattern
  - `internal/ai/gemini.go:53-79` - ParseExpense method to update
  - `internal/ai/gemini.go:438-458` - SuggestCategory method to update

  **Acceptance Criteria**:

  **Build verification:**
  - [ ] `go build ./...` exits 0 (all callers updated)

  **Interface verification:**
  - [ ] `go doc github.com/riverlin/aiexpense/internal/ai.Service` shows new signatures

  **Agent-Executed QA Scenarios:**

  ```
  Scenario: Interface updated and builds
    Tool: Bash
    Preconditions: Task 1 completed
    Steps:
      1. go build ./...
      2. Assert: exit code is 0 (no interface mismatch)
      3. go doc github.com/riverlin/aiexpense/internal/ai.Service
      4. Assert: output contains "ParseExpense" with "UserContext"
      5. Assert: output does NOT contain "userID string" for ParseExpense
    Expected Result: Interface updated, builds clean
    Evidence: Command outputs captured
  ```

  **Commit**: NO (groups with Task 5)

---

- [x] 5. Update prompt construction with user context

  **What to do**:
  - Modify `internal/ai/gemini.go` `callGeminiAPI()` method (line ~199)
    - Accept `userCtx *domain.UserContext` parameter
    - Extract `homeCurrency` from `userCtx.User.HomeCurrency` (default "TWD" if empty)
    - Extract category names from `userCtx.Categories` (default to hardcoded list if empty)
    - Update prompt template to use extracted values
  - Modify `callGeminiCategoryAPI()` method (line ~297)
    - Accept `userCtx *domain.UserContext` parameter
    - Use user's categories in the category list if available
  - Handle nil `userCtx` gracefully (use defaults)

  **Must NOT do**:
  - Add locale-based translation
  - Add caching
  - Change prompt structure beyond injecting user data

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: String interpolation changes in existing code
  - **Skills**: `[]`
    - Standard Go string formatting

  **Parallelization**:
  - **Can Run In Parallel**: YES (after Task 4)
  - **Parallel Group**: Wave 2 (after interface updated)
  - **Blocks**: Task 6
  - **Blocked By**: Task 4

  **References**:
  - `internal/ai/gemini.go:199-217` - Current hardcoded prompt to modify
  - `internal/ai/gemini.go:297-312` - Category suggestion prompt to modify
  - `internal/ai/gemini.go:53-79` - ParseExpense method structure

  **Acceptance Criteria**:

  **Build verification:**
  - [ ] `go build ./...` exits 0

  **Test verification:**
  - [ ] `go test ./internal/ai -v` passes

  **Prompt content verification:**
  - [ ] Grep for "TWD" hardcoded - should only appear in fallback/default logic

  **Agent-Executed QA Scenarios:**

  ```
  Scenario: Prompts use user context when provided
    Tool: Bash
    Preconditions: Tasks 1, 4 completed
    Steps:
      1. go test ./internal/ai -v
      2. Assert: exit code is 0
      3. grep -n "homeCurrency" internal/ai/gemini.go
      4. Assert: output shows variable usage in prompt construction
    Expected Result: Tests pass, prompts use variables
    Evidence: Test output and grep results

  Scenario: Prompts fall back to defaults when UserContext is nil
    Tool: Bash
    Preconditions: Code handles nil check
    Steps:
      1. grep -A5 "userCtx == nil" internal/ai/gemini.go
      2. Assert: output shows fallback to default values
    Expected Result: Nil handling exists
    Evidence: Grep output showing fallback logic
  ```

  **Commit**: YES
  - Message: `feat(ai): inject user context into expense parsing prompts`
  - Files: `internal/ai/service.go`, `internal/ai/gemini.go`
  - Pre-commit: `go test ./internal/ai -v`

---

- [x] 6. Wire UserContext into ParseConversationUseCase

  **What to do**:
  - Modify `internal/usecase/parse_conversation.go`
  - Add `userRepo domain.UserRepository` and `categoryRepo domain.CategoryRepository` fields to struct
  - Update constructor `NewParseConversationUseCase()` to accept repos
  - In `Execute()` method, load UserContext before calling AI service
    - Create inline `GetUserContext` logic OR inject `GetUserContextUseCase`
    - Pass loaded context to `aiService.ParseExpense()`
  - Update all callers of the constructor (use `lsp_find_references` first)

  **Must NOT do**:
  - Modify ProcessMessageUseCase beyond constructor call updates
  - Add error handling beyond what's needed (if user not found, log and use nil context)

  **Recommended Agent Profile**:
  - **Category**: `unspecified-low`
    - Reason: Wiring task touching multiple files
  - **Skills**: `[]`
    - Standard Go dependency injection

  **Parallelization**:
  - **Can Run In Parallel**: NO (depends on multiple tasks)
  - **Parallel Group**: Wave 3 (sequential)
  - **Blocks**: Task 7
  - **Blocked By**: Tasks 2, 4, 5

  **References**:
  - `internal/usecase/parse_conversation.go:15-39` - Current struct and constructor to modify
  - `internal/usecase/parse_conversation.go:42-83` - Execute method to update
  - Use `lsp_find_references` on `NewParseConversationUseCase` to find all callers
  - `cmd/server/main.go` or similar - likely caller to update

  **Acceptance Criteria**:

  **Build verification:**
  - [ ] `go build ./...` exits 0

  **Test verification:**
  - [ ] `go test ./internal/usecase -v` passes
  - [ ] `go test ./... -v` all tests pass

  **Agent-Executed QA Scenarios:**

  ```
  Scenario: ParseConversation uses user context
    Tool: Bash
    Preconditions: Tasks 1-5 completed, server can start
    Steps:
      1. go build ./...
      2. Assert: exit code is 0
      3. go test ./internal/usecase -run TestParseConversation -v
      4. Assert: tests pass
    Expected Result: Integration works, tests pass
    Evidence: Build and test outputs

  Scenario: Full integration test with server
    Tool: Bash
    Preconditions: .env configured, server can start
    Steps:
      1. Start server in background: `go run ./cmd/server &`
      2. Wait 3 seconds for startup
      3. curl -s -X POST http://localhost:8080/api/chat/terminal \
           -H "Content-Type: application/json" \
           -d '{"user_id": "test_user", "message": "coffee 50"}'
      4. Assert: response contains parsed expense (status 200 or message field)
      5. Kill server process
    Expected Result: End-to-end flow works
    Evidence: curl response captured
  ```

  **Commit**: YES
  - Message: `feat(usecase): wire user context into parse conversation flow`
  - Files: `internal/usecase/parse_conversation.go`, any caller files
  - Pre-commit: `go test ./... -v`

---

- [x] 7. Final verification and OpenSpec tasks.md update

  **What to do**:
  - Run full test suite: `go test ./... -v`
  - Run OpenSpec validation: `openspec validate add-user-context-aggregate --strict`
  - Update `openspec/changes/add-user-context-aggregate/tasks.md` to mark all items `[x]`
  - Verify no regressions in existing functionality

  **Must NOT do**:
  - Archive the OpenSpec change (that's a separate deployment step)
  - Add new features

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: Verification and checklist update
  - **Skills**: `[]`
    - No special skills needed

  **Parallelization**:
  - **Can Run In Parallel**: NO
  - **Parallel Group**: Wave 4 (final)
  - **Blocks**: None
  - **Blocked By**: Task 6

  **References**:
  - `openspec/changes/add-user-context-aggregate/tasks.md` - File to update

  **Acceptance Criteria**:

  **Test suite passes:**
  - [ ] `go test ./... -v` exits 0

  **OpenSpec valid:**
  - [ ] `openspec validate add-user-context-aggregate --strict` shows "is valid"

  **Tasks updated:**
  - [ ] All items in `tasks.md` marked `[x]`

  **Agent-Executed QA Scenarios:**

  ```
  Scenario: All tests pass
    Tool: Bash
    Preconditions: All previous tasks completed
    Steps:
      1. go test ./... -v
      2. Assert: exit code is 0
      3. Assert: output contains "PASS" for all packages
    Expected Result: Full test suite green
    Evidence: Test output captured

  Scenario: OpenSpec validation passes
    Tool: Bash
    Preconditions: All code changes complete
    Steps:
      1. openspec validate add-user-context-aggregate --strict
      2. Assert: output contains "is valid"
    Expected Result: OpenSpec requirements met
    Evidence: Validation output

  Scenario: tasks.md fully checked
    Tool: Bash
    Preconditions: tasks.md updated
    Steps:
      1. grep -c "\[x\]" openspec/changes/add-user-context-aggregate/tasks.md
      2. Assert: count equals 6 (all tasks checked)
      3. grep -c "\[ \]" openspec/changes/add-user-context-aggregate/tasks.md
      4. Assert: count equals 0 (no unchecked tasks)
    Expected Result: All tasks marked complete
    Evidence: grep output
  ```

  **Commit**: YES
  - Message: `docs(openspec): mark user-context-aggregate tasks complete`
  - Files: `openspec/changes/add-user-context-aggregate/tasks.md`
  - Pre-commit: `openspec validate add-user-context-aggregate --strict`

---

## Commit Strategy

| After Task | Message | Files | Verification |
|------------|---------|-------|--------------|
| 2 | `feat(usecase): add GetUserContext use case for prompt personalization` | `internal/domain/models.go`, `internal/usecase/get_user_context.go`, `internal/usecase/get_user_context_test.go` | `go test ./internal/usecase -v` |
| 5 | `feat(ai): inject user context into expense parsing prompts` | `internal/ai/service.go`, `internal/ai/gemini.go` | `go test ./internal/ai -v` |
| 6 | `feat(usecase): wire user context into parse conversation flow` | `internal/usecase/parse_conversation.go`, caller files | `go test ./... -v` |
| 7 | `docs(openspec): mark user-context-aggregate tasks complete` | `openspec/changes/add-user-context-aggregate/tasks.md` | `openspec validate` |

---

## Success Criteria

### Verification Commands
```bash
# Build succeeds
go build ./...  # Expected: exit 0, no errors

# All tests pass
go test ./... -v  # Expected: PASS for all packages

# UserContext type exists
go doc github.com/riverlin/aiexpense/internal/domain.UserContext
# Expected: Shows struct with User and Categories fields

# OpenSpec is valid
openspec validate add-user-context-aggregate --strict
# Expected: "is valid"
```

### Final Checklist
- [x] `UserContext` struct defined in `internal/domain/models.go`
- [x] `GetUserContextUseCase` implemented and tested
- [x] `ai.Service` interface accepts `*domain.UserContext`
- [x] Prompts use user's HomeCurrency (not hardcoded TWD)
- [x] Prompts use user's category names (not hardcoded list)
- [x] Graceful fallback when UserContext is nil
- [x] All tests pass
- [x] OpenSpec tasks.md fully checked
