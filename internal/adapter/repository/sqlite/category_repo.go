package sqlite

import (
	"context"
	"database/sql"
	"errors"

	"github.com/riverlin/aiexpense/internal/domain"
)

var _ domain.CategoryRepository = (*CategoryRepository)(nil)

type CategoryRepository struct {
	db *sql.DB
}

// NewCategoryRepository creates a new category repository
func NewCategoryRepository(db *sql.DB) *CategoryRepository {
	return &CategoryRepository{db: db}
}

// Create creates a new category
func (r *CategoryRepository) Create(ctx context.Context, category *domain.Category) error {
	const query = `
		INSERT INTO categories (id, user_id, name, is_default, created_at)
		VALUES (?, ?, ?, ?, ?)
	`
	_, err := r.db.ExecContext(ctx, query, category.ID, category.UserID, category.Name, category.IsDefault, category.CreatedAt)
	return err
}

// GetByID retrieves a category by ID
func (r *CategoryRepository) GetByID(ctx context.Context, id string) (*domain.Category, error) {
	const query = `
		SELECT id, user_id, name, is_default, created_at
		FROM categories
		WHERE id = ?
	`
	category := &domain.Category{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&category.ID,
		&category.UserID,
		&category.Name,
		&category.IsDefault,
		&category.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return category, nil
}

// GetByUserID retrieves all categories for a user
func (r *CategoryRepository) GetByUserID(ctx context.Context, userID string) ([]*domain.Category, error) {
	const query = `
		SELECT id, user_id, name, is_default, created_at
		FROM categories
		WHERE user_id = ?
		ORDER BY is_default DESC, name ASC
	`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []*domain.Category
	for rows.Next() {
		category := &domain.Category{}
		if err := rows.Scan(&category.ID, &category.UserID, &category.Name, &category.IsDefault, &category.CreatedAt); err != nil {
			return nil, err
		}
		categories = append(categories, category)
	}
	return categories, rows.Err()
}

// GetByUserIDAndName retrieves a category by user and name
func (r *CategoryRepository) GetByUserIDAndName(ctx context.Context, userID, name string) (*domain.Category, error) {
	const query = `
		SELECT id, user_id, name, is_default, created_at
		FROM categories
		WHERE user_id = ? AND name = ?
	`
	category := &domain.Category{}
	err := r.db.QueryRowContext(ctx, query, userID, name).Scan(
		&category.ID,
		&category.UserID,
		&category.Name,
		&category.IsDefault,
		&category.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return category, nil
}

// Update updates a category
func (r *CategoryRepository) Update(ctx context.Context, category *domain.Category) error {
	const query = `
		UPDATE categories
		SET name = ?, is_default = ?
		WHERE id = ?
	`
	_, err := r.db.ExecContext(ctx, query, category.Name, category.IsDefault, category.ID)
	return err
}

// Delete deletes a category
func (r *CategoryRepository) Delete(ctx context.Context, id string) error {
	const query = `DELETE FROM categories WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

// CreateKeyword creates a keyword mapping
func (r *CategoryRepository) CreateKeyword(ctx context.Context, keyword *domain.CategoryKeyword) error {
	const query = `
		INSERT INTO category_keywords (id, category_id, keyword, priority, created_at)
		VALUES (?, ?, ?, ?, ?)
	`
	_, err := r.db.ExecContext(ctx, query, keyword.ID, keyword.CategoryID, keyword.Keyword, keyword.Priority, keyword.CreatedAt)
	return err
}

// GetKeywordsByCategory retrieves keywords for a category
func (r *CategoryRepository) GetKeywordsByCategory(ctx context.Context, categoryID string) ([]*domain.CategoryKeyword, error) {
	const query = `
		SELECT id, category_id, keyword, priority, created_at
		FROM category_keywords
		WHERE category_id = ?
		ORDER BY priority DESC
	`
	rows, err := r.db.QueryContext(ctx, query, categoryID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var keywords []*domain.CategoryKeyword
	for rows.Next() {
		kw := &domain.CategoryKeyword{}
		if err := rows.Scan(&kw.ID, &kw.CategoryID, &kw.Keyword, &kw.Priority, &kw.CreatedAt); err != nil {
			return nil, err
		}
		keywords = append(keywords, kw)
	}
	return keywords, rows.Err()
}

// DeleteKeyword deletes a keyword mapping
func (r *CategoryRepository) DeleteKeyword(ctx context.Context, id string) error {
	const query = `DELETE FROM category_keywords WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}
