package http

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/riverlin/aiexpense/internal/ai"
	"github.com/riverlin/aiexpense/internal/domain"
	"github.com/riverlin/aiexpense/internal/usecase"
)

// Handler holds all HTTP request handlers
type Handler struct {
	autoSignupUC         *usecase.AutoSignupUseCase
	parseConversationUC  *usecase.ParseConversationUseCase
	createExpenseUC      *usecase.CreateExpenseUseCase
	getExpensesUC        *usecase.GetExpensesUseCase
	userRepo             domain.UserRepository
	categoryRepo         domain.CategoryRepository
	expenseRepo          domain.ExpenseRepository
	metricsRepo          domain.MetricsRepository
	adminAPIKey          string
}

// NewHandler creates a new HTTP handler
func NewHandler(
	autoSignupUC *usecase.AutoSignupUseCase,
	parseConversationUC *usecase.ParseConversationUseCase,
	createExpenseUC *usecase.CreateExpenseUseCase,
	getExpensesUC *usecase.GetExpensesUseCase,
	userRepo domain.UserRepository,
	categoryRepo domain.CategoryRepository,
	expenseRepo domain.ExpenseRepository,
	metricsRepo domain.MetricsRepository,
	adminAPIKey string,
) *Handler {
	return &Handler{
		autoSignupUC:        autoSignupUC,
		parseConversationUC: parseConversationUC,
		createExpenseUC:     createExpenseUC,
		getExpensesUC:       getExpensesUC,
		userRepo:            userRepo,
		categoryRepo:        categoryRepo,
		expenseRepo:         expenseRepo,
		metricsRepo:         metricsRepo,
		adminAPIKey:         adminAPIKey,
	}
}

