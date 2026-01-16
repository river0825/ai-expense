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
		DatabasePath:     getEnv("DATABASE_PATH", "./aiexpense.db"),
		LineChannelToken: getEnv("LINE_CHANNEL_TOKEN", ""),
		LineChannelID:    getEnv("LINE_CHANNEL_ID", ""),
		GeminiAPIKey:     getEnv("GEMINI_API_KEY", ""),
		AIProvider:       getEnv("AI_PROVIDER", "gemini"),
		ServerPort:       getEnv("SERVER_PORT", "8080"),
		AdminAPIKey:      getEnv("ADMIN_API_KEY", ""),
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
