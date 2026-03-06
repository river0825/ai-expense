package http

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/riverlin/aiexpense/internal/domain"
	"github.com/riverlin/aiexpense/internal/pkg/jwtutil"
	"github.com/riverlin/aiexpense/internal/usecase"
)

// Handler holds all HTTP request handlers
type Handler struct {
	autoSignupUC          *usecase.AutoSignupUseCase
	parseConversationUC   *usecase.ParseConversationUseCase
	createExpenseUC       *usecase.CreateExpenseUseCase
	getExpensesUC         *usecase.GetExpensesUseCase
	updateExpenseUC       *usecase.UpdateExpenseUseCase
	deleteExpenseUC       *usecase.DeleteExpenseUseCase
	manageCategoryUC      *usecase.ManageCategoryUseCase
	generateReportUC      *usecase.GenerateReportUseCase
	budgetManagementUC    *usecase.BudgetManagementUseCase
	dataExportUC          *usecase.DataExportUseCase
	recurringExpenseUC    *usecase.RecurringExpenseUseCase
	notificationUC        *usecase.NotificationUseCase
	searchExpenseUC       *usecase.SearchExpenseUseCase
	archiveUC             *usecase.ArchiveUseCase
	metricsUC             *usecase.MetricsUseCase
	getPolicyUC           *usecase.GetPolicyUseCase
	getUserAggregateUC    *usecase.GetUserAggregateUseCase
	updateUserAggregateUC *usecase.UpdateUserAggregateUseCase
	adminVerifyTokenUC    *usecase.AdminVerifyTokenUseCase
	exchangeRateSvc       domain.ExchangeRateService
	userRepo              domain.UserRepository
	categoryRepo          domain.CategoryRepository
	expenseRepo           domain.ExpenseRepository
	metricsRepo           domain.MetricsRepository
	adminAPIKey           string
	tokenManager          *jwtutil.TokenManager
	isDev                 bool
}

// NewHandler creates a new HTTP handler
func NewHandler(
	autoSignupUC *usecase.AutoSignupUseCase,
	parseConversationUC *usecase.ParseConversationUseCase,
	createExpenseUC *usecase.CreateExpenseUseCase,
	getExpensesUC *usecase.GetExpensesUseCase,
	updateExpenseUC *usecase.UpdateExpenseUseCase,
	deleteExpenseUC *usecase.DeleteExpenseUseCase,
	manageCategoryUC *usecase.ManageCategoryUseCase,
	generateReportUC *usecase.GenerateReportUseCase,
	budgetManagementUC *usecase.BudgetManagementUseCase,
	dataExportUC *usecase.DataExportUseCase,
	recurringExpenseUC *usecase.RecurringExpenseUseCase,
	notificationUC *usecase.NotificationUseCase,
	searchExpenseUC *usecase.SearchExpenseUseCase,
	archiveUC *usecase.ArchiveUseCase,
	metricsUC *usecase.MetricsUseCase,
	getPolicyUC *usecase.GetPolicyUseCase,
	getUserAggregateUC *usecase.GetUserAggregateUseCase,
	updateUserAggregateUC *usecase.UpdateUserAggregateUseCase,
	adminVerifyTokenUC *usecase.AdminVerifyTokenUseCase,
	exchangeRateSvc domain.ExchangeRateService,
	userRepo domain.UserRepository,
	categoryRepo domain.CategoryRepository,
	expenseRepo domain.ExpenseRepository,
	metricsRepo domain.MetricsRepository,
	adminAPIKey string,
	tokenManager *jwtutil.TokenManager,
	isDev bool,
) *Handler {
	h := &Handler{
		autoSignupUC:          autoSignupUC,
		parseConversationUC:   parseConversationUC,
		createExpenseUC:       createExpenseUC,
		getExpensesUC:         getExpensesUC,
		updateExpenseUC:       updateExpenseUC,
		deleteExpenseUC:       deleteExpenseUC,
		manageCategoryUC:      manageCategoryUC,
		generateReportUC:      generateReportUC,
		budgetManagementUC:    budgetManagementUC,
		dataExportUC:          dataExportUC,
		recurringExpenseUC:    recurringExpenseUC,
		notificationUC:        notificationUC,
		searchExpenseUC:       searchExpenseUC,
		archiveUC:             archiveUC,
		metricsUC:             metricsUC,
		getPolicyUC:           getPolicyUC,
		getUserAggregateUC:    getUserAggregateUC,
		updateUserAggregateUC: updateUserAggregateUC,
		adminVerifyTokenUC:    adminVerifyTokenUC,
		exchangeRateSvc:       exchangeRateSvc,
		userRepo:              userRepo,
		categoryRepo:          categoryRepo,
		expenseRepo:           expenseRepo,
		metricsRepo:           metricsRepo,
		adminAPIKey:           adminAPIKey,
		tokenManager:          tokenManager,
		isDev:                 isDev,
	}
	return h
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

	// Validate required fields
	if req.UserID == "" || req.MessengerType == "" {
		h.WriteJSON(w, http.StatusBadRequest, &Response{Status: "error", Error: "Missing required fields: user_id and messenger_type"})
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
	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		h.WriteJSON(w, http.StatusUnauthorized, &Response{Status: "error", Error: "Unauthorized"})
		return
	}

	type ParseRequest struct {
		Text string `json:"text"`
	}

	var req ParseRequest
	if err := h.ReadJSON(r, &req); err != nil {
		h.WriteJSON(w, http.StatusBadRequest, &Response{Status: "error", Error: "Invalid request"})
		return
	}

	expenses, err := h.parseConversationUC.Execute(ctx, req.Text, userID)
	if err != nil {
		h.WriteJSON(w, http.StatusInternalServerError, &Response{Status: "error", Error: err.Error()})
		return
	}

	h.WriteJSON(w, http.StatusOK, &Response{Status: "success", Data: expenses})
}

// CreateExpense godoc
func (h *Handler) CreateExpense(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		h.WriteJSON(w, http.StatusUnauthorized, &Response{Status: "error", Error: "Unauthorized"})
		return
	}

	type CreateRequest struct {
		Description      string     `json:"description"`
		Amount           float64    `json:"amount"`
		Currency         string     `json:"currency,omitempty"`
		CurrencyOriginal string     `json:"currency_original,omitempty"`
		ConvertedAmount  float64    `json:"converted_amount,omitempty"`
		HomeCurrency     string     `json:"home_currency,omitempty"`
		ExchangeRate     float64    `json:"exchange_rate,omitempty"`
		CategoryID       *string    `json:"category_id,omitempty"`
		Account          string     `json:"account,omitempty"`
		Date             *time.Time `json:"date,omitempty"`
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
		UserID:           userID,
		Description:      req.Description,
		Amount:           req.Amount,
		Currency:         req.Currency,
		CurrencyOriginal: req.CurrencyOriginal,
		ConvertedAmount:  req.ConvertedAmount,
		HomeCurrency:     req.HomeCurrency,
		ExchangeRate:     req.ExchangeRate,
		CategoryID:       req.CategoryID,
		Account:          req.Account,
		Date:             date,
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
	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		h.WriteJSON(w, http.StatusUnauthorized, &Response{Status: "error", Error: "Unauthorized"})
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
	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		h.WriteJSON(w, http.StatusUnauthorized, &Response{Status: "error", Error: "Unauthorized"})
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
	// Authentication is handled by AdminMiddleware
	ctx := r.Context()

	resp, err := h.metricsUC.GetDailyActiveUsers(ctx, &usecase.DailyActiveUsersRequest{Days: 30})
	if err != nil {
		h.WriteJSON(w, http.StatusInternalServerError, &Response{Status: "error", Error: err.Error()})
		return
	}

	h.WriteJSON(w, http.StatusOK, &Response{Status: "success", Data: resp})
}

