package main

import (
	"encoding/json"
	"net/http"

	"github.com/kfilin/massage-bot/internal/logging"
	"github.com/kfilin/massage-bot/internal/ports"
)

// NewTranscribeHandler creates the handler for Voice Transcription API
func NewTranscribeHandler(transcriptionService ports.TranscriptionService, botToken string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			_ = json.NewEncoder(w).Encode(map[string]string{"status": "error", "error": "Method not allowed"})
			return
		}

		// Parse multipart form (max 10MB)
		if err := r.ParseMultipartForm(10 << 20); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{"status": "error", "error": "Failed to parse form"})
			return
		}

		// Auth Check
		initData := r.Header.Get("X-Telegram-Init-Data")
		if initData == "" {
			initData = r.FormValue("initData")
		}

		if initData == "" {
			w.WriteHeader(http.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(map[string]string{"status": "error", "error": "Unauthorized: missing initData"})
			return
		}

		_, _, err := validateInitData(initData, botToken)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(map[string]string{"status": "error", "error": "Unauthorized: invalid initData"})
			return
		}

		// Get File
		// We expect the client to send the file in a field named 'voice' or 'file'
		file, header, err := r.FormFile("voice")
		if err != nil {
			// Try "file" key as fallback
			file, header, err = r.FormFile("file")
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				_ = json.NewEncoder(w).Encode(map[string]string{"status": "error", "error": "Missing 'voice' or 'file' in form"})
				return
			}
		}
		defer file.Close()

		// Transcribe
		text, err := transcriptionService.Transcribe(r.Context(), file, header.Filename)
		if err != nil {
			logging.Errorf("Transcription failed: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(map[string]string{"status": "error", "error": "Transcription failed"})
			return
		}

		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok", "text": text})
	}
}
