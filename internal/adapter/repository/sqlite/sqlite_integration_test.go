package sqlite

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/riverlin/aiexpense/internal/domain"
)

// TestSQLiteUserRepository integration tests
func TestSQLiteUserRepository(t *testing.T) {
	// Create temporary database
	tmpfile, err := os.CreateTemp("", "test_*.db")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpfile.Close()
	dbPath := tmpfile.Name()
	defer os.Remove(dbPath)

	// Change to project root to find migrations
	wd, err := os.Getwd()
	if err == nil {
		defer os.Chdir(wd)
	}

	// Open database
	db, err := OpenDB(dbPath)
	if err != nil {
		t.Skipf("Skipping integration test: could not open database: %v (run from project root)", err)
		return
	}
	defer db.Close()

	repo := NewUserRepository(db)
	ctx := context.Background()

	t.Run("CreateAndGetUser", func(t *testing.T) {
		user := &domain.User{
			UserID:        "test_user_1",
			MessengerType: "line",
			CreatedAt:     time.Now(),
		}

		// Create user
		err := repo.Create(ctx, user)
		if err != nil {
			t.Fatalf("Failed to create user: %v", err)
		}

		// Retrieve user
		retrieved, err := repo.GetByID(ctx, "test_user_1")
		if err != nil {
			t.Fatalf("Failed to get user: %v", err)
		}

		if retrieved.UserID != "test_user_1" {
			t.Errorf("Expected user ID 'test_user_1', got '%s'", retrieved.UserID)
		}
		if retrieved.MessengerType != "line" {
			t.Errorf("Expected messenger type 'line', got '%s'", retrieved.MessengerType)
		}
	})

	t.Run("UserExists", func(t *testing.T) {
		// Check existing user
		exists, err := repo.Exists(ctx, "test_user_1")
		if err != nil {
			t.Fatalf("Failed to check existence: %v", err)
		}
		if !exists {
			t.Error("Expected user to exist")
		}

		// Check non-existent user
		exists, err = repo.Exists(ctx, "nonexistent")
		if err != nil {
			t.Fatalf("Failed to check existence: %v", err)
		}
		if exists {
			t.Error("Expected user to not exist")
		}
	})

	t.Run("MultipleUsers", func(t *testing.T) {
		// Create multiple users
		for i := 2; i <= 5; i++ {
			userID := "test_user_" + string(rune(i))
			user := &domain.User{
				UserID:        userID,
				MessengerType: "telegram",
				CreatedAt:     time.Now(),
			}
			if err := repo.Create(ctx, user); err != nil {
				t.Fatalf("Failed to create user %s: %v", userID, err)
			}
		}

		// Verify all users exist
		for i := 1; i <= 5; i++ {
			userID := "test_user_" + string(rune(i))
			exists, err := repo.Exists(ctx, userID)
			if err != nil {
				t.Fatalf("Failed to check user %s: %v", userID, err)
			}
			if !exists {
				t.Errorf("Expected user %s to exist", userID)
			}
		}
	})
}

