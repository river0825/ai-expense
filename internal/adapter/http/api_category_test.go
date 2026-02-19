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

// setupCategoryTest initializes the test environment and returns the handler and repositories
func setupCategoryTest() (*Handler, *TestCategoryRepository, *TestUserRepository, *TestExpenseRepository) {
	userRepo := &TestUserRepository{users: make(map[string]*domain.User)}
	categoryRepo := &TestCategoryRepository{categories: make(map[string]*domain.Category)}
	expenseRepo := &TestExpenseRepository{expenses: make(map[string]*domain.Expense)}
	policyRepo := &TestPolicyRepository{policies: make(map[string]*domain.Policy)}

	handler := NewHandler(
		usecase.NewAutoSignupUseCase(userRepo, categoryRepo, nil),
		nil, nil, nil, nil, nil,
		usecase.NewManageCategoryUseCase(categoryRepo, expenseRepo),
		nil, nil, nil, nil, nil, nil, nil, nil,
		usecase.NewGetPolicyUseCase(policyRepo),
		nil, nil, nil, nil,
		userRepo, categoryRepo, expenseRepo, nil, "", nil, false,
	)

	return handler, categoryRepo, userRepo, expenseRepo
}

func Test_Category_Create(t *testing.T) {
	handler, categoryRepo, userRepo, _ := setupCategoryTest()
	userID := "test_user_create"

	// Setup User
	userRepo.Create(context.Background(), &domain.User{UserID: userID})

	t.Run("Success", func(t *testing.T) {
		body := map[string]interface{}{
			"user_id": userID,
			"name":    "New Category",
		}
		bodyBytes, _ := json.Marshal(body)
		req := httptest.NewRequest("POST", "/api/categories", bytes.NewReader(bodyBytes))
		req = authenticatedRequest(req, userID)
		w := httptest.NewRecorder()

		handler.CreateCategory(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected 200 OK, got %d", w.Code)
		}

		// Verify in repo
		cats, _ := categoryRepo.GetByUserID(context.Background(), userID)
		found := false
		for _, c := range cats {
			if c.Name == "New Category" {
				found = true
				break
			}
		}
		if !found {
			t.Error("Category not found in repository")
		}
	})

	t.Run("MissingName", func(t *testing.T) {
		body := map[string]interface{}{
			"user_id": userID,
			"name":    "",
		}
		bodyBytes, _ := json.Marshal(body)
		req := httptest.NewRequest("POST", "/api/categories", bytes.NewReader(bodyBytes))
		req = authenticatedRequest(req, userID)
		w := httptest.NewRecorder()

		handler.CreateCategory(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected 400 Bad Request, got %d", w.Code)
		}
	})
}

func Test_Category_List(t *testing.T) {
	handler, categoryRepo, userRepo, _ := setupCategoryTest()
	userID := "test_user_list"

	// Setup User and Categories
	userRepo.Create(context.Background(), &domain.User{UserID: userID})
	categoryRepo.Create(context.Background(), &domain.Category{ID: "cat1", UserID: userID, Name: "Food"})
	categoryRepo.Create(context.Background(), &domain.Category{ID: "cat2", UserID: userID, Name: "Transport"})

	req := httptest.NewRequest("GET", "/api/categories", nil)
	req = authenticatedRequest(req, userID)
	w := httptest.NewRecorder()

	handler.GetCategories(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200 OK, got %d", w.Code)
	}

	var resp Response
	json.NewDecoder(w.Body).Decode(&resp)

	// In the real handler, Data might be of type []*domain.Category, so we need to assert correctly
	// JSON unmarshaling to interface{} makes it a []interface{}
	// Let's re-marshal and unmarshal to specific type or just check count loosely
	
	// Better: Decode Data to []*domain.Category by re-marshalling
	dataBytes, _ := json.Marshal(resp.Data)
	var categories []*domain.Category
	json.Unmarshal(dataBytes, &categories)

	if len(categories) != 2 {
		t.Errorf("Expected 2 categories, got %d", len(categories))
	}
}

func Test_Category_Update(t *testing.T) {
	handler, categoryRepo, userRepo, _ := setupCategoryTest()
	userID := "test_user_update"
	catID := "cat_to_update"

	// Setup User and Category
	userRepo.Create(context.Background(), &domain.User{UserID: userID})
	categoryRepo.Create(context.Background(), &domain.Category{ID: catID, UserID: userID, Name: "Old Name"})

	body := map[string]interface{}{
		"id":   catID,
		"name": "New Name",
	}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequest("PUT", "/api/categories", bytes.NewReader(bodyBytes))
	req = authenticatedRequest(req, userID)
	w := httptest.NewRecorder()

	handler.UpdateCategory(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200 OK, got %d", w.Code)
	}

	// Verify update
	cat, _ := categoryRepo.GetByID(context.Background(), catID)
	if cat.Name != "New Name" {
		t.Errorf("Expected name 'New Name', got '%s'", cat.Name)
	}
}

func Test_Category_Delete(t *testing.T) {
	handler, categoryRepo, userRepo, _ := setupCategoryTest()
	userID := "test_user_delete"
	catID := "cat_to_delete"

	// Setup User and Category
	userRepo.Create(context.Background(), &domain.User{UserID: userID})
	categoryRepo.Create(context.Background(), &domain.Category{ID: catID, UserID: userID, Name: "To Delete"})

	// Verify existence before delete
	_, err := categoryRepo.GetByID(context.Background(), catID)
	if err != nil {
		t.Fatal("Setup failed: Category not found")
	}

	body := map[string]interface{}{
		"id": catID,
	}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequest("DELETE", "/api/categories", bytes.NewReader(bodyBytes))
	req = authenticatedRequest(req, userID)
	w := httptest.NewRecorder()

	handler.DeleteCategory(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200 OK, got %d", w.Code)
	}

	// Verify deletion
	_, err = categoryRepo.GetByID(context.Background(), catID)
	if err == nil {
		t.Error("Category should have been deleted")
	}
}
