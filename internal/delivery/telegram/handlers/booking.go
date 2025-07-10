package handlers

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/kfilin/massage-bot/internal/domain"
	"github.com/kfilin/massage-bot/internal/ports"
	apnt_svc "github.com/kfilin/massage-bot/internal/services/appointment"
	"gopkg.in/telebot.v3"
)

type BookingHandler struct {
	appointmentService ports.AppointmentService
	sessionStorage     ports.SessionStorage
}

func NewBookingHandler(appointmentService ports.AppointmentService, sessionStorage ports.SessionStorage) *BookingHandler {
	return &BookingHandler{
		appointmentService: appointmentService,
		sessionStorage:     sessionStorage,
	}
}

// HandleStart handles the /start command, greeting the user and offering services.
func (h *BookingHandler) HandleStart(c telebot.Context) error {
	// Clear any previous session for the user
	h.sessionStorage.ClearSession(c.Sender().ID)

	services, err := h.appointmentService.GetAvailableServices(context.Background())
	if err != nil {
		log.Printf("Error getting available services: %v", err)
		return c.Send("К сожалению, произошла ошибка при получении списка услуг. Пожалуйста, попробуйте позже.")
	}

	if len(services) == 0 {
		return c.Send("К сожалению, в данный момент нет доступных услуг. Пожалуйста, попробуйте позже.")
	}

	// Build keyboard for services using telebot.ReplyButton
	var replyButtons [][]telebot.ReplyButton
	for _, service := range services {
		btn := telebot.ReplyButton{Text: fmt.Sprintf("%s (%d мин) - %.2f руб.", service.Name, service.DurationMinutes, service.Price)}
		replyButtons = append(replyButtons, []telebot.ReplyButton{btn}) // Each inner slice represents a row of buttons
	}

	// Create a ReplyMarkup with the prepared buttons
	replyMarkup := &telebot.ReplyMarkup{
		ReplyKeyboard:   replyButtons,
		ResizeKeyboard:  true,
		OneTimeKeyboard: true,
	}

	return c.Send("Привет! Я бот для записи на массаж. Выберите услугу:", replyMarkup)
}

// HandleServiceSelection processes the chosen service
func (h *BookingHandler) HandleServiceSelection(c telebot.Context) error {
	serviceName := c.Text()

	allServices, err := h.appointmentService.GetAvailableServices(context.Background())
	if err != nil {
		log.Printf("Error getting services for selection: %v", err)
		return c.Send("Произошла ошибка при обработке выбора услуги. Пожалуйста, попробуйте позже.")
	}

	var selectedService domain.Service
	found := false
	for _, s := range allServices {
		// Assuming the button text exactly matches service.Name for selection
		if fmt.Sprintf("%s (%d мин) - %.2f руб.", s.Name, s.DurationMinutes, s.Price) == serviceName {
			selectedService = s
			found = true
			break
		}
	}

	if !found {
		return c.Send("Неизвестная услуга. Пожалуйста, выберите услугу из списка, используя кнопки.")
	}

	h.sessionStorage.Set(c.Sender().ID, "service", selectedService)
	return c.Send("Отлично! Теперь введите желаемую дату записи в формате ДД.ММ.ГГГГ (например, 08.07.2025):")
}

// HandleDateSelection processes the chosen date
func (h *BookingHandler) HandleDateSelection(c telebot.Context) error {
	dateString := c.Text()
	parsedDate, err := time.Parse("02.01.2006", dateString)
	if err != nil {
		log.Printf("Error parsing date string %s: %v", dateString, err)
		return c.Send("Неверный формат даты. Пожалуйста, введите дату в формате ДД.ММ.ГГГГ (например, 08.07.2025).")
	}

	if parsedDate.Before(time.Now().Truncate(24 * time.Hour)) {
		return c.Send("Вы не можете выбрать дату в прошлом. Пожалуйста, выберите будущую дату.")
	}

	h.sessionStorage.Set(c.Sender().ID, "date", parsedDate)

	return h.sendAvailableTimeSlots(c)
}

// HandleTimeSelection processes the chosen time
func (h *BookingHandler) HandleTimeSelection(c telebot.Context) error {
	timeString := c.Text()

	_, err := time.Parse("15:04", timeString)
	if err != nil {
		return c.Send("Неверный формат времени. Пожалуйста, выберите время из списка или введите в формате ЧЧ:ММ (например, 10:30).")
	}

	h.sessionStorage.Set(c.Sender().ID, "time", timeString)

	return h.sendConfirmation(c)
}

