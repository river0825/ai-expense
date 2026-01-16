package slack

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
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

// Helper to create valid Slack webhook payload with signature
func createSlackWebhookPayload(userID, text string, recentTimestamp bool) ([]byte, string, string) {
	// Use current timestamp or old one
	ts := time.Now().Unix()
	if !recentTimestamp {
		ts = time.Now().Unix() - 600 // 10 minutes old
	}

	event := map[string]interface{}{
		"type": "event_callback",
		"event": map[string]interface{}{
			"type":    "message",
			"user":    userID,
			"text":    text,
			"channel": "D123456",
			"ts":      "1234567890.123456",
		},
		"team_id":   "T123456",
		"api_app_id": "A123456",
		"event_id":  "Ev123456",
		"event_time": ts,
	}

	body, _ := json.Marshal(event)
	timestamp := strconv.FormatInt(ts, 10)

	// Compute Slack signature
	basestring := fmt.Sprintf("v0:%s:%s", timestamp, string(body))
	mac := hmac.New(sha256.New, []byte("test_signing_secret"))
	mac.Write([]byte(basestring))
	signature := "v0=" + hex.EncodeToString(mac.Sum(nil))

	return body, signature, timestamp
}

// Helper to create Slack URL verification payload
func createSlackUrlVerificationPayload() ([]byte, string, string) {
	event := map[string]interface{}{
		"type":      "url_verification",
		"challenge": "test_challenge_string_123",
		"team_id":   "T123456",
		"api_app_id": "A123456",
		"event_id":  "Ev123456",
		"event_time": time.Now().Unix(),
	}

	body, _ := json.Marshal(event)
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)

	// Compute Slack signature
	basestring := fmt.Sprintf("v0:%s:%s", timestamp, string(body))
	mac := hmac.New(sha256.New, []byte("test_signing_secret"))
	mac.Write([]byte(basestring))
	signature := "v0=" + hex.EncodeToString(mac.Sum(nil))

	return body, signature, timestamp
}

