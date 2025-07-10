package telegram

import (
	"log"
	"time"

	"github.com/kfilin/massage-bot/cmd/bot/config"
	"github.com/kfilin/massage-bot/internal/adapters/googlecalendar"
	"github.com/kfilin/massage-bot/internal/delivery/telegram/handlers"
	"github.com/kfilin/massage-bot/internal/domain"
	"github.com/kfilin/massage-bot/internal/services/appointment"
	"gopkg.in/telebot.v3"
)

// StartBot initializes and runs the Telegram bot.
func StartBot(cfg *config.Config) {
	pref := telebot.Settings{
		Token:  cfg.TgBotToken,
		Poller: &telebot.LongPoller{Timeout: 10 * time.Second},
	}

	b, err := telebot.NewBot(pref)
	if err != nil {
		log.Fatalf("Error creating bot: %v", err)
		return
	}

	// Initialize Google Calendar Client
	googleCalendarClient, err := googlecalendar.NewGoogleCalendarClient()
	if err != nil {
		log.Fatalf("Error initializing Google Calendar client: %v", err)
		return
	}

	// Initialize AppointmentRepository (Google Calendar Adapter implements this)
	appointmentRepo := googlecalendar.NewAdapter(googleCalendarClient)

	// Initialize AppointmentService (business logic)
	appointmentService := appointment.NewService(appointmentRepo)

	// Initialize SessionStorage (using the in-memory implementation for now)
	sessionStorage := NewInMemorySessionStorage() // Now correctly refers to the function in session.go

	// Initialize Handlers, injecting required services
	bookingHandler := handlers.NewBookingHandler(appointmentService, sessionStorage)

	// Register Handlers for commands and text messages
	b.Handle("/start", bookingHandler.HandleStart)
	b.Handle("/cancel", bookingHandler.HandleCancel)

	// Generic text message handler to manage booking flow state
	b.Handle(telebot.OnText, func(c telebot.Context) error {
		userID := c.Sender().ID
		session := sessionStorage.Get(userID)

		if _, ok := session["service"].(domain.Service); !ok {
			return bookingHandler.HandleServiceSelection(c)
		} else if _, ok := session["date"].(time.Time); !ok {
			return bookingHandler.HandleDateSelection(c)
		} else if _, ok := session["time"].(string); !ok {
			return bookingHandler.HandleTimeSelection(c)
		} else {
			text := c.Text()
			if text == "Подтвердить" {
				return bookingHandler.HandleConfirmBooking(c)
			} else if text == "Отменить запись" {
				return bookingHandler.HandleCancel(c)
			} else {
				return c.Send("Неизвестная команда или некорректный ввод. Вы можете начать заново командой /start.")
			}
		}
	})

	log.Println("Bot started!")
	b.Start()
}
