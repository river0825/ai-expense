package http

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/riverlin/aiexpense/internal/domain"
	"github.com/riverlin/aiexpense/internal/usecase"
)
// TestAPIConcurrentRequests tests concurrent request handling
func Test_WhenConcurrentRequests_GivenMultipleUsers_ShouldHandleAll(t *testing.T) {
	userRepo := &TestUserRepository{users: make(map[string]*domain.User)}
	categoryRepo := &TestCategoryRepository{categories: make(map[string]*domain.Category)}
	policyRepo := &TestPolicyRepository{policies: make(map[string]*domain.Policy)}

	handler := NewHandler(
		usecase.NewAutoSignupUseCase(userRepo, categoryRepo, nil),
		nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
		usecase.NewGetPolicyUseCase(policyRepo),
		nil, nil, nil, nil,
		userRepo, categoryRepo, nil, nil, "", nil, false,
	)

	// Simulate concurrent signup requests
	done := make(chan bool, 3)

	for i := 1; i <= 3; i++ {
		go func(index int) {
			userID := "concurrent_user_" + string(rune('0'+byte(index)))
			bodyMap := map[string]string{
				"user_id":        userID,
				"messenger_type": "line",
			}
			bodyBytes, _ := json.Marshal(bodyMap)

			req := httptest.NewRequest("POST", "/api/users/auto-signup", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			handler.AutoSignup(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("Concurrent signup %d: expected %d, got %d", index, http.StatusOK, w.Code)
			}

			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 3; i++ {
		<-done
	}
}
func Test_WhenRefreshRates_GivenAdminKey_ShouldRefresh(t *testing.T) {
	newHandler := func(svc domain.ExchangeRateService, adminKey string) *Handler {
		policyRepo := &TestPolicyRepository{policies: make(map[string]*domain.Policy)}
		return NewHandler(
			nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
			usecase.NewGetPolicyUseCase(policyRepo),
			nil, nil, nil,
			svc,
			nil, nil, nil, nil,
			adminKey,
			nil,
			false,
		)
	}

	t.Run("Success", func(t *testing.T) {
		svc := &TestExchangeRateService{}
		handler := newHandler(svc, "secret")
		req := httptest.NewRequest("POST", "/api/exchange-rates/refresh", nil)
		req.Header.Set("X-API-Key", "secret")
		w := httptest.NewRecorder()
		handler.RefreshExchangeRates(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected %d, got %d", http.StatusOK, w.Code)
		}
		if !svc.refreshCalled {
			t.Fatal("expected refresh to be called")
		}
	})



	t.Run("RefreshError", func(t *testing.T) {
		svc := &TestExchangeRateService{refreshErr: errors.New("boom")}
		handler := newHandler(svc, "secret")
		req := httptest.NewRequest("POST", "/api/exchange-rates/refresh", nil)
		req.Header.Set("X-API-Key", "secret")
		w := httptest.NewRecorder()
		handler.RefreshExchangeRates(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Fatalf("expected %d, got %d", http.StatusInternalServerError, w.Code)
		}
	})

	t.Run("ServiceUnavailable", func(t *testing.T) {
		handler := newHandler(nil, "secret")
		req := httptest.NewRequest("POST", "/api/exchange-rates/refresh", nil)
		req.Header.Set("X-API-Key", "secret")
		w := httptest.NewRecorder()
		handler.RefreshExchangeRates(w, req)

		if w.Code != http.StatusServiceUnavailable {
			t.Fatalf("expected %d, got %d", http.StatusServiceUnavailable, w.Code)
		}
	})
}
// TestRouteRegistration verifies the routing configuration, specifically auth protection
func Test_WhenRegisterRoutes_GivenMux_ShouldProtectAuthEndpoints(t *testing.T) {
	// Setup minimal handler
	userRepo := &TestUserRepository{users: make(map[string]*domain.User)}
	categoryRepo := &TestCategoryRepository{categories: make(map[string]*domain.Category)}
	
	handler := NewHandler(
		usecase.NewAutoSignupUseCase(userRepo, categoryRepo, nil),
		nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
		nil, nil, nil, nil,
		nil, // ExchangeRateService (passed as nil, but GetCurrencies handles it self-contained)
		userRepo, categoryRepo, nil, nil, "", nil, false,
	)

	mux := http.NewServeMux()
	RegisterRoutes(mux, handler, nil, nil, nil, nil, nil, nil)
	
	// Test Server
	ts := httptest.NewServer(mux)
	defer ts.Close()

	client := ts.Client()

	t.Run("GenerateReport_RequiresAuth", func(t *testing.T) {
		req, _ := http.NewRequest("POST", ts.URL+"/api/reports/generate", nil)
		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("Expected 401 Unauthorized for GenerateReport without token, got %v", resp.StatusCode)
		}
	})

	t.Run("GetCurrencies_Public", func(t *testing.T) {
		req, _ := http.NewRequest("GET", ts.URL+"/api/currencies", nil)
		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected 200 OK for GetCurrencies without token, got %v", resp.StatusCode)
		}
	})
}
