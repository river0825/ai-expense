package postgresql

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/riverlin/aiexpense/internal/domain"
)

var _ domain.ExpenseRepository = (*ExpenseRepository)(nil)

type ExpenseRepository struct {
	db *sql.DB
}

func NewExpenseRepository(db *sql.DB) *ExpenseRepository {
	return &ExpenseRepository{db: db}
}

func (r *ExpenseRepository) Create(ctx context.Context, expense *domain.Expense) error {
	const query = `
		INSERT INTO expenses (id, user_id, description, amount, category_id, expense_date, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := r.db.ExecContext(ctx, query,
		expense.ID, expense.UserID, expense.Description,
		expense.Amount, expense.CategoryID, expense.ExpenseDate,
		expense.CreatedAt, expense.UpdatedAt,
	)
	return err
}

func (r *ExpenseRepository) GetByID(ctx context.Context, id string) (*domain.Expense, error) {
	const query = `
		SELECT id, user_id, description, amount, category_id, expense_date, created_at, updated_at
		FROM expenses
		WHERE id = $1
	`

	expense := &domain.Expense{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&expense.ID, &expense.UserID, &expense.Description,
		&expense.Amount, &expense.CategoryID, &expense.ExpenseDate,
		&expense.CreatedAt, &expense.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return expense, nil
}

func (r *ExpenseRepository) GetByUserID(ctx context.Context, userID string) ([]*domain.Expense, error) {
	const query = `
		SELECT id, user_id, description, amount, category_id, expense_date, created_at, updated_at
		FROM expenses
		WHERE user_id = $1
		ORDER BY expense_date DESC, created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var expenses []*domain.Expense
	for rows.Next() {
		expense := &domain.Expense{}
		if err := rows.Scan(
			&expense.ID, &expense.UserID, &expense.Description,
			&expense.Amount, &expense.CategoryID, &expense.ExpenseDate,
			&expense.CreatedAt, &expense.UpdatedAt,
		); err != nil {
			return nil, err
		}
		expenses = append(expenses, expense)
	}
	return expenses, rows.Err()
}

func (r *ExpenseRepository) Update(ctx context.Context, expense *domain.Expense) error {
	const query = `
		UPDATE expenses
		SET description = $2, amount = $3, category_id = $4, expense_date = $5, updated_at = $6
		WHERE id = $1
	`

	_, err := r.db.ExecContext(ctx, query,
		expense.ID,
		expense.Description, expense.Amount, expense.CategoryID,
		expense.ExpenseDate, time.Now(),
	)
	return err
}

func (r *ExpenseRepository) GetByUserIDAndDateRange(ctx context.Context, userID string, from, to time.Time) ([]*domain.Expense, error) {
	const query = `
		SELECT id, user_id, description, amount, category_id, expense_date, created_at, updated_at
		FROM expenses
		WHERE user_id = $1 AND expense_date BETWEEN $2 AND $3
		ORDER BY expense_date DESC, created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var expenses []*domain.Expense
	for rows.Next() {
		expense := &domain.Expense{}
		if err := rows.Scan(
			&expense.ID, &expense.UserID, &expense.Description,
			&expense.Amount, &expense.CategoryID, &expense.ExpenseDate,
			&expense.CreatedAt, &expense.UpdatedAt,
		); err != nil {
			return nil, err
		}
		expenses = append(expenses, expense)
	}
	return expenses, rows.Err()
}

func (r *ExpenseRepository) GetByUserIDAndCategory(ctx context.Context, userID, categoryID string) ([]*domain.Expense, error) {
	const query = `
		SELECT id, user_id, description, amount, category_id, expense_date, created_at, updated_at
		FROM expenses
		WHERE user_id = $1 AND category_id = $2
		ORDER BY expense_date DESC, created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID, categoryID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var expenses []*domain.Expense
	for rows.Next() {
		expense := &domain.Expense{}
		if err := rows.Scan(
			&expense.ID, &expense.UserID, &expense.Description,
			&expense.Amount, &expense.CategoryID, &expense.ExpenseDate,
			&expense.CreatedAt, &expense.UpdatedAt,
		); err != nil {
			return nil, err
		}
		expenses = append(expenses, expense)
	}
	return expenses, rows.Err()
}

func (r *ExpenseRepository) Delete(ctx context.Context, id string) error {
	const query = `DELETE FROM expenses WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}
