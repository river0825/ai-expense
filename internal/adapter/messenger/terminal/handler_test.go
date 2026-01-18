package terminal

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/riverlin/aiexpense/internal/domain"
	"github.com/riverlin/aiexpense/internal/usecase"
)

func TestTerminalHandler_HandleMessage_Success(t *testing.T) {
	// Setup
	userRepo := usecase.NewMockUserRepository()
	categoryRepo := usecase.NewMockCategoryRepository()
	expenseRepo := usecase.NewMockExpenseRepository()
	aiService := usecase.NewMockAIService()

	autoSignup := usecase.NewAutoSignupUseCase(userRepo, categoryRepo)
	parseConv := usecase.NewParseConversationUseCase(aiService)
	createExp := usecase.NewCreateExpenseUseCase(expenseRepo, categoryRepo, aiService)
	getExp := usecase.NewGetExpensesUseCase(expenseRepo, categoryRepo)

	tc := NewTerminalUseCase(autoSignup, parseConv, createExp, getExp, userRepo)
	handler := NewHandler(tc)

	// Create request
	req := TerminalRequest{
		UserID:  "test_user",
		Message: "breakfast $20 lunch $30",
	}

	body, _ := json.Marshal(req)
	httpReq := httptest.NewRequest("POST", "/api/chat/terminal", bytes.NewReader(body))
	w := httptest.NewRecorder()

	// Execute
	handler.HandleMessage(w, httpReq)

	// Assert
	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp TerminalResponse
	json.NewDecoder(w.Body).Decode(&resp)

	if resp.Status != "success" {
		t.Errorf("expected status success, got %s", resp.Status)
	}

	if resp.Message == "" {
		t.Error("expected non-empty message")
	}
}

func TestTerminalHandler_HandleMessage_MethodNotAllowed(t *testing.T) {
	// Setup
	userRepo := usecase.NewMockUserRepository()
	categoryRepo := usecase.NewMockCategoryRepository()
	expenseRepo := usecase.NewMockExpenseRepository()
	aiService := usecase.NewMockAIService()

	autoSignup := usecase.NewAutoSignupUseCase(userRepo, categoryRepo)
	parseConv := usecase.NewParseConversationUseCase(aiService)
	createExp := usecase.NewCreateExpenseUseCase(expenseRepo, categoryRepo, aiService)
	getExp := usecase.NewGetExpensesUseCase(expenseRepo, categoryRepo)

	tc := NewTerminalUseCase(autoSignup, parseConv, createExp, getExp, userRepo)
	handler := NewHandler(tc)

	// Create GET request instead of POST
	httpReq := httptest.NewRequest("GET", "/api/chat/terminal", nil)
	w := httptest.NewRecorder()

	// Execute
	handler.HandleMessage(w, httpReq)

	// Assert
	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", w.Code)
	}

	var resp TerminalResponse
	json.NewDecoder(w.Body).Decode(&resp)

	if resp.Status != "error" {
		t.Errorf("expected status error, got %s", resp.Status)
	}
}

func TestTerminalHandler_HandleMessage_InvalidJSON(t *testing.T) {
	// Setup
	userRepo := usecase.NewMockUserRepository()
	categoryRepo := usecase.NewMockCategoryRepository()
	expenseRepo := usecase.NewMockExpenseRepository()
	aiService := usecase.NewMockAIService()

	autoSignup := usecase.NewAutoSignupUseCase(userRepo, categoryRepo)
	parseConv := usecase.NewParseConversationUseCase(aiService)
	createExp := usecase.NewCreateExpenseUseCase(expenseRepo, categoryRepo, aiService)
	getExp := usecase.NewGetExpensesUseCase(expenseRepo, categoryRepo)

	tc := NewTerminalUseCase(autoSignup, parseConv, createExp, getExp, userRepo)
	handler := NewHandler(tc)

	// Create request with invalid JSON
	httpReq := httptest.NewRequest("POST", "/api/chat/terminal", bytes.NewReader([]byte("invalid json")))
	w := httptest.NewRecorder()

	// Execute
	handler.HandleMessage(w, httpReq)

	// Assert
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}

	var resp TerminalResponse
	json.NewDecoder(w.Body).Decode(&resp)

	if resp.Status != "error" {
		t.Errorf("expected status error, got %s", resp.Status)
	}
}

