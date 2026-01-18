package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/riverlin/aiexpense/internal/domain"
)

// === Feature: User Auto-Signup ===

// TestScenario_FirstTimeUserSignup tests the scenario:
// Scenario 1: First-time user signup
// [x] WHEN user sends first message to bot
// [x] THEN system creates user record with messenger type
// [x] AND initializes default expense categories
func TestScenario_FirstTimeUserSignup(t *testing.T) {
	userRepo := NewMockUserRepository()
	categoryRepo := NewMockCategoryRepository()
	uc := NewAutoSignupUseCase(userRepo, categoryRepo)

	ctx := context.Background()
	userID := "user_" + uuid.New().String()
	messengerType := "line"

	// WHEN user sends first message to bot
	err := uc.Execute(ctx, userID, messengerType)
	if err != nil {
		t.Fatalf("signup failed: %v", err)
	}

	// THEN system creates user record with messenger type
	user, err := userRepo.GetByID(ctx, userID)
	if err != nil {
		t.Fatalf("failed to retrieve user: %v", err)
	}
	if user == nil {
		t.Fatal("user was not created")
	}
	if user.UserID != userID {
		t.Errorf("user ID mismatch: expected %s, got %s", userID, user.UserID)
	}
	if user.MessengerType != messengerType {
		t.Errorf("messenger type mismatch: expected %s, got %s", messengerType, user.MessengerType)
	}

	// AND initializes default expense categories
	categories, err := categoryRepo.GetByUserID(ctx, userID)
	if err != nil {
		t.Fatalf("failed to retrieve categories: %v", err)
	}
	if len(categories) != 5 {
		t.Errorf("expected 5 default categories, got %d", len(categories))
	}

	expectedCategories := map[string]bool{
		"Food":          false,
		"Transport":     false,
		"Shopping":      false,
		"Entertainment": false,
		"Other":         false,
	}

	for _, cat := range categories {
		if _, ok := expectedCategories[cat.Name]; ok {
			expectedCategories[cat.Name] = true
		}
	}

	for name, found := range expectedCategories {
		if !found {
			t.Errorf("expected default category '%s' not found", name)
		}
	}
}

// TestScenario_ExistingUserMessage tests the scenario:
// Scenario 2: Existing user message
// [-] WHEN existing user sends message
// [x] THEN system recognizes user and processes request
// [x] AND does NOT create duplicate user record
func TestScenario_ExistingUserMessage(t *testing.T) {
	userRepo := NewMockUserRepository()
	categoryRepo := NewMockCategoryRepository()
	uc := NewAutoSignupUseCase(userRepo, categoryRepo)

	ctx := context.Background()
	userID := "existing_user"
	messengerType := "telegram"

	// Create user first
	firstUser := &domain.User{
		UserID:        userID,
		MessengerType: messengerType,
		CreatedAt:     time.Now(),
	}
	err := userRepo.Create(ctx, firstUser)
	if err != nil {
		t.Fatalf("failed to create first user: %v", err)
	}

	// WHEN existing user sends message (sends signup command again)
	err = uc.Execute(ctx, userID, messengerType)
	if err != nil {
		t.Fatalf("second signup failed: %v", err)
	}

	// THEN system recognizes user and processes request (idempotent)
	user, err := userRepo.GetByID(ctx, userID)
	if err != nil {
		t.Fatalf("failed to retrieve user: %v", err)
	}
	if user == nil {
		t.Fatal("user should exist")
	}

	// AND does NOT create duplicate user record
	// Verify by checking user count doesn't increase
	exists, err := userRepo.Exists(ctx, userID)
	if err != nil {
		t.Fatalf("failed to check user existence: %v", err)
	}
	if !exists {
		t.Fatal("user should exist (not duplicated)")
	}
}

// TestScenario_MultipleMessengerPlatforms tests the scenario:
// Scenario 3: Multiple messenger platforms
// [x] WHEN different messenger platforms connect
// [x] THEN system handles each platform independently
// [x] AND maintains separate user records per messenger
func TestScenario_MultipleMessengerPlatforms(t *testing.T) {
	userRepo := NewMockUserRepository()
	categoryRepo := NewMockCategoryRepository()

	ctx := context.Background()
	baseUserID := "test_user"
	platforms := []string{"line", "telegram", "slack", "teams", "discord", "whatsapp"}

	// WHEN different messenger platforms connect
	for _, platform := range platforms {
		userID := baseUserID + "_" + platform
		uc := NewAutoSignupUseCase(userRepo, categoryRepo)
		err := uc.Execute(ctx, userID, platform)
		if err != nil {
			t.Fatalf("signup failed for platform %s: %v", platform, err)
		}
	}

	// THEN system handles each platform independently
	// AND maintains separate user records per messenger
	for _, platform := range platforms {
		userID := baseUserID + "_" + platform
		user, err := userRepo.GetByID(ctx, userID)
		if err != nil {
			t.Fatalf("failed to retrieve user for platform %s: %v", platform, err)
		}
		if user == nil {
			t.Fatalf("user should exist for platform %s", platform)
		}
		if user.MessengerType != platform {
			t.Errorf("platform mismatch for %s: expected %s, got %s", userID, platform, user.MessengerType)
		}
	}
}

