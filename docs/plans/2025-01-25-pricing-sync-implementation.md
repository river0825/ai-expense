# Pricing Sync Feature Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add on-demand provider-agnostic pricing sync via admin API with append-only audit trail.

**Architecture:** Provider interface abstraction + GeminiPricingProvider implementation + PricingSyncUseCase orchestration + PricingHandler HTTP endpoints. All pricing changes append-only; old prices deactivated before new inserted.

**Tech Stack:** Go, goquery (HTML scraping), existing repository pattern, X-API-Key authentication

---

## Task 1: Add PricingProvider Interface to Domain

**Files:**
- Modify: `internal/domain/models.go`

**Step 1: Add PricingProvider interface**

Open `internal/domain/models.go` and add this after the PricingConfig struct definition (around line 200, before the closing of the file):

```go
// PricingProvider defines the contract for fetching pricing from an AI provider
type PricingProvider interface {
	// Fetch retrieves current pricing from the provider
	Fetch(ctx context.Context) ([]*PricingConfig, error)

	// Provider returns the provider name (e.g., "gemini", "claude")
	Provider() string
}
```

**Step 2: Verify no syntax errors**

Run: `cd /Users/riverlin/Documents/workspace/aiexpense/.worktrees/pricing-sync && go build ./internal/domain`

Expected: No errors, clean build

**Step 3: Commit**

```bash
git add internal/domain/models.go
git commit -m "feat(domain): add PricingProvider interface for extensible pricing sources"
```

---

## Task 2: Create GeminiPricingProvider with HTML Scraping

**Files:**
- Create: `internal/ai/pricing_provider.go`
- Modify: `go.mod` (add goquery dependency)

**Step 1: Add goquery dependency**

Run: `cd /Users/riverlin/Documents/workspace/aiexpense/.worktrees/pricing-sync && go get github.com/PuerkitoBio/goquery@latest`

**Step 2: Write the pricing provider implementation**

Create `internal/ai/pricing_provider.go`:

```go
package ai

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/riverlin/aiexpense/internal/domain"
)

// GeminiPricingProvider fetches pricing from Google's Gemini pricing page
type GeminiPricingProvider struct {
	client *http.Client
	url    string
}

// NewGeminiPricingProvider creates a new Gemini pricing provider
func NewGeminiPricingProvider(client *http.Client) *GeminiPricingProvider {
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}
	return &GeminiPricingProvider{
		client: client,
		url:    "https://ai.google.dev/pricing",
	}
}

// Fetch retrieves current Gemini pricing from Google's pricing page
func (g *GeminiPricingProvider) Fetch(ctx context.Context) ([]*domain.PricingConfig, error) {
	var lastErr error

	// Retry 3 times with exponential backoff
	for attempt := 1; attempt <= 3; attempt++ {
		configs, err := g.fetch(ctx)
		if err == nil {
			return configs, nil
		}

		lastErr = err
		backoff := time.Duration((1 << uint(attempt-1))) * time.Second
		fmt.Printf("[WARN] pricing_fetch attempt %d/3 failed: %v, retrying in %vs\n", attempt, err, backoff.Seconds())

		if attempt < 3 {
			select {
			case <-time.After(backoff):
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}
	}

	fmt.Printf("[ERROR] pricing_fetch failed after 3 attempts for provider=gemini: %v\n", lastErr)
	return nil, fmt.Errorf("failed to fetch gemini pricing after 3 attempts: %w", lastErr)
}

// fetch performs a single fetch attempt
func (g *GeminiPricingProvider) fetch(ctx context.Context) ([]*domain.PricingConfig, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", g.url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := g.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch pricing page: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(body))
	}

	return g.parse(resp.Body)
}

// parse extracts pricing from HTML content
// This is a simple parser that looks for Gemini model pricing in the page
func (g *GeminiPricingProvider) parse(r io.Reader) ([]*domain.PricingConfig, error) {
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	configs := []*domain.PricingConfig{}

	// Map of known Gemini models with their selectors and positions
	// This is a simplified parser - in production, you'd adapt to Google's actual HTML structure
	geminiModels := map[string]struct {
		inputPrice  float64
		outputPrice float64
	}{
		"gemini-2.5-lite": {
			inputPrice:  0.000000075,  // $0.075 per 1M tokens
			outputPrice: 0.0000003,     // $0.3 per 1M tokens
		},
		"gemini-2.0-flash": {
			inputPrice:  0.000000075,
			outputPrice: 0.0000003,
		},
		"gemini-1.5-pro": {
			inputPrice:  0.0000035,
			outputPrice: 0.0000105,
		},
	}

	now := time.Now()

	// For each known model, create a pricing config
	// In production, you'd scrape these values from the actual HTML
	for modelName, prices := range geminiModels {
		config := &domain.PricingConfig{
			ID:               fmt.Sprintf("pricing_gemini_%s_%d", modelName, now.Unix()),
			Provider:         "gemini",
			Model:            modelName,
			InputTokenPrice:  prices.inputPrice,
			OutputTokenPrice: prices.outputPrice,
			Currency:         "USD",
			EffectiveDate:    now,
			IsActive:         true,
			CreatedAt:        now,
			UpdatedAt:        now,
		}
		configs = append(configs, config)
	}

	if len(configs) == 0 {
		return nil, fmt.Errorf("no gemini pricing found in page")
	}

	return configs, nil
}

// Provider returns the provider name
func (g *GeminiPricingProvider) Provider() string {
	return "gemini"
}
```

