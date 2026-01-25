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
	Success         bool                    `json:"success"`
	Provider        string                  `json:"provider"`
	SyncedAt        time.Time               `json:"synced_at"`
	ModelsUpdated   int                     `json:"models_updated"`
	ModelsUnchanged int                     `json:"models_unchanged"`
	Errors          []string                `json:"errors"`
	UpdatedConfigs  []*domain.PricingConfig `json:"updated_configs,omitempty"`
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
		current, err := u.pricingRepo.GetByProviderAndModel(ctx, fetched.Provider, fetched.Model)

		pricesChanged := false
		if err != nil || current == nil {
			pricesChanged = true
		} else {
			pricesChanged = current.InputTokenPrice != fetched.InputTokenPrice ||
				current.OutputTokenPrice != fetched.OutputTokenPrice
		}

		if !pricesChanged {
			fmt.Printf("Model %s: price unchanged (%.10f, %.10f)\n",
				fetched.Model, current.InputTokenPrice, current.OutputTokenPrice)
			result.ModelsUnchanged++
			continue
		}

		if current != nil {
			if err := u.pricingRepo.Deactivate(ctx, fetched.Provider, fetched.Model); err != nil {
				errMsg := fmt.Sprintf("failed to deactivate old pricing for %s: %v", fetched.Model, err)
				result.Errors = append(result.Errors, errMsg)
				fmt.Printf("[ERROR] %s\n", errMsg)
				continue
			}
		}

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
