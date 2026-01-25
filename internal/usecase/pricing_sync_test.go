package usecase

import (
	"context"
	"fmt"
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
			InputTokenPrice:  0.000000076, // Changed
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
			InputTokenPrice:  0.000000075, // Unchanged
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
