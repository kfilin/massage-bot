package config

import (
	"os"
	"strings"

	"github.com/kfilin/massage-bot/internal/logging"
)

// Config holds all application configuration settings.
type Config struct {
	TgBotToken                    string
	AdminTelegramID               string
	AllowedTelegramIDs            []string
	GoogleCalendarCredentialsPath string
	GoogleCalendarCredentialsJSON string
	GoogleCalendarID              string
	GroqAPIKey                    string
	TherapistIDs                  []string
	WebAppURL                     string
	WebAppSecret                  string
	WebAppPort                    string
}

// LoadConfig loads configuration from environment variables.
func LoadConfig() *Config {
	token := os.Getenv("TG_BOT_TOKEN")
	if token == "" {
		logging.Fatal("Environment variable TG_BOT_TOKEN is not set.")
	}

	adminID := os.Getenv("TG_ADMIN_ID")
	if adminID == "" {
		logging.Warn("Warning: Environment variable TG_ADMIN_ID is not set. Admin features might be limited.")
	}

	allowedIDsStr := os.Getenv("ALLOWED_TELEGRAM_IDS")
	var allowedIDs []string
	if allowedIDsStr != "" {
		ids := strings.Split(allowedIDsStr, ",")
		for _, id := range ids {
			trimmedID := strings.TrimSpace(id)
			if trimmedID != "" {
				allowedIDs = append(allowedIDs, trimmedID)
			}
		}
	} else {
		logging.Warn("Warning: Environment variable ALLOWED_TELEGRAM_IDS is not set.")
	}

	// PROFESSIONAL FIX: Support both file path and environment variable
	googleCredsPath := os.Getenv("GOOGLE_CREDENTIALS_PATH")
	googleCredsJSON := os.Getenv("GOOGLE_CREDENTIALS_JSON")

	if googleCredsPath == "" && googleCredsJSON == "" {
		logging.Fatal("Set either GOOGLE_CREDENTIALS_PATH (for Docker) or GOOGLE_CREDENTIALS_JSON (for Kubernetes)")
	}

	googleCalendarID := os.Getenv("GOOGLE_CALENDAR_ID")
	if googleCalendarID == "" {
		logging.Warn("Warning: GOOGLE_CALENDAR_ID not set. Defaulting to 'primary'.")
		googleCalendarID = "primary"
	}

	therapistIDsStr := os.Getenv("TG_THERAPIST_ID")
	var therapistIDs []string
	if therapistIDsStr != "" {
		ids := strings.Split(therapistIDsStr, ",")
		for _, id := range ids {
			trimmedID := strings.TrimSpace(id)
			if trimmedID != "" {
				therapistIDs = append(therapistIDs, trimmedID)
			}
		}
	}

	return &Config{
		TgBotToken:                    token,
		AdminTelegramID:               adminID,
		AllowedTelegramIDs:            allowedIDs,
		GoogleCalendarCredentialsPath: googleCredsPath,
		GoogleCalendarCredentialsJSON: googleCredsJSON,
		GoogleCalendarID:              googleCalendarID,
		GroqAPIKey:                    os.Getenv("GROQ_API_KEY"),
		TherapistIDs:                  therapistIDs,
		WebAppURL:                     os.Getenv("WEBAPP_URL"),
		WebAppSecret:                  os.Getenv("WEBAPP_SECRET"),
		WebAppPort:                    os.Getenv("WEBAPP_PORT"),
	}
}
