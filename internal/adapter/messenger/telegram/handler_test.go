package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/riverlin/aiexpense/internal/domain"
	"github.com/riverlin/aiexpense/internal/usecase"
)

var errNotFound = errors.New("not found")

// MockExpenseRepository for testing
type MockExpenseRepository struct {
	expenses map[string]*domain.Expense
}

func NewMockExpenseRepository() *MockExpenseRepository {
	return &MockExpenseRepository{
		expenses: make(map[string]*domain.Expense),
	}
}

func (m *MockExpenseRepository) Create(ctx context.Context, expense *domain.Expense) error {
	m.expenses[expense.ID] = expense
	return nil
}

func (m *MockExpenseRepository) GetByID(ctx context.Context, id string) (*domain.Expense, error) {
	if exp, ok := m.expenses[id]; ok {
		return exp, nil
	}
	return nil, errNotFound
}

func (m *MockExpenseRepository) GetByUserID(ctx context.Context, userID string) ([]*domain.Expense, error) {
	var result []*domain.Expense
	for _, exp := range m.expenses {
		if exp.UserID == userID {
			result = append(result, exp)
		}
	}
	return result, nil
}

func (m *MockExpenseRepository) GetByUserIDAndDateRange(ctx context.Context, userID string, from, to time.Time) ([]*domain.Expense, error) {
	var result []*domain.Expense
	for _, exp := range m.expenses {
		if exp.UserID == userID && !exp.ExpenseDate.Before(from) && !exp.ExpenseDate.After(to) {
			result = append(result, exp)
		}
	}
	return result, nil
}

func (m *MockExpenseRepository) GetByUserIDAndCategory(ctx context.Context, userID, categoryID string) ([]*domain.Expense, error) {
	var result []*domain.Expense
	for _, exp := range m.expenses {
		if exp.UserID == userID && exp.CategoryID != nil && *exp.CategoryID == categoryID {
			result = append(result, exp)
		}
	}
	return result, nil
}

func (m *MockExpenseRepository) Update(ctx context.Context, expense *domain.Expense) error {
	m.expenses[expense.ID] = expense
	return nil
}

func (m *MockExpenseRepository) Delete(ctx context.Context, id string) error {
	delete(m.expenses, id)
	return nil
}

// MockUserRepository for testing
type MockUserRepository struct {
	users map[string]*domain.User
}

func NewMockUserRepository() *MockUserRepository {
	return &MockUserRepository{
		users: make(map[string]*domain.User),
	}
}

func (m *MockUserRepository) Create(ctx context.Context, user *domain.User) error {
	m.users[user.UserID] = user
	return nil
}

func (m *MockUserRepository) GetByID(ctx context.Context, userID string) (*domain.User, error) {
	if user, ok := m.users[userID]; ok {
		return user, nil
	}
	return nil, errNotFound
}

func (m *MockUserRepository) Exists(ctx context.Context, userID string) (bool, error) {
	_, ok := m.users[userID]
	return ok, nil
}

// MockCategoryRepository for testing
type MockCategoryRepository struct {
	categories map[string]*domain.Category
}

func NewMockCategoryRepository() *MockCategoryRepository {
	return &MockCategoryRepository{
		categories: make(map[string]*domain.Category),
	}
}

func (m *MockCategoryRepository) Create(ctx context.Context, category *domain.Category) error {
	m.categories[category.ID] = category
	return nil
}

func (m *MockCategoryRepository) GetByID(ctx context.Context, id string) (*domain.Category, error) {
	if cat, ok := m.categories[id]; ok {
		return cat, nil
	}
	return nil, errNotFound
}

func (m *MockCategoryRepository) GetByUserID(ctx context.Context, userID string) ([]*domain.Category, error) {
	var result []*domain.Category
	for _, cat := range m.categories {
		if cat.UserID == userID {
			result = append(result, cat)
		}
	}
	return result, nil
}

func (m *MockCategoryRepository) GetByUserIDAndName(ctx context.Context, userID, name string) (*domain.Category, error) {
	for _, cat := range m.categories {
		if cat.UserID == userID && cat.Name == name {
			return cat, nil
		}
	}
	return nil, errNotFound
}

func (m *MockCategoryRepository) Update(ctx context.Context, category *domain.Category) error {
	m.categories[category.ID] = category
	return nil
}

func (m *MockCategoryRepository) Delete(ctx context.Context, id string) error {
	delete(m.categories, id)
	return nil
}

func (m *MockCategoryRepository) CreateKeyword(ctx context.Context, keyword *domain.CategoryKeyword) error {
	return nil
}

func (m *MockCategoryRepository) GetKeywordsByCategory(ctx context.Context, categoryID string) ([]*domain.CategoryKeyword, error) {
	return []*domain.CategoryKeyword{}, nil
}

func (m *MockCategoryRepository) DeleteKeyword(ctx context.Context, id string) error {
	return nil
}

