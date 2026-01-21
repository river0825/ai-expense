package whatsapp

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
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

// Mock repositories
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

// Helper to create WhatsApp message payload
func createWhatsAppMessagePayload(phoneNumber, text string) ([]byte, string) {
	webhook := map[string]interface{}{
		"object": "whatsapp_business_account",
		"entry": []map[string]interface{}{
			{
				"id": "entry123",
				"changes": []map[string]interface{}{
					{
						"value": map[string]interface{}{
							"messaging_product": "whatsapp",
							"metadata": map[string]string{
								"display_phone_number": "16505551234",
								"phone_number_id":      "123456789",
							},
							"messages": []map[string]interface{}{
								{
									"from":      phoneNumber,
									"id":        "msg_123",
									"timestamp": "1671497741",
									"type":      "text",
									"text": map[string]string{
										"body": text,
									},
								},
							},
						},
						"field": "messages",
					},
				},
			},
		},
	}

	body, _ := json.Marshal(webhook)
	payload := string(body)

	// Compute WhatsApp signature
	mac := hmac.New(sha256.New, []byte("test_app_secret"))
	mac.Write([]byte(payload))
	signature := hex.EncodeToString(mac.Sum(nil))

	return body, signature
}

// TestWhatsAppWebhookVerification tests GET verification request
func TestWhatsAppWebhookVerification(t *testing.T) {
	expenseRepo := &MockExpenseRepository{expenses: make(map[string]*domain.Expense)}
	userRepo := &MockUserRepository{users: make(map[string]*domain.User)}
	categoryRepo := &MockCategoryRepository{categories: make(map[string]*domain.Category)}
	mockAI := &mockAIService{}

	autoSignupUC := usecase.NewAutoSignupUseCase(userRepo, categoryRepo)
	parseUC := usecase.NewParseConversationUseCase(mockAI)
	createExpenseUC := usecase.NewCreateExpenseUseCase(expenseRepo, categoryRepo, mockAI)

	whatsappUseCase := NewWhatsAppUseCase(autoSignupUC, parseUC, createExpenseUC, nil)
	handler := NewHandler("test_app_secret", "123456789", whatsappUseCase)

	// GET request for webhook verification
	req := httptest.NewRequest("GET", "/webhook/whatsapp?hub.mode=subscribe&hub.verify_token=verify_token&hub.challenge=test_challenge_123", nil)

	w := httptest.NewRecorder()
	handler.HandleWebhook(w, req)

	// Verify challenge response
	if w.Code != http.StatusOK {
		t.Errorf("Expected %d, got %d", http.StatusOK, w.Code)
	}

	if w.Body.String() != "test_challenge_123" {
		t.Errorf("Expected challenge in response, got: %s", w.Body.String())
	}
}

// TestWhatsAppMessageWithValidSignature tests message with valid signature
func TestWhatsAppMessageWithValidSignature(t *testing.T) {
	expenseRepo := &MockExpenseRepository{expenses: make(map[string]*domain.Expense)}
	userRepo := &MockUserRepository{users: make(map[string]*domain.User)}
	categoryRepo := &MockCategoryRepository{categories: make(map[string]*domain.Category)}
	mockAI := &mockAIService{}

	autoSignupUC := usecase.NewAutoSignupUseCase(userRepo, categoryRepo)
	parseUC := usecase.NewParseConversationUseCase(mockAI)
	createExpenseUC := usecase.NewCreateExpenseUseCase(expenseRepo, categoryRepo, mockAI)

	whatsappUseCase := NewWhatsAppUseCase(autoSignupUC, parseUC, createExpenseUC, nil)
	handler := NewHandler("test_app_secret", "123456789", whatsappUseCase)

	// POST message with valid signature
	payload, signature := createWhatsAppMessagePayload("16505551234", "早餐$20")

	req := httptest.NewRequest("POST", "/webhook/whatsapp", bytes.NewReader(payload))
	req.Header.Set("X-Hub-Signature-256", "sha256="+signature)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	handler.HandleWebhook(w, req)

	// Verify success
	if w.Code != http.StatusOK {
		t.Errorf("Expected %d, got %d", http.StatusOK, w.Code)
	}
}

