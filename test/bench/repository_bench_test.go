package bench

import (
	"context"
	"testing"
	"time"

	"github.com/riverlin/aiexpense/internal/domain"
)

// BenchmarkUserRepositoryCreate benchmarks user creation
func BenchmarkUserRepositoryCreate(b *testing.B) {
	repo := &BenchUserRepository{users: make(map[string]*domain.User)}
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = repo.Create(ctx, &domain.User{
			UserID:        "user_" + string(rune(i)),
			MessengerType: "line",
			CreatedAt:     time.Now(),
		})
	}
}

// BenchmarkUserRepositoryExists benchmarks user existence check
func BenchmarkUserRepositoryExists(b *testing.B) {
	repo := &BenchUserRepository{users: make(map[string]*domain.User)}
	ctx := context.Background()

	// Populate with 100 users
	for i := 0; i < 100; i++ {
		repo.users["user_"+string(rune(i))] = &domain.User{UserID: "user_" + string(rune(i))}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = repo.Exists(ctx, "user_"+string(rune(i%100)))
	}
}

// BenchmarkExpenseRepositoryCreate benchmarks expense creation
func BenchmarkExpenseRepositoryCreate(b *testing.B) {
	repo := &BenchExpenseRepository{expenses: make(map[string]*domain.Expense)}
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = repo.Create(ctx, &domain.Expense{
			ID:          "exp_" + string(rune(i)),
			UserID:      "user_1",
			Description: "Test",
			Amount:      20.0,
			ExpenseDate: time.Now(),
			CreatedAt:   time.Now(),
		})
	}
}

// BenchmarkExpenseRepositoryGetByUserID benchmarks user expense retrieval
func BenchmarkExpenseRepositoryGetByUserID(b *testing.B) {
	repo := &BenchExpenseRepository{expenses: make(map[string]*domain.Expense)}
	ctx := context.Background()

	// Populate with 1000 expenses
	for i := 0; i < 1000; i++ {
		_ = repo.Create(ctx, &domain.Expense{
			ID:          "exp_" + string(rune(i)),
			UserID:      "user_" + string(rune(i%10)),
			Description: "Test",
			Amount:      20.0,
			ExpenseDate: time.Now(),
			CreatedAt:   time.Now(),
		})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = repo.GetByUserID(ctx, "user_"+string(rune(i%10)))
	}
}

// BenchmarkExpenseRepositoryGetByDateRange benchmarks date range queries
func BenchmarkExpenseRepositoryGetByDateRange(b *testing.B) {
	repo := &BenchExpenseRepository{expenses: make(map[string]*domain.Expense)}
	ctx := context.Background()

	// Populate with expenses
	baseTime := time.Now()
	for i := 0; i < 100; i++ {
		_ = repo.Create(ctx, &domain.Expense{
			ID:          "exp_" + string(rune(i)),
			UserID:      "user_1",
			Description: "Test",
			Amount:      20.0,
			ExpenseDate: baseTime.AddDate(0, 0, i%30),
			CreatedAt:   time.Now(),
		})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = repo.GetByUserIDAndDateRange(ctx, "user_1",
			baseTime.AddDate(0, 0, -15), baseTime.AddDate(0, 0, 15))
	}
}

// BenchmarkCategoryRepositoryGetByUserID benchmarks category retrieval
func BenchmarkCategoryRepositoryGetByUserID(b *testing.B) {
	repo := &BenchCategoryRepository{categories: make(map[string]*domain.Category)}
	ctx := context.Background()

	// Populate with categories
	for i := 0; i < 50; i++ {
		_ = repo.Create(ctx, &domain.Category{
			ID:     "cat_" + string(rune(i)),
			UserID: "user_" + string(rune(i%10)),
			Name:   "Category",
		})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = repo.GetByUserID(ctx, "user_"+string(rune(i%10)))
	}
}

// BenchmarkExpenseRepositorySequential benchmarks sequential operations
func BenchmarkExpenseRepositorySequential(b *testing.B) {
	repo := &BenchExpenseRepository{expenses: make(map[string]*domain.Expense)}
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Create
		_ = repo.Create(ctx, &domain.Expense{
			ID:          "exp_" + string(rune(i)),
			UserID:      "user_1",
			Description: "Test",
			Amount:      20.0,
			ExpenseDate: time.Now(),
			CreatedAt:   time.Now(),
		})

		// Read
		_, _ = repo.GetByID(ctx, "exp_"+string(rune(i)))

		// Update
		exp := &domain.Expense{
			ID:          "exp_" + string(rune(i)),
			UserID:      "user_1",
			Description: "Updated",
			Amount:      25.0,
			ExpenseDate: time.Now(),
			CreatedAt:   time.Now(),
		}
		_ = repo.Update(ctx, exp)
	}
}

// BenchmarkCategoryRepositoryGetByName benchmarks category lookup by name
func BenchmarkCategoryRepositoryGetByName(b *testing.B) {
	repo := &BenchCategoryRepository{categories: make(map[string]*domain.Category)}
	ctx := context.Background()

	// Populate with categories
	for i := 0; i < 20; i++ {
		_ = repo.Create(ctx, &domain.Category{
			ID:     "cat_" + string(rune(i)),
			UserID: "user_1",
			Name:   "Category_" + string(rune(i)),
		})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = repo.GetByUserIDAndName(ctx, "user_1", "Category_"+string(rune(i%20)))
	}
}
