package ports

import (
	"io"

	"gopkg.in/telebot.v3"
)

// BotAPI is the subset of *telebot.Bot methods that handlers use.
// Defining it as an interface lets tests inject a mock implementation
// without needing a real Telegram connection.
//
// The real *telebot.Bot satisfies this interface implicitly — existing
// call sites that pass `b *telebot.Bot` (or `c.Bot()`) keep working.
//
// Lives in `internal/ports` so both `internal/delivery/telegram` and
// `internal/delivery/telegram/handlers` can depend on it without
// creating an import cycle.
type BotAPI interface {
	Send(to telebot.Recipient, what interface{}, opts ...interface{}) (*telebot.Message, error)
	Copy(to telebot.Recipient, msg telebot.Editable, opts ...interface{}) (*telebot.Message, error)
	Delete(msg telebot.Editable) error
	EditReplyMarkup(msg telebot.Editable, markup *telebot.ReplyMarkup) (*telebot.Message, error)
	File(file *telebot.File) (io.ReadCloser, error)
	Raw(method string, payload interface{}) ([]byte, error)
}