// GetMetricsExpenses retrieves expense summary
func (h *Handler) GetMetricsExpenses(w http.ResponseWriter, r *http.Request) {
	// Authentication is handled by AdminMiddleware
	ctx := r.Context()

	resp, err := h.metricsUC.GetExpensesSummary(ctx, &usecase.ExpensesSummaryRequest{Days: 30})
	if err != nil {
		h.WriteJSON(w, http.StatusInternalServerError, &Response{Status: "error", Error: err.Error()})
		return
	}

	h.WriteJSON(w, http.StatusOK, &Response{Status: "success", Data: resp})
}

// GetMetricsGrowth retrieves growth metrics
func (h *Handler) GetMetricsGrowth(w http.ResponseWriter, r *http.Request) {
	// Authentication is handled by AdminMiddleware
	ctx := r.Context()

	resp, err := h.metricsUC.GetGrowthMetrics(ctx, &usecase.GrowthMetricsRequest{Days: 30})
	if err != nil {
		h.WriteJSON(w, http.StatusInternalServerError, &Response{Status: "error", Error: err.Error()})
		return
	}

	h.WriteJSON(w, http.StatusOK, &Response{Status: "success", Data: resp})
}

// RefreshExchangeRates triggers a manual exchange rate refresh
func (h *Handler) RefreshExchangeRates(w http.ResponseWriter, r *http.Request) {
	// Authentication is handled by AdminMiddleware

	if h.exchangeRateSvc == nil {
		h.WriteJSON(w, http.StatusServiceUnavailable, &Response{Status: "error", Error: "Exchange rate service not configured"})
		return
	}

	ctx := r.Context()
	if err := h.exchangeRateSvc.RefreshRates(ctx); err != nil {
		h.WriteJSON(w, http.StatusInternalServerError, &Response{Status: "error", Error: err.Error()})
		return
	}

	h.WriteJSON(w, http.StatusOK, &Response{Status: "success", Message: "Exchange rates refreshed"})
}

// authenticateAdmin checks if request has valid admin API key
func (h *Handler) authenticateAdmin(r *http.Request) bool {
	if h.adminAPIKey == "" {
		return true // No auth required if key not set
	}

	key := r.Header.Get("X-API-Key")
	return key == h.adminAPIKey
}

// UpdateExpense godoc
func (h *Handler) UpdateExpense(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		h.WriteJSON(w, http.StatusUnauthorized, &Response{Status: "error", Error: "Unauthorized"})
		return
	}

	type UpdateExpenseRequest struct {
		ID             string     `json:"id"`
		Description    *string    `json:"description,omitempty"`
		Amount         *float64   `json:"amount,omitempty"`
		OriginalAmount *float64   `json:"original_amount,omitempty"`
		Currency       *string    `json:"currency,omitempty"`
		CategoryID     *string    `json:"category_id,omitempty"`
		Account        *string    `json:"account,omitempty"`
		ExpenseDate    *time.Time `json:"expense_date,omitempty"`
	}

	var req UpdateExpenseRequest
	if err := h.ReadJSON(r, &req); err != nil {
		h.WriteJSON(w, http.StatusBadRequest, &Response{Status: "error", Error: "Invalid request"})
		return
	}

	if req.ID == "" {
		h.WriteJSON(w, http.StatusBadRequest, &Response{Status: "error", Error: "id is required"})
		return
	}

	originalAmount := req.OriginalAmount
	if originalAmount == nil {
		originalAmount = req.Amount
	}

	resp, err := h.updateExpenseUC.Execute(ctx, &usecase.UpdateRequest{
		ID:             req.ID,
		UserID:         userID,
		Description:    req.Description,
		OriginalAmount: originalAmount,
		Currency:       req.Currency,
		CategoryID:     req.CategoryID,
		Account:        req.Account,
		ExpenseDate:    req.ExpenseDate,
	})

	if err != nil {
		h.WriteJSON(w, http.StatusBadRequest, &Response{Status: "error", Error: err.Error()})
		return
	}

	h.WriteJSON(w, http.StatusOK, &Response{Status: "success", Data: resp})
}

// DeleteExpense godoc
func (h *Handler) DeleteExpense(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		h.WriteJSON(w, http.StatusUnauthorized, &Response{Status: "error", Error: "Unauthorized"})
		return
	}

	type DeleteExpenseRequest struct {
		ID string `json:"id"`
	}

	var req DeleteExpenseRequest
	if err := h.ReadJSON(r, &req); err != nil {
		h.WriteJSON(w, http.StatusBadRequest, &Response{Status: "error", Error: "Invalid request"})
		return
	}

	if req.ID == "" {
		h.WriteJSON(w, http.StatusBadRequest, &Response{Status: "error", Error: "id is required"})
		return
	}

	resp, err := h.deleteExpenseUC.Execute(ctx, &usecase.DeleteRequest{
		ID:     req.ID,
		UserID: userID,
	})

	if err != nil {
		h.WriteJSON(w, http.StatusBadRequest, &Response{Status: "error", Error: err.Error()})
		return
	}

	h.WriteJSON(w, http.StatusOK, &Response{Status: "success", Data: resp})
}

// CreateCategory godoc
func (h *Handler) CreateCategory(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		h.WriteJSON(w, http.StatusUnauthorized, &Response{Status: "error", Error: "Unauthorized"})
		return
	}

	type CreateCategoryRequest struct {
		Name        string   `json:"name"`
		Description string   `json:"description,omitempty"`
		Keywords    []string `json:"keywords,omitempty"`
	}

	var req CreateCategoryRequest
	if err := h.ReadJSON(r, &req); err != nil {
		h.WriteJSON(w, http.StatusBadRequest, &Response{Status: "error", Error: "Invalid request"})
		return
	}

	if req.Name == "" {
		h.WriteJSON(w, http.StatusBadRequest, &Response{Status: "error", Error: "name is required"})
		return
	}

	resp, err := h.manageCategoryUC.CreateCategory(ctx, &usecase.CreateCategoryRequest{
		UserID:      userID,
		Name:        req.Name,
		Description: req.Description,
		Keywords:    req.Keywords,
	})

	if err != nil {
		h.WriteJSON(w, http.StatusBadRequest, &Response{Status: "error", Error: err.Error()})
		return
	}

	h.WriteJSON(w, http.StatusOK, &Response{Status: "success", Data: resp})
}