func TestTerminalHandler_HandleMessage_MissingUserID(t *testing.T) {
	// Setup
	userRepo := usecase.NewMockUserRepository()
	categoryRepo := usecase.NewMockCategoryRepository()
	expenseRepo := usecase.NewMockExpenseRepository()
	aiService := usecase.NewMockAIService()

	autoSignup := usecase.NewAutoSignupUseCase(userRepo, categoryRepo)
	parseConv := usecase.NewParseConversationUseCase(aiService)
	createExp := usecase.NewCreateExpenseUseCase(expenseRepo, categoryRepo, aiService)
	getExp := usecase.NewGetExpensesUseCase(expenseRepo, categoryRepo)

	tc := NewTerminalUseCase(autoSignup, parseConv, createExp, getExp, userRepo)
	handler := NewHandler(tc)

	// Create request with missing user_id
	req := TerminalRequest{
		UserID:  "",
		Message: "breakfast $20",
	}

	body, _ := json.Marshal(req)
	httpReq := httptest.NewRequest("POST", "/api/chat/terminal", bytes.NewReader(body))
	w := httptest.NewRecorder()

	// Execute
	handler.HandleMessage(w, httpReq)

	// Assert
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}

	var resp TerminalResponse
	json.NewDecoder(w.Body).Decode(&resp)

	if resp.Status != "error" {
		t.Errorf("expected status error, got %s", resp.Status)
	}
}

func TestTerminalHandler_HandleMessage_MissingMessage(t *testing.T) {
	// Setup
	userRepo := usecase.NewMockUserRepository()
	categoryRepo := usecase.NewMockCategoryRepository()
	expenseRepo := usecase.NewMockExpenseRepository()
	aiService := usecase.NewMockAIService()

	autoSignup := usecase.NewAutoSignupUseCase(userRepo, categoryRepo)
	parseConv := usecase.NewParseConversationUseCase(aiService)
	createExp := usecase.NewCreateExpenseUseCase(expenseRepo, categoryRepo, aiService)
	getExp := usecase.NewGetExpensesUseCase(expenseRepo, categoryRepo)

	tc := NewTerminalUseCase(autoSignup, parseConv, createExp, getExp, userRepo)
	handler := NewHandler(tc)

	// Create request with missing message
	req := TerminalRequest{
		UserID:  "test_user",
		Message: "",
	}

	body, _ := json.Marshal(req)
	httpReq := httptest.NewRequest("POST", "/api/chat/terminal", bytes.NewReader(body))
	w := httptest.NewRecorder()

	// Execute
	handler.HandleMessage(w, httpReq)

	// Assert
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}

	var resp TerminalResponse
	json.NewDecoder(w.Body).Decode(&resp)

	if resp.Status != "error" {
		t.Errorf("expected status error, got %s", resp.Status)
	}
}

func TestTerminalHandler_GetUserInfo_Success(t *testing.T) {
	// Setup
	userRepo := usecase.NewMockUserRepository()
	categoryRepo := usecase.NewMockCategoryRepository()
	expenseRepo := usecase.NewMockExpenseRepository()
	aiService := usecase.NewMockAIService()

	ctx := context.Background()
	userID := "test_user"
	now := time.Now()

	// Pre-create user
	userRepo.Create(ctx, &domain.User{
		UserID:        userID,
		MessengerType: "terminal",
		CreatedAt:     now,
	})

	// Create category
	categoryRepo.Create(ctx, &domain.Category{
		ID:        "cat_food",
		UserID:    userID,
		Name:      "Food",
		IsDefault: true,
		CreatedAt: now,
	})

	// Create expense
	expenseRepo.Create(ctx, &domain.Expense{
		ID:          "exp1",
		UserID:      userID,
		Description: "breakfast",
		Amount:      20.0,
		CategoryID:  ptrString("cat_food"),
		ExpenseDate: now,
		CreatedAt:   now,
		UpdatedAt:   now,
	})

	autoSignup := usecase.NewAutoSignupUseCase(userRepo, categoryRepo)
	parseConv := usecase.NewParseConversationUseCase(aiService)
	createExp := usecase.NewCreateExpenseUseCase(expenseRepo, categoryRepo, aiService)
	getExp := usecase.NewGetExpensesUseCase(expenseRepo, categoryRepo)

	tc := NewTerminalUseCase(autoSignup, parseConv, createExp, getExp, userRepo)
	handler := NewHandler(tc)

	// Create request
	httpReq := httptest.NewRequest("GET", "/api/chat/terminal/user?user_id="+userID, nil)
	w := httptest.NewRecorder()

	// Execute
	handler.GetUserInfo(w, httpReq)

	// Assert
	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp TerminalResponse
	json.NewDecoder(w.Body).Decode(&resp)

	if resp.Status != "success" {
		t.Errorf("expected status success, got %s", resp.Status)
	}

	if resp.Data == nil {
		t.Fatal("expected data in response")
	}
}