// JSON response wrapper
type Response struct {
	Status  string      `json:"status"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Message string      `json:"message,omitempty"`
}

// WriteJSON writes a JSON response
func (h *Handler) WriteJSON(w http.ResponseWriter, status int, resp *Response) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(resp)
}

// ReadJSON reads a JSON request body
func (h *Handler) ReadJSON(r *http.Request, v interface{}) error {
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(v)
}

// AutoSignup godoc
// @Summary Auto-signup a user
// @Description Create a new user if not exists
// @Tags users
// @Accept json
// @Produce json
// @Param req body AutoSignupRequest true "Signup request"
// @Success 200 {object} Response
// @Router /api/users/auto-signup [post]
func (h *Handler) AutoSignup(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	type AutoSignupRequest struct {
		UserID        string `json:"user_id"`
		MessengerType string `json:"messenger_type"`
	}

	var req AutoSignupRequest
	if err := h.ReadJSON(r, &req); err != nil {
		h.WriteJSON(w, http.StatusBadRequest, &Response{Status: "error", Error: "Invalid request"})
		return
	}

	if err := h.autoSignupUC.Execute(ctx, req.UserID, req.MessengerType); err != nil {
		h.WriteJSON(w, http.StatusInternalServerError, &Response{Status: "error", Error: err.Error()})
		return
	}

	h.WriteJSON(w, http.StatusOK, &Response{Status: "success", Message: "User signed up successfully"})
}

// ParseExpenses godoc
func (h *Handler) ParseExpenses(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	type ParseRequest struct {
		UserID string `json:"user_id"`
		Text   string `json:"text"`
	}

	var req ParseRequest
	if err := h.ReadJSON(r, &req); err != nil {
		h.WriteJSON(w, http.StatusBadRequest, &Response{Status: "error", Error: "Invalid request"})
		return
	}

	expenses, err := h.parseConversationUC.Execute(ctx, req.Text, req.UserID)
	if err != nil {
		h.WriteJSON(w, http.StatusInternalServerError, &Response{Status: "error", Error: err.Error()})
		return
	}

	h.WriteJSON(w, http.StatusOK, &Response{Status: "success", Data: expenses})
}

// CreateExpense godoc
func (h *Handler) CreateExpense(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	type CreateRequest struct {
		UserID      string     `json:"user_id"`
		Description string     `json:"description"`
		Amount      float64    `json:"amount"`
		CategoryID  *string    `json:"category_id,omitempty"`
		Date        *time.Time `json:"date,omitempty"`
	}

	var req CreateRequest
	if err := h.ReadJSON(r, &req); err != nil {
		h.WriteJSON(w, http.StatusBadRequest, &Response{Status: "error", Error: "Invalid request"})
		return
	}

	// Set default date to now
	date := time.Now()
	if req.Date != nil {
		date = *req.Date
	}

	ucReq := &usecase.CreateRequest{
		UserID:      req.UserID,
		Description: req.Description,
		Amount:      req.Amount,
		CategoryID:  req.CategoryID,
		Date:        date,
	}

	resp, err := h.createExpenseUC.Execute(ctx, ucReq)
	if err != nil {
		h.WriteJSON(w, http.StatusInternalServerError, &Response{Status: "error", Error: err.Error()})
		return
	}

	h.WriteJSON(w, http.StatusCreated, &Response{Status: "success", Data: resp})
}

// GetExpenses godoc
func (h *Handler) GetExpenses(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := r.URL.Query().Get("user_id")

	if userID == "" {
		h.WriteJSON(w, http.StatusBadRequest, &Response{Status: "error", Error: "user_id is required"})
		return
	}

	req := &usecase.GetAllRequest{UserID: userID}
	resp, err := h.getExpensesUC.ExecuteGetAll(ctx, req)
	if err != nil {
		h.WriteJSON(w, http.StatusInternalServerError, &Response{Status: "error", Error: err.Error()})
		return
	}

	h.WriteJSON(w, http.StatusOK, &Response{Status: "success", Data: resp})
}

// GetCategories retrieves all categories for a user
func (h *Handler) GetCategories(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := r.URL.Query().Get("user_id")

	if userID == "" {
		h.WriteJSON(w, http.StatusBadRequest, &Response{Status: "error", Error: "user_id is required"})
		return
	}

	categories, err := h.categoryRepo.GetByUserID(ctx, userID)
	if err != nil {
		h.WriteJSON(w, http.StatusInternalServerError, &Response{Status: "error", Error: err.Error()})
		return
	}

	h.WriteJSON(w, http.StatusOK, &Response{Status: "success", Data: categories})
}

// GetMetricsDAU retrieves daily active users
func (h *Handler) GetMetricsDAU(w http.ResponseWriter, r *http.Request) {
	// Check authentication
	if !h.authenticateAdmin(r) {
		h.WriteJSON(w, http.StatusUnauthorized, &Response{Status: "error", Error: "Unauthorized"})
		return
	}

	ctx := r.Context()

	// Get last 30 days
	to := time.Now()
	from := to.AddDate(0, 0, -30)

	metrics, err := h.metricsRepo.GetDailyActiveUsers(ctx, from, to)
	if err != nil {
		h.WriteJSON(w, http.StatusInternalServerError, &Response{Status: "error", Error: err.Error()})
		return
	}

	h.WriteJSON(w, http.StatusOK, &Response{Status: "success", Data: metrics})
}

// GetMetricsExpenses retrieves expense summary
func (h *Handler) GetMetricsExpenses(w http.ResponseWriter, r *http.Request) {
	if !h.authenticateAdmin(r) {
		h.WriteJSON(w, http.StatusUnauthorized, &Response{Status: "error", Error: "Unauthorized"})
		return
	}

	ctx := r.Context()

	to := time.Now()
	from := to.AddDate(0, 0, -30)

	metrics, err := h.metricsRepo.GetExpensesSummary(ctx, from, to)
	if err != nil {
		h.WriteJSON(w, http.StatusInternalServerError, &Response{Status: "error", Error: err.Error()})
		return
	}

	h.WriteJSON(w, http.StatusOK, &Response{Status: "success", Data: metrics})
}

// GetMetricsGrowth retrieves growth metrics
func (h *Handler) GetMetricsGrowth(w http.ResponseWriter, r *http.Request) {
	if !h.authenticateAdmin(r) {
		h.WriteJSON(w, http.StatusUnauthorized, &Response{Status: "error", Error: "Unauthorized"})
		return
	}

	ctx := r.Context()

	metrics, err := h.metricsRepo.GetGrowthMetrics(ctx, 30)
	if err != nil {
		h.WriteJSON(w, http.StatusInternalServerError, &Response{Status: "error", Error: err.Error()})
		return
	}

	h.WriteJSON(w, http.StatusOK, &Response{Status: "success", Data: metrics})
}

// authenticateAdmin checks if request has valid admin API key
func (h *Handler) authenticateAdmin(r *http.Request) bool {
	if h.adminAPIKey == "" {
		return true // No auth required if key not set
	}

	key := r.Header.Get("X-API-Key")
	return key == h.adminAPIKey
}

// Health check
func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	h.WriteJSON(w, http.StatusOK, &Response{Status: "ok"})
}

// RegisterRoutes registers all HTTP routes
func RegisterRoutes(mux *http.ServeMux, handler *Handler) {
	mux.HandleFunc("POST /api/users/auto-signup", handler.AutoSignup)
	mux.HandleFunc("POST /api/expenses/parse", handler.ParseExpenses)
	mux.HandleFunc("POST /api/expenses", handler.CreateExpense)
	mux.HandleFunc("GET /api/expenses", handler.GetExpenses)
	mux.HandleFunc("GET /api/categories", handler.GetCategories)
	mux.HandleFunc("GET /api/metrics/dau", handler.GetMetricsDAU)
	mux.HandleFunc("GET /api/metrics/expenses-summary", handler.GetMetricsExpenses)
	mux.HandleFunc("GET /api/metrics/growth", handler.GetMetricsGrowth)
	mux.HandleFunc("GET /health", handler.Health)
}
