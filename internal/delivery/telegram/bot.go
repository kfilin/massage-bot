package telegram

import (
	"github.com/kfilin/massage-bot/internal/appointment"
	"github.com/kfilin/massage-bot/internal/delivery/telegram/handlers" // Add this import
	"gopkg.in/telebot.v3"
)

type Bot struct {
	*telebot.Bot
	apptService *appointment.Service
}

func New(token string, apptService *appointment.Service) (*Bot, error) {
	b, err := telebot.NewBot(telebot.Settings{Token: token})
	if err != nil {
		return nil, err
	}

	bot := &Bot{
		Bot:         b,
		apptService: apptService,
	}

	bot.registerHandlers()
	return bot, nil
}

func (b *Bot) registerHandlers() {
	// Common
	b.Handle("/start", b.handleStart)

	// Booking
	bookingHandler := handlers.NewBookingHandler(b.apptService) // Use handlers.NewBookingHandler
	b.Handle("/book", bookingHandler.HandleBookStart)
}

func (b *Bot) handleStart(c telebot.Context) error {
	return c.Send("Welcome! Use /book to schedule an appointment.")
}
