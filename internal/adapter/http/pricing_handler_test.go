package http

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

// Test mock provider for handler testing
type TestPricingProvider struct {
	configs []*domain.PricingConfig
}

func (tp *TestPricingProvider) Fetch(ctx context.Context) ([]*domain.PricingConfig, error) {
	return tp.configs, nil
}

func (tp *TestPricingProvider) Provider() string {
	return "gemini"
}

// Test mock repository for handler testing
type TestPricingRepositoryHandler struct {
	data []*domain.PricingConfig
}

func (tr *TestPricingRepositoryHandler) GetByProviderAndModel(ctx context.Context, provider, model string) (*domain.PricingConfig, error) {
	for _, c := range tr.data {
		if c.Provider == provider && c.Model == model && c.IsActive {
			return c, nil
		}
	}
	return nil, nil
}

func (tr *TestPricingRepositoryHandler) GetAll(ctx context.Context) ([]*domain.PricingConfig, error) {
	return tr.data, nil
}

func (tr *TestPricingRepositoryHandler) Create(ctx context.Context, config *domain.PricingConfig) error {
	tr.data = append(tr.data, config)
	return nil
}

func (tr *TestPricingRepositoryHandler) Update(ctx context.Context, config *domain.PricingConfig) error {
	for i, c := range tr.data {
		if c.ID == config.ID {
			tr.data[i] = config
			return nil
		}
	}
	return nil
}

func (tr *TestPricingRepositoryHandler) Deactivate(ctx context.Context, provider, model string) error {
	for i, c := range tr.data {
		if c.Provider == provider && c.Model == model && c.IsActive {
			c.IsActive = false
			tr.data[i] = c
		}
	}
	return nil
}

// TestSyncEndpoint_Success tests successful sync
func TestSyncEndpoint_Success(t *testing.T) {
	now := time.Now()
	fetched := []*domain.PricingConfig{
		{
			ID:               "test1",
			Provider:         "gemini",
			Model:            "gemini-2.5-lite",
			InputTokenPrice:  0.000000075,
			OutputTokenPrice: 0.0000003,
			EffectiveDate:    now,
			IsActive:         true,
			CreatedAt:        now,
			UpdatedAt:        now,
		},
	}

	repo := &TestPricingRepositoryHandler{data: []*domain.PricingConfig{}}
	provider := &TestPricingProvider{configs: fetched}
	handler := NewPricingHandler(repo, "test-key", map[string]domain.PricingProvider{"gemini": provider})

	mux := http.NewServeMux()
	RegisterPricingRoutes(mux, handler)

	req := httptest.NewRequest("POST", "/api/pricing/sync?provider=gemini", nil)
	req.Header.Set("X-API-Key", "test-key")
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", w.Code)
	}

	var result *usecase.SyncResult
	json.NewDecoder(w.Body).Decode(&result)

	if !result.Success {
		t.Error("Expected success=true")
	}

	if result.ModelsUpdated != 1 {
		t.Errorf("Expected 1 updated, got %d", result.ModelsUpdated)
	}
}

// TestSyncEndpoint_Unauthorized tests missing API key
func TestSyncEndpoint_Unauthorized(t *testing.T) {
	repo := &TestPricingRepositoryHandler{data: []*domain.PricingConfig{}}
	handler := NewPricingHandler(repo, "test-key", map[string]domain.PricingProvider{})

	mux := http.NewServeMux()
	RegisterPricingRoutes(mux, handler)

	req := httptest.NewRequest("POST", "/api/pricing/sync?provider=gemini", nil)
	// No X-API-Key header
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected 401, got %d", w.Code)
	}
}

// TestSyncEndpoint_InvalidProvider tests invalid provider
func TestSyncEndpoint_InvalidProvider(t *testing.T) {
	repo := &TestPricingRepositoryHandler{data: []*domain.PricingConfig{}}
	handler := NewPricingHandler(repo, "test-key", map[string]domain.PricingProvider{})

	mux := http.NewServeMux()
	RegisterPricingRoutes(mux, handler)

	req := httptest.NewRequest("POST", "/api/pricing/sync?provider=unknown", nil)
	req.Header.Set("X-API-Key", "test-key")
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected 400, got %d", w.Code)
	}
}

// TestListEndpoint tests GET /api/pricing
func TestListEndpoint(t *testing.T) {
	now := time.Now()
	repo := &TestPricingRepositoryHandler{
		data: []*domain.PricingConfig{
			{
				ID:               "test1",
				Provider:         "gemini",
				Model:            "gemini-2.5-lite",
				InputTokenPrice:  0.000000075,
				OutputTokenPrice: 0.0000003,
				IsActive:         true,
				CreatedAt:        now,
				UpdatedAt:        now,
			},
		},
	}
	handler := NewPricingHandler(repo, "test-key", map[string]domain.PricingProvider{})

	mux := http.NewServeMux()
	RegisterPricingRoutes(mux, handler)

	req := httptest.NewRequest("GET", "/api/pricing", nil)
	req.Header.Set("X-API-Key", "test-key")
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", w.Code)
	}

	var result map[string]interface{}
	json.NewDecoder(w.Body).Decode(&result)

	if result["status"] != "success" {
		t.Error("Expected success status")
	}
}

// TestCreateEndpoint tests POST /api/pricing
func TestCreateEndpoint(t *testing.T) {
	repo := &TestPricingRepositoryHandler{data: []*domain.PricingConfig{}}
	handler := NewPricingHandler(repo, "test-key", map[string]domain.PricingProvider{})

	mux := http.NewServeMux()
	RegisterPricingRoutes(mux, handler)

	body := []byte(`{
		"provider": "gemini",
		"model": "gemini-2.5-lite",
		"input_token_price": 0.000000075,
		"output_token_price": 0.0000003
	}`)

	req := httptest.NewRequest("POST", "/api/pricing", bytes.NewReader(body))
	req.Header.Set("X-API-Key", "test-key")
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected 201, got %d", w.Code)
	}

	if len(repo.data) != 1 {
		t.Errorf("Expected 1 config created, got %d", len(repo.data))
	}
}
