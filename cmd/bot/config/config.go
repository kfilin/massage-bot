package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	TgBotToken string
	TgAdminID  int64
}

// LoadConfig loads configuration from environment variables.
func LoadConfig() *Config {
	// Load .env file for local development (if it exists)
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found or error loading it, proceeding with environment variables:", err)
	}

	cfg := &Config{
		TgBotToken: os.Getenv("TG_BOT_TOKEN"),
	}

	if cfg.TgBotToken == "" {
		log.Fatal("TG_BOT_TOKEN environment variable is required.")
	}

	adminIDStr := os.Getenv("TG_ADMIN_ID")
	if adminIDStr != "" {
		adminID, err := strconv.ParseInt(adminIDStr, 10, 64)
		if err != nil {
			log.Printf("Warning: Could not parse TG_ADMIN_ID from env: %v. Admin features may not work.", err)
		} else {
			cfg.TgAdminID = adminID
		}
	} else {
		log.Println("TG_ADMIN_ID environment variable not set. Admin features will be unavailable.")
	}

	return cfg
}
