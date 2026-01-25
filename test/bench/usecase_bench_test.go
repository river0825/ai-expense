package bench

import (
	"context"
	"testing"
	"time"

	"github.com/riverlin/aiexpense/internal/ai"
	"github.com/riverlin/aiexpense/internal/domain"
	"github.com/riverlin/aiexpense/internal/usecase"
)

// Benchmark repositories - minimal implementation for performance testing
type BenchExpenseRepository struct {
	expenses map[string]*domain.Expense
}

func (r *BenchExpenseRepository) Create(ctx context.Context, expense *domain.Expense) error {
	r.expenses[expense.ID] = expense
	return nil
}

func (r *BenchExpenseRepository) GetByID(ctx context.Context, id string) (*domain.Expense, error) {
	if exp, ok := r.expenses[id]; ok {
		return exp, nil
	}
	return nil, nil
}

func (r *BenchExpenseRepository) GetByUserID(ctx context.Context, userID string) ([]*domain.Expense, error) {
	var result []*domain.Expense
	for _, exp := range r.expenses {
		if exp.UserID == userID {
			result = append(result, exp)
		}
	}
	return result, nil
}

func (r *BenchExpenseRepository) GetByUserIDAndDateRange(ctx context.Context, userID string, from, to time.Time) ([]*domain.Expense, error) {
	var result []*domain.Expense
	for _, exp := range r.expenses {
		if exp.UserID == userID && !exp.ExpenseDate.Before(from) && !exp.ExpenseDate.After(to) {
			result = append(result, exp)
		}
	}
	return result, nil
}

func (r *BenchExpenseRepository) GetByUserIDAndCategory(ctx context.Context, userID, categoryID string) ([]*domain.Expense, error) {
	var result []*domain.Expense
	for _, exp := range r.expenses {
		if exp.UserID == userID && exp.CategoryID != nil && *exp.CategoryID == categoryID {
			result = append(result, exp)
		}
	}
	return result, nil
}

func (r *BenchExpenseRepository) Update(ctx context.Context, expense *domain.Expense) error {
	r.expenses[expense.ID] = expense
	return nil
}

func (r *BenchExpenseRepository) Delete(ctx context.Context, id string) error {
	delete(r.expenses, id)
	return nil
}

type BenchUserRepository struct {
	users map[string]*domain.User
}

func (r *BenchUserRepository) Create(ctx context.Context, user *domain.User) error {
	r.users[user.UserID] = user
	return nil
}

func (r *BenchUserRepository) GetByID(ctx context.Context, userID string) (*domain.User, error) {
	if user, ok := r.users[userID]; ok {
		return user, nil
	}
	return nil, nil
}

func (r *BenchUserRepository) Exists(ctx context.Context, userID string) (bool, error) {
	_, ok := r.users[userID]
	return ok, nil
}

type BenchCategoryRepository struct {
	categories map[string]*domain.Category
}

func (r *BenchCategoryRepository) Create(ctx context.Context, category *domain.Category) error {
	r.categories[category.ID] = category
	return nil
}

func (r *BenchCategoryRepository) GetByID(ctx context.Context, id string) (*domain.Category, error) {
	if cat, ok := r.categories[id]; ok {
		return cat, nil
	}
	return nil, nil
}

func (r *BenchCategoryRepository) GetByUserID(ctx context.Context, userID string) ([]*domain.Category, error) {
	var result []*domain.Category
	for _, cat := range r.categories {
		if cat.UserID == userID {
			result = append(result, cat)
		}
	}
	return result, nil
}

func (r *BenchCategoryRepository) GetByUserIDAndName(ctx context.Context, userID, name string) (*domain.Category, error) {
	for _, cat := range r.categories {
		if cat.UserID == userID && cat.Name == name {
			return cat, nil
		}
	}
	return nil, nil
}

func (r *BenchCategoryRepository) Update(ctx context.Context, category *domain.Category) error {
	r.categories[category.ID] = category
	return nil
}

func (r *BenchCategoryRepository) Delete(ctx context.Context, id string) error {
	delete(r.categories, id)
	return nil
}

func (r *BenchCategoryRepository) CreateKeyword(ctx context.Context, keyword *domain.CategoryKeyword) error {
	return nil
}

func (r *BenchCategoryRepository) GetKeywordsByCategory(ctx context.Context, categoryID string) ([]*domain.CategoryKeyword, error) {
	return []*domain.CategoryKeyword{}, nil
}

func (r *BenchCategoryRepository) DeleteKeyword(ctx context.Context, id string) error {
	return nil
}

// BenchPricingRepository for benchmark tests
type BenchPricingRepository struct {
	pricing map[string]*domain.PricingConfig
}

func (r *BenchPricingRepository) GetByProviderAndModel(ctx context.Context, provider, model string) (*domain.PricingConfig, error) {
	key := provider + ":" + model
	if p, ok := r.pricing[key]; ok {
		return p, nil
	}
	return nil, nil
}