// TestWhatsAppMessageWithInvalidSignature tests rejection of invalid signature
func TestWhatsAppMessageWithInvalidSignature(t *testing.T) {
	expenseRepo := &MockExpenseRepository{expenses: make(map[string]*domain.Expense)}
	userRepo := &MockUserRepository{users: make(map[string]*domain.User)}
	categoryRepo := &MockCategoryRepository{categories: make(map[string]*domain.Category)}
	mockAI := &mockAIService{}

	autoSignupUC := usecase.NewAutoSignupUseCase(userRepo, categoryRepo)
	parseUC := usecase.NewParseConversationUseCase(mockAI)
	createExpenseUC := usecase.NewCreateExpenseUseCase(expenseRepo, categoryRepo, mockAI)

	whatsappUseCase := NewWhatsAppUseCase(autoSignupUC, parseUC, createExpenseUC, nil)
	handler := NewHandler("test_app_secret", "123456789", whatsappUseCase)

	payload, _ := createWhatsAppMessagePayload("16505551234", "早餐$20")

	req := httptest.NewRequest("POST", "/webhook/whatsapp", bytes.NewReader(payload))
	req.Header.Set("X-Hub-Signature", "sha256=invalid_signature")
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	handler.HandleWebhook(w, req)

	// Verify rejection
	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

// TestWhatsAppMissingSignature tests rejection without signature
func TestWhatsAppMissingSignature(t *testing.T) {
	expenseRepo := &MockExpenseRepository{expenses: make(map[string]*domain.Expense)}
	userRepo := &MockUserRepository{users: make(map[string]*domain.User)}
	categoryRepo := &MockCategoryRepository{categories: make(map[string]*domain.Category)}
	mockAI := &mockAIService{}

	autoSignupUC := usecase.NewAutoSignupUseCase(userRepo, categoryRepo)
	parseUC := usecase.NewParseConversationUseCase(mockAI)
	createExpenseUC := usecase.NewCreateExpenseUseCase(expenseRepo, categoryRepo, mockAI)

	whatsappUseCase := NewWhatsAppUseCase(autoSignupUC, parseUC, createExpenseUC, nil)
	handler := NewHandler("test_app_secret", "123456789", whatsappUseCase)

	payload, _ := createWhatsAppMessagePayload("16505551234", "早餐$20")

	req := httptest.NewRequest("POST", "/webhook/whatsapp", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	handler.HandleWebhook(w, req)

	// Verify rejection
	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

// TestWhatsAppInvalidVerifyToken tests wrong verify token
func TestWhatsAppInvalidVerifyToken(t *testing.T) {
	expenseRepo := &MockExpenseRepository{expenses: make(map[string]*domain.Expense)}
	userRepo := &MockUserRepository{users: make(map[string]*domain.User)}
	categoryRepo := &MockCategoryRepository{categories: make(map[string]*domain.Category)}
	mockAI := &mockAIService{}

	autoSignupUC := usecase.NewAutoSignupUseCase(userRepo, categoryRepo)
	parseUC := usecase.NewParseConversationUseCase(mockAI)
	createExpenseUC := usecase.NewCreateExpenseUseCase(expenseRepo, categoryRepo, mockAI)

	whatsappUseCase := NewWhatsAppUseCase(autoSignupUC, parseUC, createExpenseUC, nil)
	handler := NewHandler("test_app_secret", "123456789", whatsappUseCase)

	// GET with wrong verify token
	req := httptest.NewRequest("GET", "/webhook/whatsapp?hub.mode=subscribe&hub.verify_token=wrong_token&hub.challenge=test_challenge_123", nil)

	w := httptest.NewRecorder()
	handler.HandleWebhook(w, req)

	// Verify rejection
	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

// TestWhatsAppMultipleMessages tests webhook with multiple messages
func TestWhatsAppMultipleMessages(t *testing.T) {
	expenseRepo := &MockExpenseRepository{expenses: make(map[string]*domain.Expense)}
	userRepo := &MockUserRepository{users: make(map[string]*domain.User)}
	categoryRepo := &MockCategoryRepository{categories: make(map[string]*domain.Category)}
	mockAI := &mockAIService{}

	autoSignupUC := usecase.NewAutoSignupUseCase(userRepo, categoryRepo)
	parseUC := usecase.NewParseConversationUseCase(mockAI)
	createExpenseUC := usecase.NewCreateExpenseUseCase(expenseRepo, categoryRepo, mockAI)

	whatsappUseCase := NewWhatsAppUseCase(autoSignupUC, parseUC, createExpenseUC, nil)
	handler := NewHandler("test_app_secret", "123456789", whatsappUseCase)

	payload, signature := createWhatsAppMessagePayload("16505551234", "早餐$20午餐$30晚餐$50")

	req := httptest.NewRequest("POST", "/webhook/whatsapp", bytes.NewReader(payload))
	req.Header.Set("X-Hub-Signature-256", "sha256="+signature)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	handler.HandleWebhook(w, req)

	// Verify success
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

func (m *mockAIService) SuggestCategory(ctx context.Context, description string, userID string) (string, error) {
	return "Food", nil
}
