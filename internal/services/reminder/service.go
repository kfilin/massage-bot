package reminder

import (
	"context"
	"strconv"
	"time"

	"github.com/kfilin/massage-bot/internal/logging"
	"github.com/kfilin/massage-bot/internal/presentation"

	"github.com/kfilin/massage-bot/internal/domain"
	"github.com/kfilin/massage-bot/internal/ports"
	"gopkg.in/telebot.v3"
)

// BotSender is a minimal interface for sending Telegram messages.
// *telebot.Bot satisfies this interface automatically.
type BotSender interface {
	Send(to telebot.Recipient, what interface{}, opts ...interface{}) (*telebot.Message, error)
}

type Service struct {
	apptService ports.AppointmentService
	repo        ports.Repository
	bot         BotSender
	adminIDs    []string
	presenter   *presentation.BotPresenter
}

func NewService(as ports.AppointmentService, repo ports.Repository, bot BotSender, adminIDs []string, p *presentation.BotPresenter) *Service {
	return &Service{
		apptService: as,
		repo:        repo,
		bot:         bot,
		adminIDs:    adminIDs,
		presenter:   p,
	}
}

func (s *Service) Start(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Minute)
	logging.Infof("Reminder Service started.")

	go func() {
		for {
			select {
			case <-ticker.C:
				s.ScanAndSendReminders(ctx)
			case <-ctx.Done():
				ticker.Stop()
				return
			}
		}
	}()
}

func (s *Service) ScanAndSendReminders(ctx context.Context) {
	logging.Infof("Scanning for appointments to send reminders...")

	now := time.Now().In(domain.ApptTimeZone)
	// Scan window: from now up to 73 hours ahead (to catch 72h + 24h)
	timeMax := now.Add(73 * time.Hour)

	appts, err := s.apptService.GetUpcomingAppointments(ctx, now, timeMax)
	if err != nil {
		logging.Errorf(": Failed to fetch upcoming appointments for reminders: %v", err)
		return
	}

	for _, appt := range appts {
		if appt.CustomerTgID == "" {
			continue
		}

		// Skip if cancelled
		if appt.Status == "cancelled" {
			continue
		}

		// Enforce timezone just in case
		apptTime := appt.StartTime.In(domain.ApptTimeZone)
		timeToAppt := apptTime.Sub(now)

		// 1. 72h Reminder (3 days)
		if timeToAppt <= 72*time.Hour && timeToAppt > 71*time.Hour {
			s.sendReminder(appt, "72h")
		}

		// 2. 24h Reminder (1 day)
		if timeToAppt <= 24*time.Hour && timeToAppt > 23*time.Hour {
			s.sendReminder(appt, "24h")
		}
	}
}

func (s *Service) sendReminder(appt domain.Appointment, reminderType string) {
	// Check if already sent
	confirmedAt, sentMap, err := s.repo.GetAppointmentMetadata(appt.ID)
	if err == nil && sentMap[reminderType] {
		return
	}

	// If already confirmed, don't send 24h reminder (or send a different one)
	if confirmedAt != nil && reminderType == "24h" {
		// Could send a "See you tomorrow!" message instead
		return
	}

	userID, _ := strconv.ParseInt(appt.CustomerTgID, 10, 64)
	user := &telebot.User{ID: userID}

	var msg string
	var menu *telebot.ReplyMarkup

	// Use presenter for clinical style
	msg = s.presenter.FormatAppointment(appt, false)
	if reminderType == "72h" {
		msg = "🔔 <b>Напоминание (за 3 дня)</b>\n" + msg + "\n\n<i>Пожалуйста, подтвердите ваше присутствие.</i>"
	} else if reminderType == "24h" {
		msg = "🔔 <b>Напоминание (завтра)</b>\n" + msg + "\n\n<i>Пожалуйста, подтвердите ваше присутствие.</i>"
	}

	menu = &telebot.ReplyMarkup{}
	btnConfirm := menu.Data("✅ Подтвердить", "confirm_appt_reminder", appt.ID)
	btnCancel := menu.Data("❌ Отменить", "cancel_appt_reminder", appt.ID)
	menu.Inline(menu.Row(btnConfirm, btnCancel))

	_, err = s.bot.Send(user, msg, telebot.ModeHTML, menu)
	if err != nil {
		logging.Errorf(": Failed to send %s reminder to patient %s: %v", reminderType, appt.CustomerTgID, err)
		return
	}

	// Persist sent status
	if sentMap == nil {
		sentMap = make(map[string]bool)
	}
	sentMap[reminderType] = true
	if err := s.repo.SaveAppointmentMetadata(appt.ID, confirmedAt, sentMap); err != nil {
		logging.Errorf("Failed to save appointment metadata: %v", err)
	}
}
