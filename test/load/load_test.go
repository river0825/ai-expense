package load

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/riverlin/aiexpense/internal/domain"
	"github.com/riverlin/aiexpense/internal/usecase"
)

// LoadTestExpenseRepository implements in-memory expense repository for load testing
type LoadTestExpenseRepository struct {
	expenses map[string]*domain.Expense
	mu       sync.RWMutex
}

func (r *LoadTestExpenseRepository) Create(ctx context.Context, expense *domain.Expense) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.expenses[expense.ID] = expense
	return nil
}

func (r *LoadTestExpenseRepository) GetByID(ctx context.Context, id string) (*domain.Expense, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if exp, ok := r.expenses[id]; ok {
		return exp, nil
	}
	return nil, nil
}

func (r *LoadTestExpenseRepository) GetByUserID(ctx context.Context, userID string) ([]*domain.Expense, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []*domain.Expense
	for _, exp := range r.expenses {
		if exp.UserID == userID {
			result = append(result, exp)
		}
	}
	return result, nil
}

func (r *LoadTestExpenseRepository) GetByUserIDAndDateRange(ctx context.Context, userID string, from, to time.Time) ([]*domain.Expense, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []*domain.Expense
	for _, exp := range r.expenses {
		if exp.UserID == userID && !exp.ExpenseDate.Before(from) && !exp.ExpenseDate.After(to) {
			result = append(result, exp)
		}
	}
	return result, nil
}

func (r *LoadTestExpenseRepository) GetByUserIDAndCategory(ctx context.Context, userID, categoryID string) ([]*domain.Expense, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []*domain.Expense
	for _, exp := range r.expenses {
		if exp.UserID == userID && exp.CategoryID != nil && *exp.CategoryID == categoryID {
			result = append(result, exp)
		}
	}
	return result, nil
}

func (r *LoadTestExpenseRepository) Update(ctx context.Context, expense *domain.Expense) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.expenses[expense.ID] = expense
	return nil
}

func (r *LoadTestExpenseRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.expenses, id)
	return nil
}

// LoadTestUserRepository implements in-memory user repository for load testing
type LoadTestUserRepository struct {
	users map[string]*domain.User
	mu    sync.RWMutex
}

func (r *LoadTestUserRepository) Create(ctx context.Context, user *domain.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.users[user.UserID] = user
	return nil
}

func (r *LoadTestUserRepository) GetByID(ctx context.Context, userID string) (*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if user, ok := r.users[userID]; ok {
		return user, nil
	}
	return nil, nil
}

func (r *LoadTestUserRepository) Exists(ctx context.Context, userID string) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, ok := r.users[userID]
	return ok, nil
}

// LoadTestCategoryRepository implements in-memory category repository for load testing
type LoadTestCategoryRepository struct {
	categories map[string]*domain.Category
	mu         sync.RWMutex
}

func (r *LoadTestCategoryRepository) Create(ctx context.Context, category *domain.Category) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.categories[category.ID] = category
	return nil
}

func (r *LoadTestCategoryRepository) GetByID(ctx context.Context, id string) (*domain.Category, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if cat, ok := r.categories[id]; ok {
		return cat, nil
	}
	return nil, nil
}

func (r *LoadTestCategoryRepository) GetByUserID(ctx context.Context, userID string) ([]*domain.Category, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []*domain.Category
	for _, cat := range r.categories {
		if cat.UserID == userID {
			result = append(result, cat)
		}
	}
	return result, nil
}

func (r *LoadTestCategoryRepository) GetByUserIDAndName(ctx context.Context, userID, name string) (*domain.Category, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, cat := range r.categories {
		if cat.UserID == userID && cat.Name == name {
			return cat, nil
		}
	}
	return nil, nil
}

func (r *LoadTestCategoryRepository) Update(ctx context.Context, category *domain.Category) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.categories[category.ID] = category
	return nil
}

func (r *LoadTestCategoryRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.categories, id)
	return nil
}

func (r *LoadTestCategoryRepository) CreateKeyword(ctx context.Context, keyword *domain.CategoryKeyword) error {
	return nil
}

func (r *LoadTestCategoryRepository) GetKeywordsByCategory(ctx context.Context, categoryID string) ([]*domain.CategoryKeyword, error) {
	return []*domain.CategoryKeyword{}, nil
}

