package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/riverlin/aiexpense/internal/adapter/repository/sqlite"
	"github.com/riverlin/aiexpense/internal/ai"
	"github.com/riverlin/aiexpense/internal/config"
	"github.com/riverlin/aiexpense/internal/usecase"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Open database
	db, err := sqlite.OpenDB(cfg.DatabasePath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Initialize repositories
	userRepo := sqlite.NewUserRepository(db)
	categoryRepo := sqlite.NewCategoryRepository(db)
	expenseRepo := sqlite.NewExpenseRepository(db)
	metricsRepo := sqlite.NewMetricsRepository(db)

	// Initialize AI service
	aiService, err := ai.Factory(cfg.AIProvider, cfg.GeminiAPIKey)
	if err != nil {
		log.Fatalf("Failed to initialize AI service: %v", err)
	}

	// Initialize use cases
	autoSignupUseCase := usecase.NewAutoSignupUseCase(userRepo, categoryRepo)
	parseConversationUseCase := usecase.NewParseConversationUseCase(aiService)
	createExpenseUseCase := usecase.NewCreateExpenseUseCase(expenseRepo, categoryRepo, aiService)
	getExpensesUseCase := usecase.NewGetExpensesUseCase(expenseRepo, categoryRepo)

	// Initialize HTTP server
	mux := http.NewServeMux()

	// Health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"status":"ok"}`)
	})

	// TODO: Add HTTP handlers for:
	// - POST /api/users/auto-signup
	// - POST /api/expenses
	// - GET /api/expenses
	// - POST /api/expenses/parse
	// - GET /api/categories
	// - POST /api/categories
	// - GET /api/reports/summary
	// - GET /api/reports/breakdown
	// - GET /api/metrics/dau
	// - GET /api/metrics/expenses-summary
	// - GET /api/metrics/category-trends
	// - GET /api/metrics/growth

	// TODO: Add LINE webhook endpoint
	// POST /webhook/line

	// Log use case initialization for debugging
	_ = autoSignupUseCase
	_ = parseConversationUseCase
	_ = createExpenseUseCase
	_ = getExpensesUseCase
	_ = metricsRepo

	// Start server
	addr := ":" + cfg.ServerPort
	log.Printf("Starting server on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
