package http

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/riverlin/aiexpense/internal/ai"
	"github.com/riverlin/aiexpense/internal/domain"
	"github.com/riverlin/aiexpense/internal/usecase"
)

var errTestNotFound = errors.New("not found")

// Test repositories for API integration tests
type TestExpenseRepository struct {
	expenses map[string]*domain.Expense
}

func (r *TestExpenseRepository) Create(ctx context.Context, expense *domain.Expense) error {
	r.expenses[expense.ID] = expense
	return nil
}

func (r *TestExpenseRepository) GetByID(ctx context.Context, id string) (*domain.Expense, error) {
	if exp, ok := r.expenses[id]; ok {
		return exp, nil
	}
	return nil, errTestNotFound
}

func (r *TestExpenseRepository) GetByUserID(ctx context.Context, userID string) ([]*domain.Expense, error) {
	var result []*domain.Expense
	for _, exp := range r.expenses {
		if exp.UserID == userID {
			result = append(result, exp)
		}
	}
	return result, nil
}

func (r *TestExpenseRepository) GetByUserIDAndDateRange(ctx context.Context, userID string, from, to time.Time) ([]*domain.Expense, error) {
	var result []*domain.Expense
	for _, exp := range r.expenses {
		if exp.UserID == userID && !exp.ExpenseDate.Before(from) && !exp.ExpenseDate.After(to) {
			result = append(result, exp)
		}
	}
	return result, nil
}

func (r *TestExpenseRepository) GetByUserIDAndCategory(ctx context.Context, userID, categoryID string) ([]*domain.Expense, error) {
	var result []*domain.Expense
	for _, exp := range r.expenses {
		if exp.UserID == userID && exp.CategoryID != nil && *exp.CategoryID == categoryID {
			result = append(result, exp)
		}
	}
	return result, nil
}

func (r *TestExpenseRepository) Update(ctx context.Context, expense *domain.Expense) error {
	r.expenses[expense.ID] = expense
	return nil
}

func (r *TestExpenseRepository) Delete(ctx context.Context, id string) error {
	delete(r.expenses, id)
	return nil
}

type TestUserRepository struct {
	users map[string]*domain.User
}

func (r *TestUserRepository) Create(ctx context.Context, user *domain.User) error {
	r.users[user.UserID] = user
	return nil
}

func (r *TestUserRepository) GetByID(ctx context.Context, userID string) (*domain.User, error) {
	if user, ok := r.users[userID]; ok {
		return user, nil
	}
	return nil, errTestNotFound
}

func (r *TestUserRepository) Exists(ctx context.Context, userID string) (bool, error) {
	_, ok := r.users[userID]
	return ok, nil
}

type TestCategoryRepository struct {
	categories map[string]*domain.Category
}

func (r *TestCategoryRepository) Create(ctx context.Context, category *domain.Category) error {
	r.categories[category.ID] = category
	return nil
}

func (r *TestCategoryRepository) GetByID(ctx context.Context, id string) (*domain.Category, error) {
	if cat, ok := r.categories[id]; ok {
		return cat, nil
	}
	return nil, errTestNotFound
}

func (r *TestCategoryRepository) GetByUserID(ctx context.Context, userID string) ([]*domain.Category, error) {
	var result []*domain.Category
	for _, cat := range r.categories {
		if cat.UserID == userID {
			result = append(result, cat)
		}
	}
	return result, nil
}

func (r *TestCategoryRepository) GetByUserIDAndName(ctx context.Context, userID, name string) (*domain.Category, error) {
	for _, cat := range r.categories {
		if cat.UserID == userID && cat.Name == name {
			return cat, nil
		}
	}
	return nil, errTestNotFound
}

func (r *TestCategoryRepository) Update(ctx context.Context, category *domain.Category) error {
	r.categories[category.ID] = category
	return nil
}

func (r *TestCategoryRepository) Delete(ctx context.Context, id string) error {
	delete(r.categories, id)
	return nil
}

func (r *TestCategoryRepository) CreateKeyword(ctx context.Context, keyword *domain.CategoryKeyword) error {
	return nil
}