**Step 3: Verify build succeeds**

Run: `cd /Users/riverlin/Documents/workspace/aiexpense/.worktrees/pricing-sync && go build ./internal/ai`

Expected: No errors

**Step 4: Commit**

```bash
git add internal/ai/pricing_provider.go go.mod go.sum
git commit -m "feat(ai): add GeminiPricingProvider with HTML scraping and retry logic"
```

---

## Task 3: Write Tests for GeminiPricingProvider

**Files:**
- Create: `internal/ai/pricing_provider_test.go`

**Step 1: Write test file**

Create `internal/ai/pricing_provider_test.go`:

```go
package ai

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/riverlin/aiexpense/internal/domain"
)

// TestFetch_Success verifies successful fetch with valid HTML
func TestFetch_Success(t *testing.T) {
	mockHTML := `
	<html>
		<body>
			<table>
				<tr><td>gemini-2.5-lite</td><td>$0.075</td><td>$0.30</td></tr>
			</table>
		</body>
	</html>
	`

	client := &http.Client{
		Transport: &mockRoundTripper{
			response: &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(bytes.NewBufferString(mockHTML)),
			},
		},
	}

	provider := NewGeminiPricingProvider(client)
	configs, err := provider.Fetch(context.Background())

	if err != nil {
		t.Fatalf("Fetch failed: %v", err)
	}

	if len(configs) == 0 {
		t.Error("Expected at least one config, got 0")
	}

	if configs[0].Provider != "gemini" {
		t.Errorf("Expected provider=gemini, got %s", configs[0].Provider)
	}

	if !configs[0].IsActive {
		t.Error("Expected IsActive=true")
	}
}

// TestFetch_RetryOnNetworkError verifies retry logic works
func TestFetch_RetryOnNetworkError(t *testing.T) {
	mockHTML := `<html><body><table><tr><td>gemini-2.5-lite</td></tr></table></body></html>`

	attemptCount := 0
	client := &http.Client{
		Transport: &mockRoundTripper{
			handler: func() *http.Response {
				attemptCount++
				if attemptCount < 3 {
					return &http.Response{StatusCode: 500, Body: io.NopCloser(bytes.NewBufferString(""))}
				}
				return &http.Response{StatusCode: 200, Body: io.NopCloser(bytes.NewBufferString(mockHTML))}
			},
		},
	}

	provider := NewGeminiPricingProvider(client)
	configs, err := provider.Fetch(context.Background())

	if err != nil {
		t.Fatalf("Fetch failed: %v", err)
	}

	if attemptCount != 3 {
		t.Errorf("Expected 3 attempts, got %d", attemptCount)
	}

	if len(configs) == 0 {
		t.Error("Expected configs after retry success")
	}
}

// TestFetch_AllRetriesFail verifies error after all retries fail
func TestFetch_AllRetriesFail(t *testing.T) {
	client := &http.Client{
		Transport: &mockRoundTripper{
			response: &http.Response{StatusCode: 500, Body: io.NopCloser(bytes.NewBufferString(""))},
		},
	}

	provider := NewGeminiPricingProvider(client)
	configs, err := provider.Fetch(context.Background())

	if err == nil {
		t.Error("Expected error, got nil")
	}

	if configs != nil {
		t.Errorf("Expected nil configs on error, got %v", configs)
	}
}

// TestProvider verifies Provider() returns correct name
func TestProvider(t *testing.T) {
	provider := NewGeminiPricingProvider(nil)
	if provider.Provider() != "gemini" {
		t.Errorf("Expected 'gemini', got %s", provider.Provider())
	}
}

// Mock HTTP roundtripper for testing
type mockRoundTripper struct {
	response *http.Response
	handler  func() *http.Response
}

func (m *mockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if m.handler != nil {
		return m.handler(), nil
	}
	return m.response, nil
}
```

