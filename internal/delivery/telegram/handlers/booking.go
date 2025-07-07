package handlers

import (
	"fmt"

	"github.com/kfilin/massage-bot/internal/appointment"
	"gopkg.in/telebot.v3"
)

type BookingHandler struct {
	ApptService *appointment.Service
}

func NewBookingHandler(svc *appointment.Service) *BookingHandler {
	return &BookingHandler{ApptService: svc}
}

func (h *BookingHandler) HandleBookStart(c telebot.Context) error {
	services := h.ApptService.GetAvailableServices()

	var rows [][]telebot.ReplyButton
	for _, svc := range services {
		btn := telebot.ReplyButton{Text: fmt.Sprintf("%s - %.2f руб", svc.Name, svc.Price)}
		rows = append(rows, []telebot.ReplyButton{btn})
	}

	return c.Send("Choose a service:", &telebot.ReplyMarkup{
		ReplyKeyboard:   rows,
		ResizeKeyboard:  true,
		OneTimeKeyboard: true,
	})
}
