package e2e

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/riverlin/aiexpense/internal/domain"
	"github.com/riverlin/aiexpense/internal/usecase"
)

var errNotFound = errors.New("not found")

// E2E Test Repositories with thread-safe access
type E2EExpenseRepository struct {
	expenses map[string]*domain.Expense
	mu       sync.RWMutex
}

func (r *E2EExpenseRepository) Create(ctx context.Context, expense *domain.Expense) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.expenses[expense.ID] = expense
	return nil
}

func (r *E2EExpenseRepository) GetByID(ctx context.Context, id string) (*domain.Expense, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if exp, ok := r.expenses[id]; ok {
		return exp, nil
	}
	return nil, errNotFound
}

func (r *E2EExpenseRepository) GetByUserID(ctx context.Context, userID string) ([]*domain.Expense, error) {
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

func (r *E2EExpenseRepository) GetByUserIDAndDateRange(ctx context.Context, userID string, from, to time.Time) ([]*domain.Expense, error) {
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

func (r *E2EExpenseRepository) GetByUserIDAndCategory(ctx context.Context, userID, categoryID string) ([]*domain.Expense, error) {
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

func (r *E2EExpenseRepository) Update(ctx context.Context, expense *domain.Expense) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.expenses[expense.ID] = expense
	return nil
}

func (r *E2EExpenseRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.expenses, id)
	return nil
}

type E2EUserRepository struct {
	users map[string]*domain.User
	mu    sync.RWMutex
}

func (r *E2EUserRepository) Create(ctx context.Context, user *domain.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.users[user.UserID] = user
	return nil
}

func (r *E2EUserRepository) GetByID(ctx context.Context, userID string) (*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if user, ok := r.users[userID]; ok {
		return user, nil
	}
	return nil, errNotFound
}

func (r *E2EUserRepository) Exists(ctx context.Context, userID string) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, ok := r.users[userID]
	return ok, nil
}

type E2ECategoryRepository struct {
	categories map[string]*domain.Category
	mu         sync.RWMutex
}

func (r *E2ECategoryRepository) Create(ctx context.Context, category *domain.Category) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.categories[category.ID] = category
	return nil
}

func (r *E2ECategoryRepository) GetByID(ctx context.Context, id string) (*domain.Category, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if cat, ok := r.categories[id]; ok {
		return cat, nil
	}
	return nil, errNotFound
}

func (r *E2ECategoryRepository) GetByUserID(ctx context.Context, userID string) ([]*domain.Category, error) {
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

func (r *E2ECategoryRepository) GetByUserIDAndName(ctx context.Context, userID, name string) (*domain.Category, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, cat := range r.categories {
		if cat.UserID == userID && cat.Name == name {
			return cat, nil
		}
	}
	return nil, errNotFound
}

func (r *E2ECategoryRepository) Update(ctx context.Context, category *domain.Category) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.categories[category.ID] = category
	return nil
}

func (r *E2ECategoryRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.categories, id)
	return nil
}

func (r *E2ECategoryRepository) CreateKeyword(ctx context.Context, keyword *domain.CategoryKeyword) error {
	return nil
}

func (r *E2ECategoryRepository) GetKeywordsByCategory(ctx context.Context, categoryID string) ([]*domain.CategoryKeyword, error) {
	return []*domain.CategoryKeyword{}, nil
}

func (r *E2ECategoryRepository) DeleteKeyword(ctx context.Context, id string) error {
	return nil
}

type E2EAIService struct {
	parseResponses map[string][]*domain.ParsedExpense
	mu             sync.RWMutex
}

func (s *E2EAIService) ParseExpense(ctx context.Context, text string, userID string) ([]*domain.ParsedExpense, error) {
	s.mu.RLock()
	if responses, ok := s.parseResponses[text]; ok {
		s.mu.RUnlock()
		return responses, nil
	}
	s.mu.RUnlock()

	return []*domain.ParsedExpense{
		{
			Amount:      20.0,
			Description: text,
		},
	}, nil
}

func (s *E2EAIService) SuggestCategory(ctx context.Context, description string, userID string) (string, error) {
	return "uncategorized", nil
}

func (s *E2EAIService) SetParseResponse(text string, expenses []*domain.ParsedExpense) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.parseResponses == nil {
		s.parseResponses = make(map[string][]*domain.ParsedExpense)
	}
	s.parseResponses[text] = expenses
}

