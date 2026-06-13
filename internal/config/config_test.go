package config

import (
	"os"
	"reflect"
	"testing"
)

// envSnapshot captures the current values of the env vars used by LoadConfig.
type envSnapshot struct {
	token          string
	adminID        string
	allowedIDs     string
	credsPath      string
	credsJSON      string
	calendarID     string
	therapistID    string
	webAppURL      string
	webAppSecret   string
	webAppPort     string
	groqAPIKey     string
}

func captureEnv(t *testing.T) envSnapshot {
	t.Helper()
	return envSnapshot{
		token:        os.Getenv("TG_BOT_TOKEN"),
		adminID:      os.Getenv("TG_ADMIN_ID"),
		allowedIDs:   os.Getenv("ALLOWED_TELEGRAM_IDS"),
		credsPath:    os.Getenv("GOOGLE_CREDENTIALS_PATH"),
		credsJSON:    os.Getenv("GOOGLE_CREDENTIALS_JSON"),
		calendarID:   os.Getenv("GOOGLE_CALENDAR_ID"),
		therapistID:  os.Getenv("TG_THERAPIST_ID"),
		webAppURL:    os.Getenv("WEBAPP_URL"),
		webAppSecret: os.Getenv("WEBAPP_SECRET"),
		webAppPort:   os.Getenv("WEBAPP_PORT"),
		groqAPIKey:   os.Getenv("GROQ_API_KEY"),
	}
}

func restoreEnv(t *testing.T, snap envSnapshot) {
	t.Helper()
	setOrUnset := func(key, value string) {
		if value == "" {
			_ = os.Unsetenv(key)
		} else {
			_ = os.Setenv(key, value)
		}
	}
	setOrUnset("TG_BOT_TOKEN", snap.token)
	setOrUnset("TG_ADMIN_ID", snap.adminID)
	setOrUnset("ALLOWED_TELEGRAM_IDS", snap.allowedIDs)
	setOrUnset("GOOGLE_CREDENTIALS_PATH", snap.credsPath)
	setOrUnset("GOOGLE_CREDENTIALS_JSON", snap.credsJSON)
	setOrUnset("GOOGLE_CALENDAR_ID", snap.calendarID)
	setOrUnset("TG_THERAPIST_ID", snap.therapistID)
	setOrUnset("WEBAPP_URL", snap.webAppURL)
	setOrUnset("WEBAPP_SECRET", snap.webAppSecret)
	setOrUnset("WEBAPP_PORT", snap.webAppPort)
	setOrUnset("GROQ_API_KEY", snap.groqAPIKey)
}

func clearConfigEnv(t *testing.T) {
	t.Helper()
	snap := captureEnv(t)
	t.Cleanup(func() { restoreEnv(t, snap) })

	_ = os.Unsetenv("TG_BOT_TOKEN")
	_ = os.Unsetenv("TG_ADMIN_ID")
	_ = os.Unsetenv("ALLOWED_TELEGRAM_IDS")
	_ = os.Unsetenv("GOOGLE_CREDENTIALS_PATH")
	_ = os.Unsetenv("GOOGLE_CREDENTIALS_JSON")
	_ = os.Unsetenv("GOOGLE_CALENDAR_ID")
	_ = os.Unsetenv("TG_THERAPIST_ID")
	_ = os.Unsetenv("WEBAPP_URL")
	_ = os.Unsetenv("WEBAPP_SECRET")
	_ = os.Unsetenv("WEBAPP_PORT")
	_ = os.Unsetenv("GROQ_API_KEY")
}

func TestLoadConfigDefaults(t *testing.T) {
	clearConfigEnv(t)

	_ = os.Setenv("TG_BOT_TOKEN", "test_token")
	_ = os.Setenv("GOOGLE_CREDENTIALS_JSON", "{}")

	cfg := LoadConfig()

	if cfg.TgBotToken != "test_token" {
		t.Errorf("Expected token test_token, got %s", cfg.TgBotToken)
	}

	if cfg.GoogleCalendarID != "primary" {
		t.Errorf("Expected default calendar ID 'primary', got %s", cfg.GoogleCalendarID)
	}
}

func TestLoadConfigAllowedIDs(t *testing.T) {
	clearConfigEnv(t)
	_ = os.Setenv("TG_BOT_TOKEN", "test_token")
	_ = os.Setenv("GOOGLE_CREDENTIALS_JSON", "{}")
	_ = os.Setenv("ALLOWED_TELEGRAM_IDS", "123, 456 ,789")

	cfg := LoadConfig()

	expected := []string{"123", "456", "789"}
	if !reflect.DeepEqual(cfg.AllowedTelegramIDs, expected) {
		t.Errorf("Expected allowed IDs %v, got %v", expected, cfg.AllowedTelegramIDs)
	}
}

