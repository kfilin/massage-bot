package handlers

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time" // Ensure time is imported

	"github.com/kfilin/massage-bot/internal/domain"
	"github.com/kfilin/massage-bot/internal/monitoring"
	"github.com/kfilin/massage-bot/internal/ports"   // Alias to avoid conflict with package name "appointment"
	"github.com/kfilin/massage-bot/internal/storage" // Import storage package
	"gopkg.in/telebot.v3"                            // Ensure telebot.v3 is correctly imported
)

// Session keys for storing booking state
const (
	SessionKeyService              = "service"
	SessionKeyDate                 = "date"
	SessionKeyTime                 = "time"
	SessionKeyName                 = "name"
	SessionKeyAwaitingConfirmation = "awaiting_confirmation" // NEW: Key to indicate awaiting confirmation
)

// BookingHandler handles booking-related commands and callbacks.
type BookingHandler struct {
	appointmentService ports.AppointmentService
	sessionStorage     ports.SessionStorage
}

// NewBookingHandler creates a new BookingHandler.
func NewBookingHandler(appointmentService ports.AppointmentService, sessionStorage ports.SessionStorage) *BookingHandler {
	return &BookingHandler{
		appointmentService: appointmentService,
		sessionStorage:     sessionStorage,
	}
}

// HandleStart handles the /start command, greeting the user and offering services.
func (h *BookingHandler) HandleStart(c telebot.Context) error {
	log.Printf("DEBUG: Entered HandleStart for user %d", c.Sender().ID)
	// Clear any previous session for the user
	h.sessionStorage.ClearSession(c.Sender().ID)

	services, err := h.appointmentService.GetAvailableServices(context.Background())
	if err != nil {
		log.Printf("Error getting available services: %v", err)
		return c.Send("Произошла ошибка при получении списка услуг. Пожалуйста, попробуйте позже.")
	}

	if len(services) == 0 {
		return c.Send("В настоящее время услуги недоступны. Пожалуйста, попробуйте позже.")
	}

	selector := &telebot.ReplyMarkup{}
	var rows []telebot.Row
	for _, svc := range services {
		label := fmt.Sprintf("%s - %.0f ₺", svc.Name, svc.Price)
		if svc.Description != "" {
			label = fmt.Sprintf("%s (%s)", label, svc.Description)
		}
		rows = append(rows, selector.Row(selector.Data(label, "select_service", svc.ID)))
	}
	selector.Inline(rows...)
	return c.Send("Привет! Это VERA BOT 💆✨\nВыберите услугу для записи:", selector)
}

// HandleServiceSelection handles the callback query for service selection.
func (h *BookingHandler) HandleServiceSelection(c telebot.Context) error {
	log.Printf("DEBUG: Entered HandleServiceSelection for user %d. Callback Data: '%s'", c.Sender().ID, c.Callback().Data)

	// Callback data is "select_service|SERVICE_ID". We need to split it.
	data := strings.TrimSpace(c.Callback().Data) // Trim spaces just in case
	parts := strings.Split(data, "|")

	log.Printf("DEBUG: HandleServiceSelection - Parsed parts: %v (length: %d)", parts, len(parts))

	if len(parts) != 2 || parts[0] != "select_service" {
		log.Printf("ERROR: HandleServiceSelection - Malformed service selection callback data. Expected 'select_service|ID', got: '%s'", data)
		return c.Edit("Некорректный выбор услуги. Пожалуйста, попробуйте /start снова.")
	}
	serviceID := parts[1]
	log.Printf("DEBUG: HandleServiceSelection - Extracted serviceID: '%s'", serviceID)

	userID := c.Sender().ID

	services, err := h.appointmentService.GetAvailableServices(context.Background())
	if err != nil {
		log.Printf("Error getting services in HandleServiceSelection: %v", err)
		return c.Edit("Произошла ошибка при получении списка услуг. Пожалуйста, попробуйте /start снова.")
	}

	var chosenService domain.Service
	found := false
	for _, svc := range services {
		if svc.ID == serviceID { // Match by ID
			chosenService = svc
			found = true
			break
		}
	}

	if !found {
		log.Printf("ERROR: Service with ID '%s' not found in available services for user %d", serviceID, userID)
		return c.Edit("Выбранная услуга не найдена. Пожалуйста, выберите из предложенных.")
	}

	h.sessionStorage.Set(userID, SessionKeyService, chosenService)
	log.Printf("DEBUG: Service selected and stored in session for user %d: %s (ID: %s)", userID, chosenService.Name, chosenService.ID)

	// Ask for date
	return h.askForDate(c, chosenService.Name)
}