func (r *LoadTestCategoryRepository) DeleteKeyword(ctx context.Context, id string) error {
	return nil
}

// LoadTestAIService implements minimal AI service for load testing
type LoadTestAIService struct{}

func (s *LoadTestAIService) ParseExpense(ctx context.Context, text string, userID string) ([]*domain.ParsedExpense, error) {
	return []*domain.ParsedExpense{
		{Amount: 20.0, Description: "Test"},
	}, nil
}

func (s *LoadTestAIService) SuggestCategory(ctx context.Context, description string, userID string) (string, error) {
	return "food", nil
}

// LoadTestMetrics tracks performance metrics during load tests
type LoadTestMetrics struct {
	totalRequests   int64
	successRequests int64
	failedRequests  int64
	totalDuration   int64 // nanoseconds
	minDuration     int64 // nanoseconds
	maxDuration     int64 // nanoseconds
	mu              sync.RWMutex
}

func (m *LoadTestMetrics) recordRequest(duration time.Duration, success bool) {
	durNs := duration.Nanoseconds()

	atomic.AddInt64(&m.totalRequests, 1)
	if success {
		atomic.AddInt64(&m.successRequests, 1)
	} else {
		atomic.AddInt64(&m.failedRequests, 1)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.totalDuration += durNs
	if m.minDuration == 0 || durNs < m.minDuration {
		m.minDuration = durNs
	}
	if durNs > m.maxDuration {
		m.maxDuration = durNs
	}
}

func (m *LoadTestMetrics) getStats() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	total := atomic.LoadInt64(&m.totalRequests)
	var avgDuration int64
	if total > 0 {
		avgDuration = m.totalDuration / total
	}

	return map[string]interface{}{
		"total_requests":   total,
		"success_requests": atomic.LoadInt64(&m.successRequests),
		"failed_requests":  atomic.LoadInt64(&m.failedRequests),
		"avg_duration_ns":  avgDuration,
		"avg_duration_ms":  float64(avgDuration) / 1e6,
		"min_duration_ms":  float64(m.minDuration) / 1e6,
		"max_duration_ms":  float64(m.maxDuration) / 1e6,
		"success_rate": func() float64 {
			if total == 0 {
				return 0
			}
			return float64(atomic.LoadInt64(&m.successRequests)) / float64(total) * 100
		}(),
	}
}

// LoadTestConcurrentSignups tests concurrent user registration
func TestLoadConcurrentSignups(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping load test in short mode")
	}

	userRepo := &LoadTestUserRepository{users: make(map[string]*domain.User)}
	categoryRepo := &LoadTestCategoryRepository{categories: make(map[string]*domain.Category)}

	uc := usecase.NewAutoSignupUseCase(userRepo, categoryRepo)
	ctx := context.Background()
	metrics := &LoadTestMetrics{}

	concurrency := 50
	requestsPerGoroutine := 20
	wg := sync.WaitGroup{}

	start := time.Now()

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()
			for j := 0; j < requestsPerGoroutine; j++ {
				userID := fmt.Sprintf("load_user_%d_%d", goroutineID, j)
				opStart := time.Now()
				err := uc.Execute(ctx, userID, "telegram")
				duration := time.Since(opStart)
				metrics.recordRequest(duration, err == nil)
			}
		}(i)
	}

	wg.Wait()
	totalDuration := time.Since(start)

	stats := metrics.getStats()
	t.Logf("Concurrent Signups Load Test")
	t.Logf("  Total Requests: %d", stats["total_requests"])
	t.Logf("  Success Rate: %.2f%%", stats["success_rate"])
	t.Logf("  Avg Duration: %.2f ms", stats["avg_duration_ms"])
	t.Logf("  Min Duration: %.2f ms", stats["min_duration_ms"])
	t.Logf("  Max Duration: %.2f ms", stats["max_duration_ms"])
	t.Logf("  Total Duration: %v", totalDuration)

	// Verify all operations succeeded
	if success := stats["success_requests"].(int64); success != int64(concurrency*requestsPerGoroutine) {
		t.Errorf("Expected %d successes, got %d", concurrency*requestsPerGoroutine, success)
	}

	// Verify average duration is reasonable
	avgMs := stats["avg_duration_ms"].(float64)
	if avgMs > 20 {
		t.Logf("Warning: Average signup duration %.2f ms is higher than expected (target: < 10ms)", avgMs)
	}
}

