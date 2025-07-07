package main

import (
	"context"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/kfilin/massage-bot/internal/appointment"
	"github.com/kfilin/massage-bot/internal/delivery/telegram"
	"github.com/kfilin/massage-bot/internal/infra/googlecalendar"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

func main() {
	_ = godotenv.Load()

	ctx := context.Background()

	// 1. Initialize Google Calendar Service
	calendarService, err := calendar.NewService(
		ctx,
		option.WithCredentialsFile(os.Getenv("GOOGLE_CREDENTIALS_FILE")),
	)
	if err != nil {
		log.Fatal("Failed to create calendar service: ", err)
	}

	// 2. Create our adapter
	calendarAdapter := googlecalendar.NewAdapter(calendarService)

	// 3. Create services
	apptService := appointment.New(calendarAdapter)

	// 4. Start bot
	bot, err := telegram.New(os.Getenv("TELEGRAM_TOKEN"), apptService)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Bot is running...")
	bot.Start()
}
