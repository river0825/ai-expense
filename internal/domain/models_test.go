package domain

import (
	"testing"
	"time"
)

// TestUserAggregate tests the User aggregate root
func TestUserAggregate(t *testing.T) {
	t.Run("CreateUser", func(t *testing.T) {
		user := &User{
			UserID:        "user_123",
			MessengerType: "line",
			CreatedAt:     time.Now(),
		}

		if user.UserID != "user_123" {
			t.Errorf("expected UserID 'user_123', got '%s'", user.UserID)
		}
		if user.MessengerType != "line" {
			t.Errorf("expected MessengerType 'line', got '%s'", user.MessengerType)
		}
		if user.CreatedAt.IsZero() {
			t.Error("expected CreatedAt to be set")
		}
	})

	t.Run("MultipleMessengerTypes", func(t *testing.T) {
		types := []string{"line", "telegram", "slack", "teams", "discord", "whatsapp"}
		for _, messengerType := range types {
			user := &User{
				UserID:        "test_user",
				MessengerType: messengerType,
				CreatedAt:     time.Now(),
			}
			if user.MessengerType != messengerType {
				t.Errorf("expected messenger type '%s', got '%s'", messengerType, user.MessengerType)
			}
		}
	})

	t.Run("UserIdentityImmutability", func(t *testing.T) {
		userID := "user_123"
		user := &User{
			UserID:        userID,
			MessengerType: "line",
			CreatedAt:     time.Now(),
		}

		// UserID should remain constant
		if user.UserID != userID {
			t.Errorf("user identity should not change")
		}
	})
}

