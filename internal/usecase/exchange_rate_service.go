package usecase

import (
	"context"
	"strings"
	"time"

	"github.com/riverlin/aiexpense/internal/domain"
)

var _ domain.ExchangeRateService = (*ExchangeRateService)(nil)

// ExchangeRateProvider describes external fetch capability
type ExchangeRateProvider interface {
	Name() string
	Fetch(ctx context.Context, baseCurrency string, symbols []string) ([]*domain.ExchangeRate, error)
}

// ExchangeRateService implements domain.ExchangeRateService
type ExchangeRateService struct {
	repo     domain.ExchangeRateRepository
	provider ExchangeRateProvider
}

// NewExchangeRateService creates a new service
func NewExchangeRateService(repo domain.ExchangeRateRepository, provider ExchangeRateProvider) *ExchangeRateService {
	return &ExchangeRateService{repo: repo, provider: provider}
}

// Convert converts an amount by ensuring a cached rate exists
func (s *ExchangeRateService) Convert(ctx context.Context, amount float64, fromCurrency, toCurrency string, txTime time.Time) (float64, float64, error) {
	from := strings.ToUpper(fromCurrency)
	to := strings.ToUpper(toCurrency)
	if from == "" {
		from = "TWD"
	}
	if to == "" {
		to = from
	}
	if amount == 0 || from == to {
		return amount, 1.0, nil
	}
	rate, err := s.ensureRate(ctx, from, to, txTime)
	if err != nil {
		return amount, 1.0, err
	}
	converted := amount * rate.Rate
	return converted, rate.Rate, nil
}

// RefreshRates pulls latest rates for supported base currencies
func (s *ExchangeRateService) RefreshRates(ctx context.Context) error {
	if s.provider == nil || s.repo == nil {
		return nil
	}
	bases := []string{"USD", "EUR", "TWD", "JPY", "CNY"}
	for _, base := range bases {
		if err := s.fetchAndStore(ctx, base, nil); err != nil {
			return err
		}
	}
	return nil
}

// GetRate retrieves cached rate (most recent fallback)
func (s *ExchangeRateService) GetRate(ctx context.Context, fromCurrency, toCurrency string, txTime time.Time) (*domain.ExchangeRate, error) {
	if s.repo == nil {
		return nil, nil
	}
	rate, err := s.repo.GetRate(ctx, strings.ToUpper(fromCurrency), strings.ToUpper(toCurrency), txTime)
	if err != nil || rate != nil {
		return rate, err
	}
	return s.repo.GetMostRecentRate(ctx, strings.ToUpper(fromCurrency), strings.ToUpper(toCurrency), txTime)
}

func (s *ExchangeRateService) ensureRate(ctx context.Context, fromCurrency, toCurrency string, txTime time.Time) (*domain.ExchangeRate, error) {
	if s.repo == nil {
		return &domain.ExchangeRate{BaseCurrency: fromCurrency, TargetCurrency: toCurrency, Rate: 1.0, RateDate: txTime}, nil
	}
	rate, err := s.repo.GetRate(ctx, fromCurrency, toCurrency, txTime)
	if err != nil {
		return nil, err
	}
	if rate != nil {
		return rate, nil
	}
	rate, err = s.repo.GetMostRecentRate(ctx, fromCurrency, toCurrency, txTime)
	if err != nil {
		return nil, err
	}
	if rate != nil {
		return rate, nil
	}
	if err := s.fetchAndStore(ctx, fromCurrency, []string{toCurrency}); err != nil {
		return nil, err
	}
	return s.repo.GetRate(ctx, fromCurrency, toCurrency, txTime)
}

func (s *ExchangeRateService) fetchAndStore(ctx context.Context, baseCurrency string, symbols []string) error {
	if s.provider == nil || s.repo == nil {
		return nil
	}
	rates, err := s.provider.Fetch(ctx, baseCurrency, symbols)
	if err != nil {
		return err
	}
	for _, rate := range rates {
		_ = s.repo.SaveRate(ctx, rate)
	}
	return nil
}
