package web

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/kfilin/massage-bot/internal/ports"
)

// NewDraftHandler handles Draft Approval/Discard for voice transcripts.
// Admin-only: approves (appends transcript to therapist notes) or discards
// the pending media, then updates its status.
func NewDraftHandler(repo ports.Repository, botToken string, adminIDs []string, secret string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req struct {
			ID       string `json:"id"`
			InitData string `json:"initData"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		userID, _, err := validateInitData(req.InitData, botToken)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		isAdmin := false
		for _, adminID := range adminIDs {
			if adminID == userID {
				isAdmin = true
				break
			}
		}
		if !isAdmin {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		media, err := repo.GetMediaByID(req.ID)
		if err != nil {
			http.Error(w, "Draft not found", http.StatusNotFound)
			return
		}

		if strings.HasSuffix(r.URL.Path, "/approve") {
			// Approve: Add to therapist notes
			patient, _ := repo.GetPatient(media.PatientID)
			newNotes := patient.TherapistNotes + "\n\n--- 🎤 Расшифровка ---\n" + media.Transcript
			_ = repo.UpdatePatientProfile(media.PatientID, patient.Name, newNotes)
			_ = repo.UpdateMediaStatus(req.ID, "approved", media.Transcript)
		} else {
			// Discard
			_ = repo.UpdateMediaStatus(req.ID, "discarded", media.Transcript)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}
}