// LoadTestConcurrentExpenseCreation tests concurrent expense creation
func TestLoadConcurrentExpenseCreation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping load test in short mode")
	}

	expenseRepo := &LoadTestExpenseRepository{expenses: make(map[string]*domain.Expense)}
	userRepo := &LoadTestUserRepository{users: make(map[string]*domain.User)}
	categoryRepo := &LoadTestCategoryRepository{categories: make(map[string]*domain.Category)}

	// Setup initial user and category
	userRepo.Create(context.Background(), &domain.User{
		UserID:        "load_test_user",
		MessengerType: "telegram",
		CreatedAt:     time.Now(),
	})

	categoryRepo.Create(context.Background(), &domain.Category{
		ID:     "cat_food",
		UserID: "load_test_user",
		Name:   "Food",
	})

	aiService := &LoadTestAIService{}
	uc := usecase.NewCreateExpenseUseCase(expenseRepo, categoryRepo, aiService)
	ctx := context.Background()
	metrics := &LoadTestMetrics{}

	concurrency := 30
	requestsPerGoroutine := 15
	wg := sync.WaitGroup{}
	var expenseCounter int64

	start := time.Now()

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()
			for j := 0; j < requestsPerGoroutine; j++ {
				opStart := time.Now()
				_, err := uc.Execute(ctx, &usecase.CreateRequest{
					UserID:      "load_test_user",
					Description: fmt.Sprintf("Expense_%d_%d", goroutineID, j),
					Amount:      float64(20 + j),
				})
				duration := time.Since(opStart)
				metrics.recordRequest(duration, err == nil)
				if err == nil {
					atomic.AddInt64(&expenseCounter, 1)
				}
			}
		}(i)
	}

	wg.Wait()
	totalDuration := time.Since(start)

	stats := metrics.getStats()
	t.Logf("Concurrent Expense Creation Load Test")
	t.Logf("  Total Requests: %d", stats["total_requests"])
	t.Logf("  Success Rate: %.2f%%", stats["success_rate"])
	t.Logf("  Avg Duration: %.2f ms", stats["avg_duration_ms"])
	t.Logf("  Min Duration: %.2f ms", stats["min_duration_ms"])
	t.Logf("  Max Duration: %.2f ms", stats["max_duration_ms"])
	t.Logf("  Total Duration: %v", totalDuration)
	t.Logf("  Expenses Created: %d", atomic.LoadInt64(&expenseCounter))

	// Verify data integrity
	expenses, _ := expenseRepo.GetByUserID(ctx, "load_test_user")
	if len(expenses) != int(atomic.LoadInt64(&expenseCounter)) {
		t.Errorf("Data integrity check failed: expected %d expenses in repo, got %d", atomic.LoadInt64(&expenseCounter), len(expenses))
	}

	// Verify average duration is reasonable
	avgMs := stats["avg_duration_ms"].(float64)
	if avgMs > 100 {
		t.Logf("Warning: Average expense creation duration %.2f ms is higher than expected (target: < 50ms)", avgMs)
	}
}

// LoadTestConcurrentRetrieval tests concurrent expense retrieval
func TestLoadConcurrentRetrieval(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping load test in short mode")
	}

	expenseRepo := &LoadTestExpenseRepository{expenses: make(map[string]*domain.Expense)}
	categoryRepo := &LoadTestCategoryRepository{categories: make(map[string]*domain.Category)}

	// Populate with test data
	for i := 0; i < 500; i++ {
		expenseRepo.Create(context.Background(), &domain.Expense{
			ID:          fmt.Sprintf("exp_%d", i),
			UserID:      "load_test_user",
			Description: "Test",
			Amount:      20.0,
			ExpenseDate: time.Now(),
			CreatedAt:   time.Now(),
		})
	}

	uc := usecase.NewGetExpensesUseCase(expenseRepo, categoryRepo)
	ctx := context.Background()
	metrics := &LoadTestMetrics{}

	concurrency := 40
	requestsPerGoroutine := 25
	wg := sync.WaitGroup{}

	start := time.Now()

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < requestsPerGoroutine; j++ {
				opStart := time.Now()
				_, err := uc.ExecuteGetAll(ctx, &usecase.GetAllRequest{UserID: "load_test_user"})
				duration := time.Since(opStart)
				metrics.recordRequest(duration, err == nil)
			}
		}()
	}

	wg.Wait()
	totalDuration := time.Since(start)

	stats := metrics.getStats()
	t.Logf("Concurrent Retrieval Load Test")
	t.Logf("  Total Requests: %d", stats["total_requests"])
	t.Logf("  Success Rate: %.2f%%", stats["success_rate"])
	t.Logf("  Avg Duration: %.2f ms", stats["avg_duration_ms"])
	t.Logf("  Min Duration: %.2f ms", stats["min_duration_ms"])
	t.Logf("  Max Duration: %.2f ms", stats["max_duration_ms"])
	t.Logf("  Total Duration: %v", totalDuration)

	// Verify all operations succeeded
	if success := stats["success_requests"].(int64); success != int64(concurrency*requestsPerGoroutine) {
		t.Errorf("Expected %d successes, got %d", concurrency*requestsPerGoroutine, success)
	}

	// Verify average duration is reasonable
	avgMs := stats["avg_duration_ms"].(float64)
	if avgMs > 50 {
		t.Logf("Warning: Average retrieval duration %.2f ms is higher than expected (target: < 20ms)", avgMs)
	}
}

