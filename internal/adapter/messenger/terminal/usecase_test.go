package terminal

import (
	"context"
	"testing"
	"time"

	"github.com/riverlin/aiexpense/internal/domain"
	"github.com/riverlin/aiexpense/internal/usecase"
)

// MockUserRepo is a mock UserRepository for testing
type MockUserRepo struct {
	users map[string]*domain.User
}

func NewMockUserRepo() *MockUserRepo {
	return &MockUserRepo{
		users: make(map[string]*domain.User),
	}
}

func (m *MockUserRepo) Create(ctx context.Context, user *domain.User) error {
	m.users[user.UserID] = user
	return nil
}

func (m *MockUserRepo) GetByID(ctx context.Context, userID string) (*domain.User, error) {
	return m.users[userID], nil
}

func (m *MockUserRepo) Exists(ctx context.Context, userID string) (bool, error) {
	_, exists := m.users[userID]
	return exists, nil
}

func (m *MockUserRepo) GetAll(ctx context.Context) ([]*domain.User, error) {
	var users []*domain.User
	for _, u := range m.users {
		users = append(users, u)
	}
	return users, nil
}

func TestTerminalUseCase_HandleMessage_NewUser(t *testing.T) {
	// Setup
	userRepo := usecase.NewMockUserRepository()
	categoryRepo := usecase.NewMockCategoryRepository()
	expenseRepo := usecase.NewMockExpenseRepository()
	aiService := usecase.NewMockAIService()

	autoSignup := usecase.NewAutoSignupUseCase(userRepo, categoryRepo)
	parseConv := usecase.NewParseConversationUseCase(aiService)
	createExp := usecase.NewCreateExpenseUseCase(expenseRepo, categoryRepo, aiService)
	getExp := usecase.NewGetExpensesUseCase(expenseRepo, categoryRepo)

	tc := NewTerminalUseCase(autoSignup, parseConv, createExp, getExp, userRepo)

	ctx := context.Background()
	userID := "test_user_1"
	message := "breakfast $20 lunch $30"

	// Execute
	result, err := tc.HandleMessage(ctx, userID, message)

	// Assert
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Status != "success" {
		t.Errorf("expected status success, got %s", result.Status)
	}

	if result.Data == nil {
		t.Fatal("expected data in response")
	}

	data := result.Data.(map[string]interface{})
	if data["expenses_created"].(int) != 2 {
		t.Errorf("expected 2 expenses created, got %d", data["expenses_created"])
	}

	// Verify user was created
	user, _ := userRepo.GetByID(ctx, userID)
	if user == nil {
		t.Fatal("user should be created")
	}

	if user.MessengerType != "terminal" {
		t.Errorf("expected messenger_type terminal, got %s", user.MessengerType)
	}
}

func TestTerminalUseCase_HandleMessage_ExistingUser(t *testing.T) {
	// Setup with pre-existing user
	userRepo := usecase.NewMockUserRepository()
	categoryRepo := usecase.NewMockCategoryRepository()
	expenseRepo := usecase.NewMockExpenseRepository()
	aiService := usecase.NewMockAIService()

	// Pre-create user
	ctx := context.Background()
	userID := "existing_user"
	userRepo.Create(ctx, &domain.User{
		UserID:        userID,
		MessengerType: "terminal",
		CreatedAt:     time.Now(),
	})

	// Initialize default categories
	defaultCats := []string{"Food", "Transport", "Shopping", "Entertainment", "Other"}
	for _, name := range defaultCats {
		categoryRepo.Create(ctx, &domain.Category{
			ID:        name,
			UserID:    userID,
			Name:      name,
			IsDefault: true,
			CreatedAt: time.Now(),
		})
	}

	autoSignup := usecase.NewAutoSignupUseCase(userRepo, categoryRepo)
	parseConv := usecase.NewParseConversationUseCase(aiService)
	createExp := usecase.NewCreateExpenseUseCase(expenseRepo, categoryRepo, aiService)
	getExp := usecase.NewGetExpensesUseCase(expenseRepo, categoryRepo)

	tc := NewTerminalUseCase(autoSignup, parseConv, createExp, getExp, userRepo)

	// Execute
	message := "taxi $15"
	result, err := tc.HandleMessage(ctx, userID, message)

	// Assert
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Status != "success" {
		t.Errorf("expected status success, got %s", result.Status)
	}

	data := result.Data.(map[string]interface{})
	if data["expenses_created"].(int) != 1 {
		t.Errorf("expected 1 expense created, got %d", data["expenses_created"])
	}
}

