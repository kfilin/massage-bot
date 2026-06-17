package handlers

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/kfilin/massage-bot/internal/domain"
	"github.com/kfilin/massage-bot/internal/logging"
	"github.com/kfilin/massage-bot/internal/monitoring"
	"gopkg.in/telebot.v3"
)

// Admin-only methods on BookingHandler. These commands are gated by IsAdmin
// and assume h.adminIDs is populated.
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

func (h *BookingHandler) HandleManualAppointment(c telebot.Context) error {
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

	h.sessionStorage.ClearSession(userID)
	h.sessionStorage.Set(userID, SessionKeyIsAdminManual, true)

	// If name is provided directly in command arguments, store it
	if len(c.Args()) > 0 {
		nameFromArgs := strings.Join(c.Args(), " ")
		h.sessionStorage.Set(userID, SessionKeyName, nameFromArgs)
		logging.Debugf(": Manual appointment name captured from args: %s", nameFromArgs)
	}

	return h.showCategories(c)
}

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

	if err := c.Send("📦 Подготавливаю резервную копию данных..."); err != nil {
		logging.Warnf("Failed to send backup status: %v", err)
	}

	zipPath, err := h.repository.CreateBackup()
	if err != nil {
		logging.Errorf(": Failed to create backup: %v", err)
		return c.Send("❌ Ошибка при создании резервной копии.")
	}

	doc := &telebot.Document{
		File:     telebot.FromDisk(zipPath),
		FileName: filepath.Base(zipPath),
		Caption:  fmt.Sprintf("💾 Резервная копия данных от %s", time.Now().Format("02.01.2006 15:04")),
	}

	err = c.Send(doc)
	// Cleanup temporary zip
	os.Remove(zipPath)
	return err
}

func (h *BookingHandler) BotNotify(b *telebot.Bot, to int64, message string, opts ...interface{}) {
	_, err := b.Send(&telebot.User{ID: to}, message, append([]interface{}{telebot.ModeHTML}, opts...)...)
	if err != nil {
		logging.Errorf(": Failed to send notification to admin %d: %v", to, err)
	}
}

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

