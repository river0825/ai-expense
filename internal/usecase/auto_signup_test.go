package usecase

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/riverlin/aiexpense/internal/domain"
)

// MockProfileFetcher is a mock implementation for testing
type MockProfileFetcher struct {
	language string
	err      error
}

func (m *MockProfileFetcher) GetLanguage(ctx context.Context, userID string) (string, error) {
	return m.language, m.err
}

func TestAutoSignup(t *testing.T) {
	userRepo := NewMockUserRepository()
	categoryRepo := NewMockCategoryRepository()
	uc := NewAutoSignupUseCase(userRepo, categoryRepo, nil)

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
	uc := NewAutoSignupUseCase(userRepo, categoryRepo, nil)

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
	uc := NewAutoSignupUseCase(userRepo, categoryRepo, nil)

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
	uc := NewAutoSignupUseCase(userRepo, categoryRepo, nil)

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

func TestAutoSignup_WithProfileFetcher_SetsLocale(t *testing.T) {
	userRepo := NewMockUserRepository()
	categoryRepo := NewMockCategoryRepository()
	fetcher := &MockProfileFetcher{language: "zh-TW"}
	uc := NewAutoSignupUseCase(userRepo, categoryRepo, fetcher)

	ctx := context.Background()
	userID := "locale_user_1"

	err := uc.Execute(ctx, userID, "line")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	user, _ := userRepo.GetByID(ctx, userID)
	if user == nil {
		t.Fatal("expected user to be created")
	}
	if user.Locale != "zh-TW" {
		t.Errorf("expected locale zh-TW, got %s", user.Locale)
	}
}

func TestAutoSignup_ProfileFetcherError_StillCreatesUser(t *testing.T) {
	userRepo := NewMockUserRepository()
	categoryRepo := NewMockCategoryRepository()
	fetcher := &MockProfileFetcher{err: fmt.Errorf("api error")}
	uc := NewAutoSignupUseCase(userRepo, categoryRepo, fetcher)

	ctx := context.Background()
	userID := "locale_user_2"

	err := uc.Execute(ctx, userID, "line")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	user, _ := userRepo.GetByID(ctx, userID)
	if user == nil {
		t.Fatal("expected user to be created")
	}
	if user.Locale != "" {
		t.Errorf("expected empty locale on error, got %s", user.Locale)
	}
}

func TestAutoSignup_NilProfileFetcher_NoLocale(t *testing.T) {
	userRepo := NewMockUserRepository()
	categoryRepo := NewMockCategoryRepository()
	uc := NewAutoSignupUseCase(userRepo, categoryRepo, nil)

	ctx := context.Background()
	userID := "locale_user_3"

	err := uc.Execute(ctx, userID, "line")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	user, _ := userRepo.GetByID(ctx, userID)
	if user == nil {
		t.Fatal("expected user to be created")
	}
	if user.Locale != "" {
		t.Errorf("expected empty locale with nil fetcher, got %s", user.Locale)
	}
}
