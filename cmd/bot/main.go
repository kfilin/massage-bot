package main

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/kfilin/massage-bot/internal/adapters/googlecalendar"
	"github.com/kfilin/massage-bot/internal/adapters/transcription"
	"github.com/kfilin/massage-bot/internal/config"
	"github.com/kfilin/massage-bot/internal/delivery/telegram"
	"github.com/kfilin/massage-bot/internal/domain"
	"github.com/kfilin/massage-bot/internal/logging"
	"github.com/kfilin/massage-bot/internal/services/appointment"
	"github.com/kfilin/massage-bot/internal/storage"
	"github.com/kfilin/massage-bot/internal/version"
)

func main() {
	// Force application timezone to Europe/Istanbul
	time.Local = domain.ApptTimeZone
	// Загружаем переменные окружения из файла .env
	// Это должно быть самой первой операцией в main(), чтобы другие части могли получить доступ к env vars.
	if err := godotenv.Load(); err != nil {
		// Can't use structured logging yet as config not loaded, but acceptable for init
		// However, let's use standard log here strictly before Init
		logging.Infof("No .env file found or error loading .env file: %v", err)
		// Не фатальная ошибка, если .env не найден, так как env vars могут быть установлены в системе.
	}

	// 1. Load Configuration
	cfg := config.LoadConfig()
	logging.Init(os.Getenv("LOG_LEVEL") == "DEBUG")
	logging.Info("Configuration loaded.")
	logging.Infof("Bot version: %s", version.FullName)

	// Start health server
	// Health server started later with context

	// 1b. Initialize Database
	db, err := storage.InitDB()
	if err != nil {
		logging.Fatalf("CRITICAL: Error initializing database: %v", err)
	}
	patientRepo := storage.NewPostgresRepository(db, os.Getenv("DATA_DIR"))
	patientRepo.BotVersion = version.Version
	logging.Info("Database initialized.")

	// 1c. Run Migration (Idempotent)
	if err := storage.MigrateJSONToPostgres(patientRepo, os.Getenv("DATA_DIR")); err != nil {
		logging.Errorf("ERROR during migration: %v", err)
	}

	// 2. Initialize Google Calendar Client
	googleCalendarClient, err := googlecalendar.NewGoogleCalendarClient()
	if err != nil {
		logging.Fatalf("Error initializing Google Calendar client: %v", err)
	}
	logging.Info("Google Calendar client initialized.")

	// 3. Initialize AppointmentRepository (Google Calendar Adapter implements this)
	// Pass the client and the calendar ID from config
	appointmentRepo := googlecalendar.NewAdapter(googleCalendarClient, cfg.GoogleCalendarID)
	logging.Info("Appointment repository (Google Calendar adapter) initialized.")

	// 4. Initialize AppointmentService (business logic)
	appointmentService := appointment.NewService(appointmentRepo, patientRepo)
	logging.Info("Appointment service initialized.")

	// 5. Initialize SessionStorage (using PostgreSQL persistence)
	sessionStorage := storage.NewPostgresSessionStorage(db)
	logging.Info("Postgres session storage initialized.")

	// 6. Initialize Advanced Adapters (Transcription)
	transcriptionAdapter := transcription.NewGroqAdapter(cfg.GroqAPIKey)
	logging.Info("Advanced adapters (Groq) initialized.")

	// 6b. Start Web App server (now with apptService for sync)
	// 8. Graceful Shutdown Orchestration
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	var wg sync.WaitGroup

	// Start health server
	wg.Add(1)
	go func() {
		defer wg.Done()
		startHealthServer(ctx)
	}()

	// 7. Initialize and Start the Telegram Bot
	logging.Info("Starting Telegram bot...")
	bot, err := telegram.InitBot(cfg.TgBotToken)
	if err != nil {
		logging.Fatalf("CRITICAL: Failed to initialize Telegram bot: %v", err)
	}
	botUsername := bot.Me.Username
	logging.Infof("Authenticated as @%s", botUsername)

	// Set metadata in repository for dynamic link generation
	patientRepo.BotUsername = botUsername

	// 8. Start Web App server
	if cfg.WebAppSecret != "" {
		adminMap := make(map[string]struct{})
		if cfg.AdminTelegramID != "" {
			adminMap[cfg.AdminTelegramID] = struct{}{}
		}
		for _, id := range cfg.AllowedTelegramIDs {
			if id != "" {
				adminMap[id] = struct{}{}
			}
		}
		for _, id := range cfg.TherapistIDs {
			if id != "" {
				adminMap[id] = struct{}{}
			}
		}

		var allAdmins []string
		for id := range adminMap {
			allAdmins = append(allAdmins, id)
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			startWebAppServer(ctx, cfg.WebAppPort, cfg.WebAppSecret, cfg.TgBotToken, allAdmins, patientRepo, appointmentService, transcriptionAdapter, os.Getenv("DATA_DIR"), botUsername)
		}()
	} else {
		logging.Warn("Warning: WEBAPP_SECRET not set, Web App server not started.")
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		telegram.RunBot(
			ctx,
			bot,
			appointmentService,
			sessionStorage,
			cfg.AdminTelegramID,
			cfg.AllowedTelegramIDs,
			transcriptionAdapter,
			patientRepo,
			cfg.WebAppURL,
			cfg.WebAppSecret,
			cfg.TherapistIDs,
		)
	}()

	// Wait for interruption signal
	<-ctx.Done()
	logging.Info("Shutdown signal received. Shutting down components...")

	// Wait for all goroutines to finish
	// We might want to add a timeout for the wait itself
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		logging.Info("Graceful shutdown completed.")
	case <-shutdownCtx.Done():
		logging.Warn("Shutdown timed out. forcing exit.")
	}
}
