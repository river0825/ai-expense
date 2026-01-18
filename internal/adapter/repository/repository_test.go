package repository

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/riverlin/aiexpense/internal/domain"
	"github.com/riverlin/aiexpense/internal/usecase"
)

// === Tests for In-Memory Mock Repositories ===

// TestUserRepositoryMock tests UserRepository mock implementation
func TestUserRepositoryMock(t *testing.T) {
	t.Run("CreateAndRetrieveUser", func(t *testing.T) {
		repo := usecase.NewMockUserRepository()
		ctx := context.Background()

		user := &domain.User{
			UserID:        "user_123",
			MessengerType: "line",
			CreatedAt:     time.Now(),
		}

		err := repo.Create(ctx, user)
		if err != nil {
			t.Fatalf("failed to create user: %v", err)
		}

		retrieved, err := repo.GetByID(ctx, user.UserID)
		if err != nil {
			t.Fatalf("failed to retrieve user: %v", err)
		}

		if retrieved == nil {
			t.Fatal("expected user to be retrieved")
		}
		if retrieved.UserID != user.UserID {
			t.Errorf("user ID mismatch: expected %s, got %s", user.UserID, retrieved.UserID)
		}
		if retrieved.MessengerType != user.MessengerType {
			t.Errorf("messenger type mismatch: expected %s, got %s", user.MessengerType, retrieved.MessengerType)
		}
	})

	t.Run("CheckUserExists", func(t *testing.T) {
		repo := usecase.NewMockUserRepository()
		ctx := context.Background()

		user := &domain.User{
			UserID:        "user_exists",
			MessengerType: "telegram",
			CreatedAt:     time.Now(),
		}
		repo.Create(ctx, user)

		exists, err := repo.Exists(ctx, user.UserID)
		if err != nil {
			t.Fatalf("failed to check existence: %v", err)
		}
		if !exists {
			t.Error("expected user to exist")
		}

		exists, err = repo.Exists(ctx, "nonexistent_user")
		if err != nil {
			t.Fatalf("failed to check existence: %v", err)
		}
		if exists {
			t.Error("expected user to not exist")
		}
	})

	t.Run("MultipleUsers", func(t *testing.T) {
		repo := usecase.NewMockUserRepository()
		ctx := context.Background()

		users := []*domain.User{
			{UserID: "user_1", MessengerType: "line", CreatedAt: time.Now()},
			{UserID: "user_2", MessengerType: "telegram", CreatedAt: time.Now()},
			{UserID: "user_3", MessengerType: "slack", CreatedAt: time.Now()},
		}

		for _, user := range users {
			repo.Create(ctx, user)
		}

		for _, user := range users {
			exists, _ := repo.Exists(ctx, user.UserID)
			if !exists {
				t.Errorf("user %s should exist", user.UserID)
			}
		}
	})
}

