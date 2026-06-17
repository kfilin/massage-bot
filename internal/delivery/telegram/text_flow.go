package telegram

import (
	"fmt"
	"strconv"
	"time"

	"github.com/kfilin/massage-bot/internal/delivery/telegram/handlers"
	"github.com/kfilin/massage-bot/internal/domain"
	"github.com/kfilin/massage-bot/internal/logging"
	"github.com/kfilin/massage-bot/internal/ports"
	"gopkg.in/telebot.v3"
)

// handleAdminReply delivers the admin's text reply to the patient they were
// addressing, logs the exchange to the patient's Med-Card, and clears the
// admin-reply session state. Called from the OnText handler when the router
// resolves TextActionAdminReply.
//
// Side effects: sends a message via the bot, writes to Med-Card, mutates
// session. Not unit-tested directly because it requires a real *telebot.Bot.
func handleAdminReply(
	c telebot.Context,
	b *telebot.Bot,
	repo ports.Repository,
	sessionStorage ports.SessionStorage,
	adminUserID int64,
	text string,
) error {
	replyingToID := sessionString(sessionStorage.Get(adminUserID), handlers.SessionKeyAdminReplyingTo)
	if replyingToID == "" {
		logging.Warnf("handleAdminReply invoked with empty replyingToID for admin %d", adminUserID)
		return c.Send("Не удалось определить пациента для ответа.")
	}

	logging.Debugf("DEBUG: handleAdminReply: Admin %d is replying to patient %s.", adminUserID, replyingToID)

	patientID, _ := strconv.ParseInt(replyingToID, 10, 64)
	patientUser := &telebot.User{ID: patientID}

	replyMsg := fmt.Sprintf("📩 <b>Сообщение от Веры:</b>\n\n%s", text)
	if _, err := b.Send(patientUser, replyMsg, telebot.ModeHTML); err != nil {
		logging.Errorf("ERROR: Failed to deliver admin reply to patient %s: %v", replyingToID, err)
		return c.Send("❌ Не удалось доставить сообщение пациенту.")
	}

	if patient, err := repo.GetPatient(replyingToID); err == nil {
		prefix := fmt.Sprintf("\n\n[👩‍⚕️ Вера %s]: ", time.Now().In(domain.ApptTimeZone).Format("02.01.2006 15:04"))
		patient.TherapistNotes += prefix + text
		if saveErr := repo.SavePatient(patient); saveErr != nil {
			logging.Errorf("Failed to save admin reply to patient record: %v", saveErr)
		}
	}

	sessionStorage.Set(adminUserID, handlers.SessionKeyAdminReplyingTo, nil)
	return c.Send("✅ Сообщение доставлено и сохранено в мед-карте.")
}

// forwardPatientMessageToAdmins handles a free-text patient message that
// arrives while the booking session is fully populated (service + name set).
// It acknowledges the patient, persists the exchange to the Med-Card, and
// notifies every configured admin with a deep-link to the Med-Card and a
// "Reply" inline button.
//
// Side effects: sends messages via the bot, writes to Med-Card. Not
// unit-tested directly because it requires a real *telebot.Bot.
func forwardPatientMessageToAdmins(
	c telebot.Context,
	b *telebot.Bot,
	repo ports.Repository,
	bookingHandler *handlers.BookingHandler,
	adminIDs []string,
	text string,
) error {
	if err := c.Send("Ваше сообщение получено и передано Вере."); err != nil {
		logging.Warnf("Failed to send confirmation to patient: %v", err)
	}

	telegramID := strconv.FormatInt(c.Sender().ID, 10)
	customerName := c.Sender().FirstName + " " + c.Sender().LastName
	if c.Sender().Username != "" {
		customerName += " (@" + c.Sender().Username + ")"
	}

	notification := fmt.Sprintf(
		"📩 <b>Новое сообщение от пациента!</b>\n\n<b>Пациент:</b> %s (ID: %s)\n<b>Текст:</b> %s",
		customerName, telegramID, text,
	)

	if patient, err := repo.GetPatient(telegramID); err == nil {
		prefix := fmt.Sprintf("\n\n[💬 Пациент %s]: ", time.Now().In(domain.ApptTimeZone).Format("02.01.2006 15:04"))
		patient.TherapistNotes += prefix + text
		if saveErr := repo.SavePatient(patient); saveErr != nil {
			logging.Errorf("Failed to save patient message: %v", saveErr)
		}
	}

	selector := &telebot.ReplyMarkup{}
	btnReply := selector.Data("✍️ Ответить", "admin_reply", telegramID)

	if bookingHandler.WebAppURL != "" {
		cardURL := bookingHandler.GenerateWebAppURL(telegramID)
		notification += fmt.Sprintf("\n\n📄 <a href=\"%s\">Открыть мед-карту</a>", cardURL)
	}
	selector.Inline(selector.Row(btnReply))

	for _, adminIDStr := range adminIDs {
		adminID, _ := strconv.ParseInt(adminIDStr, 10, 64)
		_, _ = b.Send(&telebot.User{ID: adminID}, notification, telebot.ModeHTML, selector)
	}
	return nil
}
