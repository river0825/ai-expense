package usecase

import (
	"context"
	"fmt"
	"io"
	"log/slog"
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

type mockUserRepoForProcess struct{ mock.Mock }

func (m *mockUserRepoForProcess) Create(ctx context.Context, user *domain.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *mockUserRepoForProcess) GetByID(ctx context.Context, userID string) (*domain.User, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *mockUserRepoForProcess) Exists(ctx context.Context, userID string) (bool, error) {
	args := m.Called(ctx, userID)
	return args.Bool(0), args.Error(1)
}

func (m *mockUserRepoForProcess) Update(ctx context.Context, user *domain.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

type mockConversationStateRepo struct{ mock.Mock }

func (m *mockConversationStateRepo) GetByUserID(ctx context.Context, userID string) (*domain.ConversationState, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.ConversationState), args.Error(1)
}

func (m *mockConversationStateRepo) Upsert(ctx context.Context, state *domain.ConversationState) error {
	args := m.Called(ctx, state)
	return args.Error(0)
}

func (m *mockConversationStateRepo) DeleteByUserID(ctx context.Context, userID string) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func TestProcessMessageUseCase_Execute(t *testing.T) {
	t.Run("Success - Single Expense", func(t *testing.T) {
		// Setup
		autoSignup := new(mockAutoSignup)
		parser := new(mockParseConversation)
		creator := new(mockCreateExpense)
		reportLink := new(mockGenerateReportLink)

		expenseRepo := NewMockExpenseRepository()

		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		uc := NewProcessMessageUseCase(autoSignup, parser, creator, nil, reportLink, nil, expenseRepo, logger)
		uc := NewProcessMessageUseCase(autoSignup, parser, creator, nil, nil, reportLink, nil, nil)

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
		assert.Equal(t, domain.ResponseTypeExpense, resp.Type)
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

		expenseRepo := NewMockExpenseRepository()

		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		uc := NewProcessMessageUseCase(autoSignup, parser, creator, nil, reportLink, nil, expenseRepo, logger)
		uc := NewProcessMessageUseCase(autoSignup, parser, creator, nil, nil, reportLink, nil, nil)

		// Expectations
		autoSignup.On("Execute", mock.Anything, "user1", "terminal").Return(nil)
		parser.On("Execute", mock.Anything, "Bad input", "user1").Return(nil, fmt.Errorf("parse error"))

		// Execute
		msg := &domain.UserMessage{UserID: "user1", Content: "Bad input", Source: "terminal"}
		resp, err := uc.Execute(context.Background(), msg)

		// Verify
		assert.NoError(t, err) // Should not return error to caller, but handle it in response
		assert.Equal(t, domain.ResponseTypeError, resp.Type)
		assert.Contains(t, resp.Text, "Failed to parse message")
	})

	t.Run("Success - Idempotency (Duplicate Message)", func(t *testing.T) {
		// Setup
	t.Run("Currency intent asks clarification and stores pending state", func(t *testing.T) {
		autoSignup := new(mockAutoSignup)
		parser := new(mockParseConversation)
		creator := new(mockCreateExpense)
		reportLink := new(mockGenerateReportLink)
		expenseRepo := NewMockExpenseRepository()

		// Helper to create string pointer
		strPtr := func(s string) *string { return &s }

		// Pre-populate duplicate expense
		existingExpense := &domain.Expense{
			ID:              "exp_1",
			UserID:          "user1",
			Description:     "Lunch",
			OriginalAmount:  100,
			Currency:        "TWD",
			HomeAmount:      100,
			HomeCurrency:    "TWD",
			SourceMessageID: strPtr("msg_123_0"),
			Account:         "Taishin",
			ExpenseDate:     time.Now(),
		}
		expenseRepo.Create(context.Background(), existingExpense)

		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		uc := NewProcessMessageUseCase(autoSignup, parser, creator, nil, reportLink, nil, expenseRepo, logger)

		// Expectations
		autoSignup.On("Execute", mock.Anything, "user1", "terminal").Return(nil)
		// Parser and creator should NOT be called

		// Execute
		msg := &domain.UserMessage{
			UserID:    "user1",
			MessageID: "msg_123",
			Content:   "Lunch 100 Taishin",
			Source:    "terminal",
		}
		resp, err := uc.Execute(context.Background(), msg)

		// Verify
		assert.NoError(t, err)
		assert.Equal(t, domain.ResponseTypeExpense, resp.Type)
		assert.Contains(t, resp.Text, "Recorded 1 expense")
		assert.Contains(t, resp.Text, "Lunch")
		// Verify no extra calls
		parser.AssertNotCalled(t, "Execute", mock.Anything, mock.Anything, mock.Anything)
		creator.AssertNotCalled(t, "Execute", mock.Anything, mock.Anything)
		userRepo := new(mockUserRepoForProcess)
		stateRepo := new(mockConversationStateRepo)

		uc := NewProcessMessageUseCase(autoSignup, parser, creator, nil, userRepo, reportLink, nil, stateRepo)

		autoSignup.On("Execute", mock.Anything, "user1", "terminal").Return(nil)
		stateRepo.On("GetByUserID", mock.Anything, "user1").Return(nil, nil)
		stateRepo.On("Upsert", mock.Anything, mock.MatchedBy(func(state *domain.ConversationState) bool {
			return state.UserID == "user1" && state.ActiveIntent == "settings.currency.set"
		})).Return(nil)
		userRepo.On("GetByID", mock.Anything, "user1").Return(&domain.User{UserID: "user1", Locale: "zh-TW"}, nil)

		msg := &domain.UserMessage{UserID: "user1", Content: "把預設幣別改成", Source: "terminal"}
		resp, err := uc.Execute(context.Background(), msg)

		assert.NoError(t, err)
		assert.Equal(t, domain.ResponseTypeInfo, resp.Type)
		assert.Contains(t, resp.Text, "你要切換成哪個幣別")
	})

	t.Run("Pending currency follow-up applies update and clears state", func(t *testing.T) {
		autoSignup := new(mockAutoSignup)
		parser := new(mockParseConversation)
		creator := new(mockCreateExpense)
		reportLink := new(mockGenerateReportLink)
		userRepo := new(mockUserRepoForProcess)
		stateRepo := new(mockConversationStateRepo)

		uc := NewProcessMessageUseCase(autoSignup, parser, creator, nil, userRepo, reportLink, nil, stateRepo)

		autoSignup.On("Execute", mock.Anything, "user1", "terminal").Return(nil)
		stateRepo.On("GetByUserID", mock.Anything, "user1").Return(&domain.ConversationState{
			UserID:       "user1",
			ActiveIntent: "settings.currency.set",
			PendingSlots: map[string]string{"target_currency": ""},
			Status:       "pending",
			ExpiresAt:    time.Now().Add(5 * time.Minute),
		}, nil)
		userRepo.On("GetByID", mock.Anything, "user1").Return(&domain.User{UserID: "user1", Locale: "zh-TW", HomeCurrency: "TWD"}, nil).Twice()
		userRepo.On("Update", mock.Anything, mock.MatchedBy(func(user *domain.User) bool {
			return user.UserID == "user1" && user.HomeCurrency == "JPY"
		})).Return(nil)
		stateRepo.On("DeleteByUserID", mock.Anything, "user1").Return(nil)

		msg := &domain.UserMessage{UserID: "user1", Content: "JPY", Source: "terminal"}
		resp, err := uc.Execute(context.Background(), msg)

		assert.NoError(t, err)
		assert.Equal(t, domain.ResponseTypeInfo, resp.Type)
		assert.Contains(t, resp.Text, "JPY")
	})

	t.Run("Expired pending state asks to restate request", func(t *testing.T) {
		autoSignup := new(mockAutoSignup)
		parser := new(mockParseConversation)
		creator := new(mockCreateExpense)
		reportLink := new(mockGenerateReportLink)
		userRepo := new(mockUserRepoForProcess)
		stateRepo := new(mockConversationStateRepo)

		uc := NewProcessMessageUseCase(autoSignup, parser, creator, nil, userRepo, reportLink, nil, stateRepo)

		autoSignup.On("Execute", mock.Anything, "user1", "terminal").Return(nil)
		stateRepo.On("GetByUserID", mock.Anything, "user1").Return(&domain.ConversationState{
			UserID:       "user1",
			ActiveIntent: "settings.currency.set",
			PendingSlots: map[string]string{"target_currency": ""},
			Status:       "pending",
			ExpiresAt:    time.Now().Add(-1 * time.Minute),
		}, nil)
		stateRepo.On("DeleteByUserID", mock.Anything, "user1").Return(nil)
		userRepo.On("GetByID", mock.Anything, "user1").Return(&domain.User{UserID: "user1", Locale: "zh-TW"}, nil)

		msg := &domain.UserMessage{UserID: "user1", Content: "JPY", Source: "terminal"}
		resp, err := uc.Execute(context.Background(), msg)

		assert.NoError(t, err)
		assert.Equal(t, domain.ResponseTypeInfo, resp.Type)
		assert.Contains(t, resp.Text, "逾時")
	})
}
