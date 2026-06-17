package telegram

// Tests for the wiring helpers extracted from RunBot (bot.go). These
// functions are the parts of the bot startup that can be exercised
// without a real *telebot.Bot — they consume the ports.BotAPI
// interface so a mock implementation is enough.
//
// Scope:
//   - setupMenuButton: configures the chat menu button via b.Raw
//   - runScheduledBackup: one cycle of the daily backup job
//
// The full RunBot / InitBot flow is not testable in this package
// because *telebot.Bot.Use/Handle/Start/Stop are not on the BotAPI
// interface (they're registration methods on the concrete type).

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/telebot.v3"
)

// =====================================================================
// setupMenuButton
// =====================================================================

// TestSetupMenuButton_EmptyURL exercises the "feature disabled" branch:
// an empty webAppURL means we should not call the Telegram API at all.
func TestSetupMenuButton_EmptyURL(t *testing.T) {
	bot := &mockBotAPI{}

	if err := setupMenuButton(bot, ""); err != nil {
		t.Fatalf("setupMenuButton with empty URL: %v", err)
	}
	if len(bot.rawCalls) != 0 {
		t.Errorf("expected no Raw calls when webAppURL is empty, got %d: %+v",
			len(bot.rawCalls), bot.rawCalls)
	}
}

// TestSetupMenuButton_HappyPath: when a URL is configured, the function
// must call b.Raw("setChatMenuButton", ...) with the expected payload
// shape (web_app type, Russian "Открыть карту" label).
func TestSetupMenuButton_HappyPath(t *testing.T) {
	bot := &mockBotAPI{}

	if err := setupMenuButton(bot, "https://vera.example/app"); err != nil {
		t.Fatalf("setupMenuButton: %v", err)
	}
	if len(bot.rawCalls) != 1 {
		t.Fatalf("expected 1 Raw call, got %d", len(bot.rawCalls))
	}
	got := bot.rawCalls[0]
	if got.method != "setChatMenuButton" {
		t.Errorf("Raw method: got %q, want %q", got.method, "setChatMenuButton")
	}
	payload, ok := got.payload.(map[string]interface{})
	if !ok {
		t.Fatalf("Raw payload type: got %T, want map[string]interface{}", got.payload)
	}
	btn, ok := payload["menu_button"].(map[string]interface{})
	if !ok {
		t.Fatalf("payload[menu_button] type: got %T", payload["menu_button"])
	}
	if btn["type"] != "web_app" {
		t.Errorf("menu_button.type: got %v, want \"web_app\"", btn["type"])
	}
	if btn["text"] != "Открыть карту" {
		t.Errorf("menu_button.text: got %v, want \"Открыть карту\"", btn["text"])
	}
	webapp, ok := btn["web_app"].(map[string]string)
	if !ok {
		t.Fatalf("menu_button.web_app type: got %T", btn["web_app"])
	}
	if webapp["url"] != "https://vera.example/app" {
		t.Errorf("menu_button.web_app.url: got %q, want %q",
			webapp["url"], "https://vera.example/app")
	}
}

// TestSetupMenuButton_RawError: when b.Raw returns an error, setupMenuButton
// must propagate it so the caller (RunBot) can decide to log/warn.
func TestSetupMenuButton_RawError(t *testing.T) {
	bot := &mockBotAPI{rawErr: errors.New("telegram unreachable")}

	err := setupMenuButton(bot, "https://vera.example/app")
	if err == nil {
		t.Fatal("expected error from setupMenuButton when Raw fails, got nil")
	}
	if !strings.Contains(err.Error(), "telegram unreachable") {
		t.Errorf("error should wrap rawErr, got %q", err.Error())
	}
}

// =====================================================================
// runScheduledBackup
// =====================================================================

// TestRunScheduledBackup_CreateBackupError: when the repository fails to
// create a ZIP, the function should not attempt to send anything and
// should return a non-nil error.
func TestRunScheduledBackup_CreateBackupError(t *testing.T) {
	bot := &mockBotAPI{}
	repo := &mockRepository{
		createBackupFunc: func() (string, error) {
			return "", errors.New("disk full")
		},
	}

	err := runScheduledBackup(repo, bot, "1001")
	if err == nil {
		t.Fatal("expected error when CreateBackup fails, got nil")
	}
	if len(bot.sentMessages) != 0 {
		t.Errorf("expected no Send calls when backup creation failed, got %d",
			len(bot.sentMessages))
	}
}

