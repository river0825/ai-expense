package http

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/riverlin/aiexpense/internal/domain"
	"github.com/riverlin/aiexpense/internal/usecase"
)

// Handler holds all HTTP request handlers
type Handler struct {
	autoSignupUC        *usecase.AutoSignupUseCase
	parseConversationUC *usecase.ParseConversationUseCase
	createExpenseUC     *usecase.CreateExpenseUseCase
	getExpensesUC       *usecase.GetExpensesUseCase
	updateExpenseUC     *usecase.UpdateExpenseUseCase
	deleteExpenseUC     *usecase.DeleteExpenseUseCase
	manageCategoryUC    *usecase.ManageCategoryUseCase
	generateReportUC    *usecase.GenerateReportUseCase
	budgetManagementUC  *usecase.BudgetManagementUseCase
	dataExportUC        *usecase.DataExportUseCase
	recurringExpenseUC  *usecase.RecurringExpenseUseCase
	notificationUC      *usecase.NotificationUseCase
	searchExpenseUC     *usecase.SearchExpenseUseCase
	archiveUC           *usecase.ArchiveUseCase
	metricsUC           *usecase.MetricsUseCase
	getPolicyUC         *usecase.GetPolicyUseCase
	userRepo            domain.UserRepository
	categoryRepo        domain.CategoryRepository
	expenseRepo         domain.ExpenseRepository
	metricsRepo         domain.MetricsRepository
	adminAPIKey         string
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
		updateExpenseUC:     updateExpenseUC,
		deleteExpenseUC:     deleteExpenseUC,
		manageCategoryUC:    manageCategoryUC,
		generateReportUC:    generateReportUC,
		budgetManagementUC:  budgetManagementUC,
		dataExportUC:        dataExportUC,
		recurringExpenseUC:  recurringExpenseUC,
		notificationUC:      notificationUC,
		searchExpenseUC:     searchExpenseUC,
		archiveUC:           archiveUC,
		metricsUC:           metricsUC,
		getPolicyUC:         getPolicyUC,
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

	resp, err := h.metricsUC.GetDailyActiveUsers(ctx, &usecase.DailyActiveUsersRequest{Days: 30})
	if err != nil {
		h.WriteJSON(w, http.StatusInternalServerError, &Response{Status: "error", Error: err.Error()})
		return
	}

	h.WriteJSON(w, http.StatusOK, &Response{Status: "success", Data: resp})
}