// askForDate sends a calendar to the user for date selection.
func (h *BookingHandler) askForDate(c telebot.Context, serviceName string) error {
	log.Printf("DEBUG: Entered askForDate for user %d. Service: %s", c.Sender().ID, serviceName)

	now := time.Now()
	year, month, _ := now.Date()
	// Use domain.ApptTimeZone for consistency across the application
	currentMonth := time.Date(year, month, 1, 0, 0, 0, 0, domain.ApptTimeZone)

	calendarKeyboard := generateCalendar(currentMonth)

	return c.EditOrSend(
		fmt.Sprintf("Отлично, услуга '%s' выбрана. Теперь выберите дату:", serviceName),
		calendarKeyboard,
	)
}

// generateCalendar creates an inline keyboard for month navigation and date selection.
func generateCalendar(month time.Time) *telebot.ReplyMarkup {
	log.Printf("DEBUG: Generating calendar for month: %s", month.Format("2006-01"))
	selector := &telebot.ReplyMarkup{}
	var rows []telebot.Row

	// Navigation row
	prevMonth := month.AddDate(0, -1, 0)
	nextMonth := month.AddDate(0, 1, 0)
	rows = append(rows, selector.Row(
		selector.Data("⬅️", "navigate_month", prevMonth.Format("2006-01")),
		// Используем "January" для форматирования месяца, чтобы Go перевел его
		selector.Data(month.Format("January 2006"), "ignore"), // Current month, no action
		selector.Data("➡️", "navigate_month", nextMonth.Format("2006-01")),
	))

	// Weekday headers
	weekdays := selector.Row(
		selector.Data("Пн", "ignore"),
		selector.Data("Вт", "ignore"),
		selector.Data("Ср", "ignore"),
		selector.Data("Чт", "ignore"),
		selector.Data("Пт", "ignore"),
		selector.Data("Сб", "ignore"),
		selector.Data("Вс", "ignore"),
	)
	rows = append(rows, weekdays)

	// Dates
	firstDayOfMonth := month
	// Adjust to Monday
	offset := (int(firstDayOfMonth.Weekday()) + 6) % 7 // Monday = 0, Sunday = 6
	startDay := firstDayOfMonth.AddDate(0, 0, -offset)

	for week := 0; week < 6; week++ { // Max 6 weeks for a month
		var weekBtns []telebot.Btn
		for day := 0; day < 7; day++ {
			currentDay := startDay.AddDate(0, 0, week*7+day)
			// Check if the current day is not in the past
			// Using domain.ApptTimeZone for consistency
			loc := domain.ApptTimeZone
			if loc == nil {
				log.Println("WARNING: domain.ApptTimeZone is nil during calendar generation, defaulting to Local time.")
				loc = time.Local
			}
			nowInLoc := time.Now().In(loc).Truncate(24 * time.Hour) // Truncate to start of day in local time

			if currentDay.Month() != month.Month() {
				// Empty button for days outside the current month
				weekBtns = append(weekBtns, selector.Data(" ", "ignore"))
			} else if currentDay.Truncate(24 * time.Hour).Before(nowInLoc) { // Disable past dates
				weekBtns = append(weekBtns, selector.Data(fmt.Sprintf("%d", currentDay.Day()), "ignore"))
			} else {
				// Callback data format: "select_date|YYYY-MM-DD"
				weekBtns = append(weekBtns, selector.Data(fmt.Sprintf("%d", currentDay.Day()), "select_date", currentDay.Format("2006-01-02")))
			}
		}
		rows = append(rows, selector.Row(weekBtns...))
	}

	selector.Inline(rows...)
	return selector
}