func (r *TestCategoryRepository) GetKeywordsByCategory(ctx context.Context, categoryID string) ([]*domain.CategoryKeyword, error) {
	return []*domain.CategoryKeyword{}, nil
}

func (r *TestCategoryRepository) DeleteKeyword(ctx context.Context, id string) error {
	return nil
}

// Test AI Service
type TestAIService struct{}

var _ ai.Service = (*TestAIService)(nil)

func (s *TestAIService) ParseExpense(ctx context.Context, text string, userID string) (*ai.ParseExpenseResponse, error) {
	return &ai.ParseExpenseResponse{
		Expenses: []*domain.ParsedExpense{
			{
				Amount:      20.0,
				Description: "Test expense",
			},
		},
		Tokens: &ai.TokenMetadata{
			InputTokens:  10,
			OutputTokens: 20,
			TotalTokens:  30,
		},
	}, nil
}

func (s *TestAIService) SuggestCategory(ctx context.Context, description string, userID string) (*ai.SuggestCategoryResponse, error) {
	return &ai.SuggestCategoryResponse{
		Category: "food",
		Tokens: &ai.TokenMetadata{
			InputTokens:  5,
			OutputTokens: 5,
			TotalTokens:  10,
		},
	}, nil
}

// Test Metrics Repository
type TestMetricsRepository struct{}

func (r *TestMetricsRepository) GetDailyActiveUsers(ctx context.Context, from, to time.Time) ([]*domain.DailyMetrics, error) {
	return []*domain.DailyMetrics{}, nil
}

func (r *TestMetricsRepository) GetExpensesSummary(ctx context.Context, from, to time.Time) ([]*domain.DailyMetrics, error) {
	return []*domain.DailyMetrics{}, nil
}

func (r *TestMetricsRepository) GetCategoryTrends(ctx context.Context, userID string, from, to time.Time) ([]*domain.CategoryMetrics, error) {
	return []*domain.CategoryMetrics{}, nil
}

func (r *TestMetricsRepository) GetGrowthMetrics(ctx context.Context, days int) (map[string]interface{}, error) {
	return make(map[string]interface{}), nil
}

func (r *TestMetricsRepository) GetNewUsersPerDay(ctx context.Context, from, to time.Time) ([]*domain.DailyMetrics, error) {
	return []*domain.DailyMetrics{}, nil
}

// TestPolicyRepository for API integration tests
type TestPolicyRepository struct {
	policies map[string]*domain.Policy
}

func (r *TestPolicyRepository) GetByKey(ctx context.Context, key string) (*domain.Policy, error) {
	if p, ok := r.policies[key]; ok {
		return p, nil
	}
	return nil, nil // Return nil if not found (matching sqlite behavior)
}

// TestPricingRepository for API integration tests
type TestPricingRepository struct {
	pricing map[string]*domain.PricingConfig
}

func (r *TestPricingRepository) GetByProviderAndModel(ctx context.Context, provider, model string) (*domain.PricingConfig, error) {
	key := provider + ":" + model
	if p, ok := r.pricing[key]; ok {
		return p, nil
	}
	return nil, nil
}

func (r *TestPricingRepository) GetAll(ctx context.Context) ([]*domain.PricingConfig, error) {
	var result []*domain.PricingConfig
	for _, p := range r.pricing {
		result = append(result, p)
	}
	return result, nil
}

func (r *TestPricingRepository) Create(ctx context.Context, pricing *domain.PricingConfig) error {
	key := pricing.Provider + ":" + pricing.Model
	r.pricing[key] = pricing
	return nil
}

func (r *TestPricingRepository) Update(ctx context.Context, pricing *domain.PricingConfig) error {
	key := pricing.Provider + ":" + pricing.Model
	r.pricing[key] = pricing
	return nil
}

func (r *TestPricingRepository) Deactivate(ctx context.Context, provider, model string) error {
	key := provider + ":" + model
	if p, ok := r.pricing[key]; ok {
		p.IsActive = false
	}
	return nil
}