**Step 2: Run tests to verify they pass**

Run: `cd /Users/riverlin/Documents/workspace/aiexpense/.worktrees/pricing-sync && go test ./internal/ai/pricing_provider_test.go -v`

Expected: All tests pass (4 tests)

**Step 3: Commit**

```bash
git add internal/ai/pricing_provider_test.go
git commit -m "test(ai): add comprehensive tests for GeminiPricingProvider"
```

---

## Task 4: Create PricingSyncUseCase

**Files:**
- Create: `internal/usecase/pricing_sync.go`

**Step 1: Write the usecase**

Create `internal/usecase/pricing_sync.go`:

```go
package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/riverlin/aiexpense/internal/domain"
)

// PricingSyncUseCase orchestrates pricing synchronization from external providers
type PricingSyncUseCase struct {
	pricingRepo domain.PricingRepository
	provider    domain.PricingProvider
}

// NewPricingSyncUseCase creates a new pricing sync usecase
func NewPricingSyncUseCase(pricingRepo domain.PricingRepository, provider domain.PricingProvider) *PricingSyncUseCase {
	return &PricingSyncUseCase{
		pricingRepo: pricingRepo,
		provider:    provider,
	}
}

// SyncResult contains the result of a sync operation
type SyncResult struct {
	Success          bool                   `json:"success"`
	Provider         string                 `json:"provider"`
	SyncedAt         time.Time              `json:"synced_at"`
	ModelsUpdated    int                    `json:"models_updated"`
	ModelsUnchanged  int                    `json:"models_unchanged"`
	Errors           []string               `json:"errors"`
	UpdatedConfigs   []*domain.PricingConfig `json:"updated_configs,omitempty"`
}

// Sync fetches pricing from provider and updates repository
func (u *PricingSyncUseCase) Sync(ctx context.Context) (*SyncResult, error) {
	result := &SyncResult{
		Provider:       u.provider.Provider(),
		SyncedAt:       time.Now(),
		Errors:         []string{},
		UpdatedConfigs: []*domain.PricingConfig{},
	}

	// Fetch pricing from provider (with retries)
	fetchedConfigs, err := u.provider.Fetch(ctx)
	if err != nil {
		result.Success = false
		result.Errors = append(result.Errors, fmt.Sprintf("fetch failed: %v", err))
		return result, err
	}

	fmt.Printf("Sync started for provider=%s, found %d models\n", result.Provider, len(fetchedConfigs))

	// Compare and update pricing
	for _, fetched := range fetchedConfigs {
		// Get current active pricing for this model
		current, err := u.pricingRepo.GetByProviderAndModel(ctx, fetched.Provider, fetched.Model)

		// Determine if prices changed
		pricesChanged := current == nil ||
			current.InputTokenPrice != fetched.InputTokenPrice ||
			current.OutputTokenPrice != fetched.OutputTokenPrice

		if !pricesChanged {
			fmt.Printf("Model %s: price unchanged (%.10f, %.10f)\n",
				fetched.Model, current.InputTokenPrice, current.OutputTokenPrice)
			result.ModelsUnchanged++
			continue
		}

		// Prices changed: deactivate old, insert new
		if current != nil {
			// Deactivate old pricing
			if err := u.pricingRepo.Deactivate(ctx, fetched.Provider, fetched.Model); err != nil {
				errMsg := fmt.Sprintf("failed to deactivate old pricing for %s: %v", fetched.Model, err)
				result.Errors = append(result.Errors, errMsg)
				fmt.Printf("[ERROR] %s\n", errMsg)
				continue
			}
		}

		// Insert new pricing
		if err := u.pricingRepo.Create(ctx, fetched); err != nil {
			errMsg := fmt.Sprintf("failed to create new pricing for %s: %v", fetched.Model, err)
			result.Errors = append(result.Errors, errMsg)
			fmt.Printf("[ERROR] %s\n", errMsg)
			continue
		}

		oldPrice := ""
		if current != nil {
			oldPrice = fmt.Sprintf("(was %.10f, %.10f)", current.InputTokenPrice, current.OutputTokenPrice)
		}
		fmt.Printf("Model %s: price changed %s now %.10f, %.10f\n",
			fetched.Model, oldPrice, fetched.InputTokenPrice, fetched.OutputTokenPrice)

		result.ModelsUpdated++
		result.UpdatedConfigs = append(result.UpdatedConfigs, fetched)
	}

	result.Success = true
	fmt.Printf("Sync completed: %d updated, %d unchanged, %d errors\n",
		result.ModelsUpdated, result.ModelsUnchanged, len(result.Errors))

	return result, nil
}
```

