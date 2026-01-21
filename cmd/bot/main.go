package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/kfilin/massage-bot/cmd/bot/config"
	"github.com/kfilin/massage-bot/internal/adapters/googlecalendar"
	"github.com/kfilin/massage-bot/internal/adapters/transcription"
	"github.com/kfilin/massage-bot/internal/delivery/telegram"
	"github.com/kfilin/massage-bot/internal/services/appointment"
	"github.com/kfilin/massage-bot/internal/storage"
)

func main() {
	// Загружаем переменные окружения из файла .env
	// Это должно быть самой первой операцией в main(), чтобы другие части могли получить доступ к env vars.
	if err := godotenv.Load(); err != nil {
		log.Printf("No .env file found or error loading .env file: %v", err)
		// Не фатальная ошибка, если .env не найден, так как env vars могут быть установлены в системе.
	}

	// 1. Load Configuration
	cfg := config.LoadConfig()
	log.Println("Configuration loaded.")
	log.Println("Bot version: v3.1.7")

	// Start health server
	go startHealthServer()

	// 1b. Initialize Database
	db, err := storage.InitDB()
	if err != nil {
		log.Fatalf("Error initializing database: %v", err)
	}
	patientRepo := storage.NewPostgresRepository(db, os.Getenv("DATA_DIR"))
	log.Println("Database initialized.")

	// 1c. Run Migration (Idempotent)
	if err := storage.MigrateJSONToPostgres(patientRepo, os.Getenv("DATA_DIR")); err != nil {
		log.Printf("ERROR during migration: %v", err)
	}

	// Start Web App server
	if cfg.WebAppSecret != "" {
		go startWebAppServer(cfg.WebAppPort, cfg.WebAppSecret, patientRepo)
	} else {
		log.Println("Warning: WEBAPP_SECRET not set, Web App server not started.")
	}

	// 2. Initialize Google Calendar Client
	googleCalendarClient, err := googlecalendar.NewGoogleCalendarClient()
	if err != nil {
		log.Fatalf("Error initializing Google Calendar client: %v", err)
	}
	log.Println("Google Calendar client initialized.")

	// 3. Initialize AppointmentRepository (Google Calendar Adapter implements this)
	// Pass the client and the calendar ID from config
	appointmentRepo := googlecalendar.NewAdapter(googleCalendarClient, cfg.GoogleCalendarID)
	log.Println("Appointment repository (Google Calendar adapter) initialized.")

	// 4. Initialize AppointmentService (business logic)
	appointmentService := appointment.NewService(appointmentRepo)
	log.Println("Appointment service initialized.")

	// 5. Initialize SessionStorage (using PostgreSQL persistence)
	sessionStorage := storage.NewPostgresSessionStorage(db)
	log.Println("Postgres session storage initialized.")

	// 6. Initialize Advanced Adapters (Transcription)
	transcriptionAdapter := transcription.NewGroqAdapter(cfg.GroqAPIKey)
	log.Println("Advanced adapters (Groq) initialized.")

	// 7. Start the Telegram Bot
	// Pass all initialized dependencies to the bot's start function
	log.Println("Starting Telegram bot...")
	telegram.StartBot(
		cfg.TgBotToken,
		appointmentService,
		sessionStorage,
		cfg.AdminTelegramID,
		cfg.AllowedTelegramIDs,
		transcriptionAdapter,
		patientRepo,
		cfg.WebAppURL,
		cfg.WebAppSecret,
	)
}
