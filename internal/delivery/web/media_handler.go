package web

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/kfilin/massage-bot/internal/logging"
	"github.com/kfilin/massage-bot/internal/ports"
)

// MediaHandler serves patient media files (/api/media/<id>) after verifying
// the auth cookie (telegramID:timestamp:HMAC) and checking access control.
type MediaHandler struct {
	repo     ports.Repository
	secret   string
	adminIDs []string
}

// NewMediaHandler constructs a MediaHandler.
func NewMediaHandler(repo ports.Repository, secret string, adminIDs []string) *MediaHandler {
	return &MediaHandler{
		repo:     repo,
		secret:   secret,
		adminIDs: adminIDs,
	}
}

// GetMedia serves a single media file by ID. Assumes StripPrefix is used,
// so valid path is just the ID.
func (h *MediaHandler) GetMedia(w http.ResponseWriter, r *http.Request) {
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
	if len(parts) != 3 {
		http.Error(w, "Invalid auth format", http.StatusUnauthorized)
		return
	}
	telegramID := parts[0]
	timestamp := parts[1]
	signature := parts[2]

	if !h.validateSignature(telegramID, timestamp, signature) {
		http.Error(w, "Invalid or expired signature", http.StatusForbidden)
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
	dataDir := os.Getenv("DATA_DIR")
	if dataDir == "" {
		dataDir = "data"
	}

	finalPath := media.FilePath
	if !filepath.IsAbs(finalPath) {
		finalPath = filepath.Join(dataDir, finalPath)
	}

	// SECURITY: Verify resolved path stays within dataDir (prevent path traversal)
	absPath, err := filepath.Abs(finalPath)
	if err != nil {
		logging.Warnf("Media path resolution failed for %s: %v", mediaID, err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	absDataDir, err := filepath.Abs(dataDir)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if !strings.HasPrefix(absPath, absDataDir+string(filepath.Separator)) {
		logging.Warnf("Path traversal attempt blocked: mediaID=%s resolved to %s (dataDir=%s)", mediaID, absPath, absDataDir)
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	http.ServeFile(w, r, absPath)
}

// validateSignature verifies the HMAC signature and checks the timestamp is within the 24h TTL.
// Cookie format: telegramID:unixTimestamp:signature
func (h *MediaHandler) validateSignature(telegramID, timestamp, signature string) bool {
	ts, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		logging.Warnf("[validateSignature]: Invalid timestamp for ID=%s", telegramID)
		return false
	}
	const tokenTTL = 24 * 60 * 60 // 24 hours in seconds
	if time.Now().Unix()-ts > tokenTTL {
		logging.Warnf("[validateSignature]: Expired token for ID=%s (age=%ds)", telegramID, time.Now().Unix()-ts)
		return false
	}
	mac := hmac.New(sha256.New, []byte(h.secret))
	mac.Write([]byte(telegramID + ":" + timestamp))
	expectedSignature := hex.EncodeToString(mac.Sum(nil))
	match := hmac.Equal([]byte(signature), []byte(expectedSignature))
	if !match {
		logging.Debugf("[validateSignature]: Signature mismatch for ID=%s", telegramID)
	}
	return match
}