func (h *BookingHandler) HandleStatus(c telebot.Context) error {
	if !h.IsAdmin(c.Sender().ID) {
		return c.Send("❌ Эта команда доступна только администраторам.")
	}

	uptime := time.Since(monitoring.StartTime)
	totalAppts, err := h.appointmentService.GetTotalUpcomingCount(context.Background())
	if err != nil {
		logging.Errorf(": Failed to get total upcoming count in status: %v", err)
		totalAppts = 0 // Fallback
	}

	accountInfo, err := h.appointmentService.GetCalendarAccountInfo(context.Background())
	if err != nil {
		logging.Errorf(": Failed to get calendar account info: %v", err)
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

func (h *BookingHandler) generateGoogleCalendarLink(appt domain.Appointment) string {
	// Format: YYYYMMDDTHHMMSS
	start := appt.StartTime.In(domain.ApptTimeZone).Format("20060102T150405")
	end := appt.StartTime.Add(time.Duration(appt.Duration) * time.Minute).In(domain.ApptTimeZone).Format("20060102T150405")

	title := fmt.Sprintf("Массаж: %s", appt.Service.Name)
	details := fmt.Sprintf("Услуга: %s\nАдрес: Fethiye, Turkey\n\n(Generated by Vera Massage Bot)", appt.Service.Name)
	location := "Fethiye, Turkey"

	baseURL := "https://calendar.google.com/calendar/render"
	params := url.Values{}
	params.Add("action", "TEMPLATE")
	params.Add("text", title)
	params.Add("dates", start+"/"+end)
	params.Add("details", details)
	params.Add("location", location)
	params.Add("ctz", "Europe/Istanbul") // Explicit timezone

	return baseURL + "?" + params.Encode()
}

func (h *BookingHandler) HandleEditName(c telebot.Context) error {
	if !h.IsAdmin(c.Sender().ID) {
		return c.Send("⛔ Доступ запрещен.")
	}

	args := c.Args()
	if len(args) < 2 {
		return c.Send("Использование: /edit_name {telegram_id} {Новое Имя}")
	}

	targetID := args[0]
	newName := strings.Join(args[1:], " ")

	// 1. Check if patient exists
	patient, err := h.repository.GetPatient(targetID)
	if err != nil {
		return c.Send("❌ Пациент не найден в базе данных.")
	}

	oldName := patient.Name
	patient.Name = newName

	// 2. Save (Updates DB and Markdown)
	if err := h.repository.SavePatient(patient); err != nil {
		logging.Errorf(": Failed to save updated patient name: %v", err)
		return c.Send("❌ Ошибка при сохранении данных.")
	}

	logging.Infof("[ADMIN] Name updated for %s: %s -> %s", targetID, oldName, newName)
	return c.Send(fmt.Sprintf("✅ Имя пациента обновлено:\n<b>Old:</b> %s\n<b>New:</b> %s", oldName, newName), telebot.ModeHTML)
}

// GenerateWebAppURL creates a signed URL for the Telegram Web App.
// Includes a unix timestamp in both the URL and the HMAC payload, so links
// expire after 7 days (configured via web.hmacMaxAge) and the ts cannot be
// rolled without invalidating the signature.
func (h *BookingHandler) GenerateWebAppURL(telegramID string) string {
	if h.WebAppURL == "" || h.webAppSecret == "" {
		return ""
	}

	// Rigidly enforce HTTPS for Telegram compatibility
	url := h.WebAppURL
	if !strings.HasPrefix(url, "https://") && !strings.HasPrefix(url, "http://") {
		url = "https://" + url
	} else if strings.HasPrefix(url, "http://") {
		url = strings.Replace(url, "http://", "https://", 1)
	}

	ts := strconv.FormatInt(time.Now().Unix(), 10)
	mac := hmac.New(sha256.New, []byte(h.webAppSecret))
	mac.Write([]byte(strings.TrimSpace(telegramID) + ":" + ts))
	token := hex.EncodeToString(mac.Sum(nil))

	logging.Infof("[URL_GEN] ID: %s, TS: %s, SecretLen: %d, Token: %s", telegramID, ts, len(h.webAppSecret), token)

	return fmt.Sprintf("%s/card?id=%s&ts=%s&token=%s", url, telegramID, ts, token)
}

func (h *BookingHandler) HandleListPatients(c telebot.Context) error {
	if !h.IsAdmin(c.Sender().ID) {
		return c.Send("⛔ Доступ запрещен.")
	}

	// Search with empty query to get recents
	patients, err := h.repository.SearchPatients("")
	if err != nil {
		logging.Errorf("Failed to list patients: %v", err)
		return c.Send("❌ Ошибка при получении списка пациентов.")
	}

	if len(patients) == 0 {
		return c.Send("Список пациентов пуст.")
	}

	var sb strings.Builder
	sb.WriteString("📋 <b>Список пациентов:</b>\n\n")

	for i, p := range patients {
		name := p.Name
		if name == "" {
			name = "Без имени"
		}
		// Clean name for HTML
		name = strings.ReplaceAll(name, "<", "&lt;")
		name = strings.ReplaceAll(name, ">", "&gt;")

		sb.WriteString(fmt.Sprintf("%d. <b>%s</b> (ID: <code>%s</code>)\nVisits: %d\n", i+1, name, p.TelegramID, p.TotalVisits))

		webAppURL := h.GenerateWebAppURL(p.TelegramID)
		if webAppURL != "" {
			sb.WriteString(fmt.Sprintf("<a href=\"%s\">🔗 Открыть Карту</a>\n", webAppURL))
		}
		sb.WriteString("\n")
	}

	return c.Send(sb.String(), telebot.ModeHTML)
}

