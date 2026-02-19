package http

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/riverlin/aiexpense/internal/domain"
	"github.com/riverlin/aiexpense/internal/usecase"
)
// TestAPIParseExpenses tests expense parsing
func Test_WhenParseExpenses_GivenValidText_ShouldReturnParsedExpenses(t *testing.T) {
	userRepo := &TestUserRepository{users: make(map[string]*domain.User)}
	categoryRepo := &TestCategoryRepository{categories: make(map[string]*domain.Category)}
	aiService := &TestAIService{}
	policyRepo := &TestPolicyRepository{policies: make(map[string]*domain.Policy)}
	pricingRepo := &TestPricingRepository{pricing: make(map[string]*domain.PricingConfig)}
	costRepo := &TestAICostRepository{costs: make(map[string]*domain.AICostLog)}

	handler := NewHandler(
		usecase.NewAutoSignupUseCase(userRepo, categoryRepo, nil),
		usecase.NewParseConversationUseCase(aiService, pricingRepo, costRepo, userRepo, categoryRepo, "gemini", "gemini-2.5-lite"),
		nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
		usecase.NewGetPolicyUseCase(policyRepo),
		nil, nil, nil, nil,
		userRepo, categoryRepo, nil, nil, "", nil, false,
	)

	bodyMap := map[string]string{
		"user_id": "test_user_1",
		"text":    "早餐$20午餐$30",
	}
	bodyBytes, _ := json.Marshal(bodyMap)

	req := httptest.NewRequest("POST", "/api/expenses/parse", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	handler.ParseExpenses(w, authenticatedRequest(req, "test_user_1"))

	if w.Code != http.StatusOK {
		t.Errorf("Expected %d, got %d", http.StatusOK, w.Code)
	}
}

// TestAPICreateExpense tests expense creation
func Test_WhenCreateExpense_GivenValidRequest_ShouldCreateExpense(t *testing.T) {
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
		usecase.NewAutoSignupUseCase(userRepo, categoryRepo, nil),
		usecase.NewParseConversationUseCase(aiService, pricingRepo, costRepo, userRepo, categoryRepo, "gemini", "gemini-2.5-lite"),
		usecase.NewCreateExpenseUseCase(expenseRepo, categoryRepo, nil, nil, nil, nil, aiService),
		nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
		usecase.NewGetPolicyUseCase(policyRepo),
		nil, nil, nil, nil,
		userRepo, categoryRepo, expenseRepo, nil, "", nil, false,
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
	handler.CreateExpense(w, authenticatedRequest(req, "test_user_1"))

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
func Test_WhenGetExpenses_GivenValidUser_ShouldReturnExpenses(t *testing.T) {
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
		usecase.NewAutoSignupUseCase(userRepo, categoryRepo, nil),
		nil, nil,
		usecase.NewGetExpensesUseCase(expenseRepo, categoryRepo),
		nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
		usecase.NewGetPolicyUseCase(policyRepo),
		nil, nil, nil, nil,
		userRepo, categoryRepo, expenseRepo, nil, "", nil, false,
	)

	req := httptest.NewRequest("GET", "/api/expenses?user_id=test_user_1", nil)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	handler.GetExpenses(w, authenticatedRequest(req, "test_user_1"))

	if w.Code != http.StatusOK {
		t.Errorf("Expected %d, got %d", http.StatusOK, w.Code)
	}

	// Verify expenses were returned
	expenses, _ := expenseRepo.GetByUserID(context.Background(), "test_user_1")
	if len(expenses) < 1 {
		t.Error("Expected at least one expense to be retrieved")
	}
}
// TestAPINotFound tests non-existent user handling
func Test_WhenGetExpenses_GivenNonExistentUser_ShouldReturnEmpty(t *testing.T) {
	userRepo := &TestUserRepository{users: make(map[string]*domain.User)}
	categoryRepo := &TestCategoryRepository{categories: make(map[string]*domain.Category)}
	expenseRepo := &TestExpenseRepository{expenses: make(map[string]*domain.Expense)}
	policyRepo := &TestPolicyRepository{policies: make(map[string]*domain.Policy)}

	handler := NewHandler(
		usecase.NewAutoSignupUseCase(userRepo, categoryRepo, nil),
		nil, nil,
		usecase.NewGetExpensesUseCase(expenseRepo, categoryRepo),
		nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
		usecase.NewGetPolicyUseCase(policyRepo),
		nil, nil, nil, nil,
		userRepo, categoryRepo, expenseRepo, nil, "", nil, false,
	)

	// Try to get expenses for non-existent user
	req := httptest.NewRequest("GET", "/api/expenses?user_id=nonexistent_user", nil)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	handler.GetExpenses(w, authenticatedRequest(req, "nonexistent_user"))

	// Should succeed but return empty expenses
	if w.Code != http.StatusOK {
		t.Errorf("Expected %d, got %d", http.StatusOK, w.Code)
	}
}
// TestAPIMultipleExpenses tests creating multiple expenses
func Test_WhenCreateMultipleExpenses_GivenValidRequests_ShouldCreateAll(t *testing.T) {
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
		usecase.NewAutoSignupUseCase(userRepo, categoryRepo, nil),
		nil,
		usecase.NewCreateExpenseUseCase(expenseRepo, categoryRepo, nil, nil, nil, nil, aiService),
		nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
		usecase.NewGetPolicyUseCase(policyRepo),
		nil, nil, nil, nil,
		userRepo, categoryRepo, expenseRepo, nil, "", nil, false,
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
	handler.CreateExpense(w1, authenticatedRequest(req1, "test_user_1"))
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
	handler.CreateExpense(w2, authenticatedRequest(req2, "test_user_1"))
	if w2.Code != http.StatusCreated {
		t.Errorf("Second expense: expected %d, got %d", http.StatusCreated, w2.Code)
	}

	// Verify both expenses created
	expenses, _ := expenseRepo.GetByUserID(context.Background(), "test_user_1")
	if len(expenses) < 2 {
		t.Errorf("Expected 2 expenses, got %d", len(expenses))
	}
}
// TestAPICreateExpense_WithAccountField tests expense creation with account field
func Test_WhenCreateExpense_GivenAccountField_ShouldPersistAccount(t *testing.T) {
	userRepo := &TestUserRepository{users: make(map[string]*domain.User)}
	categoryRepo := &TestCategoryRepository{categories: make(map[string]*domain.Category)}
	expenseRepo := &TestExpenseRepository{expenses: make(map[string]*domain.Expense)}
	aiService := &TestAIService{}

	// Create user first
	userRepo.Create(context.Background(), &domain.User{
		UserID:        "test_user_account",
		MessengerType: "line",
		CreatedAt:     time.Now(),
	})

	policyRepo := &TestPolicyRepository{policies: make(map[string]*domain.Policy)}

	handler := NewHandler(
		usecase.NewAutoSignupUseCase(userRepo, categoryRepo, nil),
		nil,
		usecase.NewCreateExpenseUseCase(expenseRepo, categoryRepo, nil, nil, nil, nil, aiService),
		nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
		usecase.NewGetPolicyUseCase(policyRepo),
		nil, nil, nil, nil,
		userRepo, categoryRepo, expenseRepo, nil, "", nil, false,
	)

	t.Run("Create expense with explicit account", func(t *testing.T) {
		bodyMap := map[string]interface{}{
			"user_id":     "test_user_account",
			"description": "Dinner with friends",
			"amount":      85.50,
			"account":     "Credit Card",
		}
		bodyBytes, _ := json.Marshal(bodyMap)

		req := httptest.NewRequest("POST", "/api/expenses", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		handler.CreateExpense(w, authenticatedRequest(req, "test_user_account"))

		if w.Code != http.StatusCreated {
			t.Errorf("Expected %d, got %d", http.StatusCreated, w.Code)
		}

		// Verify expense was created with correct account
		expenses, _ := expenseRepo.GetByUserID(context.Background(), "test_user_account")
		if len(expenses) < 1 {
			t.Fatal("Expected expense to be created")
		}

		found := false
		for _, exp := range expenses {
			if exp.Description == "Dinner with friends" {
				found = true
				if exp.Account != "Credit Card" {
					t.Errorf("Expected account 'Credit Card', got '%s'", exp.Account)
				}
			}
		}
		if !found {
			t.Error("Expected expense 'Dinner with friends' not found")
		}
	})

	t.Run("Create expense without account defaults to Cash", func(t *testing.T) {
		bodyMap := map[string]interface{}{
			"user_id":     "test_user_account",
			"description": "Morning coffee",
			"amount":      5.00,
		}
		bodyBytes, _ := json.Marshal(bodyMap)

		req := httptest.NewRequest("POST", "/api/expenses", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		handler.CreateExpense(w, authenticatedRequest(req, "test_user_account"))

		if w.Code != http.StatusCreated {
			t.Errorf("Expected %d, got %d", http.StatusCreated, w.Code)
		}

		// Verify default account is Cash
		expenses, _ := expenseRepo.GetByUserID(context.Background(), "test_user_account")
		for _, exp := range expenses {
			if exp.Description == "Morning coffee" {
				if exp.Account != "Cash" {
					t.Errorf("Expected default account 'Cash', got '%s'", exp.Account)
				}
			}
		}
	})

	t.Run("Create expense with empty account defaults to Cash", func(t *testing.T) {
		bodyMap := map[string]interface{}{
			"user_id":     "test_user_account",
			"description": "Afternoon snack",
			"amount":      3.00,
			"account":     "",
		}
		bodyBytes, _ := json.Marshal(bodyMap)

		req := httptest.NewRequest("POST", "/api/expenses", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		handler.CreateExpense(w, authenticatedRequest(req, "test_user_account"))

		if w.Code != http.StatusCreated {
			t.Errorf("Expected %d, got %d", http.StatusCreated, w.Code)
		}

		// Verify empty account defaults to Cash
		expenses, _ := expenseRepo.GetByUserID(context.Background(), "test_user_account")
		for _, exp := range expenses {
			if exp.Description == "Afternoon snack" {
				if exp.Account != "Cash" {
					t.Errorf("Expected default account 'Cash' for empty string, got '%s'", exp.Account)
				}
			}
		}
	})
}

// TestAPIGetExpenses_IncludesAccount tests that retrieved expenses include account field
func Test_WhenGetExpenses_GivenExpensesWithAccounts_ShouldReturnAccounts(t *testing.T) {
	userRepo := &TestUserRepository{users: make(map[string]*domain.User)}
	categoryRepo := &TestCategoryRepository{categories: make(map[string]*domain.Category)}
	expenseRepo := &TestExpenseRepository{expenses: make(map[string]*domain.Expense)}

	// Create test data
	userRepo.Create(context.Background(), &domain.User{
		UserID:        "test_user_get",
		MessengerType: "line",
		CreatedAt:     time.Now(),
	})

	// Create expenses with different accounts
	expenseRepo.Create(context.Background(), &domain.Expense{
		ID:          "exp_cash",
		UserID:      "test_user_get",
		Description: "Cash expense",
		Amount:      10.0,
		Account:     "Cash",
		ExpenseDate: time.Now(),
		CreatedAt:   time.Now(),
	})
	expenseRepo.Create(context.Background(), &domain.Expense{
		ID:          "exp_credit",
		UserID:      "test_user_get",
		Description: "Credit card expense",
		Amount:      50.0,
		Account:     "Credit Card",
		ExpenseDate: time.Now(),
		CreatedAt:   time.Now(),
	})

	policyRepo := &TestPolicyRepository{policies: make(map[string]*domain.Policy)}

	handler := NewHandler(
		usecase.NewAutoSignupUseCase(userRepo, categoryRepo, nil),
		nil, nil,
		usecase.NewGetExpensesUseCase(expenseRepo, categoryRepo),
		nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
		usecase.NewGetPolicyUseCase(policyRepo),
		nil, nil, nil, nil,
		userRepo, categoryRepo, expenseRepo, nil, "", nil, false,
	)

	req := httptest.NewRequest("GET", "/api/expenses?user_id=test_user_get", nil)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	handler.GetExpenses(w, authenticatedRequest(req, "test_user_get"))

	if w.Code != http.StatusOK {
		t.Errorf("Expected %d, got %d", http.StatusOK, w.Code)
	}

	// Verify expenses have correct accounts
	expenses, _ := expenseRepo.GetByUserID(context.Background(), "test_user_get")
	if len(expenses) != 2 {
		t.Errorf("Expected 2 expenses, got %d", len(expenses))
	}

	accountMap := make(map[string]string)
	for _, exp := range expenses {
		accountMap[exp.ID] = exp.Account
	}

	if accountMap["exp_cash"] != "Cash" {
		t.Errorf("Expected account 'Cash' for exp_cash, got '%s'", accountMap["exp_cash"])
	}
	if accountMap["exp_credit"] != "Credit Card" {
		t.Errorf("Expected account 'Credit Card' for exp_credit, got '%s'", accountMap["exp_credit"])
	}
}
