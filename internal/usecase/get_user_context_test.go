package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/riverlin/aiexpense/internal/domain"
)

func TestGetUserContext_HappyPath(t *testing.T) {
	userRepo := NewMockUserRepository()
	categoryRepo := NewMockCategoryRepository()

	ctx := context.Background()
	userID := "test_user"

	user := &domain.User{
		UserID:       userID,
		HomeCurrency: "TWD",
		Locale:       "zh-TW",
		CreatedAt:    time.Now(),
	}
	userRepo.Create(ctx, user)

	cat1 := &domain.Category{ID: "cat_1", UserID: userID, Name: "Food", IsDefault: true}
	cat2 := &domain.Category{ID: "cat_2", UserID: userID, Name: "Transport", IsDefault: true}
	categoryRepo.Create(ctx, cat1)
	categoryRepo.Create(ctx, cat2)

	uc := NewGetUserContextUseCase(userRepo, categoryRepo)
	result, err := uc.Execute(ctx, userID)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected result, got nil")
	}

	if result.User == nil {
		t.Fatal("expected user in result")
	}

	if result.User.UserID != userID {
		t.Errorf("expected userID %s, got %s", userID, result.User.UserID)
	}

	if result.User.HomeCurrency != "TWD" {
		t.Errorf("expected HomeCurrency TWD, got %s", result.User.HomeCurrency)
	}

	if len(result.Categories) != 2 {
		t.Errorf("expected 2 categories, got %d", len(result.Categories))
	}
}

func TestGetUserContext_UserWithNoCategories(t *testing.T) {
	userRepo := NewMockUserRepository()
	categoryRepo := NewMockCategoryRepository()

	ctx := context.Background()
	userID := "user_no_cats"

	user := &domain.User{
		UserID:       userID,
		HomeCurrency: "USD",
		Locale:       "en-US",
		CreatedAt:    time.Now(),
	}
	userRepo.Create(ctx, user)

	uc := NewGetUserContextUseCase(userRepo, categoryRepo)
	result, err := uc.Execute(ctx, userID)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected result, got nil")
	}

	if result.User == nil {
		t.Fatal("expected user in result")
	}

	if result.Categories == nil {
		t.Error("expected categories to be initialized (empty slice), got nil")
	}

	if len(result.Categories) != 0 {
		t.Errorf("expected 0 categories, got %d", len(result.Categories))
	}
}

func TestGetUserContext_UserNotFound(t *testing.T) {
	userRepo := NewMockUserRepository()
	categoryRepo := NewMockCategoryRepository()

	ctx := context.Background()
	userID := "nonexistent_user"

	uc := NewGetUserContextUseCase(userRepo, categoryRepo)
	result, err := uc.Execute(ctx, userID)

	if err == nil {
		t.Fatal("expected error for nonexistent user, got nil")
	}

	if result != nil {
		t.Errorf("expected nil result for nonexistent user, got %+v", result)
	}
}
