package exchangerate

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/riverlin/aiexpense/internal/domain"
)

// Provider defines external exchange rate fetching behavior
type Provider interface {
	Name() string
	Fetch(ctx context.Context, baseCurrency string, symbols []string) ([]*domain.ExchangeRate, error)
}

// ExchangeRateAPIProvider fetches rates from exchangerate-api.com
type ExchangeRateAPIProvider struct {
	httpClient *http.Client
	baseURL    string
	apiKey     string
}

// NewExchangeRateAPIProvider creates a new provider
func NewExchangeRateAPIProvider(apiKey string, client *http.Client) *ExchangeRateAPIProvider {
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}
	return &ExchangeRateAPIProvider{
		httpClient: client,
		baseURL:    "https://v6.exchangerate-api.com/v6",
		apiKey:     apiKey,
	}
}

// Name returns provider identifier
func (p *ExchangeRateAPIProvider) Name() string {
	return "exchange-rate-api"
}

// Fetch retrieves latest rates for base currency
func (p *ExchangeRateAPIProvider) Fetch(ctx context.Context, baseCurrency string, symbols []string) ([]*domain.ExchangeRate, error) {
	if p.apiKey == "" {
		return nil, fmt.Errorf("exchange rate API key is not configured")
	}
	base := strings.TrimSpace(strings.ToUpper(baseCurrency))
	if base == "" {
		base = "USD"
	}
	endpoint := fmt.Sprintf("%s/%s/latest/%s", p.baseURL, p.apiKey, base)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return nil, fmt.Errorf("exchange API responded %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	var payload struct {
		Result             string             `json:"result"`
		BaseCode           string             `json:"base_code"`
		TimeLastUpdateUnix int64              `json:"time_last_update_unix"`
		ConversionRates    map[string]float64 `json:"conversion_rates"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, err
	}
	if strings.ToLower(payload.Result) != "success" {
		return nil, fmt.Errorf("exchange rate API returned result=%s", payload.Result)
	}
	rateDate := time.Unix(payload.TimeLastUpdateUnix, 0).UTC()
	var rates []*domain.ExchangeRate
	for code, rate := range payload.ConversionRates {
		upperCode := strings.ToUpper(code)
		if strings.EqualFold(upperCode, payload.BaseCode) {
			continue
		}
		if len(symbols) > 0 && !containsSymbol(symbols, upperCode) {
			continue
		}
		rates = append(rates, &domain.ExchangeRate{
			Provider:       p.Name(),
			BaseCurrency:   strings.ToUpper(payload.BaseCode),
			TargetCurrency: upperCode,
			Rate:           rate,
			RateDate:       rateDate,
			FetchedAt:      time.Now().UTC(),
		})
	}
	return rates, nil
}

func containsSymbol(symbols []string, target string) bool {
	for _, sym := range symbols {
		if strings.EqualFold(sym, target) {
			return true
		}
	}
	return false
}
