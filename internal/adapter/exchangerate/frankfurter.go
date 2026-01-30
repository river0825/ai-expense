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

// FrankfurterProvider fetches rates from frankfurter.app
type FrankfurterProvider struct {
	httpClient *http.Client
	baseURL    string
}

// NewFrankfurterProvider creates a new provider
func NewFrankfurterProvider(client *http.Client) *FrankfurterProvider {
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}
	return &FrankfurterProvider{httpClient: client, baseURL: "https://open.er-api.com/v6"}
}

// Name returns provider identifier
func (p *FrankfurterProvider) Name() string {
	return "frankfurter"
}

// Fetch retrieves latest rates for base currency
func (p *FrankfurterProvider) Fetch(ctx context.Context, baseCurrency string, symbols []string) ([]*domain.ExchangeRate, error) {
	base := strings.TrimSpace(strings.ToUpper(baseCurrency))
	if base == "" {
		base = "EUR"
	}
	endpoint := fmt.Sprintf("%s/latest/%s", p.baseURL, base)
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
		Rates              map[string]float64 `json:"rates"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, err
	}
	if payload.Result != "success" {
		return nil, fmt.Errorf("exchange API returned result=%s", payload.Result)
	}
	rateDate := time.Unix(payload.TimeLastUpdateUnix, 0).UTC()
	var rates []*domain.ExchangeRate
	for code, rate := range payload.Rates {
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
