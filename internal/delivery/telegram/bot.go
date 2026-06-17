package telegram

import (
	"context"
	"fmt"

	// "log" // Replaced by internal/logging
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/kfilin/massage-bot/internal/config"
	"github.com/kfilin/massage-bot/internal/delivery/telegram/handlers"
	"github.com/kfilin/massage-bot/internal/logging"
	"github.com/kfilin/massage-bot/internal/monitoring"
	"github.com/kfilin/massage-bot/internal/ports" // Import ports for interfaces
	"github.com/kfilin/massage-bot/internal/presentation"
	"github.com/kfilin/massage-bot/internal/services/reminder"

	// Added reminder service
	// Import storage pkg for ban check
	"gopkg.in/telebot.v3"
)

// Callback prefix constants for inline button handling
const (
	CallbackPrefixCategory        = "select_category|"
	CallbackPrefixService         = "select_service|"
	CallbackPrefixDate            = "select_date|"
	CallbackPrefixNavigateMonth   = "navigate_month|"
	CallbackPrefixTime            = "select_time|"
	CallbackPrefixCancelAppt      = "cancel_appt|"
	CallbackPrefixConfirmReminder = "confirm_appt_reminder|"
	CallbackPrefixCancelReminder  = "cancel_appt_reminder|"
	CallbackPrefixAdminReply      = "admin_reply|"
	CallbackConfirmBooking        = "confirm_booking"
	CallbackCancelBooking         = "cancel_booking"
	CallbackBackToServices        = "back_to_services"
	CallbackBackToDate            = "back_to_date"
	CallbackApproveDraft          = "approve_draft"
	CallbackDiscardDraft          = "discard_draft"
	CallbackIgnore                = "ignore"
)

// StartBot initializes and runs the Telegram bot.
// It now receives all necessary services and configuration from the main package.
// InitBot initializes the bot and returns the instance.
func InitBot(token string) (*telebot.Bot, error) {
	pref := telebot.Settings{
		Token:  token,
		Poller: &telebot.LongPoller{Timeout: 10 * time.Second},
	}

	var b *telebot.Bot
	var err error

	// Resilience: Retry loop for bot initialization
	for i := 0; i < 10; i++ {
		b, err = telebot.NewBot(pref)
		if err == nil {
			return b, nil
		}
		logging.Debugf("DEBUG_RETRY: Error creating bot (attempt %d/10): %v", i+1, err)
		time.Sleep(time.Duration(i*2+5) * time.Second) // Exponential-ish backoff
	}
	return nil, err
}

// setupMenuButton configures the chat menu button for the bot via the
// Telegram Bot API. When webAppURL is empty, this is a no-op (the
// feature is disabled). On success, it logs an info line. On failure
// it logs a warning and returns the underlying error so the caller
// (typically RunBot) can decide how to react.
//
// Extracted from RunBot so it can be unit-tested with a mock BotAPI
// (see bot_wiring_test.go). The real *telebot.Bot satisfies BotAPI.
func setupMenuButton(b ports.BotAPI, webAppURL string) error {
	if webAppURL == "" {
		return nil
	}
	// Use raw API call to avoid nil pointer issues with SetMenuButton
	params := map[string]interface{}{
		"menu_button": map[string]interface{}{
			"type": "web_app",
			"text": "Открыть карту",
			"web_app": map[string]string{
				"url": webAppURL,
			},
		},
	}
	if _, err := b.Raw("setChatMenuButton", params); err != nil {
		logging.Warnf("Failed to set menu button: %v", err)
		return err
	}
	logging.Info("Menu button configured successfully")
	return nil
}