// TestExpenseRepositoryMock tests ExpenseRepository mock implementation
func TestExpenseRepositoryMock(t *testing.T) {
	t.Run("CreateAndRetrieveExpense", func(t *testing.T) {
		repo := usecase.NewMockExpenseRepository()
		ctx := context.Background()

		expense := &domain.Expense{
			ID:          uuid.New().String(),
			UserID:      "user_123",
			Description: "breakfast",
			Amount:      20.50,
			ExpenseDate: time.Now(),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		err := repo.Create(ctx, expense)
		if err != nil {
			t.Fatalf("failed to create expense: %v", err)
		}

		retrieved, err := repo.GetByID(ctx, expense.ID)
		if err != nil {
			t.Fatalf("failed to retrieve expense: %v", err)
		}

		if retrieved == nil {
			t.Fatal("expected expense to be retrieved")
		}
		if retrieved.Amount != expense.Amount {
			t.Errorf("amount mismatch: expected %f, got %f", expense.Amount, retrieved.Amount)
		}
	})

	t.Run("GetExpensesByUserID", func(t *testing.T) {
		repo := usecase.NewMockExpenseRepository()
		ctx := context.Background()

		userID := "user_123"
		expenses := []*domain.Expense{
			{ID: uuid.New().String(), UserID: userID, Amount: 20, ExpenseDate: time.Now(), CreatedAt: time.Now(), UpdatedAt: time.Now()},
			{ID: uuid.New().String(), UserID: userID, Amount: 30, ExpenseDate: time.Now(), CreatedAt: time.Now(), UpdatedAt: time.Now()},
			{ID: uuid.New().String(), UserID: "other_user", Amount: 40, ExpenseDate: time.Now(), CreatedAt: time.Now(), UpdatedAt: time.Now()},
		}

		for _, exp := range expenses {
			repo.Create(ctx, exp)
		}

		retrieved, err := repo.GetByUserID(ctx, userID)
		if err != nil {
			t.Fatalf("failed to retrieve expenses: %v", err)
		}

		if len(retrieved) != 2 {
			t.Errorf("expected 2 expenses for user, got %d", len(retrieved))
		}
	})

	t.Run("UpdateExpense", func(t *testing.T) {
		repo := usecase.NewMockExpenseRepository()
		ctx := context.Background()

		expense := &domain.Expense{
			ID:          uuid.New().String(),
			UserID:      "user_123",
			Description: "breakfast",
			Amount:      20.50,
			ExpenseDate: time.Now(),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		repo.Create(ctx, expense)

		// Update expense
		expense.Description = "breakfast + coffee"
		expense.Amount = 25.00
		time.Sleep(10 * time.Millisecond)
		expense.UpdatedAt = time.Now()

		err := repo.Update(ctx, expense)
		if err != nil {
			t.Fatalf("failed to update expense: %v", err)
		}

		retrieved, _ := repo.GetByID(ctx, expense.ID)
		if retrieved.Amount != 25.00 {
			t.Errorf("update failed: expected amount 25.00, got %f", retrieved.Amount)
		}
	})

	t.Run("DeleteExpense", func(t *testing.T) {
		repo := usecase.NewMockExpenseRepository()
		ctx := context.Background()

		expense := &domain.Expense{
			ID:          uuid.New().String(),
			UserID:      "user_123",
			Description: "breakfast",
			Amount:      20.50,
			ExpenseDate: time.Now(),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		repo.Create(ctx, expense)

		err := repo.Delete(ctx, expense.ID)
		if err != nil {
			t.Fatalf("failed to delete expense: %v", err)
		}

		retrieved, _ := repo.GetByID(ctx, expense.ID)
		if retrieved != nil {
			t.Error("expected expense to be deleted")
		}
	})

	t.Run("GetByUserIDAndDateRange", func(t *testing.T) {
		repo := usecase.NewMockExpenseRepository()
		ctx := context.Background()

		userID := "user_123"
		now := time.Now()
		oneWeekAgo := now.AddDate(0, 0, -7)
		twoWeeksAgo := now.AddDate(0, 0, -14)

		expenses := []*domain.Expense{
			{ID: uuid.New().String(), UserID: userID, Amount: 20, ExpenseDate: now, CreatedAt: now, UpdatedAt: now},
			{ID: uuid.New().String(), UserID: userID, Amount: 30, ExpenseDate: oneWeekAgo, CreatedAt: oneWeekAgo, UpdatedAt: oneWeekAgo},
			{ID: uuid.New().String(), UserID: userID, Amount: 40, ExpenseDate: twoWeeksAgo, CreatedAt: twoWeeksAgo, UpdatedAt: twoWeeksAgo},
		}

		for _, exp := range expenses {
			repo.Create(ctx, exp)
		}

		startDate := oneWeekAgo.AddDate(0, 0, -1)
		endDate := now.AddDate(0, 0, 1)

		retrieved, err := repo.GetByUserIDAndDateRange(ctx, userID, startDate, endDate)
		if err != nil {
			t.Fatalf("failed to retrieve expenses: %v", err)
		}

		if len(retrieved) != 2 {
			t.Errorf("expected 2 expenses in range, got %d", len(retrieved))
		}
	})

	t.Run("GetByUserIDAndCategory", func(t *testing.T) {
		repo := usecase.NewMockExpenseRepository()
		ctx := context.Background()

		userID := "user_123"
		catID := "cat_food"

		expenses := []*domain.Expense{
			{ID: uuid.New().String(), UserID: userID, Amount: 20, CategoryID: &catID, ExpenseDate: time.Now(), CreatedAt: time.Now(), UpdatedAt: time.Now()},
			{ID: uuid.New().String(), UserID: userID, Amount: 30, CategoryID: nil, ExpenseDate: time.Now(), CreatedAt: time.Now(), UpdatedAt: time.Now()},
			{ID: uuid.New().String(), UserID: userID, Amount: 40, CategoryID: &catID, ExpenseDate: time.Now(), CreatedAt: time.Now(), UpdatedAt: time.Now()},
		}

		for _, exp := range expenses {
			repo.Create(ctx, exp)
		}

		retrieved, err := repo.GetByUserIDAndCategory(ctx, userID, catID)
		if err != nil {
			t.Fatalf("failed to retrieve expenses: %v", err)
		}

		if len(retrieved) != 2 {
			t.Errorf("expected 2 expenses in category, got %d", len(retrieved))
		}
	})
}

// TestCategoryRepositoryMock tests CategoryRepository mock implementation
func TestCategoryRepositoryMock(t *testing.T) {
	t.Run("CreateAndRetrieveCategory", func(t *testing.T) {
		repo := usecase.NewMockCategoryRepository()
		ctx := context.Background()

		category := &domain.Category{
			ID:        "cat_food",
			UserID:    "user_123",
			Name:      "Food",
			IsDefault: true,
			CreatedAt: time.Now(),
		}

		err := repo.Create(ctx, category)
		if err != nil {
			t.Fatalf("failed to create category: %v", err)
		}

		retrieved, err := repo.GetByID(ctx, category.ID)
		if err != nil {
			t.Fatalf("failed to retrieve category: %v", err)
		}

		if retrieved == nil {
			t.Fatal("expected category to be retrieved")
		}
		if retrieved.Name != category.Name {
			t.Errorf("name mismatch: expected %s, got %s", category.Name, retrieved.Name)
		}
	})

	t.Run("GetCategoriesByUserID", func(t *testing.T) {
		repo := usecase.NewMockCategoryRepository()
		ctx := context.Background()

		userID := "user_123"
		categories := []*domain.Category{
			{ID: "cat_1", UserID: userID, Name: "Food", IsDefault: true, CreatedAt: time.Now()},
			{ID: "cat_2", UserID: userID, Name: "Transport", IsDefault: true, CreatedAt: time.Now()},
			{ID: "cat_3", UserID: "other_user", Name: "Other", IsDefault: false, CreatedAt: time.Now()},
		}

		for _, cat := range categories {
			repo.Create(ctx, cat)
		}

		retrieved, err := repo.GetByUserID(ctx, userID)
		if err != nil {
			t.Fatalf("failed to retrieve categories: %v", err)
		}

		if len(retrieved) != 2 {
			t.Errorf("expected 2 categories for user, got %d", len(retrieved))
		}
	})

	t.Run("GetByUserIDAndName", func(t *testing.T) {
		repo := usecase.NewMockCategoryRepository()
		ctx := context.Background()

		userID := "user_123"
		category := &domain.Category{
			ID:        "cat_food",
			UserID:    userID,
			Name:      "Food",
			IsDefault: true,
			CreatedAt: time.Now(),
		}

		repo.Create(ctx, category)

		retrieved, err := repo.GetByUserIDAndName(ctx, userID, "Food")
		if err != nil {
			t.Fatalf("failed to retrieve category: %v", err)
		}

		if retrieved == nil {
			t.Fatal("expected category to be retrieved")
		}
		if retrieved.Name != "Food" {
			t.Errorf("name mismatch: expected Food, got %s", retrieved.Name)
		}
	})

	t.Run("CreateAndRetrieveKeywords", func(t *testing.T) {
		repo := usecase.NewMockCategoryRepository()
		ctx := context.Background()

		categoryID := "cat_food"
		keywords := []*domain.CategoryKeyword{
			{ID: "kw_1", CategoryID: categoryID, Keyword: "breakfast", Priority: 1, CreatedAt: time.Now()},
			{ID: "kw_2", CategoryID: categoryID, Keyword: "lunch", Priority: 2, CreatedAt: time.Now()},
			{ID: "kw_3", CategoryID: categoryID, Keyword: "dinner", Priority: 3, CreatedAt: time.Now()},
		}

		for _, kw := range keywords {
			repo.CreateKeyword(ctx, kw)
		}

		retrieved, err := repo.GetKeywordsByCategory(ctx, categoryID)
		if err != nil {
			t.Fatalf("failed to retrieve keywords: %v", err)
		}

		if len(retrieved) != 3 {
			t.Errorf("expected 3 keywords, got %d", len(retrieved))
		}
	})

	t.Run("UpdateCategory", func(t *testing.T) {
		repo := usecase.NewMockCategoryRepository()
		ctx := context.Background()

		category := &domain.Category{
			ID:        "cat_food",
			UserID:    "user_123",
			Name:      "Food",
			IsDefault: true,
			CreatedAt: time.Now(),
		}

		repo.Create(ctx, category)

		// Update category
		category.Name = "Dining"
		err := repo.Update(ctx, category)
		if err != nil {
			t.Fatalf("failed to update category: %v", err)
		}

		retrieved, _ := repo.GetByID(ctx, category.ID)
		if retrieved.Name != "Dining" {
			t.Errorf("update failed: expected Dining, got %s", retrieved.Name)
		}
	})

	t.Run("DeleteCategory", func(t *testing.T) {
		repo := usecase.NewMockCategoryRepository()
		ctx := context.Background()

		category := &domain.Category{
			ID:        "cat_food",
			UserID:    "user_123",
			Name:      "Food",
			IsDefault: true,
			CreatedAt: time.Now(),
		}

		repo.Create(ctx, category)

		err := repo.Delete(ctx, category.ID)
		if err != nil {
			t.Fatalf("failed to delete category: %v", err)
		}

		retrieved, _ := repo.GetByID(ctx, category.ID)
		if retrieved != nil {
			t.Error("expected category to be deleted")
		}
	})

	t.Run("DeleteKeyword", func(t *testing.T) {
		repo := usecase.NewMockCategoryRepository()
		ctx := context.Background()

		keyword := &domain.CategoryKeyword{
			ID:         "kw_1",
			CategoryID: "cat_food",
			Keyword:    "breakfast",
			Priority:   1,
			CreatedAt:  time.Now(),
		}

		repo.CreateKeyword(ctx, keyword)

		err := repo.DeleteKeyword(ctx, keyword.ID)
		if err != nil {
			t.Fatalf("failed to delete keyword: %v", err)
		}

		keywords, _ := repo.GetKeywordsByCategory(ctx, keyword.CategoryID)
		if len(keywords) != 0 {
			t.Errorf("expected 0 keywords after deletion, got %d", len(keywords))
		}
	})
}

// === Repository Contract Tests ===

// RepositoryContractTest runs common tests for any repository implementation
type RepositoryContractTest struct {
	UserRepo       domain.UserRepository
	ExpenseRepo    domain.ExpenseRepository
	CategoryRepo   domain.CategoryRepository
	MetricsRepo    domain.MetricsRepository
}

// TestRepositoryContract tests standard repository operations
func TestRepositoryContract(t *testing.T) {
	t.Run("UserRepositoryContract", func(t *testing.T) {
		repo := usecase.NewMockUserRepository()
		ctx := context.Background()

		// Test Create
		user := &domain.User{UserID: "test_user", MessengerType: "line", CreatedAt: time.Now()}
		if err := repo.Create(ctx, user); err != nil {
			t.Fatalf("Create failed: %v", err)
		}

		// Test GetByID
		retrieved, err := repo.GetByID(ctx, user.UserID)
		if err != nil {
			t.Fatalf("GetByID failed: %v", err)
		}
		if retrieved == nil {
			t.Fatal("GetByID returned nil")
		}

		// Test Exists
		exists, err := repo.Exists(ctx, user.UserID)
		if err != nil {
			t.Fatalf("Exists failed: %v", err)
		}
		if !exists {
			t.Error("Exists returned false for existing user")
		}
	})

	t.Run("ExpenseRepositoryContract", func(t *testing.T) {
		repo := usecase.NewMockExpenseRepository()
		ctx := context.Background()

		now := time.Now()
		expense := &domain.Expense{
			ID:          uuid.New().String(),
			UserID:      "test_user",
			Description: "test",
			Amount:      20.0,
			ExpenseDate: now,
			CreatedAt:   now,
			UpdatedAt:   now,
		}

		// Test Create
		if err := repo.Create(ctx, expense); err != nil {
			t.Fatalf("Create failed: %v", err)
		}

		// Test GetByID
		retrieved, err := repo.GetByID(ctx, expense.ID)
		if err != nil {
			t.Fatalf("GetByID failed: %v", err)
		}
		if retrieved == nil {
			t.Fatal("GetByID returned nil")
		}

		// Test GetByUserID
		expenses, err := repo.GetByUserID(ctx, "test_user")
		if err != nil {
			t.Fatalf("GetByUserID failed: %v", err)
		}
		if len(expenses) == 0 {
			t.Error("GetByUserID returned no expenses")
		}

		// Test Update
		expense.Description = "updated"
		if err := repo.Update(ctx, expense); err != nil {
			t.Fatalf("Update failed: %v", err)
		}

		// Test Delete
		if err := repo.Delete(ctx, expense.ID); err != nil {
			t.Fatalf("Delete failed: %v", err)
		}

		// Verify deletion
		deleted, _ := repo.GetByID(ctx, expense.ID)
		if deleted != nil {
			t.Error("Delete did not remove expense")
		}
	})

	t.Run("CategoryRepositoryContract", func(t *testing.T) {
		repo := usecase.NewMockCategoryRepository()
		ctx := context.Background()

		category := &domain.Category{
			ID:        "cat_test",
			UserID:    "test_user",
			Name:      "Test",
			IsDefault: false,
			CreatedAt: time.Now(),
		}

		// Test Create
		if err := repo.Create(ctx, category); err != nil {
			t.Fatalf("Create failed: %v", err)
		}

		// Test GetByID
		retrieved, err := repo.GetByID(ctx, category.ID)
		if err != nil {
			t.Fatalf("GetByID failed: %v", err)
		}
		if retrieved == nil {
			t.Fatal("GetByID returned nil")
		}

		// Test GetByUserID
		categories, err := repo.GetByUserID(ctx, "test_user")
		if err != nil {
			t.Fatalf("GetByUserID failed: %v", err)
		}
		if len(categories) == 0 {
			t.Error("GetByUserID returned no categories")
		}

		// Test CreateKeyword
		keyword := &domain.CategoryKeyword{
			ID:         "kw_test",
			CategoryID: category.ID,
			Keyword:    "test",
			Priority:   1,
			CreatedAt:  time.Now(),
		}
		if err := repo.CreateKeyword(ctx, keyword); err != nil {
			t.Fatalf("CreateKeyword failed: %v", err)
		}

		// Test GetKeywordsByCategory
		keywords, err := repo.GetKeywordsByCategory(ctx, category.ID)
		if err != nil {
			t.Fatalf("GetKeywordsByCategory failed: %v", err)
		}
		if len(keywords) == 0 {
			t.Error("GetKeywordsByCategory returned no keywords")
		}

		// Test Update
		category.Name = "Updated"
		if err := repo.Update(ctx, category); err != nil {
			t.Fatalf("Update failed: %v", err)
		}

		// Test Delete
		if err := repo.Delete(ctx, category.ID); err != nil {
			t.Fatalf("Delete failed: %v", err)
		}

		// Verify deletion
		deleted, _ := repo.GetByID(ctx, category.ID)
		if deleted != nil {
			t.Error("Delete did not remove category")
		}
	})
}