// LoadTestConcurrentMixedOperations tests mixed concurrent operations (CRUD)
func TestLoadConcurrentMixedOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping load test in short mode")
	}

	expenseRepo := &LoadTestExpenseRepository{expenses: make(map[string]*domain.Expense)}
	userRepo := &LoadTestUserRepository{users: make(map[string]*domain.User)}
	categoryRepo := &LoadTestCategoryRepository{categories: make(map[string]*domain.Category)}

	// Setup users
	for i := 0; i < 5; i++ {
		userID := fmt.Sprintf("mixed_user_%d", i)
		userRepo.Create(context.Background(), &domain.User{
			UserID:        userID,
			MessengerType: "telegram",
			CreatedAt:     time.Now(),
		})
		categoryRepo.Create(context.Background(), &domain.Category{
			ID:     fmt.Sprintf("cat_%d", i),
			UserID: userID,
			Name:   "Food",
		})
	}

	aiService := &LoadTestAIService{}
	createUC := usecase.NewCreateExpenseUseCase(expenseRepo, categoryRepo, aiService)
	getUC := usecase.NewGetExpensesUseCase(expenseRepo, categoryRepo)
	ctx := context.Background()
	metrics := &LoadTestMetrics{}

	concurrency := 30
	operationsPerGoroutine := 40
	wg := sync.WaitGroup{}
	var expenseCounter int64

	start := time.Now()

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()
			userID := fmt.Sprintf("mixed_user_%d", goroutineID%5)

			for j := 0; j < operationsPerGoroutine; j++ {
				// 60% create, 40% read
				if j%10 < 6 {
					// Create operation
					opStart := time.Now()
					_, err := createUC.Execute(ctx, &usecase.CreateRequest{
						UserID:      userID,
						Description: fmt.Sprintf("Mixed_%d_%d", goroutineID, j),
						Amount:      float64(20 + j%30),
					})
					duration := time.Since(opStart)
					metrics.recordRequest(duration, err == nil)
					if err == nil {
						atomic.AddInt64(&expenseCounter, 1)
					}
				} else {
					// Read operation
					opStart := time.Now()
					_, err := getUC.ExecuteGetAll(ctx, &usecase.GetAllRequest{UserID: userID})
					duration := time.Since(opStart)
					metrics.recordRequest(duration, err == nil)
				}
			}
		}(i)
	}

	wg.Wait()
	totalDuration := time.Since(start)

	stats := metrics.getStats()
	t.Logf("Concurrent Mixed Operations Load Test")
	t.Logf("  Total Requests: %d", stats["total_requests"])
	t.Logf("  Success Rate: %.2f%%", stats["success_rate"])
	t.Logf("  Avg Duration: %.2f ms", stats["avg_duration_ms"])
	t.Logf("  Min Duration: %.2f ms", stats["min_duration_ms"])
	t.Logf("  Max Duration: %.2f ms", stats["max_duration_ms"])
	t.Logf("  Total Duration: %v", totalDuration)
	t.Logf("  Expenses Created: %d", atomic.LoadInt64(&expenseCounter))

	// Verify success rate is high
	successRate := stats["success_rate"].(float64)
	if successRate < 99.0 {
		t.Errorf("Success rate %.2f%% is below expected threshold (99%%)", successRate)
	}
}