// HandleDateSelection handles the callback query for date selection or month navigation.
func (h *BookingHandler) HandleDateSelection(c telebot.Context) error {
	log.Printf("DEBUG: Entered HandleDateSelection for user %d. Callback Data: '%s'", c.Sender().ID, c.Callback().Data)

	data := strings.TrimSpace(c.Callback().Data) // Trim spaces
	userID := c.Sender().ID

	if strings.HasPrefix(data, "navigate_month|") {
		parts := strings.Split(data, "|")
		if len(parts) != 2 || parts[0] != "navigate_month" {
			log.Printf("ERROR: Malformed month navigation callback data: %s", data)
			return c.Edit("Некорректная навигация. Попробуйте снова.")
		}
		monthStr := parts[1]
		selectedMonth, err := time.Parse("2006-01", monthStr)
		if err != nil {
			log.Printf("ERROR: Invalid month format in navigation: %s, error: %v", monthStr, err)
			return c.Edit("Некорректная дата. Попробуйте снова.")
		}
		calendarKeyboard := generateCalendar(selectedMonth)
		return c.Edit(c.Message().Text, calendarKeyboard) // Edit the existing message
	} else if strings.HasPrefix(data, "select_date|") {
		parts := strings.Split(data, "|")
		if len(parts) != 2 || parts[0] != "select_date" {
			log.Printf("ERROR: Malformed date selection callback data: %s", data)
			return c.Edit("Некорректный выбор даты. Попробуйте /start снова.")
		}
		dateStr := parts[1]
		selectedDate, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			log.Printf("ERROR: Invalid date format in selection: %s, error: %v", dateStr, err)
			return c.Edit("Некорректная дата. Попробуйте /start снова.")
		}

		h.sessionStorage.Set(userID, SessionKeyDate, selectedDate)
		log.Printf("DEBUG: Date selected and stored in session for user %d: %s", userID, selectedDate.Format("2006-01-02"))

		// Now ask for time
		return h.askForTime(c)
	}
	return c.Send("Неизвестное действие с датой. Пожалуйста, попробуйте /start снова.")
}

