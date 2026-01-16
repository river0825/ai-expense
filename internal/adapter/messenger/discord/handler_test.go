package discord

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

// Mock repositories for Discord tests
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

// Helper to create Discord PING interaction
func createDiscordPingInteraction() []byte {
	interaction := map[string]interface{}{
		"type": 1, // PING
		"id":   "ping_123",
		"token": "ping_token",
	}
	body, _ := json.Marshal(interaction)
	return body
}

// Helper to create Discord message interaction
func createDiscordMessageInteraction(userID, content string) []byte {
	interaction := map[string]interface{}{
		"type": 2, // APPLICATION_COMMAND
		"id":   "interaction_123",
		"token": "response_token",
		"user": map[string]string{
			"id":       userID,
			"username": "testuser",
		},
		"data": map[string]string{
			"content": content,
		},
	}
	body, _ := json.Marshal(interaction)
	return body
}

// Helper to create Discord interaction with guild member
func createDiscordGuildInteraction(userID, content string) []byte {
	interaction := map[string]interface{}{
		"type": 2,
		"id":   "interaction_456",
		"token": "response_token",
		"member": map[string]interface{}{
			"user": map[string]string{
				"id":       userID,
				"username": "testuser",
			},
		},
		"data": map[string]string{
			"content": content,
		},
	}
	body, _ := json.Marshal(interaction)
	return body
}