// LoadTestConcurrentStress tests high-concurrency stress scenario
func TestLoadConcurrentStress(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping load test in short mode")
	}

	expenseRepo := &LoadTestExpenseRepository{expenses: make(map[string]*domain.Expense)}
	userRepo := &LoadTestUserRepository{users: make(map[string]*domain.User)}
	categoryRepo := &LoadTestCategoryRepository{categories: make(map[string]*domain.Category)}

	// Setup initial data
	for i := 0; i < 10; i++ {
		userID := fmt.Sprintf("stress_user_%d", i)
		userRepo.Create(context.Background(), &domain.User{
			UserID:        userID,
			MessengerType: "telegram",
			CreatedAt:     time.Now(),
		})
		categoryRepo.Create(context.Background(), &domain.Category{
			ID:     fmt.Sprintf("stress_cat_%d", i),
			UserID: userID,
			Name:   "Food",
		})
	}

	aiService := &LoadTestAIService{}
	uc := usecase.NewCreateExpenseUseCase(expenseRepo, categoryRepo, aiService)
	ctx := context.Background()
	metrics := &LoadTestMetrics{}

	// High concurrency scenario
	concurrency := 100
	requestsPerGoroutine := 10
	wg := sync.WaitGroup{}

	start := time.Now()

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()
			userID := fmt.Sprintf("stress_user_%d", goroutineID%10)

			for j := 0; j < requestsPerGoroutine; j++ {
				opStart := time.Now()
				_, err := uc.Execute(ctx, &usecase.CreateRequest{
					UserID:      userID,
					Description: fmt.Sprintf("Stress_%d_%d", goroutineID, j),
					Amount:      float64(20),
				})
				duration := time.Since(opStart)
				metrics.recordRequest(duration, err == nil)
			}
		}(i)
	}

	wg.Wait()
	totalDuration := time.Since(start)

	stats := metrics.getStats()
	t.Logf("High Concurrency Stress Load Test")
	t.Logf("  Total Requests: %d", stats["total_requests"])
	t.Logf("  Success Rate: %.2f%%", stats["success_rate"])
	t.Logf("  Avg Duration: %.2f ms", stats["avg_duration_ms"])
	t.Logf("  Min Duration: %.2f ms", stats["min_duration_ms"])
	t.Logf("  Max Duration: %.2f ms", stats["max_duration_ms"])
	t.Logf("  Total Duration: %v", totalDuration)

	// Verify operations complete under stress
	successRate := stats["success_rate"].(float64)
	if successRate < 98.0 {
		t.Errorf("Success rate %.2f%% is below stress test threshold (98%%)", successRate)
	}
}

// LoadTestRampUp tests gradual load increase (ramp-up pattern)
func TestLoadRampUp(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping load test in short mode")
	}

	expenseRepo := &LoadTestExpenseRepository{expenses: make(map[string]*domain.Expense)}
	userRepo := &LoadTestUserRepository{users: make(map[string]*domain.User)}
	categoryRepo := &LoadTestCategoryRepository{categories: make(map[string]*domain.Category)}

	userRepo.Create(context.Background(), &domain.User{
		UserID:        "rampup_user",
		MessengerType: "telegram",
		CreatedAt:     time.Now(),
	})

	categoryRepo.Create(context.Background(), &domain.Category{
		ID:     "rampup_cat",
		UserID: "rampup_user",
		Name:   "Food",
	})

	aiService := &LoadTestAIService{}
	uc := usecase.NewCreateExpenseUseCase(expenseRepo, categoryRepo, aiService)
	ctx := context.Background()

	// Ramp-up stages
	stages := []struct {
		concurrency    int
		requestsPerGo  int
		durationTarget time.Duration // Expected max duration for stage
	}{
		{5, 10, 50 * time.Millisecond},
		{10, 10, 75 * time.Millisecond},
		{20, 10, 100 * time.Millisecond},
		{50, 10, 150 * time.Millisecond},
	}

	for stageIdx, stage := range stages {
		metrics := &LoadTestMetrics{}
		wg := sync.WaitGroup{}

		stageStart := time.Now()

		for i := 0; i < stage.concurrency; i++ {
			wg.Add(1)
			go func(goroutineID int) {
				defer wg.Done()
				for j := 0; j < stage.requestsPerGo; j++ {
					opStart := time.Now()
					_, err := uc.Execute(ctx, &usecase.CreateRequest{
						UserID:      "rampup_user",
						Description: fmt.Sprintf("Rampup_s%d_%d_%d", stageIdx, goroutineID, j),
						Amount:      20.0,
					})
					duration := time.Since(opStart)
					metrics.recordRequest(duration, err == nil)
				}
			}(i)
		}

		wg.Wait()
		stageDuration := time.Since(stageStart)

		stats := metrics.getStats()
		t.Logf("Ramp-Up Stage %d (concurrency=%d, total=%d requests)", stageIdx+1, stage.concurrency, stage.concurrency*stage.requestsPerGo)
		t.Logf("  Success Rate: %.2f%%", stats["success_rate"])
		t.Logf("  Avg Duration: %.2f ms", stats["avg_duration_ms"])
		t.Logf("  Max Duration: %.2f ms", stats["max_duration_ms"])
		t.Logf("  Stage Duration: %v", stageDuration)

		// Verify stage completes within expected time
		if stageDuration > stage.durationTarget*2 {
			t.Logf("Warning: Stage %d took %.2f seconds, expected < %.2f seconds", stageIdx+1, stageDuration.Seconds(), (stage.durationTarget * 2).Seconds())
		}
	}
}

