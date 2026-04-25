package presentation

import (
	"embed"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"sort"
	"strings"
	"time"

	"github.com/kfilin/massage-bot/internal/domain"
)

//go:embed templates/*
var templatesFS embed.FS

// StaticFS provides access to the templates directory
var StaticFS, _ = fs.Sub(templatesFS, "templates")

type WebPresenter struct {
	templates *template.Template
}

func NewWebPresenter() (*WebPresenter, error) {
	tmpl, err := template.ParseFS(templatesFS, "templates/*.html")
	if err != nil {
		return nil, fmt.Errorf("failed to parse templates: %w", err)
	}
	return &WebPresenter{templates: tmpl}, nil
}

// RenderCard renders the TWA patient card
func (p *WebPresenter) RenderCard(w io.Writer, data interface{}) error {
	return p.templates.ExecuteTemplate(w, "card.html", data)
}

func (p *WebPresenter) RenderSearch(w io.Writer, data interface{}) error {
	return p.templates.ExecuteTemplate(w, "search.html", data)
}

// BotPresenter handles Telegram message formatting
type BotPresenter struct{}

func NewBotPresenter() *BotPresenter {
	return &BotPresenter{}
}

// FormatAppointment formats a clinical concierge-style appointment message
func (p *BotPresenter) FormatAppointment(appt domain.Appointment, isAdmin bool) string {
	var sb strings.Builder
	if isAdmin {
		sb.WriteString("🆕 <b>НОВАЯ ЗАПИСЬ</b>\n")
	} else {
		sb.WriteString("✅ <b>ЗАПИСЬ ПОДТВЕРЖДЕНА</b>\n")
	}
	sb.WriteString("──────────────────\n")
	sb.WriteString(fmt.Sprintf("👤 <b>Пациент:</b> %s\n", appt.CustomerName))
	sb.WriteString(fmt.Sprintf("💆 <b>Услуга:</b> %s\n", appt.Service.Name))
	sb.WriteString(fmt.Sprintf("🕒 <b>Время:</b> %s в %s\n", 
		appt.StartTime.Format("02.01.2006"),
		appt.StartTime.Format("15:04")))
	
	if appt.Duration > 0 {
		sb.WriteString(fmt.Sprintf("⏳ <b>Длительность:</b> %d мин\n", appt.Duration))
	}

	if appt.MeetLink != "" {
		sb.WriteString(fmt.Sprintf("💻 <b>Meet:</b> <a href=\"%s\">Перейти</a>\n", appt.MeetLink))
	}
	
	sb.WriteString("──────────────────\n")
	if !isAdmin {
		sb.WriteString("<i>💡 Приходите за 5 минут до начала. До встречи! 💙</i>")
	}
	
	return sb.String()
}

// FormatCancellation formats a cancellation message
func (p *BotPresenter) FormatCancellation(appt domain.Appointment, isAdmin bool) string {
	var sb strings.Builder
	if isAdmin {
		sb.WriteString("⚠️ <b>ЗАПИСЬ ОТМЕНЕНА</b>\n")
	} else {
		sb.WriteString("🚫 <b>ВАША ЗАПИСЬ ОТМЕНЕНА</b>\n")
	}
	sb.WriteString("──────────────────\n")
	sb.WriteString(fmt.Sprintf("👤 <b>Пациент:</b> %s\n", appt.CustomerName))
	sb.WriteString(fmt.Sprintf("🕒 <b>Было назначено:</b> %s в %s\n", 
		appt.StartTime.Format("02.01.2006"),
		appt.StartTime.Format("15:04")))
	sb.WriteString("──────────────────\n")
	if !isAdmin {
		sb.WriteString("<i>Для выбора другого времени используйте /start</i>")
	}
	return sb.String()
}