// UpdateCategory godoc
func (h *Handler) UpdateCategory(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		h.WriteJSON(w, http.StatusUnauthorized, &Response{Status: "error", Error: "Unauthorized"})
		return
	}

	type UpdateCategoryRequest struct {
		ID          string   `json:"id"`
		Name        *string  `json:"name,omitempty"`
		Description *string  `json:"description,omitempty"`
		Keywords    []string `json:"keywords,omitempty"`
	}

	var req UpdateCategoryRequest
	if err := h.ReadJSON(r, &req); err != nil {
		h.WriteJSON(w, http.StatusBadRequest, &Response{Status: "error", Error: "Invalid request"})
		return
	}

	if req.ID == "" {
		h.WriteJSON(w, http.StatusBadRequest, &Response{Status: "error", Error: "id is required"})
		return
	}

	resp, err := h.manageCategoryUC.UpdateCategory(ctx, &usecase.UpdateCategoryRequest{
		UserID:      userID,
		ID:          req.ID,
		Name:        req.Name,
		Description: req.Description,
		Keywords:    req.Keywords,
	})

	if err != nil {
		h.WriteJSON(w, http.StatusBadRequest, &Response{Status: "error", Error: err.Error()})
		return
	}

	h.WriteJSON(w, http.StatusOK, &Response{Status: "success", Data: resp})
}

// DeleteCategory godoc
func (h *Handler) DeleteCategory(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		h.WriteJSON(w, http.StatusUnauthorized, &Response{Status: "error", Error: "Unauthorized"})
		return
	}

	type DeleteCategoryRequest struct {
		ID string `json:"id"`
	}

	var req DeleteCategoryRequest
	if err := h.ReadJSON(r, &req); err != nil {
		h.WriteJSON(w, http.StatusBadRequest, &Response{Status: "error", Error: "Invalid request"})
		return
	}

	if req.ID == "" {
		h.WriteJSON(w, http.StatusBadRequest, &Response{Status: "error", Error: "id is required"})
		return
	}

	resp, err := h.manageCategoryUC.DeleteCategory(ctx, &usecase.DeleteCategoryRequest{
		UserID: userID,
		ID:     req.ID,
	})

	if err != nil {
		h.WriteJSON(w, http.StatusBadRequest, &Response{Status: "error", Error: err.Error()})
		return
	}

	h.WriteJSON(w, http.StatusOK, &Response{Status: "success", Data: resp})
}

// ListCategories godoc
func (h *Handler) ListCategories(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		h.WriteJSON(w, http.StatusUnauthorized, &Response{Status: "error", Error: "Unauthorized"})
		return
	}

	resp, err := h.manageCategoryUC.ListCategories(ctx, &usecase.ListCategoriesRequest{
		UserID: userID,
	})

	if err != nil {
		h.WriteJSON(w, http.StatusInternalServerError, &Response{Status: "error", Error: err.Error()})
		return
	}

	h.WriteJSON(w, http.StatusOK, &Response{Status: "success", Data: resp.Categories})
}

// MergeCategories godoc
func (h *Handler) MergeCategories(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		h.WriteJSON(w, http.StatusUnauthorized, &Response{Status: "error", Error: "Unauthorized"})
		return
	}

	type MergeCategoriesRequest struct {
		SourceID string `json:"source_id"`
		TargetID string `json:"target_id"`
	}

	var req MergeCategoriesRequest
	if err := h.ReadJSON(r, &req); err != nil {
		h.WriteJSON(w, http.StatusBadRequest, &Response{Status: "error", Error: "Invalid request"})
		return
	}

	if req.SourceID == "" || req.TargetID == "" {
		h.WriteJSON(w, http.StatusBadRequest, &Response{Status: "error", Error: "source_id and target_id are required"})
		return
	}

	resp, err := h.manageCategoryUC.MergeCategories(ctx, &usecase.MergeCategoriesRequest{
		UserID:   userID,
		SourceID: req.SourceID,
		TargetID: req.TargetID,
	})

	if err != nil {
		h.WriteJSON(w, http.StatusBadRequest, &Response{Status: "error", Error: err.Error()})
		return
	}

	h.WriteJSON(w, http.StatusOK, &Response{Status: "success", Data: resp})
}

// GenerateReport godoc
func (h *Handler) GenerateReport(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		h.WriteJSON(w, http.StatusUnauthorized, &Response{Status: "error", Error: "Unauthorized"})
		return
	}

	type GenerateReportRequest struct {
		ReportType string    `json:"report_type"`
		StartDate  time.Time `json:"start_date"`
		EndDate    time.Time `json:"end_date"`
	}

	var req GenerateReportRequest
	if err := h.ReadJSON(r, &req); err != nil {
		h.WriteJSON(w, http.StatusBadRequest, &Response{Status: "error", Error: "Invalid request"})
		return
	}

	if req.ReportType == "" {
		req.ReportType = "monthly"
	}

	if req.StartDate.IsZero() {
		req.StartDate = time.Now().AddDate(0, -1, 0)
	}

	if req.EndDate.IsZero() {
		req.EndDate = time.Now()
	}

	resp, err := h.generateReportUC.Execute(ctx, &usecase.ReportRequest{
		UserID:     userID,
		ReportType: req.ReportType,
		StartDate:  req.StartDate,
		EndDate:    req.EndDate,
	})

	if err != nil {
		h.WriteJSON(w, http.StatusInternalServerError, &Response{Status: "error", Error: err.Error()})
		return
	}

	h.WriteJSON(w, http.StatusOK, &Response{Status: "success", Data: resp})
}

// GetBudgetStatus godoc
// GetBudgetStatus godoc
func (h *Handler) GetBudgetStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		h.WriteJSON(w, http.StatusUnauthorized, &Response{Status: "error", Error: "Unauthorized"})
		return
	}

	resp, err := h.budgetManagementUC.GetBudgetStatus(ctx, &usecase.GetBudgetStatusRequest{
		UserID: userID,
	})

	if err != nil {
		h.WriteJSON(w, http.StatusInternalServerError, &Response{Status: "error", Error: err.Error()})
		return
	}

	h.WriteJSON(w, http.StatusOK, &Response{Status: "success", Data: resp})
}

// CompareToBudget godoc
func (h *Handler) CompareToBudget(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		h.WriteJSON(w, http.StatusUnauthorized, &Response{Status: "error", Error: "Unauthorized"})
		return
	}

	categoryID := r.URL.Query().Get("category_id")
	period := r.URL.Query().Get("period")

	var category *string
	if categoryID != "" {
		category = &categoryID
	}

	resp, err := h.budgetManagementUC.CompareToBudget(ctx, &usecase.CompareToBudgetRequest{
		UserID:     userID,
		CategoryID: category,
		Period:     period,
	})

	if err != nil {
		h.WriteJSON(w, http.StatusInternalServerError, &Response{Status: "error", Error: err.Error()})
		return
	}

	h.WriteJSON(w, http.StatusOK, &Response{Status: "success", Data: resp})
}

