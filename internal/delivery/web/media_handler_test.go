package web

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/kfilin/massage-bot/internal/domain"
)

type mockMediaRepo struct {
	media map[string]domain.PatientMedia
}

func (m *mockMediaRepo) GetMediaByID(id string) (*domain.PatientMedia, error) {
	if media, ok := m.media[id]; ok {
		return &media, nil
	}
	return nil, fmt.Errorf("not found")
}

func (m *mockMediaRepo) GetPatient(id string) (domain.Patient, error) {
	return domain.Patient{}, nil
}
func (m *mockMediaRepo) GetAllPatients() ([]domain.Patient, error)        { return nil, nil }
func (m *mockMediaRepo) SearchPatients(q string) ([]domain.Patient, error) { return nil, nil }
func (m *mockMediaRepo) SavePatient(p domain.Patient) error               { return nil }
func (m *mockMediaRepo) UpdatePatientProfile(id, name, notes string) error { return nil }
func (m *mockMediaRepo) IsUserBanned(id, username string) (bool, error)   { return false, nil }
func (m *mockMediaRepo) BanUser(id string) error                          { return nil }
func (m *mockMediaRepo) UnbanUser(id string) error                        { return nil }
func (m *mockMediaRepo) LogEvent(id, et string, d map[string]interface{}) error { return nil }
func (m *mockMediaRepo) SaveMedia(media domain.PatientMedia) error        { return nil }
func (m *mockMediaRepo) GetPatientMedia(id string) ([]domain.PatientMedia, error) {
	return nil, nil
}
func (m *mockMediaRepo) UpdateMediaStatus(id, status, transcript string) error { return nil }
func (m *mockMediaRepo) CreateBackup() (string, error)                     { return "", nil }
func (m *mockMediaRepo) GetAppointmentHistory(id string) ([]domain.Appointment, error) {
	return nil, nil
}
func (m *mockMediaRepo) GetAppointmentHistoryPaginated(id string, limit, offset int) ([]domain.Appointment, bool, error) {
	return nil, false, nil
}
func (m *mockMediaRepo) UpsertAppointments(a []domain.Appointment) error { return nil }
func (m *mockMediaRepo) SaveAppointmentMetadata(id string, t *time.Time, reminders map[string]bool) error {
	return nil
}
func (m *mockMediaRepo) GetAppointmentMetadata(id string) (*time.Time, map[string]bool, error) {
	return nil, nil, nil
}
func (m *mockMediaRepo) DeleteAppointment(id string) error { return nil }

func TestGenerateAuthCookie(t *testing.T) {
	cookie := GenerateAuthCookie("12345", "test-secret")

	parts := strings.Split(cookie, ":")
	if len(parts) != 3 {
		t.Fatalf("Expected 3 parts in cookie, got %d", len(parts))
	}
	if parts[0] != "12345" {
		t.Errorf("Expected telegramID 12345, got %s", parts[0])
	}
}

func TestValidateSignature_Valid(t *testing.T) {
	handler := &MediaHandler{secret: "test-secret"}

	cookie := GenerateAuthCookie("12345", "test-secret")
	parts := strings.Split(cookie, ":")

	valid := handler.validateSignature(parts[0], parts[1], parts[2])
	if !valid {
		t.Error("Expected valid signature")
	}
}

func TestValidateSignature_Expired(t *testing.T) {
	handler := &MediaHandler{secret: "test-secret"}

	oldTimestamp := fmt.Sprintf("%d", time.Now().Unix()-86401) // 24h+ ago
	mac := generateTestHMAC("12345", oldTimestamp, "test-secret")

	valid := handler.validateSignature("12345", oldTimestamp, mac)
	if valid {
		t.Error("Expected expired token to be invalid")
	}
}

func TestValidateSignature_InvalidFormat(t *testing.T) {
	handler := &MediaHandler{secret: "test-secret"}

	valid := handler.validateSignature("12345", "not-a-number", "sig")
	if valid {
		t.Error("Expected invalid timestamp to fail")
	}
}

func TestValidateSignature_WrongSecret(t *testing.T) {
	handler := &MediaHandler{secret: "wrong-secret"}

	cookie := GenerateAuthCookie("12345", "correct-secret")
	parts := strings.Split(cookie, ":")

	valid := handler.validateSignature(parts[0], parts[1], parts[2])
	if valid {
		t.Error("Expected wrong secret to fail")
	}
}