func TestLoadConfigAllowedIDsSkipsEmpty(t *testing.T) {
	clearConfigEnv(t)
	_ = os.Setenv("TG_BOT_TOKEN", "test_token")
	_ = os.Setenv("GOOGLE_CREDENTIALS_JSON", "{}")
	_ = os.Setenv("ALLOWED_TELEGRAM_IDS", "123,, 456 , , 789")

	cfg := LoadConfig()

	expected := []string{"123", "456", "789"}
	if !reflect.DeepEqual(cfg.AllowedTelegramIDs, expected) {
		t.Errorf("Expected allowed IDs %v, got %v", expected, cfg.AllowedTelegramIDs)
	}
}

func TestLoadConfigTherapistIDs(t *testing.T) {
	clearConfigEnv(t)
	_ = os.Setenv("TG_BOT_TOKEN", "test_token")
	_ = os.Setenv("GOOGLE_CREDENTIALS_JSON", "{}")
	_ = os.Setenv("TG_THERAPIST_ID", "1122, 3344 ,5566")

	cfg := LoadConfig()

	expected := []string{"1122", "3344", "5566"}
	if !reflect.DeepEqual(cfg.TherapistIDs, expected) {
		t.Errorf("Expected therapist IDs %v, got %v", expected, cfg.TherapistIDs)
	}
}

func TestLoadConfigPreservesOptionalFields(t *testing.T) {
	clearConfigEnv(t)
	_ = os.Setenv("TG_BOT_TOKEN", "test_token")
	_ = os.Setenv("GOOGLE_CREDENTIALS_JSON", "creds")
	_ = os.Setenv("TG_ADMIN_ID", "admin42")
	_ = os.Setenv("GROQ_API_KEY", "key99")
	_ = os.Setenv("WEBAPP_URL", "https://example.com")
	_ = os.Setenv("WEBAPP_SECRET", "shh")
	_ = os.Setenv("WEBAPP_PORT", "8080")

	cfg := LoadConfig()

	if cfg.AdminTelegramID != "admin42" {
		t.Errorf("Expected admin ID admin42, got %s", cfg.AdminTelegramID)
	}
	if cfg.GroqAPIKey != "key99" {
		t.Errorf("Expected GroqAPIKey key99, got %s", cfg.GroqAPIKey)
	}
	if cfg.WebAppURL != "https://example.com" {
		t.Errorf("Expected WebAppURL https://example.com, got %s", cfg.WebAppURL)
	}
	if cfg.WebAppSecret != "shh" {
		t.Errorf("Expected WebAppSecret shh, got %s", cfg.WebAppSecret)
	}
	if cfg.WebAppPort != "8080" {
		t.Errorf("Expected WebAppPort 8080, got %s", cfg.WebAppPort)
	}
}

func TestLoadConfigMissingToken(t *testing.T) {
	clearConfigEnv(t)
	_ = os.Setenv("GOOGLE_CREDENTIALS_JSON", "{}")

	var fatalCalled bool
	fatal := func(args ...interface{}) {
		fatalCalled = true
		panic(args[0])
	}

	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected fatal callback to panic")
		}
		if !fatalCalled {
			t.Error("Expected fatal callback to be called")
		}
	}()

	_ = LoadConfigWithFatal(fatal)
}

func TestLoadConfigMissingGoogleCredentials(t *testing.T) {
	clearConfigEnv(t)
	_ = os.Setenv("TG_BOT_TOKEN", "test_token")

	var fatalCalled bool
	fatal := func(args ...interface{}) {
		fatalCalled = true
		panic(args[0])
	}

	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected fatal callback to panic")
		}
		if !fatalCalled {
			t.Error("Expected fatal callback to be called")
		}
	}()

	_ = LoadConfigWithFatal(fatal)
}

func TestLoadConfigGooglePathCredential(t *testing.T) {
	clearConfigEnv(t)
	_ = os.Setenv("TG_BOT_TOKEN", "test_token")
	_ = os.Setenv("GOOGLE_CREDENTIALS_PATH", "/tmp/creds.json")

	cfg := LoadConfig()

	if cfg.GoogleCalendarCredentialsPath != "/tmp/creds.json" {
		t.Errorf("Expected credentials path /tmp/creds.json, got %s", cfg.GoogleCalendarCredentialsPath)
	}
}
