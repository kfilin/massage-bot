package web

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/kfilin/massage-bot/internal/logging"
	"github.com/kfilin/massage-bot/internal/ports"
	"github.com/kfilin/massage-bot/internal/presentation"
	"golang.org/x/net/webdav"
)

// StartServer launches the HTTP server for the WebApp on the given port.
// It registers all webapp routes (patient card, search, draft, cancel,
// update, transcribe, media, WebDAV) and blocks until ctx is cancelled.
func StartServer(
	ctx context.Context,
	port string,
	secret string,
	botToken string,
	adminIDs []string,
	repo ports.Repository,
	apptService ports.AppointmentService,
	transcriptionService ports.TranscriptionService,
	dataDir string,
	botUsername string,
) {
	if port == "" {
		port = "8082"
	}

	if dataDir == "" {
		dataDir = "data"
	}

	mux := http.NewServeMux()

	// Initialize Presenters
	webPresenter, err := presentation.NewWebPresenter()
	if err != nil {
		log.Fatalf("Failed to initialize web presenter: %v", err)
	}
	botPresenter := presentation.NewBotPresenter()

	// Static Assets (using internal/presentation/templates)
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(presentation.StaticFS))))

	// Handle both root and /card with the same logic
	handler := NewWebAppHandler(repo, apptService, webPresenter, botToken, adminIDs, secret)

	mux.HandleFunc("/", handler)
	mux.HandleFunc("/card", handler)

	// API Handlers
	mux.HandleFunc("/api/search", NewSearchHandler(repo, botToken, adminIDs))
	mux.HandleFunc("/api/patient/update", NewUpdatePatientHandler(repo, botToken, adminIDs))
	mux.HandleFunc("/cancel", NewCancelHandler(apptService, botToken, adminIDs, botPresenter))
	mux.HandleFunc("/api/transcribe", NewTranscribeHandler(transcriptionService, botToken))

	// Draft Handlers
	draftHandler := NewDraftHandler(repo, botToken, adminIDs, secret)
	mux.HandleFunc("/api/draft/approve", draftHandler)
	mux.HandleFunc("/api/draft/discard", draftHandler)

	mediaHandler := NewMediaHandler(repo, secret, adminIDs)
	mux.Handle("/api/media/", http.StripPrefix("/api/media/", http.HandlerFunc(mediaHandler.GetMedia)))

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
			allowedOrigin := os.Getenv("WEBAPP_URL")
			if allowedOrigin == "" {
				allowedOrigin = "*"
			}
			w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PROPFIND, PROPPATCH, MKCOL, COPY, MOVE, LOCK, UNLOCK")
			w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type, Depth, Destination, If-Modified-Since, Overwrite, User-Agent, X-Expected-Entity-Length")
			w.Header().Set("Access-Control-Expose-Headers", "DAV, content-length, Allow")

			if r.Method == "OPTIONS" {
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

			if r.Method == "GET" && r.URL.Path == "/webdav/" && !(strings.Contains(r.Header.Get("User-Agent"), "Obsidian") || strings.Contains(r.Header.Get("User-Agent"), "DAV")) {
				w.Header().Set("Content-Type", "text/html; charset=utf-8")

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

// telegramAPIBase is the base URL for the Telegram Bot HTTP API.
// Exposed as a package var so tests can override it to point at an
// httptest server and assert on actual request shape (URL, body,
// status code) without depending on the public Telegram API.
var telegramAPIBase = "https://api.telegram.org"

// sendTelegramMessage posts a text message to a single chat via the Bot HTTP API.
func sendTelegramMessage(token, chatID, text string) {
	apiURL := fmt.Sprintf("%s/bot%s/sendMessage", telegramAPIBase, token)
	payload, _ := json.Marshal(map[string]string{
		"chat_id":    chatID,
		"text":       text,
		"parse_mode": "HTML",
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