// askForTime sends available time slots to the user.
func (h *BookingHandler) askForTime(c telebot.Context) error {
	log.Printf("DEBUG: Entered askForTime for user %d", c.Sender().ID)
	userID := c.Sender().ID
	sessionData := h.sessionStorage.Get(userID)

	service, okS := sessionData[SessionKeyService].(domain.Service)
	date, okD := sessionData[SessionKeyDate].(time.Time)

	if !okS || !okD {
		log.Printf("ERROR: Missing session data for time selection for user %d. Service OK: %t, Date OK: %t", userID, okS, okD)
		h.sessionStorage.ClearSession(userID)
		return c.Send("⚠️ Сессия истекла из-за перезагрузки бота.\nПожалуйста, начните заново командой /start", telebot.RemoveKeyboard)
	}

	// Make sure the selected date is at the beginning of the day in the correct timezone
	loc := domain.ApptTimeZone
	if loc == nil {
		log.Println("WARNING: domain.ApptTimeZone is nil, defaulting to Local time.")
		loc = time.Local
	}
	selectedDateInLoc := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, loc)

	log.Printf("DEBUG: Calling GetAvailableTimeSlots for user %d with date %s and duration %d", userID, selectedDateInLoc.Format("2006-01-02"), service.DurationMinutes)
	timeSlots, err := h.appointmentService.GetAvailableTimeSlots(context.Background(), selectedDateInLoc, service.DurationMinutes)
	if err != nil {
		log.Printf("ERROR: Error getting available time slots for user %d: %v", userID, err)
		// Clean up the calendar keyboard before showing the error
		if c.Message() != nil {
			c.Bot().EditReplyMarkup(c.Message(), nil)
		}
		return c.Send("❌ Ошибка при получении слотов: " + err.Error() + "\n\nПожалуйста, начните заново: /start")
	}
	log.Printf("DEBUG: Received %d time slots for user %d.", len(timeSlots), userID)

	if len(timeSlots) == 0 {
		// Используем c.EditOrSend для обновления сообщения, если слотов нет
		return c.EditOrSend("На эту дату нет доступных временных слотов. Пожалуйста, выберите другую дату.", &telebot.ReplyMarkup{
			ReplyKeyboard: [][]telebot.ReplyButton{
				{{Text: "⬅️ Выбрать другую дату"}},
			},
			ResizeKeyboard:  true,
			OneTimeKeyboard: true,
		})
	}

	selector := &telebot.ReplyMarkup{}
	var rows []telebot.Row
	for _, slot := range timeSlots {
		// Callback data format: "select_time|HH:MM"
		rows = append(rows, selector.Row(
			selector.Data(slot.Start.Format("15:04"), "select_time", slot.Start.Format("15:04")),
		))
	}
	selector.Inline(rows...)

	// Создаем ReplyKeyboard для кнопки "Выбрать другую дату"
	replyKeyboard := &telebot.ReplyMarkup{
		ReplyKeyboard: [][]telebot.ReplyButton{
			{{Text: "⬅️ Выбрать другую дату"}},
		},
		ResizeKeyboard:  true,
		OneTimeKeyboard: true,
	}

	// Редактируем предыдущее сообщение (календарь) с новой инлайн-клавиатурой (слоты времени)
	err = c.Edit(
		fmt.Sprintf("Отлично, доступны следующие временные слоты для '%s' %s:", service.Name, date.Format("02.01.2006")),
		selector, // Inline keyboard for time slots
	)
	if err != nil {
		log.Printf("ERROR: Failed to edit message with time slots: %v", err)
		// Если не удалось отредактировать (например, сообщение слишком старое), отправляем новое.
		// В этом случае ReplyKeyboard также будет в этом сообщении.
		return c.Send(
			fmt.Sprintf("Отлично, доступны следующие временные слоты для '%s' %s:", service.Name, date.Format("02.01.2006")),
			selector,
			&telebot.SendOptions{ReplyMarkup: replyKeyboard}, // Reply keyboard as SendOption for new message
		)
	}

	// Если редактирование прошло успешно, отправляем ReplyKeyboard отдельным сообщением.
	// Это важно, чтобы ReplyKeyboard появилась под полем ввода, а не как часть InlineKeyboard.
	return c.Send("Или выберите другую дату:", replyKeyboard)
}

// HandleTimeSelection handles the callback query for time slot selection.
func (h *BookingHandler) HandleTimeSelection(c telebot.Context) error {
	log.Printf("DEBUG: Entered HandleTimeSelection for user %d. Callback Data: '%s'", c.Sender().ID, c.Callback().Data)

	data := strings.TrimSpace(c.Callback().Data) // Trim spaces
	userID := c.Sender().ID

	parts := strings.Split(data, "|")
	if len(parts) != 2 || parts[0] != "select_time" {
		log.Printf("ERROR: Malformed time selection callback data: %s", data)
		return c.Edit("Некорректный выбор времени. Пожалуйста, попробуйте /start снова.")
	}
	timeStr := parts[1] // e.g., "15:04"

	// Validate time format. We expect "HH:MM"
	_, err := time.Parse("15:04", timeStr)
	if err != nil {
		log.Printf("ERROR: Invalid time format in selection: %s, error: %v", timeStr, err)
		return c.Edit("Некорректное время. Пожалуйста, попробуйте /start снова.")
	}
	h.sessionStorage.Set(userID, SessionKeyTime, timeStr)
	log.Printf("DEBUG: Time selected and stored in session for user %d: %s", userID, timeStr)

	// Удаляем инлайн-клавиатуру со слотами времени из предыдущего сообщения
	if c.Message() != nil {
		_, err := c.Bot().EditReplyMarkup(c.Message(), nil) // Pass nil to remove inline keyboard
		if err != nil {
			log.Printf("WARNING: Failed to remove inline keyboard from message %d: %v", c.Message().ID, err)
		}
	}

	// Теперь переходим к запросу имени.
	// Используем c.Send для отправки нового сообщения и удаления ReplyKeyboard
	return c.Send("Пожалуйста, введите ваше имя и фамилию для записи (например, Иван Иванов).", telebot.RemoveKeyboard)
}

