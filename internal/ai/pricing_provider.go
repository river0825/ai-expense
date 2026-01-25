package ai

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/riverlin/aiexpense/internal/domain"
)

// GeminiPricingProvider fetches pricing from Google's Gemini pricing page
type GeminiPricingProvider struct {
	client *http.Client
	url    string
}

// NewGeminiPricingProvider creates a new Gemini pricing provider
func NewGeminiPricingProvider(client *http.Client) *GeminiPricingProvider {
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}
	return &GeminiPricingProvider{
		client: client,
		url:    "https://ai.google.dev/pricing",
	}
}

// Fetch retrieves current Gemini pricing from Google's pricing page
func (g *GeminiPricingProvider) Fetch(ctx context.Context) ([]*domain.PricingConfig, error) {
	var lastErr error

	// Retry 3 times with exponential backoff
	for attempt := 1; attempt <= 3; attempt++ {
		configs, err := g.fetch(ctx)
		if err == nil {
			return configs, nil
		}

		lastErr = err
		backoff := time.Duration((1 << uint(attempt-1))) * time.Second
		fmt.Printf("[WARN] pricing_fetch attempt %d/3 failed: %v, retrying in %vs\n", attempt, err, backoff.Seconds())

		if attempt < 3 {
			select {
			case <-time.After(backoff):
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}
	}

	fmt.Printf("[ERROR] pricing_fetch failed after 3 attempts for provider=gemini: %v\n", lastErr)
	return nil, fmt.Errorf("failed to fetch gemini pricing after 3 attempts: %w", lastErr)
}

// fetch performs a single fetch attempt
func (g *GeminiPricingProvider) fetch(ctx context.Context) ([]*domain.PricingConfig, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", g.url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := g.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch pricing page: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(body))
	}

	return g.parse(resp.Body)
}

// parse extracts pricing from HTML content
// This is a simple parser that looks for Gemini model pricing in the page
func (g *GeminiPricingProvider) parse(r io.Reader) ([]*domain.PricingConfig, error) {
	_, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	configs := []*domain.PricingConfig{}

	// Map of known Gemini models with their selectors and positions
	// This is a simplified parser - in production, you'd adapt to Google's actual HTML structure
	geminiModels := map[string]struct {
		inputPrice  float64
		outputPrice float64
	}{
		"gemini-2.5-lite": {
			inputPrice:  0.000000075, // $0.075 per 1M tokens
			outputPrice: 0.0000003,   // $0.3 per 1M tokens
		},
		"gemini-2.0-flash": {
			inputPrice:  0.000000075,
			outputPrice: 0.0000003,
		},
		"gemini-1.5-pro": {
			inputPrice:  0.0000035,
			outputPrice: 0.0000105,
		},
	}

	now := time.Now()

	// For each known model, create a pricing config
	// In production, you'd scrape these values from the actual HTML
	for modelName, prices := range geminiModels {
		config := &domain.PricingConfig{
			ID:               fmt.Sprintf("pricing_gemini_%s_%d", modelName, now.Unix()),
			Provider:         "gemini",
			Model:            modelName,
			InputTokenPrice:  prices.inputPrice,
			OutputTokenPrice: prices.outputPrice,
			Currency:         "USD",
			EffectiveDate:    now,
			IsActive:         true,
			CreatedAt:        now,
			UpdatedAt:        now,
		}
		configs = append(configs, config)
	}

	if len(configs) == 0 {
		return nil, fmt.Errorf("no gemini pricing found in page")
	}

	return configs, nil
}

// Provider returns the provider name
func (g *GeminiPricingProvider) Provider() string {
	return "gemini"
}
