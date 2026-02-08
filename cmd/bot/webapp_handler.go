package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/kfilin/massage-bot/internal/domain"
	"github.com/kfilin/massage-bot/internal/logging"
	"github.com/kfilin/massage-bot/internal/ports"
)

// NewWebAppHandler creates the main handler for the WebApp
func NewWebAppHandler(repo ports.Repository, apptService ports.AppointmentService, botToken string, adminIDs []string, secret string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logging.Debugf(" [WebApp]: Incoming Request: %s %s RemoteAddr: %s", r.Method, r.URL.String(), r.RemoteAddr)
		// Prepare paths for query parsing (supports both root and /card)
		id := r.URL.Query().Get("id")
		token := r.URL.Query().Get("token")
		initData := r.URL.Query().Get("initData")

		var finalID string
		var telegramName string

		// 1. Authenticate Request
		// We support two methods: HMAC (legacy/direct links) and InitData (TWA native).
		// We prefer InitData as it is more secure and session-based.

		if initData != "" {
			var err error
			finalID, telegramName, err = validateInitData(initData, botToken)
			if err != nil {
				logging.Errorf("AUTH ERROR: InitData Validation Failed! Err: %v", err)
				// If we have HMAC params, we can still fall back to them
			}
		}

		if finalID == "" && id != "" && token != "" {
			if validateHMAC(id, token, secret) {
				finalID = id
				logging.Debugf(" [WebApp]: Authenticated via legacy HMAC for ID %s", id)
			} else {
				logging.Warnf("AUTH ERROR: HMAC Mismatch for ID %s. Token may be stale.", id)
			}
		}

		if finalID == "" {
			// Serve basic TWA loading page to attempt auth via JS
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			fmt.Fprint(w, `<!DOCTYPE html><html><head><script src="https://telegram.org/js/telegram-web-app.js"></script><style>body{background:#0f172a;color:white;display:flex;justify-content:center;align-items:center;height:100vh;margin:0;font-family:sans-serif;}</style></head><body><div id="status">⏳ Авторизация...</div><script>const tg = window.Telegram.WebApp; tg.expand(); const url = new URL(window.location.href); if (url.searchParams.get('initData')) { document.getElementById('status').innerHTML = "❌ Ошибка проверки данных.<br><small>Попробуйте перезапустить бота.</small>"; } else if(tg.initData) { url.searchParams.set('initData', tg.initData); window.location.replace(url.toString()); } else { document.getElementById('status').innerHTML = "❌ Ошибка авторизации<br><small>Откройте карту через бота</small>"; }</script></body></html>`)
			return
		}

		// Set Auth Cookie for Media Access
		if finalID != "" {
			cookieVal := GenerateAuthCookie(finalID, secret)
			http.SetCookie(w, &http.Cookie{
				Name:     "vera_auth",
				Value:    cookieVal,
				Path:     "/",
				HttpOnly: true,
				Secure:   true,
				SameSite: http.SameSiteNoneMode,
				MaxAge:   86400 * 30, // 30 days
			})
		}

		// 2. Check Admin Status (Important: check this BEFORE we potentially change finalID to a patient ID!)
		isAdmin := false
		authUserID := finalID // Keep track of WHO is logged in
		for _, adminID := range adminIDs {
			if adminID == authUserID {
				isAdmin = true
				break
			}
		}

		// 3. Admin Routing Logic
		if isAdmin {
			targetID := r.URL.Query().Get("id")

			// If we are authenticated via InitData
			if initData != "" {
				if targetID != "" && targetID != authUserID {
					// We are Admin, authenticated via InitData, and requesting to view `targetID`.
					// Allow viewing this patient.
					finalID = targetID
				} else {
					// Admin, no specific target (or target is self) -> Show Search Page
					w.Header().Set("Content-Type", "text/html; charset=utf-8")
					fmt.Fprint(w, repo.GenerateAdminSearchPage())
					return
				}
			} else {
				// HMAC Auth (Development/Legacy)
				// In HMAC flow, the 'id' param is the authenticated user.
				// If that user is admin, show search page.
				w.Header().Set("Content-Type", "text/html; charset=utf-8")
				fmt.Fprint(w, repo.GenerateAdminSearchPage())
				return
			}
		}

		patient, err := repo.GetPatient(finalID)
		if err != nil {
			logging.Infof("Patient %s not found in DB, attempting self-heal from GCal...", finalID)

			name := telegramName
			if name == "" {
				name = "Пациент"
			}

			// Self-heal: Create a basic record and we will populate it below via GCal sync
			patient = domain.Patient{
				TelegramID:     finalID,
				Name:           name,
				HealthStatus:   "initial",
				TherapistNotes: fmt.Sprintf("Зарегистрирован через TWA: %s", time.Now().Format("02.01.2006")),
			}
		}

		// Optimize Speed: Use cached data from DB instead of blocking GCal sync

		// 1. Try to load appointments purely from local DB for performance
		dbAppts, err := repo.GetAppointmentHistory(finalID)
		if err != nil {
			logging.Errorf("DB Error loading history for %s: %v", finalID, err)
			dbAppts = []domain.Appointment{}
		}

		// 2. Smart Sync Logic:
		// If DB is empty -> We MUST sync synchronously (blocking) to show data
		// If DB has data -> We sync asynchronously (non-blocking) to update cache

		var appts []domain.Appointment

		if len(dbAppts) == 0 {
			// EMPTY CACHE: Blocking Sync
			logging.Infof("Cache miss for %s. Performing blocking sync...", finalID)
			fetchedAppts, err := apptService.GetCustomerHistory(r.Context(), finalID)
			if err == nil {
				appts = fetchedAppts
				// Save to DB immediately
				if len(appts) > 0 {
					if err := repo.UpsertAppointments(appts); err != nil {
						logging.Infof("Failed to cache synced appointments: %v", err)
					}
				}
			} else {
				logging.Infof("Failed to fetch history from GCal: %v", err)
			}
		} else {
			// CACHE HIT: Fast Return + Background Sync
			appts = dbAppts
			go func() {
				// Background Sync
				logging.Infof("Background syncing history for %s...", finalID)
				// Create a new context as the request context will be cancelled
				bgCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
				defer cancel()

				fetchedAppts, err := apptService.GetCustomerHistory(bgCtx, finalID)
				if err == nil && len(fetchedAppts) > 0 {
					if err := repo.UpsertAppointments(fetchedAppts); err != nil {
						logging.Infof("Failed to update cache in background: %v", err)
					} else {
						logging.Infof("Background sync successful for %s", finalID)
					}
				}
			}()
		}

		// Recalculate stats based on appts
		if len(appts) > 0 {
			var lastVisit, firstVisit time.Time
			confirmedCount := 0
			for _, a := range appts {
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
			patient.TotalVisits = confirmedCount

			// Save stats back to be safe (optional, but keeps consistency)
			// repo.SavePatient(patient)
		}

		html := repo.GenerateHTMLRecord(patient, appts, isAdmin)
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, html)
	}
}

