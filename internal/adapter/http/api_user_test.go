package http

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/riverlin/aiexpense/internal/domain"
	"github.com/riverlin/aiexpense/internal/usecase"
)
// TestAPIAutoSignupFlow tests complete auto-signup flow
func Test_WhenAutoSignupFlow_GivenValidRequest_ShouldCreateUser(t *testing.T) {
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

	// Create request body
	bodyMap := map[string]interface{}{
		"user_id":        "test_api_user",
		"messenger_type": "line",
	}
	bodyBytes, _ := json.Marshal(bodyMap)

	req := httptest.NewRequest("POST", "/api/users/auto-signup", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	handler.AutoSignup(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected %d, got %d", http.StatusOK, w.Code)
	}

	// Verify user was created
	exists, _ := userRepo.Exists(context.Background(), "test_api_user")
	if !exists {
		t.Error("Expected user to be created")
	}
}

// TestAPIAutoSignup tests user auto-signup flow
func Test_WhenAutoSignup_GivenValidRequest_ShouldCreateUserAndCategories(t *testing.T) {
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

	bodyMap := map[string]string{
		"user_id":        "test_user_1",
		"messenger_type": "line",
	}
	bodyBytes, _ := json.Marshal(bodyMap)

	req := httptest.NewRequest("POST", "/api/users/auto-signup", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	handler.AutoSignup(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected %d, got %d", http.StatusOK, w.Code)
	}

	// Verify user was created
	exists, _ := userRepo.Exists(context.Background(), "test_user_1")
	if !exists {
		t.Error("Expected user to be created")
	}

	// Verify default categories created
	categories, _ := categoryRepo.GetByUserID(context.Background(), "test_user_1")
	if len(categories) < 1 {
		t.Error("Expected default categories to be created")
	}
}
// TestAPIMissingRequired tests error handling for missing required fields
func Test_WhenAutoSignup_GivenMissingFields_ShouldReturnError(t *testing.T) {
	policyRepo := &TestPolicyRepository{policies: make(map[string]*domain.Policy)}

	handler := NewHandler(
		usecase.NewAutoSignupUseCase(
			&TestUserRepository{users: make(map[string]*domain.User)},
			&TestCategoryRepository{categories: make(map[string]*domain.Category)},
			nil,
		),
		nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
		usecase.NewGetPolicyUseCase(policyRepo),
		nil, nil, nil, nil,
		nil, nil, nil, nil, "", nil, false,
	)

	// Missing user_id
	bodyMap := map[string]string{
		"messenger_type": "line",
	}
	bodyBytes, _ := json.Marshal(bodyMap)

	req := httptest.NewRequest("POST", "/api/users/auto-signup", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	handler.AutoSignup(w, req)

	// Should fail due to missing user_id - expect 4xx status
	if w.Code >= 200 && w.Code < 300 {
		t.Errorf("Expected error status (4xx) for missing required field, got %d", w.Code)
	}
}
