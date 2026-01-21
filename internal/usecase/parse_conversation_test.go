package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/riverlin/aiexpense/internal/domain"
)

// MockAIService is already defined in mocks.go
// We will reuse it or define a specific one locally if needed but with a different name to avoid collision
type TestMockAIService struct {
	shouldFail bool
}

func (m *TestMockAIService) ParseExpense(ctx context.Context, text string, userID string) ([]*domain.ParsedExpense, error) {
	if m.shouldFail {
		return nil, nil
	}
	return []*domain.ParsedExpense{
		{
			Description:       "mock expense",
			Amount:            20,
			SuggestedCategory: "Food",
			Date:              time.Now(),
		},
	}, nil
}

func (m *TestMockAIService) SuggestCategory(ctx context.Context, description string, userID string) (string, error) {
	return "Other", nil
}

func TestParseDateLogic(t *testing.T) {
	tests := []struct {
		name string
		text string
		want time.Time
	}{
		{
			name: "Yesterday",
			text: "昨天 lunch $15",
			want: time.Now().AddDate(0, 0, -1),
		},
		{
			name: "Day before yesterday",
			text: "前天 lunch $15",
			want: time.Now().AddDate(0, 0, -2),
		},
		{
			name: "Tomorrow",
			text: "明天 lunch $15",
			want: time.Now().AddDate(0, 0, 1),
		},
		{
			name: "Day after tomorrow",
			text: "後天 lunch $15",
			want: time.Now().AddDate(0, 0, 2),
		},
		{
			name: "Last week",
			text: "上週 lunch $15",
			want: time.Now().AddDate(0, 0, -7),
		},
		{
			name: "Last month",
			text: "上月 lunch $15",
			want: time.Now().AddDate(0, -1, 0),
		},
		{
			name: "Default (Today)",
			text: "lunch $15",
			want: time.Now(),
		},
	}

	aiService := &TestMockAIService{shouldFail: true} // Use regex fallback to test date logic
	uc := NewParseConversationUseCase(aiService)
	ctx := context.Background()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expenses, err := uc.Execute(ctx, tt.text, "user")
			if err != nil {
				t.Fatalf("Execute() error = %v", err)
			}
			if len(expenses) == 0 {
				t.Fatalf("Execute() returned no expenses")
			}

			got := expenses[0].Date
			// Compare year, month, day only
			if got.Year() != tt.want.Year() || got.Month() != tt.want.Month() || got.Day() != tt.want.Day() {
				t.Errorf("parseDate() = %v, want %v (day comparison)", got.Format("2006-01-02"), tt.want.Format("2006-01-02"))
			}
		})
	}
}
