package postgresql

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

func NewCategoryRepository(db *sql.DB) *CategoryRepository {
	return &CategoryRepository{db: db}
}

func (r *CategoryRepository) Create(ctx context.Context, category *domain.Category) error {
	const query = `
		INSERT INTO categories (id, user_id, name, is_default, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`

	_, err := r.db.ExecContext(ctx, query,
		category.ID, category.UserID, category.Name,
		category.IsDefault, category.CreatedAt,
	)
	return err
}

func (r *CategoryRepository) GetByID(ctx context.Context, id string) (*domain.Category, error) {
	const query = `
		SELECT id, user_id, name, is_default, created_at
		FROM categories
		WHERE id = $1
	`

	category := &domain.Category{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&category.ID, &category.UserID, &category.Name,
		&category.IsDefault, &category.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return category, nil
}

func (r *CategoryRepository) GetByUserID(ctx context.Context, userID string) ([]*domain.Category, error) {
	const query = `
		SELECT id, user_id, name, is_default, created_at
		FROM categories
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []*domain.Category
	for rows.Next() {
		category := &domain.Category{}
		if err := rows.Scan(
			&category.ID, &category.UserID, &category.Name,
			&category.IsDefault, &category.CreatedAt,
		); err != nil {
			return nil, err
		}
		categories = append(categories, category)
	}
	return categories, rows.Err()
}

func (r *CategoryRepository) GetByUserIDAndName(ctx context.Context, userID, name string) (*domain.Category, error) {
	const query = `
		SELECT id, user_id, name, is_default, created_at
		FROM categories
		WHERE user_id = $1 AND name = $2
	`

	category := &domain.Category{}
	err := r.db.QueryRowContext(ctx, query, userID, name).Scan(
		&category.ID, &category.UserID, &category.Name,
		&category.IsDefault, &category.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return category, nil
}

func (r *CategoryRepository) Update(ctx context.Context, category *domain.Category) error {
	const query = `
		UPDATE categories
		SET name = $2, is_default = $3
		WHERE id = $1
	`

	_, err := r.db.ExecContext(ctx, query,
		category.ID, category.Name, category.IsDefault,
	)
	return err
}

func (r *CategoryRepository) Delete(ctx context.Context, id string) error {
	const query = `DELETE FROM categories WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *CategoryRepository) CreateKeyword(ctx context.Context, keyword *domain.CategoryKeyword) error {
	const query = `
		INSERT INTO category_keywords (id, category_id, keyword, priority, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`

	_, err := r.db.ExecContext(ctx, query,
		keyword.ID, keyword.CategoryID, keyword.Keyword,
		keyword.Priority, keyword.CreatedAt,
	)
	return err
}

func (r *CategoryRepository) GetKeywordsByCategory(ctx context.Context, categoryID string) ([]*domain.CategoryKeyword, error) {
	const query = `
		SELECT id, category_id, keyword, priority, created_at
		FROM category_keywords
		WHERE category_id = $1
		ORDER BY priority DESC, created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, categoryID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var keywords []*domain.CategoryKeyword
	for rows.Next() {
		keyword := &domain.CategoryKeyword{}
		if err := rows.Scan(
			&keyword.ID, &keyword.CategoryID, &keyword.Keyword,
			&keyword.Priority, &keyword.CreatedAt,
		); err != nil {
			return nil, err
		}
		keywords = append(keywords, keyword)
	}
	return keywords, rows.Err()
}

func (r *CategoryRepository) DeleteKeyword(ctx context.Context, id string) error {
	const query = `DELETE FROM category_keywords WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}