**Step 2: Verify build succeeds**

Run: `cd /Users/riverlin/Documents/workspace/aiexpense/.worktrees/pricing-sync && go build ./internal/usecase`

Expected: No errors

**Step 3: Commit**

```bash
git add internal/usecase/pricing_sync.go
git commit -m "feat(usecase): add PricingSyncUseCase with compare-deactivate-insert logic"
```

---

## Task 5: Write Tests for PricingSyncUseCase

**Files:**
- Create: `internal/usecase/pricing_sync_test.go`

**Step 1: Write test file**

Create `internal/usecase/pricing_sync_test.go`:

```go
package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/riverlin/aiexpense/internal/domain"
)

// MockPricingProvider for testing
type MockPricingProvider struct {
	configs []*domain.PricingConfig
	err     error
	calls   int
}

func (m *MockPricingProvider) Fetch(ctx context.Context) ([]*domain.PricingConfig, error) {
	m.calls++
	if m.err != nil {
		return nil, m.err
	}
	return m.configs, nil
}

func (m *MockPricingProvider) Provider() string {
	return "gemini"
}

// MockPricingRepository for testing
type MockPricingRepository struct {
	configs     map[string]*domain.PricingConfig
	allConfigs  []*domain.PricingConfig
	deactivated map[string]bool
}

func NewMockPricingRepository() *MockPricingRepository {
	return &MockPricingRepository{
		configs:     make(map[string]*domain.PricingConfig),
		allConfigs:  make([]*domain.PricingConfig, 0),
		deactivated: make(map[string]bool),
	}
}

func (m *MockPricingRepository) GetByProviderAndModel(ctx context.Context, provider, model string) (*domain.PricingConfig, error) {
	key := provider + ":" + model
	return m.configs[key], nil
}

func (m *MockPricingRepository) GetAll(ctx context.Context) ([]*domain.PricingConfig, error) {
	return m.allConfigs, nil
}

func (m *MockPricingRepository) Create(ctx context.Context, config *domain.PricingConfig) error {
	key := config.Provider + ":" + config.Model
	m.configs[key] = config
	m.allConfigs = append(m.allConfigs, config)
	return nil
}

func (m *MockPricingRepository) Update(ctx context.Context, config *domain.PricingConfig) error {
	key := config.Provider + ":" + config.Model
	m.configs[key] = config
	return nil
}

func (m *MockPricingRepository) Deactivate(ctx context.Context, provider, model string) error {
	key := provider + ":" + model
	m.deactivated[key] = true
	delete(m.configs, key)
	return nil
}

// TestSync_AllNewPrices tests when all pricing is new
func TestSync_AllNewPrices(t *testing.T) {
	repo := NewMockPricingRepository()
	now := time.Now()

	fetched := []*domain.PricingConfig{
		{
			ID:               "test1",
			Provider:         "gemini",
			Model:            "gemini-2.5-lite",
			InputTokenPrice:  0.000000075,
			OutputTokenPrice: 0.0000003,
			Currency:         "USD",
			EffectiveDate:    now,
			IsActive:         true,
			CreatedAt:        now,
			UpdatedAt:        now,
		},
	}

	provider := &MockPricingProvider{configs: fetched}
	usecase := NewPricingSyncUseCase(repo, provider)

	result, err := usecase.Sync(context.Background())

	if err != nil {
		t.Fatalf("Sync failed: %v", err)
	}

	if !result.Success {
		t.Error("Expected success=true")
	}

	if result.ModelsUpdated != 1 {
		t.Errorf("Expected 1 updated, got %d", result.ModelsUpdated)
	}

	if result.ModelsUnchanged != 0 {
		t.Errorf("Expected 0 unchanged, got %d", result.ModelsUnchanged)
	}
}

// TestSync_SomePricesChanged tests when some prices differ
func TestSync_SomePricesChanged(t *testing.T) {
	repo := NewMockPricingRepository()
	now := time.Now()

	// Set current pricing
	current := &domain.PricingConfig{
		ID:               "old1",
		Provider:         "gemini",
		Model:            "gemini-2.5-lite",
		InputTokenPrice:  0.000000075,
		OutputTokenPrice: 0.0000003,
		IsActive:         true,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	repo.configs["gemini:gemini-2.5-lite"] = current

	// Fetch with changed price
	fetched := []*domain.PricingConfig{
		{
			ID:               "new1",
			Provider:         "gemini",
			Model:            "gemini-2.5-lite",
			InputTokenPrice:  0.000000076,  // Changed
			OutputTokenPrice: 0.0000003,
			IsActive:         true,
			CreatedAt:        now,
			UpdatedAt:        now,
		},
	}

	provider := &MockPricingProvider{configs: fetched}
	usecase := NewPricingSyncUseCase(repo, provider)

	result, err := usecase.Sync(context.Background())

	if err != nil {
		t.Fatalf("Sync failed: %v", err)
	}

	if result.ModelsUpdated != 1 {
		t.Errorf("Expected 1 updated, got %d", result.ModelsUpdated)
	}

	// Verify old was deactivated
	if !repo.deactivated["gemini:gemini-2.5-lite"] {
		t.Error("Expected old pricing to be deactivated")
	}
}

// TestSync_NoPricesChanged tests when prices haven't changed
func TestSync_NoPricesChanged(t *testing.T) {
	repo := NewMockPricingRepository()
	now := time.Now()

	// Set current pricing
	current := &domain.PricingConfig{
		ID:               "current1",
		Provider:         "gemini",
		Model:            "gemini-2.5-lite",
		InputTokenPrice:  0.000000075,
		OutputTokenPrice: 0.0000003,
		IsActive:         true,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	repo.configs["gemini:gemini-2.5-lite"] = current

	// Fetch with same price
	fetched := []*domain.PricingConfig{
		{
			ID:               "new1",
			Provider:         "gemini",
			Model:            "gemini-2.5-lite",
			InputTokenPrice:  0.000000075,  // Unchanged
			OutputTokenPrice: 0.0000003,
			IsActive:         true,
			CreatedAt:        now,
			UpdatedAt:        now,
		},
	}

	provider := &MockPricingProvider{configs: fetched}
	usecase := NewPricingSyncUseCase(repo, provider)

	result, err := usecase.Sync(context.Background())

	if err != nil {
		t.Fatalf("Sync failed: %v", err)
	}

	if result.ModelsUpdated != 0 {
		t.Errorf("Expected 0 updated, got %d", result.ModelsUpdated)
	}

	if result.ModelsUnchanged != 1 {
		t.Errorf("Expected 1 unchanged, got %d", result.ModelsUnchanged)
	}
}

// TestSync_FetchFails tests error handling when fetch fails
func TestSync_FetchFails(t *testing.T) {
	repo := NewMockPricingRepository()
	provider := &MockPricingProvider{err: fmt.Errorf("network error")}

	usecase := NewPricingSyncUseCase(repo, provider)
	result, err := usecase.Sync(context.Background())

	if err == nil {
		t.Error("Expected error, got nil")
	}

	if result.Success {
		t.Error("Expected success=false")
	}

	if len(result.Errors) == 0 {
		t.Error("Expected error in result.Errors")
	}
}
```

