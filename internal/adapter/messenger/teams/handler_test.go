package teams

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

// Mock repositories (shared for all handler tests)
type MockExpenseRepository struct {
	expenses map[string]*domain.Expense
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

type MockUserRepository struct {
	users map[string]*domain.User
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

type MockCategoryRepository struct {
	categories map[string]*domain.Category
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

// Helper to create Teams activity payload
func createTeamsActivityPayload(userID, text string, activityType string) ([]byte, string) {
	activity := map[string]interface{}{
		"type":        activityType,
		"id":          "activity_123",
		"timestamp":   time.Now().Format(time.RFC3339),
		"serviceUrl":  "https://smba.trafficmanager.net/teams/",
		"channelId":   "personal",
		"from": map[string]interface{}{
			"id":   userID,
			"name": "Test User",
		},
		"conversation": map[string]interface{}{
			"conversationType": "personal",
			"id":               "conv_123",
			"isGroup":          false,
		},
		"text": text,
	}

	body, _ := json.Marshal(activity)

	// Compute Teams signature
	mac := hmac.New(sha256.New, []byte("test_app_password"))
	mac.Write(body)
	signature := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	return body, signature
}

// TestTeamsHandlerValidSignature tests message with valid signature
func TestTeamsHandlerValidSignature(t *testing.T) {
	expenseRepo := &MockExpenseRepository{expenses: make(map[string]*domain.Expense)}
	userRepo := &MockUserRepository{users: make(map[string]*domain.User)}
	categoryRepo := &MockCategoryRepository{categories: make(map[string]*domain.Category)}
	mockAI := &mockAIService{}

	autoSignupUC := usecase.NewAutoSignupUseCase(userRepo, categoryRepo)
	parseUC := usecase.NewParseConversationUseCase(mockAI)
	createExpenseUC := usecase.NewCreateExpenseUseCase(expenseRepo, categoryRepo, mockAI)

	teamsUseCase := NewTeamsUseCase(autoSignupUC, parseUC, createExpenseUC, nil)
	handler := NewHandler("test_app_id", "test_app_password", teamsUseCase)

	payload, signature := createTeamsActivityPayload("teams_U123456", "早餐$20", "message")

	req := httptest.NewRequest("POST", "/webhook/teams", bytes.NewReader(payload))
	req.Header.Set("Authorization", "Bearer "+signature)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	handler.HandleWebhook(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected %d, got %d", http.StatusOK, w.Code)
	}
}

// TestTeamsHandlerInvalidSignature tests rejection with invalid signature
func TestTeamsHandlerInvalidSignature(t *testing.T) {
	expenseRepo := &MockExpenseRepository{expenses: make(map[string]*domain.Expense)}
	userRepo := &MockUserRepository{users: make(map[string]*domain.User)}
	categoryRepo := &MockCategoryRepository{categories: make(map[string]*domain.Category)}
	mockAI := &mockAIService{}

	autoSignupUC := usecase.NewAutoSignupUseCase(userRepo, categoryRepo)
	parseUC := usecase.NewParseConversationUseCase(mockAI)
	createExpenseUC := usecase.NewCreateExpenseUseCase(expenseRepo, categoryRepo, mockAI)

	teamsUseCase := NewTeamsUseCase(autoSignupUC, parseUC, createExpenseUC, nil)
	handler := NewHandler("test_app_id", "test_app_password", teamsUseCase)

	payload, _ := createTeamsActivityPayload("teams_U123456", "早餐$20", "message")

	req := httptest.NewRequest("POST", "/webhook/teams", bytes.NewReader(payload))
	req.Header.Set("Authorization", "Bearer invalid_signature")
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	handler.HandleWebhook(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

// TestTeamsHandlerMissingSignature tests rejection without Authorization header
func TestTeamsHandlerMissingSignature(t *testing.T) {
	expenseRepo := &MockExpenseRepository{expenses: make(map[string]*domain.Expense)}
	userRepo := &MockUserRepository{users: make(map[string]*domain.User)}
	categoryRepo := &MockCategoryRepository{categories: make(map[string]*domain.Category)}
	mockAI := &mockAIService{}

	autoSignupUC := usecase.NewAutoSignupUseCase(userRepo, categoryRepo)
	parseUC := usecase.NewParseConversationUseCase(mockAI)
	createExpenseUC := usecase.NewCreateExpenseUseCase(expenseRepo, categoryRepo, mockAI)

	teamsUseCase := NewTeamsUseCase(autoSignupUC, parseUC, createExpenseUC, nil)
	handler := NewHandler("test_app_id", "test_app_password", teamsUseCase)

	payload, _ := createTeamsActivityPayload("teams_U123456", "早餐$20", "message")

	req := httptest.NewRequest("POST", "/webhook/teams", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	handler.HandleWebhook(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

// TestTeamsHandlerDirectMessage tests direct message handling
func TestTeamsHandlerDirectMessage(t *testing.T) {
	expenseRepo := &MockExpenseRepository{expenses: make(map[string]*domain.Expense)}
	userRepo := &MockUserRepository{users: make(map[string]*domain.User)}
	categoryRepo := &MockCategoryRepository{categories: make(map[string]*domain.Category)}
	mockAI := &mockAIService{}

	autoSignupUC := usecase.NewAutoSignupUseCase(userRepo, categoryRepo)
	parseUC := usecase.NewParseConversationUseCase(mockAI)
	createExpenseUC := usecase.NewCreateExpenseUseCase(expenseRepo, categoryRepo, mockAI)

	teamsUseCase := NewTeamsUseCase(autoSignupUC, parseUC, createExpenseUC, nil)
	handler := NewHandler("test_app_id", "test_app_password", teamsUseCase)

	payload, signature := createTeamsActivityPayload("teams_U123456", "早餐$20", "message")

	req := httptest.NewRequest("POST", "/webhook/teams", bytes.NewReader(payload))
	req.Header.Set("Authorization", "Bearer "+signature)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	handler.HandleWebhook(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected %d, got %d", http.StatusOK, w.Code)
	}
}

// TestTeamsHandlerConversationUpdate tests conversationUpdate event
func TestTeamsHandlerConversationUpdate(t *testing.T) {
	expenseRepo := &MockExpenseRepository{expenses: make(map[string]*domain.Expense)}
	userRepo := &MockUserRepository{users: make(map[string]*domain.User)}
	categoryRepo := &MockCategoryRepository{categories: make(map[string]*domain.Category)}
	mockAI := &mockAIService{}

	autoSignupUC := usecase.NewAutoSignupUseCase(userRepo, categoryRepo)
	parseUC := usecase.NewParseConversationUseCase(mockAI)
	createExpenseUC := usecase.NewCreateExpenseUseCase(expenseRepo, categoryRepo, mockAI)

	teamsUseCase := NewTeamsUseCase(autoSignupUC, parseUC, createExpenseUC, nil)
	handler := NewHandler("test_app_id", "test_app_password", teamsUseCase)

	payload, signature := createTeamsActivityPayload("teams_U123456", "", "conversationUpdate")

	req := httptest.NewRequest("POST", "/webhook/teams", bytes.NewReader(payload))
	req.Header.Set("Authorization", "Bearer "+signature)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	handler.HandleWebhook(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected %d, got %d", http.StatusOK, w.Code)
	}
}

// Mock AI Service
type mockAIService struct{}

func (m *mockAIService) ParseExpense(ctx context.Context, text string, userID string) ([]*domain.ParsedExpense, error) {
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
