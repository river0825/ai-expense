package http

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/riverlin/aiexpense/internal/usecase"
)

type AICostHandler struct {
	aiCostUC    *usecase.AICostUseCase
	adminAPIKey string
}

func NewAICostHandler(aiCostUC *usecase.AICostUseCase, adminAPIKey string) *AICostHandler {
	return &AICostHandler{
		aiCostUC:    aiCostUC,
		adminAPIKey: adminAPIKey,
	}
}

func (h *AICostHandler) authenticateAdmin(r *http.Request) bool {
	if h.adminAPIKey == "" {
		return true
	}
	key := r.Header.Get("X-API-Key")
	return key == h.adminAPIKey
}

func (h *AICostHandler) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (h *AICostHandler) GetAICostMetrics(w http.ResponseWriter, r *http.Request) {
	if !h.authenticateAdmin(r) {
		h.writeJSON(w, http.StatusUnauthorized, map[string]string{"status": "error", "error": "Unauthorized"})
		return
	}

	ctx := r.Context()

	daysStr := r.URL.Query().Get("days")
	days := 30
	if daysStr != "" {
		if d, err := strconv.Atoi(daysStr); err == nil && d > 0 {
			days = d
		}
	}

	resp, err := h.aiCostUC.GetAICostMetrics(ctx, &usecase.AICostMetricsRequest{Days: days})
	if err != nil {
		h.writeJSON(w, http.StatusInternalServerError, map[string]string{"status": "error", "error": err.Error()})
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]interface{}{"status": "success", "data": resp})
}

func (h *AICostHandler) GetAICostSummary(w http.ResponseWriter, r *http.Request) {
	if !h.authenticateAdmin(r) {
		h.writeJSON(w, http.StatusUnauthorized, map[string]string{"status": "error", "error": "Unauthorized"})
		return
	}

	ctx := r.Context()

	daysStr := r.URL.Query().Get("days")
	days := 30
	if daysStr != "" {
		if d, err := strconv.Atoi(daysStr); err == nil && d > 0 {
			days = d
		}
	}

	resp, err := h.aiCostUC.GetSummary(ctx, &usecase.AICostSummaryRequest{Days: days})
	if err != nil {
		h.writeJSON(w, http.StatusInternalServerError, map[string]string{"status": "error", "error": err.Error()})
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]interface{}{"status": "success", "data": resp})
}

func (h *AICostHandler) GetAICostDaily(w http.ResponseWriter, r *http.Request) {
	if !h.authenticateAdmin(r) {
		h.writeJSON(w, http.StatusUnauthorized, map[string]string{"status": "error", "error": "Unauthorized"})
		return
	}

	ctx := r.Context()

	daysStr := r.URL.Query().Get("days")
	days := 30
	if daysStr != "" {
		if d, err := strconv.Atoi(daysStr); err == nil && d > 0 {
			days = d
		}
	}

	resp, err := h.aiCostUC.GetDailyStats(ctx, &usecase.AICostDailyRequest{Days: days})
	if err != nil {
		h.writeJSON(w, http.StatusInternalServerError, map[string]string{"status": "error", "error": err.Error()})
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]interface{}{"status": "success", "data": resp})
}

func (h *AICostHandler) GetAICostByOperation(w http.ResponseWriter, r *http.Request) {
	if !h.authenticateAdmin(r) {
		h.writeJSON(w, http.StatusUnauthorized, map[string]string{"status": "error", "error": "Unauthorized"})
		return
	}

	ctx := r.Context()

	daysStr := r.URL.Query().Get("days")
	days := 30
	if daysStr != "" {
		if d, err := strconv.Atoi(daysStr); err == nil && d > 0 {
			days = d
		}
	}

	resp, err := h.aiCostUC.GetByOperation(ctx, &usecase.AICostByOperationRequest{Days: days})
	if err != nil {
		h.writeJSON(w, http.StatusInternalServerError, map[string]string{"status": "error", "error": err.Error()})
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]interface{}{"status": "success", "data": resp})
}

func (h *AICostHandler) GetAICostTopUsers(w http.ResponseWriter, r *http.Request) {
	if !h.authenticateAdmin(r) {
		h.writeJSON(w, http.StatusUnauthorized, map[string]string{"status": "error", "error": "Unauthorized"})
		return
	}

	ctx := r.Context()

	daysStr := r.URL.Query().Get("days")
	days := 30
	if daysStr != "" {
		if d, err := strconv.Atoi(daysStr); err == nil && d > 0 {
			days = d
		}
	}

	limitStr := r.URL.Query().Get("limit")
	limit := 10
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	resp, err := h.aiCostUC.GetTopUsers(ctx, &usecase.AICostByUserRequest{Days: days, Limit: limit})
	if err != nil {
		h.writeJSON(w, http.StatusInternalServerError, map[string]string{"status": "error", "error": err.Error()})
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]interface{}{"status": "success", "data": resp})
}

func RegisterAICostRoutes(mux *http.ServeMux, handler *AICostHandler) {
	mux.HandleFunc("GET /api/metrics/ai-costs", handler.GetAICostMetrics)
	mux.HandleFunc("GET /api/metrics/ai-costs/summary", handler.GetAICostSummary)
	mux.HandleFunc("GET /api/metrics/ai-costs/daily", handler.GetAICostDaily)
	mux.HandleFunc("GET /api/metrics/ai-costs/by-operation", handler.GetAICostByOperation)
	mux.HandleFunc("GET /api/metrics/ai-costs/top-users", handler.GetAICostTopUsers)
}
