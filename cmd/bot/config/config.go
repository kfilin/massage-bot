package config

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Telegram struct {
		Token   string
		AdminID int64
	}
	Google struct {
		CredentialsFile string
	}
	Timezone *time.Location
}

func Load() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	tz, _ := time.LoadLocation(os.Getenv("TIMEZONE"))

	adminID, _ := strconv.ParseInt(os.Getenv("ADMIN_ID"), 10, 64)

	return &Config{
		Telegram: struct {
			Token   string
			AdminID int64
		}{
			Token:   os.Getenv("TELEGRAM_TOKEN"),
			AdminID: adminID,
		},
		Google: struct {
			CredentialsFile string
		}{
			CredentialsFile: os.Getenv("GOOGLE_CREDENTIALS_FILE"),
		},
		Timezone: tz,
	}
}
