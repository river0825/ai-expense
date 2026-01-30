package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/riverlin/aiexpense/internal/domain"
)

var _ domain.CurrencyRepository = (*CurrencyRepository)(nil)

// CurrencyRepository provides access to currencies stored in SQLite
type CurrencyRepository struct {
	db *sql.DB
}

// NewCurrencyRepository creates a new CurrencyRepository
func NewCurrencyRepository(db *sql.DB) *CurrencyRepository {
	return &CurrencyRepository{db: db}
}

// GetAll returns all currencies ordered by code
func (r *CurrencyRepository) GetAll(ctx context.Context) ([]*domain.Currency, error) {
	const query = `SELECT code, symbol, aliases, is_active, created_at, updated_at FROM currencies ORDER BY code`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var currencies []*domain.Currency
	for rows.Next() {
		currency, err := scanCurrency(rows)
		if err != nil {
			return nil, err
		}
		currencies = append(currencies, currency)
	}
	return currencies, rows.Err()
}

// GetByCode returns a single currency
func (r *CurrencyRepository) GetByCode(ctx context.Context, code string) (*domain.Currency, error) {
	const query = `SELECT code, symbol, aliases, is_active, created_at, updated_at FROM currencies WHERE code = ?`
	row := r.db.QueryRowContext(ctx, query, code)
	currency, err := scanCurrency(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return currency, nil
}

// GetName returns the localized currency name
func (r *CurrencyRepository) GetName(ctx context.Context, code, locale string) (string, error) {
	const query = `SELECT name FROM currency_translations WHERE currency_code = ? AND locale = ?`
	var name string
	err := r.db.QueryRowContext(ctx, query, code, locale).Scan(&name)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", nil
		}
		return "", err
	}
	return name, nil
}

// Upsert inserts or updates a currency definition
func (r *CurrencyRepository) Upsert(ctx context.Context, currency *domain.Currency) error {
	aliasesJSON, err := json.Marshal(currency.Aliases)
	if err != nil {
		return err
	}
	const query = `INSERT INTO currencies (code, symbol, aliases, is_active)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(code) DO UPDATE SET
			symbol = excluded.symbol,
			aliases = excluded.aliases,
			is_active = excluded.is_active,
			updated_at = CURRENT_TIMESTAMP`
	_, err = r.db.ExecContext(ctx, query, currency.Code, currency.Symbol, string(aliasesJSON), currency.IsActive)
	return err
}

func scanCurrency(scanner interface {
	Scan(dest ...interface{}) error
}) (*domain.Currency, error) {
	var (
		aliasesRaw sql.NullString
		currency   domain.Currency
	)
	if err := scanner.Scan(&currency.Code, &currency.Symbol, &aliasesRaw, &currency.IsActive, &currency.CreatedAt, &currency.UpdatedAt); err != nil {
		return nil, err
	}
	if aliasesRaw.Valid && aliasesRaw.String != "" {
		_ = json.Unmarshal([]byte(aliasesRaw.String), &currency.Aliases)
	}
	return &currency, nil
}