func TestTerminalUseCase_HandleMessage_NoExpensesDetected(t *testing.T) {
	// Setup
	userRepo := usecase.NewMockUserRepository()
	categoryRepo := usecase.NewMockCategoryRepository()
	expenseRepo := usecase.NewMockExpenseRepository()
	aiService := usecase.NewMockAIService()

	autoSignup := usecase.NewAutoSignupUseCase(userRepo, categoryRepo)
	parseConv := usecase.NewParseConversationUseCase(aiService)
	createExp := usecase.NewCreateExpenseUseCase(expenseRepo, categoryRepo, aiService)
	getExp := usecase.NewGetExpensesUseCase(expenseRepo, categoryRepo)

	tc := NewTerminalUseCase(autoSignup, parseConv, createExp, getExp, userRepo)

	ctx := context.Background()
	userID := "test_user"
	message := "hello world"

	// Execute
	result, err := tc.HandleMessage(ctx, userID, message)

	// Assert
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Status != "success" {
		t.Errorf("expected status success, got %s", result.Status)
	}

	data := result.Data.(map[string]interface{})
	if data["expenses_parsed"].(int) != 0 {
		t.Errorf("expected 0 expenses parsed, got %d", data["expenses_parsed"])
	}
}

func TestTerminalUseCase_GetUserInfo_Success(t *testing.T) {
	// Setup
	userRepo := usecase.NewMockUserRepository()
	categoryRepo := usecase.NewMockCategoryRepository()
	expenseRepo := usecase.NewMockExpenseRepository()
	aiService := usecase.NewMockAIService()

	ctx := context.Background()
	userID := "test_user"
	now := time.Now()

	// Create user
	userRepo.Create(ctx, &domain.User{
		UserID:        userID,
		MessengerType: "terminal",
		CreatedAt:     now,
	})

	// Create categories
	categoryRepo.Create(ctx, &domain.Category{
		ID:        "cat_food",
		UserID:    userID,
		Name:      "Food",
		IsDefault: true,
		CreatedAt: now,
	})

	// Create expenses
	expenseRepo.Create(ctx, &domain.Expense{
		ID:          "exp1",
		UserID:      userID,
		Description: "breakfast",
		Amount:      20.0,
		CategoryID:  ptrString("cat_food"),
		ExpenseDate: now,
		CreatedAt:   now,
		UpdatedAt:   now,
	})

	expenseRepo.Create(ctx, &domain.Expense{
		ID:          "exp2",
		UserID:      userID,
		Description: "lunch",
		Amount:      30.0,
		CategoryID:  ptrString("cat_food"),
		ExpenseDate: now,
		CreatedAt:   now,
		UpdatedAt:   now,
	})

	autoSignup := usecase.NewAutoSignupUseCase(userRepo, categoryRepo)
	parseConv := usecase.NewParseConversationUseCase(aiService)
	createExp := usecase.NewCreateExpenseUseCase(expenseRepo, categoryRepo, aiService)
	getExp := usecase.NewGetExpensesUseCase(expenseRepo, categoryRepo)

	tc := NewTerminalUseCase(autoSignup, parseConv, createExp, getExp, userRepo)

	// Execute
	result, err := tc.GetUserInfo(ctx, userID)

	// Assert
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result["user_id"].(string) != userID {
		t.Errorf("expected user_id %s, got %s", userID, result["user_id"])
	}

	if result["messenger_type"].(string) != "terminal" {
		t.Errorf("expected messenger_type terminal, got %s", result["messenger_type"])
	}

	if result["expense_count"].(int) != 2 {
		t.Errorf("expected 2 expenses, got %d", result["expense_count"])
	}

	totalExpense := result["total_expenses"].(float64)
	if totalExpense != 50.0 {
		t.Errorf("expected total 50.0, got %f", totalExpense)
	}

	avgExpense := result["average_expense"].(float64)
	if avgExpense != 25.0 {
		t.Errorf("expected average 25.0, got %f", avgExpense)
	}
}

func TestTerminalUseCase_GetUserInfo_UserNotFound(t *testing.T) {
	// Setup
	userRepo := usecase.NewMockUserRepository()
	categoryRepo := usecase.NewMockCategoryRepository()
	expenseRepo := usecase.NewMockExpenseRepository()
	aiService := usecase.NewMockAIService()

	autoSignup := usecase.NewAutoSignupUseCase(userRepo, categoryRepo)
	parseConv := usecase.NewParseConversationUseCase(aiService)
	createExp := usecase.NewCreateExpenseUseCase(expenseRepo, categoryRepo, aiService)
	getExp := usecase.NewGetExpensesUseCase(expenseRepo, categoryRepo)

	tc := NewTerminalUseCase(autoSignup, parseConv, createExp, getExp, userRepo)

	ctx := context.Background()

	// Execute
	_, err := tc.GetUserInfo(ctx, "nonexistent_user")

	// Assert
	if err == nil {
		t.Fatal("expected error for nonexistent user")
	}
}

// Helper function
func ptrString(s string) *string {
	return &s
}
