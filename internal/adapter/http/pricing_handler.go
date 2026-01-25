package http

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/riverlin/aiexpense/internal/domain"
	"github.com/riverlin/aiexpense/internal/usecase"
)

type PricingHandler struct {
	syncUC      *usecase.PricingSyncUseCase
	pricingRepo domain.PricingRepository
	adminAPIKey string
	providers   map[string]domain.PricingProvider
}

func NewPricingHandler(
	pricingRepo domain.PricingRepository,
	adminAPIKey string,
	providers map[string]domain.PricingProvider,
) *PricingHandler {
	return &PricingHandler{
		pricingRepo: pricingRepo,
		adminAPIKey: adminAPIKey,
		providers:   providers,
	}
}

func (h *PricingHandler) authenticateAdmin(r *http.Request) bool {
	if h.adminAPIKey == "" {
		return true
	}
	key := r.Header.Get("X-API-Key")
	return key == h.adminAPIKey
}

func (h *PricingHandler) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// SyncPricing handles POST /api/pricing/sync?provider=gemini
func (h *PricingHandler) SyncPricing(w http.ResponseWriter, r *http.Request) {
	if !h.authenticateAdmin(r) {
		h.writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
		return
	}

	ctx := r.Context()
	provider := r.URL.Query().Get("provider")

	if provider == "" {
		h.writeJSON(w, http.StatusBadRequest, map[string]string{"error": "provider parameter required"})
		return
	}

	prov, exists := h.providers[provider]
	if !exists {
		h.writeJSON(w, http.StatusBadRequest, map[string]string{"error": "provider '" + provider + "' not supported"})
		return
	}

	syncUC := usecase.NewPricingSyncUseCase(h.pricingRepo, prov)
	result, err := syncUC.Sync(ctx)

	if err != nil && !result.Success {
		h.writeJSON(w, http.StatusInternalServerError, result)
		return
	}

	h.writeJSON(w, http.StatusOK, result)
}

// ListPricing handles GET /api/pricing
func (h *PricingHandler) ListPricing(w http.ResponseWriter, r *http.Request) {
	if !h.authenticateAdmin(r) {
		h.writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
		return
	}

	ctx := r.Context()
	configs, err := h.pricingRepo.GetAll(ctx)

	if err != nil {
		h.writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	activeOnly := r.URL.Query().Get("active") == "true"
	if activeOnly {
		active := []*domain.PricingConfig{}
		for _, c := range configs {
			if c.IsActive {
				active = append(active, c)
			}
		}
		configs = active
	}

	h.writeJSON(w, http.StatusOK, map[string]interface{}{"status": "success", "data": configs})
}

// CreatePricing handles POST /api/pricing
func (h *PricingHandler) CreatePricing(w http.ResponseWriter, r *http.Request) {
	if !h.authenticateAdmin(r) {
		h.writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
		return
	}

	ctx := r.Context()
	var req struct {
		Provider         string  `json:"provider"`
		Model            string  `json:"model"`
		InputTokenPrice  float64 `json:"input_token_price"`
		OutputTokenPrice float64 `json:"output_token_price"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}

	now := time.Now()
	config := &domain.PricingConfig{
		ID:               req.Provider + "_" + req.Model + "_" + now.Format("20060102150405"),
		Provider:         req.Provider,
		Model:            req.Model,
		InputTokenPrice:  req.InputTokenPrice,
		OutputTokenPrice: req.OutputTokenPrice,
		Currency:         "USD",
		EffectiveDate:    now,
		IsActive:         true,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	if err := h.pricingRepo.Create(ctx, config); err != nil {
		h.writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	h.writeJSON(w, http.StatusCreated, map[string]interface{}{"status": "success", "data": config})
}

// UpdatePricing handles PUT /api/pricing/{id}
func (h *PricingHandler) UpdatePricing(w http.ResponseWriter, r *http.Request) {
	if !h.authenticateAdmin(r) {
		h.writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
		return
	}

	ctx := r.Context()
	id := r.PathValue("id")

	var req struct {
		InputTokenPrice  float64 `json:"input_token_price"`
		OutputTokenPrice float64 `json:"output_token_price"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}

	config := &domain.PricingConfig{
		ID:               id,
		InputTokenPrice:  req.InputTokenPrice,
		OutputTokenPrice: req.OutputTokenPrice,
		UpdatedAt:        time.Now(),
	}

	if err := h.pricingRepo.Update(ctx, config); err != nil {
		h.writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]interface{}{"status": "success", "data": config})
}

// DeletePricing handles DELETE /api/pricing/{id}
func (h *PricingHandler) DeletePricing(w http.ResponseWriter, r *http.Request) {
	if !h.authenticateAdmin(r) {
		h.writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]interface{}{"status": "success", "message": "pricing deactivated (simulated)"})
}

// RegisterPricingRoutes registers all pricing routes
func RegisterPricingRoutes(mux *http.ServeMux, handler *PricingHandler) {
	mux.HandleFunc("POST /api/pricing/sync", handler.SyncPricing)
	mux.HandleFunc("GET /api/pricing", handler.ListPricing)
	mux.HandleFunc("POST /api/pricing", handler.CreatePricing)
	mux.HandleFunc("PUT /api/pricing/{id}", handler.UpdatePricing)
	mux.HandleFunc("DELETE /api/pricing/{id}", handler.DeletePricing)
}
