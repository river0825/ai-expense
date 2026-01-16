package line

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
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


// Helper to create valid LINE webhook payload
func createLineWebhookPayload(userID, text string) ([]byte, string) {
	payload := map[string]interface{}{
		"events": []map[string]interface{}{
			{
				"type": "message",
				"source": map[string]string{
					"type":   "user",
					"userId": userID,
				},
				"message": map[string]string{
					"type": "text",
					"text": text,
				},
				"replyToken": "test_reply_token_123",
			},
		},
	}
	body, _ := json.Marshal(payload)

	// Compute LINE signature
	hash := hmac.New(sha256.New, []byte("test_channel_secret"))
	hash.Write(body)
	signature := base64.StdEncoding.EncodeToString(hash.Sum(nil))

	return body, signature
}

// TestLineHandlerValidSignature tests webhook with correct signature
func TestLineHandlerValidSignature(t *testing.T) {
	// Setup
	expenseRepo := NewMockExpenseRepository()
	userRepo := NewMockUserRepository()
	categoryRepo := NewMockCategoryRepository()
	mockAI := setupMockAIService()

	// Create use cases
	autoSignupUC := usecase.NewAutoSignupUseCase(userRepo, categoryRepo)
	parseUC := usecase.NewParseConversationUseCase(mockAI)
	createExpenseUC := usecase.NewCreateExpenseUseCase(expenseRepo, categoryRepo, mockAI)

	lineUseCase := NewLineUseCase(autoSignupUC, parseUC, createExpenseUC, nil)
	handler := NewHandler("test_channel_secret", lineUseCase)

	// Create valid webhook
	payload, signature := createLineWebhookPayload("line_test_user", "早餐$20")

	req := httptest.NewRequest("POST", "/webhook/line", bytes.NewReader(payload))
	req.Header.Set("X-Line-Signature", signature)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	handler.HandleWebhook(w, req)

	// Verify response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	// Verify user was created
	userExists, _ := userRepo.Exists(context.Background(), "line_test_user")
	if !userExists {
		t.Error("Expected user to be created")
	}
}

// TestLineHandlerInvalidSignature tests webhook rejection with invalid signature
func TestLineHandlerInvalidSignature(t *testing.T) {
	// Setup
	expenseRepo := NewMockExpenseRepository()
	userRepo := NewMockUserRepository()
	categoryRepo := NewMockCategoryRepository()
	mockAI := setupMockAIService()

	autoSignupUC := usecase.NewAutoSignupUseCase(userRepo, categoryRepo)
	parseUC := usecase.NewParseConversationUseCase(mockAI)
	createExpenseUC := usecase.NewCreateExpenseUseCase(expenseRepo, categoryRepo, mockAI)

	lineUseCase := NewLineUseCase(autoSignupUC, parseUC, createExpenseUC, nil)
	handler := NewHandler("test_channel_secret", lineUseCase)

	// Create webhook with invalid signature
	payload, _ := createLineWebhookPayload("line_test_user", "早餐$20")

	req := httptest.NewRequest("POST", "/webhook/line", bytes.NewReader(payload))
	req.Header.Set("X-Line-Signature", "invalid_signature_value")
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	handler.HandleWebhook(w, req)

	// Verify rejection
	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}

	// Verify user was NOT created
	userExists, _ := userRepo.Exists(context.Background(), "line_test_user")
	if userExists {
		t.Error("Expected user to NOT be created with invalid signature")
	}
}