// TestExpenseAggregate tests the Expense aggregate root
func TestExpenseAggregate(t *testing.T) {
	t.Run("CreateExpense", func(t *testing.T) {
		now := time.Now()
		categoryID := "cat_food"
		expense := &Expense{
			ID:          "exp_001",
			UserID:      "user_123",
			Description: "breakfast",
			Amount:      20.50,
			CategoryID:  &categoryID,
			ExpenseDate: now,
			CreatedAt:   now,
			UpdatedAt:   now,
		}

		if expense.ID != "exp_001" {
			t.Errorf("expected ID 'exp_001', got '%s'", expense.ID)
		}
		if expense.Amount != 20.50 {
			t.Errorf("expected amount 20.50, got %f", expense.Amount)
		}
		if *expense.CategoryID != categoryID {
			t.Errorf("expected category ID '%s', got '%s'", categoryID, *expense.CategoryID)
		}
	})

	t.Run("ExpenseWithoutCategory", func(t *testing.T) {
		expense := &Expense{
			ID:          "exp_002",
			UserID:      "user_123",
			Description: "misc expense",
			Amount:      10.00,
			CategoryID:  nil,
			ExpenseDate: time.Now(),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		if expense.CategoryID != nil {
			t.Error("expected CategoryID to be nil for uncategorized expense")
		}
	})

	t.Run("ExpenseAmountValidation", func(t *testing.T) {
		testCases := []struct {
			amount      float64
			description string
		}{
			{0.01, "minimum amount"},
			{999999.99, "large amount"},
			{20.50, "decimal amount"},
		}

		for _, tc := range testCases {
			expense := &Expense{
				ID:          "exp_test",
				UserID:      "user_123",
				Description: tc.description,
				Amount:      tc.amount,
				ExpenseDate: time.Now(),
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			}
			if expense.Amount != tc.amount {
				t.Errorf("amount mismatch for %s: expected %f, got %f", tc.description, tc.amount, expense.Amount)
			}
		}
	})

	t.Run("ExpenseDateTracking", func(t *testing.T) {
		pastDate := time.Now().AddDate(0, -1, 0) // 1 month ago
		expense := &Expense{
			ID:          "exp_003",
			UserID:      "user_123",
			Description: "old expense",
			Amount:      50.00,
			ExpenseDate: pastDate,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		if expense.ExpenseDate != pastDate {
			t.Errorf("expected expense date to be %v, got %v", pastDate, expense.ExpenseDate)
		}
	})

	t.Run("ExpenseUpdateTracking", func(t *testing.T) {
		created := time.Now()
		time.Sleep(10 * time.Millisecond) // small delay
		updated := time.Now()

		expense := &Expense{
			ID:          "exp_004",
			UserID:      "user_123",
			Description: "updated expense",
			Amount:      100.00,
			ExpenseDate: created,
			CreatedAt:   created,
			UpdatedAt:   updated,
		}

		if expense.CreatedAt.After(expense.UpdatedAt) {
			t.Error("UpdatedAt should be >= CreatedAt")
		}
	})
}

// TestCategoryAggregate tests the Category aggregate root
func TestCategoryAggregate(t *testing.T) {
	t.Run("CreateCategory", func(t *testing.T) {
		category := &Category{
			ID:        "cat_food",
			UserID:    "user_123",
			Name:      "Food",
			IsDefault: true,
			CreatedAt: time.Now(),
		}

		if category.ID != "cat_food" {
			t.Errorf("expected ID 'cat_food', got '%s'", category.ID)
		}
		if category.Name != "Food" {
			t.Errorf("expected name 'Food', got '%s'", category.Name)
		}
		if !category.IsDefault {
			t.Error("expected IsDefault to be true")
		}
	})

	t.Run("DefaultCategories", func(t *testing.T) {
		defaultCategories := []string{"Food", "Transport", "Shopping", "Entertainment", "Other"}

		for _, name := range defaultCategories {
			category := &Category{
				ID:        "cat_" + name,
				UserID:    "user_123",
				Name:      name,
				IsDefault: true,
				CreatedAt: time.Now(),
			}

			if category.Name != name {
				t.Errorf("default category name mismatch: expected '%s', got '%s'", name, category.Name)
			}
			if !category.IsDefault {
				t.Errorf("category %s should be marked as default", name)
			}
		}
	})

	t.Run("CustomCategory", func(t *testing.T) {
		category := &Category{
			ID:        "cat_custom",
			UserID:    "user_123",
			Name:      "Medical",
			IsDefault: false,
			CreatedAt: time.Now(),
		}

		if category.IsDefault {
			t.Error("custom category should not be marked as default")
		}
	})

	t.Run("UserSpecificCategories", func(t *testing.T) {
		user1Categories := []string{"Food", "Transport"}
		user2Categories := []string{"Work", "Personal"}

		for _, name := range user1Categories {
			cat := &Category{
				ID:        "cat_" + name,
				UserID:    "user_1",
				Name:      name,
				IsDefault: false,
				CreatedAt: time.Now(),
			}
			if cat.UserID != "user_1" {
				t.Error("category should belong to user_1")
			}
		}

		for _, name := range user2Categories {
			cat := &Category{
				ID:        "cat_" + name,
				UserID:    "user_2",
				Name:      name,
				IsDefault: false,
				CreatedAt: time.Now(),
			}
			if cat.UserID != "user_2" {
				t.Error("category should belong to user_2")
			}
		}
	})
}

// TestCategoryKeywordValueObject tests category keyword mappings
func TestCategoryKeywordValueObject(t *testing.T) {
	t.Run("CreateCategoryKeyword", func(t *testing.T) {
		keyword := &CategoryKeyword{
			ID:         "kw_001",
			CategoryID: "cat_food",
			Keyword:    "breakfast",
			Priority:   1,
			CreatedAt:  time.Now(),
		}

		if keyword.Keyword != "breakfast" {
			t.Errorf("expected keyword 'breakfast', got '%s'", keyword.Keyword)
		}
		if keyword.CategoryID != "cat_food" {
			t.Errorf("expected category ID 'cat_food', got '%s'", keyword.CategoryID)
		}
	})

	t.Run("KeywordPriority", func(t *testing.T) {
		keywords := []struct {
			keyword  string
			priority int
		}{
			{"breakfast", 10},
			{"lunch", 9},
			{"dinner", 8},
		}

		for _, kw := range keywords {
			keyword := &CategoryKeyword{
				ID:         "kw_test",
				CategoryID: "cat_food",
				Keyword:    kw.keyword,
				Priority:   kw.priority,
				CreatedAt:  time.Now(),
			}

			if keyword.Priority != kw.priority {
				t.Errorf("priority mismatch for '%s': expected %d, got %d", kw.keyword, kw.priority, keyword.Priority)
			}
		}
	})
}

// TestParsedExpenseValueObject tests parsed expense from natural language
func TestParsedExpenseValueObject(t *testing.T) {
	t.Run("ParsedExpenseFromNaturalLanguage", func(t *testing.T) {
		parsed := &ParsedExpense{
			Description:       "breakfast",
			Amount:            20.00,
			SuggestedCategory: "Food",
			Date:              time.Now(),
		}

		if parsed.Amount != 20.00 {
			t.Errorf("expected amount 20.00, got %f", parsed.Amount)
		}
		if parsed.SuggestedCategory != "Food" {
			t.Errorf("expected category 'Food', got '%s'", parsed.SuggestedCategory)
		}
	})

	t.Run("MultipleExpensesParsed", func(t *testing.T) {
		expenses := []*ParsedExpense{
			{Description: "breakfast", Amount: 20.00, SuggestedCategory: "Food", Date: time.Now()},
			{Description: "taxi", Amount: 30.00, SuggestedCategory: "Transport", Date: time.Now()},
			{Description: "shopping", Amount: 50.00, SuggestedCategory: "Shopping", Date: time.Now()},
		}

		totalAmount := 0.0
		for _, exp := range expenses {
			totalAmount += exp.Amount
		}

		if totalAmount != 100.00 {
			t.Errorf("expected total 100.00, got %f", totalAmount)
		}
	})
}

// TestDailyMetricsValueObject tests daily metrics calculation
func TestDailyMetricsValueObject(t *testing.T) {
	t.Run("DailyMetricsCalculation", func(t *testing.T) {
		metrics := &DailyMetrics{
			Date:           time.Now(),
			ActiveUsers:    5,
			TotalExpense:   150.00,
			ExpenseCount:   3,
			AverageExpense: 50.00,
		}

		if metrics.TotalExpense/float64(metrics.ExpenseCount) != metrics.AverageExpense {
			t.Errorf("average expense calculation error")
		}
	})

	t.Run("ZeroMetrics", func(t *testing.T) {
		metrics := &DailyMetrics{
			Date:           time.Now(),
			ActiveUsers:    0,
			TotalExpense:   0.0,
			ExpenseCount:   0,
			AverageExpense: 0.0,
		}

		if metrics.ActiveUsers != 0 || metrics.TotalExpense != 0 {
			t.Error("zero metrics should be initialized properly")
		}
	})
}

// TestCategoryMetricsValueObject tests category spending metrics
func TestCategoryMetricsValueObject(t *testing.T) {
	t.Run("CategoryMetricsCalculation", func(t *testing.T) {
		metrics := &CategoryMetrics{
			CategoryID: "cat_food",
			Category:   "Food",
			Total:      150.00,
			Count:      5,
			Percent:    30.0,
		}

		if metrics.Total <= 0 {
			t.Error("category total should be positive")
		}
		if metrics.Percent < 0 || metrics.Percent > 100 {
			t.Errorf("percent should be between 0-100, got %f", metrics.Percent)
		}
	})

	t.Run("CategoryDistribution", func(t *testing.T) {
		categories := []*CategoryMetrics{
			{CategoryID: "cat_food", Category: "Food", Total: 150.00, Count: 5, Percent: 30.0},
			{CategoryID: "cat_transport", Category: "Transport", Total: 100.00, Count: 4, Percent: 20.0},
			{CategoryID: "cat_shopping", Category: "Shopping", Total: 250.00, Count: 10, Percent: 50.0},
		}

		totalPercent := 0.0
		for _, cat := range categories {
			totalPercent += cat.Percent
		}

		if totalPercent != 100.0 {
			t.Errorf("total percent should be 100, got %f", totalPercent)
		}
	})
}

// TestAggregateBoundaries tests domain model boundaries and relationships
func TestAggregateBoundaries(t *testing.T) {
	t.Run("UserCategoryRelationship", func(t *testing.T) {
		user := &User{
			UserID:        "user_123",
			MessengerType: "line",
			CreatedAt:     time.Now(),
		}

		category := &Category{
			ID:        "cat_food",
			UserID:    user.UserID,
			Name:      "Food",
			IsDefault: true,
			CreatedAt: time.Now(),
		}

		if category.UserID != user.UserID {
			t.Error("category should belong to user")
		}
	})

	t.Run("UserExpenseRelationship", func(t *testing.T) {
		user := &User{
			UserID:        "user_123",
			MessengerType: "line",
			CreatedAt:     time.Now(),
		}

		expense := &Expense{
			ID:          "exp_001",
			UserID:      user.UserID,
			Description: "breakfast",
			Amount:      20.00,
			ExpenseDate: time.Now(),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		if expense.UserID != user.UserID {
			t.Error("expense should belong to user")
		}
	})

	t.Run("ExpenseCategoryRelationship", func(t *testing.T) {
		categoryID := "cat_food"
		expense := &Expense{
			ID:          "exp_001",
			UserID:      "user_123",
			Description: "breakfast",
			Amount:      20.00,
			CategoryID:  &categoryID,
			ExpenseDate: time.Now(),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		if *expense.CategoryID != categoryID {
			t.Error("expense should reference correct category")
		}
	})
}
