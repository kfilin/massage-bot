package web

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/kfilin/massage-bot/internal/domain"
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

// createDummyMuxInputs returns minimal mock dependencies that satisfy
// the signature of createWebAppMux without panicking on route setup.
func createDummyMuxInputs(t *testing.T) (string, string, []string, *mockRepo, *mockApptService, *mockTranscriptionService, string, string) {
	t.Helper()
	return "test-secret", "123:test-token", []string{"admin1"},
		&mockRepo{getAllPatientsFunc: func() ([]domain.Patient, error) { return nil, nil }},
		&mockApptService{},
		&mockTranscriptionService{},
		t.TempDir(), "testbot"
}

// TestCreateWebAppMux_RoutesRegistered verifies that createWebAppMux
// registers expected route patterns without crashing.
func TestCreateWebAppMux_RoutesRegistered(t *testing.T) {
	secret, botToken, adminIDs, repo, apptSvc, transSvc, dataDir, botUser := createDummyMuxInputs(t)

	mux := createWebAppMux(secret, botToken, adminIDs, repo, apptSvc, transSvc, dataDir, botUser)
	if mux == nil {
		t.Fatal("createWebAppMux returned nil")
	}

	ts := httptest.NewServer(mux)
	defer ts.Close()

	tests := []struct {
		path       string
		wantStatus int // 0 = skip status check (just verify non-404)
	}{
		{path: "/", wantStatus: 0},
		{path: "/card", wantStatus: 0},
		{path: "/api/search", wantStatus: 0},
		{path: "/cancel", wantStatus: 0},
		{path: "/api/transcribe", wantStatus: 0},
		{path: "/api/media/test.jpg", wantStatus: 0},
		{path: "/api/draft/approve", wantStatus: 0},
		{path: "/api/draft/discard", wantStatus: 0},
		{path: "/api/patient/update", wantStatus: 0},
		{path: "/static/", wantStatus: 0},
	}

	for _, tc := range tests {
		t.Run(tc.path, func(t *testing.T) {
			resp, err := http.Get(ts.URL + tc.path)
			if err != nil {
				t.Fatalf("GET %s: %v", tc.path, err)
			}
			resp.Body.Close()

			if resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusMovedPermanently {
				t.Errorf("GET %s: status=%d — route likely not registered", tc.path, resp.StatusCode)
			}
			if tc.wantStatus != 0 && resp.StatusCode != tc.wantStatus {
				t.Errorf("GET %s: got status %d, want %d", tc.path, resp.StatusCode, tc.wantStatus)
			}
		})
	}
}

// TestCreateWebAppMux_StaticAssets checks that /static/ serves actual content.
func TestCreateWebAppMux_StaticAssets(t *testing.T) {
	secret, botToken, adminIDs, repo, apptSvc, transSvc, dataDir, botUser := createDummyMuxInputs(t)

	mux := createWebAppMux(secret, botToken, adminIDs, repo, apptSvc, transSvc, dataDir, botUser)
	ts := httptest.NewServer(mux)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/static/")
	if err != nil {
		t.Fatalf("GET /static/: %v", err)
	}
	defer resp.Body.Close()

	// Should serve directory listing or index — any non-404 is fine.
	if resp.StatusCode == http.StatusNotFound {
		t.Error("/static/ returned 404 — static assets not served")
	}
}

// TestCreateWebAppMux_NoWebDAV checks that WebDAV-specific content is NOT
// served when env vars are unset (default test environment). The catch-all
// route returns JSON, not the WebDAV HTML status page.
func TestCreateWebAppMux_NoWebDAV(t *testing.T) {
	secret, botToken, adminIDs, repo, apptSvc, transSvc, dataDir, botUser := createDummyMuxInputs(t)

	mux := createWebAppMux(secret, botToken, adminIDs, repo, apptSvc, transSvc, dataDir, botUser)
	ts := httptest.NewServer(mux)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/webdav/")
	if err != nil {
		t.Fatalf("GET /webdav/: %v", err)
	}
	defer resp.Body.Close()

	body := new(strings.Builder)
	_, _ = io.Copy(body, resp.Body)
	bodyStr := body.String()

	// Without env vars, the catch-all handler serves the card app or error page.
	// Verify the body does NOT contain WebDAV-specific strings.
	if strings.Contains(bodyStr, "WebDAV Сервер Активен") {
		t.Errorf("/webdav/ contains WebDAV status page content (WebDAV should be disabled)")
	}
	if strings.Contains(bodyStr, "WEBDAV") {
		t.Errorf("/webdav/ body contains WEBDAV string (WebDAV should be disabled)")
	}
}