// GetMetricsExpenses retrieves expense summary
func (h *Handler) GetMetricsExpenses(w http.ResponseWriter, r *http.Request) {
	if !h.authenticateAdmin(r) {
		h.WriteJSON(w, http.StatusUnauthorized, &Response{Status: "error", Error: "Unauthorized"})
		return
	}

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
	if !h.authenticateAdmin(r) {
		h.WriteJSON(w, http.StatusUnauthorized, &Response{Status: "error", Error: "Unauthorized"})
		return
	}

	ctx := r.Context()

	resp, err := h.metricsUC.GetGrowthMetrics(ctx, &usecase.GrowthMetricsRequest{Days: 30})
	if err != nil {
		h.WriteJSON(w, http.StatusInternalServerError, &Response{Status: "error", Error: err.Error()})
		return
	}

	h.WriteJSON(w, http.StatusOK, &Response{Status: "success", Data: resp})
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

	type UpdateExpenseRequest struct {
		ID          string     `json:"id"`
		UserID      string     `json:"user_id"`
		Description *string    `json:"description,omitempty"`
		Amount      *float64   `json:"amount,omitempty"`
		CategoryID  *string    `json:"category_id,omitempty"`
		ExpenseDate *time.Time `json:"expense_date,omitempty"`
	}

	var req UpdateExpenseRequest
	if err := h.ReadJSON(r, &req); err != nil {
		h.WriteJSON(w, http.StatusBadRequest, &Response{Status: "error", Error: "Invalid request"})
		return
	}

	if req.ID == "" || req.UserID == "" {
		h.WriteJSON(w, http.StatusBadRequest, &Response{Status: "error", Error: "id and user_id are required"})
		return
	}

	resp, err := h.updateExpenseUC.Execute(ctx, &usecase.UpdateRequest{
		ID:          req.ID,
		UserID:      req.UserID,
		Description: req.Description,
		Amount:      req.Amount,
		CategoryID:  req.CategoryID,
		ExpenseDate: req.ExpenseDate,
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

	type DeleteExpenseRequest struct {
		ID     string `json:"id"`
		UserID string `json:"user_id"`
	}

	var req DeleteExpenseRequest
	if err := h.ReadJSON(r, &req); err != nil {
		h.WriteJSON(w, http.StatusBadRequest, &Response{Status: "error", Error: "Invalid request"})
		return
	}

	if req.ID == "" || req.UserID == "" {
		h.WriteJSON(w, http.StatusBadRequest, &Response{Status: "error", Error: "id and user_id are required"})
		return
	}

	resp, err := h.deleteExpenseUC.Execute(ctx, &usecase.DeleteRequest{
		ID:     req.ID,
		UserID: req.UserID,
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

	type CreateCategoryRequest struct {
		UserID   string   `json:"user_id"`
		Name     string   `json:"name"`
		Keywords []string `json:"keywords,omitempty"`
	}

	var req CreateCategoryRequest
	if err := h.ReadJSON(r, &req); err != nil {
		h.WriteJSON(w, http.StatusBadRequest, &Response{Status: "error", Error: "Invalid request"})
		return
	}

	if req.UserID == "" || req.Name == "" {
		h.WriteJSON(w, http.StatusBadRequest, &Response{Status: "error", Error: "user_id and name are required"})
		return
	}

	resp, err := h.manageCategoryUC.CreateCategory(ctx, &usecase.CreateCategoryRequest{
		UserID:   req.UserID,
		Name:     req.Name,
		Keywords: req.Keywords,
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

	type UpdateCategoryRequest struct {
		ID       string   `json:"id"`
		UserID   string   `json:"user_id"`
		Name     *string  `json:"name,omitempty"`
		Keywords []string `json:"keywords,omitempty"`
	}

	var req UpdateCategoryRequest
	if err := h.ReadJSON(r, &req); err != nil {
		h.WriteJSON(w, http.StatusBadRequest, &Response{Status: "error", Error: "Invalid request"})
		return
	}

	if req.ID == "" || req.UserID == "" {
		h.WriteJSON(w, http.StatusBadRequest, &Response{Status: "error", Error: "id and user_id are required"})
		return
	}

	resp, err := h.manageCategoryUC.UpdateCategory(ctx, &usecase.UpdateCategoryRequest{
		UserID:   req.UserID,
		ID:       req.ID,
		Name:     req.Name,
		Keywords: req.Keywords,
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

	type DeleteCategoryRequest struct {
		ID     string `json:"id"`
		UserID string `json:"user_id"`
	}

	var req DeleteCategoryRequest
	if err := h.ReadJSON(r, &req); err != nil {
		h.WriteJSON(w, http.StatusBadRequest, &Response{Status: "error", Error: "Invalid request"})
		return
	}

	if req.ID == "" || req.UserID == "" {
		h.WriteJSON(w, http.StatusBadRequest, &Response{Status: "error", Error: "id and user_id are required"})
		return
	}

	resp, err := h.manageCategoryUC.DeleteCategory(ctx, &usecase.DeleteCategoryRequest{
		UserID: req.UserID,
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
	userID := r.URL.Query().Get("user_id")

	if userID == "" {
		h.WriteJSON(w, http.StatusBadRequest, &Response{Status: "error", Error: "user_id is required"})
		return
	}

	resp, err := h.manageCategoryUC.ListCategories(ctx, &usecase.ListCategoriesRequest{
		UserID: userID,
	})

	if err != nil {
		h.WriteJSON(w, http.StatusInternalServerError, &Response{Status: "error", Error: err.Error()})
		return
	}

	h.WriteJSON(w, http.StatusOK, &Response{Status: "success", Data: resp})
}

// GenerateReport godoc
func (h *Handler) GenerateReport(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	type GenerateReportRequest struct {
		UserID     string    `json:"user_id"`
		ReportType string    `json:"report_type"`
		StartDate  time.Time `json:"start_date"`
		EndDate    time.Time `json:"end_date"`
	}

	var req GenerateReportRequest
	if err := h.ReadJSON(r, &req); err != nil {
		h.WriteJSON(w, http.StatusBadRequest, &Response{Status: "error", Error: "Invalid request"})
		return
	}

	if req.UserID == "" {
		h.WriteJSON(w, http.StatusBadRequest, &Response{Status: "error", Error: "user_id is required"})
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
		UserID:     req.UserID,
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
func (h *Handler) GetBudgetStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := r.URL.Query().Get("user_id")

	if userID == "" {
		h.WriteJSON(w, http.StatusBadRequest, &Response{Status: "error", Error: "user_id is required"})
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
	userID := r.URL.Query().Get("user_id")
	categoryID := r.URL.Query().Get("category_id")
	period := r.URL.Query().Get("period")

	if userID == "" {
		h.WriteJSON(w, http.StatusBadRequest, &Response{Status: "error", Error: "user_id is required"})
		return
	}

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
	userID := r.URL.Query().Get("user_id")
	format := r.URL.Query().Get("format")
	startDate := r.URL.Query().Get("start_date")
	endDate := r.URL.Query().Get("end_date")

	if userID == "" {
		h.WriteJSON(w, http.StatusBadRequest, &Response{Status: "error", Error: "user_id is required"})
		return
	}

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
	userID := r.URL.Query().Get("user_id")
	startDate := r.URL.Query().Get("start_date")
	endDate := r.URL.Query().Get("end_date")

	if userID == "" {
		h.WriteJSON(w, http.StatusBadRequest, &Response{Status: "error", Error: "user_id is required"})
		return
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
	userID := r.URL.Query().Get("user_id")
	query := r.URL.Query().Get("q")
	categoryID := r.URL.Query().Get("category_id")
	sortBy := r.URL.Query().Get("sort_by")
	limit := 20

	if userID == "" {
		h.WriteJSON(w, http.StatusBadRequest, &Response{Status: "error", Error: "user_id is required"})
		return
	}

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
	userID := r.URL.Query().Get("user_id")
	period := r.URL.Query().Get("period")
	categoryID := r.URL.Query().Get("category_id")

	if userID == "" {
		h.WriteJSON(w, http.StatusBadRequest, &Response{Status: "error", Error: "user_id is required"})
		return
	}

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

	type CreateRecurringRequest struct {
		UserID      string    `json:"user_id"`
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
		UserID:      req.UserID,
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
	userID := r.URL.Query().Get("user_id")

	if userID == "" {
		h.WriteJSON(w, http.StatusBadRequest, &Response{Status: "error", Error: "user_id is required"})
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

	type UpdateRecurringRequest struct {
		UserID      string   `json:"user_id"`
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
		UserID:      req.UserID,
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
	userID := r.URL.Query().Get("user_id")
	id := r.URL.Query().Get("id")

	if userID == "" || id == "" {
		h.WriteJSON(w, http.StatusBadRequest, &Response{Status: "error", Error: "user_id and id are required"})
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
	userID := r.URL.Query().Get("user_id")

	if userID == "" {
		h.WriteJSON(w, http.StatusBadRequest, &Response{Status: "error", Error: "user_id is required"})
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

	type ProcessRecurringRequest struct {
		UserID string    `json:"user_id"`
		Date   time.Time `json:"date"`
	}

	var req ProcessRecurringRequest
	if err := h.ReadJSON(r, &req); err != nil {
		h.WriteJSON(w, http.StatusBadRequest, &Response{Status: "error", Error: "Invalid request"})
		return
	}

	resp, err := h.recurringExpenseUC.ProcessRecurring(ctx, &usecase.ProcessRecurringRequest{
		UserID: req.UserID,
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

	type CreateNotificationRequest struct {
		UserID  string                 `json:"user_id"`
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
		UserID:  req.UserID,
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
	userID := r.URL.Query().Get("user_id")

	if userID == "" {
		h.WriteJSON(w, http.StatusBadRequest, &Response{Status: "error", Error: "user_id is required"})
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

	type MarkAsReadRequest struct {
		UserID         string `json:"user_id"`
		NotificationID string `json:"notification_id"`
	}

	var req MarkAsReadRequest
	if err := h.ReadJSON(r, &req); err != nil {
		h.WriteJSON(w, http.StatusBadRequest, &Response{Status: "error", Error: "Invalid request"})
		return
	}

	resp, err := h.notificationUC.MarkAsRead(ctx, &usecase.MarkAsReadRequest{
		UserID:         req.UserID,
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

	type MarkAllAsReadRequest struct {
		UserID string `json:"user_id"`
	}

	var req MarkAllAsReadRequest
	if err := h.ReadJSON(r, &req); err != nil {
		h.WriteJSON(w, http.StatusBadRequest, &Response{Status: "error", Error: "Invalid request"})
		return
	}

	resp, err := h.notificationUC.MarkAllAsRead(ctx, &usecase.MarkAllAsReadRequest{
		UserID: req.UserID,
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
	userID := r.URL.Query().Get("user_id")
	notificationID := r.URL.Query().Get("id")

	if userID == "" || notificationID == "" {
		h.WriteJSON(w, http.StatusBadRequest, &Response{Status: "error", Error: "user_id and id are required"})
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
	userID := r.URL.Query().Get("user_id")

	if userID == "" {
		h.WriteJSON(w, http.StatusBadRequest, &Response{Status: "error", Error: "user_id is required"})
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

	type UpdatePreferencesRequest struct {
		UserID              string `json:"user_id"`
		BudgetAlerts        *bool  `json:"budget_alerts,omitempty"`
		RecurringReminders  *bool  `json:"recurring_reminders,omitempty"`
		ReportNotifications *bool  `json:"report_notifications,omitempty"`
		ExpenseReminders    *bool  `json:"expense_reminders,omitempty"`
		DailyDigest         *bool  `json:"daily_digest,omitempty"`
		WeeklyReport        *bool  `json:"weekly_report,omitempty"`
	}

	var req UpdatePreferencesRequest
	if err := h.ReadJSON(r, &req); err != nil {
		h.WriteJSON(w, http.StatusBadRequest, &Response{Status: "error", Error: "Invalid request"})
		return
	}

	resp, err := h.notificationUC.UpdatePreferences(ctx, &usecase.UpdatePreferencesRequest{
		UserID:              req.UserID,
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

	type CreateArchiveRequest struct {
		UserID        string    `json:"user_id"`
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
		UserID:        req.UserID,
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
	userID := r.URL.Query().Get("user_id")

	if userID == "" {
		h.WriteJSON(w, http.StatusBadRequest, &Response{Status: "error", Error: "user_id is required"})
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
	userID := r.URL.Query().Get("user_id")

	if userID == "" {
		h.WriteJSON(w, http.StatusBadRequest, &Response{Status: "error", Error: "user_id is required"})
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
	userID := r.URL.Query().Get("user_id")
	archiveID := r.URL.Query().Get("archive_id")

	if userID == "" || archiveID == "" {
		h.WriteJSON(w, http.StatusBadRequest, &Response{Status: "error", Error: "user_id and archive_id are required"})
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

	type RestoreArchiveRequest struct {
		UserID    string `json:"user_id"`
		ArchiveID string `json:"archive_id"`
		Strategy  string `json:"strategy,omitempty"`
	}

	var req RestoreArchiveRequest
	if err := h.ReadJSON(r, &req); err != nil {
		h.WriteJSON(w, http.StatusBadRequest, &Response{Status: "error", Error: "Invalid request"})
		return
	}

	resp, err := h.archiveUC.RestoreArchive(ctx, &usecase.RestoreArchiveRequest{
		UserID:    req.UserID,
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

// Health check
func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	h.WriteJSON(w, http.StatusOK, &Response{Status: "ok"})
}

// RegisterRoutes registers all HTTP routes
func RegisterRoutes(mux *http.ServeMux, handler *Handler) {
	// User endpoints
	mux.HandleFunc("POST /api/users/auto-signup", handler.AutoSignup)

	// Expense endpoints
	mux.HandleFunc("POST /api/expenses/parse", handler.ParseExpenses)
	mux.HandleFunc("POST /api/expenses", handler.CreateExpense)
	mux.HandleFunc("PUT /api/expenses", handler.UpdateExpense)
	mux.HandleFunc("DELETE /api/expenses", handler.DeleteExpense)
	mux.HandleFunc("GET /api/expenses", handler.GetExpenses)
	mux.HandleFunc("GET /api/expenses/search", handler.SearchExpenses)
	mux.HandleFunc("GET /api/expenses/filter", handler.FilterExpenses)

	// Category endpoints
	mux.HandleFunc("POST /api/categories", handler.CreateCategory)
	mux.HandleFunc("PUT /api/categories", handler.UpdateCategory)
	mux.HandleFunc("DELETE /api/categories", handler.DeleteCategory)
	mux.HandleFunc("GET /api/categories", handler.GetCategories)
	mux.HandleFunc("GET /api/categories/list", handler.ListCategories)

	// Recurring expense endpoints
	mux.HandleFunc("POST /api/recurring", handler.CreateRecurring)
	mux.HandleFunc("GET /api/recurring", handler.ListRecurring)
	mux.HandleFunc("PUT /api/recurring", handler.UpdateRecurring)
	mux.HandleFunc("DELETE /api/recurring", handler.DeleteRecurring)
	mux.HandleFunc("GET /api/recurring/upcoming", handler.GetUpcomingRecurring)
	mux.HandleFunc("POST /api/recurring/process", handler.ProcessRecurring)

	// Notification endpoints
	mux.HandleFunc("POST /api/notifications", handler.CreateNotification)
	mux.HandleFunc("GET /api/notifications", handler.ListNotifications)
	mux.HandleFunc("PUT /api/notifications", handler.MarkNotificationAsRead)
	mux.HandleFunc("PUT /api/notifications/mark-all", handler.MarkAllNotificationsAsRead)
	mux.HandleFunc("DELETE /api/notifications", handler.DeleteNotification)
	mux.HandleFunc("GET /api/notifications/preferences", handler.GetNotificationPreferences)
	mux.HandleFunc("PUT /api/notifications/preferences", handler.UpdateNotificationPreferences)

	// Archive endpoints
	mux.HandleFunc("POST /api/archives", handler.CreateArchive)
	mux.HandleFunc("GET /api/archives", handler.ListArchives)
	mux.HandleFunc("GET /api/archives/stats", handler.GetArchiveStats)
	mux.HandleFunc("GET /api/archives/details", handler.GetArchiveDetails)
	mux.HandleFunc("POST /api/archives/restore", handler.RestoreArchive)
	mux.HandleFunc("POST /api/archives/purge", handler.PurgeArchive)
	mux.HandleFunc("POST /api/archives/export", handler.ExportArchive)

	// Report endpoints
	mux.HandleFunc("POST /api/reports/generate", handler.GenerateReport)

	// Budget endpoints
	mux.HandleFunc("GET /api/budgets/status", handler.GetBudgetStatus)
	mux.HandleFunc("GET /api/budgets/compare", handler.CompareToBudget)

	// Export endpoints
	mux.HandleFunc("GET /api/export/expenses", handler.ExportExpenses)
	mux.HandleFunc("GET /api/export/summary", handler.ExportSummary)

	// Metrics endpoints
	mux.HandleFunc("GET /api/metrics/dau", handler.GetMetricsDAU)
	mux.HandleFunc("GET /api/metrics/expenses-summary", handler.GetMetricsExpenses)
	mux.HandleFunc("GET /api/metrics/growth", handler.GetMetricsGrowth)

	// Legal endpoints
	mux.HandleFunc("GET /api/policies/{key}", handler.GetPolicy)

	// Health endpoint
	mux.HandleFunc("GET /health", handler.Health)
}