// TestLineHandlerMalformedJSON tests webhook with invalid JSON
func TestLineHandlerMalformedJSON(t *testing.T) {
	// Setup
	expenseRepo := NewMockExpenseRepository()
	userRepo := NewMockUserRepository()
	categoryRepo := NewMockCategoryRepository()
	mockAI := setupMockAIService()

	autoSignupUC := usecase.NewAutoSignupUseCase(userRepo, categoryRepo)
	parseUC := usecase.NewParseConversationUseCase(mockAI)
	createExpenseUC := usecase.NewCreateExpenseUseCase(expenseRepo, categoryRepo, mockAI)

	lineUseCase := NewLineUseCase(autoSignupUC, parseUC, createExpenseUC, nil)
	handler := NewHandler("test_channel_secret", lineUseCase)

	// Create malformed JSON
	malformedPayload := []byte(`{"events": [invalid json}`)

	// Compute signature for malformed payload
	hash := hmac.New(sha256.New, []byte("test_channel_secret"))
	hash.Write(malformedPayload)
	signature := base64.StdEncoding.EncodeToString(hash.Sum(nil))

	req := httptest.NewRequest("POST", "/webhook/line", bytes.NewReader(malformedPayload))
	req.Header.Set("X-Line-Signature", signature)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	handler.HandleWebhook(w, req)

	// Verify error response
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

// TestLineHandlerEmptyMessage tests webhook with empty message
func TestLineHandlerEmptyMessage(t *testing.T) {
	// Setup
	expenseRepo := NewMockExpenseRepository()
	userRepo := NewMockUserRepository()
	categoryRepo := NewMockCategoryRepository()
	mockAI := setupMockAIService()

	autoSignupUC := usecase.NewAutoSignupUseCase(userRepo, categoryRepo)
	parseUC := usecase.NewParseConversationUseCase(mockAI)
	createExpenseUC := usecase.NewCreateExpenseUseCase(expenseRepo, categoryRepo, mockAI)

	lineUseCase := NewLineUseCase(autoSignupUC, parseUC, createExpenseUC, nil)
	handler := NewHandler("test_channel_secret", lineUseCase)

	// Create webhook with empty message
	payload, signature := createLineWebhookPayload("line_test_user", "")

	req := httptest.NewRequest("POST", "/webhook/line", bytes.NewReader(payload))
	req.Header.Set("X-Line-Signature", signature)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	handler.HandleWebhook(w, req)

	// Verify it doesn't crash (should return 200 but not create expense)
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	// Verify user was created but no expense
	userExists, _ := userRepo.Exists(context.Background(), "line_test_user")
	if !userExists {
		t.Error("Expected user to be created")
	}

	expenses, _ := expenseRepo.GetByUserID(context.Background(), "line_test_user")
	if len(expenses) > 0 {
		t.Errorf("Expected no expenses for empty message, got %d", len(expenses))
	}
}

// TestLineHandlerNonMessageEvent tests webhook with non-message event
func TestLineHandlerNonMessageEvent(t *testing.T) {
	// Setup
	expenseRepo := NewMockExpenseRepository()
	userRepo := NewMockUserRepository()
	categoryRepo := NewMockCategoryRepository()
	mockAI := setupMockAIService()

	autoSignupUC := usecase.NewAutoSignupUseCase(userRepo, categoryRepo)
	parseUC := usecase.NewParseConversationUseCase(mockAI)
	createExpenseUC := usecase.NewCreateExpenseUseCase(expenseRepo, categoryRepo, mockAI)

	lineUseCase := NewLineUseCase(autoSignupUC, parseUC, createExpenseUC, nil)
	handler := NewHandler("test_channel_secret", lineUseCase)

	// Create webhook with non-message event
	payload := map[string]interface{}{
		"events": []map[string]interface{}{
			{
				"type": "join",
				"source": map[string]string{
					"type": "group",
					"groupId": "test_group",
				},
			},
		},
	}
	body, _ := json.Marshal(payload)

	hash := hmac.New(sha256.New, []byte("test_channel_secret"))
	hash.Write(body)
	signature := base64.StdEncoding.EncodeToString(hash.Sum(nil))

	req := httptest.NewRequest("POST", "/webhook/line", bytes.NewReader(body))
	req.Header.Set("X-Line-Signature", signature)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	handler.HandleWebhook(w, req)

	// Verify success (join events are ignored but handler returns 200)
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

// TestLineHandlerMultipleExpenses tests parsing multiple expenses in one message
func TestLineHandlerMultipleExpenses(t *testing.T) {
	// Setup
	expenseRepo := NewMockExpenseRepository()
	userRepo := NewMockUserRepository()
	categoryRepo := NewMockCategoryRepository()
	mockAI := setupMockAIService()

	autoSignupUC := usecase.NewAutoSignupUseCase(userRepo, categoryRepo)
	parseUC := usecase.NewParseConversationUseCase(mockAI)
	createExpenseUC := usecase.NewCreateExpenseUseCase(expenseRepo, categoryRepo, mockAI)

	lineUseCase := NewLineUseCase(autoSignupUC, parseUC, createExpenseUC, nil)
	handler := NewHandler("test_channel_secret", lineUseCase)

	// Create webhook with multiple expenses
	payload, signature := createLineWebhookPayload("line_test_user", "早餐$20午餐$30晚餐$50")

	req := httptest.NewRequest("POST", "/webhook/line", bytes.NewReader(payload))
	req.Header.Set("X-Line-Signature", signature)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	handler.HandleWebhook(w, req)

	// Verify success
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	// Verify expenses were created
	expenses, _ := expenseRepo.GetByUserID(context.Background(), "line_test_user")
	if len(expenses) < 1 {
		t.Errorf("Expected at least 1 expense, got %d", len(expenses))
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

func (m *mockAIService) SuggestCategory(ctx context.Context, description string) (string, error) {
	return "Food", nil
}