// TestRunScheduledBackup_InvalidAdminID: when the admin ID cannot be
// parsed as int64, the function should return an error and not call
// b.Send. (Validates the parseInt branch.)
func TestRunScheduledBackup_InvalidAdminID(t *testing.T) {
	bot := &mockBotAPI{}
	repo := &mockRepository{
		createBackupFunc: func() (string, error) {
			return "/tmp/backup.zip", nil
		},
	}

	err := runScheduledBackup(repo, bot, "not-a-number")
	if err == nil {
		t.Fatal("expected error when admin ID is not a valid int64, got nil")
	}
	if len(bot.sentMessages) != 0 {
		t.Errorf("expected no Send calls when admin ID is invalid, got %d",
			len(bot.sentMessages))
	}
}

// TestRunScheduledBackup_HappyPath: a real temp file is created, the
// backup is sent to the admin as a Document, and the temp file is
// removed after sending.
func TestRunScheduledBackup_HappyPath(t *testing.T) {
	// Real temp file so os.Remove (called by the function) succeeds.
	dir := t.TempDir()
	zipPath := filepath.Join(dir, "patient_backup_2025-01-01.zip")
	if err := os.WriteFile(zipPath, []byte("PK\x03\x04 fake zip body"), 0o600); err != nil {
		t.Fatalf("setup temp file: %v", err)
	}

	bot := &mockBotAPI{}
	repo := &mockRepository{
		createBackupFunc: func() (string, error) {
			return zipPath, nil
		},
	}

	if err := runScheduledBackup(repo, bot, "1001"); err != nil {
		t.Fatalf("runScheduledBackup: %v", err)
	}
	if len(bot.sentMessages) != 1 {
		t.Fatalf("expected 1 Send call to admin, got %d", len(bot.sentMessages))
	}
	rec := bot.sentMessages[0]

	// Recipient should be the admin
	u, ok := rec.to.(*telebot.User)
	if !ok {
		t.Fatalf("recipient type: got %T, want *telebot.User", rec.to)
	}
	if u.ID != 1001 {
		t.Errorf("recipient ID: got %d, want 1001", u.ID)
	}

	// Payload should be a Document
	doc, ok := rec.what.(*telebot.Document)
	if !ok {
		t.Fatalf("payload type: got %T, want *telebot.Document", rec.what)
	}
	if doc.FileName != "patient_backup_2025-01-01.zip" {
		t.Errorf("Document.FileName: got %q, want %q",
			doc.FileName, "patient_backup_2025-01-01.zip")
	}
	if !strings.Contains(doc.Caption, "Ежедневная копия данных") {
		t.Errorf("Document.Caption should mention daily backup, got %q", doc.Caption)
	}

	// Temp file should have been removed (server disk hygiene)
	if _, err := os.Stat(zipPath); !os.IsNotExist(err) {
		t.Errorf("temp zip should have been removed, stat err = %v", err)
	}
}

// TestRunScheduledBackup_SendError: when b.Send fails, the function
// should return the error. The temp file should still be removed.
func TestRunScheduledBackup_SendError(t *testing.T) {
	dir := t.TempDir()
	zipPath := filepath.Join(dir, "backup.zip")
	if err := os.WriteFile(zipPath, []byte("x"), 0o600); err != nil {
		t.Fatalf("setup temp file: %v", err)
	}

	bot := &mockBotAPI{sendErr: errors.New("network down")}
	repo := &mockRepository{
		createBackupFunc: func() (string, error) {
			return zipPath, nil
		},
	}

	err := runScheduledBackup(repo, bot, "1001")
	if err == nil {
		t.Fatal("expected error when b.Send fails, got nil")
	}
	if _, statErr := os.Stat(zipPath); !os.IsNotExist(statErr) {
		t.Errorf("temp zip should have been removed even on send failure, stat err = %v", statErr)
	}
}