// TestAICostRepository for API integration tests
type TestAICostRepository struct {
	costs map[string]*domain.AICostLog
}

func (r *TestAICostRepository) Create(ctx context.Context, log *domain.AICostLog) error {
	r.costs[log.ID] = log
	return nil
}

func (r *TestAICostRepository) GetByUserID(ctx context.Context, userID string, limit int) ([]*domain.AICostLog, error) {
	var result []*domain.AICostLog
	for _, log := range r.costs {
		if log.UserID == userID {
			result = append(result, log)
			if len(result) >= limit {
				break
			}
		}
	}
	return result, nil
}

func (r *TestAICostRepository) GetSummary(ctx context.Context, from, to time.Time) (*domain.AICostSummary, error) {
	return &domain.AICostSummary{}, nil
}

func (r *TestAICostRepository) GetDailyStats(ctx context.Context, from, to time.Time) ([]*domain.AICostDailyStats, error) {
	return []*domain.AICostDailyStats{}, nil
}

func (r *TestAICostRepository) GetByOperation(ctx context.Context, from, to time.Time) ([]*domain.AICostByOperation, error) {
	return []*domain.AICostByOperation{}, nil
}

func (r *TestAICostRepository) GetByUserSummary(ctx context.Context, from, to time.Time, limit int) ([]*domain.AICostByUser, error) {
	return []*domain.AICostByUser{}, nil
}

