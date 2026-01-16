package main

import (
	"log"
	"net/http"

	httpAdapter "github.com/riverlin/aiexpense/internal/adapter/http"
	"github.com/riverlin/aiexpense/internal/adapter/messenger/line"
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

	// Initialize HTTP handler
	handler := httpAdapter.NewHandler(
		autoSignupUseCase,
		parseConversationUseCase,
		createExpenseUseCase,
		getExpensesUseCase,
		userRepo,
		categoryRepo,
		expenseRepo,
		metricsRepo,
		cfg.AdminAPIKey,
	)

	// Initialize LINE client
	lineClient, err := line.NewClient(cfg.LineChannelToken)
	if err != nil {
		log.Fatalf("Failed to initialize LINE client: %v", err)
	}

	// Initialize LINE use case
	lineUseCase := line.NewLineUseCase(
		autoSignupUseCase,
		parseConversationUseCase,
		createExpenseUseCase,
		lineClient,
	)

	// Initialize LINE webhook handler
	lineHandler := line.NewHandler(cfg.LineChannelID, lineUseCase)

	// Initialize HTTP server
	mux := http.NewServeMux()
	httpAdapter.RegisterRoutes(mux, handler)

	// Add LINE webhook endpoint
	mux.HandleFunc("POST /webhook/line", lineHandler.HandleWebhook)

	// TODO: Add more use cases and handlers:
	// - UpdateExpenseUseCase
	// - DeleteExpenseUseCase
	// - ManageCategoryUseCase
	// - GenerateReportUseCase
	// - MetricsAggregatorUseCase

	// Start server
	addr := ":" + cfg.ServerPort
	log.Printf("Starting server on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
