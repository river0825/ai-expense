package config

import (
	"fmt"
	"os"
	"strings"
)

type Config struct {
	// Database
	DatabasePath string

	// LINE Bot
	LineChannelToken  string
	LineChannelID     string
	LineChannelSecret string

	// Telegram Bot
	TelegramBotToken string

	// Discord Bot
	DiscordBotToken string

	// WhatsApp Business API
	WhatsAppPhoneNumberID string
	WhatsAppAccessToken   string

	// Slack Bot
	SlackBotToken      string
	SlackSigningSecret string

	// Microsoft Teams Bot
	TeamsAppID       string
	TeamsAppPassword string

	// Gemini AI
	GeminiAPIKey string
	AIProvider   string // "gemini", "claude", "openai"

	// Server
	ServerPort string

	// Admin API Key for metrics
	AdminAPIKey string

	// Enabled Messengers
	EnabledMessengers []string
}

func Load() (*Config, error) {
	cfg := &Config{
		DatabasePath:          getEnv("DATABASE_PATH", "./aiexpense.db"),
		DatabaseURL:          getEnv("DATABASE_URL", ""),
		LineChannelToken:      getEnv("LINE_CHANNEL_TOKEN", ""),
		LineChannelID:         getEnv("LINE_CHANNEL_ID", ""),
		LineChannelSecret:     getEnv("LINE_CHANNEL_SECRET", ""),
		TelegramBotToken:      getEnv("TELEGRAM_BOT_TOKEN", ""),
		DiscordBotToken:       getEnv("DISCORD_BOT_TOKEN", ""),
		WhatsAppPhoneNumberID: getEnv("WHATSAPP_PHONE_NUMBER_ID", ""),
		WhatsAppAccessToken:   getEnv("WHATSAPP_ACCESS_TOKEN", ""),
		SlackBotToken:         getEnv("SLACK_BOT_TOKEN", ""),
		SlackSigningSecret:    getEnv("SLACK_SIGNING_SECRET", ""),
		TeamsAppID:            getEnv("TEAMS_APP_ID", ""),
		TeamsAppPassword:      getEnv("TEAMS_APP_PASSWORD", ""),
		GeminiAPIKey:          getEnv("GEMINI_API_KEY", ""),
		AIProvider:            getEnv("AI_PROVIDER", "gemini"),
		ServerPort:            getEnv("SERVER_PORT", "8080"),
		AdminAPIKey:           getEnv("ADMIN_API_KEY", ""),
	}

	// Parse enabled messengers
	enabledMessengersEnv := getEnv("ENABLED_MESSENGERS", "")
	if enabledMessengersEnv == "" {
		cfg.EnabledMessengers = []string{"terminal"}
	} else {
		cfg.EnabledMessengers = strings.Split(enabledMessengersEnv, ",")
		for i, m := range cfg.EnabledMessengers {
			cfg.EnabledMessengers[i] = strings.TrimSpace(m)
		}
	}

	// Validate required fields
	if cfg.IsMessengerEnabled("line") && cfg.LineChannelToken == "" {
		return nil, fmt.Errorf("LINE_CHANNEL_TOKEN is required when line messenger is enabled")
	}

	if cfg.GeminiAPIKey == "" && cfg.AIProvider == "gemini" {
		return nil, fmt.Errorf("GEMINI_API_KEY is required when using gemini AI provider")
	}

	// Validate database configuration - mutually exclusive for SQLite and PostgreSQL
	if cfg.DatabasePath == "" && cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("Either DATABASE_PATH or DATABASE_URL must be set")
	}

	if cfg.DatabasePath != "" && cfg.DatabaseURL != "" {
		return nil, fmt.Errorf("Only one of DATABASE_PATH or DATABASE_URL can be set, not both")
	}

	return cfg, nil
}

	// Parse enabled messengers
	enabledMessengersEnv := getEnv("ENABLED_MESSENGERS", "")
	if enabledMessengersEnv == "" {
		cfg.EnabledMessengers = []string{"terminal"}
	} else {
		cfg.EnabledMessengers = strings.Split(enabledMessengersEnv, ",")
		// Trim spaces
		for i, m := range cfg.EnabledMessengers {
			cfg.EnabledMessengers[i] = strings.TrimSpace(m)
		}
	}

	// Validate required fields
	if cfg.IsMessengerEnabled("line") && cfg.LineChannelToken == "" {
		return nil, fmt.Errorf("LINE_CHANNEL_TOKEN is required when line messenger is enabled")
	}
	if cfg.GeminiAPIKey == "" && cfg.AIProvider == "gemini" {
		return nil, fmt.Errorf("GEMINI_API_KEY is required when using gemini AI provider")
	}

	return cfg, nil
}

// IsMessengerEnabled checks if a specific messenger is enabled
func (c *Config) IsMessengerEnabled(name string) bool {
	for _, m := range c.EnabledMessengers {
		if m == name {
			return true
		}
	}
	return false
}

func getEnv(key, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultVal
}
