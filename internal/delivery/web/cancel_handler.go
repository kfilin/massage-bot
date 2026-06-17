package web

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/kfilin/massage-bot/internal/domain"
	"github.com/kfilin/massage-bot/internal/ports"
	"github.com/kfilin/massage-bot/internal/presentation"
)

// NewCancelHandler creates the handler for Appointment Cancellation.
// Admins can cancel any appointment; patients can cancel their own with
// at least 72 hours notice. Sends a Telegram notification to all admins
// on success.
func NewCancelHandler(apptService ports.AppointmentService, botToken string, adminIDs []string, presenter *presentation.BotPresenter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			_ = json.NewEncoder(w).Encode(map[string]string{"status": "error", "error": "Method not allowed"})
			return
		}

		var reqBody struct {
			InitData string `json:"initData"`
			ApptID   string `json:"apptId"`
		}
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{"status": "error", "error": "Invalid request"})
			return
		}

		if reqBody.InitData == "" || reqBody.ApptID == "" {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{"status": "error", "error": "Missing parameters"})
			return
		}

		userID, _, err := validateInitData(reqBody.InitData, botToken)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(map[string]string{"status": "error", "error": "Сессия недействительна."})
			return
		}

		isAdmin := false
		for _, adminID := range adminIDs {
			if adminID == userID {
				isAdmin = true
				break
			}
		}

		appt, err := apptService.FindByID(r.Context(), reqBody.ApptID)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]string{"status": "error", "error": "Запись не найдена"})
			return
		}

		if !isAdmin && appt.CustomerTgID != userID {
			w.WriteHeader(http.StatusForbidden)
			_ = json.NewEncoder(w).Encode(map[string]string{"status": "error", "error": "Доступ запрещен"})
			return
		}

		now := time.Now().In(domain.ApptTimeZone)
		if !isAdmin && appt.StartTime.Sub(now) < 72*time.Hour {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"status": "error",
				"error":  "До приема менее 72 часов. Напишите терапевту.",
			})
			return
		}

		err = apptService.CancelAppointment(r.Context(), reqBody.ApptID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(map[string]string{"status": "error", "error": "Не удалось отменить запись"})
			return
		}

		notificationMsg := presenter.FormatCancellation(appt, true)
		for _, adminID := range adminIDs {
			sendTelegramMessage(botToken, adminID, notificationMsg)
		}

		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}
}
