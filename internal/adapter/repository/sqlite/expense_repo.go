package sqlite

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

func hydrateExpenseAmounts(expense *domain.Expense) {
	if expense == nil {
		return
	}
	expense.Amount = expense.HomeAmount
}

func normalizeExpenseForWrite(expense *domain.Expense) {
	if expense == nil {
		return
	}
	if expense.Account == "" {
		expense.Account = "Cash"
	}
	if expense.OriginalAmount == 0 && expense.Amount != 0 {
		expense.OriginalAmount = expense.Amount
	}
	if expense.HomeAmount == 0 {
		if expense.Amount != 0 {
			expense.HomeAmount = expense.Amount
		} else if expense.OriginalAmount != 0 {
			expense.HomeAmount = expense.OriginalAmount
		}
	}
	if expense.HomeCurrency == "" {
		expense.HomeCurrency = expense.Currency
	}
	if expense.Currency == "" {
		expense.Currency = expense.HomeCurrency
	}
	if expense.Currency == "" {
		expense.Currency = "TWD"
	}
	if expense.HomeCurrency == "" {
		expense.HomeCurrency = expense.Currency
	}
	if expense.ExchangeRate == 0 {
		expense.ExchangeRate = 1.0
	}
	if expense.HomeAmount == 0 {
		expense.HomeAmount = expense.OriginalAmount
	}
	if expense.Amount == 0 {
		expense.Amount = expense.HomeAmount
	}
}

// NewExpenseRepository creates a new expense repository
func NewExpenseRepository(db *sql.DB) *ExpenseRepository {
	return &ExpenseRepository{db: db}
}

// Create creates a new expense
func (r *ExpenseRepository) Create(ctx context.Context, expense *domain.Expense) error {
	const query = `
		INSERT INTO expenses (
			id,
			user_id,
			description,
			original_amount,
			currency,
			home_amount,
			home_currency,
			exchange_rate,
			category_id,
			account,
			expense_date,
			created_at,
			updated_at
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	normalizeExpenseForWrite(expense)
	_, err := r.db.ExecContext(
		ctx,
		query,
		expense.ID,
		expense.UserID,
		expense.Description,
		expense.OriginalAmount,
		expense.Currency,
		expense.HomeAmount,
		expense.HomeCurrency,
		expense.ExchangeRate,
		expense.CategoryID,
		expense.Account,
		expense.ExpenseDate,
		expense.CreatedAt,
		expense.UpdatedAt,
	)
	return err
}

// GetByID retrieves an expense by ID
func (r *ExpenseRepository) GetByID(ctx context.Context, id string) (*domain.Expense, error) {
	const query = `
		SELECT id, user_id, description, original_amount, currency, home_amount, home_currency, exchange_rate, category_id, account, expense_date, created_at, updated_at
		FROM expenses
		WHERE id = ?
	`
	expense := &domain.Expense{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&expense.ID,
		&expense.UserID,
		&expense.Description,
		&expense.OriginalAmount,
		&expense.Currency,
		&expense.HomeAmount,
		&expense.HomeCurrency,
		&expense.ExchangeRate,
		&expense.CategoryID,
		&expense.Account,
		&expense.ExpenseDate,
		&expense.CreatedAt,
		&expense.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	hydrateExpenseAmounts(expense)
	return expense, nil
}

// GetByUserID retrieves all expenses for a user
func (r *ExpenseRepository) GetByUserID(ctx context.Context, userID string) ([]*domain.Expense, error) {
	const query = `
		SELECT id, user_id, description, original_amount, currency, home_amount, home_currency, exchange_rate, category_id, account, expense_date, created_at, updated_at
		FROM expenses
		WHERE user_id = ?
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
			&expense.ID,
			&expense.UserID,
			&expense.Description,
			&expense.OriginalAmount,
			&expense.Currency,
			&expense.HomeAmount,
			&expense.HomeCurrency,
			&expense.ExchangeRate,
			&expense.CategoryID,
			&expense.Account,
			&expense.ExpenseDate,
			&expense.CreatedAt,
			&expense.UpdatedAt,
		); err != nil {
			return nil, err
		}
		hydrateExpenseAmounts(expense)
		expenses = append(expenses, expense)
	}
	return expenses, rows.Err()
}

// GetByUserIDAndDateRange retrieves expenses for a user within a date range
func (r *ExpenseRepository) GetByUserIDAndDateRange(ctx context.Context, userID string, from, to time.Time) ([]*domain.Expense, error) {
	const query = `
		SELECT id, user_id, description, original_amount, currency, home_amount, home_currency, exchange_rate, category_id, account, expense_date, created_at, updated_at
		FROM expenses
		WHERE user_id = ? AND expense_date >= ? AND expense_date <= ?
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
			&expense.ID,
			&expense.UserID,
			&expense.Description,
			&expense.OriginalAmount,
			&expense.Currency,
			&expense.HomeAmount,
			&expense.HomeCurrency,
			&expense.ExchangeRate,
			&expense.CategoryID,
			&expense.Account,
			&expense.ExpenseDate,
			&expense.CreatedAt,
			&expense.UpdatedAt,
		); err != nil {
			return nil, err
		}
		hydrateExpenseAmounts(expense)
		expenses = append(expenses, expense)
	}
	return expenses, rows.Err()
}

// GetByUserIDAndCategory retrieves expenses for a user in a category
func (r *ExpenseRepository) GetByUserIDAndCategory(ctx context.Context, userID, categoryID string) ([]*domain.Expense, error) {
	const query = `
		SELECT id, user_id, description, original_amount, currency, home_amount, home_currency, exchange_rate, category_id, account, expense_date, created_at, updated_at
		FROM expenses
		WHERE user_id = ? AND category_id = ?
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
			&expense.ID,
			&expense.UserID,
			&expense.Description,
			&expense.OriginalAmount,
			&expense.Currency,
			&expense.HomeAmount,
			&expense.HomeCurrency,
			&expense.ExchangeRate,
			&expense.CategoryID,
			&expense.Account,
			&expense.ExpenseDate,
			&expense.CreatedAt,
			&expense.UpdatedAt,
		); err != nil {
			return nil, err
		}
		hydrateExpenseAmounts(expense)
		expenses = append(expenses, expense)
	}
	return expenses, rows.Err()
}

// Update updates an existing expense
func (r *ExpenseRepository) Update(ctx context.Context, expense *domain.Expense) error {
	const query = `
		UPDATE expenses
		SET description = ?, original_amount = ?, currency = ?, home_amount = ?, home_currency = ?, exchange_rate = ?, category_id = ?, account = ?, expense_date = ?, updated_at = ?
		WHERE id = ?
	`
	normalizeExpenseForWrite(expense)
	_, err := r.db.ExecContext(ctx, query,
		expense.Description,
		expense.OriginalAmount,
		expense.Currency,
		expense.HomeAmount,
		expense.HomeCurrency,
		expense.ExchangeRate,
		expense.CategoryID,
		expense.Account,
		expense.ExpenseDate,
		time.Now(),
		expense.ID,
	)
	return err
}

// Delete deletes an expense
func (r *ExpenseRepository) Delete(ctx context.Context, id string) error {
	const query = `DELETE FROM expenses WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}