// ExportExpenses godoc
func (h *Handler) ExportExpenses(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		h.WriteJSON(w, http.StatusUnauthorized, &Response{Status: "error", Error: "Unauthorized"})
		return
	}

	format := r.URL.Query().Get("format")
	startDate := r.URL.Query().Get("start_date")
	endDate := r.URL.Query().Get("end_date")

	if format == "" {
		format = "json"
	}

	// Parse dates
	var start, end time.Time
	if startDate != "" {
		start, _ = time.Parse("2006-01-02", startDate)
	} else {
		start = time.Now().AddDate(-1, 0, 0)
	}

	if endDate != "" {
		end, _ = time.Parse("2006-01-02", endDate)
	} else {
		end = time.Now()
	}

	req := &usecase.ExportRequest{
		UserID:    userID,
		Format:    format,
		StartDate: start,
		EndDate:   end,
	}

	if format == "csv" {
		data, err := h.dataExportUC.ExportAsCSV(ctx, req)
		if err != nil {
			h.WriteJSON(w, http.StatusInternalServerError, &Response{Status: "error", Error: err.Error()})
			return
		}

		w.Header().Set("Content-Type", "text/csv")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-API-Key, Authorization")
		w.Header().Set("Content-Disposition", "attachment; filename=expenses.csv")
		w.WriteHeader(http.StatusOK)
		w.Write(data)
	} else {
		data, err := h.dataExportUC.ExportAsJSON(ctx, req)
		if err != nil {
			h.WriteJSON(w, http.StatusInternalServerError, &Response{Status: "error", Error: err.Error()})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(data)
	}
}

// ExportSummary godoc
func (h *Handler) ExportSummary(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		h.WriteJSON(w, http.StatusUnauthorized, &Response{Status: "error", Error: "Unauthorized"})
		return
	}

	startDate := r.URL.Query().Get("start_date")
	endDate := r.URL.Query().Get("end_date")

	// Parse dates
	var start, end time.Time
	if startDate != "" {
		start, _ = time.Parse("2006-01-02", startDate)
	} else {
		start = time.Now().AddDate(-1, 0, 0)
	}

	if endDate != "" {
		end, _ = time.Parse("2006-01-02", endDate)
	} else {
		end = time.Now()
	}

	resp, err := h.dataExportUC.ExportSummary(ctx, &usecase.SummaryExportRequest{
		UserID:    userID,
		StartDate: start,
		EndDate:   end,
	})

	if err != nil {
		h.WriteJSON(w, http.StatusInternalServerError, &Response{Status: "error", Error: err.Error()})
		return
	}

	h.WriteJSON(w, http.StatusOK, &Response{Status: "success", Data: resp})
}

// SearchExpenses godoc
func (h *Handler) SearchExpenses(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		h.WriteJSON(w, http.StatusUnauthorized, &Response{Status: "error", Error: "Unauthorized"})
		return
	}

	query := r.URL.Query().Get("q")
	categoryID := r.URL.Query().Get("category_id")
	sortBy := r.URL.Query().Get("sort_by")
	limit := 20

	var category *string
	if categoryID != "" {
		category = &categoryID
	}

	resp, err := h.searchExpenseUC.Search(ctx, &usecase.SearchRequest{
		UserID:     userID,
		Query:      query,
		CategoryID: category,
		SortBy:     sortBy,
		Limit:      limit,
		Offset:     0,
	})

	if err != nil {
		h.WriteJSON(w, http.StatusInternalServerError, &Response{Status: "error", Error: err.Error()})
		return
	}

	h.WriteJSON(w, http.StatusOK, &Response{Status: "success", Data: resp})
}

// FilterExpenses godoc
func (h *Handler) FilterExpenses(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		h.WriteJSON(w, http.StatusUnauthorized, &Response{Status: "error", Error: "Unauthorized"})
		return
	}

	period := r.URL.Query().Get("period")
	categoryID := r.URL.Query().Get("category_id")

	resp, err := h.searchExpenseUC.Filter(ctx, &usecase.FilterRequest{
		UserID:     userID,
		CategoryID: categoryID,
		Period:     period,
	})

	if err != nil {
		h.WriteJSON(w, http.StatusInternalServerError, &Response{Status: "error", Error: err.Error()})
		return
	}

	h.WriteJSON(w, http.StatusOK, &Response{Status: "success", Data: resp})
}

// CreateRecurring godoc
func (h *Handler) CreateRecurring(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		h.WriteJSON(w, http.StatusUnauthorized, &Response{Status: "error", Error: "Unauthorized"})
		return
	}

	type CreateRecurringRequest struct {
		Description string    `json:"description"`
		Amount      float64   `json:"amount"`
		CategoryID  *string   `json:"category_id,omitempty"`
		Frequency   string    `json:"frequency"`
		StartDate   time.Time `json:"start_date"`
	}

	var req CreateRecurringRequest
	if err := h.ReadJSON(r, &req); err != nil {
		h.WriteJSON(w, http.StatusBadRequest, &Response{Status: "error", Error: "Invalid request"})
		return
	}

	resp, err := h.recurringExpenseUC.CreateRecurring(ctx, &usecase.CreateRecurringRequest{
		UserID:      userID,
		Description: req.Description,
		Amount:      req.Amount,
		CategoryID:  req.CategoryID,
		Frequency:   req.Frequency,
		StartDate:   req.StartDate,
	})

	if err != nil {
		h.WriteJSON(w, http.StatusBadRequest, &Response{Status: "error", Error: err.Error()})
		return
	}

	h.WriteJSON(w, http.StatusOK, &Response{Status: "success", Data: resp})
}

// ListRecurring godoc
func (h *Handler) ListRecurring(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		h.WriteJSON(w, http.StatusUnauthorized, &Response{Status: "error", Error: "Unauthorized"})
		return
	}

	resp, err := h.recurringExpenseUC.ListRecurring(ctx, &usecase.ListRecurringRequest{
		UserID: userID,
	})

	if err != nil {
		h.WriteJSON(w, http.StatusInternalServerError, &Response{Status: "error", Error: err.Error()})
		return
	}

	h.WriteJSON(w, http.StatusOK, &Response{Status: "success", Data: resp})
}

// UpdateRecurring godoc
func (h *Handler) UpdateRecurring(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		h.WriteJSON(w, http.StatusUnauthorized, &Response{Status: "error", Error: "Unauthorized"})
		return
	}

	type UpdateRecurringRequest struct {
		ID          string   `json:"id"`
		Description *string  `json:"description,omitempty"`
		Amount      *float64 `json:"amount,omitempty"`
		Frequency   *string  `json:"frequency,omitempty"`
	}

	var req UpdateRecurringRequest
	if err := h.ReadJSON(r, &req); err != nil {
		h.WriteJSON(w, http.StatusBadRequest, &Response{Status: "error", Error: "Invalid request"})
		return
	}

	resp, err := h.recurringExpenseUC.UpdateRecurring(ctx, &usecase.UpdateRecurringRequest{
		UserID:      userID,
		ID:          req.ID,
		Description: req.Description,
		Amount:      req.Amount,
		Frequency:   req.Frequency,
	})

	if err != nil {
		h.WriteJSON(w, http.StatusBadRequest, &Response{Status: "error", Error: err.Error()})
		return
	}

	h.WriteJSON(w, http.StatusOK, &Response{Status: "success", Data: resp})
}

