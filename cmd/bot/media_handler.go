package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/kfilin/massage-bot/internal/logging"
	"github.com/kfilin/massage-bot/internal/ports"
)

type MediaHandler struct {
	repo     ports.Repository
	secret   string
	adminIDs []string
}

func NewMediaHandler(repo ports.Repository, secret string, adminIDs []string) *MediaHandler {
	return &MediaHandler{
		repo:     repo,
		secret:   secret,
		adminIDs: adminIDs,
	}
}

func (h *MediaHandler) GetMedia(w http.ResponseWriter, r *http.Request) {
	// Assumes StripPrefix is used, so valid path is just the ID
	mediaID := strings.TrimPrefix(r.URL.Path, "/")
	if mediaID == "" {
		http.Error(w, "Missing media ID", http.StatusBadRequest)
		return
	}

	// 1. Auth Check via Cookie
	cookie, err := r.Cookie("vera_auth")
	if err != nil || cookie.Value == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	parts := strings.Split(cookie.Value, ":")
	if len(parts) != 2 {
		http.Error(w, "Invalid auth format", http.StatusUnauthorized)
		return
	}
	telegramID := parts[0]
	signature := parts[1]

	if !h.validateSignature(telegramID, signature) {
		http.Error(w, "Invalid signature", http.StatusForbidden)
		return
	}

	// 2. Fetch Media Metadata
	media, err := h.repo.GetMediaByID(mediaID)
	if err != nil {
		logging.Warnf("Media not found: %s (err: %v)", mediaID, err)
		http.Error(w, "Media not found", http.StatusNotFound)
		return
	}

	// 3. Access Control
	isAdmin := false
	for _, id := range h.adminIDs {
		if id == telegramID {
			isAdmin = true
			break
		}
	}

	if media.PatientID != telegramID && !isAdmin {
		logging.Warnf("Access denied for user %s to media %s (owner: %s)", telegramID, mediaID, media.PatientID)
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	// 4. Serve File
	// Resolve path against DATA_DIR
	dataDir := os.Getenv("DATA_DIR")
	if dataDir == "" {
		dataDir = "data"
	}

	// If the stored path is absolute, use it. If relative, join with dataDir.
	finalPath := media.FilePath
	if !filepath.IsAbs(finalPath) {
		finalPath = filepath.Join(dataDir, finalPath)
	}

	http.ServeFile(w, r, finalPath)
}

func (h *MediaHandler) validateSignature(telegramID, signature string) bool {
	mac := hmac.New(sha256.New, []byte(h.secret))
	mac.Write([]byte(telegramID))
	expectedSignature := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(signature), []byte(expectedSignature))
}

// GenerateAuthCookie creates the cookie value
func GenerateAuthCookie(telegramID, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(telegramID))
	signature := hex.EncodeToString(mac.Sum(nil))
	return fmt.Sprintf("%s:%s", telegramID, signature)
}
