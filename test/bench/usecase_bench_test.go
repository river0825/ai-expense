package bench

import (
	"context"
	"testing"
	"time"

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

type BenchAIService struct{}

func (s *BenchAIService) ParseExpense(ctx context.Context, text string, userID string) ([]*domain.ParsedExpense, error) {
	return []*domain.ParsedExpense{
		{Amount: 20.0, Description: "Test"},
	}, nil
}

func (s *BenchAIService) SuggestCategory(ctx context.Context, description string) (string, error) {
	return "food", nil
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

	uc := usecase.NewCreateExpenseUseCase(expenseRepo, categoryRepo, aiService)
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
	uc := usecase.NewParseConversationUseCase(aiService)
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

	uc := usecase.NewCreateExpenseUseCase(expenseRepo, categoryRepo, aiService)
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

	uc := usecase.NewCreateExpenseUseCase(expenseRepo, categoryRepo, aiService)
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