func TestTerminalHandler_GetUserInfo_MethodNotAllowed(t *testing.T) {
	// Setup
	userRepo := usecase.NewMockUserRepository()
	categoryRepo := usecase.NewMockCategoryRepository()
	expenseRepo := usecase.NewMockExpenseRepository()
	aiService := usecase.NewMockAIService()

	autoSignup := usecase.NewAutoSignupUseCase(userRepo, categoryRepo)
	parseConv := usecase.NewParseConversationUseCase(aiService)
	createExp := usecase.NewCreateExpenseUseCase(expenseRepo, categoryRepo, aiService)
	getExp := usecase.NewGetExpensesUseCase(expenseRepo, categoryRepo)

	tc := NewTerminalUseCase(autoSignup, parseConv, createExp, getExp, userRepo)
	handler := NewHandler(tc)

	// Create POST request instead of GET
	httpReq := httptest.NewRequest("POST", "/api/chat/terminal/user", nil)
	w := httptest.NewRecorder()

	// Execute
	handler.GetUserInfo(w, httpReq)

	// Assert
	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", w.Code)
	}

	var resp TerminalResponse
	json.NewDecoder(w.Body).Decode(&resp)

	if resp.Status != "error" {
		t.Errorf("expected status error, got %s", resp.Status)
	}
}

func TestTerminalHandler_GetUserInfo_MissingUserID(t *testing.T) {
	// Setup
	userRepo := usecase.NewMockUserRepository()
	categoryRepo := usecase.NewMockCategoryRepository()
	expenseRepo := usecase.NewMockExpenseRepository()
	aiService := usecase.NewMockAIService()

	autoSignup := usecase.NewAutoSignupUseCase(userRepo, categoryRepo)
	parseConv := usecase.NewParseConversationUseCase(aiService)
	createExp := usecase.NewCreateExpenseUseCase(expenseRepo, categoryRepo, aiService)
	getExp := usecase.NewGetExpensesUseCase(expenseRepo, categoryRepo)

	tc := NewTerminalUseCase(autoSignup, parseConv, createExp, getExp, userRepo)
	handler := NewHandler(tc)

	// Create request without user_id parameter
	httpReq := httptest.NewRequest("GET", "/api/chat/terminal/user", nil)
	w := httptest.NewRecorder()

	// Execute
	handler.GetUserInfo(w, httpReq)

	// Assert
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}

	var resp TerminalResponse
	json.NewDecoder(w.Body).Decode(&resp)

	if resp.Status != "error" {
		t.Errorf("expected status error, got %s", resp.Status)
	}
}

func TestTerminalHandler_GetUserInfo_UserNotFound(t *testing.T) {
	// Setup
	userRepo := usecase.NewMockUserRepository()
	categoryRepo := usecase.NewMockCategoryRepository()
	expenseRepo := usecase.NewMockExpenseRepository()
	aiService := usecase.NewMockAIService()

	autoSignup := usecase.NewAutoSignupUseCase(userRepo, categoryRepo)
	parseConv := usecase.NewParseConversationUseCase(aiService)
	createExp := usecase.NewCreateExpenseUseCase(expenseRepo, categoryRepo, aiService)
	getExp := usecase.NewGetExpensesUseCase(expenseRepo, categoryRepo)

	tc := NewTerminalUseCase(autoSignup, parseConv, createExp, getExp, userRepo)
	handler := NewHandler(tc)

	// Create request with nonexistent user
	httpReq := httptest.NewRequest("GET", "/api/chat/terminal/user?user_id=nonexistent", nil)
	w := httptest.NewRecorder()

	// Execute
	handler.GetUserInfo(w, httpReq)

	// Assert
	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", w.Code)
	}

	var resp TerminalResponse
	json.NewDecoder(w.Body).Decode(&resp)

	if resp.Status != "error" {
		t.Errorf("expected status error, got %s", resp.Status)
	}
}