// TestDiscordPingInteraction tests PING/PONG webhook validation
func TestDiscordPingInteraction(t *testing.T) {
	expenseRepo := &MockExpenseRepository{expenses: make(map[string]*domain.Expense)}
	userRepo := &MockUserRepository{users: make(map[string]*domain.User)}
	categoryRepo := &MockCategoryRepository{categories: make(map[string]*domain.Category)}
	mockAI := &mockAIService{}

	autoSignupUC := usecase.NewAutoSignupUseCase(userRepo, categoryRepo)
	parseUC := usecase.NewParseConversationUseCase(mockAI)
	createExpenseUC := usecase.NewCreateExpenseUseCase(expenseRepo, categoryRepo, mockAI)

	discordUseCase := NewDiscordUseCase(autoSignupUC, parseUC, createExpenseUC, nil)
	handler := NewHandler("test_bot_token", discordUseCase)

	payload := createDiscordPingInteraction()

	req := httptest.NewRequest("POST", "/webhook/discord", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	handler.HandleWebhook(w, req)

	// Verify PONG response
	if w.Code != http.StatusOK {
		t.Errorf("Expected %d, got %d", http.StatusOK, w.Code)
	}

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	if responseType, ok := response["type"].(float64); !ok || int(responseType) != 1 {
		t.Error("Expected PONG response type 1")
	}
}

// TestDiscordMessageInteraction tests message command interaction
func TestDiscordMessageInteraction(t *testing.T) {
	expenseRepo := &MockExpenseRepository{expenses: make(map[string]*domain.Expense)}
	userRepo := &MockUserRepository{users: make(map[string]*domain.User)}
	categoryRepo := &MockCategoryRepository{categories: make(map[string]*domain.Category)}
	mockAI := &mockAIService{}

	autoSignupUC := usecase.NewAutoSignupUseCase(userRepo, categoryRepo)
	parseUC := usecase.NewParseConversationUseCase(mockAI)
	createExpenseUC := usecase.NewCreateExpenseUseCase(expenseRepo, categoryRepo, mockAI)

	discordUseCase := NewDiscordUseCase(autoSignupUC, parseUC, createExpenseUC, nil)
	handler := NewHandler("test_bot_token", discordUseCase)

	payload := createDiscordMessageInteraction("discord_987654321", "早餐$20")

	req := httptest.NewRequest("POST", "/webhook/discord", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	handler.HandleWebhook(w, req)

	// Verify deferred response
	if w.Code != http.StatusOK {
		t.Errorf("Expected %d, got %d", http.StatusOK, w.Code)
	}

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	if responseType, ok := response["type"].(float64); !ok || int(responseType) != 5 {
		t.Error("Expected deferred response type 5")
	}
}

// TestDiscordEmptyMessage tests empty message handling
func TestDiscordEmptyMessage(t *testing.T) {
	expenseRepo := &MockExpenseRepository{expenses: make(map[string]*domain.Expense)}
	userRepo := &MockUserRepository{users: make(map[string]*domain.User)}
	categoryRepo := &MockCategoryRepository{categories: make(map[string]*domain.Category)}
	mockAI := &mockAIService{}

	autoSignupUC := usecase.NewAutoSignupUseCase(userRepo, categoryRepo)
	parseUC := usecase.NewParseConversationUseCase(mockAI)
	createExpenseUC := usecase.NewCreateExpenseUseCase(expenseRepo, categoryRepo, mockAI)

	discordUseCase := NewDiscordUseCase(autoSignupUC, parseUC, createExpenseUC, nil)
	handler := NewHandler("test_bot_token", discordUseCase)

	payload := createDiscordMessageInteraction("discord_987654321", "")

	req := httptest.NewRequest("POST", "/webhook/discord", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	handler.HandleWebhook(w, req)

	// Verify error response for empty message
	if w.Code != http.StatusOK {
		t.Errorf("Expected %d, got %d", http.StatusOK, w.Code)
	}

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	if responseType, ok := response["type"].(float64); !ok || int(responseType) != 4 {
		t.Error("Expected immediate error response type 4")
	}
}

// TestDiscordGuildMemberInteraction tests guild member interaction
func TestDiscordGuildMemberInteraction(t *testing.T) {
	expenseRepo := &MockExpenseRepository{expenses: make(map[string]*domain.Expense)}
	userRepo := &MockUserRepository{users: make(map[string]*domain.User)}
	categoryRepo := &MockCategoryRepository{categories: make(map[string]*domain.Category)}
	mockAI := &mockAIService{}

	autoSignupUC := usecase.NewAutoSignupUseCase(userRepo, categoryRepo)
	parseUC := usecase.NewParseConversationUseCase(mockAI)
	createExpenseUC := usecase.NewCreateExpenseUseCase(expenseRepo, categoryRepo, mockAI)

	discordUseCase := NewDiscordUseCase(autoSignupUC, parseUC, createExpenseUC, nil)
	handler := NewHandler("test_bot_token", discordUseCase)

	payload := createDiscordGuildInteraction("discord_112233445", "午餐$30")

	req := httptest.NewRequest("POST", "/webhook/discord", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	handler.HandleWebhook(w, req)

	// Verify success
	if w.Code != http.StatusOK {
		t.Errorf("Expected %d, got %d", http.StatusOK, w.Code)
	}
}

// TestDiscordInvalidMethod tests non-POST requests
func TestDiscordInvalidMethod(t *testing.T) {
	expenseRepo := &MockExpenseRepository{expenses: make(map[string]*domain.Expense)}
	userRepo := &MockUserRepository{users: make(map[string]*domain.User)}
	categoryRepo := &MockCategoryRepository{categories: make(map[string]*domain.Category)}
	mockAI := &mockAIService{}

	autoSignupUC := usecase.NewAutoSignupUseCase(userRepo, categoryRepo)
	parseUC := usecase.NewParseConversationUseCase(mockAI)
	createExpenseUC := usecase.NewCreateExpenseUseCase(expenseRepo, categoryRepo, mockAI)

	discordUseCase := NewDiscordUseCase(autoSignupUC, parseUC, createExpenseUC, nil)
	handler := NewHandler("test_bot_token", discordUseCase)

	payload := createDiscordPingInteraction()

	req := httptest.NewRequest("GET", "/webhook/discord", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	handler.HandleWebhook(w, req)

	// Verify method not allowed
	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected %d, got %d", http.StatusMethodNotAllowed, w.Code)
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
