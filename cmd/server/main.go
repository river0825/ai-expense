package main

import (
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
	aiCostRepo := sqlite.NewAICostRepository(db)
	policyRepo := sqlite.NewPolicyRepository(db)

	// Initialize AI service
	aiService, err := ai.Factory(cfg.AIProvider, cfg.GeminiAPIKey, aiCostRepo)
	if err != nil {
		log.Fatalf("Failed to initialize AI service: %v", err)
	}

	// Initialize use cases
	autoSignupUseCase := usecase.NewAutoSignupUseCase(userRepo, categoryRepo)
	parseConversationUseCase := usecase.NewParseConversationUseCase(aiService)
	createExpenseUseCase := usecase.NewCreateExpenseUseCase(expenseRepo, categoryRepo, aiService)
	getExpensesUseCase := usecase.NewGetExpensesUseCase(expenseRepo, categoryRepo)
	updateExpenseUseCase := usecase.NewUpdateExpenseUseCase(expenseRepo, categoryRepo)
	deleteExpenseUseCase := usecase.NewDeleteExpenseUseCase(expenseRepo)
	manageCategoryUseCase := usecase.NewManageCategoryUseCase(categoryRepo)
	generateReportUseCase := usecase.NewGenerateReportUseCase(expenseRepo, categoryRepo, metricsRepo)
	budgetManagementUseCase := usecase.NewBudgetManagementUseCase(categoryRepo, expenseRepo)
	dataExportUseCase := usecase.NewDataExportUseCase(expenseRepo, categoryRepo)
	metricsUseCase := usecase.NewMetricsUseCase(metricsRepo)
	recurringExpenseUseCase := usecase.NewRecurringExpenseUseCase(expenseRepo, categoryRepo)
	notificationUseCase := usecase.NewNotificationUseCase()
	searchExpenseUseCase := usecase.NewSearchExpenseUseCase(expenseRepo, categoryRepo)
	archiveUseCase := usecase.NewArchiveUseCase(expenseRepo)
	getPolicyUseCase := usecase.NewGetPolicyUseCase(policyRepo)

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

		// Initialize LINE use case
		lineUseCase := line.NewLineUseCase(
			autoSignupUseCase,
			parseConversationUseCase,
			createExpenseUseCase,
			lineClient,
		)

		// Initialize LINE webhook handler
		lineHandler = line.NewHandler(cfg.LineChannelID, lineUseCase)
	}

	// Initialize Terminal messenger (if enabled)
	var terminalHandler *terminal.Handler
	if cfg.IsMessengerEnabled("terminal") {
		terminalUseCase := terminal.NewTerminalUseCase(
			autoSignupUseCase,
			parseConversationUseCase,
			createExpenseUseCase,
			getExpensesUseCase,
			userRepo,
		)
		terminalHandler = terminal.NewHandler(terminalUseCase)
		log.Printf("Terminal messenger initialized")
	}

	// Initialize Telegram client (optional)
	var telegramHandler *telegram.Handler
	if cfg.IsMessengerEnabled("telegram") && cfg.TelegramBotToken != "" {
		telegramClient, err := telegram.NewClient(cfg.TelegramBotToken)
		if err != nil {
			log.Fatalf("Failed to initialize Telegram client: %v", err)
		}

		// Initialize Telegram use case
		telegramUseCase := telegram.NewTelegramUseCase(
			autoSignupUseCase,
			parseConversationUseCase,
			createExpenseUseCase,
			telegramClient,
		)

		// Initialize Telegram webhook handler
		telegramHandler = telegram.NewHandler(cfg.TelegramBotToken, telegramUseCase)
	}

	// Initialize Discord client (optional)
	var discordHandler *discord.Handler
	if cfg.IsMessengerEnabled("discord") && cfg.DiscordBotToken != "" {
		discordClient, err := discord.NewClient(cfg.DiscordBotToken)
		if err != nil {
			log.Fatalf("Failed to initialize Discord client: %v", err)
		}

		// Initialize Discord use case
		discordUseCase := discord.NewDiscordUseCase(
			autoSignupUseCase,
			parseConversationUseCase,
			createExpenseUseCase,
			discordClient,
		)

		// Initialize Discord webhook handler
		discordHandler = discord.NewHandler(cfg.DiscordBotToken, discordUseCase)
	}

	// Initialize WhatsApp client (optional)
	var whatsappHandler *whatsapp.Handler
	if cfg.IsMessengerEnabled("whatsapp") && cfg.WhatsAppPhoneNumberID != "" && cfg.WhatsAppAccessToken != "" {
		whatsappClient, err := whatsapp.NewClient(cfg.WhatsAppPhoneNumberID, cfg.WhatsAppAccessToken)
		if err != nil {
			log.Fatalf("Failed to initialize WhatsApp client: %v", err)
		}

		// Initialize WhatsApp use case
		whatsappUseCase := whatsapp.NewWhatsAppUseCase(
			autoSignupUseCase,
			parseConversationUseCase,
			createExpenseUseCase,
			whatsappClient,
		)

		// Initialize WhatsApp webhook handler with app secret
		appSecret := "" // In production, this would be the app secret from Meta
		whatsappHandler = whatsapp.NewHandler(appSecret, cfg.WhatsAppPhoneNumberID, whatsappUseCase)
	}

	// Initialize Slack client (optional)
	var slackHandler *slack.Handler
	if cfg.IsMessengerEnabled("slack") && cfg.SlackBotToken != "" {
		slackClient, err := slack.NewClient(cfg.SlackBotToken)
		if err != nil {
			log.Fatalf("Failed to initialize Slack client: %v", err)
		}

		// Initialize Slack use case
		slackUseCase := slack.NewSlackUseCase(
			autoSignupUseCase,
			parseConversationUseCase,
			createExpenseUseCase,
			slackClient,
		)

		// Initialize Slack webhook handler
		slackHandler = slack.NewHandler(cfg.SlackSigningSecret, slackUseCase)
	}

	// Initialize Microsoft Teams client (optional)
	var teamsHandler *teams.Handler
	if cfg.IsMessengerEnabled("teams") && cfg.TeamsAppID != "" && cfg.TeamsAppPassword != "" {
		teamsClient, err := teams.NewClient(cfg.TeamsAppID, cfg.TeamsAppPassword)
		if err != nil {
			log.Fatalf("Failed to initialize Teams client: %v", err)
		}

		// Initialize Teams use case
		teamsUseCase := teams.NewTeamsUseCase(
			autoSignupUseCase,
			parseConversationUseCase,
			createExpenseUseCase,
			teamsClient,
		)

		// Initialize Teams webhook handler
		teamsHandler = teams.NewHandler(cfg.TeamsAppID, cfg.TeamsAppPassword, teamsUseCase)
	}

	// Initialize HTTP server
	mux := http.NewServeMux()
	httpAdapter.RegisterRoutes(mux, handler)

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