func (r *BenchPricingRepository) GetAll(ctx context.Context) ([]*domain.PricingConfig, error) {
	var result []*domain.PricingConfig
	for _, p := range r.pricing {
		result = append(result, p)
	}
	return result, nil
}

func (r *BenchPricingRepository) Create(ctx context.Context, pricing *domain.PricingConfig) error {
	key := pricing.Provider + ":" + pricing.Model
	r.pricing[key] = pricing
	return nil
}

func (r *BenchPricingRepository) Update(ctx context.Context, pricing *domain.PricingConfig) error {
	key := pricing.Provider + ":" + pricing.Model
	r.pricing[key] = pricing
	return nil
}

func (r *BenchPricingRepository) Deactivate(ctx context.Context, provider, model string) error {
	key := provider + ":" + model
	if p, ok := r.pricing[key]; ok {
		p.IsActive = false
	}
	return nil
}

// BenchAICostRepository for benchmark tests
type BenchAICostRepository struct {
	costs map[string]*domain.AICostLog
}

func (r *BenchAICostRepository) Create(ctx context.Context, log *domain.AICostLog) error {
	r.costs[log.ID] = log
	return nil
}

func (r *BenchAICostRepository) GetByUserID(ctx context.Context, userID string, limit int) ([]*domain.AICostLog, error) {
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

func (r *BenchAICostRepository) GetSummary(ctx context.Context, from, to time.Time) (*domain.AICostSummary, error) {
	return &domain.AICostSummary{}, nil
}

func (r *BenchAICostRepository) GetDailyStats(ctx context.Context, from, to time.Time) ([]*domain.AICostDailyStats, error) {
	return []*domain.AICostDailyStats{}, nil
}

func (r *BenchAICostRepository) GetByOperation(ctx context.Context, from, to time.Time) ([]*domain.AICostByOperation, error) {
	return []*domain.AICostByOperation{}, nil
}

func (r *BenchAICostRepository) GetByUserSummary(ctx context.Context, from, to time.Time, limit int) ([]*domain.AICostByUser, error) {
	return []*domain.AICostByUser{}, nil
}

type BenchAIService struct{}

var _ ai.Service = (*BenchAIService)(nil)

func (s *BenchAIService) ParseExpense(ctx context.Context, text string, userID string) (*ai.ParseExpenseResponse, error) {
	return &ai.ParseExpenseResponse{
		Expenses: []*domain.ParsedExpense{
			{Amount: 20.0, Description: "Test"},
		},
		Tokens: &ai.TokenMetadata{
			InputTokens:  10,
			OutputTokens: 20,
			TotalTokens:  30,
		},
	}, nil
}

func (s *BenchAIService) SuggestCategory(ctx context.Context, description string, userID string) (*ai.SuggestCategoryResponse, error) {
	return &ai.SuggestCategoryResponse{
		Category: "food",
		Tokens: &ai.TokenMetadata{
			InputTokens:  5,
			OutputTokens: 5,
			TotalTokens:  10,
		},
	}, nil
}

// BenchmarkAutoSignup benchmarks the auto-signup use case
func BenchmarkAutoSignup(b *testing.B) {
	userRepo := &BenchUserRepository{users: make(map[string]*domain.User)}
	categoryRepo := &BenchCategoryRepository{categories: make(map[string]*domain.Category)}
	uc := usecase.NewAutoSignupUseCase(userRepo, categoryRepo)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = uc.Execute(ctx, "user_"+string(rune(i%100)), "line")
	}
}

// BenchmarkCreateExpense benchmarks expense creation
func BenchmarkCreateExpense(b *testing.B) {
	expenseRepo := &BenchExpenseRepository{expenses: make(map[string]*domain.Expense)}
	userRepo := &BenchUserRepository{users: make(map[string]*domain.User)}
	categoryRepo := &BenchCategoryRepository{categories: make(map[string]*domain.Category)}
	aiService := &BenchAIService{}

	// Setup user and category
	userRepo.Create(context.Background(), &domain.User{
		UserID:        "bench_user",
		MessengerType: "line",
		CreatedAt:     time.Now(),
	})
	categoryRepo.Create(context.Background(), &domain.Category{
		ID:     "cat_food",
		UserID: "bench_user",
		Name:   "Food",
	})

	uc := usecase.NewCreateExpenseUseCase(expenseRepo, categoryRepo, nil, nil, aiService)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = uc.Execute(ctx, &usecase.CreateRequest{
			UserID:      "bench_user",
			Description: "Expense",
			Amount:      20.0,
		})
	}
}

