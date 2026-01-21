package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/kfilin/massage-bot/internal/ports"
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

func startWebAppServer(port string, secret string, repo ports.Repository, apptService ports.AppointmentService) {
	if port == "" {
		port = "8082"
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/card", func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")
		token := r.URL.Query().Get("token")

		if id == "" || token == "" {
			http.Error(w, "Missing id or token", http.StatusBadRequest)
			return
		}

		if !validateHMAC(id, token, secret) {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		patient, err := repo.GetPatient(id)
		if err != nil {
			log.Printf("Error fetching patient %s: %v", id, err)
			http.Error(w, "Patient not found", http.StatusNotFound)
			return
		}

		// Sync logic: Fetch actual appointments from GCal to ensure Medical Card is up-to-date
		appts, err := apptService.GetCustomerAppointments(r.Context(), id)
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

	log.Printf("Starting Web App server on :%s", port)
	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Web App server failed: %v", err)
	}
}