// Helper to create valid Telegram webhook payload
func createTelegramWebhookPayload(userID, chatID int64, text string) []byte {
	update := map[string]interface{}{
		"update_id": 123456,
		"message": map[string]interface{}{
			"message_id": 789,
			"from": map[string]interface{}{
				"id":         userID,
				"is_bot":     false,
				"first_name": "Test",
				"username":   "testuser",
			},
			"chat": map[string]interface{}{
				"id":   chatID,
				"type": "private",
			},
			"date": int64(time.Now().Unix()),
			"text": text,
		},
	}
	body, _ := json.Marshal(update)
	return body
}

// TestTelegramHandlerValidMessage tests webhook with valid message
func TestTelegramHandlerValidMessage(t *testing.T) {
	// Setup
	expenseRepo := NewMockExpenseRepository()
	userRepo := NewMockUserRepository()
	categoryRepo := NewMockCategoryRepository()
	mockAI := setupMockAIService()

	// Create use cases
	autoSignupUC := usecase.NewAutoSignupUseCase(userRepo, categoryRepo)
	parseUC := usecase.NewParseConversationUseCase(mockAI)
	createExpenseUC := usecase.NewCreateExpenseUseCase(expenseRepo, categoryRepo, mockAI)

	telegramUseCase := NewTelegramUseCase(autoSignupUC, parseUC, createExpenseUC, nil)
	handler := NewHandler("test_bot_token", telegramUseCase)

	// Create valid webhook
	payload := createTelegramWebhookPayload(987654321, 123456, "早餐$20")

	req := httptest.NewRequest("POST", "/webhook/telegram", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	handler.HandleWebhook(w, req)

	// Verify response (Telegram always expects 200 OK)
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	// Verify user was created with telegram_ prefix
	userExists, _ := userRepo.Exists(context.Background(), "telegram_987654321")
	if !userExists {
		t.Error("Expected user to be created with telegram_ prefix")
	}
}

// TestTelegramHandlerMalformedJSON tests webhook with invalid JSON
func TestTelegramHandlerMalformedJSON(t *testing.T) {
	// Setup
	expenseRepo := NewMockExpenseRepository()
	userRepo := NewMockUserRepository()
	categoryRepo := NewMockCategoryRepository()
	mockAI := setupMockAIService()

	autoSignupUC := usecase.NewAutoSignupUseCase(userRepo, categoryRepo)
	parseUC := usecase.NewParseConversationUseCase(mockAI)
	createExpenseUC := usecase.NewCreateExpenseUseCase(expenseRepo, categoryRepo, mockAI)

	telegramUseCase := NewTelegramUseCase(autoSignupUC, parseUC, createExpenseUC, nil)
	handler := NewHandler("test_bot_token", telegramUseCase)

	// Create malformed JSON
	malformedPayload := []byte(`{"update_id": 123, "message": invalid json}`)

	req := httptest.NewRequest("POST", "/webhook/telegram", bytes.NewReader(malformedPayload))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	handler.HandleWebhook(w, req)

	// Verify error response (Telegram expects 400 for bad request)
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

// TestTelegramHandlerEmptyMessage tests webhook with empty message
func TestTelegramHandlerEmptyMessage(t *testing.T) {
	// Setup
	expenseRepo := NewMockExpenseRepository()
	userRepo := NewMockUserRepository()
	categoryRepo := NewMockCategoryRepository()
	mockAI := setupMockAIService()

	autoSignupUC := usecase.NewAutoSignupUseCase(userRepo, categoryRepo)
	parseUC := usecase.NewParseConversationUseCase(mockAI)
	createExpenseUC := usecase.NewCreateExpenseUseCase(expenseRepo, categoryRepo, mockAI)

	telegramUseCase := NewTelegramUseCase(autoSignupUC, parseUC, createExpenseUC, nil)
	handler := NewHandler("test_bot_token", telegramUseCase)

	// Create webhook with empty message
	payload := createTelegramWebhookPayload(987654321, 123456, "")

	req := httptest.NewRequest("POST", "/webhook/telegram", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	handler.HandleWebhook(w, req)

	// Verify success (empty messages are ignored)
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	// Verify user was NOT created (no message means no processing)
	userExists, _ := userRepo.Exists(context.Background(), "telegram_987654321")
	if userExists {
		t.Error("Expected user to NOT be created for empty message")
	}
}

// TestTelegramHandlerNoMessage tests webhook with no message field
func TestTelegramHandlerNoMessage(t *testing.T) {
	// Setup
	expenseRepo := NewMockExpenseRepository()
	userRepo := NewMockUserRepository()
	categoryRepo := NewMockCategoryRepository()
	mockAI := setupMockAIService()

	autoSignupUC := usecase.NewAutoSignupUseCase(userRepo, categoryRepo)
	parseUC := usecase.NewParseConversationUseCase(mockAI)
	createExpenseUC := usecase.NewCreateExpenseUseCase(expenseRepo, categoryRepo, mockAI)

	telegramUseCase := NewTelegramUseCase(autoSignupUC, parseUC, createExpenseUC, nil)
	handler := NewHandler("test_bot_token", telegramUseCase)

	// Create update with no message
	update := map[string]interface{}{
		"update_id": 123456,
		// No message field
	}
	body, _ := json.Marshal(update)

	req := httptest.NewRequest("POST", "/webhook/telegram", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	handler.HandleWebhook(w, req)

	// Verify success (non-message updates are ignored)
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

// TestTelegramHandlerMissingUserInfo tests webhook with missing user info
func TestTelegramHandlerMissingUserInfo(t *testing.T) {
	// Setup
	expenseRepo := NewMockExpenseRepository()
	userRepo := NewMockUserRepository()
	categoryRepo := NewMockCategoryRepository()
	mockAI := setupMockAIService()

	autoSignupUC := usecase.NewAutoSignupUseCase(userRepo, categoryRepo)
	parseUC := usecase.NewParseConversationUseCase(mockAI)
	createExpenseUC := usecase.NewCreateExpenseUseCase(expenseRepo, categoryRepo, mockAI)

	telegramUseCase := NewTelegramUseCase(autoSignupUC, parseUC, createExpenseUC, nil)
	handler := NewHandler("test_bot_token", telegramUseCase)

	// Create update with missing from field
	update := map[string]interface{}{
		"update_id": 123456,
		"message": map[string]interface{}{
			"message_id": 789,
			// No 'from' field
			"chat": map[string]interface{}{
				"id":   int64(123456),
				"type": "private",
			},
			"date": int64(time.Now().Unix()),
			"text": "早餐$20",
		},
	}
	body, _ := json.Marshal(update)

	req := httptest.NewRequest("POST", "/webhook/telegram", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	handler.HandleWebhook(w, req)

	// Verify success (message is ignored but no error)
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

// TestTelegramHandlerMultipleExpenses tests parsing multiple expenses
func TestTelegramHandlerMultipleExpenses(t *testing.T) {
	// Setup
	expenseRepo := NewMockExpenseRepository()
	userRepo := NewMockUserRepository()
	categoryRepo := NewMockCategoryRepository()
	mockAI := setupMockAIService()

	autoSignupUC := usecase.NewAutoSignupUseCase(userRepo, categoryRepo)
	parseUC := usecase.NewParseConversationUseCase(mockAI)
	createExpenseUC := usecase.NewCreateExpenseUseCase(expenseRepo, categoryRepo, mockAI)

	telegramUseCase := NewTelegramUseCase(autoSignupUC, parseUC, createExpenseUC, nil)
	handler := NewHandler("test_bot_token", telegramUseCase)

	// Create webhook with multiple expenses
	payload := createTelegramWebhookPayload(987654321, 123456, "早餐$20午餐$30晚餐$50")

	req := httptest.NewRequest("POST", "/webhook/telegram", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	handler.HandleWebhook(w, req)

	// Verify success
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	// Verify user was created
	userExists, _ := userRepo.Exists(context.Background(), "telegram_987654321")
	if !userExists {
		t.Error("Expected user to be created")
	}
}

// TestTelegramHandlerDifferentChatTypes tests different Telegram chat types
func TestTelegramHandlerDifferentChatTypes(t *testing.T) {
	chatTypes := []string{"private", "group", "supergroup", "channel"}

	for _, chatType := range chatTypes {
		t.Run(chatType, func(t *testing.T) {
			// Setup
			expenseRepo := NewMockExpenseRepository()
			userRepo := NewMockUserRepository()
			categoryRepo := NewMockCategoryRepository()
			mockAI := setupMockAIService()

			autoSignupUC := usecase.NewAutoSignupUseCase(userRepo, categoryRepo)
			parseUC := usecase.NewParseConversationUseCase(mockAI)
			createExpenseUC := usecase.NewCreateExpenseUseCase(expenseRepo, categoryRepo, mockAI)

			telegramUseCase := NewTelegramUseCase(autoSignupUC, parseUC, createExpenseUC, nil)
			handler := NewHandler("test_bot_token", telegramUseCase)

			// Create update with specific chat type
			update := map[string]interface{}{
				"update_id": 123456,
				"message": map[string]interface{}{
					"message_id": 789,
					"from": map[string]interface{}{
						"id":         int64(987654321),
						"is_bot":     false,
						"first_name": "Test",
					},
					"chat": map[string]interface{}{
						"id":   int64(123456),
						"type": chatType,
					},
					"date": int64(time.Now().Unix()),
					"text": "早餐$20",
				},
			}
			body, _ := json.Marshal(update)

			req := httptest.NewRequest("POST", "/webhook/telegram", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			handler.HandleWebhook(w, req)

			// Verify success for all chat types
			if w.Code != http.StatusOK {
				t.Errorf("Expected status %d for chat type %s, got %d", http.StatusOK, chatType, w.Code)
			}
		})
	}
}

// Helper to setup mock AI service
func setupMockAIService() domain.AIService {
	return &mockAIService{}
}

// mockAIService implements domain.AIService for testing
type mockAIService struct{}

func (m *mockAIService) ParseExpense(ctx context.Context, text string, userID string) ([]*domain.ParsedExpense, error) {
	// Simple mock that returns a basic expense
	return []*domain.ParsedExpense{
		{
			Amount:      20.0,
			Description: "Test expense",
		},
	}, nil
}

func (m *mockAIService) SuggestCategory(ctx context.Context, description string, userID string) (string, error) {
	return "Food", nil
}
