# Plan: Integration Tests for User Aggregate API

## TL;DR

> **Quick Summary**: Create integration tests for the new `GET` and `PUT` `/api/user/aggregate` endpoints to verify end-to-end functionality including database interactions and JSON responses.
>
> **Deliverables**:
> - New integration test file `test/integration/user_aggregate_test.go`
> - Tests for `GET /api/user/aggregate`
> - Tests for `PUT /api/user/aggregate` (renaming, adding accounts)
>
> **Estimated Effort**: Medium
> **Parallel Execution**: NO
> **Critical Path**: Task 1 (Create File) -> Task 2 (GET Test) -> Task 3 (PUT Test)

---

## Context

### Goal
Verify the newly created "User Aggregate Settings API" works as expected.

### Scope
- **IN**: Integration tests for `/api/user/aggregate`.
- **OUT**: Unit tests (already covered or out of scope for this specific request).

---

## Work Objectives

### Core Objective
Ensure the aggregate API correctly handles data retrieval and updates, including the complex account renaming logic.

### Concrete Deliverables
- `test/integration/user_aggregate_test.go`

### Definition of Done
- `go test ./test/integration/...` passes.

---

## Verification Strategy

### Automated Tests
- Run `go test -v ./test/integration/`

### Agent-Executed QA Scenarios
```
Scenario: Run Integration Tests
  Tool: Bash
  Steps:
    1. go test -v ./test/integration/ -run TestUserAggregateAPI
  Expected Result: PASS
```

---

## TODOs

- [x] 1. Create Integration Test File
  **What to do**:
  - Create `test/integration/user_aggregate_test.go`.
  - Setup test suite structure (setup/teardown).

  **Recommended Agent Profile**:
  - **Category**: `quick`

- [x] 2. Implement GET Test
  **What to do**:
  - Add test case `TestGetUserAggregate`.
  - Seed data (User, Categories, Accounts).
  - Call `GET /api/user/aggregate`.
  - Assert response matches seeded data.

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`

- [x] 3. Implement PUT Test
  **What to do**:
  - Add test case `TestUpdateUserAggregate`.
  - Test scenario: Rename account "Cash" -> "Gold".
  - Verify `accounts` table updated.
  - Verify `expenses` table updated (historical data sync).

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`

---

## Success Criteria
- All integration tests pass.
