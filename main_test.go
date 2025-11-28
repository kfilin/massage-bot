package main

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	os.Setenv("TELEGRAM_BOT_TOKEN", "test-token")
	os.Setenv("ADMIN_USER_ID", "304528450")
	code := m.Run()
	os.Exit(code)
}

func TestEnvironmentSetup(t *testing.T) {
	if os.Getenv("TELEGRAM_BOT_TOKEN") != "test-token" {
		t.Error("TELEGRAM_BOT_TOKEN not set correctly")
	}
}