// NewSearchHandler creates the handler for the Patient Search API
func NewSearchHandler(repo ports.Repository, botToken string, adminIDs []string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		initData := r.Header.Get("X-Telegram-Init-Data")
		if initData == "" {
			initData = r.URL.Query().Get("initData")
		}

		if initData == "" {
			http.Error(w, "Unauthorized: missing initData", http.StatusUnauthorized)
			return
		}

		userID, _, err := validateInitData(initData, botToken)
		if err != nil {
			http.Error(w, "Unauthorized: invalid initData", http.StatusUnauthorized)
			return
		}

		// Check Admin
		isAdmin := false
		for _, id := range adminIDs {
			if id == userID {
				isAdmin = true
				break
			}
		}
		if !isAdmin {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		query := r.URL.Query().Get("q")
		patients, err := repo.SearchPatients(query)
		if err != nil {
			logging.Errorf("Search failed: %v", err)
			http.Error(w, "Search failed", http.StatusInternalServerError)
			return
		}

		type patResult struct {
			TelegramID  string `json:"telegram_id"`
			Name        string `json:"name"`
			TotalVisits int    `json:"total_visits"`
		}
		var results []patResult
		for _, p := range patients {
			results = append(results, patResult{
				TelegramID:  p.TelegramID,
				Name:        p.Name,
				TotalVisits: p.TotalVisits,
			})
		}
		json.NewEncoder(w).Encode(results)
	}
}

// NewCancelHandler creates the handler for Appointment Cancellation
func NewCancelHandler(apptService ports.AppointmentService, botToken string, adminIDs []string) http.HandlerFunc {
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

		notificationMsg := fmt.Sprintf("⚠️ *Запись отменена!*\n\nПациент: %s\nДата: %s", appt.CustomerName, appt.StartTime.Format("02.01 15:04"))
		for _, adminID := range adminIDs {
			sendTelegramMessage(botToken, adminID, notificationMsg)
		}

		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}
}
