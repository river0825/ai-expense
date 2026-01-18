package usecase

import (
	"context"
	"strings"
	"time"

	"github.com/riverlin/aiexpense/internal/domain"
)

// MockUserRepository is a mock implementation for testing
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
	return m.users[userID], nil
}

func (m *MockUserRepository) Exists(ctx context.Context, userID string) (bool, error) {
	_, exists := m.users[userID]
	return exists, nil
}

// MockCategoryRepository is a mock implementation for testing
type MockCategoryRepository struct {
	categories map[string]*domain.Category
	keywords   map[string]*domain.CategoryKeyword
}

func NewMockCategoryRepository() *MockCategoryRepository {
	return &MockCategoryRepository{
		categories: make(map[string]*domain.Category),
		keywords:   make(map[string]*domain.CategoryKeyword),
	}
}

func (m *MockCategoryRepository) Create(ctx context.Context, category *domain.Category) error {
	m.categories[category.ID] = category
	return nil
}

func (m *MockCategoryRepository) GetByID(ctx context.Context, id string) (*domain.Category, error) {
	return m.categories[id], nil
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
	return nil, nil
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
	m.keywords[keyword.ID] = keyword
	return nil
}

func (m *MockCategoryRepository) GetKeywordsByCategory(ctx context.Context, categoryID string) ([]*domain.CategoryKeyword, error) {
	var result []*domain.CategoryKeyword
	for _, kw := range m.keywords {
		if kw.CategoryID == categoryID {
			result = append(result, kw)
		}
	}
	return result, nil
}

func (m *MockCategoryRepository) DeleteKeyword(ctx context.Context, id string) error {
	delete(m.keywords, id)
	return nil
}

// MockExpenseRepository is a mock implementation for testing
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
	return m.expenses[id], nil
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
		if exp.UserID == userID && exp.ExpenseDate.After(from) && exp.ExpenseDate.Before(to) {
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

// MockAIService is a mock implementation for testing
type MockAIService struct {
	shouldFail bool
}

func (m *MockAIService) ParseExpense(ctx context.Context, text string, userID string) ([]*domain.ParsedExpense, error) {
	if m.shouldFail {
		return nil, nil // Falls back to regex
	}

	// Mock parsing
	return []*domain.ParsedExpense{
		{
			Description: "early_breakfast",
			Amount:      20,
			Date:        time.Now(),
		},
	}, nil
}

func (m *MockAIService) SuggestCategory(ctx context.Context, description string) (string, error) {
	description = strings.ToLower(description)

	// Food keywords
	if strings.Contains(description, "breakfast") || strings.Contains(description, "lunch") ||
		strings.Contains(description, "dinner") || strings.Contains(description, "restaurant") ||
		strings.Contains(description, "cafe") || strings.Contains(description, "meal") ||
		strings.Contains(description, "food") || strings.Contains(description, "eat") {
		return "Food", nil
	}

	// Transport keywords
	if strings.Contains(description, "taxi") || strings.Contains(description, "uber") ||
		strings.Contains(description, "transport") || strings.Contains(description, "gas") ||
		strings.Contains(description, "fuel") || strings.Contains(description, "bus") ||
		strings.Contains(description, "train") || strings.Contains(description, "airport") ||
		strings.Contains(description, "car") {
		return "Transport", nil
	}

	// Shopping keywords
	if strings.Contains(description, "shopping") || strings.Contains(description, "shirt") ||
		strings.Contains(description, "clothes") || strings.Contains(description, "buy") ||
		strings.Contains(description, "shop") || strings.Contains(description, "store") ||
		strings.Contains(description, "mall") {
		return "Shopping", nil
	}

	// Entertainment keywords
	if strings.Contains(description, "movie") || strings.Contains(description, "cinema") ||
		strings.Contains(description, "entertainment") || strings.Contains(description, "tickets") ||
		strings.Contains(description, "concert") || strings.Contains(description, "show") {
		return "Entertainment", nil
	}

	return "Other", nil
}
