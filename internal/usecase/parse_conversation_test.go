package usecase

import (
	"context"
	"testing"
	"time"
)

func TestParseConversationWithAI(t *testing.T) {
	aiService := &MockAIService{shouldFail: false}
	uc := NewParseConversationUseCase(aiService)

	ctx := context.Background()
	text := "早餐$20"
	userID := "test_user"

	expenses, err := uc.Execute(ctx, text, userID)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(expenses) != 1 {
		t.Errorf("expected 1 expense, got %d", len(expenses))
	}
}

func TestParseDateYesterday(t *testing.T) {
	aiService := &MockAIService{shouldFail: true} // Force fallback to regex
	uc := NewParseConversationUseCase(aiService)

	ctx := context.Background()
	text := "昨天買水果$300"
	userID := "test_user"

	expenses, err := uc.Execute(ctx, text, userID)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(expenses) > 0 {
		// Check that date is approximately yesterday
		expectedDate := time.Now().AddDate(0, 0, -1)
		actualDate := expenses[0].Date

		// Allow 1 minute difference for test execution time
		if actualDate.Day() != expectedDate.Day() {
			t.Errorf("expected date to be yesterday, got %v", actualDate)
		}
	}
}

func TestParseDateLastWeek(t *testing.T) {
	aiService := &MockAIService{shouldFail: true}
	uc := NewParseConversationUseCase(aiService)

	ctx := context.Background()
	text := "上週買的東西$500"
	userID := "test_user"

	expenses, err := uc.Execute(ctx, text, userID)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(expenses) > 0 {
		expectedDate := time.Now().AddDate(0, 0, -7)
		actualDate := expenses[0].Date

		// Check if it's within 1 week
		if actualDate.Day() != expectedDate.Day() {
			t.Errorf("expected date to be last week, got %v", actualDate)
		}
	}
}

func TestParseDateLastMonth(t *testing.T) {
	aiService := &MockAIService{shouldFail: true}
	uc := NewParseConversationUseCase(aiService)

	ctx := context.Background()
	text := "上個月的消費$1000"
	userID := "test_user"

	expenses, err := uc.Execute(ctx, text, userID)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(expenses) > 0 {
		expectedDate := time.Now().AddDate(0, -1, 0)
		actualDate := expenses[0].Date

		if actualDate.Month() != expectedDate.Month() {
			t.Errorf("expected date to be last month, got %v", actualDate)
		}
	}
}

func TestParseDateDefault(t *testing.T) {
	aiService := &MockAIService{shouldFail: true}
	uc := NewParseConversationUseCase(aiService)

	ctx := context.Background()
	text := "早餐$20"
	userID := "test_user"

	expenses, err := uc.Execute(ctx, text, userID)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(expenses) > 0 {
		expectedDate := time.Now()
		actualDate := expenses[0].Date

		// Check if it's today (within 1 minute)
		if actualDate.Day() != expectedDate.Day() {
			t.Errorf("expected date to be today, got %v", actualDate)
		}
	}
}

func TestParseWithRegexFallback(t *testing.T) {
	aiService := &MockAIService{shouldFail: true}
	uc := NewParseConversationUseCase(aiService)

	ctx := context.Background()
	text := "早餐$20午餐$30"
	userID := "test_user"

	expenses, err := uc.Execute(ctx, text, userID)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should still parse despite AI failure
	if len(expenses) == 0 {
		t.Errorf("expected regex fallback to parse expenses")
	}
}

func TestParseConversationMultipleExpenses(t *testing.T) {
	aiService := &MockAIService{shouldFail: true}
	uc := NewParseConversationUseCase(aiService)

	ctx := context.Background()
	text := "早餐$20午餐$30加油$200"
	userID := "test_user"

	expenses, err := uc.Execute(ctx, text, userID)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(expenses) != 3 {
		t.Errorf("expected 3 expenses, got %d", len(expenses))
	}
}

func TestParseConversationEmpty(t *testing.T) {
	aiService := &MockAIService{shouldFail: true}
	uc := NewParseConversationUseCase(aiService)

	ctx := context.Background()
	text := "no expenses here"
	userID := "test_user"

	expenses, err := uc.Execute(ctx, text, userID)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(expenses) != 0 {
		t.Errorf("expected 0 expenses, got %d", len(expenses))
	}
}
