package config

import (
	"os"
	"testing"
)

func TestLoadConfigDefaults(t *testing.T) {
	// Clear relevant env vars
	os.Unsetenv("TG_BOT_TOKEN")
	os.Unsetenv("GOOGLE_CREDENTIALS_JSON")
	os.Unsetenv("GOOGLE_CREDENTIALS_PATH")
	os.Unsetenv("GOOGLE_CALENDAR_ID")

	// Set required ones to avoid immediate fatal
	os.Setenv("TG_BOT_TOKEN", "test_token")
	os.Setenv("GOOGLE_CREDENTIALS_JSON", "{}")

	cfg := LoadConfig()

	if cfg.TgBotToken != "test_token" {
		t.Errorf("Expected token test_token, got %s", cfg.TgBotToken)
	}

	if cfg.GoogleCalendarID != "primary" {
		t.Errorf("Expected default calendar ID 'primary', got %s", cfg.GoogleCalendarID)
	}
}

func TestLoadConfigAllowedIDs(t *testing.T) {
	os.Setenv("TG_BOT_TOKEN", "test_token")
	os.Setenv("GOOGLE_CREDENTIALS_JSON", "{}")
	os.Setenv("ALLOWED_TELEGRAM_IDS", "123, 456 ,789")

	cfg := LoadConfig()

	expected := []string{"123", "456", "789"}
	if len(cfg.AllowedTelegramIDs) != len(expected) {
		t.Fatalf("Expected %d allowed IDs, got %d", len(expected), len(cfg.AllowedTelegramIDs))
	}

	for i, id := range cfg.AllowedTelegramIDs {
		if id != expected[i] {
			t.Errorf("At index %d: expected %s, got %s", i, expected[i], id)
		}
	}
}
