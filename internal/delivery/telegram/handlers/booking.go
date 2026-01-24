package handlers

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"html"
	"log"
	"path/filepath"
	"strconv"
	"strings"
	"time" // Ensure time is imported

	"github.com/kfilin/massage-bot/internal/domain"
	"github.com/kfilin/massage-bot/internal/monitoring"
	"github.com/kfilin/massage-bot/internal/ports" // Alias to avoid conflict with package name "appointment"
	"gopkg.in/telebot.v3"                          // Ensure telebot.v3 is correctly imported
)

// BookingHandler handles booking-related commands and callbacks.
type BookingHandler struct {
	appointmentService   ports.AppointmentService
	sessionStorage       ports.SessionStorage
	adminIDs             []string
	therapistID          string // Added to notify Vera
	transcriptionService ports.TranscriptionService
	repository           ports.Repository
	webAppURL            string
	webAppSecret         string
}

// Session keys
const (
	SessionKeyService              = "service"
	SessionKeyDate                 = "date"
	SessionKeyTime                 = "time"
	SessionKeyName                 = "name"
	SessionKeyAwaitingConfirmation = "awaiting_confirmation"
	SessionKeyCategory             = "category" // New for categorized menu
	SessionKeyIsAdminBlock         = "is_admin_block"
)

// NewBookingHandler creates a new BookingHandler.
func NewBookingHandler(as ports.AppointmentService, ss ports.SessionStorage, admins []string, therapistID string, trans ports.TranscriptionService, repo ports.Repository, webAppURL string, webAppSecret string) *BookingHandler {
	return &BookingHandler{
		appointmentService:   as,
		sessionStorage:       ss,
		adminIDs:             admins,
		therapistID:          therapistID,
		transcriptionService: trans,
		repository:           repo,
		webAppURL:            webAppURL,
		webAppSecret:         webAppSecret,
	}
}

