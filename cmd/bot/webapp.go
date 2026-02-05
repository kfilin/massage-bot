package main

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/kfilin/massage-bot/internal/logging"

	"github.com/kfilin/massage-bot/internal/domain"
	"github.com/kfilin/massage-bot/internal/ports"
	"golang.org/x/net/webdav"
)

func generateHMAC(id string, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(id))
	return hex.EncodeToString(h.Sum(nil))
}

func validateHMAC(id string, token string, secret string) bool {
	expected := generateHMAC(id, secret)
	match := hmac.Equal([]byte(token), []byte(expected))
	if !match {
		logging.Debugf(" [validateHMAC]: Mismatch for ID=%s. Provided=%s, Expected=%s, SecretLen=%d", id, token, expected, len(secret))
	}
	return match
}

// validateInitData validates Telegram WebApp initData
func validateInitData(initData string, botToken string) (string, string, error) {
	values, err := url.ParseQuery(initData)
	if err != nil {
		return "", "", err
	}

	hash := values.Get("hash")
	if hash == "" {
		return "", "", fmt.Errorf("missing hash")
	}
	values.Del("hash")

	// Sort keys
	var keys []string
	for k := range values {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Build data check string
	var dataCheckArr []string
	for _, k := range keys {
		dataCheckArr = append(dataCheckArr, fmt.Sprintf("%s=%s", k, values.Get(k)))
	}
	dataCheckString := strings.Join(dataCheckArr, "\n")

	// Calculate HMAC
	// Step 1: secret_key = HMAC_SHA256("WebAppData", botToken)
	h1 := hmac.New(sha256.New, []byte("WebAppData"))
	h1.Write([]byte(botToken))
	secretKey := h1.Sum(nil)

	// Step 2: result = HMAC_SHA256(secret_key, dataCheckString)
	h2 := hmac.New(sha256.New, secretKey)
	h2.Write([]byte(dataCheckString))
	expectedHash := hex.EncodeToString(h2.Sum(nil))

	if expectedHash != hash {
		return "", "", fmt.Errorf("hash mismatch")
	}

	// Extract user data
	userJSON := values.Get("user")
	if userJSON == "" {
		return "", "", fmt.Errorf("missing user data")
	}

	var user struct {
		ID        int64  `json:"id"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
	}
	if err := json.Unmarshal([]byte(userJSON), &user); err != nil {
		return "", "", err
	}

	fullName := strings.TrimSpace(user.FirstName + " " + user.LastName)
	if fullName == "" {
		fullName = "Пациент"
	}

	return fmt.Sprintf("%d", user.ID), fullName, nil
}

func startWebAppServer(ctx context.Context, port string, secret string, botToken string, adminIDs []string, repo ports.Repository, apptService ports.AppointmentService, dataDir string) {
	if port == "" {
		port = "8082"
	}

	if dataDir == "" {
		dataDir = "data"
	}

	mux := http.NewServeMux()

	// Handle both root and /card with the same logic
	handler := func(w http.ResponseWriter, r *http.Request) {
		logging.Debugf(" [WebApp]: Incoming Request: %s %s RemoteAddr: %s", r.Method, r.URL.String(), r.RemoteAddr)
		// Prepare paths for query parsing (supports both root and /card)
		id := r.URL.Query().Get("id")
		token := r.URL.Query().Get("token")
		initData := r.URL.Query().Get("initData")

		var finalID string
		var telegramName string

		// 1. Authenticate Request
		if id != "" && token != "" {
			if !validateHMAC(id, token, secret) {
				expected := generateHMAC(id, secret)
				logging.Errorf("AUTH ERROR: HMAC Mismatch! ID=%s, Token=%s, Expected=%s, SecretLen=%d", id, token, expected, len(secret))
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				return
			}
			finalID = id
		} else if initData != "" {
			var err error
			finalID, telegramName, err = validateInitData(initData, botToken)
			if err != nil {
				logging.Errorf("AUTH ERROR: InitData Validation Failed! Err: %v\nInitData: %s\nBotTokenPrefix: %s...", err, initData, botToken[:5])
				http.Error(w, "Authentication failed", http.StatusUnauthorized)
				return
			}
		} else {
			// Serve basic TWA loading page
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, `<!DOCTYPE html><html><head><script src="https://telegram.org/js/telegram-web-app.js"></script></head><body><script>const tg = window.Telegram.WebApp; if(tg.initData) { const url = new URL(window.location.href); url.searchParams.set('initData', tg.initData); window.location.href = url.toString(); } else { document.body.innerHTML = "❌ Ошибка авторизации"; }</script></body></html>`)
			return
		}

		// 2. Check Admin Status
		isAdmin := false
		for _, adminID := range adminIDs {
			if adminID == finalID {
				isAdmin = true
				break
			}
		}

		// 3. Admin Routing Logic
		if isAdmin {

			// If we arrived via initData (no explicit ID param in URL, just auth), or if ID matches admin ID
			// AND we are not explicitly trying to "view self" via some specific link
			// Then show search page.
			// However, standard flow uses ?initData=... and no ID param usually.
			// But the code above sets finalID from id param if valid HMAC present.

			// Simplified: If 'id' query param is NOT present (or empty), check if we want to show search.
			// But 'id' might be used for HMAC auth.
			// Let's rely on a specific 'view' param or lack of target.

			// Case A: HMAC Auth (Legacy/Dev) - usually has ?id=...&token=...
			// In this case finalID is set. If finalID == Admin, do we show Search?
			// Probably yes, unless they want to see their own card.

			// Case B: InitData Auth - usually no ?id=... initially.

			// Logic: If query param 'id' is PRESENT and DIFFERENT from finalID (Admin acting as User) -> Show User.
			// logic failure: HMAC uses ?id=ADMIN_ID.

			// Better Logic:
			// If query param 'view_user_id' is set -> Show that user.
			// Else if URL param 'id' matches 'finalID' (or missing) -> Show Search Page.

			// But wait, the previous code used 'id' query param as THE target ID if HMAC valid.
			// If I am admin, and I click a link ?id=PATIENT_ID&token=..., validation fails because token is for ADMIN_ID?
			// No, HMAC is ID+Secret. So I can only generate links for myself unless I know the secret.
			// So `id` param in HMAC flow IS the authenticated user.

			// So:
			// 1. Authenticated User = finalID.
			// 2. If Admin, and no 'id' param provided in URL (meaning we are identifying via initData) -> Show Search.
			// 3. If Admin, and 'id' param provided MATCHES finalID -> Show Search.
			// 4. If Admin wants to view a patient, they need to supply target ID.
			//    But we need to distinguish "I am Admin X" from "I want to view Patient Y".
			//    Let's use the 'id' param as the TARGET view if we are already authenticated via initData.
			//    But wait, initData is stripped in some flows?

			// Let's settle on:
			// If Admin:
			//   Check 'id' param.
			//   If 'id' != finalID -> We are viewing someone else (assuming we allow admins to view anyone without token if auth via initData).
			//   If 'id' == finalID or empty -> Show Search Page.

			// BUT: The existing code uses 'id' path for HMAC check.
			// Let's stick to: Authenticated User is `finalID`.
			// The `id` query param is used for auth if token present.
			// If we use TWA `initData`, `id` param might be used for Target.

			targetID := r.URL.Query().Get("id")

			// If we are authenticated via InitData
			if initData != "" {
				if targetID != "" && targetID != finalID {
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
				// If id == finalID (Admin), show Search
				// But we might want to debug specific user.
				// Let's just say for HMAC, if Admin -> Search.
				// If we want to view user via HMAC, we must generate HMAC for THAT user.
				// So if I authenticate as Admin via HMAC -> Search Page.
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

		html := repo.GenerateHTMLRecord(patient, appts)
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, html)
	}

	mux.HandleFunc("/", handler)
	mux.HandleFunc("/card", handler)

	// API: Search Patients (Admin Only)
	mux.HandleFunc("/api/search", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// Auth check (InitData only for API)
		// We could support HMAC if needed, but TWA passes InitData cleaner
		// Actually, let's reuse query param extraction logic or simplify.
		// For fetch() we might not easily pass initData in headers without custom logic.
		// The simple way is: The Fetch request should include ?initData=... or check Referer/Cookie?
		// TWA doesn't set cookies.
		// We must inspect the request.

		// NOTE: Our frontend search script calls '/api/search?q=...'
		// It inherits the initData from the window location if we were redirected?
		// No, fetch() does NOT automatically send window.location parameters.
		// We need to fix the template to include initData in the fetch call.
		// BUT: I can't fix the template right now in this step easily without another tool call.
		// Wait, I just edited the template in the previous step. Is it sending initData?
		// `fetch('/api/search?q=' + encodeURIComponent(query))` -> NO initData sent.

		// Quick Fix: Assuming the session cookie or we skip auth for this iteration?
		// NO, unsafe.
		// I must fix the template to send initData.
		// AND checking initData here.
		// If I cannot fix the template now, I will break it.
		// I MUST fix the template.

		// However, I can implement unsafe search temporarily or use a secret token if I had one.
		// I will assume for now I can read the Referer or rely on "Sec-Fetch-Site" but that's weak.

		// CORRECT APPROACH:
		// I will modify the template in the NEXT step (or same step if I could).
		// But I am in `webapp.go` editing.
		// I'll implement validation expecting `initData` in query or header `X-Telegram-Init-Data`.

		// For now, let's just log and skip validation to unblock, OR fail if critical.
		// "Admin Only" -> MUST VALIDATE.
		// I'll read `X-Telegram-Init-Data` header.
		// I will update the template later to send this header.

		initData := r.Header.Get("X-Telegram-Init-Data")
		if initData == "" {
			// try query param
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
		if query == "" {
			json.NewEncoder(w).Encode([]interface{}{})
			return
		}

		patients, err := repo.SearchPatients(query)
		if err != nil {
			logging.Errorf("Search failed: %v", err)
			http.Error(w, "Search failed", http.StatusInternalServerError)
			return
		}

		// Map to JSON safe struct
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
	})

	mux.HandleFunc("/cancel", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			_ = json.NewEncoder(w).Encode(map[string]string{"status": "error", "error": "Method not allowed"})
			return
		}

		// Parse JSON body
		var reqBody struct {
			InitData string `json:"initData"`
			ApptID   string `json:"apptId"`
		}
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
			logging.Debugf(" [WebApp]: Failed to parse /cancel body: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{"status": "error", "error": "Invalid request"})
			return
		}

		if reqBody.InitData == "" || reqBody.ApptID == "" {
			logging.Debugf(" [WebApp]: Missing parameters in /cancel")
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{"status": "error", "error": "Missing parameters"})
			return
		}

		// Validate using Telegram's native initData (bulletproof, never expires)
		userID, userName, err := validateInitData(reqBody.InitData, botToken)
		if err != nil {
			logging.Debugf(" [WebApp]: initData validation failed: %v", err)
			w.WriteHeader(http.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(map[string]string{"status": "error", "error": "Сессия недействительна. Закройте и откройте мед-карту заново."})
			return
		}

		logging.Debugf(" [WebApp]: Validated cancel request from user %s (%s)", userID, userName)

		// Security: Ensure the appointment belongs to the user
		appt, err := apptService.FindByID(r.Context(), reqBody.ApptID)
		if err != nil {
			logging.Errorf("Cancel Error: Appt %s not found: %v", reqBody.ApptID, err)
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]string{"status": "error", "error": "Запись не найдена"})
			return
		}

		if appt.CustomerTgID != userID {
			logging.Errorf("Cancel Error: Appt %s (Owner: %s) access denied for User %s", reqBody.ApptID, appt.CustomerTgID, userID)
			w.WriteHeader(http.StatusForbidden)
			_ = json.NewEncoder(w).Encode(map[string]string{"status": "error", "error": "Доступ запрещен"})
			return
		}

		// Enforce 72h rule
		now := time.Now().In(domain.ApptTimeZone)
		if appt.StartTime.Sub(now) < 72*time.Hour {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"status": "error",
				"error":  "До приема менее 72 часов. Пожалуйста, напишите терапевту напрямую.",
			})
			return
		}

		err = apptService.CancelAppointment(r.Context(), reqBody.ApptID)
		if err != nil {
			logging.Errorf("Cancel Error: Failed to cancel appt %s: %v", reqBody.ApptID, err)
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(map[string]string{"status": "error", "error": "Не удалось отменить запись"})
			return
		}

		logging.Infof("TWA: User %s cancelled appointment %s", userID, reqBody.ApptID)

		// NOTIFY VIA BOT
		notificationMsg := fmt.Sprintf("⚠️ *Запись отменена!*\n\nПациент: %s\nДата: %s\nУслуга: %s",
			appt.CustomerName, appt.StartTime.In(domain.ApptTimeZone).Format("02.01.2006 15:04"), appt.Service.Name)

		// 1. Notify Admins
		for _, adminID := range adminIDs {
			sendTelegramMessage(botToken, adminID, notificationMsg)
		}

		// 2. Notify Patient (Push confirmation)
		patientMsg := fmt.Sprintf("✅ *Вы успешно отменили запись:*\n\n📅 %s\n💆 %s\n\nЖдем вас в другой раз!",
			appt.StartTime.In(domain.ApptTimeZone).Format("02.01.2006 15:04"), appt.Service.Name)
		sendTelegramMessage(botToken, userID, patientMsg)

		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	// WebDAV Handler for Obsidian Sync
	davUser := os.Getenv("WEBDAV_USER")
	davPass := os.Getenv("WEBDAV_PASSWORD")

	if davUser != "" && davPass != "" {
		davHandler := &webdav.Handler{
			Prefix:     "/webdav/",
			FileSystem: webdav.Dir(dataDir),
			LockSystem: webdav.NewMemLS(),
			Logger: func(r *http.Request, err error) {
				if err != nil {
					logging.Errorf("WebDAV [Err] %s %s: %v", r.Method, r.URL.Path, err)
				}
			},
		}

		webdavAuthHandler := func(w http.ResponseWriter, r *http.Request) {
			// Add CORS headers for Obsidian/Browser compatibility
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PROPFIND, PROPPATCH, MKCOL, COPY, MOVE, LOCK, UNLOCK")
			w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type, Depth, Destination, If-Modified-Since, Overwrite, User-Agent, X-Expected-Entity-Length")
			w.Header().Set("Access-Control-Expose-Headers", "DAV, content-length, Allow")

			if r.Method == "OPTIONS" {
				// Let WebDAV handler provide the DAV: 1, 2 headers required for client discovery
				davHandler.ServeHTTP(w, r)
				return
			}

			user, pass, ok := r.BasicAuth()
			if !ok || user != davUser || pass != davPass {
				if !ok {
					logging.Warnf("WebDAV [Auth] Missing] %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)
				} else {
					logging.Warnf("WebDAV [Auth] Denied] user=%s from %s", user, r.RemoteAddr)
				}
				w.Header().Set("WWW-Authenticate", `Basic realm="Vera Bot Medical Records"`)
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// Browser-friendly landing page if accessing /webdav/ directly via GET
			if r.Method == "GET" && r.URL.Path == "/webdav/" && !(strings.Contains(r.Header.Get("User-Agent"), "Obsidian") || strings.Contains(r.Header.Get("User-Agent"), "DAV")) {
				w.Header().Set("Content-Type", "text/html; charset=utf-8")

				// Diagnostic check of the storage directory
				info, err := os.Stat(dataDir)
				storageStatus := "✅ Доступно"
				if err != nil {
					storageStatus = fmt.Sprintf("❌ Ошибка доступа: %v", err)
				} else if !info.IsDir() {
					storageStatus = "❌ Ошибка: Путь не является папкой"
				}

				fmt.Fprintf(w, `
					<html>
					<head><style>body{font-family:sans-serif;padding:20px;line-height:1.6}code{background:#eee;padding:2px 5px}</style></head>
					<body>
						<h1>✅ WebDAV Сервер Активен</h1>
						<p>Пользователь: <b>%s</b></p>
						<p>Статус хранилища: %s</p>
						<hr>
						<p><b>Для настройки в Obsidian:</b></p>
						<ul>
							<li>Remote Service: <code>WebDAV</code></li>
							<li>Server Address: <code>https://%s/webdav/</code></li>
						</ul>
					</body>
					</html>
				`, davUser, storageStatus, r.Host)
				return
			}

			logging.Infof("WebDAV [%s] %s (User: %s)", r.Method, r.URL.Path, user)
			davHandler.ServeHTTP(w, r)
		}

		// Use a single handler for /webdav/ and redirect /webdav (no slash)
		mux.HandleFunc("/webdav", func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, "/webdav/", http.StatusMovedPermanently)
		})
		mux.HandleFunc("/webdav/", webdavAuthHandler)
		logging.Infof("WebDAV server enabled at /webdav/ (User: %s)", davUser)
	} else {
		log.Println("Warning: WEBDAV_USER or WEBDAV_PASSWORD not set. WebDAV disabled.")
	}

	logging.Infof("Starting Web App server on :%s", port)
	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Web App server failed: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("Shutting down Web App server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logging.Infof("Web App server shutdown error: %v", err)
	}
}

// Helper to send simple Telegram messages without complex dependencies
func sendTelegramMessage(token, chatID, text string) {
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", token)
	payload, _ := json.Marshal(map[string]string{
		"chat_id":    chatID,
		"text":       text,
		"parse_mode": "Markdown",
	})

	resp, err := http.Post(apiURL, "application/json", strings.NewReader(string(payload)))
	if err != nil {
		logging.Infof("Failed to send bot notification: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logging.Infof("Telegram API error: %s", resp.Status)
	}
}