// TestSQLiteCategoryRepository integration tests
func TestSQLiteCategoryRepository(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "test_*.db")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())
	tmpfile.Close()

	db, err := OpenDB(tmpfile.Name())
	if err != nil {
		t.Skipf("Skipping integration test: could not open database: %v (run from project root)", err)
		return
	}
	defer db.Close()

	userRepo := NewUserRepository(db)
	categoryRepo := NewCategoryRepository(db)
	ctx := context.Background()

	// Create test user
	user := &domain.User{
		UserID:        "cat_test_user",
		MessengerType: "slack",
		CreatedAt:     time.Now(),
	}
	userRepo.Create(ctx, user)

	t.Run("CreateAndGetCategory", func(t *testing.T) {
		category := &domain.Category{
			ID:        "cat_food",
			UserID:    "cat_test_user",
			Name:      "Food & Dining",
			IsDefault: true,
			CreatedAt: time.Now(),
		}

		err := categoryRepo.Create(ctx, category)
		if err != nil {
			t.Fatalf("Failed to create category: %v", err)
		}

		retrieved, err := categoryRepo.GetByID(ctx, "cat_food")
		if err != nil {
			t.Fatalf("Failed to get category: %v", err)
		}

		if retrieved.Name != "Food & Dining" {
			t.Errorf("Expected category name 'Food & Dining', got '%s'", retrieved.Name)
		}
		if !retrieved.IsDefault {
			t.Error("Expected category to be default")
		}
	})

	t.Run("GetCategoriesByUserID", func(t *testing.T) {
		// Create additional categories
		categories := []string{"Transport", "Shopping", "Entertainment"}
		for i, name := range categories {
			cat := &domain.Category{
				ID:        "cat_" + name,
				UserID:    "cat_test_user",
				Name:      name,
				IsDefault: false,
				CreatedAt: time.Now(),
			}
			if err := categoryRepo.Create(ctx, cat); err != nil {
				t.Fatalf("Failed to create category: %v", err)
			}
			_ = i
		}

		// Retrieve all categories for user
		retrieved, err := categoryRepo.GetByUserID(ctx, "cat_test_user")
		if err != nil {
			t.Fatalf("Failed to get categories: %v", err)
		}

		if len(retrieved) < 1 {
			t.Error("Expected to retrieve at least 1 category")
		}
	})

	t.Run("CategoryKeywords", func(t *testing.T) {
		// Create keyword
		keyword := &domain.CategoryKeyword{
			ID:         "kw_1",
			CategoryID: "cat_food",
			Keyword:    "breakfast",
			Priority:   1,
			CreatedAt:  time.Now(),
		}

		err := categoryRepo.CreateKeyword(ctx, keyword)
		if err != nil {
			t.Fatalf("Failed to create keyword: %v", err)
		}

		// Retrieve keywords
		keywords, err := categoryRepo.GetKeywordsByCategory(ctx, "cat_food")
		if err != nil {
			t.Fatalf("Failed to get keywords: %v", err)
		}

		if len(keywords) == 0 {
			t.Error("Expected to retrieve keywords")
		}
	})

	t.Run("UpdateCategory", func(t *testing.T) {
		category := &domain.Category{
			ID:        "cat_food",
			UserID:    "cat_test_user",
			Name:      "Food & Beverages (Updated)",
			IsDefault: true,
			CreatedAt: time.Now(),
		}

		err := categoryRepo.Update(ctx, category)
		if err != nil {
			t.Fatalf("Failed to update category: %v", err)
		}

		retrieved, err := categoryRepo.GetByID(ctx, "cat_food")
		if err != nil {
			t.Fatalf("Failed to get updated category: %v", err)
		}

		if retrieved.Name != "Food & Beverages (Updated)" {
			t.Errorf("Expected updated name, got '%s'", retrieved.Name)
		}
	})
}

// TestSQLiteExpenseRepository integration tests
func TestSQLiteExpenseRepository(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "test_*.db")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())
	tmpfile.Close()

	db, err := OpenDB(tmpfile.Name())
	if err != nil {
		t.Skipf("Skipping integration test: could not open database: %v (run from project root)", err)
		return
	}
	defer db.Close()

	userRepo := NewUserRepository(db)
	categoryRepo := NewCategoryRepository(db)
	expenseRepo := NewExpenseRepository(db)
	ctx := context.Background()

	// Setup test data
	user := &domain.User{
		UserID:        "exp_test_user",
		MessengerType: "teams",
		CreatedAt:     time.Now(),
	}
	userRepo.Create(ctx, user)

	category := &domain.Category{
		ID:        "cat_exp_test",
		UserID:    "exp_test_user",
		Name:      "Test Category",
		IsDefault: false,
		CreatedAt: time.Now(),
	}
	categoryRepo.Create(ctx, category)

	catID := "cat_exp_test"

	t.Run("CreateAndGetExpense", func(t *testing.T) {
		expense := &domain.Expense{
			ID:          "exp_001",
			UserID:      "exp_test_user",
			Description: "Lunch at restaurant",
			Amount:      25.50,
			CategoryID:  &catID,
			ExpenseDate: time.Now(),
			CreatedAt:   time.Now(),
		}

		err := expenseRepo.Create(ctx, expense)
		if err != nil {
			t.Fatalf("Failed to create expense: %v", err)
		}

		retrieved, err := expenseRepo.GetByID(ctx, "exp_001")
		if err != nil {
			t.Fatalf("Failed to get expense: %v", err)
		}

		if retrieved.Amount != 25.50 {
			t.Errorf("Expected amount 25.50, got %f", retrieved.Amount)
		}
		if retrieved.Description != "Lunch at restaurant" {
			t.Errorf("Expected description 'Lunch at restaurant', got '%s'", retrieved.Description)
		}
	})

	t.Run("GetExpensesByUserID", func(t *testing.T) {
		// Create additional expenses
		for i := 2; i <= 3; i++ {
			expense := &domain.Expense{
				ID:          "exp_00" + string(rune(i)),
				UserID:      "exp_test_user",
				Description: "Expense " + string(rune(i)),
				Amount:      float64(10 * i),
				CategoryID:  &catID,
				ExpenseDate: time.Now().AddDate(0, 0, -i),
				CreatedAt:   time.Now(),
			}
			if err := expenseRepo.Create(ctx, expense); err != nil {
				t.Fatalf("Failed to create expense: %v", err)
			}
		}

		// Retrieve all expenses for user
		expenses, err := expenseRepo.GetByUserID(ctx, "exp_test_user")
		if err != nil {
			t.Fatalf("Failed to get expenses: %v", err)
		}

		if len(expenses) < 1 {
			t.Error("Expected to retrieve at least 1 expense")
		}
	})

	t.Run("UpdateExpense", func(t *testing.T) {
		updated := &domain.Expense{
			ID:          "exp_001",
			UserID:      "exp_test_user",
			Description: "Lunch at fancy restaurant (Updated)",
			Amount:      35.00,
			CategoryID:  &catID,
			ExpenseDate: time.Now(),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		err := expenseRepo.Update(ctx, updated)
		if err != nil {
			t.Fatalf("Failed to update expense: %v", err)
		}

		retrieved, err := expenseRepo.GetByID(ctx, "exp_001")
		if err != nil {
			t.Fatalf("Failed to get updated expense: %v", err)
		}

		if retrieved.Amount != 35.00 {
			t.Errorf("Expected updated amount 35.00, got %f", retrieved.Amount)
		}
	})

	t.Run("DeleteExpense", func(t *testing.T) {
		err := expenseRepo.Delete(ctx, "exp_001")
		if err != nil {
			t.Fatalf("Failed to delete expense: %v", err)
		}

		// Verify deletion
		_, err = expenseRepo.GetByID(ctx, "exp_001")
		if err == nil {
			t.Error("Expected expense to be deleted")
		}
	})

	t.Run("GetByDateRange", func(t *testing.T) {
		now := time.Now()
		from := now.AddDate(0, 0, -5)
		to := now.AddDate(0, 0, 1)

		expenses, err := expenseRepo.GetByUserIDAndDateRange(ctx, "exp_test_user", from, to)
		if err != nil {
			t.Fatalf("Failed to get expenses by date range: %v", err)
		}

		if len(expenses) == 0 {
			t.Error("Expected to retrieve expenses in date range")
		}
	})

	t.Run("GetByCategory", func(t *testing.T) {
		expenses, err := expenseRepo.GetByUserIDAndCategory(ctx, "exp_test_user", catID)
		if err != nil {
			t.Fatalf("Failed to get expenses by category: %v", err)
		}

		if len(expenses) == 0 {
			t.Error("Expected to retrieve expenses in category")
		}
	})
}

