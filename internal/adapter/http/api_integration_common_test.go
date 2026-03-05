package http

import (
	"context"
	"errors"
	"net/http"
	"sync"
	"time"

	"github.com/riverlin/aiexpense/internal/ai"
	"github.com/riverlin/aiexpense/internal/domain"
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

func (r *TestExpenseRepository) ReassignExpenses(ctx context.Context, sourceID, targetID string) (int, error) {
	count := 0
	for _, exp := range r.expenses {
		if exp.CategoryID != nil && *exp.CategoryID == sourceID {
			exp.CategoryID = &targetID
			count++
		}
	}
	return count, nil
}

func (r *TestExpenseRepository) GetBySourceMessageID(ctx context.Context, messageID string) ([]*domain.Expense, error) {
	var result []*domain.Expense
	// Simple mock implementation
	return result, nil
}

type TestUserRepository struct {
	users map[string]*domain.User
	mu    sync.RWMutex
}

func (r *TestUserRepository) Create(ctx context.Context, user *domain.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.users[user.UserID] = user
	return nil
}

func (r *TestUserRepository) GetByID(ctx context.Context, userID string) (*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if user, ok := r.users[userID]; ok {
		return user, nil
	}
	return nil, errTestNotFound
}

func (r *TestUserRepository) Exists(ctx context.Context, userID string) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, ok := r.users[userID]
	return ok, nil
}

func (r *TestUserRepository) Update(ctx context.Context, user *domain.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.users[user.UserID] = user
	return nil
}

type TestCategoryRepository struct {
	categories map[string]*domain.Category
	mu         sync.RWMutex
}

func (r *TestCategoryRepository) Create(ctx context.Context, category *domain.Category) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.categories[category.ID] = category
	return nil
}

func (r *TestCategoryRepository) GetByID(ctx context.Context, id string) (*domain.Category, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if cat, ok := r.categories[id]; ok {
		return cat, nil
	}
	return nil, errTestNotFound
}

func (r *TestCategoryRepository) GetByUserID(ctx context.Context, userID string) ([]*domain.Category, error) {
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

func (r *TestCategoryRepository) GetByUserIDAndName(ctx context.Context, userID, name string) (*domain.Category, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, cat := range r.categories {
		if cat.UserID == userID && cat.Name == name {
			return cat, nil
		}
	}
	return nil, errTestNotFound
}

func (r *TestCategoryRepository) Update(ctx context.Context, category *domain.Category) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.categories[category.ID] = category
	return nil
}

func (r *TestCategoryRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
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

func (s *TestAIService) ParseExpense(ctx context.Context, text string, userCtx *domain.UserContext) (*ai.ParseExpenseResponse, error) {
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

// TestExchangeRateService is a stub for triggering refresh
type TestExchangeRateService struct {
	refreshCalled bool
	refreshErr    error
}

var _ domain.ExchangeRateService = (*TestExchangeRateService)(nil)

func (s *TestExchangeRateService) Convert(ctx context.Context, amount float64, fromCurrency, toCurrency string, txTime time.Time) (float64, float64, error) {
	return amount, 1.0, nil
}

func (s *TestExchangeRateService) RefreshRates(ctx context.Context) error {
	s.refreshCalled = true
	return s.refreshErr
}

func (s *TestExchangeRateService) GetRate(ctx context.Context, fromCurrency, toCurrency string, txTime time.Time) (*domain.ExchangeRate, error) {
	return nil, nil
}

func (s *TestAIService) SuggestCategory(ctx context.Context, description string, userCtx *domain.UserContext) (*ai.SuggestCategoryResponse, error) {
	return &ai.SuggestCategoryResponse{
		Category: "food",
		Tokens: &ai.TokenMetadata{
			InputTokens:  5,
			OutputTokens: 5,
			TotalTokens:  10,
		},
	}, nil
}

func (s *TestAIService) ClassifyIntent(ctx context.Context, text string, userCtx *domain.UserContext) (*ai.ClassifyIntentResponse, error) {
	return &ai.ClassifyIntentResponse{Intent: &domain.ClassifiedIntent{Type: domain.IntentUnknown}}, nil
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


// TestShortLinkRepository
type TestShortLinkRepository struct {
	links map[string]*domain.ShortLink
	mu    sync.RWMutex
}

func (r *TestShortLinkRepository) Create(ctx context.Context, link *domain.ShortLink) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.links[link.ID] = link
	return nil
}

func (r *TestShortLinkRepository) Get(ctx context.Context, id string) (*domain.ShortLink, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if link, ok := r.links[id]; ok {
		return link, nil
	}
	return nil, errors.New("short link not found")
}



func (r *TestShortLinkRepository) DeprecateByUserID(ctx context.Context, userID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	// No-op for test or iterate and delete
	return nil
}

func (r *TestShortLinkRepository) DeleteExpired(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	// No-op for test
	return nil
}

// TestInteractionLogRepository
type TestInteractionLogRepository struct {
	logs []*domain.InteractionLog
	mu   sync.RWMutex
}

func (r *TestInteractionLogRepository) Create(ctx context.Context, log *domain.InteractionLog) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.logs = append(r.logs, log)
	return nil
}

// TestAccountRepository
type TestAccountRepository struct {
	accounts []*domain.Account
	mu       sync.RWMutex
}

func (r *TestAccountRepository) GetByUserID(userID string) ([]*domain.Account, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []*domain.Account
	for _, acc := range r.accounts {
		if acc.UserID == userID {
			result = append(result, acc)
		}
	}
	return result, nil
}

func (r *TestAccountRepository) Create(account *domain.Account) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.accounts = append(r.accounts, account)
	return nil
}

func (r *TestAccountRepository) Update(userID string, oldName string, newName string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, acc := range r.accounts {
		if acc.UserID == userID && acc.Name == oldName {
			acc.Name = newName
			return nil
		}
	}
	return errors.New("account not found")
}

func (r *TestAccountRepository) Delete(userID string, name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i, acc := range r.accounts {
		if acc.UserID == userID && acc.Name == name {
			r.accounts = append(r.accounts[:i], r.accounts[i+1:]...)
			return nil
		}
	}
	return errors.New("account not found")
}





// Helper for authenticated requests
func authenticatedRequest(req *http.Request, userID string) *http.Request {
	ctx := context.WithValue(req.Context(), "user_id", userID)
	return req.WithContext(ctx)
}