// HandleNameInput handles the user's name input (regular text message).
func (h *BookingHandler) HandleNameInput(c telebot.Context) error {
	log.Printf("DEBUG: Entered HandleNameInput for user %d. Text: '%s'", c.Sender().ID, c.Text())

	userID := c.Sender().ID
	userName := strings.TrimSpace(c.Text())

	if userName == "" {
		return c.Send("Имя не может быть пустым. Пожалуйста, введите ваше имя и фамилию.")
	}

	h.sessionStorage.Set(userID, SessionKeyName, userName)
	log.Printf("DEBUG: Name stored in session for user %d: %s", userID, userName)

	// All data collected, ask for confirmation
	return h.askForConfirmation(c)
}

// askForConfirmation asks the user to confirm the booking details.
func (h *BookingHandler) askForConfirmation(c telebot.Context) error {
	log.Printf("DEBUG: Entered askForConfirmation for user %d", c.Sender().ID)

	userID := c.Sender().ID
	sessionData := h.sessionStorage.Get(userID)

	service, okS := sessionData[SessionKeyService].(domain.Service)
	date, okD := sessionData[SessionKeyDate].(time.Time)
	timeStr, okT := sessionData[SessionKeyTime].(string)
	name, okN := sessionData[SessionKeyName].(string)

	if !okS || !okD || !okT || !okN {
		log.Printf("ERROR: Missing session data for confirmation for user %d: service=%t, date=%t, time=%t, name=%t", userID, okS, okD, okT, okN)
		h.sessionStorage.ClearSession(userID)
		return c.Send("Ошибка сессии. Пожалуйста, начните /start снова.", telebot.RemoveKeyboard)
	}

	// Combine date and time string into a time.Time object for display
	appointmentTime, err := time.Parse("2006-01-02 15:04", fmt.Sprintf("%s %s", date.Format("2006-01-02"), timeStr))
	if err != nil {
		log.Printf("ERROR: Failed to parse appointment time for user %d: %v", userID, err)
		h.sessionStorage.ClearSession(userID)
		return c.Send("Ошибка форматирования времени. Пожалуйста, начните /start снова.", telebot.RemoveKeyboard)
	}

	confirmMessage := fmt.Sprintf(
		"<b>Пожалуйста, подтвердите вашу запись:</b>\n\n"+
			"Услуга: <b>%s</b>\n"+
			"Цена: <b>%.0f ₺</b>\n"+
			"Дата: <b>%s</b>\n"+
			"Время: <b>%s</b>\n"+
			"Имя: <b>%s</b>\n\n"+
			"Всё верно?",
		service.Name,
		service.Price,
		appointmentTime.Format("02.01.2006"),
		appointmentTime.Format("15:04"),
		name,
	)

	// Inline Keyboard - One button per row for maximum prominence
	selector := &telebot.ReplyMarkup{}
	selector.Inline(
		selector.Row(selector.Data("✅ ПОДТВЕРДИТЬ", "confirm_booking")),
		selector.Row(selector.Data("❌ ОТМЕНИТЬ", "cancel_booking")),
	)

	// Set session flag indicating awaiting confirmation (keep for fallback/cleanup)
	h.sessionStorage.Set(userID, SessionKeyAwaitingConfirmation, true)
	log.Printf("DEBUG: Set SessionKeyAwaitingConfirmation for user %d to true.", userID)

	return c.Send(confirmMessage, selector, telebot.ModeHTML)
}

