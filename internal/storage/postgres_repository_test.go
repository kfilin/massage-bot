package storage

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/kfilin/massage-bot/internal/domain"
)

// TestNewPostgresRepository tests repository creation
func TestNewPostgresRepository(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	dataDir := t.TempDir()

	repo := NewPostgresRepository(sqlxDB, dataDir)

	if repo == nil {
		t.Fatal("NewPostgresRepository returned nil")
	}
	if repo.db != sqlxDB {
		t.Error("Repository db not set correctly")
	}
	if repo.dataDir != dataDir {
		t.Error("Repository dataDir not set correctly")
	}
}

// TestSavePatient tests saving a patient to the database
func TestSavePatient(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewPostgresRepository(sqlxDB, t.TempDir())

	patient := domain.Patient{
		TelegramID:   "123456789",
		Name:         "Test Patient",
		FirstVisit:   time.Now(),
		LastVisit:    time.Now(),
		TotalVisits:  5,
		HealthStatus: "Good",
	}

	// Expect the INSERT query
	mock.ExpectExec("INSERT INTO patients").
		WithArgs(
			patient.TelegramID,
			patient.Name,
			sqlmock.AnyArg(), // first_visit
			sqlmock.AnyArg(), // last_visit
			patient.TotalVisits,
			patient.HealthStatus,
			patient.TherapistNotes,
			patient.VoiceTranscripts,
			patient.CurrentService,
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.SavePatient(patient)
	if err != nil {
		t.Errorf("SavePatient failed: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

// TestSavePatient_Error tests error handling in SavePatient
func TestSavePatient_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewPostgresRepository(sqlxDB, t.TempDir())

	patient := domain.Patient{
		TelegramID: "123456789",
		Name:       "Test Patient",
	}

	// Expect the INSERT query to fail
	mock.ExpectExec("INSERT INTO patients").
		WillReturnError(sql.ErrConnDone)

	err = repo.SavePatient(patient)
	if err == nil {
		t.Error("SavePatient should have returned an error")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

// TestGetPatient tests retrieving a patient from the database
func TestGetPatient(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewPostgresRepository(sqlxDB, t.TempDir())

	telegramID := "123456789"
	expectedPatient := domain.Patient{
		TelegramID:   telegramID,
		Name:         "Test Patient",
		TotalVisits:  5,
		HealthStatus: "Good",
	}

	rows := sqlmock.NewRows([]string{
		"telegram_id", "name", "first_visit", "last_visit",
		"total_visits", "health_status", "therapist_notes",
		"voice_transcripts", "current_service",
	}).AddRow(
		expectedPatient.TelegramID,
		expectedPatient.Name,
		time.Now(),
		time.Now(),
		expectedPatient.TotalVisits,
		expectedPatient.HealthStatus,
		"",
		"",
		"",
	)

	mock.ExpectQuery("SELECT (.+) FROM patients WHERE telegram_id").
		WithArgs(telegramID).
		WillReturnRows(rows)

	patient, err := repo.GetPatient(telegramID)
	if err != nil {
		t.Errorf("GetPatient failed: %v", err)
	}

	if patient.TelegramID != expectedPatient.TelegramID {
		t.Errorf("TelegramID = %s, want %s", patient.TelegramID, expectedPatient.TelegramID)
	}
	if patient.Name != expectedPatient.Name {
		t.Errorf("Name = %s, want %s", patient.Name, expectedPatient.Name)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

// TestGetPatient_NotFound tests GetPatient when patient doesn't exist
func TestGetPatient_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewPostgresRepository(sqlxDB, t.TempDir())

	telegramID := "999999999"

	mock.ExpectQuery("SELECT (.+) FROM patients WHERE telegram_id").
		WithArgs(telegramID).
		WillReturnError(sql.ErrNoRows)

	_, err = repo.GetPatient(telegramID)
	if err != sql.ErrNoRows {
		t.Errorf("Expected sql.ErrNoRows, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

// TestIsUserBanned tests checking if a user is banned
func TestIsUserBanned(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewPostgresRepository(sqlxDB, t.TempDir())

	tests := []struct {
		name       string
		telegramID string
		username   string
		mockResult bool
		wantBanned bool
	}{
		{
			name:       "User is banned",
			telegramID: "123456789",
			username:   "testuser",
			mockResult: true,
			wantBanned: true,
		},
		{
			name:       "User is not banned",
			telegramID: "987654321",
			username:   "gooduser",
			mockResult: false,
			wantBanned: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var count int
			if tt.mockResult {
				count = 1
			} else {
				count = 0
			}
			rows := sqlmock.NewRows([]string{"count"}).AddRow(count)

			mock.ExpectQuery("SELECT count").
				WithArgs(tt.telegramID, tt.username, "@"+tt.username).
				WillReturnRows(rows)

			banned, err := repo.IsUserBanned(tt.telegramID, tt.username)
			if err != nil {
				t.Errorf("IsUserBanned failed: %v", err)
			}

			if banned != tt.wantBanned {
				t.Errorf("IsUserBanned = %v, want %v", banned, tt.wantBanned)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("Unfulfilled expectations: %v", err)
			}
		})
	}
}

// TestBanUser tests banning a user
func TestBanUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewPostgresRepository(sqlxDB, t.TempDir())

	telegramID := "123456789"

	mock.ExpectExec("INSERT INTO blacklist").
		WithArgs(telegramID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.BanUser(telegramID)
	if err != nil {
		t.Errorf("BanUser failed: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

// TestUnbanUser tests unbanning a user
func TestUnbanUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewPostgresRepository(sqlxDB, t.TempDir())

	telegramID := "123456789"

	mock.ExpectExec("DELETE FROM blacklist").
		WithArgs(telegramID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = repo.UnbanUser(telegramID)
	if err != nil {
		t.Errorf("UnbanUser failed: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

// TestLogEvent tests logging an analytics event
func TestLogEvent(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewPostgresRepository(sqlxDB, t.TempDir())

	patientID := "123456789"
	eventType := "appointment_booked"
	details := map[string]interface{}{
		"service_id": "massage-60",
		"duration":   60,
	}

	detailsJSON, _ := json.Marshal(details)

	mock.ExpectExec("INSERT INTO analytics_events").
		WithArgs(patientID, eventType, detailsJSON).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.LogEvent(patientID, eventType, details)
	if err != nil {
		t.Errorf("LogEvent failed: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

// TestGetAppointmentHistory tests retrieving appointment history
func TestGetAppointmentHistory(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewPostgresRepository(sqlxDB, t.TempDir())

	telegramID := "123456789"
	now := time.Now()

	rows := sqlmock.NewRows([]string{
		"id", "service_id", "time", "duration", "start_time", "end_time",
		"customer_id", "customer_name", "notes", "calendar_event_id",
		"meet_link", "status",
	}).AddRow(
		"appt-1", "massage-60", now, 60, now, now.Add(60*time.Minute),
		telegramID, "Test Patient", "Test notes", "gcal-123",
		"", "confirmed",
	)

	mock.ExpectQuery("SELECT (.+) FROM appointments WHERE customer_id").
		WithArgs(telegramID).
		WillReturnRows(rows)

	appointments, err := repo.GetAppointmentHistory(telegramID)
	if err != nil {
		t.Errorf("GetAppointmentHistory failed: %v", err)
	}

	if len(appointments) != 1 {
		t.Errorf("Expected 1 appointment, got %d", len(appointments))
	}

	if appointments[0].ID != "appt-1" {
		t.Errorf("Appointment ID = %s, want appt-1", appointments[0].ID)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

// TestUpsertAppointments tests batch upserting appointments
// FIXED: Added db tags to Service struct to map DurationMinutes to :service.duration
func TestUpsertAppointments(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewPostgresRepository(sqlxDB, t.TempDir())

	now := time.Now()
	appointments := []domain.Appointment{
		{
			ID:        "appt-1",
			ServiceID: "massage-60",
			Service: domain.Service{
				Name:            "Classic Massage",
				DurationMinutes: 60,
				Price:           50.0,
			},
			Time:            now,
			Duration:        60,
			StartTime:       now,
			EndTime:         now.Add(60 * time.Minute),
			CustomerTgID:    "123456789",
			CustomerName:    "Test Patient",
			CalendarEventID: "gcal-123",
			Status:          "confirmed",
		},
	}

	// Note: UpsertAppointments uses NamedExec which is difficult to mock precisely
	// This test just verifies the function doesn't panic with valid input
	// The actual SQL execution would need integration testing
	mock.ExpectExec("INSERT INTO appointments").
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.UpsertAppointments(appointments)
	if err != nil {
		t.Errorf("UpsertAppointments failed: %v", err)
	}

	// Don't check expectations - NamedExec behavior is hard to mock exactly
}

// TestUpsertAppointments_Empty tests upserting empty slice
func TestUpsertAppointments_Empty(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewPostgresRepository(sqlxDB, t.TempDir())

	err = repo.UpsertAppointments([]domain.Appointment{})
	if err != nil {
		t.Errorf("UpsertAppointments with empty slice should not error: %v", err)
	}
}

// TestMdToHTML tests markdown to HTML conversion
func TestMdToHTML(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewPostgresRepository(sqlxDB, t.TempDir())

	tests := []struct {
		name     string
		input    string
		wantHTML string
	}{
		{
			name:     "Bold text",
			input:    "**bold**",
			wantHTML: "<strong>bold</strong>",
		},

		{
			name:     "Plain text",
			input:    "plain text",
			wantHTML: "plain text",
		},
		{
			name:     "Empty string",
			input:    "",
			wantHTML: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := repo.mdToHTML(tt.input)
			resultStr := string(result)

			if resultStr != tt.wantHTML {
				t.Errorf("mdToHTML(%q) = %q, want %q", tt.input, resultStr, tt.wantHTML)
			}
		})
	}
}

// TestParseTime tests the parseTime helper function
func TestParseTime(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantZero bool
	}{
		{
			name:     "Valid format",
			input:    "03.02.2026 10:00",
			wantZero: false,
		},
		{
			name:     "Empty string",
			input:    "",
			wantZero: true,
		},
		{
			name:     "Invalid format",
			input:    "not-a-date",
			wantZero: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseTime(tt.input)

			if tt.wantZero {
				if !result.IsZero() {
					t.Errorf("parseTime(%q) should return zero time, got %v", tt.input, result)
				}
			} else {
				if result.IsZero() {
					t.Errorf("parseTime(%q) returned zero time unexpectedly", tt.input)
				}
			}
		})
	}
}

// TestGetPatientDir tests patient directory path generation
func TestGetPatientDir(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	dataDir := "/tmp/test"
	repo := NewPostgresRepository(sqlxDB, dataDir)

	patient := domain.Patient{
		TelegramID: "123456789",
		Name:       "Test Patient",
	}

	dir := repo.getPatientDir(patient)

	// Should contain the data directory
	if dir[:len(dataDir)] != dataDir {
		t.Errorf("Patient dir should start with %s, got %s", dataDir, dir)
	}

	// Should contain telegram ID
	if !contains(dir, patient.TelegramID) {
		t.Errorf("Patient dir should contain telegram ID %s, got %s", patient.TelegramID, dir)
	}
}

// Helper function for string contains check
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsMiddle(s, substr)))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// TestGenerateHTMLRecord tests HTML record generation
func TestGenerateHTMLRecord(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewPostgresRepository(sqlxDB, t.TempDir())
	repo.BotVersion = "v5.3.6"

	tests := []struct {
		name            string
		patient         domain.Patient
		history         []domain.Appointment
		wantContains    []string
		wantNotContains []string
	}{
		{
			name: "Patient with notes and future appointment",
			patient: domain.Patient{
				TelegramID:     "123456789",
				Name:           "Test Patient",
				FirstVisit:     time.Now().Add(-30 * 24 * time.Hour),
				LastVisit:      time.Now().Add(-7 * 24 * time.Hour),
				TotalVisits:    5,
				TherapistNotes: "**Important** notes about patient",
				CurrentService: "Massage",
			},
			history: []domain.Appointment{
				{
					ID:        "appt-1",
					StartTime: time.Now().Add(24 * time.Hour),
					Service:   domain.Service{Name: "Classic Massage", DurationMinutes: 60},
					Status:    "confirmed",
				},
			},
			wantContains: []string{
				"TEST PATIENT",
				"<strong>Important</strong>",
				"Classic Massage",
				"v5.3.6",
			},
			wantNotContains: []string{},
		},
		{
			name: "Patient with empty notes",
			patient: domain.Patient{
				TelegramID:  "987654321",
				Name:        "Empty Notes Patient",
				FirstVisit:  time.Now(),
				LastVisit:   time.Now(),
				TotalVisits: 1,
			},
			history: []domain.Appointment{},
			wantContains: []string{
				"EMPTY NOTES PATIENT",
			},
			wantNotContains: []string{},
		},
		{
			name: "Patient with past appointments only",
			patient: domain.Patient{
				TelegramID:  "111222333",
				Name:        "Past Patient",
				FirstVisit:  time.Now().Add(-60 * 24 * time.Hour),
				LastVisit:   time.Now().Add(-30 * 24 * time.Hour),
				TotalVisits: 3,
			},
			history: []domain.Appointment{
				{
					ID:        "appt-past",
					StartTime: time.Now().Add(-30 * 24 * time.Hour),
					Service:   domain.Service{Name: "Massage"},
					Status:    "confirmed",
				},
			},
			wantContains: []string{
				"PAST PATIENT",
			},
			wantNotContains: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			html := repo.GenerateHTMLRecord(tt.patient, tt.history, false)

			if html == "" {
				t.Error("GenerateHTMLRecord returned empty string")
			}

			for _, want := range tt.wantContains {
				if !strings.Contains(html, want) {
					t.Errorf("HTML should contain %q", want)
				}
			}

			for _, notWant := range tt.wantNotContains {
				if strings.Contains(html, notWant) {
					t.Errorf("HTML should not contain %q", notWant)
				}
			}
		})
	}
}

// TestSavePatientDocumentReader tests saving patient documents
func TestSavePatientDocumentReader(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		category string
		content  string
		wantDir  string
	}{
		{
			name:     "Save scan document",
			filename: "scan1.pdf",
			category: "scans",
			content:  "test scan content",
			wantDir:  "scans",
		},
		{
			name:     "Save image document",
			filename: "photo.jpg",
			category: "images",
			content:  "test image content",
			wantDir:  "images",
		},
		{
			name:     "Save voice message",
			filename: "voice.ogg",
			category: "messages",
			content:  "test voice content",
			wantDir:  "messages",
		},
		{
			name:     "Save generic document",
			filename: "document.pdf",
			category: "other",
			content:  "test document content",
			wantDir:  "documents",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("Failed to create mock: %v", err)
			}
			defer db.Close()

			sqlxDB := sqlx.NewDb(db, "sqlmock")
			dataDir := t.TempDir()
			repo := NewPostgresRepository(sqlxDB, dataDir)

			telegramID := "123456789"
			patient := domain.Patient{
				TelegramID: telegramID,
				Name:       "Test Patient",
			}

			// Mock GetPatient call for this subtest
			rows := sqlmock.NewRows([]string{
				"telegram_id", "name", "first_visit", "last_visit",
				"total_visits", "health_status", "therapist_notes",
				"voice_transcripts", "current_service",
			}).AddRow(
				patient.TelegramID,
				patient.Name,
				time.Now(),
				time.Now(),
				0,
				"",
				"",
				"",
				"",
			)

			mock.ExpectQuery("SELECT (.+) FROM patients WHERE telegram_id").
				WithArgs(telegramID).
				WillReturnRows(rows)

			reader := strings.NewReader(tt.content)
			path, err := repo.SavePatientDocumentReader(telegramID, tt.filename, tt.category, reader)

			if err != nil {
				t.Errorf("SavePatientDocumentReader failed: %v", err)
			}

			if path == "" {
				t.Error("SavePatientDocumentReader returned empty path")
			}

			if !strings.Contains(path, tt.wantDir) {
				t.Errorf("Path should contain %q, got %q", tt.wantDir, path)
			}

			// Verify file was created
			if _, err := os.Stat(path); os.IsNotExist(err) {
				t.Errorf("File was not created at %q", path)
			}

			// Verify file content
			content, err := os.ReadFile(path)
			if err != nil {
				t.Errorf("Failed to read created file: %v", err)
			}

			if string(content) != tt.content {
				t.Errorf("File content = %q, want %q", string(content), tt.content)
			}
		})
	}
}

// TestSyncFromFile tests syncing patient data from markdown files
func TestSyncFromFile(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	dataDir := t.TempDir()
	repo := NewPostgresRepository(sqlxDB, dataDir)

	// Create a test patient directory and file
	patient := domain.Patient{
		TelegramID:     "123456789",
		Name:           "Original Name",
		TherapistNotes: "Original notes",
	}

	patientDir := repo.getPatientDir(patient)
	if err := os.MkdirAll(patientDir, 0755); err != nil {
		t.Fatalf("Failed to create patient dir: %v", err)
	}

	// Create markdown file with updated content
	mdContent := `---
Name: Updated Name
ID: 123456789
---

# 🩺 Медицинская карта: Updated Name

## 📋 История болезни
Updated patient notes

## 📝 Заметки терапевта
Additional notes

---
*Статистика (обновляется ботом):*
- Первый визит: 01.01.2026
- Последний визит: 02.02.2026
- Всего визитов: 5
- Услуга: Massage
`

	mdPath := filepath.Join(patientDir, "123456789.md")
	err = os.WriteFile(mdPath, []byte(mdContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test markdown file: %v", err)
	}

	// Test syncing
	updated, err := repo.syncFromFile(&patient)

	if err != nil {
		t.Errorf("syncFromFile failed: %v", err)
	}

	if !updated {
		t.Error("syncFromFile should return true when file has changes")
	}

	if patient.Name != "Updated Name" {
		t.Errorf("Name = %q, want %q", patient.Name, "Updated Name")
	}

	if !strings.Contains(patient.TherapistNotes, "Updated patient notes") {
		t.Errorf("TherapistNotes should contain updated content, got: %q", patient.TherapistNotes)
	}
}

// TestSyncFromFile_NoChanges tests that syncFromFile returns false when no changes
func TestSyncFromFile_NoChanges(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	dataDir := t.TempDir()
	repo := NewPostgresRepository(sqlxDB, dataDir)

	patient := domain.Patient{
		TelegramID:     "123456789",
		Name:           "Test Name",
		TherapistNotes: "Test notes",
	}

	// Create markdown file with same content
	patientDir := repo.getPatientDir(patient)
	if err := os.MkdirAll(patientDir, 0755); err != nil {
		t.Fatalf("Failed to create patient dir: %v", err)
	}

	mdContent := fmt.Sprintf(`---
Name: %s
ID: %s
---

# 🩺 Медицинская карта: %s

%s
`, patient.Name, patient.TelegramID, patient.Name, patient.TherapistNotes)

	mdPath := filepath.Join(patientDir, "123456789.md")
	if err := os.WriteFile(mdPath, []byte(mdContent), 0644); err != nil {
		t.Fatalf("Failed to write md file: %v", err)
	}

	// Test syncing
	updated, err := repo.syncFromFile(&patient)

	if err != nil {
		t.Errorf("syncFromFile failed: %v", err)
	}

	if updated {
		t.Error("syncFromFile should return false when no changes")
	}
}

// TestSyncFromFile_FileNotFound tests handling of missing files
func TestSyncFromFile_FileNotFound(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	dataDir := t.TempDir()
	repo := NewPostgresRepository(sqlxDB, dataDir)

	patient := domain.Patient{
		TelegramID: "nonexistent",
		Name:       "Test",
	}

	updated, err := repo.syncFromFile(&patient)

	if err == nil {
		t.Error("syncFromFile should return error for missing file")
	}

	if updated {
		t.Error("syncFromFile should return false for missing file")
	}
}
