package telegram

import (
	"log"
	"strings"
	"time"

	"github.com/kfilin/massage-bot/internal/delivery/telegram/handlers"
	"github.com/kfilin/massage-bot/internal/domain"
	"github.com/kfilin/massage-bot/internal/ports" // Import ports for interfaces
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

	bookingHandler := handlers.NewBookingHandler(appointmentService, sessionStorage)

	b.Handle("/start", bookingHandler.HandleStart)
	b.Handle("/cancel", bookingHandler.HandleCancel)

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
		if strings.HasPrefix(trimmedData, "select_service|") {
			log.Printf("DEBUG: OnCallback: Matched 'select_service' prefix.")
			return bookingHandler.HandleServiceSelection(c)
		} else if strings.HasPrefix(trimmedData, "select_date|") || strings.HasPrefix(trimmedData, "navigate_month|") {
			log.Printf("DEBUG: OnCallback: Matched 'select_date' or 'navigate_month' prefix.")
			return bookingHandler.HandleDateSelection(c)
		} else if strings.HasPrefix(trimmedData, "select_time|") {
			log.Printf("DEBUG: OnCallback: Matched 'select_time' prefix.")
			return bookingHandler.HandleTimeSelection(c)
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
		session := sessionStorage.Get(userID)
		text := c.Text()
		log.Printf("Received text: \"%s\" from user %d", text, userID)

		// Проверяем, ожидает ли бот подтверждения
		if awaitingConfirmation, ok := session[handlers.SessionKeyAwaitingConfirmation].(bool); ok && awaitingConfirmation {
			log.Printf("DEBUG: OnText: Bot is awaiting confirmation for user %d.", userID)
			switch text {
			case "Подтвердить":
				log.Printf("DEBUG: OnText: Matched 'Подтвердить' for confirmation.")
				return bookingHandler.HandleConfirmBooking(c)
			case "Отменить запись":
				log.Printf("DEBUG: OnText: Matched 'Отменить запись' for cancellation.")
				return bookingHandler.HandleCancel(c)
			default:
				log.Printf("DEBUG: OnText: Invalid text input '%s' while awaiting confirmation for user %d.", text, userID)
				return c.Send("Пожалуйста, используйте кнопки 'Подтвердить' или 'Отменить запись'.")
			}
		}

		// Оригинальная логика для других текстовых вводов (имя и т.д.)
		switch text {
		case "Подтвердить": // Этот случай будет срабатывать только если SessionKeyAwaitingConfirmation = false (чего быть не должно)
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
