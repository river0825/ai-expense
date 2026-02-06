# Fix Category Response Format

## TL;DR

> **Quick Summary**: The backend returns `ListCategoriesResponse` (an object `{categories: [], total: N}`) without JSON tags (so capitalized keys), but the frontend expects a simple array `Category[]` with snake_case keys (`user_id`, `is_default`).
>
> **Deliverables**:
> - Updated `internal/usecase/manage_category.go` with JSON tags for `CategoryResponse` and `ListCategoriesResponse`.
> - Updated `internal/adapter/http/handler.go` to return `resp.Categories` directly for `ListCategories`.
>
> **Estimated Effort**: Quick
> **Parallel Execution**: NO
> **Critical Path**: Task 1 -> Task 2

---

## Context

### The Bug
1. `handler.go` calls `ListCategories` which returns `*ListCategoriesResponse`.
2. `handler.go` writes this struct as JSON. Since no JSON tags exist, fields are capitalized: `{ "Categories": [...], "Total": ... }`.
3. Frontend `HttpCategoryRepository` expects `response.data.data` to be an array `Category[]`.
4. Instead it gets the object.
5. `page.tsx` calls `.map()` on this object -> CRASH.
6. Also, `CategoryResponse` fields are capitalized (`ID`, `UserID`, `IsDefault`), but frontend interface uses `id`, `user_id`, `is_default`.

### The Fix
1. Add JSON tags to structs to ensure correct casing (snake_case for frontend compatibility).
2. Unwrap the response in `handler.go` to return just the array, matching frontend expectation.

---

## Verification Strategy

### Automated Tests
- Build verification.
- Frontend should load settings page without crash.

### Agent-Executed QA Scenarios (MANDATORY)

```
Scenario: List categories returns correct JSON structure
  Tool: Bash (curl)
  Preconditions: Server running
  Steps:
    1. curl -s "http://localhost:8080/api/user/categories?token=test-user"
    2. Assert: response.data is an array (starts with [)
    3. Assert: array items have keys "id", "user_id", "is_default" (not capitalized)
  Expected Result: Success
  Evidence: Response body captured
```

---

## TODOs

- [ ] 1. Add JSON tags to usecase structs
  **What to do**:
  - Modify `internal/usecase/manage_category.go`
  - Update `CategoryResponse` struct with `json:"..."` tags (snake_case).
  - Update `ListCategoriesResponse` struct with `json:"..."` tags.

- [ ] 2. Update Handler to return array
  **What to do**:
  - Modify `internal/adapter/http/handler.go`
  - In `ListCategories` method, change:
    ```go
    h.WriteJSON(w, http.StatusOK, &Response{Status: "success", Data: resp})
    ```
    to
    ```go
    h.WriteJSON(w, http.StatusOK, &Response{Status: "success", Data: resp.Categories})
    ```

- [ ] 3. Verify Fix
  **What to do**:
  - Rebuild server: `go build -o server ./cmd/server`
  - Start server
  - Run curl check

---

## Success Criteria
- Frontend Settings page loads categories correctly.
- JSON response keys match frontend interface.