// DeleteRecurring godoc
func (h *Handler) DeleteRecurring(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		h.WriteJSON(w, http.StatusUnauthorized, &Response{Status: "error", Error: "Unauthorized"})
		return
	}
	id := r.URL.Query().Get("id")

	if id == "" {
		h.WriteJSON(w, http.StatusBadRequest, &Response{Status: "error", Error: "id is required"})
		return
	}

	resp, err := h.recurringExpenseUC.DeleteRecurring(ctx, &usecase.DeleteRecurringRequest{
		UserID: userID,
		ID:     id,
	})

	if err != nil {
		h.WriteJSON(w, http.StatusBadRequest, &Response{Status: "error", Error: err.Error()})
		return
	}

	h.WriteJSON(w, http.StatusOK, &Response{Status: "success", Data: resp})
}

// GetUpcomingRecurring godoc
func (h *Handler) GetUpcomingRecurring(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		h.WriteJSON(w, http.StatusUnauthorized, &Response{Status: "error", Error: "Unauthorized"})
		return
	}

	resp, err := h.recurringExpenseUC.GetUpcoming(ctx, &usecase.GetUpcomingRequest{
		UserID: userID,
		Days:   30,
	})

	if err != nil {
		h.WriteJSON(w, http.StatusInternalServerError, &Response{Status: "error", Error: err.Error()})
		return
	}

	h.WriteJSON(w, http.StatusOK, &Response{Status: "success", Data: resp})
}

// ProcessRecurring godoc
func (h *Handler) ProcessRecurring(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		h.WriteJSON(w, http.StatusUnauthorized, &Response{Status: "error", Error: "Unauthorized"})
		return
	}

	type ProcessRecurringRequest struct {
		Date time.Time `json:"date"`
	}

	var req ProcessRecurringRequest
	if err := h.ReadJSON(r, &req); err != nil {
		h.WriteJSON(w, http.StatusBadRequest, &Response{Status: "error", Error: "Invalid request"})
		return
	}

	resp, err := h.recurringExpenseUC.ProcessRecurring(ctx, &usecase.ProcessRecurringRequest{
		UserID: userID,
		Date:   req.Date,
	})

	if err != nil {
		h.WriteJSON(w, http.StatusInternalServerError, &Response{Status: "error", Error: err.Error()})
		return
	}

	h.WriteJSON(w, http.StatusOK, &Response{Status: "success", Data: resp})
}

// CreateNotification godoc
func (h *Handler) CreateNotification(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		h.WriteJSON(w, http.StatusUnauthorized, &Response{Status: "error", Error: "Unauthorized"})
		return
	}

	type CreateNotificationRequest struct {
		Type    string                 `json:"type"`
		Title   string                 `json:"title"`
		Message string                 `json:"message"`
		Data    map[string]interface{} `json:"data,omitempty"`
	}

	var req CreateNotificationRequest
	if err := h.ReadJSON(r, &req); err != nil {
		h.WriteJSON(w, http.StatusBadRequest, &Response{Status: "error", Error: "Invalid request"})
		return
	}

	resp, err := h.notificationUC.CreateNotification(ctx, &usecase.CreateNotificationRequest{
		UserID:  userID,
		Type:    req.Type,
		Title:   req.Title,
		Message: req.Message,
		Data:    req.Data,
	})

	if err != nil {
		h.WriteJSON(w, http.StatusBadRequest, &Response{Status: "error", Error: err.Error()})
		return
	}

	h.WriteJSON(w, http.StatusOK, &Response{Status: "success", Data: resp})
}

// ListNotifications godoc
func (h *Handler) ListNotifications(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		h.WriteJSON(w, http.StatusUnauthorized, &Response{Status: "error", Error: "Unauthorized"})
		return
	}

	resp, err := h.notificationUC.ListNotifications(ctx, &usecase.ListNotificationsRequest{
		UserID: userID,
		Limit:  20,
	})

	if err != nil {
		h.WriteJSON(w, http.StatusInternalServerError, &Response{Status: "error", Error: err.Error()})
		return
	}

	h.WriteJSON(w, http.StatusOK, &Response{Status: "success", Data: resp})
}

// MarkNotificationAsRead godoc
func (h *Handler) MarkNotificationAsRead(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		h.WriteJSON(w, http.StatusUnauthorized, &Response{Status: "error", Error: "Unauthorized"})
		return
	}

	type MarkAsReadRequest struct {
		NotificationID string `json:"notification_id"`
	}

	var req MarkAsReadRequest
	if err := h.ReadJSON(r, &req); err != nil {
		h.WriteJSON(w, http.StatusBadRequest, &Response{Status: "error", Error: "Invalid request"})
		return
	}

	resp, err := h.notificationUC.MarkAsRead(ctx, &usecase.MarkAsReadRequest{
		UserID:         userID,
		NotificationID: req.NotificationID,
	})

	if err != nil {
		h.WriteJSON(w, http.StatusBadRequest, &Response{Status: "error", Error: err.Error()})
		return
	}

	h.WriteJSON(w, http.StatusOK, &Response{Status: "success", Data: resp})
}

// MarkAllNotificationsAsRead godoc
func (h *Handler) MarkAllNotificationsAsRead(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		h.WriteJSON(w, http.StatusUnauthorized, &Response{Status: "error", Error: "Unauthorized"})
		return
	}

	resp, err := h.notificationUC.MarkAllAsRead(ctx, &usecase.MarkAllAsReadRequest{
		UserID: userID,
	})

	if err != nil {
		h.WriteJSON(w, http.StatusBadRequest, &Response{Status: "error", Error: err.Error()})
		return
	}

	h.WriteJSON(w, http.StatusOK, &Response{Status: "success", Data: resp})
}

// DeleteNotification godoc
func (h *Handler) DeleteNotification(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		h.WriteJSON(w, http.StatusUnauthorized, &Response{Status: "error", Error: "Unauthorized"})
		return
	}
	notificationID := r.URL.Query().Get("id")

	if notificationID == "" {
		h.WriteJSON(w, http.StatusBadRequest, &Response{Status: "error", Error: "id is required"})
		return
	}

	resp, err := h.notificationUC.DeleteNotification(ctx, &usecase.DeleteNotificationRequest{
		UserID:         userID,
		NotificationID: notificationID,
	})

	if err != nil {
		h.WriteJSON(w, http.StatusBadRequest, &Response{Status: "error", Error: err.Error()})
		return
	}

	h.WriteJSON(w, http.StatusOK, &Response{Status: "success", Data: resp})
}

// GetNotificationPreferences godoc
func (h *Handler) GetNotificationPreferences(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		h.WriteJSON(w, http.StatusUnauthorized, &Response{Status: "error", Error: "Unauthorized"})
		return
	}

	resp, err := h.notificationUC.GetPreferences(ctx, &usecase.GetPreferencesRequest{
		UserID: userID,
	})

	if err != nil {
		h.WriteJSON(w, http.StatusInternalServerError, &Response{Status: "error", Error: err.Error()})
		return
	}

	h.WriteJSON(w, http.StatusOK, &Response{Status: "success", Data: resp})
}

