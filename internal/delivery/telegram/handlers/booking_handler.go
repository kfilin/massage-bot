package handlers

import (
	"context"
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/kfilin/massage-bot/internal/domain"
	"github.com/kfilin/massage-bot/internal/logging"
	"github.com/kfilin/massage-bot/internal/monitoring"
	"github.com/kfilin/massage-bot/internal/ports"
	"github.com/kfilin/massage-bot/internal/presentation"
	"gopkg.in/telebot.v3"
)

// BookingHandler is the central handler for booking-related commands and
// callbacks. Methods on this struct are split across this and three sibling
// files (booking_admin.go, booking_file.go, booking_session.go) for
// navigability — they all belong to the same struct.
type BookingHandler struct {
	appointmentService   ports.AppointmentService
	sessionStorage       ports.SessionStorage
	adminIDs             []string
	therapistIDs         []string
	transcriptionService ports.TranscriptionService
	repository           ports.Repository
	presenter            *presentation.BotPresenter
	WebAppURL            string
	webAppSecret         string
}

func NewBookingHandler(as ports.AppointmentService, ss ports.SessionStorage, admins []string, therapistIDs []string, trans ports.TranscriptionService, repo ports.Repository, presenter *presentation.BotPresenter, webAppURL string, webAppSecret string) *BookingHandler {
	return &BookingHandler{
		appointmentService:   as,
		sessionStorage:       ss,
		adminIDs:             admins,
		therapistIDs:         therapistIDs,
		transcriptionService: trans,
		repository:           repo,
		presenter:            presenter,
		WebAppURL:            webAppURL,
		webAppSecret:         webAppSecret,
	}
}

func (h *BookingHandler) HandleStart(c telebot.Context) error {
	userID := c.Sender().ID
	logging.Debugf(": Entered HandleStart for user %d", userID)
	h.sessionStorage.ClearSession(userID)

	// 1. Handle deep links
	args := c.Args()
	if len(args) > 0 {
		arg := args[0]
		if strings.HasPrefix(arg, "manual_") {
			targetID := strings.TrimPrefix(arg, "manual_")
			isAdmin := false
			userIDStr := strconv.FormatInt(userID, 10)
			for _, id := range h.adminIDs {
				if id == userIDStr {
					isAdmin = true
					break
				}
			}
			if isAdmin {
				h.sessionStorage.Set(userID, SessionKeyIsAdminManual, true)
				h.sessionStorage.Set(userID, SessionKeyPatientID, targetID)
				patient, err := h.repository.GetPatient(targetID)
				if err == nil {
					h.sessionStorage.Set(userID, SessionKeyName, patient.Name)
					logging.Debugf(": Deep link manual booking: detected patient %s for admin %d", patient.Name, userID)
					return h.showCategories(c)
				}
			}
		} else if arg == "book" {
			// Just proceed to booking
			return h.showCategories(c)
		}
	}

	// 2. Welcome message
	welcomeMsg := h.presenter.FormatWelcome(c.Sender().FirstName)
	_ = c.Send(welcomeMsg, h.GetMainMenu(), telebot.ModeHTML)

	h.sessionStorage.Set(userID, SessionKeyIsAdminBlock, false)

	// Async analytics
	go func() {
		if err := h.repository.LogEvent(strconv.FormatInt(userID, 10), "start_bot", nil); err != nil {
			logging.Warnf("Failed to log start_bot event: %v", err)
		}
	}()

	// Tentatively register patient if not exists to capture Telegram name
	existingPatient, err := h.repository.GetPatient(strconv.FormatInt(userID, 10))
	if err != nil {
		firstName := c.Sender().FirstName
		lastName := c.Sender().LastName
		fullName := strings.TrimSpace(firstName + " " + lastName)
		if fullName == "" {
			fullName = c.Sender().Username
		}
		if fullName != "" {
			errSave := h.repository.SavePatient(domain.Patient{
				TelegramID:     strconv.FormatInt(userID, 10),
				Name:           fullName,
				HealthStatus:   "initial",
				TherapistNotes: fmt.Sprintf("Зарегистрирован через /start: %s", time.Now().Format("02.01.2006")),
			})
			if errSave != nil {
				logging.Errorf(": Failed to tentatively save new patient %d: %v", userID, errSave)
			}
		}
	} else if existingPatient.Name == "" {
		firstName := c.Sender().FirstName
		lastName := c.Sender().LastName
		fullName := strings.TrimSpace(firstName + " " + lastName)
		if fullName != "" {
			existingPatient.Name = fullName
			errSave := h.repository.SavePatient(existingPatient)
			if errSave != nil {
				logging.Errorf(": Failed to update patient name for %d: %v", userID, errSave)
			}
		}
	}

	return h.showCategories(c)
}

func (h *BookingHandler) showCategories(c telebot.Context) error {
	selector := &telebot.ReplyMarkup{}
	btnMassages := selector.Data("💆 Массаж", "select_category", "massages")
	btnConsultations := selector.Data("👥 Консультация", "select_category", "consultations")
	btnOther := selector.Data("✨ Другие услуги", "select_category", "other")

	selector.Inline(
		selector.Row(btnMassages),
		selector.Row(btnConsultations),
		selector.Row(btnOther),
	)

	msg := "Выберите категорию услуг:"
	if c.Callback() != nil {
		return c.Edit(msg, selector)
	}
	return c.Send(msg, selector)
}

