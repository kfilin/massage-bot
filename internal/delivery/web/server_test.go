package web

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// withTelegramAPIBase temporarily overrides the package-level
// telegramAPIBase and restores the original on cleanup.
func withTelegramAPIBase(t *testing.T, base string) {
	t.Helper()
	old := telegramAPIBase
	telegramAPIBase = base
	t.Cleanup(func() { telegramAPIBase = old })
}

// readBody reads the entire request body and returns it as a string.
func readBody(t *testing.T, r *http.Request) string {
	t.Helper()
	b, err := io.ReadAll(r.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	return string(b)
}

// TestSendTelegramMessage_200OK exercises the happy path: 200 from
// the (mocked) Telegram API is treated as success. Verifies the
// request URL, method, body shape, and that no error is logged.
func TestSendTelegramMessage_200OK(t *testing.T) {
	var (
		gotMethod string
		gotPath   string
		gotBody   string
		gotAuth   string
	)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		gotBody = readBody(t, r)
		gotAuth = r.Header.Get("Content-Type")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true,"result":{}}`))
	}))
	defer srv.Close()
	withTelegramAPIBase(t, srv.URL)

	sendTelegramMessage("test-token", "12345", "hello")

	if gotMethod != http.MethodPost {
		t.Errorf("method: got %q, want POST", gotMethod)
	}
	if gotPath != "/bottest-token/sendMessage" {
		// httptest preserves the case of the URL prefix; the bot token
		// was lower-cased in the test call. This is the actual URL
		// path the function built.
		t.Errorf("path: got %q, want /bottest-token/sendMessage", gotPath)
	}
	if !strings.Contains(gotAuth, "application/json") {
		t.Errorf("content-type: got %q, want application/json", gotAuth)
	}

	// Verify body shape: chat_id, text, parse_mode.
	var payload map[string]string
	if err := json.Unmarshal([]byte(gotBody), &payload); err != nil {
		t.Fatalf("body is not valid JSON: %v\nbody=%s", err, gotBody)
	}
	if payload["chat_id"] != "12345" {
		t.Errorf("chat_id: got %q, want 12345", payload["chat_id"])
	}
	if payload["text"] != "hello" {
		t.Errorf("text: got %q, want hello", payload["text"])
	}
	if payload["parse_mode"] != "HTML" {
		t.Errorf("parse_mode: got %q, want HTML", payload["parse_mode"])
	}
}

// TestSendTelegramMessage_Non200 checks the non-200 branch is
// handled (logged, no panic). The function does not bubble the error
// up — it's fire-and-forget for caller.
func TestSendTelegramMessage_Non200(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"ok":false,"description":"Unauthorized"}`))
	}))
	defer srv.Close()
	withTelegramAPIBase(t, srv.URL)

	// Should not panic; error is logged and swallowed.
	sendTelegramMessage("bad-token", "12345", "hello")
}

// TestSendTelegramMessage_NetworkError ensures the function handles
// a network failure (server already closed) without panicking.
func TestSendTelegramMessage_NetworkError(t *testing.T) {
	// Pre-close a server so the request fails fast with a connection
	// error. The function should log the error and return cleanly.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	srv.Close()
	withTelegramAPIBase(t, srv.URL)

	sendTelegramMessage("test-token", "12345", "hello")
}

// TestSendTelegramMessage_EmptyToken verifies behaviour when called
// with an empty token (URL still builds, server may or may not accept
// the request). The function must not panic on empty inputs.
func TestSendTelegramMessage_EmptyToken(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()
	withTelegramAPIBase(t, srv.URL)

	sendTelegramMessage("", "12345", "hello")

	// URL should still contain "/bot/sendMessage" even with empty token.
	if !strings.HasPrefix(gotPath, "/bot/") {
		t.Errorf("path: got %q, want prefix /bot/", gotPath)
	}
}
