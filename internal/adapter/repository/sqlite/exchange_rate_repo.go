package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/riverlin/aiexpense/internal/domain"
)

var _ domain.ExchangeRateRepository = (*ExchangeRateRepository)(nil)

const defaultRateProvider = "frankfurter"

// ExchangeRateRepository stores exchange rates in SQLite
type ExchangeRateRepository struct {
	db *sql.DB
}

// NewExchangeRateRepository creates a new ExchangeRateRepository
func NewExchangeRateRepository(db *sql.DB) *ExchangeRateRepository {
	return &ExchangeRateRepository{db: db}
}

// SaveRate upserts a rate for a given provider/base/target/day
func (r *ExchangeRateRepository) SaveRate(ctx context.Context, rate *domain.ExchangeRate) error {
	provider := rate.Provider
	if provider == "" {
		provider = defaultRateProvider
	}
	const query = `INSERT INTO exchange_rates (provider, base_currency, target_currency, rate, rate_date, fetched_at)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT(provider, base_currency, target_currency, rate_date)
		DO UPDATE SET rate = excluded.rate, fetched_at = excluded.fetched_at`
	_, err := r.db.ExecContext(ctx, query,
		provider,
		rate.BaseCurrency,
		rate.TargetCurrency,
		rate.Rate,
		rate.RateDate.Format("2006-01-02"),
		rate.FetchedAt,
	)
	return err
}

// GetRate gets an exact rate for provider/base/target/date
func (r *ExchangeRateRepository) GetRate(ctx context.Context, baseCurrency, targetCurrency string, rateDate time.Time) (*domain.ExchangeRate, error) {
	return r.getRateInternal(ctx, baseCurrency, targetCurrency, rateDate, false)
}

// GetMostRecentRate returns the latest rate before or on given date
func (r *ExchangeRateRepository) GetMostRecentRate(ctx context.Context, baseCurrency, targetCurrency string, before time.Time) (*domain.ExchangeRate, error) {
	return r.getRateInternal(ctx, baseCurrency, targetCurrency, before, true)
}

func (r *ExchangeRateRepository) getRateInternal(ctx context.Context, baseCurrency, targetCurrency string, date time.Time, allowBefore bool) (*domain.ExchangeRate, error) {
	provider := defaultRateProvider
	var query string
	var args []interface{}
	if allowBefore {
		query = `SELECT id, provider, base_currency, target_currency, rate, rate_date, fetched_at
			FROM exchange_rates
			WHERE provider = ? AND base_currency = ? AND target_currency = ? AND rate_date <= ?
			ORDER BY rate_date DESC LIMIT 1`
		args = []interface{}{provider, baseCurrency, targetCurrency, date.Format("2006-01-02")}
	} else {
		query = `SELECT id, provider, base_currency, target_currency, rate, rate_date, fetched_at
			FROM exchange_rates
			WHERE provider = ? AND base_currency = ? AND target_currency = ? AND rate_date = ?`
		args = []interface{}{provider, baseCurrency, targetCurrency, date.Format("2006-01-02")}
	}
	row := r.db.QueryRowContext(ctx, query, args...)
	var rate domain.ExchangeRate
	var rateDateStr string
	if err := row.Scan(&rate.ID, &rate.Provider, &rate.BaseCurrency, &rate.TargetCurrency, &rate.Rate, &rateDateStr, &rate.FetchedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	rate.Provider = provider
	rate.RateDate, _ = time.Parse("2006-01-02", rateDateStr)
	return &rate, nil
}