// LoadTestSustainedLoad tests sustained load over time
func TestLoadSustainedLoad(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping load test in short mode")
	}

	expenseRepo := &LoadTestExpenseRepository{expenses: make(map[string]*domain.Expense)}
	userRepo := &LoadTestUserRepository{users: make(map[string]*domain.User)}
	categoryRepo := &LoadTestCategoryRepository{categories: make(map[string]*domain.Category)}

	// Setup multiple users
	for i := 0; i < 5; i++ {
		userID := fmt.Sprintf("sustained_user_%d", i)
		userRepo.Create(context.Background(), &domain.User{
			UserID:        userID,
			MessengerType: "telegram",
			CreatedAt:     time.Now(),
		})
		categoryRepo.Create(context.Background(), &domain.Category{
			ID:     fmt.Sprintf("sustained_cat_%d", i),
			UserID: userID,
			Name:   "Food",
		})
	}

	aiService := &LoadTestAIService{}
	uc := usecase.NewCreateExpenseUseCase(expenseRepo, categoryRepo, aiService)
	ctx := context.Background()

	// Sustained load for 5 seconds at constant rate
	concurrency := 20
	duration := 5 * time.Second
	metrics := &LoadTestMetrics{}
	wg := sync.WaitGroup{}
	stopCh := make(chan struct{})

	testStart := time.Now()

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()
			userID := fmt.Sprintf("sustained_user_%d", goroutineID%5)
			counter := 0

			for {
				select {
				case <-stopCh:
					return
				default:
					opStart := time.Now()
					_, err := uc.Execute(ctx, &usecase.CreateRequest{
						UserID:      userID,
						Description: fmt.Sprintf("Sustained_%d_%d", goroutineID, counter),
						Amount:      20.0,
					})
					duration := time.Since(opStart)
					metrics.recordRequest(duration, err == nil)
					counter++
				}
			}
		}(i)
	}

	// Run for specified duration
	time.Sleep(duration)
	close(stopCh)
	wg.Wait()
	totalDuration := time.Since(testStart)

	stats := metrics.getStats()
	t.Logf("Sustained Load Test (Duration: %v)", duration)
	t.Logf("  Total Requests: %d", stats["total_requests"])
	t.Logf("  Throughput: %.2f req/sec", float64(stats["total_requests"].(int64))/totalDuration.Seconds())
	t.Logf("  Success Rate: %.2f%%", stats["success_rate"])
	t.Logf("  Avg Duration: %.2f ms", stats["avg_duration_ms"])
	t.Logf("  P95 Duration: %.2f ms", stats["max_duration_ms"]) // Approximation
	t.Logf("  Total Duration: %v", totalDuration)

	// Verify sustained performance
	successRate := stats["success_rate"].(float64)
	if successRate < 99.0 {
		t.Errorf("Success rate %.2f%% is below sustained load threshold (99%%)", successRate)
	}
}
