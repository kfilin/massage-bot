package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"

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
	return hmac.Equal([]byte(token), []byte(expected))
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

func startWebAppServer(port string, secret string, botToken string, repo ports.Repository, apptService ports.AppointmentService, dataDir string) {
	if port == "" {
		port = "8082"
	}

	if dataDir == "" {
		dataDir = "data"
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/card", func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")
		token := r.URL.Query().Get("token")
		initData := r.URL.Query().Get("initData") // Fallback for Menu Button

		var finalID string
		var telegramName string

		if id != "" && token != "" {
			if !validateHMAC(id, token, secret) {
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				return
			}
			finalID = id
		} else if initData != "" {
			var err error
			finalID, telegramName, err = validateInitData(initData, botToken)
			if err != nil {
				log.Printf("InitData validation failed: %v", err)
				http.Error(w, "Authentication failed", http.StatusUnauthorized)
				return
			}
		} else {
			// If both missing, we might still have it in the hash (fragment)
			// But Go can't see the fragment. We need a JS gateway or just show a nice error.
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, `
				<!DOCTYPE html>
				<html>
				<head>
					<meta charset="UTF-8">
					<meta name="viewport" content="width=device-width, initial-scale=1.0">
					<script src="https://telegram.org/js/telegram-web-app.js"></script>
					<title>Авторизация...</title>
					<style>
						body { font-family: sans-serif; display: flex; align-items: center; justify-content: center; height: 100vh; margin: 0; background: #f0f2f5; }
						.loader { border: 4px solid #f3f3f3; border-top: 4px solid #3498db; border-radius: 50%; width: 30px; height: 30px; animation: spin 2s linear infinite; }
						@keyframes spin { 0% { transform: rotate(0deg); } 100% { transform: rotate(360deg); } }
					</style>
				</head>
				<body>
					<div id="status">⏳ Авторизация...</div>
					<script>
						const tg = window.Telegram.WebApp;
						if (tg.initData) {
							const currentUrl = new URL(window.location.href);
							currentUrl.searchParams.set('initData', tg.initData);
							window.location.href = currentUrl.toString();
						} else {
							document.getElementById('status').innerHTML = "❌ Ошибка: Некорректная ссылка.<br><br>Пожалуйста, откройте карту через кнопку в чате @vera_massage_bot";
						}
					</script>
				</body>
				</html>
			`)
			return
		}

		patient, err := repo.GetPatient(finalID)
		if err != nil {
			log.Printf("Patient %s not found in DB, attempting self-heal from GCal...", finalID)

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

		// Sync logic: Fetch actual appointments from GCal to ensure Medical Card is up-to-date
		appts, err := apptService.GetCustomerAppointments(r.Context(), finalID)
		if err == nil {
			// Update visit stats even if zero
			var lastVisit, firstVisit time.Time
			if len(appts) > 0 {
				for _, a := range appts {
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
			patient.TotalVisits = len(appts)

			// CLEANUP LEGACY AUDIT LOGS FROM NOTES (Aggressive regex scrubbing)
			// Matches lines starting with (optional symbols) followed by Запись:, Первая запись:, or Зарегистрирован:
			scrubRegex := regexp.MustCompile(`(?m)^.*(Запись:|Первая запись:|Зарегистрирован:).*$\n?`)
			patient.TherapistNotes = scrubRegex.ReplaceAllString(patient.TherapistNotes, "")
			patient.TherapistNotes = strings.TrimSpace(patient.TherapistNotes)

			// Save back to repo to persist the sync
			repo.SavePatient(patient)
		}

		html := repo.GenerateHTMLRecord(patient)
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, html)
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
					log.Printf("WebDAV [Err] %s %s: %v", r.Method, r.URL.Path, err)
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
					log.Printf("WebDAV [Auth Missing] %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)
				} else {
					log.Printf("WebDAV [Auth Denied] user=%s from %s", user, r.RemoteAddr)
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

			log.Printf("WebDAV [%s] %s (User: %s)", r.Method, r.URL.Path, user)
			davHandler.ServeHTTP(w, r)
		}

		// Use a single handler for /webdav/ and redirect /webdav (no slash)
		mux.HandleFunc("/webdav", func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, "/webdav/", http.StatusMovedPermanently)
		})
		mux.HandleFunc("/webdav/", webdavAuthHandler)
		log.Printf("WebDAV server enabled at /webdav/ (User: %s)", davUser)
	} else {
		log.Println("Warning: WEBDAV_USER or WEBDAV_PASSWORD not set. WebDAV disabled.")
	}

	log.Printf("Starting Web App server on :%s", port)
	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Web App server failed: %v", err)
	}
}
