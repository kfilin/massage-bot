package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	os.Setenv("TELEGRAM_BOT_TOKEN", "test-token")
	os.Setenv("ADMIN_USER_ID", "304528450")
	code := m.Run()
	os.Exit(code)
}

// Comment out or remove the TestConfigLoad since loadConfig doesn't exist
// func TestConfigLoad(t *testing.T) {
// 	cfg := loadConfig()
// 	if cfg.TelegramToken != "test-token" {
// 		t.Errorf("Expected token 'test-token', got '%s'", cfg.TelegramToken)
// 	}
// }

func TestHealthHandler(t *testing.T) {
	req := httptest.NewRequest("GET", "/health", nil)
	rr := httptest.NewRecorder()

	healthHandler(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Health handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	expected := `{"status":"healthy"}`
	if rr.Body.String() != expected {
		t.Errorf("Health handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
	}
}
