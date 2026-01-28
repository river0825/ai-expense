package usecase

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/riverlin/aiexpense/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock dependencies
type mockAutoSignup struct{ mock.Mock }

func (m *mockAutoSignup) Execute(ctx context.Context, userID, sourceType string) error {
	args := m.Called(ctx, userID, sourceType)
	return args.Error(0)
}

type mockParseConversation struct{ mock.Mock }

func (m *mockParseConversation) Execute(ctx context.Context, text, userID string) (*domain.ParseResult, error) {
	args := m.Called(ctx, text, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.ParseResult), args.Error(1)
}

type mockCreateExpense struct{ mock.Mock }

func (m *mockCreateExpense) Execute(ctx context.Context, req *CreateRequest) (*CreateResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*CreateResponse), args.Error(1)
}

type mockGenerateReportLink struct{ mock.Mock }

func (m *mockGenerateReportLink) Execute(userID string) (string, error) {
	args := m.Called(userID)
	return args.String(0), args.Error(1)
}

func TestProcessMessageUseCase_Execute(t *testing.T) {
	t.Run("Success - Single Expense", func(t *testing.T) {
		// Setup
		autoSignup := new(mockAutoSignup)
		parser := new(mockParseConversation)
		creator := new(mockCreateExpense)
		reportLink := new(mockGenerateReportLink)

		uc := NewProcessMessageUseCase(autoSignup, parser, creator, nil, reportLink, nil)

		// Expectations
		autoSignup.On("Execute", mock.Anything, "user1", "terminal").Return(nil)

		parsedExpenses := []*domain.ParsedExpense{
			{Description: "Lunch", Amount: 100, Date: time.Now(), Account: "Taishin"},
		}
		parseResult := &domain.ParseResult{
			Expenses:     parsedExpenses,
			SystemPrompt: "prompt",
			RawResponse:  "response",
		}
		parser.On("Execute", mock.Anything, "Lunch 100 Taishin", "user1").Return(parseResult, nil)

		createResp := &CreateResponse{
			ID:             "1",
			Category:       "Food",
			Message:        "Saved",
			OriginalAmount: 100,
			Currency:       "TWD",
			HomeAmount:     100,
			HomeCurrency:   "TWD",
			ExchangeRate:   1,
		}
		creator.On("Execute", mock.Anything, mock.MatchedBy(func(req *CreateRequest) bool {
			return req.UserID == "user1" && req.Amount == 100
		})).Return(createResp, nil)

		// Execute
		msg := &domain.UserMessage{
			UserID:  "user1",
			Content: "Lunch 100 Taishin",
			Source:  "terminal",
		}
		resp, err := uc.Execute(context.Background(), msg)

		// Verify
		assert.NoError(t, err)
		assert.Contains(t, resp.Text, "Recorded 1 expense")
		assert.Contains(t, resp.Text, "Lunch")
		assert.Contains(t, resp.Text, "100 TWD")
		assert.Contains(t, resp.Text, "Taishin")
		assert.Contains(t, resp.Text, "100 TWD")
		assert.Contains(t, resp.Text, "Taishin")
	})

	t.Run("Failure - Parse Error", func(t *testing.T) {
		// Setup
		autoSignup := new(mockAutoSignup)
		parser := new(mockParseConversation)
		creator := new(mockCreateExpense)
		reportLink := new(mockGenerateReportLink)

		uc := NewProcessMessageUseCase(autoSignup, parser, creator, nil, reportLink, nil)

		// Expectations
		autoSignup.On("Execute", mock.Anything, "user1", "terminal").Return(nil)
		parser.On("Execute", mock.Anything, "Bad input", "user1").Return(nil, fmt.Errorf("parse error"))

		// Execute
		msg := &domain.UserMessage{UserID: "user1", Content: "Bad input", Source: "terminal"}
		resp, err := uc.Execute(context.Background(), msg)

		// Verify
		assert.NoError(t, err) // Should not return error to caller, but handle it in response
		assert.Contains(t, resp.Text, "Failed to parse message")
	})
}
