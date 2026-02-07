package usecase

import (
	"github.com/riverlin/aiexpense/internal/domain"
)

// GetUserAggregateUseCase fetches all user settings in one call
type GetUserAggregateUseCase struct {
	userRepo     domain.UserRepository
	categoryRepo domain.CategoryRepository
	accountRepo  domain.AccountRepository
}

// NewGetUserAggregateUseCase creates a new usecase instance
func NewGetUserAggregateUseCase(
	userRepo domain.UserRepository,
	categoryRepo domain.CategoryRepository,
	accountRepo domain.AccountRepository,
) *GetUserAggregateUseCase {
	return &GetUserAggregateUseCase{
		userRepo:     userRepo,
		categoryRepo: categoryRepo,
		accountRepo:  accountRepo,
	}
}

// Execute fetches aggregated user data
func (uc *GetUserAggregateUseCase) Execute(userID string) (*domain.AggregateSettings, error) {
	// TODO: Implement fetching logic
	return nil, nil
}
