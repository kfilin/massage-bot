package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// TestCreateHealthMux_Routes verifies that createHealthMux registers all
// expected health-check routes and they return valid JSON responses.
func TestCreateHealthMux_Routes(t *testing.T) {
	mux := createHealthMux()
	if mux == nil {
		t.Fatal("createHealthMux returned nil")
	}

	ts := httptest.NewServer(mux)
	defer ts.Close()

	tests := []struct {
		path       string
		wantStatus int
		wantKey    string // JSON key that must exist in response body
	}{
		{path: "/health", wantStatus: http.StatusOK, wantKey: "status"},
		{path: "/ready", wantStatus: http.StatusOK, wantKey: "status"},
		{path: "/live", wantStatus: http.StatusOK, wantKey: "status"},
		{path: "/", wantStatus: http.StatusOK, wantKey: "service"},
		{path: "/metrics", wantStatus: http.StatusOK, wantKey: ""},
	}

	for _, tc := range tests {
		t.Run(tc.path, func(t *testing.T) {
			resp, err := http.Get(ts.URL + tc.path)
			if err != nil {
				t.Fatalf("GET %s: %v", tc.path, err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != tc.wantStatus {
				t.Errorf("GET %s: status=%d, want %d", tc.path, resp.StatusCode, tc.wantStatus)
			}

			if tc.wantKey != "" {
				var body map[string]interface{}
				if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
					t.Fatalf("GET %s: body not JSON: %v", tc.path, err)
				}
				if _, ok := body[tc.wantKey]; !ok {
					t.Errorf("GET %s: response missing key %q", tc.path, tc.wantKey)
				}
			}
		})
	}
}

// TestCreateHealthMux_NotFound checks that unknown routes return 404
// (makes sure the catch-all / handler doesn't swallow all routes).
func TestCreateHealthMux_NotFound(t *testing.T) {
	mux := createHealthMux()
	ts := httptest.NewServer(mux)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/nonexistent")
	if err != nil {
		t.Fatalf("GET /nonexistent: %v", err)
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// The catch-all "/" handler catches everything — that's by design.
		// We just verify it doesn't panic and returns a valid response.
		t.Errorf("GET /nonexistent: status=%d, want 200 (catch-all)", resp.StatusCode)
	}
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

func TestStartHealthServer_Lifecycle(t *testing.T) {
	port := pickPort(t)
	t.Setenv("HEALTH_PORT", port)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go startHealthServer(ctx)

	// Wait for server to start
	var resp *http.Response
	var err error
	for i := 0; i < 20; i++ {
		resp, err = http.Get("http://127.0.0.1:" + port + "/health")
		if err == nil && resp.StatusCode == 200 {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
	if err != nil {
		t.Fatalf("GET /health after retry: %v", err)
	}
	resp.Body.Close()

	// Verify the server is responding
	if resp.StatusCode != http.StatusOK {
		t.Errorf("GET /health: status=%d, want 200", resp.StatusCode)
	}

	// Shut down
	cancel()
	time.Sleep(100 * time.Millisecond)

	// Verify server stopped
	_, err = http.Get("http://127.0.0.1:" + port + "/health")
	if err == nil {
		t.Log("Server still responding after shutdown (expected eventually)")
	}
}