// HandleConfirmBooking handles the confirmation of a booking.
func (h *BookingHandler) HandleConfirmBooking(c telebot.Context) error {
	log.Printf("DEBUG: Entered HandleConfirmBooking for user %d", c.Sender().ID)

	userID := c.Sender().ID
	sessionData := h.sessionStorage.Get(userID)

	// Clear awaiting confirmation flag
	h.sessionStorage.Set(userID, SessionKeyAwaitingConfirmation, false)
	log.Printf("DEBUG: Cleared SessionKeyAwaitingConfirmation for user %d.", userID)

	service, okS := sessionData[SessionKeyService].(domain.Service)
	date, okD := sessionData[SessionKeyDate].(time.Time)
	timeStr, okT := sessionData[SessionKeyTime].(string)
	name, okN := sessionData[SessionKeyName].(string)

	if !okS || !okD || !okT || !okN {
		log.Printf("Session data missing for user %d during confirmation: service=%t, date=%t, time=%t, name=%t", userID, okS, okD, okT, okN)
		h.sessionStorage.ClearSession(userID)
		return c.Send("Ошибка сессии. Пожалуйста, начните /start снова.", telebot.RemoveKeyboard)
	}

	// Combine date and time string into a time.Time object
	appointmentTime, err := time.Parse("2006-01-02 15:04", fmt.Sprintf("%s %s", date.Format("2006-01-02"), timeStr))
	if err != nil {
		log.Printf("Failed to parse appointment time for user %d during confirmation: %v", userID, err)
		h.sessionStorage.ClearSession(userID)
		return c.Send("Ошибка форматирования времени. Пожалуйста, начните /start снова.", telebot.RemoveKeyboard)
	}

	// Adjust appointmentTime to the correct timezone (e.g., Europe/Istanbul)
	loc := domain.ApptTimeZone
	if loc == nil {
		log.Println("WARNING: domain.ApptTimeZone is nil during appointment creation, defaulting to Local time.")
		loc = time.Local
	}
	appointmentTime = time.Date(appointmentTime.Year(), appointmentTime.Month(), appointmentTime.Day(),
		appointmentTime.Hour(), appointmentTime.Minute(), 0, 0, loc)

	// Create the Appointment object
	appointment := &domain.Appointment{
		Service:      service,
		StartTime:    appointmentTime,
		EndTime:      appointmentTime.Add(time.Duration(service.DurationMinutes) * time.Minute),
		Duration:     service.DurationMinutes,
		CustomerName: name,
		CustomerTgID: strconv.FormatInt(userID, 10), // Store Telegram User ID as string
	}

	// Call the appointment service to create the appointment
	_, err = h.appointmentService.CreateAppointment(context.Background(), appointment)
	if err != nil {
		log.Printf("Error creating appointment for user %d: %v", userID, err)
		// Handle specific errors from the service layer
		switch {
		case errors.Is(err, domain.ErrSlotUnavailable):
			return c.Send("К сожалению, выбранное время уже занято. Пожалуйста, выберите другой слот.", telebot.RemoveKeyboard)
		case errors.Is(err, domain.ErrAppointmentInPast):
			return c.Send("Выбранное время уже в прошлом. Пожалуйста, выберите будущее время.", telebot.RemoveKeyboard)
		case errors.Is(err, domain.ErrOutsideWorkingHours):
			return c.Send("Выбранное время выходит за рамки рабочего дня. Пожалуйста, выберите другое время.", telebot.RemoveKeyboard)
		case errors.Is(err, domain.ErrInvalidDuration):
			return c.Send("Некорректная длительность услуги. Пожалуйста, свяжитесь с администратором.", telebot.RemoveKeyboard)
		case errors.Is(err, domain.ErrInvalidAppointment):
			return c.Send("Некорректные данные для записи. Пожалуйста, попробуйте сначала.", telebot.RemoveKeyboard)
		default:
			return c.Send("Произошла ошибка при создании записи. Пожалуйста, попробуйте позже.", telebot.RemoveKeyboard)
		}
	}

	// Save patient record
	patient := domain.Patient{
		TelegramID:     strconv.FormatInt(userID, 10),
		Name:           name,
		FirstVisit:     time.Now(),
		LastVisit:      time.Now(),
		TotalVisits:    1,
		HealthStatus:   "initial",
		CurrentService: service.Name,
		TherapistNotes: fmt.Sprintf("Первая запись: %s на %s",
			service.Name,
			appointmentTime.Format("02.01.2006 15:04")),
	}

	if err := storage.SavePatient(patient); err != nil {
		log.Printf("WARNING: Failed to save patient record for user %d: %v", userID, err)
		// Don't fail the booking, just log the error
	} else {
		log.Printf("Patient record saved for user %d", userID)
	}

	// Increment booking metric
	monitoring.IncrementBooking(service.Name)

	// Clear session on successful booking
	h.sessionStorage.ClearSession(userID)

	// Add button to download the record
	selector := &telebot.ReplyMarkup{}
	selector.Inline(
		selector.Row(selector.Data("📄 СКАЧАТЬ МЕД-КАРТУ", "download_record")),
	)

	return c.Send(fmt.Sprintf("Ваша запись на услугу '%s' на %s в %s успешно подтверждена! Ждем вас.\n\nВы можете скачать вашу медицинскую карту ниже:",
		service.Name, appointmentTime.Format("02.01.2006"), appointmentTime.Format("15:04")), selector, telebot.RemoveKeyboard)
}

