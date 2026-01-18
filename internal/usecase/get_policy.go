package usecase

import (
	"context"
	"fmt"

	"github.com/riverlin/aiexpense/internal/domain"
)

// GetPolicyUseCase handles retrieval of policy documents
type GetPolicyUseCase struct {
	policyRepo domain.PolicyRepository
}

// NewGetPolicyUseCase creates a new use case for retrieving policies
func NewGetPolicyUseCase(policyRepo domain.PolicyRepository) *GetPolicyUseCase {
	return &GetPolicyUseCase{
		policyRepo: policyRepo,
	}
}

// Execute retrieves a policy by key
func (uc *GetPolicyUseCase) Execute(ctx context.Context, key string) (*domain.Policy, error) {
	policy, err := uc.policyRepo.GetByKey(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("failed to get policy: %w", err)
	}
	if policy == nil {
		return nil, nil // Policy not found
	}
	return policy, nil
}
