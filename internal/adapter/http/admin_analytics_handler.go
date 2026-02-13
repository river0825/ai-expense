package http

import (
	"net/http"

	"github.com/riverlin/aiexpense/internal/usecase"
)

type AdminAnalyticsHandler struct {
	metricsUC *usecase.MetricsUseCase
}

func NewAdminAnalyticsHandler(metricsUC *usecase.MetricsUseCase) *AdminAnalyticsHandler {
	return &AdminAnalyticsHandler{metricsUC: metricsUC}
}

func (h *AdminAnalyticsHandler) Overview(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	dau, err := h.metricsUC.GetDailyActiveUsers(ctx, &usecase.DailyActiveUsersRequest{Days: 30})
	if err != nil {
		writeAdminJSON(w, http.StatusInternalServerError, &Response{Status: "error", Error: err.Error()})
		return
	}

	expenses, err := h.metricsUC.GetExpensesSummary(ctx, &usecase.ExpensesSummaryRequest{Days: 30})
	if err != nil {
		writeAdminJSON(w, http.StatusInternalServerError, &Response{Status: "error", Error: err.Error()})
		return
	}

	growth, err := h.metricsUC.GetGrowthMetrics(ctx, &usecase.GrowthMetricsRequest{Days: 30})
	if err != nil {
		writeAdminJSON(w, http.StatusInternalServerError, &Response{Status: "error", Error: err.Error()})
		return
	}

	writeAdminJSON(w, http.StatusOK, &Response{Status: "success", Data: map[string]interface{}{
		"dau":      dau,
		"expenses": expenses,
		"growth":   growth,
	}})
}