// runScheduledBackup performs one cycle of the daily backup job:
//  1. ask the repository to create a ZIP of patient data
//  2. send that ZIP to the primary admin as a Telegram Document
//  3. remove the local temp file to save server disk
//
// Any failure (CreateBackup error, invalid admin ID, Send failure) is
// logged and returned to the caller. The function is safe to call
// even if the temp file was never created (os.Remove is a no-op on
// missing files). Extracted from RunBot for testability.
func runScheduledBackup(repo ports.Repository, b ports.BotAPI, adminTelegramID string) error {
	logging.Infof("[BackupWorker] Starting scheduled daily backup for Admin %s...", adminTelegramID)
	zipPath, err := repo.CreateBackup()
	if err != nil {
		logging.Errorf("[BackupWorker] FAILED to create scheduled backup: %v", err)
		return err
	}
	defer os.Remove(zipPath)

	adminIntID, err := strconv.ParseInt(adminTelegramID, 10, 64)
	if err != nil {
		logging.Errorf("[BackupWorker] Invalid admin ID %q: %v", adminTelegramID, err)
		return err
	}
	doc := &telebot.Document{
		File:     telebot.FromDisk(zipPath),
		FileName: filepath.Base(zipPath),
		Caption:  fmt.Sprintf("💾 Ежедневная копия данных (%s)\n\n<i>Примечание: Локальный файл удален для экономии места.</i>", time.Now().Format("02.01.2006")),
	}
	if _, err := b.Send(&telebot.User{ID: adminIntID}, doc, telebot.ModeHTML); err != nil {
		logging.Errorf("[BackupWorker] FAILED to send scheduled backup: %v", err)
		return err
	}
	return nil
}

