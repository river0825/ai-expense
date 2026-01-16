# Integration Testing Guide

This guide describes how to perform integration testing for AIExpense, including HTTP handler tests, repository tests, and end-to-end testing.

## Testing Strategy

AIExpense uses a layered testing approach:

1. **Unit Tests** (Completed - 60+ tests)
   - AI service parsing and categorization
   - Use case business logic
   - Mock-based isolation

2. **Integration Tests** (This guide)
   - HTTP handler request/response flow
   - Repository layer with SQLite
   - Multi-layer interaction

3. **End-to-End Tests** (Manual/Automation)
   - Full message flow (webhook to response)
   - Real database operations
   - Messenger platform simulation

## Unit Tests (Already Implemented)

### Running Unit Tests

```bash
# All unit tests
go test ./...

# Specific package
go test ./internal/ai -v
go test ./internal/usecase -v

# Specific test
go test ./internal/ai -run TestParseExpenseRegex -v
```

### Current Coverage

- **internal/ai**: 24 test functions covering parsing and categorization
- **internal/usecase**: 17 test functions covering business logic
- **Total**: 60+ test cases with 100% passing rate

### Mock Pattern Used

All unit tests use in-memory mock repositories implementing domain interfaces:

```go
type MockUserRepository struct {
    users map[string]*domain.User
}

func (m *MockUserRepository) Create(ctx context.Context, user *domain.User) error {
    m.users[user.UserID] = user
    return nil
}
```

## Integration Testing

### HTTP Handler Integration Tests

Test handlers with real repositories and SQLite in-memory database:

```bash
# Create file: internal/adapter/http/handler_integration_test.go
go test ./internal/adapter/http -run TestAutoSignupFlow -v
```

#### Test Structure

```go
func TestAutoSignupFlow(t *testing.T) {
    // 1. Set up real database
    db := createTestDB(t)
    defer db.Close()

    // 2. Create real repositories
    userRepo := sqliteAdapter.NewUserRepository(db)
    categoryRepo := sqliteAdapter.NewCategoryRepository(db)

    // 3. Initialize use cases with real repos
    autoSignupUC := usecase.NewAutoSignupUseCase(userRepo, categoryRepo)

    // 4. Create handler
    handler := NewHandler(autoSignupUC, ...)

    // 5. Make HTTP request
    req := httptest.NewRequest("POST", "/api/users/auto-signup", ...)
    w := httptest.NewRecorder()

    // 6. Verify response
    handler.AutoSignup(w, req)
    assert(w.Code == http.StatusOK)
}
```

### Repository Integration Tests

Test SQLite repositories with real database:

```bash
# Create file: internal/adapter/repository/sqlite/repositories_integration_test.go
go test ./internal/adapter/repository/sqlite -run TestUserRepository -v
```

#### Test Pattern

```go
func TestUserRepositoryCreate(t *testing.T) {
    db := createTestDB(t)  // :memory: SQLite
    repo := NewUserRepository(db)

    // Test operations
    user := &domain.User{UserID: "test_1", MessengerType: "line"}
    err := repo.Create(context.Background(), user)

    // Verify
    retrieved, _ := repo.GetByID(context.Background(), "test_1")
    assert(retrieved.UserID == "test_1")
}
```

### Full Integration Flow Test

Test end-to-end message processing:

```go
func TestExpenseCreationFlow(t *testing.T) {
    // Create in-memory database with real repos
    db := createTestDB(t)

    // Initialize all repos
    userRepo := sqliteAdapter.NewUserRepository(db)
    categoryRepo := sqliteAdapter.NewCategoryRepository(db)
    expenseRepo := sqliteAdapter.NewExpenseRepository(db)

    // Create mock AI service
    aiService := &mockAI{}

    // Initialize use cases
    autoSignupUC := usecase.NewAutoSignupUseCase(userRepo, categoryRepo)
    createExpenseUC := usecase.NewCreateExpenseUseCase(expenseRepo, categoryRepo, aiService)

    // Create handler
    handler := NewHandler(autoSignupUC, createExpenseUC, ...)

    // Step 1: Auto-signup
    signupReq := httptest.NewRequest("POST", "/api/users/auto-signup", ...)
    signupW := httptest.NewRecorder()
    handler.AutoSignup(signupW, signupReq)
    assert(signupW.Code == 200)

    // Step 2: Create expense
    expenseReq := httptest.NewRequest("POST", "/api/expenses", ...)
    expenseW := httptest.NewRecorder()
    handler.CreateExpense(expenseW, expenseReq)
    assert(expenseW.Code == 201)

    // Step 3: Verify in database
    expenses, _ := expenseRepo.GetByUserID(context.Background(), "test_user")
    assert(len(expenses) == 1)
    assert(expenses[0].Description == "breakfast")
}
```