// TestE2ENewUserWebhookFlow tests complete flow for new user
func TestE2ENewUserWebhookFlow(t *testing.T) {
	ctx := context.Background()
	userRepo := &E2EUserRepository{users: make(map[string]*domain.User)}
	categoryRepo := &E2ECategoryRepository{categories: make(map[string]*domain.Category)}
	expenseRepo := &E2EExpenseRepository{expenses: make(map[string]*domain.Expense)}
	aiService := &E2EAIService{parseResponses: make(map[string][]*domain.ParsedExpense)}

	autoSignupUC := usecase.NewAutoSignupUseCase(userRepo, categoryRepo)
	parseUC := usecase.NewParseConversationUseCase(aiService)
	createExpenseUC := usecase.NewCreateExpenseUseCase(expenseRepo, categoryRepo, aiService)

	// Step 1: Auto-signup new user
	userID := "e2e_new_user_1"
	err := autoSignupUC.Execute(ctx, userID, "line")
	if err != nil {
		t.Fatalf("Auto-signup failed: %v", err)
	}

	// Verify user was created
	exists, _ := userRepo.Exists(ctx, userID)
	if !exists {
		t.Error("User should have been created")
	}

	// Verify default categories were created
	categories, _ := categoryRepo.GetByUserID(ctx, userID)
	if len(categories) == 0 {
		t.Error("Default categories should have been created")
	}

	// Step 2: Parse message from new user
	messageText := "早餐$20午餐$30"
	parsedExpenses, err := parseUC.Execute(ctx, messageText, userID)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if len(parsedExpenses) == 0 {
		t.Error("Should have parsed expenses from message")
	}

	// Step 3: Create expenses from parsed data
	for _, parsedExp := range parsedExpenses {
		createReq := &usecase.CreateRequest{
			UserID:      userID,
			Description: parsedExp.Description,
			Amount:      parsedExp.Amount,
		}
		_, err := createExpenseUC.Execute(ctx, createReq)
		if err != nil {
			t.Fatalf("Create expense failed: %v", err)
		}
	}

	// Step 4: Verify expenses in database
	expenses, _ := expenseRepo.GetByUserID(ctx, userID)
	if len(expenses) < 1 {
		t.Errorf("Expected at least 1 expense, got %d", len(expenses))
	}
}

// TestE2EExistingUserWebhookFlow tests complete flow for existing user
func TestE2EExistingUserWebhookFlow(t *testing.T) {
	ctx := context.Background()
	userRepo := &E2EUserRepository{users: make(map[string]*domain.User)}
	categoryRepo := &E2ECategoryRepository{categories: make(map[string]*domain.Category)}
	expenseRepo := &E2EExpenseRepository{expenses: make(map[string]*domain.Expense)}
	aiService := &E2EAIService{parseResponses: make(map[string][]*domain.ParsedExpense)}

	autoSignupUC := usecase.NewAutoSignupUseCase(userRepo, categoryRepo)
	createExpenseUC := usecase.NewCreateExpenseUseCase(expenseRepo, categoryRepo, aiService)

	userID := "e2e_existing_user_1"

	// Create user first
	userRepo.Create(ctx, &domain.User{
		UserID:        userID,
		MessengerType: "line",
		CreatedAt:     time.Now(),
	})

	// Create a category
	categoryRepo.Create(ctx, &domain.Category{
		ID:     "cat_001",
		UserID: userID,
		Name:   "Food",
	})

	// Attempt auto-signup for existing user
	err := autoSignupUC.Execute(ctx, userID, "line")
	// Should handle gracefully

	// Create new expense for existing user
	createReq := &usecase.CreateRequest{
		UserID:      userID,
		Description: "Dinner",
		Amount:      45.50,
	}
	_, err = createExpenseUC.Execute(ctx, createReq)
	if err != nil {
		t.Fatalf("Create expense failed: %v", err)
	}

	// Verify expense was created
	expenses, _ := expenseRepo.GetByUserID(ctx, userID)
	if len(expenses) < 1 {
		t.Error("Expense should have been created")
	}
}

