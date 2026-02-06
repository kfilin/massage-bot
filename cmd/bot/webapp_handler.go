package main

import (
	"context"
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
			fmt.Fprint(w, `<!DOCTYPE html><html><head><script src="https://telegram.org/js/telegram-web-app.js"></script><style>body{background:#0f172a;color:white;display:flex;justify-content:center;align-items:center;height:100vh;margin:0;font-family:sans-serif;}</style></head><body><div id="status">⏳ Авторизация...</div><script>const tg = window.Telegram.WebApp; tg.expand(); if(tg.initData) { const url = new URL(window.location.href); url.searchParams.set('initData', tg.initData); window.location.replace(url.toString()); } else { document.getElementById('status').innerHTML = "❌ Ошибка авторизации<br><small>Откройте карту через бота</small>"; }</script></body></html>`)
			return
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