## Running Integration Tests

### Manual Testing with curl

1. **Start the server**:
```bash
export LINE_CHANNEL_TOKEN=test
export LINE_CHANNEL_ID=test
export GEMINI_API_KEY=test
go run ./cmd/server
```

2. **Test endpoints**:
```bash
# Auto-signup
curl -X POST http://localhost:8080/api/users/auto-signup \
  -H "Content-Type: application/json" \
  -d '{"user_id":"test_user_1","messenger_type":"line"}'

# Create expense
curl -X POST http://localhost:8080/api/expenses \
  -H "Content-Type: application/json" \
  -d '{
    "user_id":"test_user_1",
    "description":"breakfast",
    "amount":20.0
  }'

# Get expenses
curl http://localhost:8080/api/expenses?user_id=test_user_1

# Get metrics (with auth)
curl http://localhost:8080/api/metrics/dau \
  -H "X-API-Key: admin_key"
```

### Docker Integration Testing

```bash
# Start with docker-compose
docker-compose up -d

# Wait for container to be ready
sleep 5

# Run tests against running service
curl http://localhost:8080/health

# Clean up
docker-compose down
```

## Test Coverage Goals

### Current Status
- Unit Tests: ✅ 60+ tests passing
- Integration Tests: ⏳ To be implemented
- E2E Tests: ⏳ To be implemented

### Target Coverage

| Layer | Type | Goal |
|-------|------|------|
| AI Service | Unit | 100% ✅ Done |
| Use Cases | Unit | 100% ✅ Done |
| Repositories | Integration | 100% |
| HTTP Handlers | Integration | 100% |
| Messengers | Integration | 100% |
| Full System | E2E | 100% |

## Writing Integration Tests

### Template

```go
package handler

import (
    "context"
    "testing"

    "github.com/riverlin/aiexpense/internal/adapter/repository/sqlite"
    "github.com/riverlin/aiexpense/internal/usecase"
)

func TestMyIntegration(t *testing.T) {
    // 1. Create in-memory database
    db := createTestDB(t)
    defer db.Close()

    // 2. Initialize repositories
    userRepo := sqlite.NewUserRepository(db)
    expenseRepo := sqlite.NewExpenseRepository(db)

    // 3. Create use cases
    myUseCase := usecase.NewMyUseCase(userRepo, expenseRepo)

    // 4. Execute operation
    result, err := myUseCase.Execute(context.Background(), req)

    // 5. Verify result
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if result == nil {
        t.Error("expected result, got nil")
    }
}

func createTestDB(t *testing.T) *sql.DB {
    db, err := sqlite.OpenDB(":memory:")
    if err != nil {
        t.Fatalf("failed to create test db: %v", err)
    }
    return db
}
```

## CI/CD Integration

### GitHub Actions Example

```yaml
name: Integration Tests

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - uses: actions/setup-go@v2
      - run: go test -v ./...
      - run: go test -v -race ./...
      - run: go test -cover ./...
```

## Performance Testing

For integration tests, track performance metrics:

```go
func TestPerformance(t *testing.T) {
    db := createTestDB(t)
    repo := sqlite.NewExpenseRepository(db)

    // Create 1000 expenses
    start := time.Now()
    for i := 0; i < 1000; i++ {
        repo.Create(context.Background(), &domain.Expense{...})
    }
    elapsed := time.Since(start)

    t.Logf("Created 1000 expenses in %v", elapsed)
    if elapsed > 10*time.Second {
        t.Error("performance regression: too slow")
    }
}
```

## Debugging Failed Tests

### Enable verbose output
```bash
go test -v -run TestMyTest
```

### Run with race detector
```bash
go test -race ./internal/...
```

### Get more details
```bash
go test -v -run TestMyTest -timeout 30s -cpuprofile=cpu.prof -memprofile=mem.prof
go tool pprof cpu.prof
```

## Next Steps

1. **Create handler integration tests** - Test HTTP request/response flow
2. **Create repository tests** - Test SQLite operations with real queries
3. **Add E2E tests** - Test full message flow from webhook to response
4. **Performance tests** - Benchmark critical paths
5. **Load tests** - Test concurrent user handling

## References

- Go testing documentation: https://golang.org/pkg/testing/
- Table-driven tests: https://golang.org/doc/effective_go#table_driven_tests
- httptest package: https://golang.org/pkg/net/http/httptest/
- SQLite testing: Use :memory: database for isolation
