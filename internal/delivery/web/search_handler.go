package web

import (
	"encoding/json"
	"net/http"
	"sort"
	"strings"

	"github.com/kfilin/massage-bot/internal/domain"
	"github.com/kfilin/massage-bot/internal/logging"
	"github.com/kfilin/massage-bot/internal/ports"
)

// NewSearchHandler creates the handler for the Patient Search API.
// Admin-only: returns all patients when q is empty, otherwise filters by name.
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
		var patients []domain.Patient
		if query == "" {
			patients, err = repo.GetAllPatients()
		} else {
			patients, err = repo.SearchPatients(query)
		}

		if err != nil {
			logging.Errorf("Search failed: %v", err)
			http.Error(w, "Search failed", http.StatusInternalServerError)
			return
		}

		// Sort alphabetically by name
		sort.Slice(patients, func(i, j int) bool {
			return strings.ToLower(patients[i].Name) < strings.ToLower(patients[j].Name)
		})

		type patResult struct {
			TelegramID  string `json:"telegram_id"`
			Name        string `json:"name"`
			TotalVisits int    `json:"total_visits"`
		}
		results := make([]patResult, 0)
		for _, p := range patients {
			results = append(results, patResult{
				TelegramID:  p.TelegramID,
				Name:        p.Name,
				TotalVisits: p.TotalVisits,
			})
		}
		if err := json.NewEncoder(w).Encode(results); err != nil {
		logging.Errorf("Failed to encode search results: %v", err)
	}
	}
}
