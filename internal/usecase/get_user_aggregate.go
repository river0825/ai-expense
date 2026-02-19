package usecase

import (
	"context"

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
func (uc *GetUserAggregateUseCase) Execute(ctx context.Context, userID string) (*domain.AggregateSettings, error) {
	// Fetch user profile
	user, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Fetch categories
	categories, err := uc.categoryRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Fetch accounts
	accounts, err := uc.accountRepo.GetByUserID(userID)
	if err != nil {
		return nil, err
	}

	return &domain.AggregateSettings{
		Profile:    user,
		Categories: categories,
		Accounts:   accounts,
	}, nil
}