// HandleCancel handles the "Отменить запись" (Cancel booking) button
func (h *BookingHandler) HandleCancel(c telebot.Context) error {
	log.Printf("DEBUG: Entered HandleCancel for user %d", c.Sender().ID)
	userID := c.Sender().ID
	// Clear awaiting confirmation flag
	h.sessionStorage.Set(userID, SessionKeyAwaitingConfirmation, false)
	log.Printf("DEBUG: Cleared SessionKeyAwaitingConfirmation for user %d (via cancel).", userID)

	h.sessionStorage.ClearSession(userID)
	// Remove keyboard and send confirmation
	return c.Send("Запись отменена. Сессия очищена. Вы можете начать /start снова.", telebot.RemoveKeyboard)
}

// HandleMyRecords shows patient their records summary
func (h *BookingHandler) HandleMyRecords(c telebot.Context) error {
	userID := c.Sender().ID
	telegramID := strconv.FormatInt(userID, 10)

	patient, err := storage.GetPatient(telegramID)
	if err != nil {
		return c.Send(`📝 У вас еще нет медицинской карты.

После первой записи на массаж, ваша карта будет создана автоматически.

Запишитесь через /start чтобы начать!`)
	}

	message := fmt.Sprintf(`📋 *Ваша медицинская карта*

👤 *Имя:* %s
📅 *Первое посещение:* %s
📅 *Последний визит:* %s
🔢 *Всего посещений:* %d
💆 *Последняя услуга:* %s

📝 *Заметки вашего доктора:*
%s

Для получения полной записи в формате Markdown нажмите /downloadrecord`,
		patient.Name,
		patient.FirstVisit.Format("02.01.2006"),
		patient.LastVisit.Format("02.01.2006"),
		patient.TotalVisits,
		patient.CurrentService,
		patient.TherapistNotes)

	return c.Send(message, telebot.ParseMode(telebot.ModeMarkdown))
}

// HandleDownloadRecord sends the Markdown file
func (h *BookingHandler) HandleDownloadRecord(c telebot.Context) error {
	userID := c.Sender().ID
	telegramID := strconv.FormatInt(userID, 10)

	filePath, err := storage.GetPatientMarkdownFile(telegramID)
	if err != nil {
		return c.Send(`📭 Файл с вашей медицинской картой не найден.

Возможные причины:
1. Вы еще не записывались на массаж
2. Ваша карта была создана недавно

Запишитесь через /start чтобы создать вашу карту!`)
	}

	doc := &telebot.Document{
		File:     telebot.FromDisk(filePath),
		FileName: "medical_record.md",
		Caption: `📄 Ваша медицинская карта

*Как открыть этот файл:*
1. **Рекомендуем Obsidian** (бесплатно) — отличный инструмент для ваших записей. Скачайте для любого устройства на https://obsidian.md/download
2. **Или любой текстовый редактор** (Блокнот, TextEdit)

*Скачайте Obsidian для удобного ведения медицинского дневника!*`,
	}

	return c.Send(doc)
}
