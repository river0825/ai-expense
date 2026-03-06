package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/riverlin/aiexpense/internal/ai"
	"github.com/riverlin/aiexpense/internal/domain"
)

type trackingSuggestAIService struct {
	suggestCalls int
}

var _ ai.Service = (*trackingSuggestAIService)(nil)

func (s *trackingSuggestAIService) ParseExpense(ctx context.Context, text string, userCtx *domain.UserContext) (*ai.ParseExpenseResponse, error) {
	return nil, nil
}

func (s *trackingSuggestAIService) SuggestCategory(ctx context.Context, description string, userCtx *domain.UserContext) (*ai.SuggestCategoryResponse, error) {
	s.suggestCalls++
	return &ai.SuggestCategoryResponse{
		Category: "Other",
		Tokens: &ai.TokenMetadata{
			InputTokens:  1,
			OutputTokens: 1,
			TotalTokens:  2,
		},
	}, nil
}

func (s *trackingSuggestAIService) ClassifyIntent(ctx context.Context, text string, userCtx *domain.UserContext) (*ai.ClassifyIntentResponse, error) {
	return &ai.ClassifyIntentResponse{Intent: &domain.ClassifiedIntent{Type: domain.IntentUnknown}}, nil
}

func TestCreateExpenseSuccess(t *testing.T) {
	expenseRepo := NewMockExpenseRepository()
	categoryRepo := NewMockCategoryRepository()
	aiService := &MockAIService{}

	uc := NewCreateExpenseUseCase(expenseRepo, categoryRepo, nil, nil, nil, nil, aiService)

	ctx := context.Background()
	userID := "test_user"
	req := &CreateRequest{
		UserID:      userID,
		Description: "早餐",
		Amount:      20,
		Date:        time.Now(),
	}

	resp, err := uc.Execute(ctx, req)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp == nil {
		t.Errorf("expected response, got nil")
	}

	if resp.ID == "" {
		t.Errorf("expected expense ID, got empty string")
	}

	// Verify expense was created
	expense, _ := expenseRepo.GetByID(ctx, resp.ID)
	if expense == nil {
		t.Errorf("expected expense to be created")
	}

	if expense.Description != "早餐" {
		t.Errorf("expected description 早餐, got %s", expense.Description)
	}

	if expense.Amount != 20 {
		t.Errorf("expected amount 20, got %f", expense.Amount)
	}
}

func TestCreateExpenseWithCategory(t *testing.T) {
	expenseRepo := NewMockExpenseRepository()
	categoryRepo := NewMockCategoryRepository()
	aiService := &MockAIService{}

	uc := NewCreateExpenseUseCase(expenseRepo, categoryRepo, nil, nil, nil, nil, aiService)

	ctx := context.Background()
	userID := "test_user"
	categoryID := "cat_123"

	req := &CreateRequest{
		UserID:      userID,
		Description: "早餐",
		Amount:      20,
		CategoryID:  &categoryID,
		Date:        time.Now(),
	}

	resp, err := uc.Execute(ctx, req)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify expense was created with category
	expense, _ := expenseRepo.GetByID(ctx, resp.ID)
	if expense.CategoryID == nil || *expense.CategoryID != categoryID {
		t.Errorf("expected category ID %s, got %v", categoryID, expense.CategoryID)
	}
}

func TestCreateExpenseWithAICategory(t *testing.T) {
	expenseRepo := NewMockExpenseRepository()
	categoryRepo := NewMockCategoryRepository()

	// Create a Food category in the repo
	foodCat := &domain.Category{
		ID:        "cat_food",
		UserID:    "test_user",
		Name:      "Food",
		IsDefault: true,
	}
	categoryRepo.Create(context.Background(), foodCat)

	aiService := &MockAIService{}

	uc := NewCreateExpenseUseCase(expenseRepo, categoryRepo, nil, nil, nil, nil, aiService)

	ctx := context.Background()
	userID := "test_user"

	req := &CreateRequest{
		UserID:      userID,
		Description: "早餐",
		Amount:      20,
		Date:        time.Now(),
	}

	resp, err := uc.Execute(ctx, req)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify expense has Food category suggestion in message
	if resp.Category != "Food" {
		t.Errorf("expected category Food in response, got %s", resp.Category)
	}
}

func TestCreateExpenseMessage(t *testing.T) {
	expenseRepo := NewMockExpenseRepository()
	categoryRepo := NewMockCategoryRepository()
	aiService := &MockAIService{}

	uc := NewCreateExpenseUseCase(expenseRepo, categoryRepo, nil, nil, nil, nil, aiService)

	ctx := context.Background()
	req := &CreateRequest{
		UserID:      "test_user",
		Description: "早餐",
		Amount:      20,
		Date:        time.Now(),
	}

	resp, err := uc.Execute(ctx, req)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check message contains Chinese confirmation
	if !contains(resp.Message, "已儲存") {
		t.Errorf("expected message to contain 已儲存, got %s", resp.Message)
	}

	if !contains(resp.Message, "早餐") {
		t.Errorf("expected message to contain description, got %s", resp.Message)
	}

	if !contains(resp.Message, "20") {
		t.Errorf("expected message to contain amount, got %s", resp.Message)
	}
}

func TestCreateExpenseDecimalAmount(t *testing.T) {
	expenseRepo := NewMockExpenseRepository()
	categoryRepo := NewMockCategoryRepository()
	aiService := &MockAIService{}

	uc := NewCreateExpenseUseCase(expenseRepo, categoryRepo, nil, nil, nil, nil, aiService)

	ctx := context.Background()
	req := &CreateRequest{
		UserID:      "test_user",
		Description: "咖啡",
		Amount:      3.50,
		Date:        time.Now(),
	}

	resp, err := uc.Execute(ctx, req)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check message contains decimal amount
	if !contains(resp.Message, "3.5") && !contains(resp.Message, "3.50") {
		t.Errorf("expected message to contain amount 3.50, got %s", resp.Message)
	}
}

func TestCreateExpenseUsesParsedSuggestedCategory(t *testing.T) {
	expenseRepo := NewMockExpenseRepository()
	categoryRepo := NewMockCategoryRepository()
	aiService := &trackingSuggestAIService{}

	userID := "test_user"
	transportCategory := &domain.Category{
		ID:        "cat_transport",
		UserID:    userID,
		Name:      "Transport",
		IsDefault: true,
		CreatedAt: time.Now(),
	}
	if err := categoryRepo.Create(context.Background(), transportCategory); err != nil {
		t.Fatalf("failed to seed category: %v", err)
	}

	uc := NewCreateExpenseUseCase(expenseRepo, categoryRepo, nil, nil, nil, nil, aiService)

	req := &CreateRequest{
		UserID:            userID,
		Description:       "Uber ride",
		Amount:            250,
		SuggestedCategory: "Transport",
		Date:              time.Now(),
	}

	resp, err := uc.Execute(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if aiService.suggestCalls != 0 {
		t.Fatalf("expected no AI suggest calls, got %d", aiService.suggestCalls)
	}

	if resp.Category != "Transport" {
		t.Fatalf("expected category Transport, got %s", resp.Category)
	}

	expense, err := expenseRepo.GetByID(context.Background(), resp.ID)
	if err != nil {
		t.Fatalf("failed to load expense: %v", err)
	}
	if expense == nil || expense.CategoryID == nil || *expense.CategoryID != transportCategory.ID {
		t.Fatalf("expected stored category id %s, got %#v", transportCategory.ID, expense)
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