// UpdateNotificationPreferences godoc
func (h *Handler) UpdateNotificationPreferences(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		h.WriteJSON(w, http.StatusUnauthorized, &Response{Status: "error", Error: "Unauthorized"})
		return
	}

	type UpdatePreferencesRequest struct {
		BudgetAlerts        *bool `json:"budget_alerts,omitempty"`
		RecurringReminders  *bool `json:"recurring_reminders,omitempty"`
		ReportNotifications *bool `json:"report_notifications,omitempty"`
		ExpenseReminders    *bool `json:"expense_reminders,omitempty"`
		DailyDigest         *bool `json:"daily_digest,omitempty"`
		WeeklyReport        *bool `json:"weekly_report,omitempty"`
	}

	var req UpdatePreferencesRequest
	if err := h.ReadJSON(r, &req); err != nil {
		h.WriteJSON(w, http.StatusBadRequest, &Response{Status: "error", Error: "Invalid request"})
		return
	}

	resp, err := h.notificationUC.UpdatePreferences(ctx, &usecase.UpdatePreferencesRequest{
		UserID:              userID,
		BudgetAlerts:        req.BudgetAlerts,
		RecurringReminders:  req.RecurringReminders,
		ReportNotifications: req.ReportNotifications,
		ExpenseReminders:    req.ExpenseReminders,
		DailyDigest:         req.DailyDigest,
		WeeklyReport:        req.WeeklyReport,
	})

	if err != nil {
		h.WriteJSON(w, http.StatusBadRequest, &Response{Status: "error", Error: err.Error()})
		return
	}

	h.WriteJSON(w, http.StatusOK, &Response{Status: "success", Data: resp})
}

// CreateArchive godoc
func (h *Handler) CreateArchive(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		h.WriteJSON(w, http.StatusUnauthorized, &Response{Status: "error", Error: "Unauthorized"})
		return
	}

	type CreateArchiveRequest struct {
		Period        string    `json:"period"`
		StartDate     time.Time `json:"start_date"`
		EndDate       time.Time `json:"end_date"`
		RetentionDays int       `json:"retention_days,omitempty"`
	}

	var req CreateArchiveRequest
	if err := h.ReadJSON(r, &req); err != nil {
		h.WriteJSON(w, http.StatusBadRequest, &Response{Status: "error", Error: "Invalid request"})
		return
	}

	resp, err := h.archiveUC.CreateArchive(ctx, &usecase.CreateArchiveRequest{
		UserID:        userID,
		Period:        req.Period,
		StartDate:     req.StartDate,
		EndDate:       req.EndDate,
		RetentionDays: req.RetentionDays,
	})

	if err != nil {
		h.WriteJSON(w, http.StatusBadRequest, &Response{Status: "error", Error: err.Error()})
		return
	}

	h.WriteJSON(w, http.StatusOK, &Response{Status: "success", Data: resp})
}

// ListArchives godoc
func (h *Handler) ListArchives(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		h.WriteJSON(w, http.StatusUnauthorized, &Response{Status: "error", Error: "Unauthorized"})
		return
	}

	resp, err := h.archiveUC.ListArchives(ctx, &usecase.ListArchivesRequest{
		UserID: userID,
	})

	if err != nil {
		h.WriteJSON(w, http.StatusInternalServerError, &Response{Status: "error", Error: err.Error()})
		return
	}

	h.WriteJSON(w, http.StatusOK, &Response{Status: "success", Data: resp})
}

// GetArchiveStats godoc
func (h *Handler) GetArchiveStats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		h.WriteJSON(w, http.StatusUnauthorized, &Response{Status: "error", Error: "Unauthorized"})
		return
	}

	resp, err := h.archiveUC.GetStatistics(ctx, &usecase.ArchiveStatisticsRequest{
		UserID: userID,
	})

	if err != nil {
		h.WriteJSON(w, http.StatusInternalServerError, &Response{Status: "error", Error: err.Error()})
		return
	}

	h.WriteJSON(w, http.StatusOK, &Response{Status: "success", Data: resp})
}

// GetArchiveDetails godoc
func (h *Handler) GetArchiveDetails(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		h.WriteJSON(w, http.StatusUnauthorized, &Response{Status: "error", Error: "Unauthorized"})
		return
	}
	archiveID := r.URL.Query().Get("archive_id")

	if archiveID == "" {
		h.WriteJSON(w, http.StatusBadRequest, &Response{Status: "error", Error: "archive_id is required"})
		return
	}

	resp, err := h.archiveUC.GetArchive(ctx, &usecase.GetArchiveRequest{
		UserID:    userID,
		ArchiveID: archiveID,
	})

	if err != nil {
		h.WriteJSON(w, http.StatusInternalServerError, &Response{Status: "error", Error: err.Error()})
		return
	}

	h.WriteJSON(w, http.StatusOK, &Response{Status: "success", Data: resp})
}

// RestoreArchive godoc
func (h *Handler) RestoreArchive(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		h.WriteJSON(w, http.StatusUnauthorized, &Response{Status: "error", Error: "Unauthorized"})
		return
	}

	type RestoreArchiveRequest struct {
		ArchiveID string `json:"archive_id"`
		Strategy  string `json:"strategy,omitempty"`
	}

	var req RestoreArchiveRequest
	if err := h.ReadJSON(r, &req); err != nil {
		h.WriteJSON(w, http.StatusBadRequest, &Response{Status: "error", Error: "Invalid request"})
		return
	}

	resp, err := h.archiveUC.RestoreArchive(ctx, &usecase.RestoreArchiveRequest{
		UserID:    userID,
		ArchiveID: req.ArchiveID,
		Strategy:  req.Strategy,
	})

	if err != nil {
		h.WriteJSON(w, http.StatusBadRequest, &Response{Status: "error", Error: err.Error()})
		return
	}

	h.WriteJSON(w, http.StatusOK, &Response{Status: "success", Data: resp})
}

// PurgeArchive godoc
func (h *Handler) PurgeArchive(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	type PurgeArchiveRequest struct {
		UserID  string `json:"user_id"`
		DaysOld int    `json:"days_old"`
		KeepMin int    `json:"keep_min,omitempty"`
	}

	var req PurgeArchiveRequest
	if err := h.ReadJSON(r, &req); err != nil {
		h.WriteJSON(w, http.StatusBadRequest, &Response{Status: "error", Error: "Invalid request"})
		return
	}

	resp, err := h.archiveUC.PurgeArchive(ctx, &usecase.PurgeArchiveRequest{
		UserID:  req.UserID,
		DaysOld: req.DaysOld,
		KeepMin: req.KeepMin,
	})

	if err != nil {
		h.WriteJSON(w, http.StatusBadRequest, &Response{Status: "error", Error: err.Error()})
		return
	}

	h.WriteJSON(w, http.StatusOK, &Response{Status: "success", Data: resp})
}

// ExportArchive godoc
func (h *Handler) ExportArchive(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	type ExportArchiveRequest struct {
		UserID    string `json:"user_id"`
		ArchiveID string `json:"archive_id"`
		Format    string `json:"format,omitempty"`
	}

	var req ExportArchiveRequest
	if err := h.ReadJSON(r, &req); err != nil {
		h.WriteJSON(w, http.StatusBadRequest, &Response{Status: "error", Error: "Invalid request"})
		return
	}

	resp, err := h.archiveUC.ExportArchive(ctx, &usecase.ExportArchiveRequest{
		UserID:    req.UserID,
		ArchiveID: req.ArchiveID,
		Format:    req.Format,
	})

	if err != nil {
		h.WriteJSON(w, http.StatusBadRequest, &Response{Status: "error", Error: err.Error()})
		return
	}

	h.WriteJSON(w, http.StatusOK, &Response{Status: "success", Data: resp})
}