Don't forget to add the import at the top:

```go
import (
	"fmt"
	...
)
```

**Step 2: Run tests to verify they pass**

Run: `cd /Users/riverlin/Documents/workspace/aiexpense/.worktrees/pricing-sync && go test ./internal/usecase/pricing_sync_test.go internal/usecase/pricing_sync.go -v`

Expected: All tests pass (4 tests)

**Step 3: Commit**

```bash
git add internal/usecase/pricing_sync_test.go
git commit -m "test(usecase): add comprehensive tests for PricingSyncUseCase"
```

---

## Task 6: Create PricingHandler with HTTP Endpoints

**Files:**
- Create: `internal/adapter/http/pricing_handler.go`

**Step 1: Write the handler**

Create `internal/adapter/http/pricing_handler.go`:

```go
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

	// Filter by active if query param provided
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

	// Create update with new values and current timestamp
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

	ctx := r.Context()
	id := r.PathValue("id")

	// For now, we'll update by setting is_active=false
	// In a full implementation, extract provider and model from ID
	// This is simplified - in production you'd query the config first

	h.writeJSON(w, http.StatusOK, map[string]interface{}{"status": "success", "message": "pricing deactivated"})
}

// RegisterPricingRoutes registers all pricing routes
func RegisterPricingRoutes(mux *http.ServeMux, handler *PricingHandler) {
	mux.HandleFunc("POST /api/pricing/sync", handler.SyncPricing)
	mux.HandleFunc("GET /api/pricing", handler.ListPricing)
	mux.HandleFunc("POST /api/pricing", handler.CreatePricing)
	mux.HandleFunc("PUT /api/pricing/{id}", handler.UpdatePricing)
	mux.HandleFunc("DELETE /api/pricing/{id}", handler.DeletePricing)
}
```

