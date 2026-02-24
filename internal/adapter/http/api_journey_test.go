package http

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
	"time"

	"github.com/riverlin/aiexpense/internal/adapter/messenger/terminal"
	"github.com/riverlin/aiexpense/internal/domain"
	"github.com/riverlin/aiexpense/internal/pkg/jwtutil"
	"github.com/riverlin/aiexpense/internal/usecase"
)

// Test_UserJourney_RecordExpenseAndEdit performs an end-to-end integration test
// Scenario:
// 1. User sends a message via "Terminal Chat" to record an expense.
// 2. User sends a message to get a report/link.
// 3. System returns a short link.
// 4. User clicks the link and is redirected (with a token).
// 5. User accesses the dashboard using the token.
// 6. User edits the expense on the dashboard.
func Test_UserJourney_RecordExpenseAndEdit(t *testing.T) {
	// Standard Test Wiring
	userRepo := &TestUserRepository{users: make(map[string]*domain.User)}
	categoryRepo := &TestCategoryRepository{categories: make(map[string]*domain.Category)}
	expenseRepo := &TestExpenseRepository{expenses: make(map[string]*domain.Expense)}
	policyRepo := &TestPolicyRepository{policies: make(map[string]*domain.Policy)}
	shortLinkRepo := &TestShortLinkRepository{links: make(map[string]*domain.ShortLink)}
	interactionRepo := &TestInteractionLogRepository{logs: []*domain.InteractionLog{}}
	aiCostRepo := &TestAICostRepository{costs: make(map[string]*domain.AICostLog)}
	pricingRepo := &TestPricingRepository{pricing: make(map[string]*domain.PricingConfig)}
	aiService := &TestAIService{}
	exchangeRateSvc := &TestExchangeRateService{}
	tokenManager := jwtutil.NewTokenManager("test-secret")

	// Create User
	userID := "journey_user_1"
	userRepo.Create(context.Background(), &domain.User{
		UserID:        userID,
		MessengerType: "terminal",
		CreatedAt:     time.Now(),
		HomeCurrency:  "TWD", // Set home currency for conversion check
	})

	// Setup UseCases
	autoSignupUC := usecase.NewAutoSignupUseCase(userRepo, categoryRepo, nil)
	parseConversationUC := usecase.NewParseConversationUseCase(aiService, pricingRepo, aiCostRepo, userRepo, categoryRepo, "mock", "mock-model")
	createExpenseUC := usecase.NewCreateExpenseUseCaseWithAIConfig(expenseRepo, categoryRepo, userRepo, exchangeRateSvc, aiCostRepo, pricingRepo, aiService, "mock", "mock-model")
	getExpensesUC := usecase.NewGetExpensesUseCase(expenseRepo, categoryRepo)
	generateReportLinkUC := usecase.NewGenerateReportLinkUseCase("http://localhost:3000", shortLinkRepo, tokenManager)
	
	processMessageUC := usecase.NewProcessMessageUseCase(
		autoSignupUC,
		parseConversationUC,
		createExpenseUC,
		getExpensesUC,
		nil,
		generateReportLinkUC,
		interactionRepo,
		expenseRepo,
		slog.Default(),
		nil,
	)

	// Setup Handlers
	terminalHandler := terminal.NewHandler(processMessageUC)
	shortLinkHandler := NewShortLinkHandler(shortLinkRepo, "http://localhost:3000")
	apiHandler := NewHandler(
		autoSignupUC,
		parseConversationUC,
		createExpenseUC,
		getExpensesUC,
		usecase.NewUpdateExpenseUseCase(expenseRepo, categoryRepo, exchangeRateSvc),
		nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
		usecase.NewGetPolicyUseCase(policyRepo),
		nil, nil, nil,
		exchangeRateSvc,
		userRepo, categoryRepo, expenseRepo, nil, "", tokenManager, false,
	)

	// Register Routes
	mux := http.NewServeMux()
	mux.HandleFunc("/api/chat/terminal", terminalHandler.HandleMessage)
	RegisterRoutes(mux, apiHandler, nil, nil, nil, shortLinkHandler, nil, nil)

	// Start Test Server
	ts := httptest.NewServer(mux)
	defer ts.Close()
	client := ts.Client()

	// --- STEP 1: Record Expense via Chat ---
	t.Log("Step 1: Recording expense via Chat")
	chatReq1 := map[string]interface{}{
		"user_id": userID,
		"message": "Lunch ",
	}
	chatBody1, _ := json.Marshal(chatReq1)
	resp1, err := client.Post(ts.URL+"/api/chat/terminal", "application/json", bytes.NewReader(chatBody1))
	if err != nil {
		t.Fatalf("Failed chat request: %v", err)
	}
	defer resp1.Body.Close()

	if resp1.StatusCode != http.StatusOK {
		t.Fatalf("Expected 200 OK from chat, got %d", resp1.StatusCode)
	}
	
	// Verify expense recorded in repo
	expenses, _ := expenseRepo.GetByUserID(context.Background(), userID)
	if len(expenses) != 1 {
		t.Fatalf("Expected 1 expense, got %d", len(expenses))
	}
	expenseID := expenses[0].ID

	// --- STEP 2: Get Report Link via Chat ---
	t.Log("Step 2: Requesting Report Link via Chat")
	chatReq2 := map[string]interface{}{
		"user_id": userID,
		"message": "Show report",
	}
	chatBody2, _ := json.Marshal(chatReq2)
	resp2, err := client.Post(ts.URL+"/api/chat/terminal", "application/json", bytes.NewReader(chatBody2))
	if err != nil {
		t.Fatalf("Failed chat report request: %v", err)
	}
	defer resp2.Body.Close()
	
	// Parse response to find link
	var termResp terminal.TerminalResponse
	json.NewDecoder(resp2.Body).Decode(&termResp)
	
	dataMap, ok := termResp.Data.(map[string]interface{})
	if !ok {
		t.Fatalf("Unexpected data format in terminal response: %+v", termResp.Data)
	}
	
	resultMap, ok := dataMap["result"].(map[string]interface{})
	if !ok {
		// Try checking if "link" is directly in Datamap if structure differs
		link, okDirect := dataMap["link"].(string)
		if okDirect {
			t.Logf("Received Link (direct): %s", link)
		} else {
             // Debug print map content
			 t.Fatalf("Result map missing in response data: %+v", dataMap)
		}
	} else {
        link, ok := resultMap["link"].(string)
        if !ok || link == "" {
            t.Fatalf("Link not found in result: %+v", resultMap)
        }
        t.Logf("Received Link: %s", link)
    }
    
    // Extract link again cleanly for next step (assuming happy path above printed/logged it but let's be robust)
    var link string
    if resultMap != nil {
        l, _ := resultMap["link"].(string)
        link = l
    } else {
         l, _ := dataMap["link"].(string)
         link = l
    }

	// --- STEP 3: Follow Link (Redirect) ---
	t.Log("Step 3: Following Short Link")
	re := regexp.MustCompile(`/r/([a-zA-Z0-9]+)`)
	matches := re.FindStringSubmatch(link)
	if len(matches) < 2 {
		t.Fatalf("Could not extract ID from link: %s", link)
	}
	shortID := matches[1]
	
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}
	
	redirectResp, err := client.Get(ts.URL + "/r/" + shortID)
	if err != nil {
		t.Fatalf("Failed to call redirect link: %v", err)
	}
	defer redirectResp.Body.Close()
	
	if redirectResp.StatusCode != http.StatusFound {
		t.Fatalf("Expected 302 Found, got %d", redirectResp.StatusCode)
	}
	
	cookies := redirectResp.Cookies()
	var authCookie *http.Cookie
	for _, c := range cookies {
		if c.Name == "report_token" {
			authCookie = c
			break
		}
	}
	
	location, _ := redirectResp.Location()
	token := location.Query().Get("token")
	
	if authCookie == nil && token == "" {
		t.Fatalf("No auth token found in cookie or redirect URL")
	}
	
	if token == "" && authCookie != nil {
		token = authCookie.Value
	}

	// --- STEP 4: Access Dashboard (Get Expenses) ---
	t.Log("Step 4: Accessing Dashboard API with Token")
	
	dashboardReq, _ := http.NewRequest("GET", ts.URL+"/api/expenses?user_id="+userID, nil)
	dashboardReq.Header.Set("Authorization", "Bearer "+token)
	
	client.CheckRedirect = nil
	dashboardResp, err := client.Do(dashboardReq)
	if err != nil {
		t.Fatalf("Failed to access dashboard: %v", err)
	}
	defer dashboardResp.Body.Close()
	
	if dashboardResp.StatusCode != http.StatusOK {
		t.Errorf("Expected 200 OK accessing dashboard, got %d", dashboardResp.StatusCode)
	}
	
	// Use json.NewDecoder to allow standard behavior
	// Note: GetAllResponse inside Data has "Expenses" field (Capitalized) because it has no tags. 
	// However, json decoder is case-insensitive for matching.
	// But let's use map to be safe if we are unsure about GetAllResponse serialization.
	// Actually, let's just decode to map[string]interface{} to debug or verify count.
	
	var rawResp map[string]interface{}
	json.NewDecoder(dashboardResp.Body).Decode(&rawResp)
	
	data, ok := rawResp["data"].(map[string]interface{})
	if !ok {
		t.Fatalf("Dashboard response missing 'data': %+v", rawResp)
	}
	
	expensesRaw, ok := data["Expenses"].([]interface{})
	if !ok {
		// Try lowercase
		expensesRaw, ok = data["expenses"].([]interface{})
	}
	
	if len(expensesRaw) != 1 {
		t.Errorf("Dashboard should show 1 expense, got %d. Data: %+v", len(expensesRaw), data)
	}


	// --- STEP 5: Edit Expense ---
	t.Log("Step 5: Editing Expense")
	
	editBody := map[string]interface{}{
		"id":          expenseID,
		"description": "Premium Lunch",
		"amount":      100.0,
	}
	editBytes, _ := json.Marshal(editBody)
	
	// Correct URL: Update via PUT /api/expenses (ID in body)
	editReq, _ := http.NewRequest("PUT", ts.URL+"/api/expenses", bytes.NewReader(editBytes))
	editReq.Header.Set("Authorization", "Bearer "+token)
	editReq.Header.Set("Content-Type", "application/json")
	
	editResp, err := client.Do(editReq)
	if err != nil {
		t.Fatalf("Failed to edit expense: %v", err)
	}
	defer editResp.Body.Close()
	
	if editResp.StatusCode != http.StatusOK {
		t.Errorf("Expected 200 OK for edit, got %d", editResp.StatusCode)
	}
	
	updatedExp, _ := expenseRepo.GetByID(context.Background(), expenseID)
	if updatedExp.Description != "Premium Lunch" {
		t.Errorf("Expense description not updated. Got: %s", updatedExp.Description)
	}
	if updatedExp.Amount != 100.0 {
		t.Errorf("Expense amount not updated. Got: %f", updatedExp.Amount)
	}
}

func Test_UserJourney_ErrorPaths(t *testing.T) {
	shortLinkRepo := &TestShortLinkRepository{links: make(map[string]*domain.ShortLink)}
	shortLinkHandler := NewShortLinkHandler(shortLinkRepo, "http://localhost:3000")
	
	mux := http.NewServeMux()
	mux.HandleFunc("/r/{id}", shortLinkHandler.HandleRedirect)
	
	ts := httptest.NewServer(mux)
	defer ts.Close()
	client := ts.Client()

	t.Run("Invalid Short Link", func(t *testing.T) {
		resp, err := client.Get(ts.URL + "/r/invalid-id")
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("Expected 404 for invalid link, got %d", resp.StatusCode)
		}
	})
}
