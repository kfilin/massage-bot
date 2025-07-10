package main

import (
	"github.com/kfilin/massage-bot/cmd/bot/config"
	"github.com/kfilin/massage-bot/internal/delivery/telegram"
)

func main() {
	cfg := config.LoadConfig()

	telegram.StartBot(cfg) // This function will start the bot
}
