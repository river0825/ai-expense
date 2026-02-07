package usecase

import (
	"github.com/riverlin/aiexpense/internal/domain"
)

// UpdateUserAggregateUseCase handles updating all user settings
type UpdateUserAggregateUseCase struct {
	userRepo     domain.UserRepository
	categoryRepo domain.CategoryRepository
	accountRepo  domain.AccountRepository
	expenseRepo  domain.ExpenseRepository
}

// NewUpdateUserAggregateUseCase creates a new usecase instance
func NewUpdateUserAggregateUseCase(
	userRepo domain.UserRepository,
	categoryRepo domain.CategoryRepository,
	accountRepo domain.AccountRepository,
	expenseRepo domain.ExpenseRepository,
) *UpdateUserAggregateUseCase {
	return &UpdateUserAggregateUseCase{
		userRepo:     userRepo,
		categoryRepo: categoryRepo,
		accountRepo:  accountRepo,
		expenseRepo:  expenseRepo,
	}
}

// Execute updates aggregated user data
func (uc *UpdateUserAggregateUseCase) Execute(userID string, settings *domain.AggregateSettings) error {
	// TODO: Implement update logic
	return nil
}
