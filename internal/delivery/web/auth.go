// Package web implements the HTTP server for the Telegram WebApp
// (patient card, search, draft approval, etc.) and the supporting
// auth helpers (HMAC, Telegram initData validation, auth-cookie signing).
package web

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/kfilin/massage-bot/internal/logging"
)

// initDataMaxAge is the maximum tolerated age of a Telegram WebApp initData payload.
// After this window, the payload is considered stale and rejected, preventing replay
// of leaked URLs beyond initDataMaxAge.
const initDataMaxAge = 1 * time.Hour

// initDataClockSkew tolerates client/server clock drift of up to 5 minutes.
const initDataClockSkew = 5 * time.Minute

// hmacMaxAge is the maximum tolerated age of an HMAC-signed webapp link.
// Links older than this are rejected to limit replay window for shared URLs.
const hmacMaxAge = 7 * 24 * time.Hour

// generateHMAC produces a hex-encoded HMAC-SHA256 of (id+":"+ts) using secret.
// Including the timestamp in the signed payload prevents an attacker from
// rolling the ts value forward to extend a link's lifetime.
func generateHMAC(id string, ts string, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(strings.TrimSpace(id) + ":" + ts))
	return hex.EncodeToString(h.Sum(nil))
}

// validateHMAC compares a provided token to the expected HMAC for (id, ts, secret)
// and rejects payloads with a ts older than hmacMaxAge.
func validateHMAC(id string, ts string, token string, secret string) bool {
	if ts == "" {
		logging.Warnf(" [validateHMAC]: Legacy link without ts for ID=%s. Accepting without expiry check.", id)
		// Backward-compat: accept old (id-only) tokens but warn so operator
		// can see how many legacy links are still in circulation.
		expected := generateHMACNoTS(id, secret)
		match := hmac.Equal([]byte(token), []byte(expected))
		if !match {
			logging.Debugf(" [validateHMAC]: Legacy mismatch for ID=%s. Provided=%s, Expected=%s, SecretLen=%d", id, token, expected, len(secret))
		}
		return match
	}
	tsUnix, err := strconv.ParseInt(ts, 10, 64)
	if err != nil {
		logging.Warnf(" [validateHMAC]: Invalid ts=%q for ID=%s: %v", ts, id, err)
		return false
	}
	age := time.Since(time.Unix(tsUnix, 0))
	if age > hmacMaxAge {
		logging.Warnf(" [validateHMAC]: Link expired for ID=%s (age=%s, max=%s)", id, age.Round(time.Second), hmacMaxAge)
		return false
	}
	if age < -initDataClockSkew {
		logging.Warnf(" [validateHMAC]: Link ts in the future for ID=%s (ahead=%s)", id, -age.Round(time.Second))
		return false
	}
	expected := generateHMAC(id, ts, secret)
	match := hmac.Equal([]byte(token), []byte(expected))
	if !match {
		logging.Debugf(" [validateHMAC]: Mismatch for ID=%s. Provided=%s, Expected=%s, SecretLen=%d", id, token, expected, len(secret))
	}
	return match
}

// generateHMACNoTS is the legacy HMAC for backward compat with links generated
// before the ts hardening. New code should use generateHMAC.
func generateHMACNoTS(id string, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(strings.TrimSpace(id)))
	return hex.EncodeToString(h.Sum(nil))
}

// validateInitData validates Telegram WebApp initData, including signature,
// presence of auth_date, and expiry (rejecting payloads older than initDataMaxAge
// or further in the future than initDataClockSkew).
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

	// Reject stale initData: auth_date must be present and within the last hour.
	// Without this, a leaked initData URL grants permanent access to the patient card.
	authDateStr := values.Get("auth_date")
	if authDateStr == "" {
		return "", "", fmt.Errorf("missing auth_date")
	}
	authDateUnix, err := strconv.ParseInt(authDateStr, 10, 64)
	if err != nil {
		return "", "", fmt.Errorf("invalid auth_date: %w", err)
	}
	authDate := time.Unix(authDateUnix, 0)
	if age := time.Since(authDate); age > initDataMaxAge {
		return "", "", fmt.Errorf("initData expired (%s old, max %s)", age.Round(time.Second), initDataMaxAge)
	}
	if age := time.Since(authDate); age < -initDataClockSkew {
		return "", "", fmt.Errorf("initData auth_date is in the future (%s ahead)", -age.Round(time.Second))
	}

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

// GenerateAuthCookie creates a time-limited cookie value for media-access auth.
// Format: telegramID:unixTimestamp:HMAC_SHA256(telegramID:unixTimestamp, secret)
// Tokens expire after 24 hours, preventing replay attacks from leaked URLs/logs.
func GenerateAuthCookie(telegramID, secret string) string {
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(telegramID + ":" + timestamp))
	signature := hex.EncodeToString(mac.Sum(nil))
	return fmt.Sprintf("%s:%s:%s", telegramID, timestamp, signature)
}