// RunBot starts the main event loop of the Telegram bot.
func RunBot(
	ctx context.Context,
	b *telebot.Bot,
	appointmentService ports.AppointmentService,
	sessionStorage ports.SessionStorage,
	adminTelegramID string,
	allowedTelegramIDs []string,
	trans ports.TranscriptionService,
	repo ports.Repository,
	webAppURL string,
	webAppSecret string,
	therapistIDs []string,
) {
	// Set menu button for quick TWA access. The raw API call is wrapped
	// by setupMenuButton so this behaviour is unit-testable.
	_ = setupMenuButton(b, webAppURL)

	// Ensure Admin ID is in the allowed list for notifications.
	// config.ResolveAdminIDs deduplicates the three sources (primary, allowed, therapist).
	finalAdminIDs := config.ResolveAdminIDs(adminTelegramID, allowedTelegramIDs, therapistIDs)

	botPresenter := presentation.NewBotPresenter()
	bookingHandler := handlers.NewBookingHandler(appointmentService, sessionStorage, finalAdminIDs, therapistIDs, trans, repo, botPresenter, webAppURL, webAppSecret)

	// Initialize and start Reminder Service
	reminderService := reminder.NewService(appointmentService, repo, b, finalAdminIDs, botPresenter)
	reminderService.Start(ctx)

	// Start Daily Backup Worker (Sent to Primary Admin). The inner
	// work is delegated to runScheduledBackup so it is testable.
	if adminTelegramID != "" {
		go func() {
			// Wait 5 minutes after startup to avoid congestion
			time.Sleep(5 * time.Minute)
			ticker := time.NewTicker(24 * time.Hour)
			defer ticker.Stop()

			for range ticker.C {
				_ = runScheduledBackup(repo, b, adminTelegramID)
			}
		}()
	}

	// GLOBAL MIDDLEWARE: Enforce ban check on ALL entry points
	b.Use(func(next telebot.HandlerFunc) telebot.HandlerFunc {
		return func(c telebot.Context) error {
			if c.Sender() == nil {
				return next(c)
			}
			telegramID := strconv.FormatInt(c.Sender().ID, 10)
			username := c.Sender().Username
			if banned, _ := repo.IsUserBanned(telegramID, username); banned {
				logging.Warnf("BLOCKED (Middleware): Banned user %s (@%s) tried to access bot.", telegramID, username)

				// SHADOW BAN: Polite "No spots available" message
				shadowBanMsg := "К сожалению, на данный момент свободных мест для записи нет. Попробуйте позже."

				if c.Callback() != nil {
					return c.Respond(&telebot.CallbackResponse{
						Text:      shadowBanMsg,
						ShowAlert: true,
					})
				}
				return c.Send(shadowBanMsg, telebot.RemoveKeyboard)
			}
			return next(c)
		}
	})

	// GLOBAL MIDDLEWARE: Record metrics for command/callback usage
	b.Use(func(next telebot.HandlerFunc) telebot.HandlerFunc {
		return func(c telebot.Context) error {
			if c.Message() != nil {
				text := c.Message().Text
				logging.Debugf("DEBUG: Incoming Message from %d (%s): %s", c.Sender().ID, c.Sender().Username, text)
				if strings.HasPrefix(text, "/") {
					command := strings.Split(text, " ")[0]
					monitoring.BotCommandsTotal.WithLabelValues(command).Inc()
				} else {
					monitoring.BotCommandsTotal.WithLabelValues("text_message").Inc()
				}
			} else if c.Callback() != nil {
				logging.Debugf("DEBUG: Incoming Callback from %d (%s): %s", c.Sender().ID, c.Sender().Username, c.Callback().Data)
				monitoring.BotCommandsTotal.WithLabelValues("callback").Inc()
			}
			return next(c)
		}
	})

	b.Handle("/start", bookingHandler.HandleStart)
	b.Handle("/cancel", bookingHandler.HandleCancel)
	b.Handle("/myrecords", bookingHandler.HandleMyRecords)
	b.Handle("/myappointments", bookingHandler.HandleMyAppointments)
	b.Handle("/upload", bookingHandler.HandleUploadCommand)
	b.Handle("/backup", bookingHandler.HandleBackup)
	b.Handle("/ban", bookingHandler.HandleBan)
	b.Handle("/unban", bookingHandler.HandleUnban)
	b.Handle("/block", bookingHandler.HandleBlock)
	b.Handle("/status", bookingHandler.HandleStatus)
	b.Handle("/edit_name", bookingHandler.HandleEditName)
	b.Handle("/patients", bookingHandler.HandleListPatients)
	b.Handle("/create_appointment", bookingHandler.HandleManualAppointment)
	b.Handle("/manual", bookingHandler.HandleManualAppointment)
	b.Handle("/book", bookingHandler.HandleStart)

	// Register file/media handlers
	b.Handle(telebot.OnDocument, bookingHandler.HandleFileMessage)
	b.Handle(telebot.OnPhoto, bookingHandler.HandleFileMessage)
	b.Handle(telebot.OnVideo, bookingHandler.HandleFileMessage)
	b.Handle(telebot.OnAnimation, bookingHandler.HandleFileMessage)
	b.Handle(telebot.OnVoice, bookingHandler.HandleFileMessage)

	// Re-send main menu when TWA is closed (BackButton sends data back)
	b.Handle(telebot.OnWebApp, func(c telebot.Context) error {
		return c.Send("📋 Главное меню", bookingHandler.GetMainMenu())
	})

	// Обработчик для всех inline-кнопок
	b.Handle(telebot.OnCallback, func(c telebot.Context) error {
		logging.Debugf(": Entered OnCallback handler.")

		data := c.Callback().Data
		// Обрезаем пробелы в начале и конце строки данных колбэка
		trimmedData := strings.TrimSpace(data)
		logging.Debugf("Received callback: '%s' (trimmed: '%s') from user %d", data, trimmedData, c.Sender().ID)

		defer func() {
			if err := c.Respond(); err != nil {
				logging.Warnf("Failed to respond to callback: %v", err)
			}
		}() // Важно: Respond() должен быть вызван, чтобы убрать "часики" с кнопки

		action, matched := RouteCallback(trimmedData)
		if !matched {
			logging.Warnf("DEBUG: OnCallback: No specific callback prefix matched for data: '%s'", trimmedData)
			return c.Send("Неизвестное действие с кнопкой. Пожалуйста, начните /start снова.")
		}

		switch action {
		case CallbackPrefixCategory:
			return bookingHandler.HandleCategorySelection(c)
		case CallbackPrefixService:
			return bookingHandler.HandleServiceSelection(c)
		case CallbackPrefixDate, CallbackPrefixNavigateMonth, CallbackBackToServices:
			return bookingHandler.HandleDateSelection(c)
		case CallbackPrefixTime, CallbackBackToDate:
			return bookingHandler.HandleTimeSelection(c)
		case CallbackConfirmBooking:
			return bookingHandler.HandleConfirmBooking(c)
		case CallbackCancelBooking:
			return bookingHandler.HandleCancel(c)
		case CallbackPrefixCancelAppt:
			return bookingHandler.HandleCancelAppointmentCallback(c)
		case CallbackPrefixConfirmReminder:
			return bookingHandler.HandleReminderConfirmation(c)
		case CallbackPrefixCancelReminder:
			return bookingHandler.HandleReminderCancellation(c)
		case CallbackPrefixAdminReply:
			return bookingHandler.HandleAdminReplyRequest(c)
		case CallbackApproveDraft:
			return bookingHandler.HandleApproveDraft(c)
		case CallbackDiscardDraft:
			return bookingHandler.HandleDiscardDraft(c)
		case CallbackIgnore:
			return nil // Просто игнорируем кнопки-заглушки
		}
		return nil
	})

	// Обработчик для всех текстовых сообщений
	b.Handle(telebot.OnText, func(c telebot.Context) error {
		userID := c.Sender().ID
		text := strings.TrimSpace(c.Text())
		logging.Debugf("Received text: \"%s\" from user %d", text, userID)

		session := sessionStorage.Get(userID)
		view := SessionView{
			AdminReplyingTo:      sessionString(session, handlers.SessionKeyAdminReplyingTo),
			AwaitingConfirmation: sessionBool(session, handlers.SessionKeyAwaitingConfirmation),
			HasService:           sessionHasKey(session, handlers.SessionKeyService),
			HasName:              sessionHasKey(session, handlers.SessionKeyName),
		}

		switch RouteTextMessage(text, view) {
		case TextActionManualAppointment:
			return bookingHandler.HandleManualAppointment(c)
		case TextActionStart:
			return bookingHandler.HandleStart(c)
		case TextActionMyAppointments:
			return bookingHandler.HandleMyAppointments(c)
		case TextActionMyRecords:
			return bookingHandler.HandleMyRecords(c)
		case TextActionUpload:
			return bookingHandler.HandleUploadCommand(c)
		case TextActionAdminReply:
			return handleAdminReply(c, b, repo, sessionStorage, userID, text)
		case TextActionConfirmBooking:
			return bookingHandler.HandleConfirmBooking(c)
		case TextActionCancel:
			return bookingHandler.HandleCancel(c)
		case TextActionAskUseButtons:
			return c.Send("Пожалуйста, используйте кнопки под сообщением или напишите 'Да' для подтверждения.")
		case TextActionRestartFlow:
			sessionStorage.Set(userID, handlers.SessionKeyDate, nil)
			return bookingHandler.HandleStart(c) // Перезапускаем процесс, чтобы показать календарь
		case TextActionAskSelectService:
			return c.Send("Пожалуйста, выберите услугу, используя предложенные кнопки.")
		case TextActionNameInput:
			return bookingHandler.HandleNameInput(c)
		case TextActionForwardToAdmins:
			return forwardPatientMessageToAdmins(
				c, b, repo,
				bookingHandler.WebAppURL,
				bookingHandler.GenerateWebAppURL,
				finalAdminIDs, text,
			)
		}
		return nil
	})

	// Исправлен некорректный вывод имени бота при старте
	// Handle graceful shutdown
	go func() {
		<-ctx.Done()
		logging.Info("Stopping Telegram bot...")
		b.Stop()
	}()

	logging.Infof("Telegram bot started as @%s", b.Me.Username)
	b.Start()
}
