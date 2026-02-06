package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/riverlin/aiexpense/internal/domain"
)

// ManageCategoryUseCase handles managing user expense categories
type ManageCategoryUseCase struct {
	categoryRepo domain.CategoryRepository
	expenseRepo  domain.ExpenseRepository
}

// NewManageCategoryUseCase creates a new manage category use case
func NewManageCategoryUseCase(
	categoryRepo domain.CategoryRepository,
	expenseRepo domain.ExpenseRepository,
) *ManageCategoryUseCase {
	return &ManageCategoryUseCase{
		categoryRepo: categoryRepo,
		expenseRepo:  expenseRepo,
	}
}

// CreateCategoryRequest represents a request to create a category
type CreateCategoryRequest struct {
	UserID      string
	Name        string
	Description string   // Optional description
	Keywords    []string // Optional keywords to map to this category
}

// CategoryResponse represents a response with category info
type CategoryResponse struct {
	ID          string   `json:"id"`
	UserID      string   `json:"user_id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	IsDefault   bool     `json:"is_default"`
	Keywords    []string `json:"keywords"`
	Message     string   `json:"message"`
}

// CreateCategory creates a new custom category for a user
func (u *ManageCategoryUseCase) CreateCategory(ctx context.Context, req *CreateCategoryRequest) (*CategoryResponse, error) {
	if req.Name == "" {
		return nil, fmt.Errorf("category name is required")
	}

	// Check if category already exists
	existing, err := u.categoryRepo.GetByUserIDAndName(ctx, req.UserID, req.Name)
	if err == nil && existing != nil {
		return nil, fmt.Errorf("category '%s' already exists", req.Name)
	}

	// Create the category
	category := &domain.Category{
		ID:          uuid.New().String(),
		UserID:      req.UserID,
		Name:        req.Name,
		Description: req.Description,
		IsDefault:   false,
		CreatedAt:   time.Now(),
	}

	if err := u.categoryRepo.Create(ctx, category); err != nil {
		return nil, fmt.Errorf("failed to create category: %w", err)
	}

	// Add keywords if provided
	var keywords []string
	for _, keyword := range req.Keywords {
		if keyword == "" {
			continue
		}

		kw := &domain.CategoryKeyword{
			ID:         uuid.New().String(),
			CategoryID: category.ID,
			Keyword:    keyword,
			Priority:   1,
			CreatedAt:  time.Now(),
		}

		if err := u.categoryRepo.CreateKeyword(ctx, kw); err != nil {
			// Log error but continue with other keywords
			continue
		}

		keywords = append(keywords, keyword)
	}

	return &CategoryResponse{
		ID:          category.ID,
		UserID:      category.UserID,
		Name:        category.Name,
		Description: category.Description,
		IsDefault:   category.IsDefault,
		Keywords:    keywords,
		Message:     fmt.Sprintf("Category '%s' created successfully", category.Name),
	}, nil
}

// UpdateCategoryRequest represents a request to update a category
type UpdateCategoryRequest struct {
	UserID      string
	ID          string
	Name        *string
	Description *string
	Keywords    []string // If provided, replaces existing keywords
}

// UpdateCategory updates an existing category
func (u *ManageCategoryUseCase) UpdateCategory(ctx context.Context, req *UpdateCategoryRequest) (*CategoryResponse, error) {
	// Get existing category
	category, err := u.categoryRepo.GetByID(ctx, req.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get category: %w", err)
	}

	if category == nil {
		return nil, fmt.Errorf("category not found")
	}

	// Verify ownership
	if category.UserID != req.UserID {
		return nil, fmt.Errorf("unauthorized: user does not own this category")
	}

	// Don't allow updating default categories
	if category.IsDefault {
		return nil, fmt.Errorf("cannot update default categories")
	}

	if req.Name != nil && *req.Name != "" && *req.Name != category.Name {
		existing, err := u.categoryRepo.GetByUserIDAndName(ctx, req.UserID, *req.Name)
		if err == nil && existing != nil {
			return nil, fmt.Errorf("category with name '%s' already exists", *req.Name)
		}
	}

	// Update name if provided
	if req.Name != nil && *req.Name != "" {
		category.Name = *req.Name
	}

	// Update description if provided
	if req.Description != nil {
		category.Description = *req.Description
	}

	// Update in database
	if err := u.categoryRepo.Update(ctx, category); err != nil {
		return nil, fmt.Errorf("failed to update category: %w", err)
	}

	// Handle keywords if provided
	var keywords []string
	if req.Keywords != nil {
		// Get existing keywords
		existingKws, _ := u.categoryRepo.GetKeywordsByCategory(ctx, req.ID)

		// Delete all existing keywords
		for _, kw := range existingKws {
			u.categoryRepo.DeleteKeyword(ctx, kw.ID)
		}

		// Add new keywords
		for _, keyword := range req.Keywords {
			if keyword == "" {
				continue
			}

			kw := &domain.CategoryKeyword{
				ID:         uuid.New().String(),
				CategoryID: category.ID,
				Keyword:    keyword,
				Priority:   1,
				CreatedAt:  time.Now(),
			}

			if err := u.categoryRepo.CreateKeyword(ctx, kw); err != nil {
				continue
			}

			keywords = append(keywords, keyword)
		}
	} else {
		// Get existing keywords if not updating them
		existingKws, _ := u.categoryRepo.GetKeywordsByCategory(ctx, req.ID)
		for _, kw := range existingKws {
			keywords = append(keywords, kw.Keyword)
		}
	}

	return &CategoryResponse{
		ID:          category.ID,
		UserID:      category.UserID,
		Name:        category.Name,
		Description: category.Description,
		IsDefault:   category.IsDefault,
		Keywords:    keywords,
		Message:     fmt.Sprintf("Category '%s' updated successfully", category.Name),
	}, nil
}

// DeleteCategoryRequest represents a request to delete a category
type DeleteCategoryRequest struct {
	UserID string
	ID     string
}

// DeleteCategory deletes a user category (only if it has no expenses)
func (u *ManageCategoryUseCase) DeleteCategory(ctx context.Context, req *DeleteCategoryRequest) (*CategoryResponse, error) {
	// Get category
	category, err := u.categoryRepo.GetByID(ctx, req.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get category: %w", err)
	}

	if category == nil {
		return nil, fmt.Errorf("category not found")
	}

	// Verify ownership
	if category.UserID != req.UserID {
		return nil, fmt.Errorf("unauthorized: user does not own this category")
	}

	// Don't allow deleting default categories
	if category.IsDefault {
		return nil, fmt.Errorf("cannot delete default categories")
	}

	// Delete associated keywords
	keywords, _ := u.categoryRepo.GetKeywordsByCategory(ctx, req.ID)
	for _, kw := range keywords {
		u.categoryRepo.DeleteKeyword(ctx, kw.ID)
	}

	// Delete the category
	if err := u.categoryRepo.Delete(ctx, req.ID); err != nil {
		return nil, fmt.Errorf("failed to delete category: %w", err)
	}

	return &CategoryResponse{
		ID:          req.ID,
		UserID:      category.UserID,
		Name:        category.Name,
		Description: category.Description,
		IsDefault:   category.IsDefault,
		Message:     fmt.Sprintf("Category '%s' deleted successfully", category.Name),
	}, nil
}

// ListCategoriesRequest represents a request to list categories
type ListCategoriesRequest struct {
	UserID string
}

// ListCategoriesResponse represents a list of categories
type ListCategoriesResponse struct {
	Categories []*CategoryResponse `json:"categories"`
	Total      int                 `json:"total"`
}

// ListCategories retrieves all categories for a user
func (u *ManageCategoryUseCase) ListCategories(ctx context.Context, req *ListCategoriesRequest) (*ListCategoriesResponse, error) {
	categories, err := u.categoryRepo.GetByUserID(ctx, req.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to list categories: %w", err)
	}

	var result []*CategoryResponse
	for _, cat := range categories {
		keywords, _ := u.categoryRepo.GetKeywordsByCategory(ctx, cat.ID)
		var kwList []string
		for _, kw := range keywords {
			kwList = append(kwList, kw.Keyword)
		}

		result = append(result, &CategoryResponse{
			ID:          cat.ID,
			UserID:      cat.UserID,
			Name:        cat.Name,
			Description: cat.Description,
			IsDefault:   cat.IsDefault,
			Keywords:    kwList,
		})
	}

	return &ListCategoriesResponse{
		Categories: result,
		Total:      len(result),
	}, nil
}

type MergeCategoriesRequest struct {
	UserID   string
	SourceID string
	TargetID string
}

type MergeCategoriesResponse struct {
	MergedCount int
	Message     string
}

func (u *ManageCategoryUseCase) MergeCategories(ctx context.Context, req *MergeCategoriesRequest) (*MergeCategoriesResponse, error) {
	if req.SourceID == req.TargetID {
		return nil, fmt.Errorf("source and target categories must be different")
	}

	sourceCategory, err := u.categoryRepo.GetByID(ctx, req.SourceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get source category: %w", err)
	}
	if sourceCategory == nil {
		return nil, fmt.Errorf("source category not found")
	}
	if sourceCategory.UserID != req.UserID {
		return nil, fmt.Errorf("unauthorized: user does not own source category")
	}
	if sourceCategory.IsDefault {
		return nil, fmt.Errorf("cannot merge default categories")
	}

	targetCategory, err := u.categoryRepo.GetByID(ctx, req.TargetID)
	if err != nil {
		return nil, fmt.Errorf("failed to get target category: %w", err)
	}
	if targetCategory == nil {
		return nil, fmt.Errorf("target category not found")
	}
	if targetCategory.UserID != req.UserID {
		return nil, fmt.Errorf("unauthorized: user does not own target category")
	}
	if targetCategory.IsDefault {
		return nil, fmt.Errorf("cannot merge into default categories")
	}

	mergedCount, err := u.expenseRepo.ReassignExpenses(ctx, req.SourceID, req.TargetID)
	if err != nil {
		return nil, fmt.Errorf("failed to reassign expenses: %w", err)
	}

	keywords, _ := u.categoryRepo.GetKeywordsByCategory(ctx, req.SourceID)
	for _, kw := range keywords {
		u.categoryRepo.DeleteKeyword(ctx, kw.ID)
	}

	if err := u.categoryRepo.Delete(ctx, req.SourceID); err != nil {
		return nil, fmt.Errorf("failed to delete source category: %w", err)
	}

	return &MergeCategoriesResponse{
		MergedCount: mergedCount,
		Message:     fmt.Sprintf("Merged %d expenses from '%s' into '%s'", mergedCount, sourceCategory.Name, targetCategory.Name),
	}, nil
}
