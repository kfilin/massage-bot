package keyboards

import (
	"time"

	"gopkg.in/telebot.v3"
)

func NewDatePicker() *telebot.ReplyMarkup {
	kb := &telebot.ReplyMarkup{}
	now := time.Now()

	// Month header
	headerRow := kb.Row(kb.Text(now.Format("January 2006")))

	// Weekdays
	days := []string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"}
	var dayRow []telebot.Btn
	for _, day := range days {
		dayRow = append(dayRow, kb.Text(day))
	}
	weekdayRow := kb.Row(dayRow...)

	kb.Reply(headerRow, weekdayRow)

	return kb
}