// TestSQLiteMetricsRepository integration tests
func TestSQLiteMetricsRepository(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "test_*.db")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())
	tmpfile.Close()

	db, err := OpenDB(tmpfile.Name())
	if err != nil {
		t.Skipf("Skipping integration test: could not open database: %v (run from project root)", err)
		return
	}
	defer db.Close()

	userRepo := NewUserRepository(db)
	expenseRepo := NewExpenseRepository(db)
	metricsRepo := NewMetricsRepository(db)
	ctx := context.Background()

	// Create test users and expenses for metrics
	for i := 1; i <= 3; i++ {
		userID := "metrics_user_" + string(rune(i))
		user := &domain.User{
			UserID:        userID,
			MessengerType: "telegram",
			CreatedAt:     time.Now().AddDate(0, 0, -i),
		}
		userRepo.Create(ctx, user)

		// Create some expenses
		for j := 1; j <= 2; j++ {
			expense := &domain.Expense{
				ID:          userID + "_exp_" + string(rune(j)),
				UserID:      userID,
				Description: "Test expense",
				Amount:      float64(10 * j),
				ExpenseDate: time.Now().AddDate(0, 0, -j),
				CreatedAt:   time.Now(),
			}
			expenseRepo.Create(ctx, expense)
		}
	}

	t.Run("GetDailyActiveUsers", func(t *testing.T) {
		now := time.Now()
		users, err := metricsRepo.GetDailyActiveUsers(ctx, now.AddDate(0, 0, -30), now)
		if err != nil {
			t.Fatalf("Failed to get daily active users: %v", err)
		}

		if len(users) == 0 {
			t.Error("Expected to retrieve daily active users")
		}
	})

	t.Run("GetExpensesSummary", func(t *testing.T) {
		now := time.Now()
		summary, err := metricsRepo.GetExpensesSummary(ctx, now.AddDate(0, 0, -30), now)
		if err != nil {
			t.Fatalf("Failed to get expenses summary: %v", err)
		}

		if len(summary) == 0 {
			t.Error("Expected to retrieve expenses summary")
		}
	})

	t.Run("GetNewUsersPerDay", func(t *testing.T) {
		now := time.Now()
		growth, err := metricsRepo.GetNewUsersPerDay(ctx, now.AddDate(0, 0, -30), now)
		if err != nil {
			t.Fatalf("Failed to get new users per day: %v", err)
		}

		if len(growth) == 0 {
			t.Error("Expected to retrieve new users per day")
		}
	})

	t.Run("GetGrowthMetrics", func(t *testing.T) {
		metrics, err := metricsRepo.GetGrowthMetrics(ctx, 30)
		if err != nil {
			t.Fatalf("Failed to get growth metrics: %v", err)
		}

		if metrics == nil {
			t.Error("Expected to retrieve growth metrics")
		}
	})
}
