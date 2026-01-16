package config

import (
	"fmt"
	"os"
)

type Config struct {
	// Database
	DatabasePath string

	// LINE Bot
	LineChannelToken string
	LineChannelID    string

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
}

func Load() (*Config, error) {
	cfg := &Config{
		DatabasePath:          getEnv("DATABASE_PATH", "./aiexpense.db"),
		LineChannelToken:      getEnv("LINE_CHANNEL_TOKEN", ""),
		LineChannelID:         getEnv("LINE_CHANNEL_ID", ""),
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

	// Validate required fields
	if cfg.LineChannelToken == "" {
		return nil, fmt.Errorf("LINE_CHANNEL_TOKEN is required")
	}
	if cfg.GeminiAPIKey == "" && cfg.AIProvider == "gemini" {
		return nil, fmt.Errorf("GEMINI_API_KEY is required when using gemini AI provider")
	}

	return cfg, nil
}

func getEnv(key, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultVal
}
