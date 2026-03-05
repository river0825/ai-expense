package postgresql

import (
	"context"
	"database/sql"
	"errors"

	"github.com/riverlin/aiexpense/internal/domain"
)

var _ domain.UserRepository = (*UserRepository)(nil)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, user *domain.User) error {
	const query = `
		INSERT INTO users (user_id, messenger_type, created_at, home_currency, locale, default_input_currency)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	homeCurrency := user.HomeCurrency
	if homeCurrency == "" {
		homeCurrency = "TWD"
	}
	locale := user.Locale
	if locale == "" {
		locale = "zh-TW"
	}
	var defaultInputCurrency sql.NullString
	if user.DefaultInputCurrency != "" {
		defaultInputCurrency = sql.NullString{String: user.DefaultInputCurrency, Valid: true}
	}
	_, err := r.db.ExecContext(ctx, query,
		user.UserID,
		user.MessengerType,
		user.CreatedAt,
		homeCurrency,
		locale,
		defaultInputCurrency,
	)
	return err
}

func (r *UserRepository) GetByID(ctx context.Context, userID string) (*domain.User, error) {
	const query = `
		SELECT user_id, messenger_type, created_at, home_currency, locale, default_input_currency
		FROM users
		WHERE user_id = $1
	`

	user := &domain.User{}
	var defaultInputCurrency sql.NullString
	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&user.UserID,
		&user.MessengerType,
		&user.CreatedAt,
		&user.HomeCurrency,
		&user.Locale,
		&defaultInputCurrency,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	if defaultInputCurrency.Valid {
		user.DefaultInputCurrency = defaultInputCurrency.String
	}
	return user, nil
}

func (r *UserRepository) Exists(ctx context.Context, userID string) (bool, error) {
	const query = `SELECT 1 FROM users WHERE user_id = $1`
	var exists int
	err := r.db.QueryRowContext(ctx, query, userID).Scan(&exists)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (r *UserRepository) Update(ctx context.Context, user *domain.User) error {
	const query = `
		UPDATE users
		SET home_currency = $2, locale = $3, default_input_currency = $4
		WHERE user_id = $1
	`
	var defaultInputCurrency sql.NullString
	if user.DefaultInputCurrency != "" {
		defaultInputCurrency = sql.NullString{String: user.DefaultInputCurrency, Valid: true}
	}
	_, err := r.db.ExecContext(ctx, query, user.UserID, user.HomeCurrency, user.Locale, defaultInputCurrency)
	return err
}
