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

	"github.com/kfilin/massage-bot/internal/delivery/telegram/handlers"
	"github.com/kfilin/massage-bot/internal/domain"
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
	// Set menu button for quick TWA access
	if webAppURL != "" {
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
		} else {
			logging.Info("Menu button configured successfully")
		}
	}

	// Ensure Admin ID is in the allowed list for notifications
	// Use a map to deduplicate IDs ensuring no double notifications
	adminMap := make(map[string]bool)
	if adminTelegramID != "" {
		adminMap[adminTelegramID] = true
	}
	for _, id := range allowedTelegramIDs {
		if id != "" {
			adminMap[id] = true
		}
	}
	for _, id := range therapistIDs {
		if id != "" {
			adminMap[id] = true
		}
	}

	finalAdminIDs := make([]string, 0, len(adminMap))
	for id := range adminMap {
		finalAdminIDs = append(finalAdminIDs, id)
	}

	botPresenter := presentation.NewBotPresenter()
	bookingHandler := handlers.NewBookingHandler(appointmentService, sessionStorage, finalAdminIDs, therapistIDs, trans, repo, botPresenter, webAppURL, webAppSecret)

	// Initialize and start Reminder Service
	reminderService := reminder.NewService(appointmentService, repo, b, finalAdminIDs, botPresenter)
	reminderService.Start(context.Background())

	// Start Daily Backup Worker (Sent to Primary Admin)
	if adminTelegramID != "" {
		go func() {
			// Wait 5 minutes after startup to avoid congestion
			time.Sleep(5 * time.Minute)
			ticker := time.NewTicker(24 * time.Hour)
			defer ticker.Stop()

			for range ticker.C {
				logging.Infof("[BackupWorker] Starting scheduled daily backup for Admin %s...", adminTelegramID)
				zipPath, err := repo.CreateBackup()
				if err != nil {
					logging.Errorf("[BackupWorker] FAILED to create scheduled backup: %v", err)
					continue
				}

				adminIntID, err := strconv.ParseInt(adminTelegramID, 10, 64)
				if err == nil {
					doc := &telebot.Document{
						File:     telebot.FromDisk(zipPath),
						FileName: filepath.Base(zipPath),
						Caption:  fmt.Sprintf("💾 Ежедневная копия данных (%s)\n\n<i>Примечание: Локальный файл удален для экономии места.</i>", time.Now().Format("02.01.2006")),
					}
					_, err = b.Send(&telebot.User{ID: adminIntID}, doc, telebot.ModeHTML)
					if err != nil {
						logging.Errorf("[BackupWorker] FAILED to send scheduled backup: %v", err)
					}
				}
				// Cleanup temporary zip to save server disk space
				os.Remove(zipPath)
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

		// Добавляем логирование для каждой ветки if/else if
		// Используем trimmedData для проверки префикса
		if strings.HasPrefix(trimmedData, CallbackPrefixCategory) {
			logging.Debug("DEBUG: OnCallback: Matched 'select_category' prefix.")
			return bookingHandler.HandleCategorySelection(c)
		} else if strings.HasPrefix(trimmedData, CallbackPrefixService) {
			logging.Debug("DEBUG: OnCallback: Matched 'select_service' prefix.")
			return bookingHandler.HandleServiceSelection(c)
		} else if strings.HasPrefix(trimmedData, CallbackPrefixDate) || strings.HasPrefix(trimmedData, CallbackPrefixNavigateMonth) || trimmedData == CallbackBackToServices {
			logging.Debug("DEBUG: OnCallback: Matched 'select_date', 'navigate_month' or 'back_to_services'.")
			return bookingHandler.HandleDateSelection(c)
		} else if strings.HasPrefix(trimmedData, CallbackPrefixTime) || trimmedData == CallbackBackToDate {
			logging.Debug("DEBUG: OnCallback: Matched 'select_time' or 'back_to_date'.")
			return bookingHandler.HandleTimeSelection(c)
		} else if trimmedData == CallbackConfirmBooking {
			logging.Debug("DEBUG: OnCallback: Matched 'confirm_booking' data.")
			return bookingHandler.HandleConfirmBooking(c)
		} else if trimmedData == CallbackCancelBooking {
			logging.Debug("DEBUG: OnCallback: Matched 'cancel_booking' data.")
			return bookingHandler.HandleCancel(c)
		} else if strings.HasPrefix(trimmedData, CallbackPrefixCancelAppt) {
			logging.Debug("DEBUG: OnCallback: Matched 'cancel_appt' prefix.")
			return bookingHandler.HandleCancelAppointmentCallback(c)
		} else if strings.HasPrefix(trimmedData, CallbackPrefixConfirmReminder) {
			logging.Debug("DEBUG: OnCallback: Matched 'confirm_appt_reminder' prefix.")
			return bookingHandler.HandleReminderConfirmation(c)
		} else if strings.HasPrefix(trimmedData, CallbackPrefixCancelReminder) {
			logging.Debug("DEBUG: OnCallback: Matched 'cancel_appt_reminder' prefix.")
			return bookingHandler.HandleReminderCancellation(c)
		} else if strings.HasPrefix(trimmedData, CallbackPrefixAdminReply) {
			logging.Debug("DEBUG: OnCallback: Matched 'admin_reply' prefix.")
			return bookingHandler.HandleAdminReplyRequest(c)
		} else if strings.HasPrefix(trimmedData, CallbackApproveDraft) {
			logging.Debug("DEBUG: OnCallback: Matched 'approve_draft' prefix.")
			return bookingHandler.HandleApproveDraft(c)
		} else if strings.HasPrefix(trimmedData, CallbackDiscardDraft) {
			logging.Debug("DEBUG: OnCallback: Matched 'discard_draft' prefix.")
			return bookingHandler.HandleDiscardDraft(c)
		} else if trimmedData == CallbackIgnore {
			logging.Debug("DEBUG: OnCallback: Matched 'ignore' data.")
			return nil // Просто игнорируем кнопки-заглушки
		}

		logging.Warnf("DEBUG: OnCallback: No specific callback prefix matched for data: '%s'", trimmedData)
		return c.Send("Неизвестное действие с кнопкой. Пожалуйста, начните /start снова.")
	})

	// Обработчик для всех текстовых сообщений
	b.Handle(telebot.OnText, func(c telebot.Context) error {
		userID := c.Sender().ID
		text := strings.TrimSpace(c.Text())
		logging.Debugf("Received text: \"%s\" from user %d", text, userID)

		// Priority level 1: Commands fallback in OnText (helps if command handlers are bypassed)
		if strings.HasPrefix(text, "/create_appointment") {
			return bookingHandler.HandleManualAppointment(c)
		}

		// Priority level 2: Main menu buttons (always available)
		switch text {
		case "🗓 Записаться":
			return bookingHandler.HandleStart(c)
		case "📅 Мои записи":
			return bookingHandler.HandleMyAppointments(c)
		case "📄 Мед-карта":
			return bookingHandler.HandleMyRecords(c)
		case "📤 Загрузить документы":
			return bookingHandler.HandleUploadCommand(c)
		}

		session := sessionStorage.Get(userID)

		// Priority level 2: Admin Replying to Patient
		if replyingToID, ok := session[handlers.SessionKeyAdminReplyingTo].(string); ok && replyingToID != "" {
			logging.Debugf("DEBUG: OnText: Admin %d is replying to patient %s.", userID, replyingToID)

			patientID, _ := strconv.ParseInt(replyingToID, 10, 64)
			patientUser := &telebot.User{ID: patientID}

			// Send to patient
			replyMsg := fmt.Sprintf("📩 <b>Сообщение от Веры:</b>\n\n%s", text)
			_, err := b.Send(patientUser, replyMsg, telebot.ModeHTML)
			if err != nil {
				logging.Errorf("ERROR: Failed to deliver admin reply to patient %s: %v", replyingToID, err)
				return c.Send("❌ Не удалось доставить сообщение пациенту.")
			}

			// Log to Med-Card
			patient, err := repo.GetPatient(replyingToID)
			if err == nil {
				prefix := fmt.Sprintf("\n\n[👩‍⚕️ Вера %s]: ", time.Now().In(domain.ApptTimeZone).Format("02.01.2006 15:04"))
				patient.TherapistNotes += prefix + text
				if err := repo.SavePatient(patient); err != nil {
					logging.Errorf("Failed to save admin reply to patient record: %v", err)
				}
			}

			// Clear state
			sessionStorage.Set(userID, handlers.SessionKeyAdminReplyingTo, nil)
			return c.Send("✅ Сообщение доставлено и сохранено в мед-карте.")
		}

		// Priority level 3: Confirmation flow
		if awaitingConfirmation, ok := session[handlers.SessionKeyAwaitingConfirmation].(bool); ok && awaitingConfirmation {
			logging.Debugf("DEBUG: OnText: Bot is awaiting confirmation for user %d.", userID)
			cleanText := strings.ToLower(text)
			switch cleanText {
			case "подтвердить", "да", "д", "yes", "y", "ok", "ок":
				logging.Debugf("DEBUG: OnText: Matched confirmation text '%s' for user %d.", cleanText, userID)
				return bookingHandler.HandleConfirmBooking(c)
			case "отменить запись", "нет", "н", "no", "n", "отмена":
				logging.Debugf("DEBUG: OnText: Matched cancellation text '%s' for user %d.", cleanText, userID)
				return bookingHandler.HandleCancel(c)
			default:
				logging.Warnf("DEBUG: OnText: Invalid text input '%s' while awaiting confirmation for user %d.", text, userID)
				return c.Send("Пожалуйста, используйте кнопки под сообщением или напишите 'Да' для подтверждения.")
			}
		}

		// Priority level 4: Standard flow (Name input, etc.)
		switch text {
		case "Подтвердить": // Safety fallback
			logging.Debugf("DEBUG: OnText: Matched 'Подтвердить' (unexpectedly outside confirmation flow).")
			return bookingHandler.HandleConfirmBooking(c)
		case "Отменить запись":
			logging.Debugf("DEBUG: OnText: Matched 'Отменить запись'.")
			return bookingHandler.HandleCancel(c)
		case "Выбрать другую дату", "⬅️ Выбрать другую дату":
			logging.Debugf("DEBUG: OnText: Matched 'Выбрать другую дату'.")
			sessionStorage.Set(userID, handlers.SessionKeyDate, nil)
			return bookingHandler.HandleStart(c) // Перезапускаем процесс, чтобы показать календарь
		default:
			logging.Debugf(": OnText: Default case (assuming name input or initial service text).")
			if _, ok := session[handlers.SessionKeyService].(domain.Service); !ok {
				logging.Debugf("DEBUG: OnText: SessionKeyService not set. Asking to select service.")
				return c.Send("Пожалуйста, выберите услугу, используя предложенные кнопки.")
			} else if _, ok := session[handlers.SessionKeyName].(string); !ok {
				logging.Debugf("DEBUG: OnText: SessionKeyName not set. Assuming name input.")
				return bookingHandler.HandleNameInput(c)
			} else {
				logging.Debugf("DEBUG: OnText: All session data present, unknown text input. Forwarding to admins.")

				// Provide polite feedback to user
				if err := c.Send("Ваше сообщение получено и передано Вере."); err != nil {
					logging.Warnf("Failed to send confirmation to patient: %v", err)
				}

				// Forward to all admins
				telegramID := strconv.FormatInt(c.Sender().ID, 10)
				customerName := c.Sender().FirstName + " " + c.Sender().LastName
				if c.Sender().Username != "" {
					customerName += " (@" + c.Sender().Username + ")"
				}

				notification := fmt.Sprintf("📩 <b>Новое сообщение от пациента!</b>\n\n<b>Пациент:</b> %s (ID: %s)\n<b>Текст:</b> %s",
					customerName, telegramID, text)

				// Log to Med-Card automatically
				patient, err := repo.GetPatient(telegramID)
				if err == nil {
					prefix := fmt.Sprintf("\n\n[💬 Пациент %s]: ", time.Now().In(domain.ApptTimeZone).Format("02.01.2006 15:04"))
					patient.TherapistNotes += prefix + text
					if err := repo.SavePatient(patient); err != nil {
						logging.Errorf("Failed to save patient message: %v", err)
					}
				}

				// Add link to med-card and Reply button
				selector := &telebot.ReplyMarkup{}
				btnReply := selector.Data("✍️ Ответить", "admin_reply", telegramID)

				if bookingHandler.WebAppURL != "" {
					cardURL := bookingHandler.GenerateWebAppURL(telegramID)
					notification += fmt.Sprintf("\n\n📄 <a href=\"%s\">Открыть мед-карту</a>", cardURL)
				}
				selector.Inline(selector.Row(btnReply))

				for _, adminIDStr := range finalAdminIDs {
					adminID, _ := strconv.ParseInt(adminIDStr, 10, 64)
					// Helper to send notification with Reply button
					_, _ = b.Send(&telebot.User{ID: adminID}, notification, telebot.ModeHTML, selector)
				}

				return nil
			}
		}
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