// FormatNotification formats a generic clinical notification (e.g. locks, admin actions)
func (p *BotPresenter) FormatNotification(header string, details map[string]string) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("🔔 <b>%s</b>\n", strings.ToUpper(header)))
	sb.WriteString("──────────────────\n")
	
	// Sort keys for consistent output
	keys := make([]string, 0, len(details))
	for k := range details {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		sb.WriteString(fmt.Sprintf("• %s: %s\n", k, details[k]))
	}
	sb.WriteString("──────────────────\n")
	return sb.String()
}

// FormatBookingSummary formats a pre-confirmation booking summary
func (p *BotPresenter) FormatBookingSummary(title string, patientName string, serviceName string, date time.Time, duration int, price float64) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("📖 <b>%s</b>\n", strings.ToUpper(title)))
	sb.WriteString("──────────────────\n")
	sb.WriteString(fmt.Sprintf("👤 <b>Пациент:</b> %s\n", patientName))
	sb.WriteString(fmt.Sprintf("💆 <b>Услуга:</b> %s\n", serviceName))
	sb.WriteString(fmt.Sprintf("🕒 <b>Время:</b> %s в %s\n", date.Format("02.01.2006"), date.Format("15:04")))
	sb.WriteString(fmt.Sprintf("⏳ <b>Длительность:</b> %d мин\n", duration))
	if price > 0 {
		sb.WriteString(fmt.Sprintf("💰 <b>Цена:</b> %.0f ₺\n", price))
	}
	sb.WriteString("──────────────────\n")
	sb.WriteString("<i>Всё верно?</i>")
	return sb.String()
}

// FormatDraftNotification formats the admin review message for transcriptions
func (p *BotPresenter) FormatDraftNotification(patientName, transcript string) string {
	var sb strings.Builder
	sb.WriteString("🎙 <b>ЧЕРНОВИК РАСШИФРОВКИ</b>\n")
	sb.WriteString("──────────────────\n")
	sb.WriteString(fmt.Sprintf("👤 <b>Пациент:</b> %s\n", patientName))
	sb.WriteString(fmt.Sprintf("📝 <b>Текст:</b>\n<i>%s</i>\n", transcript))
	sb.WriteString("──────────────────\n")
	sb.WriteString("📥 <i>Выберите действие ниже:</i>")
	
	return sb.String()
}

// FormatWelcome formats the start message
func (p *BotPresenter) FormatWelcome(name string) string {
	return fmt.Sprintf(`👋 <b>Здравствуйте, %s!</b>

Добро пожаловать в <b>Vera Massage Clinic</b>. 🌿

Я ваш персональный ассистент. Здесь вы можете:
• Записаться на прием
• Посмотреть свою мед-карту
• Загрузить снимки или анализы

<i>Выберите нужное действие в меню ниже.</i>`, name)
}

// FormatPatientCard formats the patient's record summary for the bot
func (p *BotPresenter) FormatPatientCard(patient domain.Patient) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("📋 <b>КАРТА ПАЦИЕНТА #%s</b>\n", patient.TelegramID))
	sb.WriteString("──────────────────\n")
	sb.WriteString(fmt.Sprintf("👤 <b>ФИО:</b> %s\n", patient.Name))
	sb.WriteString(fmt.Sprintf("🔢 <b>ВИЗИТОВ:</b> %d\n", patient.TotalVisits))
	sb.WriteString(fmt.Sprintf("💆 <b>ПРОГРАММА:</b> %s\n\n", patient.CurrentService))
	
	sb.WriteString("<b>КЛИНИЧЕСКИЕ ЗАМЕТКИ:</b>\n")
	if patient.TherapistNotes == "" {
		sb.WriteString("<i>Записей пока нет</i>\n")
	} else {
		// Limit notes length for Telegram
		notes := patient.TherapistNotes
		if len(notes) > 500 {
			notes = notes[:497] + "..."
		}
		sb.WriteString(fmt.Sprintf("<i>%s</i>\n", notes))
	}
	sb.WriteString("──────────────────\n")
	sb.WriteString("📂 <i>Все файлы и анализы доступны в TWA мед-карте.</i>")
	
	return sb.String()
}