func TestGetMedia_MissingID(t *testing.T) {
	handler := &MediaHandler{repo: &mockMediaRepo{}, secret: "s"}

	req := httptest.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()

	handler.GetMedia(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected 400, got %d", rr.Code)
	}
}

func TestGetMedia_NoCookie(t *testing.T) {
	handler := &MediaHandler{repo: &mockMediaRepo{}, secret: "s"}

	req := httptest.NewRequest("GET", "/media-1", nil)
	rr := httptest.NewRecorder()

	handler.GetMedia(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("Expected 401, got %d", rr.Code)
	}
}

func TestGetMedia_InvalidCookieFormat(t *testing.T) {
	handler := &MediaHandler{repo: &mockMediaRepo{}, secret: "s"}

	req := httptest.NewRequest("GET", "/media-1", nil)
	req.AddCookie(&http.Cookie{Name: "vera_auth", Value: "invalid"})
	rr := httptest.NewRecorder()

	handler.GetMedia(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("Expected 401, got %d", rr.Code)
	}
}

func TestGetMedia_ForbiddenAccess(t *testing.T) {
	repo := &mockMediaRepo{
		media: map[string]domain.PatientMedia{
			"media-1": {ID: "media-1", PatientID: "owner-1", FilePath: "/tmp/test.txt"},
		},
	}
	handler := &MediaHandler{repo: repo, secret: "s", adminIDs: []string{"admin-1"}}

	cookie := GenerateAuthCookie("other-user", "s")
	req := httptest.NewRequest("GET", "/media-1", nil)
	req.AddCookie(&http.Cookie{Name: "vera_auth", Value: cookie})
	rr := httptest.NewRecorder()

	handler.GetMedia(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("Expected 403, got %d", rr.Code)
	}
}

func TestGetMedia_AdminAccess(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.txt")
	_ = os.WriteFile(tmpFile, []byte("hello"), 0644)

	// Set DATA_DIR to temp dir so path traversal check passes
	os.Setenv("DATA_DIR", tmpDir)
	defer os.Unsetenv("DATA_DIR")

	repo := &mockMediaRepo{
		media: map[string]domain.PatientMedia{
			"media-1": {ID: "media-1", PatientID: "owner-1", FilePath: tmpFile},
		},
	}
	handler := &MediaHandler{repo: repo, secret: "s", adminIDs: []string{"admin-1"}}

	cookie := GenerateAuthCookie("admin-1", "s")
	req := httptest.NewRequest("GET", "/media-1", nil)
	req.AddCookie(&http.Cookie{Name: "vera_auth", Value: cookie})
	rr := httptest.NewRecorder()

	handler.GetMedia(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", rr.Code)
	}
	if rr.Body.String() != "hello" {
		t.Errorf("Expected 'hello', got '%s'", rr.Body.String())
	}
}

func TestGetMedia_PathTraversal(t *testing.T) {
	repo := &mockMediaRepo{
		media: map[string]domain.PatientMedia{
			"media-1": {ID: "media-1", PatientID: "admin-1", FilePath: "../../../etc/passwd"},
		},
	}
	handler := &MediaHandler{repo: repo, secret: "s", adminIDs: []string{"admin-1"}}

	cookie := GenerateAuthCookie("admin-1", "s")
	req := httptest.NewRequest("GET", "/media-1", nil)
	req.AddCookie(&http.Cookie{Name: "vera_auth", Value: cookie})
	rr := httptest.NewRecorder()

	handler.GetMedia(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("Expected 403 for path traversal, got %d", rr.Code)
	}
}

func TestNewMediaHandler(t *testing.T) {
	handler := NewMediaHandler(&mockMediaRepo{}, "secret", []string{"1", "2"})
	if handler == nil {
		t.Fatal("Expected non-nil handler")
	}
	if handler.secret != "secret" {
		t.Errorf("Expected secret 'secret', got '%s'", handler.secret)
	}
	if len(handler.adminIDs) != 2 {
		t.Errorf("Expected 2 admin IDs, got %d", len(handler.adminIDs))
	}
}

func generateTestHMAC(id, timestamp, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(id + ":" + timestamp))
	return hex.EncodeToString(mac.Sum(nil))
}
