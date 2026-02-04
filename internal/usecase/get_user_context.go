package usecase

import (
	"context"
	"errors"

	"github.com/riverlin/aiexpense/internal/domain"
)

type GetUserContextUseCase struct {
	userRepo     domain.UserRepository
	categoryRepo domain.CategoryRepository
}

func NewGetUserContextUseCase(
	userRepo domain.UserRepository,
	categoryRepo domain.CategoryRepository,
) *GetUserContextUseCase {
	return &GetUserContextUseCase{
		userRepo:     userRepo,
		categoryRepo: categoryRepo,
	}
}

func (uc *GetUserContextUseCase) Execute(ctx context.Context, userID string) (*domain.UserContext, error) {
	user, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	categories, err := uc.categoryRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if categories == nil {
		categories = []*domain.Category{}
	}

	return &domain.UserContext{
		User:       user,
		Categories: categories,
	}, nil
}