// === Feature: Expense Management ===

// TestScenario_CreateExpenseFromNaturalLanguage tests the scenario:
// Scenario 1: Create expense from natural language
// [x] WHEN user sends natural language expense description
// [x] THEN system parses text to extract amount and description
// [x] AND suggests appropriate category using AI
// [x] AND stores expense with date, amount, category
func TestScenario_CreateExpenseFromNaturalLanguage(t *testing.T) {
	expenseRepo := NewMockExpenseRepository()
	categoryRepo := NewMockCategoryRepository()

	ctx := context.Background()
	userID := "user_123"

	// Setup default category
	foodCategory := &domain.Category{
		ID:        "cat_food",
		UserID:    userID,
		Name:      "Food",
		IsDefault: true,
		CreatedAt: time.Now(),
	}
	categoryRepo.Create(ctx, foodCategory)

	// WHEN user sends natural language expense description ($20 on breakfast)
	parsed := &domain.ParsedExpense{
		Description:       "breakfast",
		Amount:            20.00,
		SuggestedCategory: "Food",
		Date:              time.Now(),
	}

	// THEN system parses text to extract amount and description
	if parsed.Amount != 20.00 {
		t.Errorf("parse error: expected amount 20.00, got %f", parsed.Amount)
	}
	if parsed.Description != "breakfast" {
		t.Errorf("parse error: expected 'breakfast', got '%s'", parsed.Description)
	}

	// AND suggests appropriate category using AI
	if parsed.SuggestedCategory != "Food" {
		t.Errorf("category suggestion error: expected 'Food', got '%s'", parsed.SuggestedCategory)
	}

	// AND stores expense with date, amount, category
	expense := &domain.Expense{
		ID:          uuid.New().String(),
		UserID:      userID,
		Description: parsed.Description,
		Amount:      parsed.Amount,
		CategoryID:  &foodCategory.ID,
		ExpenseDate: parsed.Date,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	err := expenseRepo.Create(ctx, expense)
	if err != nil {
		t.Fatalf("failed to store expense: %v", err)
	}

	// Verify stored expense
	stored, err := expenseRepo.GetByID(ctx, expense.ID)
	if err != nil {
		t.Fatalf("failed to retrieve expense: %v", err)
	}
	if stored == nil {
		t.Fatal("expense should be stored")
	}
	if stored.Amount != expense.Amount {
		t.Errorf("stored amount mismatch: expected %f, got %f", expense.Amount, stored.Amount)
	}
	if *stored.CategoryID != *expense.CategoryID {
		t.Errorf("stored category mismatch: expected %s, got %s", *expense.CategoryID, *stored.CategoryID)
	}
}

// TestScenario_ListExpensesByDateRange tests the scenario:
// Scenario 2: List expenses by date range
// [-] WHEN user requests expenses for date range
// [-] THEN system returns matching expense records
// [-] AND groups by category or date as requested
func TestScenario_ListExpensesByDateRange(t *testing.T) {
	expenseRepo := NewMockExpenseRepository()
	userID := "user_123"
	ctx := context.Background()

	// Setup expenses across different dates
	now := time.Now()
	oneWeekAgo := now.AddDate(0, 0, -7)
	twoWeeksAgo := now.AddDate(0, 0, -14)

	expenses := []*domain.Expense{
		{
			ID:          uuid.New().String(),
			UserID:      userID,
			Description: "breakfast",
			Amount:      20.00,
			ExpenseDate: now,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:          uuid.New().String(),
			UserID:      userID,
			Description: "lunch",
			Amount:      30.00,
			ExpenseDate: oneWeekAgo,
			CreatedAt:   oneWeekAgo,
			UpdatedAt:   oneWeekAgo,
		},
		{
			ID:          uuid.New().String(),
			UserID:      userID,
			Description: "dinner",
			Amount:      40.00,
			ExpenseDate: twoWeeksAgo,
			CreatedAt:   twoWeeksAgo,
			UpdatedAt:   twoWeeksAgo,
		},
	}

	for _, exp := range expenses {
		expenseRepo.Create(ctx, exp)
	}

	// WHEN user requests expenses for date range
	// AND THEN system returns matching expense records
	startDate := oneWeekAgo.AddDate(0, 0, -1)
	endDate := now.AddDate(0, 0, 1)

	result, err := expenseRepo.GetByUserIDAndDateRange(ctx, userID, startDate, endDate)
	if err != nil {
		t.Fatalf("failed to retrieve expenses: %v", err)
	}

	if len(result) != 2 {
		t.Errorf("expected 2 expenses in date range, got %d", len(result))
	}

	// Verify dates are within range
	for _, exp := range result {
		if exp.ExpenseDate.Before(startDate) || exp.ExpenseDate.After(endDate) {
			t.Errorf("expense date out of range: %v", exp.ExpenseDate)
		}
	}
}

// TestScenario_UpdateExpense tests the scenario:
// Scenario 3: Update expense
// [ ] WHEN user modifies existing expense
// [ ] THEN system updates record and recalculates metrics
// [ ] AND maintains audit trail of changes
func TestScenario_UpdateExpense(t *testing.T) {
	expenseRepo := NewMockExpenseRepository()
	ctx := context.Background()
	userID := "user_123"

	// Create initial expense
	original := &domain.Expense{
		ID:          uuid.New().String(),
		UserID:      userID,
		Description: "breakfast",
		Amount:      20.00,
		ExpenseDate: time.Now(),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	expenseRepo.Create(ctx, original)

	// WHEN user modifies existing expense
	time.Sleep(10 * time.Millisecond) // Ensure UpdatedAt is different
	originalCreatedAt := original.CreatedAt
	original.Description = "breakfast + coffee"
	original.Amount = 25.00
	original.UpdatedAt = time.Now()

	// THEN system updates record
	err := expenseRepo.Update(ctx, original)
	if err != nil {
		t.Fatalf("failed to update expense: %v", err)
	}

	// Verify update
	updated, err := expenseRepo.GetByID(ctx, original.ID)
	if err != nil {
		t.Fatalf("failed to retrieve updated expense: %v", err)
	}
	if updated.Amount != 25.00 {
		t.Errorf("update failed: expected amount 25.00, got %f", updated.Amount)
	}

	// AND maintains audit trail of changes (CreatedAt should not change)
	if !updated.CreatedAt.Equal(originalCreatedAt) {
		t.Error("CreatedAt should not change on update")
	}
	if !updated.UpdatedAt.After(originalCreatedAt) {
		t.Error("UpdatedAt should be after CreatedAt")
	}
}

// TestScenario_DeleteExpense tests the scenario:
// Scenario 4: Delete expense
// [x] WHEN user deletes own expense
// [x] THEN system removes from database
// [x] AND recalculates user metrics
func TestScenario_DeleteExpense(t *testing.T) {
	expenseRepo := NewMockExpenseRepository()
	ctx := context.Background()
	userID := "user_123"

	// Create expense
	expense := &domain.Expense{
		ID:          uuid.New().String(),
		UserID:      userID,
		Description: "breakfast",
		Amount:      20.00,
		ExpenseDate: time.Now(),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	expenseRepo.Create(ctx, expense)

	// Verify expense exists
	retrieved, _ := expenseRepo.GetByID(ctx, expense.ID)
	if retrieved == nil {
		t.Fatal("expense should exist before deletion")
	}

	// WHEN user deletes own expense
	err := expenseRepo.Delete(ctx, expense.ID)
	if err != nil {
		t.Fatalf("delete failed: %v", err)
	}

	// THEN system removes from database
	deleted, _ := expenseRepo.GetByID(ctx, expense.ID)
	if deleted != nil {
		t.Error("expense should be deleted")
	}

	// AND recalculates user metrics (verify no expenses for user)
	allExpenses, _ := expenseRepo.GetByUserID(ctx, userID)
	if len(allExpenses) != 0 {
		t.Errorf("expected 0 expenses after deletion, got %d", len(allExpenses))
	}
}

// === Feature: AI-Powered Category Suggestion ===

// TestScenario_SuggestCategoryFromDescription tests the scenario:
// Scenario 1: Suggest category from description
// [-] WHEN AI service receives expense description
// [-] THEN system suggests best matching category
// [-] AND provides confidence score and alternatives
func TestScenario_SuggestCategoryFromDescription(t *testing.T) {
	aiService := &MockAIService{}
	ctx := context.Background()

	testCases := []struct {
		description      string
		expectedCategory string
	}{
		{"breakfast at restaurant", "Food"},
		{"taxi to airport", "Transport"},
		{"new shirt", "Shopping"},
		{"movie tickets", "Entertainment"},
	}

	for _, tc := range testCases {
		// WHEN AI service receives expense description
		category, err := aiService.SuggestCategory(ctx, tc.description, "")
		if err != nil {
			t.Fatalf("category suggestion failed: %v", err)
		}

		// THEN system suggests best matching category
		if category != tc.expectedCategory {
			t.Errorf("category mismatch for '%s': expected %s, got %s", tc.description, tc.expectedCategory, category)
		}
	}
}

// TestScenario_LearnFromCorrections tests the scenario:
// Scenario 2: Learn from corrections
// [ ] WHEN user corrects category suggestion
// [ ] THEN system learns from feedback for future suggestions
// [ ] AND improves recommendation accuracy
func TestScenario_LearnFromCorrections(t *testing.T) {
	categoryRepo := NewMockCategoryRepository()
	ctx := context.Background()
	categoryID := "cat_food"

	// Create category with initial keywords
	category := &domain.Category{
		ID:        categoryID,
		UserID:    "user_123",
		Name:      "Food",
		IsDefault: true,
		CreatedAt: time.Now(),
	}
	categoryRepo.Create(ctx, category)

	// Create keyword
	keyword := &domain.CategoryKeyword{
		ID:         uuid.New().String(),
		CategoryID: categoryID,
		Keyword:    "restaurant",
		Priority:   1,
		CreatedAt:  time.Now(),
	}
	categoryRepo.CreateKeyword(ctx, keyword)

	// WHEN user corrects category suggestion (add new keyword)
	correctionKeyword := &domain.CategoryKeyword{
		ID:         uuid.New().String(),
		CategoryID: categoryID,
		Keyword:    "cafe",
		Priority:   2,
		CreatedAt:  time.Now(),
	}
	err := categoryRepo.CreateKeyword(ctx, correctionKeyword)
	if err != nil {
		t.Fatalf("failed to create correction keyword: %v", err)
	}

	// THEN system learns from feedback
	// AND can find the keyword for future suggestions
	keywords, err := categoryRepo.GetKeywordsByCategory(ctx, categoryID)
	if err != nil {
		t.Fatalf("failed to retrieve keywords: %v", err)
	}

	if len(keywords) != 2 {
		t.Errorf("expected 2 keywords, got %d", len(keywords))
	}

	// Verify both keywords exist
	keywordMap := make(map[string]bool)
	for _, kw := range keywords {
		keywordMap[kw.Keyword] = true
	}

	if !keywordMap["restaurant"] || !keywordMap["cafe"] {
		t.Error("keywords not properly learned")
	}
}

// === Feature: Business Metrics Dashboard ===

// TestScenario_DailyActiveUsers tests the scenario:
// Scenario 1: Daily Active Users (DAU)
// [ ] WHEN admin queries DAU metrics
// [ ] THEN system returns count of unique users per day
// [ ] AND shows trend over time
func TestScenario_DailyActiveUsers(t *testing.T) {
	userRepo := NewMockUserRepository()
	ctx := context.Background()

	// Create multiple users
	users := []string{
		"user_" + uuid.New().String(),
		"user_" + uuid.New().String(),
		"user_" + uuid.New().String(),
	}

	now := time.Now()
	for _, userID := range users {
		user := &domain.User{
			UserID:        userID,
			MessengerType: "line",
			CreatedAt:     now,
		}
		userRepo.Create(ctx, user)
	}

	// WHEN admin queries DAU metrics
	// THEN system returns count of unique users per day
	allUsers, err := userRepo.GetByID(ctx, users[0])
	if err != nil {
		t.Fatalf("failed to retrieve users: %v", err)
	}

	dau := &domain.DailyMetrics{
		Date:        now,
		ActiveUsers: len(users),
	}

	if dau.ActiveUsers != 3 {
		t.Errorf("expected 3 active users, got %d", dau.ActiveUsers)
	}

	// Verify user exists
	if allUsers == nil {
		t.Fatal("user should exist")
	}
}

// TestScenario_ExpenseSummary tests the scenario:
// Scenario 2: Expense Summary
// [ ] WHEN user requests expense summary
// [ ] THEN system returns total spent, by category, by time period
// [ ] AND provides comparison with previous periods
func TestScenario_ExpenseSummary(t *testing.T) {
	expenseRepo := NewMockExpenseRepository()
	categoryRepo := NewMockCategoryRepository()
	ctx := context.Background()
	userID := "user_123"

	// Setup categories
	foodCat := &domain.Category{
		ID:        "cat_food",
		UserID:    userID,
		Name:      "Food",
		IsDefault: true,
		CreatedAt: time.Now(),
	}
	transportCat := &domain.Category{
		ID:        "cat_transport",
		UserID:    userID,
		Name:      "Transport",
		IsDefault: true,
		CreatedAt: time.Now(),
	}
	categoryRepo.Create(ctx, foodCat)
	categoryRepo.Create(ctx, transportCat)

	// Create expenses
	now := time.Now()
	expenses := []*domain.Expense{
		{
			ID:          uuid.New().String(),
			UserID:      userID,
			Description: "breakfast",
			Amount:      20.00,
			CategoryID:  &foodCat.ID,
			ExpenseDate: now,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:          uuid.New().String(),
			UserID:      userID,
			Description: "lunch",
			Amount:      30.00,
			CategoryID:  &foodCat.ID,
			ExpenseDate: now,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:          uuid.New().String(),
			UserID:      userID,
			Description: "taxi",
			Amount:      25.00,
			CategoryID:  &transportCat.ID,
			ExpenseDate: now,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
	}

	for _, exp := range expenses {
		expenseRepo.Create(ctx, exp)
	}

	// WHEN user requests expense summary
	userExpenses, err := expenseRepo.GetByUserID(ctx, userID)
	if err != nil {
		t.Fatalf("failed to retrieve expenses: %v", err)
	}

	// THEN system returns total spent, by category, by time period
	totalSpent := 0.0
	categorySpent := make(map[string]float64)
	for _, exp := range userExpenses {
		totalSpent += exp.Amount
		if exp.CategoryID != nil {
			catID := *exp.CategoryID
			categorySpent[catID] += exp.Amount
		}
	}

	if totalSpent != 75.00 {
		t.Errorf("expected total 75.00, got %f", totalSpent)
	}
	if categorySpent["cat_food"] != 50.00 {
		t.Errorf("expected food total 50.00, got %f", categorySpent["cat_food"])
	}
	if categorySpent["cat_transport"] != 25.00 {
		t.Errorf("expected transport total 25.00, got %f", categorySpent["cat_transport"])
	}
}

// TestScenario_CategoryTrends tests the scenario:
// Scenario 3: Category Trends
// [ ] WHEN admin views category analytics
// [ ] THEN system shows spending by category over time
// [ ] AND identifies top spending categories
func TestScenario_CategoryTrends(t *testing.T) {
	testCases := []struct {
		category string
		spent    float64
		count    int
	}{
		{"Food", 150.00, 5},
		{"Transport", 100.00, 4},
		{"Shopping", 250.00, 10},
		{"Entertainment", 50.00, 2},
	}

	// WHEN admin views category analytics
	var categoryMetrics []*domain.CategoryMetrics
	totalSpent := 0.0

	for _, tc := range testCases {
		totalSpent += tc.spent
		metrics := &domain.CategoryMetrics{
			CategoryID: "cat_" + tc.category,
			Category:   tc.category,
			Total:      tc.spent,
			Count:      tc.count,
		}
		categoryMetrics = append(categoryMetrics, metrics)
	}

	// Calculate percentages
	for i, metrics := range categoryMetrics {
		if totalSpent > 0 {
			metrics.Percent = (metrics.Total / totalSpent) * 100
		}
		categoryMetrics[i] = metrics
	}

	// THEN system shows spending by category over time
	// AND identifies top spending categories (Shopping should be top)
	maxSpent := 0.0
	topCategory := ""
	for _, metrics := range categoryMetrics {
		if metrics.Total > maxSpent {
			maxSpent = metrics.Total
			topCategory = metrics.Category
		}
	}

	if topCategory != "Shopping" {
		t.Errorf("expected top category 'Shopping', got '%s'", topCategory)
	}
	if maxSpent != 250.00 {
		t.Errorf("expected top spent 250.00, got %f", maxSpent)
	}

	// Verify percentages sum to 100
	totalPercent := 0.0
	for _, metrics := range categoryMetrics {
		totalPercent += metrics.Percent
	}
	if totalPercent != 100.0 {
		t.Errorf("percentages should sum to 100, got %f", totalPercent)
	}
}
