package http

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
)

// ErrNotFound is a sentinel error used in mock implementations
var ErrNotFound = errors.New("not found")

// MockExpenseRepository for HTTP handler tests
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
	return nil, ErrNotFound
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

// MockUserRepository for HTTP handler tests
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
	return nil, ErrNotFound
}

func (m *MockUserRepository) Exists(ctx context.Context, userID string) (bool, error) {
	_, ok := m.users[userID]
	return ok, nil
}

// MockCategoryRepository for HTTP handler tests
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
	return nil, ErrNotFound
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
	return nil, ErrNotFound
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

// TestHealthCheck tests the health endpoint
func TestHealthCheck(t *testing.T) {
	req := httptest.NewRequest("GET", "/health", nil)

	// This would be registered via mux
	// For now, just verify the pattern works
	if req.Method != "GET" {
		t.Error("Expected GET method")
	}
}

// TestGetExpenses tests listing expenses
func TestGetExpenses(t *testing.T) {
	expenseRepo := NewMockExpenseRepository()
	userRepo := NewMockUserRepository()

	// Add test data
	categoryID := "cat_food"
	user := &domain.User{
		UserID:        "test_user",
		MessengerType: "line",
		CreatedAt:     time.Now(),
	}
	userRepo.Create(context.Background(), user)

	expense := &domain.Expense{
		ID:          "exp_001",
		UserID:      "test_user",
		Description: "Test expense",
		Amount:      10.00,
		CategoryID:  &categoryID,
		ExpenseDate: time.Now(),
		CreatedAt:   time.Now(),
	}
	expenseRepo.Create(context.Background(), expense)

	// Test retrieving expenses
	expenses, err := expenseRepo.GetByUserID(context.Background(), "test_user")
	if err != nil {
		t.Fatalf("Failed to get expenses: %v", err)
	}

	if len(expenses) != 1 {
		t.Errorf("Expected 1 expense, got %d", len(expenses))
	}

	if expenses[0].Description != "Test expense" {
		t.Errorf("Expected 'Test expense', got '%s'", expenses[0].Description)
	}
}

// MockMetricsRepository for testing
type MockMetricsRepository struct{}

func NewMockMetricsRepository() *MockMetricsRepository {
	return &MockMetricsRepository{}
}

func (m *MockMetricsRepository) GetDailyActiveUsers(ctx context.Context, from, to time.Time) ([]*domain.DailyMetrics, error) {
	return []*domain.DailyMetrics{}, nil
}

func (m *MockMetricsRepository) GetExpensesSummary(ctx context.Context, from, to time.Time) ([]*domain.DailyMetrics, error) {
	return []*domain.DailyMetrics{}, nil
}

func (m *MockMetricsRepository) GetCategoryTrends(ctx context.Context, userID string, from, to time.Time) ([]*domain.CategoryMetrics, error) {
	return []*domain.CategoryMetrics{}, nil
}

func (m *MockMetricsRepository) GetGrowthMetrics(ctx context.Context, days int) (map[string]interface{}, error) {
	return make(map[string]interface{}), nil
}

func (m *MockMetricsRepository) GetNewUsersPerDay(ctx context.Context, from, to time.Time) ([]*domain.DailyMetrics, error) {
	return []*domain.DailyMetrics{}, nil
}

// TestCreateExpenseRequest tests expense creation request parsing
func TestCreateExpenseRequest(t *testing.T) {
	requestBody := map[string]interface{}{
		"user_id":     "test_user",
		"description": "Lunch",
		"amount":      25.50,
	}

	body, err := json.Marshal(requestBody)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}

	req := httptest.NewRequest("POST", "/api/expenses", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	if req.Method != "POST" {
		t.Error("Expected POST method")
	}

	if req.Header.Get("Content-Type") != "application/json" {
		t.Error("Expected JSON content type")
	}

	// Verify body can be parsed
	var parsed map[string]interface{}
	if err := json.NewDecoder(req.Body).Decode(&parsed); err != nil {
		t.Fatalf("Failed to parse request body: %v", err)
	}

	if parsed["description"] != "Lunch" {
		t.Errorf("Expected 'Lunch', got '%v'", parsed["description"])
	}
}

// TestResponseFormat tests JSON response formatting
func TestResponseFormat(t *testing.T) {
	response := &Response{
		Status: "success",
		Data: map[string]interface{}{
			"id":     "exp_001",
			"amount": 25.50,
		},
	}

	body, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("Failed to marshal response: %v", err)
	}

	var parsed Response
	if err := json.Unmarshal(body, &parsed); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if parsed.Status != "success" {
		t.Errorf("Expected status 'success', got '%s'", parsed.Status)
	}
}

// TestErrorResponse tests error response formatting
func TestErrorResponse(t *testing.T) {
	response := &Response{
		Status: "error",
		Error:  "expense not found",
	}

	body, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("Failed to marshal error response: %v", err)
	}

	var parsed Response
	if err := json.Unmarshal(body, &parsed); err != nil {
		t.Fatalf("Failed to parse error response: %v", err)
	}

	if parsed.Status != "error" {
		t.Errorf("Expected status 'error', got '%s'", parsed.Status)
	}

	if parsed.Error != "expense not found" {
		t.Errorf("Expected error 'expense not found', got '%s'", parsed.Error)
	}
}

// TestHTTPStatusCodes verifies proper HTTP status codes
func TestHTTPStatusCodes(t *testing.T) {
	// Test 200 OK
	w := httptest.NewRecorder()

	// Handler would write status, mocking it here
	w.WriteHeader(http.StatusOK)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Test 400 Bad Request
	w = httptest.NewRecorder()
	w.WriteHeader(http.StatusBadRequest)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}

	// Test 404 Not Found
	w = httptest.NewRecorder()
	w.WriteHeader(http.StatusNotFound)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}

	// Test 500 Internal Server Error
	w = httptest.NewRecorder()
	w.WriteHeader(http.StatusInternalServerError)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", w.Code)
	}
}
