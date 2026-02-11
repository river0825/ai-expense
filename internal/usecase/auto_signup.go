package usecase

import (
	"context"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/riverlin/aiexpense/internal/domain"
)

// ProfileFetcher fetches user profile language from a messenger platform.
type ProfileFetcher interface {
	GetLanguage(ctx context.Context, userID string) (string, error)
}

// AutoSignupUseCase handles automatic user registration
type AutoSignupUseCase struct {
	userRepo       domain.UserRepository
	categoryRepo   domain.CategoryRepository
	profileFetcher ProfileFetcher
}

// NewAutoSignupUseCase creates a new auto-signup use case
func NewAutoSignupUseCase(userRepo domain.UserRepository, categoryRepo domain.CategoryRepository, profileFetcher ProfileFetcher) *AutoSignupUseCase {
	return &AutoSignupUseCase{
		userRepo:       userRepo,
		categoryRepo:   categoryRepo,
		profileFetcher: profileFetcher,
	}
}

// Execute registers a new user and initializes default categories
func (u *AutoSignupUseCase) Execute(ctx context.Context, userID, messengerType string) error {
	// Check if user already exists
	exists, err := u.userRepo.Exists(ctx, userID)
	if err != nil {
		return err
	}

	if exists {
		// User already exists, no need to sign up
		return nil
	}

	// Create new user
	user := &domain.User{
		UserID:        userID,
		MessengerType: messengerType,
		CreatedAt:     time.Now(),
	}

	// Detect locale from messenger profile
	if u.profileFetcher != nil {
		if lang, err := u.profileFetcher.GetLanguage(ctx, userID); err == nil && lang != "" {
			user.Locale = lang
		} else if err != nil {
			log.Printf("[AutoSignup] Failed to fetch profile language for %s: %v", userID, err)
		}
	}

	if err := u.userRepo.Create(ctx, user); err != nil {
		return err
	}

	// Initialize default categories
	defaultCategoryNames := []string{"Food", "Transport", "Shopping", "Entertainment", "Other"}

	for _, name := range defaultCategoryNames {
		category := &domain.Category{
			ID:        uuid.New().String(),
			UserID:    userID,
			Name:      name,
			IsDefault: true,
			CreatedAt: time.Now(),
		}

		if err := u.categoryRepo.Create(ctx, category); err != nil {
			// Log error but continue with other categories
			continue
		}
	}

	return nil
}
