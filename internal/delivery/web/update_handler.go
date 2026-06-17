package web

import (
	"encoding/json"
	"net/http"

	"github.com/kfilin/massage-bot/internal/logging"
	"github.com/kfilin/massage-bot/internal/ports"
)

// NewUpdatePatientHandler creates the handler for updating a patient profile.
// Admin-only: updates name and notes, with a 50KB cap on notes to prevent
// payload stuffing.
func NewUpdatePatientHandler(repo ports.Repository, botToken string, adminIDs []string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			_ = json.NewEncoder(w).Encode(map[string]string{"status": "error", "error": "Method not allowed"})
			return
		}

		var reqBody struct {
			InitData string `json:"initData"`
			ID       string `json:"id"`
			Name     string `json:"name"`
			Notes    string `json:"notes"`
		}
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
			logging.Errorf("Failed to decode update request: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{"status": "error", "error": "Invalid JSON"})
			return
		}

		if reqBody.InitData == "" || reqBody.ID == "" {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{"status": "error", "error": "Missing initData or ID"})
			return
		}

		// SECURITY: Cap notes length to prevent payload stuffing
		const maxNotesLength = 50_000 // ~50KB
		if len(reqBody.Notes) > maxNotesLength {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{"status": "error", "error": "Notes too long (max 50KB)"})
			return
		}

		// Authenticate Admin
		userID, _, err := validateInitData(reqBody.InitData, botToken)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(map[string]string{"status": "error", "error": "Unauthorized"})
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
			w.WriteHeader(http.StatusForbidden)
			_ = json.NewEncoder(w).Encode(map[string]string{"status": "error", "error": "Access denied"})
			return
		}

		// Perform Update
		if err := repo.UpdatePatientProfile(reqBody.ID, reqBody.Name, reqBody.Notes); err != nil {
			logging.Errorf("Failed to update patient %s: %v", reqBody.ID, err)
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(map[string]string{"status": "error", "error": "Database update failed"})
			return
		}

		logging.Infof("Admin %s updated patient %s", userID, reqBody.ID)
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}
}
