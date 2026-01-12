package main

import (
	"log"

	"github.com/joho/godotenv" // <-- ДОБАВЛЕН ИМПОРТ godotenv
	"github.com/kfilin/massage-bot/cmd/bot/config"
	"github.com/kfilin/massage-bot/internal/adapters/googlecalendar"
	"github.com/kfilin/massage-bot/internal/delivery/telegram"
	"github.com/kfilin/massage-bot/internal/services/appointment"
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

        // Start health server
         go startHealthServer()

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

	// 5. Initialize SessionStorage (using the in-memory implementation)
	sessionStorage := telegram.NewInMemorySessionStorage()
	log.Println("In-memory session storage initialized.")

	// 6. Start the Telegram Bot
	// Pass all initialized dependencies to the bot's start function
	log.Println("Starting Telegram bot...")
	telegram.StartBot(
		cfg.TgBotToken,
		appointmentService,
		sessionStorage,
		cfg.AdminTelegramID,
		cfg.AllowedTelegramIDs,
	)
	// StartBot is a blocking call, so code after this will not execute until bot stops.
}
