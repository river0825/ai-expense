package http

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/riverlin/aiexpense/internal/domain"
	"github.com/riverlin/aiexpense/internal/usecase"
)

// TestAPIGetPolicy tests the policy retrieval endpoint
func TestAPIGetPolicy(t *testing.T) {
	policyRepo := &TestPolicyRepository{policies: make(map[string]*domain.Policy)}

	// Seed test data
	policyRepo.policies["privacy_policy"] = &domain.Policy{
		ID:        "policy_1",
		Key:       "privacy_policy",
		Title:     "Privacy Policy",
		Content:   "This is privacy policy",
		Version:   "1.0",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	handler := NewHandler(
		nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
		usecase.NewGetPolicyUseCase(policyRepo),
		nil, nil, nil, nil, "",
	)

	t.Run("GetPrivacyPolicy", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/policies/privacy_policy", nil)
		req.SetPathValue("key", "privacy_policy") // Simulate path param for 1.22+

		w := httptest.NewRecorder()
		handler.GetPolicy(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var resp Response
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if resp.Status != "success" {
			t.Errorf("Expected status success, got %s", resp.Status)
		}

		data, ok := resp.Data.(map[string]interface{})
		if !ok {
			t.Fatalf("Data is not a map")
		}

		if data["title"] != "Privacy Policy" {
			t.Errorf("Expected title 'Privacy Policy', got '%v'", data["title"])
		}
	})

	t.Run("GetNonExistentPolicy", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/policies/non_existent", nil)
		req.SetPathValue("key", "non_existent")

		w := httptest.NewRecorder()
		handler.GetPolicy(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status 404, got %d", w.Code)
		}
	})

	t.Run("MissingKey", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/policies/", nil)
		// No key set

		w := httptest.NewRecorder()
		handler.GetPolicy(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", w.Code)
		}
	})
}