// TestSlackHandlerValidSignature tests webhook with correct signature
func TestSlackHandlerValidSignature(t *testing.T) {
	// Setup
	expenseRepo := NewMockExpenseRepository()
	userRepo := NewMockUserRepository()
	categoryRepo := NewMockCategoryRepository()
	mockAI := setupMockAIService()

	// Create use cases
	autoSignupUC := usecase.NewAutoSignupUseCase(userRepo, categoryRepo)
	parseUC := usecase.NewParseConversationUseCase(mockAI)
	createExpenseUC := usecase.NewCreateExpenseUseCase(expenseRepo, categoryRepo, mockAI)

	slackUseCase := NewSlackUseCase(autoSignupUC, parseUC, createExpenseUC, nil)
	handler := NewHandler("test_signing_secret", slackUseCase)

	// Create valid webhook
	payload, signature, timestamp := createSlackWebhookPayload("U123456", "早餐$20", true)

	req := httptest.NewRequest("POST", "/webhook/slack", bytes.NewReader(payload))
	req.Header.Set("X-Slack-Request-Signature", signature)
	req.Header.Set("X-Slack-Request-Timestamp", timestamp)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	handler.HandleWebhook(w, req)

	// Verify response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

// TestSlackHandlerInvalidSignature tests webhook rejection with invalid signature
func TestSlackHandlerInvalidSignature(t *testing.T) {
	// Setup
	expenseRepo := NewMockExpenseRepository()
	userRepo := NewMockUserRepository()
	categoryRepo := NewMockCategoryRepository()
	mockAI := setupMockAIService()

	autoSignupUC := usecase.NewAutoSignupUseCase(userRepo, categoryRepo)
	parseUC := usecase.NewParseConversationUseCase(mockAI)
	createExpenseUC := usecase.NewCreateExpenseUseCase(expenseRepo, categoryRepo, mockAI)

	slackUseCase := NewSlackUseCase(autoSignupUC, parseUC, createExpenseUC, nil)
	handler := NewHandler("test_signing_secret", slackUseCase)

	// Create webhook with invalid signature
	payload, _, timestamp := createSlackWebhookPayload("U123456", "早餐$20", true)

	req := httptest.NewRequest("POST", "/webhook/slack", bytes.NewReader(payload))
	req.Header.Set("X-Slack-Request-Signature", "v0=invalid_signature")
	req.Header.Set("X-Slack-Request-Timestamp", timestamp)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	handler.HandleWebhook(w, req)

	// Verify rejection
	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

// TestSlackHandlerReplayAttack tests protection against old timestamps
func TestSlackHandlerReplayAttack(t *testing.T) {
	// Setup
	expenseRepo := NewMockExpenseRepository()
	userRepo := NewMockUserRepository()
	categoryRepo := NewMockCategoryRepository()
	mockAI := setupMockAIService()

	autoSignupUC := usecase.NewAutoSignupUseCase(userRepo, categoryRepo)
	parseUC := usecase.NewParseConversationUseCase(mockAI)
	createExpenseUC := usecase.NewCreateExpenseUseCase(expenseRepo, categoryRepo, mockAI)

	slackUseCase := NewSlackUseCase(autoSignupUC, parseUC, createExpenseUC, nil)
	handler := NewHandler("test_signing_secret", slackUseCase)

	// Create webhook with old timestamp (10+ minutes)
	payload, signature, timestamp := createSlackWebhookPayload("U123456", "早餐$20", false)

	req := httptest.NewRequest("POST", "/webhook/slack", bytes.NewReader(payload))
	req.Header.Set("X-Slack-Request-Signature", signature)
	req.Header.Set("X-Slack-Request-Timestamp", timestamp)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	handler.HandleWebhook(w, req)

	// Verify rejection (replay attack)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d for replay attack, got %d", http.StatusUnauthorized, w.Code)
	}
}

// TestSlackHandlerMissingSignature tests webhook rejection without signature
func TestSlackHandlerMissingSignature(t *testing.T) {
	// Setup
	expenseRepo := NewMockExpenseRepository()
	userRepo := NewMockUserRepository()
	categoryRepo := NewMockCategoryRepository()
	mockAI := setupMockAIService()

	autoSignupUC := usecase.NewAutoSignupUseCase(userRepo, categoryRepo)
	parseUC := usecase.NewParseConversationUseCase(mockAI)
	createExpenseUC := usecase.NewCreateExpenseUseCase(expenseRepo, categoryRepo, mockAI)

	slackUseCase := NewSlackUseCase(autoSignupUC, parseUC, createExpenseUC, nil)
	handler := NewHandler("test_signing_secret", slackUseCase)

	// Create webhook without signature headers
	payload, _, _ := createSlackWebhookPayload("U123456", "早餐$20", true)

	req := httptest.NewRequest("POST", "/webhook/slack", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	handler.HandleWebhook(w, req)

	// Verify rejection
	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

// TestSlackHandlerUrlVerification tests URL verification challenge
func TestSlackHandlerUrlVerification(t *testing.T) {
	// Setup
	expenseRepo := NewMockExpenseRepository()
	userRepo := NewMockUserRepository()
	categoryRepo := NewMockCategoryRepository()
	mockAI := setupMockAIService()

	autoSignupUC := usecase.NewAutoSignupUseCase(userRepo, categoryRepo)
	parseUC := usecase.NewParseConversationUseCase(mockAI)
	createExpenseUC := usecase.NewCreateExpenseUseCase(expenseRepo, categoryRepo, mockAI)

	slackUseCase := NewSlackUseCase(autoSignupUC, parseUC, createExpenseUC, nil)
	handler := NewHandler("test_signing_secret", slackUseCase)

	// Create URL verification payload
	payload, signature, timestamp := createSlackUrlVerificationPayload()

	req := httptest.NewRequest("POST", "/webhook/slack", bytes.NewReader(payload))
	req.Header.Set("X-Slack-Request-Signature", signature)
	req.Header.Set("X-Slack-Request-Timestamp", timestamp)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	handler.HandleWebhook(w, req)

	// Verify response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	// Verify response is the challenge string
	if w.Body.String() != "test_challenge_string_123" {
		t.Errorf("Expected challenge in response body, got: %s", w.Body.String())
	}
}

// TestSlackHandlerBotMessageFiltering tests that bot messages are ignored
func TestSlackHandlerBotMessageFiltering(t *testing.T) {
	// Setup
	expenseRepo := NewMockExpenseRepository()
	userRepo := NewMockUserRepository()
	categoryRepo := NewMockCategoryRepository()
	mockAI := setupMockAIService()

	autoSignupUC := usecase.NewAutoSignupUseCase(userRepo, categoryRepo)
	parseUC := usecase.NewParseConversationUseCase(mockAI)
	createExpenseUC := usecase.NewCreateExpenseUseCase(expenseRepo, categoryRepo, mockAI)

	slackUseCase := NewSlackUseCase(autoSignupUC, parseUC, createExpenseUC, nil)
	handler := NewHandler("test_signing_secret", slackUseCase)

	// Create bot message payload
	ts := time.Now().Unix()
	event := map[string]interface{}{
		"type": "event_callback",
		"event": map[string]interface{}{
			"type":    "message",
			"user":    "U123456",
			"text":    "早餐$20",
			"bot_id":  "B123456", // Bot message
			"channel": "D123456",
		},
		"team_id":    "T123456",
		"event_time": ts,
	}

	body, _ := json.Marshal(event)
	timestamp := strconv.FormatInt(ts, 10)

	basestring := fmt.Sprintf("v0:%s:%s", timestamp, string(body))
	mac := hmac.New(sha256.New, []byte("test_signing_secret"))
	mac.Write([]byte(basestring))
	signature := "v0=" + hex.EncodeToString(mac.Sum(nil))

	req := httptest.NewRequest("POST", "/webhook/slack", bytes.NewReader(body))
	req.Header.Set("X-Slack-Request-Signature", signature)
	req.Header.Set("X-Slack-Request-Timestamp", timestamp)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	handler.HandleWebhook(w, req)

	// Verify success (but no processing)
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	// Verify user was NOT created
	userExists, _ := userRepo.Exists(context.Background(), "slack_U123456")
	if userExists {
		t.Error("Expected user to NOT be created for bot message")
	}
}

// TestSlackHandlerAppMention tests app_mention event handling
func TestSlackHandlerAppMention(t *testing.T) {
	// Setup
	expenseRepo := NewMockExpenseRepository()
	userRepo := NewMockUserRepository()
	categoryRepo := NewMockCategoryRepository()
	mockAI := setupMockAIService()

	autoSignupUC := usecase.NewAutoSignupUseCase(userRepo, categoryRepo)
	parseUC := usecase.NewParseConversationUseCase(mockAI)
	createExpenseUC := usecase.NewCreateExpenseUseCase(expenseRepo, categoryRepo, mockAI)

	slackUseCase := NewSlackUseCase(autoSignupUC, parseUC, createExpenseUC, nil)
	handler := NewHandler("test_signing_secret", slackUseCase)

	// Create app_mention payload
	// Modify event type to app_mention
	event := map[string]interface{}{
		"type": "event_callback",
		"event": map[string]interface{}{
			"type":    "app_mention",
			"user":    "U123456",
			"text":    "<@U_BOT_ID> 早餐$20",
			"channel": "C123456",
		},
		"team_id":    "T123456",
		"event_time": time.Now().Unix(),
	}

	body, _ := json.Marshal(event)
	ts := time.Now().Unix()
	timestamp := strconv.FormatInt(ts, 10)

	basestring := fmt.Sprintf("v0:%s:%s", timestamp, string(body))
	mac := hmac.New(sha256.New, []byte("test_signing_secret"))
	mac.Write([]byte(basestring))
	signature := "v0=" + hex.EncodeToString(mac.Sum(nil))

	req := httptest.NewRequest("POST", "/webhook/slack", bytes.NewReader(body))
	req.Header.Set("X-Slack-Request-Signature", signature)
	req.Header.Set("X-Slack-Request-Timestamp", timestamp)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	handler.HandleWebhook(w, req)

	// Verify success
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
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