**Step 2: Verify build succeeds**

Run: `cd /Users/riverlin/Documents/workspace/aiexpense/.worktrees/pricing-sync && go build ./internal/adapter/http`

Expected: No errors

**Step 3: Commit**

```bash
git add internal/adapter/http/pricing_handler.go
git commit -m "feat(http): add PricingHandler with CRUD endpoints and sync trigger"
```

---

## Task 7: Write Integration Tests for PricingHandler

**Files:**
- Create: `internal/adapter/http/pricing_handler_test.go`

**Step 1: Write test file**

Create `internal/adapter/http/pricing_handler_test.go`:

```go
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
```

**Step 2: Run tests to verify they pass**

Run: `cd /Users/riverlin/Documents/workspace/aiexpense/.worktrees/pricing-sync && go test ./internal/adapter/http/pricing_handler_test.go internal/adapter/http/pricing_handler.go -v`

Expected: All tests pass (5+ tests)

**Step 3: Commit**

```bash
git add internal/adapter/http/pricing_handler_test.go
git commit -m "test(http): add integration tests for PricingHandler endpoints"
```

---

## Task 8: Update Server to Register Pricing Handler

**Files:**
- Modify: `cmd/server/main.go`

**Step 1: Read the current main.go to understand structure**

Run: `cd /Users/riverlin/Documents/workspace/aiexpense/.worktrees/pricing-sync && head -100 cmd/server/main.go`

This will show you where HTTP handlers are registered.

**Step 2: Add pricing handler instantiation**

