// internal/delivery/telegram/middleware.go
package telegram

import "gopkg.in/telebot.v3"

// List of admin user IDs
var adminUserIDs = map[int64]bool{
	123456789: true, // Replace with actual admin IDs
}

// AdminOnly middleware
func AdminOnly(next telebot.HandlerFunc) telebot.HandlerFunc {
	return func(c telebot.Context) error {
		if !adminUserIDs[c.Sender().ID] {
			return c.Send("🚫 You don't have permission to use this command")
		}
		return next(c)
	}
}