func (h *BookingHandler) HandleCategorySelection(c telebot.Context) error {
	data := strings.TrimSpace(c.Callback().Data)
	parts := strings.Split(data, "|")
	if len(parts) != 2 || parts[0] != "select_category" {
		return c.Edit("Ошибка выбора категории.")
	}

	category := parts[1]
	if category == "back" {
		return h.showCategories(c)
	}

	userID := c.Sender().ID
	h.sessionStorage.Set(userID, SessionKeyCategory, category)

	services, err := h.appointmentService.GetAvailableServices(context.Background())
	if err != nil {
		logging.Infof("Error getting services: %v", err)
		return c.Edit("Ошибка загрузки услуг.")
	}

	selector := &telebot.ReplyMarkup{}
	var rows []telebot.Row

	for _, svc := range services {
		include := false
		name := svc.Name

		switch category {
		case "massages":
			if name == "Массаж Спина + Шея" || name == "Общий массаж" || name == "Лимфодренаж" {
				include = true
			}
		case "consultations":
			if name == "Консультация офлайн" || name == "Консультация онлайн" {
				include = true
			}
		case "other":
			if name == "Иглоукалывание" || name == "Реабилитационные программы" {
				include = true
			}
		}

		if include {
			label := fmt.Sprintf("%s · %.0f₺", name, svc.Price)
			rows = append(rows, selector.Row(selector.Data(label, "select_service", svc.ID)))
		}
	}

	btnBack := selector.Data("⬅️ Назад", "select_category", "back")
	rows = append(rows, selector.Row(btnBack))

	selector.Inline(rows...)
	return c.Edit("Выберите конкретную услугу:", selector)
}

func (h *BookingHandler) getMainMenuWithBackBtn() *telebot.ReplyMarkup {
	menu := h.GetMainMenu()
	// Insert "Select another date" as the first row.
	// telebot.v3 uses ReplyButton for ReplyKeyboard.
	backBtnRow := []telebot.ReplyButton{{Text: "⬅️ Выбрать другую дату"}}
	menu.ReplyKeyboard = append([][]telebot.ReplyButton{backBtnRow}, menu.ReplyKeyboard...)
	return menu
}

func (h *BookingHandler) GetMainMenu() *telebot.ReplyMarkup {
	menu := &telebot.ReplyMarkup{ResizeKeyboard: true}
	menu.Reply(
		menu.Row(menu.Text("🗓 Записаться"), menu.Text("📅 Мои записи")),
		menu.Row(menu.Text("📄 Мед-карта"), menu.Text("📤 Загрузить документы")),
	)
	return menu
}

func (h *BookingHandler) HandleServiceSelection(c telebot.Context) error {
	logging.Debugf(": Entered HandleServiceSelection for user %d. Callback Data: '%s'", c.Sender().ID, c.Callback().Data)

	// Callback data is "select_service|SERVICE_ID". We need to split it.
	data := strings.TrimSpace(c.Callback().Data) // Trim spaces just in case
	parts := strings.Split(data, "|")

	logging.Debugf(": HandleServiceSelection - Parsed parts: %v (length: %d)", parts, len(parts))

	if len(parts) != 2 || parts[0] != "select_service" {
		logging.Errorf(": HandleServiceSelection - Malformed service selection callback data. Expected 'select_service|ID', got: '%s'", data)
		return c.Edit("Некорректный выбор услуги. Пожалуйста, попробуйте /start снова.")
	}
	serviceID := parts[1]
	logging.Debugf(": HandleServiceSelection - Extracted serviceID: '%s'", serviceID)

	userID := c.Sender().ID

	// HANDLE ADMIN BLOCKING "FAKE" SERVICES
	if strings.HasPrefix(serviceID, "block_") {
		var durationMinutes int
		var name string

		switch serviceID {
		case "block_30":
			durationMinutes = 30
			name = "⛔ Блок: 30 мин"
		case "block_60":
			durationMinutes = 60
			name = "⛔ Блок: 1 час"
		case "block_90":
			durationMinutes = 90
			name = "⛔ Блок: 1.5 часа"
		case "block_120":
			durationMinutes = 120
			name = "⛔ Блок: 2 часа"
		case "block_day":
			durationMinutes = 480 // 8 hours (work day) - or handle differently
			name = "⛔ Блок: Весь день"
		}

		fakeService := domain.Service{
			ID:              serviceID,
			Name:            name,
			DurationMinutes: durationMinutes,
			Price:           0,
		}

		// Store service struct directly in session (consistent with normal services)
		h.sessionStorage.Set(userID, SessionKeyService, fakeService)

		return h.askForDate(c, fakeService.Name) // Proceed to date selection directly
	}

	services, err := h.appointmentService.GetAvailableServices(context.Background())
	if err != nil {
		logging.Infof("Error getting services in HandleServiceSelection: %v", err)
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
		logging.Errorf(": Service with ID '%s' not found in available services for user %d", serviceID, userID)
		return c.Edit("Выбранная услуга не найдена. Пожалуйста, выберите из предложенных.")
	}

	h.sessionStorage.Set(userID, SessionKeyService, chosenService)
	logging.Debugf(": Service selected and stored in session for user %d: %s (ID: %s)", userID, chosenService.Name, chosenService.ID)

	go func() {
		if err := h.repository.LogEvent(strconv.FormatInt(userID, 10), "service_selected", map[string]interface{}{
			"service_id":   chosenService.ID,
			"service_name": chosenService.Name,
			"price":        chosenService.Price,
		}); err != nil {
			logging.Warnf("Failed to log service_selected event: %v", err)
		}
	}()

	// Ask for date
	return h.askForDate(c, chosenService.Name)
}

func (h *BookingHandler) askForDate(c telebot.Context, serviceName string) error {
	logging.Debugf(": Entered askForDate for user %d. Service: %s", c.Sender().ID, serviceName)

	now := time.Now()
	year, month, _ := now.Date()
	// Use domain.ApptTimeZone for consistency across the application
	currentMonth := time.Date(year, month, 1, 0, 0, 0, 0, domain.ApptTimeZone)

	calendarKeyboard := h.generateCalendar(currentMonth)

	return c.EditOrSend(
		fmt.Sprintf("Отлично, услуга '%s' выбрана. Теперь выберите дату:\n\n<i>░X░ — дата недоступна</i>", serviceName),
		calendarKeyboard,
		telebot.ModeHTML,
	)
}

