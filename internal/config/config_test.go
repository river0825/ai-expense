package config

import (
	"os"
	"testing"
)

func TestLoad_EnabledMessengers(t *testing.T) {
	// Clear env vars before test
	os.Unsetenv("ENABLED_MESSENGERS")
	os.Unsetenv("LINE_CHANNEL_TOKEN")
	os.Unsetenv("LINE_CHANNEL_ID")
	os.Unsetenv("GEMINI_API_KEY")

	// Set minimal required fields for other parts (Gemini is required by default if provider is gemini)
	// But provider defaults to gemini. Let's set provider to something else or provide key to avoid that error masking our test.
	os.Setenv("GEMINI_API_KEY", "dummy_key")

	t.Run("Default to terminal", func(t *testing.T) {
		os.Unsetenv("ENABLED_MESSENGERS")

		cfg, err := Load()
		if err != nil {
			t.Fatalf("Load() failed: %v", err)
		}

		if len(cfg.EnabledMessengers) != 1 || cfg.EnabledMessengers[0] != "terminal" {
			t.Errorf("Expected default enabled messengers to be ['terminal'], got %v", cfg.EnabledMessengers)
		}

		if !cfg.IsMessengerEnabled("terminal") {
			t.Error("Expected terminal to be enabled")
		}
		if cfg.IsMessengerEnabled("line") {
			t.Error("Expected line to be disabled by default")
		}
	})

	t.Run("Parse enabled messengers from env", func(t *testing.T) {
		os.Setenv("ENABLED_MESSENGERS", "line,telegram")
		// We need line token now because line is enabled
		os.Setenv("LINE_CHANNEL_TOKEN", "dummy_token")
		defer os.Unsetenv("LINE_CHANNEL_TOKEN")

		cfg, err := Load()
		if err != nil {
			t.Fatalf("Load() failed: %v", err)
		}

		if !cfg.IsMessengerEnabled("line") {
			t.Error("Expected line to be enabled")
		}
		if !cfg.IsMessengerEnabled("telegram") {
			t.Error("Expected telegram to be enabled")
		}
		if cfg.IsMessengerEnabled("terminal") {
			t.Error("Expected terminal to be disabled")
		}
	})

	t.Run("Line token not required if line disabled", func(t *testing.T) {
		os.Setenv("ENABLED_MESSENGERS", "terminal")
		os.Unsetenv("LINE_CHANNEL_TOKEN")

		_, err := Load()
		if err != nil {
			t.Errorf("Expected Load() to succeed without line token when line disabled, got error: %v", err)
		}
	})

	t.Run("Line token required if line enabled", func(t *testing.T) {
		os.Setenv("ENABLED_MESSENGERS", "line")
		os.Unsetenv("LINE_CHANNEL_TOKEN")

		_, err := Load()
		if err == nil {
			t.Error("Expected Load() to fail when line enabled but token missing")
		}
	})
}