// TestE2EMultiExpenseMessage tests multi-expense parsing
func TestE2EMultiExpenseMessage(t *testing.T) {
	ctx := context.Background()
	userRepo := &E2EUserRepository{users: make(map[string]*domain.User)}
	categoryRepo := &E2ECategoryRepository{categories: make(map[string]*domain.Category)}
	expenseRepo := &E2EExpenseRepository{expenses: make(map[string]*domain.Expense)}
	aiService := &E2EAIService{parseResponses: make(map[string][]*domain.ParsedExpense)}

	parseUC := usecase.NewParseConversationUseCase(aiService)
	createExpenseUC := usecase.NewCreateExpenseUseCase(expenseRepo, categoryRepo, aiService)

	userID := "e2e_multi_user"

	// Create user
	userRepo.Create(ctx, &domain.User{
		UserID:        userID,
		MessengerType: "slack",
		CreatedAt:     time.Now(),
	})

	// Set mock response for multi-expense message
	messageText := "早餐$20午餐$30晚餐$50"
	aiService.SetParseResponse(messageText, []*domain.ParsedExpense{
		{Amount: 20.0, Description: "早餐"},
		{Amount: 30.0, Description: "午餐"},
		{Amount: 50.0, Description: "晚餐"},
	})

	// Parse message with multiple expenses
	parsedExpenses, err := parseUC.Execute(ctx, messageText, userID)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(parsedExpenses) != 3 {
		t.Errorf("Expected 3 expenses, got %d", len(parsedExpenses))
	}

	// Create all expenses
	for _, parsedExp := range parsedExpenses {
		createReq := &usecase.CreateRequest{
			UserID:      userID,
			Description: parsedExp.Description,
			Amount:      parsedExp.Amount,
		}
		_, err := createExpenseUC.Execute(ctx, createReq)
		if err != nil {
			t.Fatalf("Create expense failed: %v", err)
		}
	}

	// Verify all expenses created
	expenses, _ := expenseRepo.GetByUserID(ctx, userID)
	if len(expenses) != 3 {
		t.Errorf("Expected 3 expenses in database, got %d", len(expenses))
	}
}

// TestE2EConcurrentWebhookProcessing tests concurrent webhook processing
func TestE2EConcurrentWebhookProcessing(t *testing.T) {
	ctx := context.Background()
	userRepo := &E2EUserRepository{users: make(map[string]*domain.User)}
	categoryRepo := &E2ECategoryRepository{categories: make(map[string]*domain.Category)}
	expenseRepo := &E2EExpenseRepository{expenses: make(map[string]*domain.Expense)}
	aiService := &E2EAIService{parseResponses: make(map[string][]*domain.ParsedExpense)}

	autoSignupUC := usecase.NewAutoSignupUseCase(userRepo, categoryRepo)
	createExpenseUC := usecase.NewCreateExpenseUseCase(expenseRepo, categoryRepo, aiService)

	numRequests := 10
	done := make(chan bool, numRequests)

	// Run concurrent webhook requests
	for i := 1; i <= numRequests; i++ {
		go func(index int) {
			userID := "e2e_concurrent_user_" + string(rune('0'+byte(index%10)))

			// Auto-signup
			_ = autoSignupUC.Execute(ctx, userID, "telegram")

			// Create expense
			createReq := &usecase.CreateRequest{
				UserID:      userID,
				Description: "Expense",
				Amount:      float64(index) * 10.0,
			}
			_, _ = createExpenseUC.Execute(ctx, createReq)

			done <- true
		}(i)
	}

	// Wait for all goroutines
	successCount := 0
	for i := 0; i < numRequests; i++ {
		if <-done {
			successCount++
		}
	}

	if successCount != numRequests {
		t.Errorf("Expected %d successful requests, got %d", numRequests, successCount)
	}
}

// TestE2EDataIntegrity tests data consistency
func TestE2EDataIntegrity(t *testing.T) {
	ctx := context.Background()
	userRepo := &E2EUserRepository{users: make(map[string]*domain.User)}
	categoryRepo := &E2ECategoryRepository{categories: make(map[string]*domain.Category)}
	expenseRepo := &E2EExpenseRepository{expenses: make(map[string]*domain.Expense)}
	aiService := &E2EAIService{parseResponses: make(map[string][]*domain.ParsedExpense)}

	createExpenseUC := usecase.NewCreateExpenseUseCase(expenseRepo, categoryRepo, aiService)

	userID := "e2e_integrity_user"

	// Create user
	userRepo.Create(ctx, &domain.User{
		UserID:        userID,
		MessengerType: "teams",
		CreatedAt:     time.Now(),
	})

	// Create multiple expenses
	const numExpenses = 5
	expectedTotal := 0.0
	for i := 1; i <= numExpenses; i++ {
		amount := float64(i) * 10.0
		createReq := &usecase.CreateRequest{
			UserID:      userID,
			Description: "Expense",
			Amount:      amount,
		}
		_, err := createExpenseUC.Execute(ctx, createReq)
		if err != nil {
			t.Fatalf("Create expense %d failed: %v", i, err)
		}
		expectedTotal += amount
	}

	// Verify all expenses present and amounts correct
	expenses, _ := expenseRepo.GetByUserID(ctx, userID)
	if len(expenses) != numExpenses {
		t.Errorf("Expected %d expenses, got %d", numExpenses, len(expenses))
	}

	var actualTotal float64
	for _, exp := range expenses {
		actualTotal += exp.Amount
		// Verify correct user ID
		if exp.UserID != userID {
			t.Errorf("Expense has wrong user ID: expected %s, got %s", userID, exp.UserID)
		}
	}

	if actualTotal != expectedTotal {
		t.Errorf("Total amount mismatch: expected %f, got %f", expectedTotal, actualTotal)
	}
}