func (h *BookingHandler) generateCalendar(month time.Time) *telebot.ReplyMarkup {
	logging.Debugf(": Generating calendar for month: %s", month.Format("2006-01"))
	selector := &telebot.ReplyMarkup{}
	var rows []telebot.Row

	// Navigation row
	prevMonth := month.AddDate(0, -1, 0)
	nextMonth := month.AddDate(0, 1, 0)
	rows = append(rows, selector.Row(
		selector.Data("⬅️", "navigate_month", prevMonth.Format("2006-01")),
		selector.Data(month.Format("January 2006"), "ignore"),
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
	offset := (int(firstDayOfMonth.Weekday()) + 6) % 7
	startDay := firstDayOfMonth.AddDate(0, 0, -offset)

	for week := 0; week < 6; week++ {
		var weekBtns []telebot.Btn
		for day := 0; day < 7; day++ {
			currentDay := startDay.AddDate(0, 0, week*7+day)
			loc := domain.ApptTimeZone
			if loc == nil {
				loc = time.Local
			}
			nowInLoc := time.Now().In(loc).Truncate(24 * time.Hour)

			if currentDay.Month() != month.Month() {
				weekBtns = append(weekBtns, selector.Data(" ", "ignore"))
			} else {
				dayStr := fmt.Sprintf("%d", currentDay.Day())
				isPast := currentDay.Truncate(24 * time.Hour).Before(nowInLoc)
				isWeekend := currentDay.Weekday() == time.Saturday || currentDay.Weekday() == time.Sunday

				if isPast || isWeekend {
					// Use a "faded" look for unavailable dates
					fadedDay := fmt.Sprintf("░%d░", currentDay.Day())
					weekBtns = append(weekBtns, selector.Data(fadedDay, "ignore"))
				} else {
					weekBtns = append(weekBtns, selector.Data(dayStr, "select_date", currentDay.Format("2006-01-02")))
				}
			}
		}
		rows = append(rows, selector.Row(weekBtns...))
	}

	// Back button to return to service selection
	rows = append(rows, selector.Row(selector.Data("⬅️ Назад к выбору услуги", "back_to_services")))

	selector.Inline(rows...)
	return selector
}

func (h *BookingHandler) HandleDateSelection(c telebot.Context) error {
	logging.Debugf(": Entered HandleDateSelection for user %d. Callback Data: '%s'", c.Sender().ID, c.Callback().Data)

	data := strings.TrimSpace(c.Callback().Data) // Trim spaces
	userID := c.Sender().ID

	if strings.HasPrefix(data, "navigate_month|") {
		parts := strings.Split(data, "|")
		if len(parts) != 2 || parts[0] != "navigate_month" {
			logging.Errorf(": Malformed month navigation callback data: %s", data)
			return c.Edit("Некорректная навигация. Попробуйте снова.")
		}
		monthStr := parts[1]
		selectedMonth, err := time.Parse("2006-01", monthStr)
		if err != nil {
			logging.Errorf(": Invalid month format in navigation: %s, error: %v", monthStr, err)
			return c.Edit("Некорректная дата. Попробуйте снова.")
		}
		calendarKeyboard := h.generateCalendar(selectedMonth)
		return c.Edit(c.Message().Text, calendarKeyboard, telebot.ModeHTML) // Edit the existing message
	} else if strings.HasPrefix(data, "select_date|") {
		parts := strings.Split(data, "|")
		if len(parts) != 2 || parts[0] != "select_date" {
			logging.Errorf(": Malformed date selection callback data: %s", data)
			return c.Edit("Некорректный выбор даты. Попробуйте /start снова.")
		}
		dateStr := parts[1]
		selectedDate, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			logging.Errorf(": Invalid date format in selection: %s, error: %v", dateStr, err)
			return c.Edit("Некорректная дата. Попробуйте /start снова.")
		}

		h.sessionStorage.Set(userID, SessionKeyDate, selectedDate)
		logging.Debugf(": Date selected and stored in session for user %d: %s", userID, selectedDate.Format("2006-01-02"))

		// Now ask for time
		return h.askForTime(c)
	} else if data == "back_to_services" {
		// Return to service selection for the stored category
		session := h.sessionStorage.Get(userID)
		category, ok := session[SessionKeyCategory].(string)
		if !ok || category == "" {
			return h.showCategories(c)
		}
		// Mock callback data for HandleCategorySelection
		c.Callback().Data = "select_category|" + category
		return h.HandleCategorySelection(c)
	}
	return c.Send("Неизвестное действие с датой. Пожалуйста, попробуйте /start снова.")
}

func (h *BookingHandler) HandleReminderConfirmation(c telebot.Context) error {
	apptID := strings.TrimPrefix(c.Callback().Data, "confirm_appt_reminder|")
	logging.Debugf(": HandleReminderConfirmation called for apptID: %s", apptID)

	now := time.Now()
	err := h.repository.SaveAppointmentMetadata(apptID, &now, nil)
	if err != nil {
		logging.Errorf(": Failed to save confirmation for appt %s: %v", apptID, err)
		return c.Send("❌ Ошибка при подтверждении записи.")
	}

	// Notify Vera
	appt, err := h.appointmentService.FindByID(context.Background(), apptID)
	if err == nil {
		notification := h.presenter.FormatAppointment(appt, true)

		for _, adminIDStr := range h.adminIDs {
			adminID, _ := strconv.ParseInt(adminIDStr, 10, 64)
			h.BotNotify(c.Bot(), adminID, notification)
		}
	}

	return c.Edit("✅ Спасибо! Ваша запись подтверждена. Ждем вас!")
}

func (h *BookingHandler) HandleReminderCancellation(c telebot.Context) error {
	apptID := strings.TrimPrefix(c.Callback().Data, "cancel_appt_reminder|")
	logging.Debugf(": HandleReminderCancellation called for apptID: %s", apptID)

	// Since we already have the ID, we can directly cancel or use the existing callback handler
	// For consistency, let's use the existing callback handler logic
	c.Callback().Data = "cancel_appt|" + apptID
	return h.HandleCancelAppointmentCallback(c)
}

func (h *BookingHandler) HandleAdminReplyRequest(c telebot.Context) error {
	patientID := strings.TrimPrefix(c.Callback().Data, "admin_reply|")
	// Trim any potential leading/trailing whitespace including hidden characters
	patientID = strings.TrimSpace(patientID)
	// Remove the unique prefix if it was duplicated by telebot (rare but possible: "admin_reply|admin_reply|id")
	patientID = strings.TrimPrefix(patientID, "admin_reply|")

	logging.Debugf(": HandleAdminReplyRequest called. Raw Data: '%s', Extracted ID: '%s'", c.Callback().Data, patientID)

	patient, err := h.repository.GetPatient(patientID)
	if err != nil {
		return c.Send("❌ Пациент не найден.")
	}

	h.sessionStorage.Set(c.Sender().ID, SessionKeyAdminReplyingTo, patientID)
	h.sessionStorage.Set(c.Sender().ID, SessionKeyAdminReplyingTo, patientID)
	return c.Send(fmt.Sprintf("✍️ Введите ответ для пациента <b>%s</b> (ID: %s):", patient.Name, patient.TelegramID), telebot.ModeHTML, telebot.ForceReply)
}

func (h *BookingHandler) askForTime(c telebot.Context) error {
	logging.Debugf(": Entered askForTime for user %d", c.Sender().ID)
	userID := c.Sender().ID
	sessionData := h.sessionStorage.Get(userID)

	service, okS := sessionData[SessionKeyService].(domain.Service)
	date, okD := sessionData[SessionKeyDate].(time.Time)

	if !okS || !okD {
		logging.Errorf(": Missing session data for time selection for user %d. Service OK: %t, Date OK: %t", userID, okS, okD)
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

	logging.Debugf(": Calling GetAvailableTimeSlots for user %d with date %s and duration %d", userID, selectedDateInLoc.Format("2006-01-02"), service.DurationMinutes)
	timeSlots, err := h.appointmentService.GetAvailableTimeSlots(context.Background(), selectedDateInLoc, service.DurationMinutes)
	if err != nil {
		logging.Errorf(": Error getting available time slots for user %d: %v", userID, err)
		// Clean up the calendar keyboard before showing the error
		if c.Message() != nil {
			if _, err := c.Bot().EditReplyMarkup(c.Message(), nil); err != nil {
				logging.Warnf("Failed to remove inline keyboard: %v", err)
			}
		}
		return c.Send("❌ Ошибка при получении слотов: " + err.Error() + "\n\nПожалуйста, начните заново: /start")
	}
	logging.Debugf(": Received %d time slots for user %d.", len(timeSlots), userID)

	if len(timeSlots) == 0 {
		// Используем c.EditOrSend для обновления сообщения, если слотов нет
		return c.EditOrSend("На эту дату нет доступных временных слотов. Пожалуйста, выберите другую дату.", h.getMainMenuWithBackBtn())
	}

	selector := &telebot.ReplyMarkup{}
	var rows []telebot.Row
	for _, slot := range timeSlots {
		// Callback data format: "select_time|HH:MM"
		rows = append(rows, selector.Row(
			selector.Data(slot.Start.Format("15:04"), "select_time", slot.Start.Format("15:04")),
		))
	}
	rows = append(rows, selector.Row(selector.Data("⬅️ Назад к выбору даты", "back_to_date")))
	selector.Inline(rows...)

	// Используем специальную клавиатуру: Кнопка "Назад" + Главное меню
	replyKeyboard := h.getMainMenuWithBackBtn()

	// Редактируем предыдущее сообщение (календарь) с новой инлайн-клавиатурой (слоты времени)
	_, err = c.Bot().EditReplyMarkup(c.Message(), nil)
	if err != nil {
		logging.Warnf("Failed to clear previous markup: %v", err)
	}

	err = c.Edit(
		fmt.Sprintf("Отлично, доступны следующие временные слоты для '%s' %s:", service.Name, date.Format("02.01.2006")),
		selector, // Inline keyboard for time slots
	)
	if err != nil {
		logging.Errorf(": Failed to edit message with time slots: %v", err)
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

func (h *BookingHandler) HandleTimeSelection(c telebot.Context) error {
	logging.Debugf(": Entered HandleTimeSelection for user %d. Callback Data: '%s'", c.Sender().ID, c.Callback().Data)

	data := strings.TrimSpace(c.Callback().Data) // Trim spaces
	userID := c.Sender().ID

	if data == "back_to_date" {
		userID := c.Sender().ID
		session := h.sessionStorage.Get(userID)
		service, ok := session[SessionKeyService].(domain.Service)
		if !ok {
			return h.showCategories(c)
		}
		return h.askForDate(c, service.Name)
	}

	parts := strings.Split(data, "|")
	if len(parts) != 2 || parts[0] != "select_time" {
		logging.Errorf(": Malformed time selection callback data: %s", data)
		return c.Edit("Некорректный выбор времени. Пожалуйста, попробуйте /start снова.")
	}
	timeStr := parts[1] // e.g., "15:04"

	// Validate time format. We expect "HH:MM"
	_, err := time.Parse("15:04", timeStr)
	if err != nil {
		logging.Errorf(": Invalid time format in selection: %s, error: %v", timeStr, err)
		return c.Edit("Некорректное время. Пожалуйста, попробуйте /start снова.")
	}
	h.sessionStorage.Set(userID, SessionKeyTime, timeStr)
	logging.Debugf(": Time selected and stored in session for user %d: %s", userID, timeStr)

	// Удаляем инлайн-клавиатуру со слотами времени из предыдущего сообщения
	if c.Message() != nil {
		_, err := c.Bot().EditReplyMarkup(c.Message(), nil) // Pass nil to remove inline keyboard
		if err != nil {
			logging.Warnf("ING: Failed to remove inline keyboard from message %d: %v", c.Message().ID, err)
		}
	}

	// Check if this is a block service (skip name input)
	sessionData := h.sessionStorage.Get(userID)
	if service, ok := sessionData[SessionKeyService].(domain.Service); ok {
		if strings.HasPrefix(service.ID, "block_") {
			h.sessionStorage.Set(userID, SessionKeyName, "Admin")
			logging.Debugf(": Block service detected, skipping name input for user %d", userID)
			return h.askForConfirmation(c)
		}
	}

	// Check if this is a manual admin booking
	if val, ok := sessionData[SessionKeyIsAdminManual].(bool); ok && val {
		// If name is already set (from HandleStart deep link lookup), skip input
		if name, okName := sessionData[SessionKeyName].(string); okName && name != "" {
			logging.Debugf(": Manual admin booking with pre-filled name '%s', skipping input", name)
			return h.askForConfirmation(c)
		}

		logging.Debugf(": Manual admin booking detected for user %d, asking for patient name", userID)
		return c.Send("✍️ Введите <b>имя и фамилию пациента</b> для записи:", telebot.ModeHTML)
	}

	// Check for returning patient (with at least one visit)
	patient, errRepo := h.repository.GetPatient(strconv.FormatInt(userID, 10))
	if errRepo == nil && patient.Name != "" && patient.TotalVisits > 0 {
		h.sessionStorage.Set(userID, SessionKeyName, patient.Name)
		logging.Debugf(": Returning patient %d detected (Name: %s), skipping name input", userID, patient.Name)
		return h.askForConfirmation(c)
	}

	// Теперь переходим к запросу имени.
	// Используем c.Send + RemoveKeyboard чтобы клавиатура не закрывала промпт (Bug #3)
	return c.Send("✍️ Пожалуйста, введите ваше <b>имя и фамилию</b>.\n\nЭто имя будет использоваться в вашей медицинской карте.", telebot.ModeHTML, telebot.RemoveKeyboard)
}

func (h *BookingHandler) HandleNameInput(c telebot.Context) error {
	logging.Debugf(": Entered HandleNameInput for user %d. Text: '%s'", c.Sender().ID, c.Text())

	userID := c.Sender().ID
	userName := strings.TrimSpace(c.Text())

	if userName == "" {
		return c.Send("Имя не может быть пустым. Пожалуйста, введите ваше имя и фамилию.")
	}

	h.sessionStorage.Set(userID, SessionKeyName, userName)
	logging.Debugf(": Name stored in session for user %d: %s", userID, userName)

	// All data collected, ask for confirmation
	return h.askForConfirmation(c)
}

func (h *BookingHandler) askForConfirmation(c telebot.Context) error {
	logging.Debugf(": Entered askForConfirmation for user %d", c.Sender().ID)

	userID := c.Sender().ID
	sessionData := h.sessionStorage.Get(userID)

	service, okS := sessionData[SessionKeyService].(domain.Service)
	date, okD := sessionData[SessionKeyDate].(time.Time)
	timeStr, okT := sessionData[SessionKeyTime].(string)
	name, okN := sessionData[SessionKeyName].(string)

	if !okS || !okD || !okT || !okN {
		logging.Errorf(": Missing session data for confirmation for user %d: service=%t, date=%t, time=%t, name=%t", userID, okS, okD, okT, okN)
		h.sessionStorage.ClearSession(userID)
		return c.Send("Ошибка сессии. Пожалуйста, начните /start снова.", telebot.RemoveKeyboard)
	}

	// Combine date and time string into a time.Time object for display
	appointmentTime, err := time.Parse("2006-01-02 15:04", fmt.Sprintf("%s %s", date.Format("2006-01-02"), timeStr))
	if err != nil {
		logging.Errorf(": Failed to parse appointment time for user %d: %v", userID, err)
		h.sessionStorage.ClearSession(userID)
		return c.Send("Ошибка форматирования времени. Пожалуйста, начните /start снова.", telebot.RemoveKeyboard)
	}

	title := "<b>Пожалуйста, подтвердите вашу запись:</b>"
	if val, ok := sessionData[SessionKeyIsAdminManual].(bool); ok && val {
		title = "<b>Подтвердите создание ручной записи:</b>"
	}

	confirmMessage := h.presenter.FormatBookingSummary(title, name, service.Name, appointmentTime, service.DurationMinutes, service.Price)

	// Inline Keyboard - One button per row for maximum prominence
	selector := &telebot.ReplyMarkup{}
	selector.Inline(
		selector.Row(selector.Data("✅ ПОДТВЕРДИТЬ", "confirm_booking")),
		selector.Row(selector.Data("❌ ОТМЕНИТЬ", "cancel_booking")),
	)

	// Set session flag indicating awaiting confirmation (keep for fallback/cleanup)
	h.sessionStorage.Set(userID, SessionKeyAwaitingConfirmation, true)
	logging.Debugf(": Set SessionKeyAwaitingConfirmation for user %d to true.", userID)

	return c.Send(confirmMessage, selector, telebot.ModeHTML)
}

func (h *BookingHandler) HandleConfirmBooking(c telebot.Context) error {
	logging.Debugf(": Entered HandleConfirmBooking for user %d", c.Sender().ID)

	userID := c.Sender().ID
	sessionData := h.sessionStorage.Get(userID)

	// Clear awaiting confirmation flag
	h.sessionStorage.Set(userID, SessionKeyAwaitingConfirmation, false)
	logging.Debugf(": Cleared SessionKeyAwaitingConfirmation for user %d.", userID)

	service, okS := sessionData[SessionKeyService].(domain.Service)
	date, okD := sessionData[SessionKeyDate].(time.Time)
	timeStr, okT := sessionData[SessionKeyTime].(string)
	name, okN := sessionData[SessionKeyName].(string)

	if !okS || !okD || !okT || !okN {
		logging.Infof("Session data missing for user %d during confirmation: service=%t, date=%t, time=%t, name=%t", userID, okS, okD, okT, okN)
		h.sessionStorage.ClearSession(userID)
		return c.Send("Ошибка сессии. Пожалуйста, начните /start снова.", telebot.RemoveKeyboard)
	}

	// Combine date and time string into a time.Time object
	appointmentTime, err := time.Parse("2006-01-02 15:04", fmt.Sprintf("%s %s", date.Format("2006-01-02"), timeStr))
	if err != nil {
		logging.Infof("Failed to parse appointment time for user %d during confirmation: %v", userID, err)
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

	// Check if this is an Admin Block action
	isAdminBlock := false
	session := h.sessionStorage.Get(userID)
	if val, ok := session[SessionKeyIsAdminBlock].(bool); ok && val {
		isAdminBlock = true
	}
	// Check if this is an Admin manual booking
	isAdminManual := false
	if val, ok := session[SessionKeyIsAdminManual].(bool); ok && val {
		isAdminManual = true
	}

	// Create appointment model
	appt := domain.Appointment{
		ServiceID:    service.ID,
		Service:      service,
		Time:         appointmentTime,
		StartTime:    appointmentTime,
		Duration:     service.DurationMinutes,
		CustomerTgID: strconv.FormatInt(userID, 10),
		CustomerName: name,
		Notes:        "Telegram Bot Booking",
	}

	if isAdminManual {
		// If manual booking, use the SessionKeyPatientID if available
		if targetID, ok := session[SessionKeyPatientID].(string); ok && targetID != "" {
			appt.CustomerTgID = targetID
			logging.Debugf(": Manual booking using stored PatientID: %s", targetID)
		} else {
			// Fallback if ID is missing (should not happen in deep-link flow)
			normalized := strings.ToLower(strings.Join(strings.Fields(name), ""))
			appt.CustomerTgID = "manual_" + normalized
			logging.Warnf(": Manual booking fallback to generated ID: %s", appt.CustomerTgID)
		}
		appt.Notes = "Manual Appointment by Admin"
	}

	if isAdminBlock {
		appt.Notes = "Manual Block by Admin"
		appt.CustomerName = "Admin Block"
		// Use a distinct summary for blocks
		// The service name is already "⛔ Block: X min"
	}

	// Save to Google Calendar (and internal DB via adapter)
	_, err = h.appointmentService.CreateAppointment(context.Background(), &appt)
	if err != nil {
		logging.Infof("Error creating appointment: %v", err)
		if strings.Contains(err.Error(), "slot is not available") {
			return c.Send("❌ К сожалению, это время уже занято. Пожалуйста, выберите другое время.", telebot.RemoveKeyboard)
		}
		if isAdminBlock {
			return c.Send(fmt.Sprintf("❌ Ошибка при создании блокировки: %v", err), telebot.RemoveKeyboard)
		}
		return c.Send("Произошла ошибка при создании записи. Пожалуйста, попробуйте позже.", telebot.RemoveKeyboard)
	}

	if isAdminBlock {
		// For blocks, we are done. Confirm to admin.
		// Clear session
		h.sessionStorage.ClearSession(userID)

		// Notify OTHER admin(s) about the block
		blockerName := c.Sender().FirstName
		if blockerName == "" {
			blockerName = c.Sender().Username
		}
		for _, adminIDStr := range h.adminIDs {
			adminID, _ := strconv.ParseInt(adminIDStr, 10, 64)
			if adminID != userID { // Don't notify the admin who created the block
				details := map[string]string{
					"Админ":  blockerName,
					"Дата":   appointmentTime.Format("02.01.2006"),
					"Время":  appointmentTime.Format("15:04"),
					"Услуга": service.Name,
				}
				h.BotNotify(c.Bot(), adminID, h.presenter.FormatNotification("Время заблокировано", details))
			}
		}

		// Use createdAppt info if available, otherwise use request data
		details := map[string]string{
			"Дата":         appointmentTime.Format("02.01.2006"),
			"Время":        appointmentTime.Format("15:04"),
			"Услуга":       service.Name,
		}
		return c.Send(h.presenter.FormatNotification("Время заблокировано", details), telebot.ModeHTML)
	}
	// Update or create patient record using robust sync
	var nameInSync string
	if n, ok := session[SessionKeyName].(string); ok {
		nameInSync = n
	}
	patient, errSync := h.syncPatientStats(context.Background(), appt.CustomerTgID, nameInSync)
	if errSync != nil {
		logging.Warnf("ING: Failed to sync patient record for user %d: %v", userID, errSync)
		// Fallback to minimal update if sync fails
		existingPatient, errRepo := h.repository.GetPatient(appt.CustomerTgID)
		if errRepo == nil {
			patient = existingPatient
			patient.LastVisit = appointmentTime
			patient.TotalVisits++
			if err := h.repository.SavePatient(patient); err != nil {
				logging.Errorf("Failed to save patient fallback: %v", err)
			}
		}
	} else {
		logging.Infof("Patient record synced for user %d (TotalVisits: %d)", userID, patient.TotalVisits)
		// Record patient loyalty metric
		if patient.TotalVisits <= 1 {
			monitoring.AppointmentTypeTotal.WithLabelValues("first_visit").Inc()
		} else {
			monitoring.AppointmentTypeTotal.WithLabelValues("returning").Inc()
		}

		// Log analytics event
		if err := h.repository.LogEvent(patient.TelegramID, "booking_confirmed", map[string]interface{}{
			"service_id":     service.ID,
			"service_name":   service.Name,
			"time":           appointmentTime.Format(time.RFC3339),
			"is_admin_block": isAdminBlock,
			"visit_count":    patient.TotalVisits,
		}); err != nil {
			logging.Warnf("Failed to log booking_confirmed event: %v", err)
		}
	}

	// 1. Notify Admin(s)
	adminMsg := h.presenter.FormatAppointment(&appt, true)
	for _, adminIDStr := range h.adminIDs {
		adminID, _ := strconv.ParseInt(adminIDStr, 10, 64)
		h.BotNotify(c.Bot(), adminID, adminMsg)
	}

	// 2. Notify Therapists
	for _, tID := range h.therapistIDs {
		therapistID, _ := strconv.ParseInt(tID, 10, 64)
		h.BotNotify(c.Bot(), therapistID, adminMsg)
	}

	// Increment booking metric
	monitoring.IncrementBooking(service.Name)

	// Clear session on successful booking
	h.sessionStorage.ClearSession(userID)

	// 3. Confirm to User (Admin or Patient)
	confirmationMsg := h.presenter.FormatAppointment(&appt, false)
	if isAdminManual {
		confirmationMsg = "✅ <b>РУЧНАЯ ЗАПИСЬ СОЗДАНА</b>\n" + confirmationMsg
	}

	// Add Calendar Link
	calendarLink := h.generateGoogleCalendarLink(appt)
	confirmationMsg += fmt.Sprintf("\n\n<a href=\"%s\">📅 Добавить в Календарь</a>", calendarLink)

	selector := &telebot.ReplyMarkup{}
	url := h.GenerateWebAppURL(patient.TelegramID)
	if url != "" {
		selector.Inline(
			selector.Row(selector.WebApp("📱 ОТКРЫТЬ МЕД-КАРТУ (LIVE)", &telebot.WebApp{URL: url})),
		)
	}

	// 4. Notify patient if manual
	if isAdminManual {
		patientIDStr, ok := session[SessionKeyPatientID].(string)
		if ok && patientIDStr != "" {
			patientID, _ := strconv.ParseInt(patientIDStr, 10, 64)
			h.BotNotify(c.Bot(), patientID, h.presenter.FormatAppointment(&appt, false))
		}
	}

	return c.Send(confirmationMsg, h.GetMainMenu(), selector, telebot.ModeHTML)
}

func (h *BookingHandler) syncPatientStats(ctx context.Context, telegramID string, name string) (domain.Patient, error) {
	patient, err := h.repository.GetPatient(telegramID)
	if err != nil {
		// If patient not found, initialize a new one
		if name == "" {
			name = "Пациент"
		}
		patient = domain.Patient{
			TelegramID:     telegramID,
			Name:           name,
			HealthStatus:   "initial",
			TherapistNotes: fmt.Sprintf("Зарегистрирован: %s", time.Now().Format("02.01.2006")),
		}
	} else if name != "" {
		// Update name if patient provided a new one during booking
		patient.Name = name
	}

	// Fetch ALL history from GCal
	appts, err := h.appointmentService.GetCustomerHistory(ctx, telegramID)
	if err != nil {
		return patient, err
	}

	var lastVisit, firstVisit time.Time
	confirmedCount := 0
	if len(appts) > 0 {
		for _, a := range appts {
			// Filter: Only confirmed visits, skip cancellations and admin blocks
			if a.Status == "cancelled" || strings.Contains(strings.ToLower(a.Service.Name), "block") || strings.Contains(strings.ToLower(a.CustomerName), "admin block") {
				continue
			}

			confirmedCount++
			if firstVisit.IsZero() || a.StartTime.Before(firstVisit) {
				firstVisit = a.StartTime
			}
			if lastVisit.IsZero() || a.StartTime.After(lastVisit) {
				lastVisit = a.StartTime
			}
		}
		patient.FirstVisit = firstVisit
		patient.LastVisit = lastVisit
	}
	patient.TotalVisits = confirmedCount

	// Save back to repository
	if err := h.repository.SavePatient(patient); err != nil {
		return patient, err
	}

	return patient, nil
}

func (h *BookingHandler) HandleCancel(c telebot.Context) error {
	logging.Debugf(": Entered HandleCancel for user %d", c.Sender().ID)
	userID := c.Sender().ID
	// Clear awaiting confirmation flag
	h.sessionStorage.Set(userID, SessionKeyAwaitingConfirmation, false)
	logging.Debugf(": Cleared SessionKeyAwaitingConfirmation for user %d (via cancel).", userID)

	h.sessionStorage.ClearSession(userID)
	// Remove keyboard and send confirmation
	return c.Send("Запись отменена. Сессия очищена. Вы можете начать /start снова.", h.GetMainMenu())
}

func (h *BookingHandler) HandleMyRecords(c telebot.Context) error {
	userID := c.Sender().ID
	telegramID := strconv.FormatInt(userID, 10)

	patient, err := h.repository.GetPatient(telegramID)
	if err != nil {
		return c.Send(`📊 <b>Ваша медицинская карта</b>

У вас еще нет активной медицинской карты. Она создается автоматически после первого посещения.

Запишитесь на прием через меню бота!`, telebot.ModeHTML)
	}

	card := h.presenter.FormatPatientCard(patient)

	// Compact menu for record management
	selector := &telebot.ReplyMarkup{}
	url := h.GenerateWebAppURL(patient.TelegramID)

	if url != "" {
		btnWebApp := selector.WebApp("📱 ОТКРЫТЬ МЕД-КАРТУ (LIVE)", &telebot.WebApp{URL: url})
		selector.Inline(
			selector.Row(btnWebApp),
		)
	}

	return c.Send(card, telebot.ModeHTML, selector)
}

func (h *BookingHandler) HandleMyAppointments(c telebot.Context) error {
	userID := c.Sender().ID
	telegramID := strconv.FormatInt(userID, 10)

	isAdmin := h.IsAdmin(userID)
	var appts []domain.Appointment
	var err error

	if isAdmin {
		appts, err = h.appointmentService.GetAllUpcomingAppointments(context.Background())
	} else {
		appts, err = h.appointmentService.GetCustomerAppointments(context.Background(), telegramID)
	}

	if err != nil {
		logging.Errorf(": Failed to get appointments for user %d: %v", userID, err)
		return c.Send("Ошибка при получении списка записей. Пожалуйста, попробуйте позже.")
	}

	if len(appts) == 0 {
		return c.Send("Активных записей не найдено.")
	}

	h.sessionStorage.ClearSession(userID)

	var message string
	if isAdmin {
		message = "📊 <b>Общее расписание записей:</b>\n\n"
	} else {
		message = "📋 <b>Ваши текущие записи:</b>\n\n"
	}

	selector := &telebot.ReplyMarkup{}
	var rows []telebot.Row
	hasLateAppts := false

	// Sort by time for display
	sort.Slice(appts, func(i, j int) bool {
		return appts[i].StartTime.Before(appts[j].StartTime)
	})

	for _, appt := range appts {
		apptTime := appt.StartTime.In(domain.ApptTimeZone)

		// For admins, show who the appointment is for
		patientInfo := ""
		if isAdmin && appt.CustomerTgID != telegramID {
			patientInfo = fmt.Sprintf("👤 %s\n", appt.CustomerName)
		}

		message += fmt.Sprintf("🗓 <b>%s</b>\n🕒 %s\n💆 %s\n%s",
			apptTime.Format("02.01.2006"),
			apptTime.Format("15:04"),
			appt.Service.Name,
			patientInfo)

		// Smart Cancellation Logic: Only show Cancel button if more than 72 hours (3 days) remain
		// OR if the user is an admin (admins can always cancel)
		now := time.Now().In(domain.ApptTimeZone)
		timeRemaining := appt.StartTime.Sub(now)

		if isAdmin || timeRemaining > 72*time.Hour {
			btn := selector.Data(fmt.Sprintf("❌ Отменить %s (%s)", apptTime.Format("02.01"), apptTime.Format("15:04")), "cancel_appt", appt.ID)
			rows = append(rows, selector.Row(btn))
		} else {
			message += "⚠️ <i>Отмена только через терапевта</i>\n"
			hasLateAppts = true
		}
		message += "\n"
	}

	if hasLateAppts && !isAdmin {
		btnContact := selector.URL("💬 Написать терапевту", "https://t.me/VeraFethiye")
		rows = append(rows, selector.Row(btnContact))
	}

	selector.Inline(rows...)

	return c.Send(message, selector, telebot.ModeHTML)
}

func (h *BookingHandler) HandleCancelAppointmentCallback(c telebot.Context) error {
	callbackData := strings.TrimSpace(c.Callback().Data)
	parts := strings.Split(callbackData, "|")
	if len(parts) < 2 {
		return c.Respond(&telebot.CallbackResponse{Text: "Ошибка: неверные данные для отмены."})
	}

	appointmentID := parts[1]
	logging.Debugf(": HandleCancelAppointmentCallback for ID: %s", appointmentID)

	// Get appointment details BEFORE deleting for block check
	appt, _ := h.appointmentService.FindByID(context.Background(), appointmentID)

	if appt != nil {
		now := time.Now().In(domain.ApptTimeZone)
		if appt.StartTime.Sub(now) < 72*time.Hour {
			logging.Infof("BLOCKED: Late cancellation attempt for user %s, appt %s", appt.CustomerTgID, appt.ID)
			return c.Respond(&telebot.CallbackResponse{
				Text:      "⛔ До записи меньше 3 дней!\nАвтоматическая отмена невозможна.\nПожалуйста, напишите терапевту напрямую.",
				ShowAlert: true,
			})
		}
	}

	err := h.appointmentService.CancelAppointment(context.Background(), appointmentID)
	if err != nil {
		logging.Errorf(": Failed to cancel appointment %s: %v", appointmentID, err)
		return c.Respond(&telebot.CallbackResponse{Text: "Не удалось отменить запись. Возможно, она уже отменена."})
	}

	// Notify admin
	if appt != nil {
		for _, adminIDStr := range h.adminIDs {
			adminID, _ := strconv.ParseInt(adminIDStr, 10, 64)
			h.BotNotify(c.Bot(), adminID, h.presenter.FormatCancellation(appt, true))
		}

		// Robust sync after cancellation
		if _, err := h.syncPatientStats(context.Background(), appt.CustomerTgID, appt.CustomerName); err != nil {
			logging.Warnf("Failed to sync patient stats after cancellation: %v", err)
		}
	}

	if err := c.Respond(&telebot.CallbackResponse{Text: "Запись успешно отменена!"}); err != nil {
		logging.Warnf("Failed to respond to callback: %v", err)
	}
	if err := c.Edit("✅ Ваша запись успешно отменена и удалена из календаря."); err != nil {
		logging.Warnf("Failed to edit cancellation message: %v", err)
	}

	return h.HandleMyAppointments(c)
}

