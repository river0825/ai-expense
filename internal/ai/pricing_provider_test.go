package ai

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"testing"
)

// TestFetch_Success verifies successful fetch with valid HTML
func TestFetch_Success(t *testing.T) {
	mockHTML := `
	<html>
		<body>
			<table>
				<tr><td>gemini-2.5-lite</td><td>$0.075</td><td>$0.30</td></tr>
			</table>
		</body>
	</html>
	`

	client := &http.Client{
		Transport: &mockRoundTripper{
			response: &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(bytes.NewBufferString(mockHTML)),
			},
		},
	}

	provider := NewGeminiPricingProvider(client)
	configs, err := provider.Fetch(context.Background())

	if err != nil {
		t.Fatalf("Fetch failed: %v", err)
	}

	if len(configs) == 0 {
		t.Error("Expected at least one config, got 0")
	}

	if configs[0].Provider != "gemini" {
		t.Errorf("Expected provider=gemini, got %s", configs[0].Provider)
	}

	if !configs[0].IsActive {
		t.Error("Expected IsActive=true")
	}
}

// TestFetch_RetryOnNetworkError verifies retry logic works
func TestFetch_RetryOnNetworkError(t *testing.T) {
	mockHTML := `<html><body><table><tr><td>gemini-2.5-lite</td></tr></table></body></html>`

	attemptCount := 0
	client := &http.Client{
		Transport: &mockRoundTripper{
			handler: func() *http.Response {
				attemptCount++
				if attemptCount < 3 {
					return &http.Response{StatusCode: 500, Body: io.NopCloser(bytes.NewBufferString(""))}
				}
				return &http.Response{StatusCode: 200, Body: io.NopCloser(bytes.NewBufferString(mockHTML))}
			},
		},
	}

	provider := NewGeminiPricingProvider(client)
	configs, err := provider.Fetch(context.Background())

	if err != nil {
		t.Fatalf("Fetch failed: %v", err)
	}

	if attemptCount != 3 {
		t.Errorf("Expected 3 attempts, got %d", attemptCount)
	}

	if len(configs) == 0 {
		t.Error("Expected configs after retry success")
	}
}

// TestFetch_AllRetriesFail verifies error after all retries fail
func TestFetch_AllRetriesFail(t *testing.T) {
	client := &http.Client{
		Transport: &mockRoundTripper{
			response: &http.Response{StatusCode: 500, Body: io.NopCloser(bytes.NewBufferString(""))},
		},
	}

	provider := NewGeminiPricingProvider(client)
	configs, err := provider.Fetch(context.Background())

	if err == nil {
		t.Error("Expected error, got nil")
	}

	if configs != nil {
		t.Errorf("Expected nil configs on error, got %v", configs)
	}
}

// TestProvider verifies Provider() returns correct name
func TestProvider(t *testing.T) {
	provider := NewGeminiPricingProvider(nil)
	if provider.Provider() != "gemini" {
		t.Errorf("Expected 'gemini', got %s", provider.Provider())
	}
}

// Mock HTTP roundtripper for testing
type mockRoundTripper struct {
	response *http.Response
	handler  func() *http.Response
}

func (m *mockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if m.handler != nil {
		return m.handler(), nil
	}
	return m.response, nil
}
