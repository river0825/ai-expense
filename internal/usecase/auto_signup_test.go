package usecase

import (
	"context"
	"testing"
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

func TestAutoSignupNewUser(t *testing.T) {
	userRepo := NewMockUserRepository()
	categoryRepo := NewMockCategoryRepository()
	uc := NewAutoSignupUseCase(userRepo, categoryRepo)

	ctx := context.Background()
	userID := "test_user_123"
	messengerType := "line"

	err := uc.Execute(ctx, userID, messengerType)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify user was created
	user, _ := userRepo.GetByID(ctx, userID)
	if user == nil {
		t.Errorf("expected user to be created")
	}
	if user.UserID != userID {
		t.Errorf("expected user ID %s, got %s", userID, user.UserID)
	}
	if user.MessengerType != messengerType {
		t.Errorf("expected messenger type %s, got %s", messengerType, user.MessengerType)
	}

	// Verify default categories were created
	categories, _ := categoryRepo.GetByUserID(ctx, userID)
	if len(categories) != 5 {
		t.Errorf("expected 5 default categories, got %d", len(categories))
	}

	// Verify all categories are marked as default
	for _, cat := range categories {
		if !cat.IsDefault {
			t.Errorf("expected category %s to be marked as default", cat.Name)
		}
	}
}

func TestAutoSignupExistingUser(t *testing.T) {
	userRepo := NewMockUserRepository()
	categoryRepo := NewMockCategoryRepository()
	uc := NewAutoSignupUseCase(userRepo, categoryRepo)

	ctx := context.Background()
	userID := "existing_user"

	// Create user manually
	existingUser := &domain.User{
		UserID:        userID,
		MessengerType: "line",
		CreatedAt:     time.Now(),
	}
	userRepo.Create(ctx, existingUser)

	// Try to signup same user again
	err := uc.Execute(ctx, userID, "line")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify no new categories were created
	categories, _ := categoryRepo.GetByUserID(ctx, userID)
	if len(categories) != 0 {
		t.Errorf("expected 0 categories for existing user, got %d", len(categories))
	}
}

func TestAutoSignupIdempotent(t *testing.T) {
	userRepo := NewMockUserRepository()
	categoryRepo := NewMockCategoryRepository()
	uc := NewAutoSignupUseCase(userRepo, categoryRepo)

	ctx := context.Background()
	userID := "test_user_456"

	// First signup
	err1 := uc.Execute(ctx, userID, "line")
	if err1 != nil {
		t.Fatalf("first signup failed: %v", err1)
	}

	// Second signup (should be idempotent)
	err2 := uc.Execute(ctx, userID, "line")
	if err2 != nil {
		t.Fatalf("second signup failed: %v", err2)
	}

	// Verify only one user exists
	user, _ := userRepo.GetByID(ctx, userID)
	if user == nil {
		t.Errorf("expected user to exist")
	}

	// Verify only 5 categories exist (not 10)
	categories, _ := categoryRepo.GetByUserID(ctx, userID)
	if len(categories) != 5 {
		t.Errorf("expected 5 categories after idempotent signup, got %d", len(categories))
	}
}

func TestDefaultCategoriesCreated(t *testing.T) {
	userRepo := NewMockUserRepository()
	categoryRepo := NewMockCategoryRepository()
	uc := NewAutoSignupUseCase(userRepo, categoryRepo)

	ctx := context.Background()
	userID := "test_user_789"

	err := uc.Execute(ctx, userID, "line")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	categories, _ := categoryRepo.GetByUserID(ctx, userID)

	expectedCategories := map[string]bool{
		"Food":          false,
		"Transport":     false,
		"Shopping":      false,
		"Entertainment": false,
		"Other":         false,
	}

	for _, cat := range categories {
		if _, exists := expectedCategories[cat.Name]; exists {
			expectedCategories[cat.Name] = true
		}
	}

	for name, found := range expectedCategories {
		if !found {
			t.Errorf("expected category %s not found", name)
		}
	}
}
