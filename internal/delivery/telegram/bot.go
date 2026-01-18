package telegram

import (
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/kfilin/massage-bot/internal/delivery/telegram/handlers"
	"github.com/kfilin/massage-bot/internal/domain"
	"github.com/kfilin/massage-bot/internal/ports"   // Import ports for interfaces
	"github.com/kfilin/massage-bot/internal/storage" // Import storage pkg for ban check
	"gopkg.in/telebot.v3"
)

// StartBot initializes and runs the Telegram bot.
// It now receives all necessary services and configuration from the main package.
func StartBot(
	token string,
	appointmentService ports.AppointmentService,
	sessionStorage ports.SessionStorage,
	adminTelegramID string,
	allowedTelegramIDs []string,
	pdfGen ports.PDFGenerator,
	trans ports.TranscriptionService,
) {
	pref := telebot.Settings{
		Token:  token,
		Poller: &telebot.LongPoller{Timeout: 10 * time.Second},
	}

	b, err := telebot.NewBot(pref)
	if err != nil {
		log.Fatalf("Error creating bot: %v", err)
		return
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

	finalAdminIDs := make([]string, 0, len(adminMap))
	for id := range adminMap {
		finalAdminIDs = append(finalAdminIDs, id)
	}

	// Retrieve therapist ID from environment
	therapistID := os.Getenv("TG_THERAPIST_ID")
	if therapistID == "" {
		log.Println("WARNING: TG_THERAPIST_ID not set in environment.")
	}

	bookingHandler := handlers.NewBookingHandler(appointmentService, sessionStorage, finalAdminIDs, therapistID, pdfGen, trans)

	// GLOBAL MIDDLEWARE: Enforce ban check on ALL entry points
	b.Use(func(next telebot.HandlerFunc) telebot.HandlerFunc {
		return func(c telebot.Context) error {
			if c.Sender() == nil {
				return next(c)
			}
			telegramID := strconv.FormatInt(c.Sender().ID, 10)
			username := c.Sender().Username
			if banned, _ := storage.IsUserBanned(telegramID, username); banned {
				log.Printf("BLOCKED (Middleware): Banned user %s (@%s) tried to access bot.", telegramID, username)

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

	b.Handle("/start", bookingHandler.HandleStart)
	b.Handle("/cancel", bookingHandler.HandleCancel)
	b.Handle("/myrecords", bookingHandler.HandleMyRecords)
	b.Handle("/myappointments", bookingHandler.HandleMyAppointments)
	b.Handle("/upload", bookingHandler.HandleUploadCommand)
	b.Handle("/backup", bookingHandler.HandleBackup)
	b.Handle("/ban", bookingHandler.HandleBan)
	b.Handle("/ban", bookingHandler.HandleBan)
	b.Handle("/unban", bookingHandler.HandleUnban)
	b.Handle("/block", bookingHandler.HandleBlock)
	b.Handle("/status", bookingHandler.HandleStatus)

	// Register file/media handlers
	b.Handle(telebot.OnDocument, bookingHandler.HandleFileMessage)
	b.Handle(telebot.OnPhoto, bookingHandler.HandleFileMessage)
	b.Handle(telebot.OnVideo, bookingHandler.HandleFileMessage)
	b.Handle(telebot.OnAnimation, bookingHandler.HandleFileMessage)
	b.Handle(telebot.OnVoice, bookingHandler.HandleFileMessage)

	// Обработчик для всех inline-кнопок
	b.Handle(telebot.OnCallback, func(c telebot.Context) error {
		log.Printf("DEBUG: Entered OnCallback handler.")

		data := c.Callback().Data
		// Обрезаем пробелы в начале и конце строки данных колбэка
		trimmedData := strings.TrimSpace(data)
		log.Printf("Received callback: '%s' (trimmed: '%s') from user %d", data, trimmedData, c.Sender().ID)

		defer c.Respond() // Важно: Respond() должен быть вызван, чтобы убрать "часики" с кнопки

		// Добавляем логирование для каждой ветки if/else if
		// Используем trimmedData для проверки префикса
		if strings.HasPrefix(trimmedData, "select_category|") {
			log.Printf("DEBUG: OnCallback: Matched 'select_category' prefix.")
			return bookingHandler.HandleCategorySelection(c)
		} else if strings.HasPrefix(trimmedData, "select_service|") {
			log.Printf("DEBUG: OnCallback: Matched 'select_service' prefix.")
			return bookingHandler.HandleServiceSelection(c)
		} else if strings.HasPrefix(trimmedData, "select_date|") || strings.HasPrefix(trimmedData, "navigate_month|") {
			log.Printf("DEBUG: OnCallback: Matched 'select_date' or 'navigate_month' prefix.")
			return bookingHandler.HandleDateSelection(c)
		} else if strings.HasPrefix(trimmedData, "select_time|") {
			log.Printf("DEBUG: OnCallback: Matched 'select_time' prefix.")
			return bookingHandler.HandleTimeSelection(c)
		} else if trimmedData == "confirm_booking" {
			log.Printf("DEBUG: OnCallback: Matched 'confirm_booking' data.")
			return bookingHandler.HandleConfirmBooking(c)
		} else if trimmedData == "cancel_booking" {
			log.Printf("DEBUG: OnCallback: Matched 'cancel_booking' data.")
			return bookingHandler.HandleCancel(c)
		} else if strings.HasPrefix(trimmedData, "cancel_appt|") {
			log.Printf("DEBUG: OnCallback: Matched 'cancel_appt' prefix.")
			return bookingHandler.HandleCancelAppointmentCallback(c)
		} else if trimmedData == "download_record" {
			log.Printf("DEBUG: OnCallback: Matched 'download_record' data.")
			return bookingHandler.HandleDownloadRecord(c)
		} else if trimmedData == "ignore" {
			log.Printf("DEBUG: OnCallback: Matched 'ignore' data.")
			return nil // Просто игнорируем кнопки-заглушки
		}

		log.Printf("DEBUG: OnCallback: No specific callback prefix matched for data: '%s'", trimmedData)
		return c.Send("Неизвестное действие с кнопкой. Пожалуйста, начните /start снова.")
	})

	// Обработчик для всех текстовых сообщений
	b.Handle(telebot.OnText, func(c telebot.Context) error {
		userID := c.Sender().ID
		text := strings.TrimSpace(c.Text())
		log.Printf("Received text: \"%s\" from user %d", text, userID)

		// Priority level 1: Main menu buttons (always available)
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

		// Priority level 2: Confirmation flow
		if awaitingConfirmation, ok := session[handlers.SessionKeyAwaitingConfirmation].(bool); ok && awaitingConfirmation {
			log.Printf("DEBUG: OnText: Bot is awaiting confirmation for user %d.", userID)
			cleanText := strings.ToLower(text)
			switch cleanText {
			case "подтвердить", "да", "д", "yes", "y", "ok", "ок":
				log.Printf("DEBUG: OnText: Matched confirmation text '%s' for user %d.", cleanText, userID)
				return bookingHandler.HandleConfirmBooking(c)
			case "отменить запись", "нет", "н", "no", "n", "отмена":
				log.Printf("DEBUG: OnText: Matched cancellation text '%s' for user %d.", cleanText, userID)
				return bookingHandler.HandleCancel(c)
			default:
				log.Printf("DEBUG: OnText: Invalid text input '%s' while awaiting confirmation for user %d.", text, userID)
				return c.Send("Пожалуйста, используйте кнопки под сообщением или напишите 'Да' для подтверждения.")
			}
		}

		// Priority level 3: Standard flow (Name input, etc.)
		switch text {
		case "Подтвердить": // Safety fallback
			log.Printf("DEBUG: OnText: Matched 'Подтвердить' (unexpectedly outside confirmation flow).")
			return bookingHandler.HandleConfirmBooking(c)
		case "Отменить запись":
			log.Printf("DEBUG: OnText: Matched 'Отменить запись'.")
			return bookingHandler.HandleCancel(c)
		case "Выбрать другую дату", "⬅️ Выбрать другую дату":
			log.Printf("DEBUG: OnText: Matched 'Выбрать другую дату'.")
			sessionStorage.Set(userID, handlers.SessionKeyDate, nil)
			return bookingHandler.HandleStart(c) // Перезапускаем процесс, чтобы показать календарь
		default:
			log.Printf("DEBUG: OnText: Default case (assuming name input or initial service text).")
			if _, ok := session[handlers.SessionKeyService].(domain.Service); !ok {
				log.Printf("DEBUG: OnText: SessionKeyService not set. Asking to select service.")
				return c.Send("Пожалуйста, выберите услугу, используя предложенные кнопки.")
			} else if _, ok := session[handlers.SessionKeyName].(string); !ok { // Только запрашиваем имя, если оно еще не установлено
				log.Printf("DEBUG: OnText: SessionKeyName not set. Assuming name input.")
				return bookingHandler.HandleNameInput(c)
			} else {
				log.Printf("DEBUG: OnText: All session data present, unknown text input.")
				return c.Send("Неизвестная команда или некорректный ввод. Вы можете начать заново командой /start.")
			}
		}
	})

	// Исправлен некорректный вывод имени бота при старте
	log.Printf("Telegram bot started as @%s", b.Me.Username)
	b.Start()
}