// GetUser godoc
func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		h.WriteJSON(w, http.StatusUnauthorized, &Response{Status: "error", Error: "Unauthorized"})
		return
	}

	user, err := h.userRepo.GetByID(ctx, userID)
	if err != nil {
		h.WriteJSON(w, http.StatusInternalServerError, &Response{Status: "error", Error: err.Error()})
		return
	}
	if user == nil {
		h.WriteJSON(w, http.StatusNotFound, &Response{Status: "error", Error: "User not found"})
		return
	}

	h.WriteJSON(w, http.StatusOK, &Response{Status: "success", Data: user})
}

// UpdateUserSettings godoc
func (h *Handler) UpdateUserSettings(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		h.WriteJSON(w, http.StatusUnauthorized, &Response{Status: "error", Error: "Unauthorized"})
		return
	}

	type UpdateSettingsRequest struct {
		HomeCurrency         string `json:"home_currency"`
		DefaultInputCurrency string `json:"default_input_currency"`
		Locale               string `json:"locale"`
	}

	var req UpdateSettingsRequest
	if err := h.ReadJSON(r, &req); err != nil {
		h.WriteJSON(w, http.StatusBadRequest, &Response{Status: "error", Error: "Invalid request"})
		return
	}

	user, err := h.userRepo.GetByID(ctx, userID)
	if err != nil {
		h.WriteJSON(w, http.StatusInternalServerError, &Response{Status: "error", Error: err.Error()})
		return
	}
	if user == nil {
		h.WriteJSON(w, http.StatusNotFound, &Response{Status: "error", Error: "User not found"})
		return
	}

	if req.HomeCurrency != "" {
		user.HomeCurrency = req.HomeCurrency
	}
	if req.DefaultInputCurrency != "" {
		user.DefaultInputCurrency = req.DefaultInputCurrency
	}
	if req.Locale != "" {
		user.Locale = req.Locale
	}

	if err := h.userRepo.Update(ctx, user); err != nil {
		h.WriteJSON(w, http.StatusInternalServerError, &Response{Status: "error", Error: err.Error()})
		return
	}

	h.WriteJSON(w, http.StatusOK, &Response{Status: "success", Message: "Settings updated"})
}

// GetCurrencies godoc
func (h *Handler) GetCurrencies(w http.ResponseWriter, r *http.Request) {
	// Assuming we have a standard list of currencies.
	// We can add a specialized CurrencyRepository later if needed for dynamic lists.
	// For now, let's return a hardcoded list or fetch from DB if available.
	// Checking the Handler struct, we have exchangeRateSvc but no specific CurrencyRepo exposed (Wait, NewHandler had categoryRepo, etc).
	// Checking line 38, NewHandler arguments includes generic repos.
	// Wait, line 59 has expenseRepo.
	// Let's check NewHandler definition again. Does it have CurrencyRepo?
	// No, it doesn't seem to have CurrencyRepo in the struct (lines 13-36).
	// But `domain/repositories.go` defined `CurrencyRepository`.
	// I should probably add CurrencyRepository to Handler struct if I want to use it.
	// Or I can just return the hardcoded list supported by Frankfurter for now.

	// Quick fix: Return a static list of common currencies.
	currencies := []domain.Currency{
		{Code: "USD", Name: "United States Dollar", Symbol: "$", IsActive: true},
		{Code: "EUR", Name: "Euro", Symbol: "€", IsActive: true},
		{Code: "GBP", Name: "British Pound", Symbol: "£", IsActive: true},
		{Code: "JPY", Name: "Japanese Yen", Symbol: "¥", IsActive: true},
		{Code: "TWD", Name: "New Taiwan Dollar", Symbol: "NT$", IsActive: true},
		{Code: "CNY", Name: "Chinese Yuan", Symbol: "¥", IsActive: true},
		{Code: "KRW", Name: "South Korean Won", Symbol: "₩", IsActive: true},
		{Code: "AUD", Name: "Australian Dollar", Symbol: "$", IsActive: true},
		{Code: "CAD", Name: "Canadian Dollar", Symbol: "$", IsActive: true},
	}

	h.WriteJSON(w, http.StatusOK, &Response{Status: "success", Data: currencies})
}

// Health check
func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	h.WriteJSON(w, http.StatusOK, &Response{Status: "ok"})
}

// HandleGetUserAggregate returns all user settings in one response
func (h *Handler) HandleGetUserAggregate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		h.WriteJSON(w, http.StatusUnauthorized, &Response{Status: "error", Error: "Unauthorized"})
		return
	}

	settings, err := h.getUserAggregateUC.Execute(ctx, userID)
	if err != nil {
		h.WriteJSON(w, http.StatusInternalServerError, &Response{Status: "error", Error: err.Error()})
		return
	}

	h.WriteJSON(w, http.StatusOK, &Response{Status: "success", Data: settings})
}

// HandleUpdateUserAggregate updates all user settings
func (h *Handler) HandleUpdateUserAggregate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		h.WriteJSON(w, http.StatusUnauthorized, &Response{Status: "error", Error: "Unauthorized"})
		return
	}

	var settings domain.AggregateSettings
	if err := h.ReadJSON(r, &settings); err != nil {
		h.WriteJSON(w, http.StatusBadRequest, &Response{Status: "error", Error: "invalid request"})
		return
	}

	if err := h.updateUserAggregateUC.Execute(userID, &settings); err != nil {
		h.WriteJSON(w, http.StatusInternalServerError, &Response{Status: "error", Error: err.Error()})
		return
	}

	h.WriteJSON(w, http.StatusOK, &Response{Status: "success", Message: "Settings updated"})
}

