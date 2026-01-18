package usecase

import (
	"context"
	"regexp"
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

func (m *MockUserRepository) GetAll(ctx context.Context) ([]*domain.User, error) {
	var users []*domain.User
	for _, u := range m.users {
		users = append(users, u)
	}
	return users, nil
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

func NewMockAIService() *MockAIService {
	return &MockAIService{
		shouldFail: false,
	}
}

func (m *MockAIService) ParseExpense(ctx context.Context, text string, userID string) ([]*domain.ParsedExpense, error) {
	if m.shouldFail {
		return nil, nil // Falls back to regex in ParseConversationUseCase
	}

	// Parse using regex pattern to extract expenses
	// This pattern matches: description$amount (e.g., "breakfast $20" or "早餐$20")
	var expenses []*domain.ParsedExpense
	re := regexp.MustCompile(`([^\d$]+)\$(\d+(?:\.\d{2})?)`)
	matches := re.FindAllStringSubmatch(text, -1)

	for _, match := range matches {
		if len(match) < 3 {
			continue
		}

		description := strings.TrimSpace(match[1])
		if description == "" {
			continue
		}

		// Parse amount
		amount := 0.0
		amountStr := match[2]
		for _, c := range amountStr {
			if c >= '0' && c <= '9' {
				amount = amount*10 + float64(c-'0')
			} else if c == '.' {
				break
			}
		}

		expenses = append(expenses, &domain.ParsedExpense{
			Description: description,
			Amount:      amount,
			Date:        time.Now(),
		})
	}

	// If no matches found, return nil to trigger regex fallback
	if len(expenses) == 0 {
		return nil, nil
	}

	return expenses, nil
}

func (m *MockAIService) SuggestCategory(ctx context.Context, description string, userID string) (string, error) {
	descLower := strings.ToLower(description)

	// Food keywords (English + Chinese)
	if strings.Contains(descLower, "breakfast") || strings.Contains(descLower, "lunch") ||
		strings.Contains(descLower, "dinner") || strings.Contains(descLower, "restaurant") ||
		strings.Contains(descLower, "cafe") || strings.Contains(descLower, "meal") ||
		strings.Contains(descLower, "food") || strings.Contains(descLower, "eat") ||
		strings.Contains(description, "早餐") || strings.Contains(description, "午餐") ||
		strings.Contains(description, "晚餐") || strings.Contains(description, "飯") ||
		strings.Contains(description, "食") {
		return "Food", nil
	}

	// Transport keywords (English + Chinese)
	if strings.Contains(descLower, "taxi") || strings.Contains(descLower, "uber") ||
		strings.Contains(descLower, "transport") || strings.Contains(descLower, "gas") ||
		strings.Contains(descLower, "fuel") || strings.Contains(descLower, "bus") ||
		strings.Contains(descLower, "train") || strings.Contains(descLower, "airport") ||
		strings.Contains(descLower, "car") ||
		strings.Contains(description, "計程車") || strings.Contains(description, "的士") ||
		strings.Contains(description, "交通") || strings.Contains(description, "油") {
		return "Transport", nil
	}

	// Shopping keywords (English + Chinese)
	if strings.Contains(descLower, "shopping") || strings.Contains(descLower, "shirt") ||
		strings.Contains(descLower, "clothes") || strings.Contains(descLower, "buy") ||
		strings.Contains(descLower, "shop") || strings.Contains(descLower, "store") ||
		strings.Contains(descLower, "mall") ||
		strings.Contains(description, "購物") || strings.Contains(description, "買") ||
		strings.Contains(description, "衣服") {
		return "Shopping", nil
	}

	// Entertainment keywords (English + Chinese)
	if strings.Contains(descLower, "movie") || strings.Contains(descLower, "cinema") ||
		strings.Contains(descLower, "entertainment") || strings.Contains(descLower, "tickets") ||
		strings.Contains(descLower, "concert") || strings.Contains(descLower, "show") ||
		strings.Contains(description, "電影") || strings.Contains(description, "電影院") ||
		strings.Contains(description, "娛樂") || strings.Contains(description, "演唱會") {
		return "Entertainment", nil
	}

	return "Other", nil
}