// HandleStart handles the /start command, greeting the user and offering services.
func (h *BookingHandler) HandleStart(c telebot.Context) error {
	userID := c.Sender().ID
	log.Printf("DEBUG: Entered HandleStart for user %d", userID)
	h.sessionStorage.ClearSession(userID)

	// First, send the persistent main menu
	c.Send("💆 Добро пожаловать!", h.GetMainMenu())

	h.sessionStorage.Set(userID, SessionKeyIsAdminBlock, false)

	h.repository.LogEvent(strconv.FormatInt(userID, 10), "start_bot", nil)

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
				log.Printf("ERROR: Failed to tentatively save new patient %d: %v", userID, errSave)
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
				log.Printf("ERROR: Failed to update patient name for %d: %v", userID, errSave)
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

// HandleCategorySelection handles the callback query for category selection.
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
		log.Printf("Error getting services: %v", err)
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

// HandleBlock initiates the admin Blocking Time flow
func (h *BookingHandler) HandleBlock(c telebot.Context) error {
	userID := c.Sender().ID

	// Check if user is admin
	isAdmin := false
	userIDStr := strconv.FormatInt(userID, 10)
	for _, id := range h.adminIDs {
		if id == userIDStr {
			isAdmin = true
			break
		}
	}

	if !isAdmin {
		return c.Send("❌ Эта команда доступна только администраторам.")
	}

	// Set session flag for Admin Block Mode
	h.sessionStorage.Set(userID, SessionKeyIsAdminBlock, true)

	// Define Fake Services for Blocking
	selector := &telebot.ReplyMarkup{}

	btn30 := selector.Data("⛔ 30 мин", "select_service", "block_30")
	btn60 := selector.Data("⛔ 1 час", "select_service", "block_60")
	btn90 := selector.Data("⛔ 1.5 часа", "select_service", "block_90")
	btn120 := selector.Data("⛔ 2 часа", "select_service", "block_120")
	btnDay := selector.Data("📅 Весь день", "select_service", "block_day") // Special handling needed?

	selector.Inline(
		selector.Row(btn30, btn60),
		selector.Row(btn90, btn120),
		selector.Row(btnDay),
	)

	return c.Send("🔒 <b>Блокировка времени</b>\nВыберите длительность:", selector, telebot.ModeHTML)
}

// getMainMenuWithBackBtn returns the main menu with an additional "Select another date" button
func (h *BookingHandler) getMainMenuWithBackBtn() *telebot.ReplyMarkup {
	menu := h.GetMainMenu()
	// Insert "Select another date" as the first row.
	// telebot.v3 uses ReplyButton for ReplyKeyboard.
	backBtnRow := []telebot.ReplyButton{{Text: "⬅️ Выбрать другую дату"}}
	menu.ReplyKeyboard = append([][]telebot.ReplyButton{backBtnRow}, menu.ReplyKeyboard...)
	return menu
}

// GetMainMenu returns the persistent Reply Keyboard for patients in a compact 2x2 grid
func (h *BookingHandler) GetMainMenu() *telebot.ReplyMarkup {
	menu := &telebot.ReplyMarkup{ResizeKeyboard: true}
	menu.Reply(
		menu.Row(menu.Text("🗓 Записаться"), menu.Text("📅 Мои записи")),
		menu.Row(menu.Text("📄 Мед-карта"), menu.Text("📤 Загрузить документы")),
	)
	return menu
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

	h.repository.LogEvent(strconv.FormatInt(userID, 10), "service_selected", map[string]interface{}{
		"service_id":   chosenService.ID,
		"service_name": chosenService.Name,
	})

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

	calendarKeyboard := h.generateCalendar(currentMonth)

	return c.EditOrSend(
		fmt.Sprintf("Отлично, услуга '%s' выбрана. Теперь выберите дату:\n\n<i>░X░ — дата недоступна</i>", serviceName),
		calendarKeyboard,
		telebot.ModeHTML,
	)
}

// generateCalendar creates an inline keyboard for month navigation and date selection.
func (h *BookingHandler) generateCalendar(month time.Time) *telebot.ReplyMarkup {
	log.Printf("DEBUG: Generating calendar for month: %s", month.Format("2006-01"))
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
		calendarKeyboard := h.generateCalendar(selectedMonth)
		return c.Edit(c.Message().Text, calendarKeyboard, telebot.ModeHTML) // Edit the existing message
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

	// Check if this is a block service (skip name input)
	sessionData := h.sessionStorage.Get(userID)
	if service, ok := sessionData[SessionKeyService].(domain.Service); ok {
		if strings.HasPrefix(service.ID, "block_") {
			h.sessionStorage.Set(userID, SessionKeyName, "Admin")
			log.Printf("DEBUG: Block service detected, skipping name input for user %d", userID)
			return h.askForConfirmation(c)
		}
	}

	// Check for returning patient (with at least one visit)
	patient, errRepo := h.repository.GetPatient(strconv.FormatInt(userID, 10))
	if errRepo == nil && patient.Name != "" && patient.TotalVisits > 0 {
		h.sessionStorage.Set(userID, SessionKeyName, patient.Name)
		log.Printf("DEBUG: Returning patient %d detected (Name: %s), skipping name input", userID, patient.Name)
		return h.askForConfirmation(c)
	}

	// Теперь переходим к запросу имени.
	// Используем c.Send для отправки нового сообщения и удаления ReplyKeyboard
	return c.Send("Пожалуйста, введите ваше имя и фамилию для записи (например, Иван Иванов).")
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
			"Длительность: <b>%d мин</b>\n"+
			"Цена: <b>%.0f ₺</b>\n"+
			"Дата: <b>%s</b>\n"+
			"Время: <b>%s</b>\n"+
			"Имя: <b>%s</b>\n\n"+
			"Всё верно?",
		service.Name,
		service.DurationMinutes,
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

	// Check if this is an Admin Block action
	isAdminBlock := false
	session := h.sessionStorage.Get(userID)
	if val, ok := session[SessionKeyIsAdminBlock].(bool); ok && val {
		isAdminBlock = true
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

	if isAdminBlock {
		appt.Notes = "Manual Block by Admin"
		appt.CustomerName = "Admin Block"
		// Use a distinct summary for blocks
		// The service name is already "⛔ Block: X min"
	}

	// Save to Google Calendar (and internal DB via adapter)
	_, err = h.appointmentService.CreateAppointment(context.Background(), &appt)
	if err != nil {
		log.Printf("Error creating appointment: %v", err)
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
				h.BotNotify(c.Bot(), adminID, fmt.Sprintf("🔒 *Время заблокировано*\n\nАдмин: %s\nДата: %s\nВремя: %s\nДлительность: %s",
					blockerName,
					appointmentTime.Format("02.01.2006"),
					appointmentTime.Format("15:04"),
					service.Name))
			}
		}

		// Use createdAppt info if available, otherwise use request data
		return c.Send(fmt.Sprintf("✅ <b>Время заблокировано!</b>\n\n📅 %s\n⏰ %s\n⏳ %s",
			appointmentTime.Format("02.01.2006"),
			appointmentTime.Format("15:04"),
			service.Name), telebot.ModeHTML)
	}
	// Update or create patient record using robust sync
	var nameInSync string
	if n, ok := session[SessionKeyName].(string); ok {
		nameInSync = n
	}
	patient, errSync := h.syncPatientStats(context.Background(), strconv.FormatInt(userID, 10), nameInSync)
	if errSync != nil {
		log.Printf("WARNING: Failed to sync patient record for user %d: %v", userID, errSync)
		// Fallback to minimal update if sync fails
		existingPatient, errRepo := h.repository.GetPatient(strconv.FormatInt(userID, 10))
		if errRepo == nil {
			patient = existingPatient
			patient.LastVisit = appointmentTime
			patient.TotalVisits++
			h.repository.SavePatient(patient)
		}
	} else {
		log.Printf("Patient record synced for user %d (TotalVisits: %d)", userID, patient.TotalVisits)
		// Record patient loyalty metric
		if patient.TotalVisits <= 1 {
			monitoring.AppointmentTypeTotal.WithLabelValues("first_visit").Inc()
		} else {
			monitoring.AppointmentTypeTotal.WithLabelValues("returning").Inc()
		}

		// Log analytics event
		h.repository.LogEvent(patient.TelegramID, "booking_confirmed", map[string]interface{}{
			"service_id":     service.ID,
			"service_name":   service.Name,
			"time":           appointmentTime.Format(time.RFC3339),
			"is_admin_block": isAdminBlock,
			"visit_count":    patient.TotalVisits,
		})
	}

	// 1. Notify Admin(s)
	for _, adminIDStr := range h.adminIDs {
		adminID, _ := strconv.ParseInt(adminIDStr, 10, 64)
		msg := fmt.Sprintf("🆕 *Новая запись!*\n\nПациент: %s (ID: %s)\nУслуга: %s\nДата: %s\nВремя: %s\nВсего посещений: %d",
			name, patient.TelegramID, service.Name,
			appointmentTime.Format("02.01.2006"),
			appointmentTime.Format("15:04"), patient.TotalVisits)
		if appt.MeetLink != "" {
			msg += fmt.Sprintf("\n\n💻 *Google Meet:* %s", appt.MeetLink)
		}
		h.BotNotify(c.Bot(), adminID, msg)
	}

	// 2. Notify Therapist (Vera)
	if h.therapistID != "" {
		therapistID, _ := strconv.ParseInt(h.therapistID, 10, 64)
		msg := fmt.Sprintf("🆕 *Вера, у вас новая запись!*\n\nПациент: %s\nУслуга: %s\nДата: %s\nВремя: %s",
			name, service.Name,
			appointmentTime.Format("02.01.2006"),
			appointmentTime.Format("15:04"))
		if appt.MeetLink != "" {
			msg += fmt.Sprintf("\n\n💻 *Google Meet:* %s", appt.MeetLink)
		}
		h.BotNotify(c.Bot(), therapistID, msg)
	}

	// Increment booking metric
	monitoring.IncrementBooking(service.Name)

	// Clear session on successful booking
	h.sessionStorage.ClearSession(userID)

	// 3. Confirm to Patient
	confirmationMsg := fmt.Sprintf("✅ <b>Запись подтверждена!</b>\n\n📅 %s\n⏰ %s\n⏳ %s\n\n⚠️ Отмена возможна за 72 часа до приема. Для отмены свяжитесь с терапевтом.",
		appointmentTime.Format("02.01.2006"),
		appointmentTime.Format("15:04"),
		service.Name)
	if appt.MeetLink != "" {
		confirmationMsg += fmt.Sprintf("\n\n💻 <b>Ссылка на Google Meet:</b>\n%s", appt.MeetLink)
	}

	selector := &telebot.ReplyMarkup{}
	url := h.generateWebAppURL(patient.TelegramID)

	if url != "" {
		selector.Inline(
			selector.Row(selector.WebApp("📱 ОТКРЫТЬ МЕД-КАРТУ (LIVE)", &telebot.WebApp{URL: url})),
		)
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

// HandleCancel handles the "Отменить запись" (Cancel booking) button
func (h *BookingHandler) HandleCancel(c telebot.Context) error {
	log.Printf("DEBUG: Entered HandleCancel for user %d", c.Sender().ID)
	userID := c.Sender().ID
	// Clear awaiting confirmation flag
	h.sessionStorage.Set(userID, SessionKeyAwaitingConfirmation, false)
	log.Printf("DEBUG: Cleared SessionKeyAwaitingConfirmation for user %d (via cancel).", userID)

	h.sessionStorage.ClearSession(userID)
	// Remove keyboard and send confirmation
	return c.Send("Запись отменена. Сессия очищена. Вы можете начать /start снова.", h.GetMainMenu())
}

// HandleMyRecords shows patient their records summary
func (h *BookingHandler) HandleMyRecords(c telebot.Context) error {
	userID := c.Sender().ID
	telegramID := strconv.FormatInt(userID, 10)

	patient, err := h.repository.GetPatient(telegramID)
	if err != nil {
		return c.Send(`📊 <b>Ваша медицинская карта</b>

У вас еще нет активной медицинской карты. Она создается автоматически после первого посещения.

Запишитесь на прием через меню бота!`, telebot.ModeHTML)
	}

	card := fmt.Sprintf(`📋 <b>КАРТА ПАЦИЕНТА #%s</b>
──────────────────
👤 <b>ФИО:</b> %s
🔢 <b>ВСЕГО ВИЗИТОВ:</b> %d
💆 <b>ПРОГРАММА:</b> %s

<b>КЛИНИЧЕСКИЕ ЗАМЕТКИ:</b>
<i>%s</i>
──────────────────
📂 <i>Все файлы и анализы доступны в онлайн мед-карте.</i>`,
		patient.TelegramID,
		html.EscapeString(patient.Name),
		patient.TotalVisits,
		html.EscapeString(patient.CurrentService),
		html.EscapeString(patient.TherapistNotes))

	// Compact menu for record management
	selector := &telebot.ReplyMarkup{}
	url := h.generateWebAppURL(patient.TelegramID)

	if url != "" {
		btnWebApp := selector.WebApp("📱 ОТКРЫТЬ МЕД-КАРТУ (LIVE)", &telebot.WebApp{URL: url})
		selector.Inline(
			selector.Row(btnWebApp),
		)
	}

	return c.Send(card, telebot.ModeHTML, selector)
}

// HandleMyAppointments lists user's upcoming appointments
func (h *BookingHandler) HandleMyAppointments(c telebot.Context) error {
	userID := c.Sender().ID
	telegramID := strconv.FormatInt(userID, 10)

	appts, err := h.appointmentService.GetCustomerAppointments(context.Background(), telegramID)
	if err != nil {
		log.Printf("ERROR: Failed to get appointments for user %d: %v", userID, err)
		return c.Send("Ошибка при получении списка ваших записей. Пожалуйста, попробуйте позже.")
	}

	if len(appts) == 0 {
		return c.Send("У вас пока нет активных записей. Вы можете записаться через /start")
	}

	h.sessionStorage.ClearSession(userID)

	var message string = "📋 *Ваши текущие записи:*\n\n"
	selector := &telebot.ReplyMarkup{}
	var rows []telebot.Row
	hasLateAppts := false

	for _, appt := range appts {
		apptTime := appt.StartTime.In(domain.ApptTimeZone)
		message += fmt.Sprintf("🗓 *%s*\n🕒 %s\n💆 %s\n",
			apptTime.Format("02.01.2006"),
			apptTime.Format("15:04"),
			appt.Service.Name)

		// Smart Cancellation Logic: Only show Cancel button if more than 72 hours (3 days) remain
		now := time.Now().In(domain.ApptTimeZone)
		timeRemaining := appt.StartTime.Sub(now)

		if timeRemaining > 72*time.Hour {
			btn := selector.Data(fmt.Sprintf("❌ Отменить %s (%s)", apptTime.Format("02.01"), apptTime.Format("15:04")), "cancel_appt", appt.ID)
			rows = append(rows, selector.Row(btn))
		} else {
			message += "⚠️ _Отмена только через терапевта_\n"
			hasLateAppts = true
		}
		message += "\n"
	}

	if hasLateAppts {
		btnContact := selector.URL("💬 Написать терапевту", "https://t.me/VeraFethiye")
		rows = append(rows, selector.Row(btnContact))
	}

	selector.Inline(rows...)

	// Send with Inline Keyboard ONLY (no Reply Keyboard here to avoid conflicts)
	return c.Send(message, selector, telebot.ParseMode(telebot.ModeMarkdown))
}

// HandleCancelAppointmentCallback handles the specific cancellation of an appointment
func (h *BookingHandler) HandleCancelAppointmentCallback(c telebot.Context) error {
	callbackData := strings.TrimSpace(c.Callback().Data)
	parts := strings.Split(callbackData, "|")
	if len(parts) < 2 {
		return c.Respond(&telebot.CallbackResponse{Text: "Ошибка: неверные данные для отмены."})
	}

	appointmentID := parts[1]
	log.Printf("DEBUG: HandleCancelAppointmentCallback for ID: %s", appointmentID)

	// Get appointment details BEFORE deleting for block check
	appt, _ := h.appointmentService.FindByID(context.Background(), appointmentID)

	if appt != nil {
		now := time.Now().In(domain.ApptTimeZone)
		if appt.StartTime.Sub(now) < 72*time.Hour {
			log.Printf("BLOCKED: Late cancellation attempt for user %s, appt %s", appt.CustomerTgID, appt.ID)
			return c.Respond(&telebot.CallbackResponse{
				Text:      "⛔ До записи меньше 3 дней!\nАвтоматическая отмена невозможна.\nПожалуйста, напишите терапевту напрямую.",
				ShowAlert: true,
			})
		}
	}

	err := h.appointmentService.CancelAppointment(context.Background(), appointmentID)
	if err != nil {
		log.Printf("ERROR: Failed to cancel appointment %s: %v", appointmentID, err)
		return c.Respond(&telebot.CallbackResponse{Text: "Не удалось отменить запись. Возможно, она уже отменена."})
	}

	// Notify admin
	if appt != nil {
		for _, adminIDStr := range h.adminIDs {
			adminID, _ := strconv.ParseInt(adminIDStr, 10, 64)
			h.BotNotify(c.Bot(), adminID, fmt.Sprintf("⚠️ *Запись отменена!*\n\nПациент: %s (ID: %s)\nУслуга: %s\nДата: %s\nВремя: %s",
				appt.CustomerName, appt.CustomerTgID, appt.Service.Name,
				appt.StartTime.In(domain.ApptTimeZone).Format("02.01.2006"),
				appt.StartTime.In(domain.ApptTimeZone).Format("15:04")))
		}

		// Robust sync after cancellation
		h.syncPatientStats(context.Background(), appt.CustomerTgID, appt.CustomerName)
	}

	c.Respond(&telebot.CallbackResponse{Text: "Запись успешно отменена!"})
	c.Edit("✅ Ваша запись успешно отменена и удалена из календаря.")

	return h.HandleMyAppointments(c)
}

// HandleUploadCommand explains how to upload documents
func (h *BookingHandler) HandleUploadCommand(c telebot.Context) error {
	return c.Send(`📤 *Загрузка медицинских документов*

Вы можете отправить мне свои результаты обследований (МРТ, КТ, рентген, анализы) в форматах **PDF, JPG, PNG** или **DICOM (.dcm)**.

*Инструкция:*
1. Просто прикрепите файл или фото к сообщению и отправьте его мне.
2. Я автоматически сохраню его в вашу медицинскую карту.
3. Доктор увидит ваши документы при следующем посещении.

⚠️ *Максимальный размер файла: 50 МБ*`, telebot.ParseMode(telebot.ModeMarkdown))
}

// HandleFileMessage processes incoming documents and photos
func (h *BookingHandler) HandleFileMessage(c telebot.Context) error {
	userID := c.Sender().ID
	telegramID := strconv.FormatInt(userID, 10)

	var fileID string
	var fileName string
	var fileSize int

	msg := c.Message()
	if doc := msg.Document; doc != nil {
		fileID = doc.FileID
		fileName = doc.FileName
		fileSize = int(doc.FileSize)
	} else if photo := msg.Photo; photo != nil {
		fileID = photo.FileID
		fileName = fmt.Sprintf("photo_%d.jpg", time.Now().Unix())
		fileSize = int(photo.FileSize)
	} else if vid := msg.Video; vid != nil {
		fileID = vid.FileID
		fileName = vid.FileName
		if fileName == "" {
			fileName = fmt.Sprintf("video_%d.mp4", time.Now().Unix())
		}
		fileSize = int(vid.FileSize)
	} else if anim := msg.Animation; anim != nil {
		fileID = anim.FileID
		fileName = anim.FileName
		if fileName == "" {
			fileName = fmt.Sprintf("animation_%d.mp4", time.Now().Unix())
		}
		fileSize = int(anim.FileSize)
	} else if voice := msg.Voice; voice != nil {
		fileID = voice.FileID
		fileName = fmt.Sprintf("voice_%d.ogg", time.Now().Unix())
		fileSize = int(voice.FileSize)
	} else {
		return nil // Not a recognized media type
	}

	// 500MB limit for all files
	if fileSize > 500*1024*1024 {
		return c.Send("❌ Файл слишком большой. Максимальный размер: 500 МБ.")
	}

	// Check if patient exists
	patient, err := h.repository.GetPatient(telegramID)
	if err != nil {
		return c.Send("❌ Сначала запишитесь на прием через /start, чтобы я мог создать вашу карту и сохранить документ.")
	}

	statusMsg, err := c.Bot().Send(c.Recipient(), "⏳ Загружаю и сохраняю ваш файл...")
	if err != nil {
		log.Printf("ERROR: Failed to send status message: %v", err)
	}

	// Get file from Telegram servers
	fileReader, err := c.Bot().File(&telebot.File{FileID: fileID})
	if err != nil {
		log.Printf("ERROR: Failed to download file from Telegram: %v", err)
		c.Bot().Delete(statusMsg)
		return c.Send("❌ Ошибка при загрузке файла. Возможно, он слишком большой для Telegram-бота (лимит 50МБ).\n\nПопробуйте отправить файл меньшего размера или ссылкой.")
	}
	defer fileReader.Close()

	// Determine category based on extension/type
	ext := strings.ToLower(filepath.Ext(fileName))
	category := "documents"
	if msg.Voice != nil || msg.Audio != nil {
		category = "messages"
	} else if msg.Photo != nil {
		category = "images"
	} else if ext == ".pdf" || ext == ".doc" || ext == ".docx" {
		category = "scans"
	}

	// Save to storage using Reader for efficiency
	_, err = h.repository.SavePatientDocumentReader(telegramID, fileName, category, fileReader)
	if err != nil {
		log.Printf("ERROR: Failed to save patient document: %v", err)
		c.Bot().Delete(statusMsg)
		return c.Send("❌ Ошибка при сохранении файла на сервере.")
	}

	c.Bot().Delete(statusMsg)

	// Special handling for voice: Transcribe and append to notes
	if voice := msg.Voice; voice != nil {
		transMsg, _ := c.Bot().Send(c.Recipient(), "📝 Расшифровываю ваше аудио-сообщение...")

		// We need a fresh reader or the content of the file
		fileReader, _ := c.Bot().File(&telebot.File{FileID: fileID})
		transcript, err := h.transcriptionService.Transcribe(context.Background(), fileReader, fileName)

		if transMsg != nil {
			c.Bot().Delete(transMsg)
		}

		if err == nil && transcript != "" {
			// Save transcripts to dedicated field instead of clinical notes
			prefix := fmt.Sprintf("\n\n[🎙 %s]: ", time.Now().Format("02.01.2006 15:04"))
			patient.VoiceTranscripts += prefix + transcript
			h.repository.SavePatient(patient)
			c.Send("✅ Аудио расшифровано и сохранено в архиве записей.")
		} else {
			log.Printf("ERROR: Transcription failed for user %d: %v", userID, err)
			c.Send("⚠️ Аудио сохранено, но не удалось его расшифровать.")
		}
	} else {
		c.Send(fmt.Sprintf("✅ Файл '%s' успешно сохранен в вашу медицинскую карту!", fileName))
	}

	// Notify admins with HTML to avoid parsing errors with underscores in filenames
	notification := fmt.Sprintf("📂 <b>Новый файл в мед-карте!</b>\n\nПациент: %s (ID: %s)\nФайл: <code>%s</code>\nРазмер: %.2f MB",
		html.EscapeString(patient.Name),
		html.EscapeString(telegramID),
		html.EscapeString(fileName),
		float64(fileSize)/(1024*1024))

	for _, adminIDStr := range h.adminIDs {
		adminID, _ := strconv.ParseInt(adminIDStr, 10, 64)
		h.BotNotify(c.Bot(), adminID, notification)
	}

	return nil
}

// HandleBackup creates a zip of the data and sends it to the admin
func (h *BookingHandler) HandleBackup(c telebot.Context) error {
	isAdmin := false
	userIDStr := strconv.FormatInt(c.Sender().ID, 10)
	for _, id := range h.adminIDs {
		if id == userIDStr {
			isAdmin = true
			break
		}
	}

	if !isAdmin {
		return c.Send("⛔ У вас нет прав для выполнения этой команды.")
	}

	c.Send("📦 Подготавливаю резервную копию данных...")

	zipPath, err := h.repository.CreateBackup()
	if err != nil {
		log.Printf("ERROR: Failed to create backup: %v", err)
		return c.Send("❌ Ошибка при создании резервной копии.")
	}

	doc := &telebot.Document{
		File:     telebot.FromDisk(zipPath),
		FileName: filepath.Base(zipPath),
		Caption:  fmt.Sprintf("💾 Резервная копия данных от %s", time.Now().Format("02.01.2006 15:04")),
	}

	return c.Send(doc)
}

// BotNotify is a helper to send notifications to admins
func (h *BookingHandler) BotNotify(b *telebot.Bot, to int64, message string) {
	_, err := b.Send(&telebot.User{ID: to}, message, telebot.ModeHTML)
	if err != nil {
		log.Printf("ERROR: Failed to send notification to admin %d: %v", to, err)
	}
}

// HandleBan adds a user to the blacklist
func (h *BookingHandler) HandleBan(c telebot.Context) error {
	if !h.IsAdmin(c.Sender().ID) {
		return c.Send("⛔ Доступ запрещен.")
	}

	args := c.Args()
	if len(args) < 1 {
		return c.Send("Использование: /ban {telegram_id}")
	}

	targetID := args[0]
	if err := h.repository.BanUser(targetID); err != nil {
		return c.Send("❌ Ошибка при блокировке пользователя.")
	}

	return c.Send(fmt.Sprintf("✅ Пользователь %s заблокирован.", targetID))
}

// HandleUnban removes a user from the blacklist
func (h *BookingHandler) HandleUnban(c telebot.Context) error {
	if !h.IsAdmin(c.Sender().ID) {
		return c.Send("⛔ Доступ запрещен.")
	}

	args := c.Args()
	if len(args) < 1 {
		return c.Send("Использование: /unban {telegram_id}")
	}

	targetID := args[0]
	if err := h.repository.UnbanUser(targetID); err != nil {
		return c.Send("❌ Ошибка при разблокировке пользователя.")
	}

	return c.Send(fmt.Sprintf("✅ Пользователь %s разблокирован.", targetID))
}

func (h *BookingHandler) IsAdmin(userID int64) bool {
	userIDStr := strconv.FormatInt(userID, 10)
	for _, id := range h.adminIDs {
		if id == userIDStr {
			return true
		}
	}
	return false
}

// HandleStatus shows bot health and metrics (admin only)
func (h *BookingHandler) HandleStatus(c telebot.Context) error {
	if !h.IsAdmin(c.Sender().ID) {
		return c.Send("❌ Эта команда доступна только администраторам.")
	}

	uptime := time.Since(monitoring.StartTime)
	totalAppts, err := h.appointmentService.GetTotalUpcomingCount(context.Background())
	if err != nil {
		log.Printf("ERROR: Failed to get total upcoming count in status: %v", err)
		totalAppts = 0 // Fallback
	}

	accountInfo, err := h.appointmentService.GetCalendarAccountInfo(context.Background())
	if err != nil {
		log.Printf("ERROR: Failed to get calendar account info: %v", err)
		accountInfo = "Unknown"
	}

	calendarID := h.appointmentService.GetCalendarID()
	allCalendars, _ := h.appointmentService.ListCalendars(context.Background())
	calendarsList := strings.Join(allCalendars, "\n  • ")

	status := fmt.Sprintf(`📊 <b>Статус бота</b>

⏱ <b>Uptime:</b> %s
📈 <b>Метрики:</b>
  • Записей в календаре: %d
  • Сессий с запуска: %d

🔗 <b>Подключения:</b>
  • Account: ✅ %s
  • Calendar ID: <code>%s</code>
  • Telegram API: ✅ OK

📂 <b>Доступные календари:</b>
  • %s`,
		uptime.Round(time.Second),
		totalAppts,
		monitoring.GetTotalBookings(),
		accountInfo,
		calendarID,
		calendarsList,
	)

	return c.Send(status, telebot.ModeHTML)
}

// generateWebAppURL creates a signed URL for the Telegram Web App
func (h *BookingHandler) generateWebAppURL(telegramID string) string {
	if h.webAppURL == "" || h.webAppSecret == "" {
		return ""
	}

	mac := hmac.New(sha256.New, []byte(h.webAppSecret))
	mac.Write([]byte(telegramID))
	token := hex.EncodeToString(mac.Sum(nil))

	return fmt.Sprintf("%s/card?id=%s&token=%s", h.webAppURL, telegramID, token)
}