// RegisterRoutes registers all HTTP routes
func RegisterRoutes(
	mux *http.ServeMux,
	handler *Handler,
	aiCostHandler *AICostHandler,
	pricingHandler *PricingHandler,
	reportHandler *ReportHandler,
	shortLinkHandler *ShortLinkHandler,
	adminAuthHandler *AdminAuthHandler,
	adminAnalyticsHandler *AdminAnalyticsHandler,
) {
	// --- Public Endpoints ---
	mux.HandleFunc("GET /health", handler.Health)
	mux.HandleFunc("POST /api/users/auto-signup", handler.AutoSignup)
	mux.HandleFunc("GET /api/currencies", handler.GetCurrencies)

	if shortLinkHandler != nil {
		mux.HandleFunc("GET /r/{id}", shortLinkHandler.HandleRedirect)
	}

	// --- Middleware Helpers ---
	withAuth := func(h http.HandlerFunc) http.Handler {
		return handler.AuthMiddleware(h)
	}
	withAdmin := func(h http.HandlerFunc) http.Handler {
		return handler.AdminMiddleware(h)
	}

	withAdminSession := func(h http.HandlerFunc) http.Handler {
		return handler.AdminSessionMiddleware(h)
	}

	// --- Admin Auth Endpoints (Bearer Token for UI) ---
	if adminAuthHandler != nil {
		mux.HandleFunc("POST /api/admin/auth/login", adminAuthHandler.Login)
		mux.HandleFunc("GET /api/admin/auth/verify", adminAuthHandler.Verify)
		mux.HandleFunc("POST /api/admin/auth/logout", adminAuthHandler.Logout)

		if adminAnalyticsHandler != nil {
			mux.Handle("GET /api/admin/analytics/overview", withAdminSession(adminAnalyticsHandler.Overview))
		}
	}

	// --- User Endpoints (Authenticated) ---
	// User
	mux.Handle("GET /api/user", withAuth(handler.GetUser))
	mux.Handle("GET /api/user/aggregate", withAuth(handler.HandleGetUserAggregate))
	mux.Handle("PUT /api/user/aggregate", withAuth(handler.HandleUpdateUserAggregate))
	mux.Handle("PUT /api/user/settings", withAuth(handler.UpdateUserSettings))

	// Categories
	mux.Handle("GET /api/categories", withAuth(handler.GetCategories))
	mux.Handle("POST /api/categories", withAuth(handler.CreateCategory))
	mux.Handle("PUT /api/categories", withAuth(handler.UpdateCategory))
	mux.Handle("DELETE /api/categories", withAuth(handler.DeleteCategory))
	mux.Handle("GET /api/categories/list", withAuth(handler.ListCategories))

	// User Categories (Legacy/Duplicate paths?)
	mux.Handle("GET /api/user/categories", withAuth(handler.ListCategories))
	mux.Handle("POST /api/user/categories", withAuth(handler.CreateCategory))
	mux.Handle("PUT /api/user/categories", withAuth(handler.UpdateCategory))
	mux.Handle("DELETE /api/user/categories", withAuth(handler.DeleteCategory))
	mux.Handle("POST /api/user/categories/merge", withAuth(handler.MergeCategories))

	// Expenses
	mux.Handle("GET /api/expenses", withAuth(handler.GetExpenses))
	mux.Handle("POST /api/expenses", withAuth(handler.CreateExpense))
	mux.Handle("PUT /api/expenses", withAuth(handler.UpdateExpense))
	mux.Handle("DELETE /api/expenses", withAuth(handler.DeleteExpense))
	mux.Handle("POST /api/expenses/parse", withAuth(handler.ParseExpenses))
	mux.Handle("GET /api/expenses/search", withAuth(handler.SearchExpenses))
	mux.Handle("GET /api/expenses/filter", withAuth(handler.FilterExpenses))

	// Recurring
	mux.Handle("GET /api/recurring", withAuth(handler.ListRecurring))
	mux.Handle("POST /api/recurring", withAuth(handler.CreateRecurring))
	mux.Handle("PUT /api/recurring", withAuth(handler.UpdateRecurring))
	mux.Handle("DELETE /api/recurring", withAuth(handler.DeleteRecurring))
	mux.Handle("GET /api/recurring/upcoming", withAuth(handler.GetUpcomingRecurring))
	mux.Handle("POST /api/recurring/process", withAuth(handler.ProcessRecurring))

	// Notifications
	mux.Handle("GET /api/notifications", withAuth(handler.ListNotifications))
	mux.Handle("POST /api/notifications", withAuth(handler.CreateNotification))
	mux.Handle("PUT /api/notifications", withAuth(handler.MarkNotificationAsRead))
	mux.Handle("PUT /api/notifications/mark-all", withAuth(handler.MarkAllNotificationsAsRead))
	mux.Handle("DELETE /api/notifications", withAuth(handler.DeleteNotification))
	mux.Handle("GET /api/notifications/preferences", withAuth(handler.GetNotificationPreferences))
	mux.Handle("PUT /api/notifications/preferences", withAuth(handler.UpdateNotificationPreferences))

	// Archives
	mux.Handle("POST /api/archives", withAuth(handler.CreateArchive))
	mux.Handle("GET /api/archives", withAuth(handler.ListArchives))
	mux.Handle("GET /api/archives/stats", withAuth(handler.GetArchiveStats))
	mux.Handle("GET /api/archives/details", withAuth(handler.GetArchiveDetails))
	mux.Handle("POST /api/archives/restore", withAuth(handler.RestoreArchive))
	mux.Handle("POST /api/archives/purge", withAuth(handler.PurgeArchive))
	mux.Handle("POST /api/archives/export", withAuth(handler.ExportArchive))

	// Budgets
	mux.Handle("GET /api/budgets/status", withAuth(handler.GetBudgetStatus))
	mux.Handle("GET /api/budgets/compare", withAuth(handler.CompareToBudget))

	// Export
	mux.Handle("GET /api/export/expenses", withAuth(handler.ExportExpenses))
	mux.Handle("GET /api/export/summary", withAuth(handler.ExportSummary))

	// Reports
	mux.Handle("POST /api/reports/generate", withAuth(handler.GenerateReport))

	// Currencies (Public)
	// mux.HandleFunc("GET /api/currencies", handler.GetCurrencies) - moved to Public group above


	// Legal/Polices - usually public
	mux.HandleFunc("GET /api/policies/{key}", handler.GetPolicy)

	// --- Admin Endpoints (Protected by API Key) ---
	mux.Handle("GET /api/metrics/dau", withAdmin(handler.GetMetricsDAU))
	mux.Handle("GET /api/metrics/expenses-summary", withAdmin(handler.GetMetricsExpenses))
	mux.Handle("GET /api/metrics/growth", withAdmin(handler.GetMetricsGrowth))
	mux.Handle("POST /api/exchange-rates/refresh", withAdmin(handler.RefreshExchangeRates))

	// AI Cost
	if aiCostHandler != nil {
		mux.Handle("GET /api/metrics/ai-costs", withAdmin(aiCostHandler.GetAICostMetrics))
		mux.Handle("GET /api/metrics/ai-costs/summary", withAdmin(aiCostHandler.GetAICostSummary))
		mux.Handle("GET /api/metrics/ai-costs/daily", withAdmin(aiCostHandler.GetAICostDaily))
		mux.Handle("GET /api/metrics/ai-costs/by-operation", withAdmin(aiCostHandler.GetAICostByOperation))
		mux.Handle("GET /api/metrics/ai-costs/top-users", withAdmin(aiCostHandler.GetAICostTopUsers))
	}

	// Pricing
	if pricingHandler != nil {
		mux.Handle("POST /api/pricing/sync", withAdmin(pricingHandler.SyncPricing))
		mux.Handle("GET /api/pricing", withAdmin(pricingHandler.ListPricing))
		mux.Handle("POST /api/pricing", withAdmin(pricingHandler.CreatePricing))
		mux.Handle("PUT /api/pricing/{id}", withAdmin(pricingHandler.UpdatePricing))
		mux.Handle("DELETE /api/pricing/{id}", withAdmin(pricingHandler.DeletePricing))
	}

	// Reports (Secure Link)
	if reportHandler != nil {
		// Used for viewing the report via shared link?
		// "GET /api/reports/summary": reportHandler.GetReportSummary
		// This likely corresponds to the short link redirect target?
		// If it's for the public view of the report, it shouldn't be auth-protected by user token.
		// It might be protected by a signed URL token (handled inside).
		// So I'll leave it Public/Ungrouped (no middleware).
		mux.HandleFunc("GET /api/reports/summary", reportHandler.GetReportSummary)
	}
}
