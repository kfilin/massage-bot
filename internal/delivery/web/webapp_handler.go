package web

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/kfilin/massage-bot/internal/domain"
	"github.com/kfilin/massage-bot/internal/logging"
	"github.com/kfilin/massage-bot/internal/ports"
	"github.com/kfilin/massage-bot/internal/presentation"
	"github.com/kfilin/massage-bot/internal/version"
)

// NewWebAppHandler creates the main handler for the WebApp.
// It performs auth (InitData preferred, HMAC fallback), enforces admin
// routing, and renders either the patient card or the admin search page.
func NewWebAppHandler(repo ports.Repository, apptService ports.AppointmentService, presenter *presentation.WebPresenter, botToken string, adminIDs []string, secret string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logging.Debugf(" [WebApp]: Incoming Request: %s %s RemoteAddr: %s", r.Method, r.URL.String(), r.RemoteAddr)
		// Prepare paths for query parsing (supports both root and /card)
		id := r.URL.Query().Get("id")
		ts := r.URL.Query().Get("ts")
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
			if validateHMAC(id, ts, token, secret) {
				finalID = id
				logging.Debugf(" [WebApp]: Authenticated via HMAC for ID %s (ts=%s)", id, ts)
			} else {
				logging.Warnf("AUTH ERROR: HMAC Mismatch for ID %s. Token may be stale or expired.", id)
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
				MaxAge:   86400, // 24 hours — clinical data re-auth daily
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
					data := struct {
						BotUsername string
					}{
						BotUsername: os.Getenv("BOT_USERNAME"), // We'll need to make sure this is available
					}
					if err := presenter.RenderSearch(w, data); err != nil {
						logging.Errorf("Search template rendering failed: %v", err)
					}
					return
				}
			} else {
				// HMAC Auth (Development/Legacy)
				// In HMAC flow, the 'id' param is the authenticated user.
				// If that user is admin, show search page.
				w.Header().Set("Content-Type", "text/html; charset=utf-8")
				data := struct{ BotUsername string }{BotUsername: os.Getenv("BOT_USERNAME")}
				if err := presenter.RenderSearch(w, data); err != nil {
					logging.Errorf("Search template rendering failed: %v", err)
				}
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
				logging.Errorf("CRITICAL SYNC ERROR for %s: %v", finalID, err)
				// If sync fails completely and we have NO data, show a slightly different page or a warning
				// For now, we continue and GenerateHTMLRecord will show an empty list
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

		// Fetch Media (including Drafts)
		allMedia, _ := repo.GetPatientMedia(finalID)
		var drafts []map[string]interface{}

		// Grouping Logic for Files Tab
		groups := make(map[string]*struct {
			Name  string
			Count int
			Files []domain.PatientMedia
		})

		initGroup := func(name string) {
			groups[name] = &struct {
				Name  string
				Count int
				Files []domain.PatientMedia
			}{Name: name}
		}
		initGroup("Снимки")
		initGroup("Фотографии")
		initGroup("Видео")
		initGroup("Голосовые заметки")
		initGroup("Прочее")

		for _, m := range allMedia {
			// 1. Separate Drafts (pending voice only)
			if m.FileType == "voice" && m.Status == "pending" {
				drafts = append(drafts, map[string]interface{}{
					"ID":         m.ID,
					"Transcript": m.Transcript,
					"Date":       m.CreatedAt.Format("02.01 15:04"),
				})
				continue // Don't show drafts in "Files" yet
			}

			// 2. Populate DocGroups (approved OR discarded)
			// Discarded files are still accessible as raw media
			var targetGroup string
			switch m.FileType {
			case "scan":
				targetGroup = "Снимки"
			case "photo", "image":
				targetGroup = "Фотографии"
			case "voice", "audio":
				targetGroup = "Голосовые заметки"
			case "video":
				targetGroup = "Видео"
			default:
				targetGroup = "Прочее"
			}

			if g, ok := groups[targetGroup]; ok {
				g.Count++
				g.Files = append(g.Files, m)
			}
		}

		// Prepare ordered list of populated groups
		var docGroups []interface{}
		order := []string{"Снимки", "Фотографии", "Видео", "Голосовые заметки", "Прочее"}
		for _, name := range order {
			if g := groups[name]; g != nil && g.Count > 0 {
				docGroups = append(docGroups, g)
			}
		}

		// Prepare Template Data
		data := struct {
			Title        string
			Patient      domain.Patient
			RecentVisits []domain.Appointment
			Drafts       []map[string]interface{}
			DocGroups    []interface{}
			BotVersion   string
			IsAdmin      bool
		}{
			Title:        "Карта пациента",
			Patient:      patient,
			RecentVisits: appts,
			Drafts:       drafts,
			DocGroups:    docGroups,
			BotVersion:   version.FullName,
			IsAdmin:      isAdmin,
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := presenter.RenderCard(w, data); err != nil {
			logging.Errorf("Template rendering failed: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	}
}
