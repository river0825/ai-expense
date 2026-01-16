package http

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/riverlin/aiexpense/internal/usecase"
)

// MetricsHandler handles metrics endpoint requests
type MetricsHandler struct {
	metricsUC *usecase.MetricsUseCase
}

// NewMetricsHandler creates a new metrics handler
func NewMetricsHandler(metricsUC *usecase.MetricsUseCase) *MetricsHandler {
	return &MetricsHandler{
		metricsUC: metricsUC,
	}
}

// GetDAU retrieves daily active users metrics
func (h *MetricsHandler) GetDAU(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get days parameter from query string
	daysStr := r.URL.Query().Get("days")
	days := 30
	if daysStr != "" {
		if d, err := strconv.Atoi(daysStr); err == nil && d > 0 {
			days = d
		}
	}

	resp, err := h.metricsUC.GetDailyActiveUsers(ctx, &usecase.DailyActiveUsersRequest{
		Days: days,
	})
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, &Response{Status: "error", Error: err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, &Response{Status: "success", Data: resp})
}

// GetExpensesSummary retrieves expense summary metrics
func (h *MetricsHandler) GetExpensesSummary(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get days parameter
	daysStr := r.URL.Query().Get("days")
	days := 30
	if daysStr != "" {
		if d, err := strconv.Atoi(daysStr); err == nil && d > 0 {
			days = d
		}
	}

	resp, err := h.metricsUC.GetExpensesSummary(ctx, &usecase.ExpensesSummaryRequest{
		Days: days,
	})
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, &Response{Status: "error", Error: err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, &Response{Status: "success", Data: resp})
}

// GetCategoryTrends retrieves category trend metrics
func (h *MetricsHandler) GetCategoryTrends(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		writeJSON(w, http.StatusBadRequest, &Response{Status: "error", Error: "user_id is required"})
		return
	}

	// Get days parameter
	daysStr := r.URL.Query().Get("days")
	days := 30
	if daysStr != "" {
		if d, err := strconv.Atoi(daysStr); err == nil && d > 0 {
			days = d
		}
	}

	resp, err := h.metricsUC.GetCategoryTrends(ctx, &usecase.CategoryTrendsRequest{
		UserID: userID,
		Days:   days,
	})
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, &Response{Status: "error", Error: err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, &Response{Status: "success", Data: resp})
}

// GetGrowth retrieves growth metrics
func (h *MetricsHandler) GetGrowth(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get days parameter
	daysStr := r.URL.Query().Get("days")
	days := 30
	if daysStr != "" {
		if d, err := strconv.Atoi(daysStr); err == nil && d > 0 {
			days = d
		}
	}

	resp, err := h.metricsUC.GetGrowthMetrics(ctx, &usecase.GrowthMetricsRequest{
		Days: days,
	})
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, &Response{Status: "error", Error: err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, &Response{Status: "success", Data: resp})
}

// Helper function for JSON writing
func writeJSON(w http.ResponseWriter, status int, resp *Response) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(resp)
}
