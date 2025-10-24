package config

import (
	"log"
	"os"
	"strings"
)

// Config holds all application configuration settings.
type Config struct {
	TgBotToken                    string
	AdminTelegramID               string // Имя поля в структуре
	AllowedTelegramIDs            []string
	GoogleCalendarCredentialsPath string
	GoogleCalendarID              string
}

// LoadConfig loads configuration from environment variables.
func LoadConfig() *Config {
	token := os.Getenv("TG_BOT_TOKEN") // Используем TG_BOT_TOKEN
	if token == "" {
		log.Fatal("Environment variable TG_BOT_TOKEN is not set.")
	}

	// ИСПРАВЛЕНО: Читаем из переменной окружения TG_ADMIN_ID
	adminID := os.Getenv("TG_ADMIN_ID")
	if adminID == "" {
		log.Println("Warning: Environment variable TG_ADMIN_ID is not set. Admin features might be limited.")
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
		log.Println("Warning: Environment variable ALLOWED_TELEGRAM_IDS is not set. All users will be allowed (or implement specific logic).")
	}

	// ИСПРАВЛЕНО: Теперь ожидаем GOOGLE_CREDENTIALS_PATH
	googleCredsPath := os.Getenv("GOOGLE_CREDENTIALS_PATH")
	if googleCredsPath == "" {
		log.Fatal("Environment variable GOOGLE_CREDENTIALS_PATH is not set. Please set it to the path of your credentials.json file (e.g., ./credentials.json).")
	}

	googleCalendarID := os.Getenv("GOOGLE_CALENDAR_ID")
	if googleCalendarID == "" {
		log.Println("Warning: Environment variable GOOGLE_CALENDAR_ID is not set. Defaulting to 'primary' calendar.")
		googleCalendarID = "primary"
	}

	return &Config{
		TgBotToken:                    token,
		AdminTelegramID:               adminID, // Сохраняем имя поля
		AllowedTelegramIDs:            allowedIDs,
		GoogleCalendarCredentialsPath: googleCredsPath,
		GoogleCalendarID:              googleCalendarID,
	}
}