// HandleConfirmBooking confirms and creates the appointment
func (h *BookingHandler) HandleConfirmBooking(c telebot.Context) error {
	userID := c.Sender().ID
	session := h.sessionStorage.Get(userID)

	selectedService, ok := session["service"].(domain.Service)
	if !ok {
		return c.Send("Ошибка: Услуга не выбрана или некорректна. Начните сначала командой /start")
	}

	dateVal, ok := session["date"].(time.Time)
	if !ok {
		return c.Send("Ошибка: Дата не выбрана. Начните сначала командой /start")
	}

	timeString, ok := session["time"].(string)
	if !ok {
		return c.Send("Ошибка: Время не выбрано. Начните сначала командой /start")
	}

	parsedTime, err := time.Parse("15:04", timeString)
	if err != nil {
		log.Printf("Error parsing time string %s: %v", timeString, err)
		return c.Send("Ошибка формата времени. Пожалуйста, попробуйте снова.")
	}

	appointmentTime := time.Date(dateVal.Year(), dateVal.Month(), dateVal.Day(), parsedTime.Hour(), parsedTime.Minute(), 0, 0, dateVal.Location())

	appointment := &domain.Appointment{
		ServiceID:    selectedService.ID,
		Service:      selectedService,
		Time:         appointmentTime,
		Duration:     selectedService.DurationMinutes,
		CustomerName: c.Sender().FirstName + " " + c.Sender().LastName,
		CustomerTgID: strconv.FormatInt(c.Sender().ID, 10),
		Notes:        fmt.Sprintf("Запись через Telegram-бот. Услуга: %s. Клиент: %s %s (ID: %s)", selectedService.Name, c.Sender().FirstName, c.Sender().LastName, strconv.FormatInt(c.Sender().ID, 10)),
	}

	createdApp, err := h.appointmentService.CreateAppointment(context.Background(), appointment)
	if err != nil {
		log.Printf("Error creating appointment: %v", err)
		switch err {
		case domain.ErrAppointmentInPast:
			return c.Send("Выбранное время уже в прошлом. Пожалуйста, выберите будущее время.")
		case domain.ErrOutsideWorkingHours:
			return c.Send("Выбранное время находится вне рабочих часов. Пожалуйста, выберите время с " + fmt.Sprintf("%d:00", apnt_svc.WorkStartHour) + " до " + fmt.Sprintf("%d:00", apnt_svc.WorkEndHour) + ".")
		case domain.ErrSlotUnavailable:
			return c.Send("Выбранное время уже занято. Пожалуйста, выберите другое время.")
		default:
			return c.Send(fmt.Sprintf("Не удалось создать запись: %v. Пожалуйста, попробуйте позже.", err))
		}
	}

	h.sessionStorage.ClearSession(userID)
	return c.Send(fmt.Sprintf("Ваша запись подтверждена!\nУслуга: %s\nДата и время: %s\nID записи: %s",
		createdApp.Service.Name, createdApp.Time.Format("02.01.2006 15:04"), createdApp.ID))
}

// sendAvailableTimeSlots fetches and sends available time slots for the chosen date and service
func (h *BookingHandler) sendAvailableTimeSlots(c telebot.Context) error {
	userID := c.Sender().ID
	session := h.sessionStorage.Get(userID)

	selectedService, ok := session["service"].(domain.Service)
	if !ok {
		return c.Send("Ошибка: Услуга не выбрана или некорректна. Начните сначала командой /start")
	}

	dateVal, ok := session["date"].(time.Time)
	if !ok {
		return c.Send("Ошибка: Дата не выбрана. Начните сначала командой /start")
	}

	availableSlots, err := h.appointmentService.GetAvailableTimeSlots(context.Background(), dateVal, selectedService.DurationMinutes)
	if err != nil {
		log.Printf("Error getting available time slots: %v", err)
		return c.Send("К сожалению, произошла ошибка при получении доступных временных слотов. Пожалуйста, попробуйте позже.")
	}

	if len(availableSlots) == 0 {
		return c.Send("На выбранную дату нет свободных слотов. Пожалуйста, выберите другую дату или попробуйте /start.")
	}

	// Build keyboard for time slots
	var replyButtons [][]telebot.ReplyButton
	for _, slot := range availableSlots {
		btn := telebot.ReplyButton{Text: slot.Start.Format("15:04")}
		replyButtons = append(replyButtons, []telebot.ReplyButton{btn})
	}

	replyMarkup := &telebot.ReplyMarkup{
		ReplyKeyboard:   replyButtons,
		ResizeKeyboard:  true,
		OneTimeKeyboard: true,
	}

	return c.Send("Выберите удобное время:", replyMarkup)
}

// sendConfirmation sends a confirmation message with chosen details
func (h *BookingHandler) sendConfirmation(c telebot.Context) error {
	userID := c.Sender().ID
	session := h.sessionStorage.Get(userID)

	selectedService, ok := session["service"].(domain.Service)
	if !ok {
		return c.Send("Ошибка: Услуга не выбрана. Начните сначала командой /start")
	}

	dateVal, ok := session["date"].(time.Time)
	if !ok {
		return c.Send("Ошибка: Дата не выбрана. Начните сначала командой /start")
	}

	timeString, ok := session["time"].(string)
	if !ok {
		return c.Send("Ошибка: Время не выбрано. Начните сначала командой /start")
	}

	// Prepare confirmation message
	confirmMsg := fmt.Sprintf("Вы выбрали:\nУслуга: %s\nДата: %s\nВремя: %s\n\nПодтвердите запись?",
		selectedService.Name, dateVal.Format("02.01.2006"), timeString)

	// Confirmation buttons
	confirmBtn := telebot.ReplyButton{Text: "Подтвердить"}
	cancelBtn := telebot.ReplyButton{Text: "Отменить запись"}

	confirmMarkup := &telebot.ReplyMarkup{
		ReplyKeyboard: [][]telebot.ReplyButton{
			{confirmBtn},
			{cancelBtn},
		},
		ResizeKeyboard:  true,
		OneTimeKeyboard: true,
	}

	return c.Send(confirmMsg, confirmMarkup)
}

// HandleCancel handles the cancellation of the current booking flow or an existing appointment.
func (h *BookingHandler) HandleCancel(c telebot.Context) error {
	userID := c.Sender().ID
	h.sessionStorage.ClearSession(userID)
	return c.Send("Запись отменена. Вы можете начать новую запись командой /start.", telebot.RemoveKeyboard)
}