// TestAPIAutoSignupFlow tests complete auto-signup flow
func TestAPIAutoSignupFlow(t *testing.T) {
	userRepo := &TestUserRepository{users: make(map[string]*domain.User)}
	categoryRepo := &TestCategoryRepository{categories: make(map[string]*domain.Category)}
	policyRepo := &TestPolicyRepository{policies: make(map[string]*domain.Policy)}

	handler := NewHandler(
		usecase.NewAutoSignupUseCase(userRepo, categoryRepo),
		nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
		usecase.NewGetPolicyUseCase(policyRepo),
		userRepo, categoryRepo, nil, nil, "",
	)

	// Create request body
	bodyMap := map[string]interface{}{
		"user_id":        "test_api_user",
		"messenger_type": "line",
	}
	bodyBytes, _ := json.Marshal(bodyMap)

	req := httptest.NewRequest("POST", "/api/users/auto-signup", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	handler.AutoSignup(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected %d, got %d", http.StatusOK, w.Code)
	}

	// Verify user was created
	exists, _ := userRepo.Exists(context.Background(), "test_api_user")
	if !exists {
		t.Error("Expected user to be created")
	}
}

// TestAPIAutoSignup tests user auto-signup flow
func TestAPIAutoSignup(t *testing.T) {
	userRepo := &TestUserRepository{users: make(map[string]*domain.User)}
	categoryRepo := &TestCategoryRepository{categories: make(map[string]*domain.Category)}
	policyRepo := &TestPolicyRepository{policies: make(map[string]*domain.Policy)}

	handler := NewHandler(
		usecase.NewAutoSignupUseCase(userRepo, categoryRepo),
		nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
		usecase.NewGetPolicyUseCase(policyRepo),
		userRepo, categoryRepo, nil, nil, "",
	)

	bodyMap := map[string]string{
		"user_id":        "test_user_1",
		"messenger_type": "line",
	}
	bodyBytes, _ := json.Marshal(bodyMap)

	req := httptest.NewRequest("POST", "/api/users/auto-signup", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	handler.AutoSignup(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected %d, got %d", http.StatusOK, w.Code)
	}

	// Verify user was created
	exists, _ := userRepo.Exists(context.Background(), "test_user_1")
	if !exists {
		t.Error("Expected user to be created")
	}

	// Verify default categories created
	categories, _ := categoryRepo.GetByUserID(context.Background(), "test_user_1")
	if len(categories) < 1 {
		t.Error("Expected default categories to be created")
	}
}

// TestAPIParseExpenses tests expense parsing
func TestAPIParseExpenses(t *testing.T) {
	userRepo := &TestUserRepository{users: make(map[string]*domain.User)}
	categoryRepo := &TestCategoryRepository{categories: make(map[string]*domain.Category)}
	aiService := &TestAIService{}
	policyRepo := &TestPolicyRepository{policies: make(map[string]*domain.Policy)}
	pricingRepo := &TestPricingRepository{pricing: make(map[string]*domain.PricingConfig)}
	costRepo := &TestAICostRepository{costs: make(map[string]*domain.AICostLog)}

	handler := NewHandler(
		usecase.NewAutoSignupUseCase(userRepo, categoryRepo),
		usecase.NewParseConversationUseCase(aiService, pricingRepo, costRepo, "gemini", "gemini-2.5-lite"),
		nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
		usecase.NewGetPolicyUseCase(policyRepo),
		userRepo, categoryRepo, nil, nil, "",
	)

	bodyMap := map[string]string{
		"user_id": "test_user_1",
		"text":    "早餐$20午餐$30",
	}
	bodyBytes, _ := json.Marshal(bodyMap)

	req := httptest.NewRequest("POST", "/api/expenses/parse", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	handler.ParseExpenses(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected %d, got %d", http.StatusOK, w.Code)
	}
}

// TestAPICreateExpense tests expense creation
func TestAPICreateExpense(t *testing.T) {
	userRepo := &TestUserRepository{users: make(map[string]*domain.User)}
	categoryRepo := &TestCategoryRepository{categories: make(map[string]*domain.Category)}
	expenseRepo := &TestExpenseRepository{expenses: make(map[string]*domain.Expense)}
	aiService := &TestAIService{}

	// Create user first
	userRepo.Create(context.Background(), &domain.User{
		UserID:        "test_user_1",
		MessengerType: "line",
		CreatedAt:     time.Now(),
	})

	policyRepo := &TestPolicyRepository{policies: make(map[string]*domain.Policy)}
	pricingRepo := &TestPricingRepository{pricing: make(map[string]*domain.PricingConfig)}
	costRepo := &TestAICostRepository{costs: make(map[string]*domain.AICostLog)}

	handler := NewHandler(
		usecase.NewAutoSignupUseCase(userRepo, categoryRepo),
		usecase.NewParseConversationUseCase(aiService, pricingRepo, costRepo, "gemini", "gemini-2.5-lite"),
		usecase.NewCreateExpenseUseCase(expenseRepo, categoryRepo, aiService),
		nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
		usecase.NewGetPolicyUseCase(policyRepo),
		userRepo, categoryRepo, expenseRepo, nil, "",
	)

	bodyMap := map[string]interface{}{
		"user_id":     "test_user_1",
		"description": "Lunch",
		"amount":      25.50,
	}
	bodyBytes, _ := json.Marshal(bodyMap)

	req := httptest.NewRequest("POST", "/api/expenses", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	handler.CreateExpense(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected %d, got %d", http.StatusCreated, w.Code)
	}

	// Verify expense was created
	expenses, _ := expenseRepo.GetByUserID(context.Background(), "test_user_1")
	if len(expenses) < 1 {
		t.Error("Expected expense to be created")
	}
}

// TestAPIGetExpenses tests expense retrieval
func TestAPIGetExpenses(t *testing.T) {
	userRepo := &TestUserRepository{users: make(map[string]*domain.User)}
	categoryRepo := &TestCategoryRepository{categories: make(map[string]*domain.Category)}
	expenseRepo := &TestExpenseRepository{expenses: make(map[string]*domain.Expense)}

	// Create test data
	userRepo.Create(context.Background(), &domain.User{
		UserID:        "test_user_1",
		MessengerType: "line",
		CreatedAt:     time.Now(),
	})

	expenseRepo.Create(context.Background(), &domain.Expense{
		ID:          "exp_001",
		UserID:      "test_user_1",
		Description: "Test expense",
		Amount:      20.0,
		ExpenseDate: time.Now(),
		CreatedAt:   time.Now(),
	})

	policyRepo := &TestPolicyRepository{policies: make(map[string]*domain.Policy)}

	handler := NewHandler(
		usecase.NewAutoSignupUseCase(userRepo, categoryRepo),
		nil, nil,
		usecase.NewGetExpensesUseCase(expenseRepo, categoryRepo),
		nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
		usecase.NewGetPolicyUseCase(policyRepo),
		userRepo, categoryRepo, expenseRepo, nil, "",
	)

	req := httptest.NewRequest("GET", "/api/expenses?user_id=test_user_1", nil)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	handler.GetExpenses(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected %d, got %d", http.StatusOK, w.Code)
	}

	// Verify expenses were returned
	expenses, _ := expenseRepo.GetByUserID(context.Background(), "test_user_1")
	if len(expenses) < 1 {
		t.Error("Expected at least one expense to be retrieved")
	}
}

// TestAPIMissingRequired tests error handling for missing required fields
func TestAPIMissingRequired(t *testing.T) {
	policyRepo := &TestPolicyRepository{policies: make(map[string]*domain.Policy)}

	handler := NewHandler(
		usecase.NewAutoSignupUseCase(
			&TestUserRepository{users: make(map[string]*domain.User)},
			&TestCategoryRepository{categories: make(map[string]*domain.Category)},
		),
		nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
		usecase.NewGetPolicyUseCase(policyRepo),
		nil, nil, nil, nil, "",
	)

	// Missing user_id
	bodyMap := map[string]string{
		"messenger_type": "line",
	}
	bodyBytes, _ := json.Marshal(bodyMap)

	req := httptest.NewRequest("POST", "/api/users/auto-signup", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	handler.AutoSignup(w, req)

	// Should fail due to missing user_id - expect 4xx status
	if w.Code >= 200 && w.Code < 300 {
		t.Errorf("Expected error status (4xx) for missing required field, got %d", w.Code)
	}
}

// TestAPINotFound tests non-existent user handling
func TestAPINotFound(t *testing.T) {
	userRepo := &TestUserRepository{users: make(map[string]*domain.User)}
	categoryRepo := &TestCategoryRepository{categories: make(map[string]*domain.Category)}
	expenseRepo := &TestExpenseRepository{expenses: make(map[string]*domain.Expense)}
	policyRepo := &TestPolicyRepository{policies: make(map[string]*domain.Policy)}

	handler := NewHandler(
		usecase.NewAutoSignupUseCase(userRepo, categoryRepo),
		nil, nil,
		usecase.NewGetExpensesUseCase(expenseRepo, categoryRepo),
		nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
		usecase.NewGetPolicyUseCase(policyRepo),
		userRepo, categoryRepo, expenseRepo, nil, "",
	)

	// Try to get expenses for non-existent user
	req := httptest.NewRequest("GET", "/api/expenses?user_id=nonexistent_user", nil)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	handler.GetExpenses(w, req)

	// Should succeed but return empty expenses
	if w.Code != http.StatusOK {
		t.Errorf("Expected %d, got %d", http.StatusOK, w.Code)
	}
}

// TestAPICategoryManagement tests category operations
func TestAPICategoryManagement(t *testing.T) {
	userRepo := &TestUserRepository{users: make(map[string]*domain.User)}
	categoryRepo := &TestCategoryRepository{categories: make(map[string]*domain.Category)}

	// Create user first
	userRepo.Create(context.Background(), &domain.User{
		UserID:        "test_user_1",
		MessengerType: "line",
		CreatedAt:     time.Now(),
	})

	policyRepo := &TestPolicyRepository{policies: make(map[string]*domain.Policy)}

	handler := NewHandler(
		usecase.NewAutoSignupUseCase(userRepo, categoryRepo),
		nil, nil, nil, nil, nil,
		usecase.NewManageCategoryUseCase(categoryRepo),
		nil, nil, nil, nil, nil, nil, nil, nil,
		usecase.NewGetPolicyUseCase(policyRepo),
		userRepo, categoryRepo, nil, nil, "",
	)

	// Create category
	bodyMap := map[string]interface{}{
		"user_id": "test_user_1",
		"name":    "Custom Category",
	}
	bodyBytes, _ := json.Marshal(bodyMap)

	req := httptest.NewRequest("POST", "/api/categories", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	handler.CreateCategory(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected %d, got %d", http.StatusOK, w.Code)
	}

	// Verify category was created
	categories, _ := categoryRepo.GetByUserID(context.Background(), "test_user_1")
	if len(categories) < 1 {
		t.Error("Expected category to be created")
	}
}

// TestAPIMultipleExpenses tests creating multiple expenses
func TestAPIMultipleExpenses(t *testing.T) {
	userRepo := &TestUserRepository{users: make(map[string]*domain.User)}
	categoryRepo := &TestCategoryRepository{categories: make(map[string]*domain.Category)}
	expenseRepo := &TestExpenseRepository{expenses: make(map[string]*domain.Expense)}
	aiService := &TestAIService{}

	userRepo.Create(context.Background(), &domain.User{
		UserID:        "test_user_1",
		MessengerType: "line",
		CreatedAt:     time.Now(),
	})

	policyRepo := &TestPolicyRepository{policies: make(map[string]*domain.Policy)}

	handler := NewHandler(
		usecase.NewAutoSignupUseCase(userRepo, categoryRepo),
		nil,
		usecase.NewCreateExpenseUseCase(expenseRepo, categoryRepo, aiService),
		nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
		usecase.NewGetPolicyUseCase(policyRepo),
		userRepo, categoryRepo, expenseRepo, nil, "",
	)

	// Create first expense
	bodyMap1 := map[string]interface{}{
		"user_id":     "test_user_1",
		"description": "Breakfast",
		"amount":      15.0,
	}
	bodyBytes1, _ := json.Marshal(bodyMap1)
	req1 := httptest.NewRequest("POST", "/api/expenses", bytes.NewReader(bodyBytes1))
	req1.Header.Set("Content-Type", "application/json")
	w1 := httptest.NewRecorder()
	handler.CreateExpense(w1, req1)
	if w1.Code != http.StatusCreated {
		t.Errorf("First expense: expected %d, got %d", http.StatusCreated, w1.Code)
	}

	// Create second expense
	bodyMap2 := map[string]interface{}{
		"user_id":     "test_user_1",
		"description": "Lunch",
		"amount":      25.0,
	}
	bodyBytes2, _ := json.Marshal(bodyMap2)
	req2 := httptest.NewRequest("POST", "/api/expenses", bytes.NewReader(bodyBytes2))
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	handler.CreateExpense(w2, req2)
	if w2.Code != http.StatusCreated {
		t.Errorf("Second expense: expected %d, got %d", http.StatusCreated, w2.Code)
	}

	// Verify both expenses created
	expenses, _ := expenseRepo.GetByUserID(context.Background(), "test_user_1")
	if len(expenses) < 2 {
		t.Errorf("Expected 2 expenses, got %d", len(expenses))
	}
}

// TestAPIConcurrentRequests tests concurrent request handling
func TestAPIConcurrentRequests(t *testing.T) {
	userRepo := &TestUserRepository{users: make(map[string]*domain.User)}
	categoryRepo := &TestCategoryRepository{categories: make(map[string]*domain.Category)}
	policyRepo := &TestPolicyRepository{policies: make(map[string]*domain.Policy)}

	handler := NewHandler(
		usecase.NewAutoSignupUseCase(userRepo, categoryRepo),
		nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
		usecase.NewGetPolicyUseCase(policyRepo),
		userRepo, categoryRepo, nil, nil, "",
	)

	// Simulate concurrent signup requests
	done := make(chan bool, 3)

	for i := 1; i <= 3; i++ {
		go func(index int) {
			userID := "concurrent_user_" + string(rune('0'+byte(index)))
			bodyMap := map[string]string{
				"user_id":        userID,
				"messenger_type": "line",
			}
			bodyBytes, _ := json.Marshal(bodyMap)

			req := httptest.NewRequest("POST", "/api/users/auto-signup", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			handler.AutoSignup(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("Concurrent signup %d: expected %d, got %d", index, http.StatusOK, w.Code)
			}

			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 3; i++ {
		<-done
	}
}