// BenchmarkParseConversation benchmarks conversation parsing
func BenchmarkParseConversation(b *testing.B) {
	aiService := &BenchAIService{}
	pricingRepo := &BenchPricingRepository{pricing: make(map[string]*domain.PricingConfig)}
	costRepo := &BenchAICostRepository{costs: make(map[string]*domain.AICostLog)}
	uc := usecase.NewParseConversationUseCase(aiService, pricingRepo, costRepo, "gemini", "gemini-2.5-lite")
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = uc.Execute(ctx, "早餐$20午餐$30", "bench_user")
	}
}

// BenchmarkGetExpenses benchmarks expense retrieval
func BenchmarkGetExpenses(b *testing.B) {
	expenseRepo := &BenchExpenseRepository{expenses: make(map[string]*domain.Expense)}
	categoryRepo := &BenchCategoryRepository{categories: make(map[string]*domain.Category)}

	// Populate with test data
	for i := 0; i < 100; i++ {
		expenseRepo.Create(context.Background(), &domain.Expense{
			ID:          "exp_" + string(rune(i)),
			UserID:      "bench_user",
			Description: "Test",
			Amount:      20.0,
			ExpenseDate: time.Now(),
			CreatedAt:   time.Now(),
		})
	}

	uc := usecase.NewGetExpensesUseCase(expenseRepo, categoryRepo)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = uc.ExecuteGetAll(ctx, &usecase.GetAllRequest{UserID: "bench_user"})
	}
}

// BenchmarkMultipleCreateExpenses benchmarks creating many expenses
func BenchmarkMultipleCreateExpenses(b *testing.B) {
	expenseRepo := &BenchExpenseRepository{expenses: make(map[string]*domain.Expense)}
	userRepo := &BenchUserRepository{users: make(map[string]*domain.User)}
	categoryRepo := &BenchCategoryRepository{categories: make(map[string]*domain.Category)}
	aiService := &BenchAIService{}

	userRepo.Create(context.Background(), &domain.User{
		UserID:        "bench_user",
		MessengerType: "line",
		CreatedAt:     time.Now(),
	})

	uc := usecase.NewCreateExpenseUseCase(expenseRepo, categoryRepo, nil, nil, aiService)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = uc.Execute(ctx, &usecase.CreateRequest{
			UserID:      "bench_user",
			Description: "Expense",
			Amount:      float64(i),
		})
	}
}

// BenchmarkUserRegistration benchmarks the complete user registration flow
func BenchmarkUserRegistration(b *testing.B) {
	userRepo := &BenchUserRepository{users: make(map[string]*domain.User)}
	categoryRepo := &BenchCategoryRepository{categories: make(map[string]*domain.Category)}
	autoSignupUC := usecase.NewAutoSignupUseCase(userRepo, categoryRepo)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = autoSignupUC.Execute(ctx, "user_"+string(rune(i)), "telegram")
	}
}

// BenchmarkExpenseRetrieval benchmarks retrieving user expenses from large dataset
func BenchmarkExpenseRetrieval(b *testing.B) {
	expenseRepo := &BenchExpenseRepository{expenses: make(map[string]*domain.Expense)}
	categoryRepo := &BenchCategoryRepository{categories: make(map[string]*domain.Category)}

	// Populate with 1000 expenses
	for i := 0; i < 1000; i++ {
		expenseRepo.Create(context.Background(), &domain.Expense{
			ID:          "exp_" + string(rune(i%100)),
			UserID:      "bench_user_" + string(rune(i%10)),
			Description: "Test",
			Amount:      20.0,
			ExpenseDate: time.Now(),
			CreatedAt:   time.Now(),
		})
	}

	uc := usecase.NewGetExpensesUseCase(expenseRepo, categoryRepo)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = uc.ExecuteGetAll(ctx, &usecase.GetAllRequest{
			UserID: "bench_user_" + string(rune(i%10)),
		})
	}
}

// BenchmarkExpenseCreationWithCategoryLookup benchmarks creation with category lookup
func BenchmarkExpenseCreationWithCategoryLookup(b *testing.B) {
	expenseRepo := &BenchExpenseRepository{expenses: make(map[string]*domain.Expense)}
	userRepo := &BenchUserRepository{users: make(map[string]*domain.User)}
	categoryRepo := &BenchCategoryRepository{categories: make(map[string]*domain.Category)}
	aiService := &BenchAIService{}

	// Setup user and multiple categories
	userRepo.Create(context.Background(), &domain.User{
		UserID:        "bench_user",
		MessengerType: "line",
		CreatedAt:     time.Now(),
	})

	for i := 0; i < 10; i++ {
		categoryRepo.Create(context.Background(), &domain.Category{
			ID:     "cat_" + string(rune(i)),
			UserID: "bench_user",
			Name:   "Category",
		})
	}

	uc := usecase.NewCreateExpenseUseCase(expenseRepo, categoryRepo, nil, nil, aiService)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = uc.Execute(ctx, &usecase.CreateRequest{
			UserID:      "bench_user",
			Description: "Expense",
			Amount:      20.0,
		})
	}
}