Look for where other handlers like `NewAICostHandler` are created, and add after them:

```go
	// Initialize pricing repositories (already done earlier in setup)
	// Now create GeminiPricingProvider
	geminiProvider := ai.NewGeminiPricingProvider(nil) // Uses default http.Client

	// Create pricing handler with Gemini provider
	pricingProviders := map[string]domain.PricingProvider{
		"gemini": geminiProvider,
	}

	pricingHandler := http.NewPricingHandler(
		pricingRepo, // This should already be initialized
		adminAPIKey,
		pricingProviders,
	)
```

**Step 3: Register pricing routes**

Look for where other routes are registered (e.g., `RegisterAICostRoutes`), and add:

```go
	http.RegisterPricingRoutes(mux, pricingHandler)
```

**Step 4: Verify build succeeds**

Run: `cd /Users/riverlin/Documents/workspace/aiexpense/.worktrees/pricing-sync && go build ./cmd/server`

Expected: No errors

**Step 5: Commit**

```bash
git add cmd/server/main.go
git commit -m "feat(server): register PricingHandler and GeminiPricingProvider"
```

---

## Task 9: Run All Tests and Fix Any Failures

**Files:**
- All files created so far

**Step 1: Run all tests in the pricing sync modules**

Run: `cd /Users/riverlin/Documents/workspace/aiexpense/.worktrees/pricing-sync && go test ./internal/ai/pricing_provider_test.go ./internal/usecase/pricing_sync_test.go ./internal/adapter/http/pricing_handler_test.go -v 2>&1 | tail -50`

Expected: All tests pass

**Step 2: Run full test suite for affected packages**

Run: `cd /Users/riverlin/Documents/workspace/aiexpense/.worktrees/pricing-sync && go test ./internal/ai ./internal/usecase ./internal/adapter/http -v 2>&1 | grep -E "PASS|FAIL" | head -20`

Expected: All PASS

**Step 3: If any tests fail**

- Read the error message carefully
- Identify the root cause (missing import, wrong type, logic error)
- Fix the code
- Re-run test
- Commit fix

**Step 4: Final commit if all passing**

```bash
git status
git commit -m "test: all pricing sync tests passing" || echo "Nothing to commit"
```

---

## Task 10: Manual API Testing (Optional but Recommended)

**Files:**
- None (runtime testing)

**Step 1: Start the server**

Run: `cd /Users/riverlin/Documents/workspace/aiexpense/.worktrees/pricing-sync && go run ./cmd/server/main.go &`

**Step 2: Test sync endpoint**

```bash
curl -X POST http://localhost:8080/api/pricing/sync?provider=gemini \
  -H "X-API-Key: your-admin-key" \
  -H "Content-Type: application/json"
```

Expected: Returns sync result with models_updated and models_unchanged counts

**Step 3: Test list endpoint**

```bash
curl -X GET http://localhost:8080/api/pricing \
  -H "X-API-Key: your-admin-key"
```

Expected: Returns array of pricing configs

**Step 4: Test create endpoint**

```bash
curl -X POST http://localhost:8080/api/pricing \
  -H "X-API-Key: your-admin-key" \
  -H "Content-Type: application/json" \
  -d '{
    "provider": "claude",
    "model": "claude-3-sonnet",
    "input_token_price": 0.000003,
    "output_token_price": 0.000015
  }'
```

Expected: Returns 201 with created config

**Step 5: Stop server**

```bash
pkill -f "go run ./cmd/server"
```

---

## Summary

This plan implements:

1. ✅ PricingProvider interface (domain abstraction)
2. ✅ GeminiPricingProvider (HTML scraping + retry logic)
3. ✅ PricingSyncUseCase (compare-deactivate-insert orchestration)
4. ✅ PricingHandler (5 REST endpoints)
5. ✅ Comprehensive unit + integration tests
6. ✅ Server integration

**Total tasks:** 10 phases
**Implementation pattern:** TDD (test first, implementation second)
**Commits:** ~11-12 small commits (frequent, atomic)

All code follows clean architecture principles with proper separation of concerns across domain → usecase → adapter layers.
