package main

import (
	"fmt"
	"log"
	"net/http"

	httpAdapter "github.com/riverlin/aiexpense/internal/adapter/http"
	"github.com/riverlin/aiexpense/internal/adapter/messenger/discord"
	"github.com/riverlin/aiexpense/internal/adapter/messenger/line"
	"github.com/riverlin/aiexpense/internal/adapter/messenger/slack"
	"github.com/riverlin/aiexpense/internal/adapter/messenger/teams"
	"github.com/riverlin/aiexpense/internal/adapter/messenger/telegram"
	"github.com/riverlin/aiexpense/internal/adapter/messenger/terminal"
	"github.com/riverlin/aiexpense/internal/adapter/messenger/whatsapp"
	postgresRepo "github.com/riverlin/aiexpense/internal/adapter/repository/postgresql"
	sqliteRepo "github.com/riverlin/aiexpense/internal/adapter/repository/sqlite"
	"github.com/riverlin/aiexpense/internal/ai"
	"github.com/riverlin/aiexpense/internal/config"
	"github.com/riverlin/aiexpense/internal/domain"
	"github.com/riverlin/aiexpense/internal/usecase"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Open database based on configuration
	var userRepo domain.UserRepository
	var categoryRepo domain.CategoryRepository
	var expenseRepo domain.ExpenseRepository
	var metricsRepo domain.MetricsRepository
	var aiCostRepo domain.AICostRepository
	var policyRepo domain.PolicyRepository
	var dbCloser interface{ Close() error }

	var pricingRepo domain.PricingRepository

	if cfg.DatabaseURL != "" {
		// Use PostgreSQL
		log.Printf("Connecting to PostgreSQL: %s", cfg.DatabaseURL)
		db, err := postgresRepo.OpenDB(cfg.DatabaseURL)
		if err != nil {
			log.Fatalf("Failed to open PostgreSQL database: %v", err)
		}
		dbCloser = db

		userRepo = postgresRepo.NewUserRepository(db)
		categoryRepo = postgresRepo.NewCategoryRepository(db)
		expenseRepo = postgresRepo.NewExpenseRepository(db)
		metricsRepo = postgresRepo.NewMetricsRepository(db)
		aiCostRepo = postgresRepo.NewAICostRepository(db)
		policyRepo = postgresRepo.NewPolicyRepository(db)
		pricingRepo = postgresRepo.NewPricingRepository(db)
		log.Printf("Connected to PostgreSQL database")
	} else {
		// Use SQLite
		log.Printf("Opening SQLite database: %s", cfg.DatabasePath)
		db, err := sqliteRepo.OpenDB(cfg.DatabasePath)
		if err != nil {
			log.Fatalf("Failed to open SQLite database: %v", err)
		}
		dbCloser = db

		userRepo = sqliteRepo.NewUserRepository(db)
		categoryRepo = sqliteRepo.NewCategoryRepository(db)
		expenseRepo = sqliteRepo.NewExpenseRepository(db)
		metricsRepo = sqliteRepo.NewMetricsRepository(db)
		aiCostRepo = sqliteRepo.NewAICostRepository(db)
		policyRepo = sqliteRepo.NewPolicyRepository(db)
		pricingRepo = sqliteRepo.NewPricingRepository(db)
		log.Printf("Connected to SQLite database")
	}

	// Ensure database is closed on exit
	defer func() {
		if dbCloser != nil {
			dbCloser.Close()
		}
	}()

	// Initialize AI service
	aiService, err := ai.Factory(cfg.AIProvider, cfg.GeminiAPIKey, aiCostRepo)
	if err != nil {
		log.Fatalf("Failed to initialize AI service: %v", err)
	}

	// Initialize use cases
	autoSignupUseCase := usecase.NewAutoSignupUseCase(userRepo, categoryRepo)
	parseConversationUseCase := usecase.NewParseConversationUseCase(
		aiService,
		pricingRepo,
		aiCostRepo,
		cfg.AIProvider,
		cfg.AIModel,
	)
	createExpenseUseCase := usecase.NewCreateExpenseUseCase(expenseRepo, categoryRepo, aiService)
	getExpensesUseCase := usecase.NewGetExpensesUseCase(expenseRepo, categoryRepo)
	updateExpenseUseCase := usecase.NewUpdateExpenseUseCase(expenseRepo, categoryRepo)
	deleteExpenseUseCase := usecase.NewDeleteExpenseUseCase(expenseRepo)
	manageCategoryUseCase := usecase.NewManageCategoryUseCase(categoryRepo)
	generateReportUseCase := usecase.NewGenerateReportUseCase(expenseRepo, categoryRepo, metricsRepo)
	budgetManagementUseCase := usecase.NewBudgetManagementUseCase(categoryRepo, expenseRepo)
	dataExportUseCase := usecase.NewDataExportUseCase(expenseRepo, categoryRepo)
	metricsUseCase := usecase.NewMetricsUseCase(metricsRepo)
	aiCostUseCase := usecase.NewAICostUseCase(aiCostRepo)
	recurringExpenseUseCase := usecase.NewRecurringExpenseUseCase(expenseRepo, categoryRepo)
	notificationUseCase := usecase.NewNotificationUseCase()
	searchExpenseUseCase := usecase.NewSearchExpenseUseCase(expenseRepo, categoryRepo)
	archiveUseCase := usecase.NewArchiveUseCase(expenseRepo)
	getPolicyUseCase := usecase.NewGetPolicyUseCase(policyRepo)

	// Initialize Unified Message Processor
	processMessageUseCase := usecase.NewProcessMessageUseCase(
		autoSignupUseCase,
		parseConversationUseCase,
		createExpenseUseCase,
		getExpensesUseCase,
	)

	// Initialize HTTP handler
	handler := httpAdapter.NewHandler(
		autoSignupUseCase,
		parseConversationUseCase,
		createExpenseUseCase,
		getExpensesUseCase,
		updateExpenseUseCase,
		deleteExpenseUseCase,
		manageCategoryUseCase,
		generateReportUseCase,
		budgetManagementUseCase,
		dataExportUseCase,
		recurringExpenseUseCase,
		notificationUseCase,
		searchExpenseUseCase,
		archiveUseCase,
		metricsUseCase,
		getPolicyUseCase,
		userRepo,
		categoryRepo,
		expenseRepo,
		metricsRepo,
		cfg.AdminAPIKey,
	)

	// Initialize LINE client (if enabled)
	var lineHandler *line.Handler
	if cfg.IsMessengerEnabled("line") {
		lineClient, err := line.NewClient(cfg.LineChannelToken)
		if err != nil {
			log.Fatalf("Failed to initialize LINE client: %v", err)
		}

		// Initialize LINE webhook handler with Unified Message Processor
		lineHandler = line.NewHandler(cfg.LineChannelSecret, processMessageUseCase, lineClient)
	}

	// Initialize Terminal messenger (if enabled)
	var terminalHandler *terminal.Handler
	if cfg.IsMessengerEnabled("terminal") {
		terminalHandler = terminal.NewHandler(processMessageUseCase)
		log.Printf("Terminal messenger initialized")
	}

	// Initialize Telegram client (optional)
	var telegramHandler *telegram.Handler
	if cfg.IsMessengerEnabled("telegram") && cfg.TelegramBotToken != "" {
		telegramClient, err := telegram.NewClient(cfg.TelegramBotToken)
		if err != nil {
			log.Fatalf("Failed to initialize Telegram client: %v", err)
		}

		// Initialize Telegram webhook handler
		telegramHandler = telegram.NewHandler(cfg.TelegramBotToken, processMessageUseCase, telegramClient)
	}

	// Initialize Discord client (optional)
	var discordHandler *discord.Handler
	if cfg.IsMessengerEnabled("discord") && cfg.DiscordBotToken != "" {
		discordClient, err := discord.NewClient(cfg.DiscordBotToken)
		if err != nil {
			log.Fatalf("Failed to initialize Discord client: %v", err)
		}

		// Initialize Discord webhook handler
		discordHandler = discord.NewHandler(cfg.DiscordBotToken, processMessageUseCase, discordClient)
	}

	// Initialize WhatsApp client (optional)
	var whatsappHandler *whatsapp.Handler
	if cfg.IsMessengerEnabled("whatsapp") && cfg.WhatsAppPhoneNumberID != "" && cfg.WhatsAppAccessToken != "" {
		// Client initialization logic removed as it's not used by handler yet
		// To re-enable client usage, update whatsapp.NewHandler to accept *Client

		// Initialize WhatsApp webhook handler with app secret
		appSecret := "" // In production, this would be the app secret from Meta
		// TODO: Get AppSecret from config
		whatsappHandler = whatsapp.NewHandler(appSecret, cfg.WhatsAppPhoneNumberID, processMessageUseCase)
	}

	// Initialize Slack client (optional)
	var slackHandler *slack.Handler
	if cfg.IsMessengerEnabled("slack") && cfg.SlackBotToken != "" {
		slackClient, err := slack.NewClient(cfg.SlackBotToken)
		if err != nil {
			log.Fatalf("Failed to initialize Slack client: %v", err)
		}

		// Initialize Slack webhook handler
		slackHandler = slack.NewHandler(cfg.SlackSigningSecret, processMessageUseCase, slackClient)
	}

	// Initialize Microsoft Teams client (optional)
	var teamsHandler *teams.Handler
	if cfg.IsMessengerEnabled("teams") && cfg.TeamsAppID != "" && cfg.TeamsAppPassword != "" {
		teamsClient, err := teams.NewClient(cfg.TeamsAppID, cfg.TeamsAppPassword)
		if err != nil {
			log.Fatalf("Failed to initialize Teams client: %v", err)
		}

		// Initialize Teams webhook handler
		teamsHandler = teams.NewHandler(cfg.TeamsAppID, cfg.TeamsAppPassword, processMessageUseCase, teamsClient)
	}

	// Initialize AI Cost handler
	aiCostHandler := httpAdapter.NewAICostHandler(aiCostUseCase, cfg.AdminAPIKey)

	// Initialize HTTP server
	mux := http.NewServeMux()
	httpAdapter.RegisterRoutes(mux, handler)
	httpAdapter.RegisterAICostRoutes(mux, aiCostHandler)

	// Add LINE webhook endpoint
	if lineHandler != nil {
		mux.HandleFunc("/webhook/line", lineHandler.HandleWebhook)
		log.Printf("LINE webhook enabled at /webhook/line")
	}

	// Add Terminal messenger endpoints
	if terminalHandler != nil {
		mux.HandleFunc("/api/chat/terminal", terminalHandler.HandleMessage)
		mux.HandleFunc("/api/chat/terminal/user", terminalHandler.GetUserInfo)
		log.Printf("Terminal messenger enabled at /api/chat/terminal")
	}

	// Add Telegram webhook endpoint (if configured)
	if telegramHandler != nil {
		mux.HandleFunc("/webhook/telegram", telegramHandler.HandleWebhook)
		log.Printf("Telegram webhook enabled at /webhook/telegram")
	}

	// Add Discord webhook endpoint (if configured)
	if discordHandler != nil {
		mux.HandleFunc("/webhook/discord", discordHandler.HandleWebhook)
		log.Printf("Discord webhook enabled at /webhook/discord")
	}

	// Add WhatsApp webhook endpoint (if configured)
	if whatsappHandler != nil {
		// WhatsApp uses GET for verification and POST for events
		mux.HandleFunc("/webhook/whatsapp", whatsappHandler.HandleWebhook)
		log.Printf("WhatsApp webhook enabled at /webhook/whatsapp")
	}

	// Add Slack webhook endpoint (if configured)
	if slackHandler != nil {
		mux.HandleFunc("/webhook/slack", slackHandler.HandleWebhook)
		log.Printf("Slack webhook enabled at /webhook/slack")
	}

	// Add Microsoft Teams webhook endpoint (if configured)
	if teamsHandler != nil {
		mux.HandleFunc("/webhook/teams", teamsHandler.HandleWebhook)
		log.Printf("Microsoft Teams webhook enabled at /webhook/teams")
	}

	// TODO: Add more use cases and handlers:
	// - UpdateExpenseUseCase
	// - DeleteExpenseUseCase
	// - ManageCategoryUseCase
	// - GenerateReportUseCase
	// - MetricsAggregatorUseCase

	// Wrap mux with CORS middleware for dashboard
	corsHandler := withCORS(mux)

	// Wrap with logging middleware
	loggingHandler := httpAdapter.LoggingMiddleware(corsHandler)

	// Start server
	addr := ":" + cfg.ServerPort
	log.Printf("Starting server on %s", addr)
	fmt.Printf("SERVER STARTED ON %s\n", addr)
	if err := http.ListenAndServe(addr, loggingHandler); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

// withCORS wraps HTTP handler with CORS headers
func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-API-Key")
		w.Header().Set("Access-Control-Max-Age", "3600")

		// Handle OPTIONS requests
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