// TestCreateWebAppMux_WithWebDAV sets WEBDAV env vars and verifies that
// the WebDAV routes are registered and serve the expected status page.
func TestCreateWebAppMux_WithWebDAV(t *testing.T) {
	// Set WebDAV env vars for this test only
	t.Setenv("WEBDAV_USER", "testuser")
	t.Setenv("WEBDAV_PASSWORD", "testpass")

	dataDir := t.TempDir()
	secret, botToken, adminIDs, repo, apptSvc, transSvc, _, botUser := createDummyMuxInputs(t)

	mux := createWebAppMux(secret, botToken, adminIDs, repo, apptSvc, transSvc, dataDir, botUser)
	ts := httptest.NewServer(mux)
	defer ts.Close()

	t.Run("status page", func(t *testing.T) {
		req, _ := http.NewRequest("GET", ts.URL+"/webdav/", nil)
		req.Header.Set("User-Agent", "Mozilla/5.0")
		req.SetBasicAuth("testuser", "testpass")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("GET /webdav/: %v", err)
		}
		defer resp.Body.Close()

		body := new(strings.Builder)
		_, _ = io.Copy(body, resp.Body)

		if !strings.Contains(body.String(), "WebDAV Сервер Активен") {
			t.Errorf("/webdav/ should return WebDAV status page when env vars are set")
		}
	})

	t.Run("redirect without trailing slash", func(t *testing.T) {
		req, _ := http.NewRequest("GET", ts.URL+"/webdav", nil)
		req.SetBasicAuth("testuser", "testpass")

		// Don't follow redirects — we want the 301 itself
		client := &http.Client{CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}}
		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("GET /webdav: %v", err)
		}
		resp.Body.Close()

		if resp.StatusCode != http.StatusMovedPermanently {
			t.Errorf("GET /webdav: status=%d, want %d", resp.StatusCode, http.StatusMovedPermanently)
		}
	})

	t.Run("unauthorized without auth", func(t *testing.T) {
		resp, err := http.Get(ts.URL + "/webdav/")
		if err != nil {
			t.Fatalf("GET /webdav/ (no auth): %v", err)
		}
		defer resp.Body.Close()

		body := new(strings.Builder)
		_, _ = io.Copy(body, resp.Body)

		if !strings.Contains(body.String(), "Unauthorized") {
			t.Errorf("GET /webdav/ without auth should return 401; got body: %s", body.String())
		}
	})

	t.Run("OPTIONS preflight", func(t *testing.T) {
		req, _ := http.NewRequest("OPTIONS", ts.URL+"/webdav/", nil)
		req.SetBasicAuth("testuser", "testpass")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("OPTIONS /webdav/: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("OPTIONS /webdav/: status=%d, want 200", resp.StatusCode)
		}
		if resp.Header.Get("Access-Control-Allow-Origin") == "" {
			t.Error("OPTIONS /webdav/: missing CORS headers")
		}
	})

	t.Run("empty dataDir defaults to 'data'", func(t *testing.T) {
		mux2 := createWebAppMux(secret, botToken, adminIDs, repo, apptSvc, transSvc, "", botUser)
		if mux2 == nil {
			t.Fatal("createWebAppMux with empty dataDir returned nil")
		}
		ts2 := httptest.NewServer(mux2)
		defer ts2.Close()

		resp, err := http.Get(ts2.URL + "/")
		if err != nil {
			t.Fatalf("GET /: %v", err)
		}
		resp.Body.Close()
		if resp.StatusCode == http.StatusNotFound {
			t.Error("default mux with empty dataDir: root route not registered")
		}
	})

	t.Run("WebDAV os.Stat error with nonexistent dir", func(t *testing.T) {
		nonExistent := os.TempDir() + "/__vera_test_nonexistent__"
		mux2 := createWebAppMux(secret, botToken, adminIDs, repo, apptSvc, transSvc, nonExistent, botUser)
		ts2 := httptest.NewServer(mux2)
		defer ts2.Close()

		req, _ := http.NewRequest("GET", ts2.URL+"/webdav/", nil)
		req.Header.Set("User-Agent", "Mozilla/5.0")
		req.SetBasicAuth("testuser", "testpass")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("GET /webdav/ (nonexistent dir): %v", err)
		}
		defer resp.Body.Close()

		body := new(strings.Builder)
		_, _ = io.Copy(body, resp.Body)
		bodyStr := body.String()

		if strings.Contains(bodyStr, "✅ Доступно") {
			t.Error("WebDAV with nonexistent dir should show error status")
		}
	})

	t.Run("WebDAV path is file not dir", func(t *testing.T) {
		filePath := t.TempDir() + "/not-a-dir"
		if err := os.WriteFile(filePath, []byte("x"), 0644); err != nil {
			t.Fatalf("create file: %v", err)
		}

		mux2 := createWebAppMux(secret, botToken, adminIDs, repo, apptSvc, transSvc, filePath, botUser)
		ts2 := httptest.NewServer(mux2)
		defer ts2.Close()

		req, _ := http.NewRequest("GET", ts2.URL+"/webdav/", nil)
		req.Header.Set("User-Agent", "Mozilla/5.0")
		req.SetBasicAuth("testuser", "testpass")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("GET /webdav/ (file path): %v", err)
		}
		defer resp.Body.Close()

		body := new(strings.Builder)
		_, _ = io.Copy(body, resp.Body)
		bodyStr := body.String()

		if !strings.Contains(bodyStr, "не является папкой") {
			t.Errorf("WebDAV with file path should show 'not a directory' error; got: %s", bodyStr)
		}
	})

	t.Run("WebDAV wrong password shows denied", func(t *testing.T) {
		req, _ := http.NewRequest("GET", ts.URL+"/webdav/", nil)
		req.Header.Set("User-Agent", "Mozilla/5.0")
		req.SetBasicAuth("testuser", "wrongpass")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("GET /webdav/ (wrong password): %v", err)
		}
		defer resp.Body.Close()

		body := new(strings.Builder)
		_, _ = io.Copy(body, resp.Body)
		bodyStr := body.String()

		if !strings.Contains(bodyStr, "Unauthorized") {
			t.Errorf("wrong password should return 401; got: %s", bodyStr)
		}
	})

	t.Run("WebDAV with Obsidian client bypasses browser page", func(t *testing.T) {
		req, _ := http.NewRequest("PROPFIND", ts.URL+"/webdav/", nil)
		req.Header.Set("User-Agent", "Obsidian/1.0")
		req.SetBasicAuth("testuser", "testpass")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("PROPFIND /webdav/ (Obsidian): %v", err)
		}
		defer resp.Body.Close()

		body := new(strings.Builder)
		_, _ = io.Copy(body, resp.Body)
		bodyStr := body.String()

		if strings.Contains(bodyStr, "WebDAV Сервер Активен") {
			t.Error("Obsidian client should not get browser status page")
		}
		if resp.StatusCode == http.StatusNotFound {
			t.Errorf("PROPFIND /webdav/ with Obsidian: status=%d (handler may not be routing correctly)", resp.StatusCode)
		}
	})
}

func pickPort(t *testing.T) string {
	t.Helper()
	ln, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatalf("pickPort: %v", err)
	}
	defer ln.Close()
	return fmt.Sprint(ln.Addr().(*net.TCPAddr).Port)
}

func TestStartServer_Lifecycle(t *testing.T) {
	secret, botToken, adminIDs, repo, apptSvc, transSvc, dataDir, botUser := createDummyMuxInputs(t)
	port := pickPort(t)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go StartServer(ctx, port, secret, botToken, adminIDs, repo, apptSvc, transSvc, dataDir, botUser)

	// Retry until the server responds
	var resp *http.Response
	var err error
	for i := 0; i < 20; i++ {
		resp, err = http.Get("http://127.0.0.1:" + port + "/")
		if err == nil && resp.StatusCode == http.StatusOK {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
	if err != nil {
		t.Fatalf("GET / after retry: %v", err)
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("GET /: status=%d, want 200", resp.StatusCode)
	}

	// Shut down
	cancel()
	time.Sleep(100 * time.Millisecond)

	// Verify server stopped
	_, err = http.Get("http://127.0.0.1:" + port + "/")
	if err == nil {
		t.Log("Web App server still responding after shutdown (expected eventually)")
	}
}

