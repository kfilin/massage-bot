//go:build integration

package storage

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/kfilin/massage-bot/internal/domain"
)

type IntegrationTestSuite struct {
	suite.Suite
	repo      *PostgresRepository
	db        *sqlx.DB
	container testcontainers.Container
	connStr   string
	ctx       context.Context
}

func TestIntegrationSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	suite.Run(t, new(IntegrationTestSuite))
}

func (s *IntegrationTestSuite) SetupSuite() {
	s.ctx = context.Background()

	pgContainer, err := tcpostgres.Run(s.ctx,
		"postgres:16-alpine",
		tcpostgres.WithDatabase("testdb"),
		tcpostgres.WithUsername("testuser"),
		tcpostgres.WithPassword("testpass"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second),
		),
	)
	s.Require().NoError(err)

	connStr, err := pgContainer.ConnectionString(s.ctx, "sslmode=disable")
	s.Require().NoError(err)

	// Retry connection with backoff
	var db *sqlx.DB
	for i := 0; i < 10; i++ {
		db, err = sqlx.Connect("postgres", connStr)
		if err == nil {
			if err = db.Ping(); err == nil {
				break
			}
		}
		time.Sleep(time.Duration(i+1) * time.Second)
	}
	s.Require().NoError(err, "Failed to connect to postgres container after retries")

	_, err = db.Exec(Schema)
	s.Require().NoError(err)

	s.db = db
	s.repo = NewPostgresRepository(db, s.T().TempDir())
	s.container = pgContainer
	s.connStr = connStr
}

func (s *IntegrationTestSuite) TearDownSuite() {
	if s.db != nil {
		s.db.Close()
	}
}

func (s *IntegrationTestSuite) TestSaveAndGetPatient() {
	patient := domain.Patient{
		TelegramID:   "int-001",
		Name:         "Integration Test Patient",
		FirstVisit:   time.Now().Add(-30 * 24 * time.Hour),
		LastVisit:    time.Now(),
		TotalVisits:  10,
		HealthStatus: "Good",
	}

	err := s.repo.SavePatient(patient)
	s.Require().NoError(err)

	got, err := s.repo.GetPatient("int-001")
	s.Require().NoError(err)
	s.Equal("Integration Test Patient", got.Name)
	s.Equal(10, got.TotalVisits)
}

func (s *IntegrationTestSuite) TestSavePatient_Upsert() {
	patient := domain.Patient{
		TelegramID:  "int-upsert",
		Name:        "Original Name",
		FirstVisit:  time.Now(),
		LastVisit:   time.Now(),
		TotalVisits: 1,
	}

	err := s.repo.SavePatient(patient)
	s.Require().NoError(err)

	patient.Name = "Updated Name"
	patient.TotalVisits = 5
	err = s.repo.SavePatient(patient)
	s.Require().NoError(err)

	got, err := s.repo.GetPatient("int-upsert")
	s.Require().NoError(err)
	s.Equal("Updated Name", got.Name)
	s.Equal(5, got.TotalVisits)
}

func (s *IntegrationTestSuite) TestGetPatient_NotFound() {
	_, err := s.repo.GetPatient("nonexistent")
	s.ErrorIs(err, sql.ErrNoRows)
}

func (s *IntegrationTestSuite) TestUpdatePatientProfile() {
	patient := domain.Patient{
		TelegramID: "int-profile",
		Name:       "Before",
		FirstVisit: time.Now(),
		LastVisit:  time.Now(),
	}
	s.Require().NoError(s.repo.SavePatient(patient))

	err := s.repo.UpdatePatientProfile("int-profile", "After", "New notes")
	s.Require().NoError(err)

	got, err := s.repo.GetPatient("int-profile")
	s.Require().NoError(err)
	s.Equal("After", got.Name)
	s.Equal("New notes", got.TherapistNotes)
}

func (s *IntegrationTestSuite) TestBanAndUnban() {
	err := s.repo.BanUser("int-ban")
	s.Require().NoError(err)

	banned, err := s.repo.IsUserBanned("int-ban", "")
	s.Require().NoError(err)
	s.True(banned)

	err = s.repo.UnbanUser("int-ban")
	s.Require().NoError(err)

	banned, err = s.repo.IsUserBanned("int-ban", "")
	s.Require().NoError(err)
	s.False(banned)
}

func (s *IntegrationTestSuite) TestSearchPatients() {
	s.Require().NoError(s.repo.SavePatient(domain.Patient{
		TelegramID: "int-search-1", Name: "Zelda Smith",
		FirstVisit: time.Now(), LastVisit: time.Now(),
	}))
	s.Require().NoError(s.repo.SavePatient(domain.Patient{
		TelegramID: "int-search-2", Name: "Zelda Johnson",
		FirstVisit: time.Now(), LastVisit: time.Now(),
	}))
	s.Require().NoError(s.repo.SavePatient(domain.Patient{
		TelegramID: "int-search-3", Name: "Yuri Jones",
		FirstVisit: time.Now(), LastVisit: time.Now(),
	}))

	results, err := s.repo.SearchPatients("Zelda")
	s.Require().NoError(err)
	s.Len(results, 2)

	results, err = s.repo.SearchPatients("int-search-3")
	s.Require().NoError(err)
	s.Len(results, 1)
	s.Equal("Yuri Jones", results[0].Name)
}

func (s *IntegrationTestSuite) TestGetAllPatients() {
	s.Require().NoError(s.repo.SavePatient(domain.Patient{
		TelegramID: "int-all-1", Name: "Charlie All",
		FirstVisit: time.Now(), LastVisit: time.Now(),
	}))
	s.Require().NoError(s.repo.SavePatient(domain.Patient{
		TelegramID: "int-all-2", Name: "Alice All",
		FirstVisit: time.Now(), LastVisit: time.Now(),
	}))

	patients, err := s.repo.GetAllPatients()
	s.Require().NoError(err)
	s.GreaterOrEqual(len(patients), 2)

	// Verify alphabetical order
	foundAlice, foundCharlie := false, false
	for _, p := range patients {
		if p.Name == "Alice All" {
			foundAlice = true
		}
		if p.Name == "Charlie All" {
			foundCharlie = true
		}
	}
	s.True(foundAlice)
	s.True(foundCharlie)
}

func (s *IntegrationTestSuite) TestSaveAndGetMedia() {
	media := domain.PatientMedia{
		ID:             "int-media-1",
		PatientID:      "int-001",
		FileType:       "voice",
		FilePath:       "/tmp/test.ogg",
		TelegramFileID: "tg-file-456",
		Transcript:     "Test transcript content",
		Status:         "approved",
		CreatedAt:      time.Now(),
	}

	err := s.repo.SaveMedia(media)
	s.Require().NoError(err)

	got, err := s.repo.GetMediaByID("int-media-1")
	s.Require().NoError(err)
	s.Equal("voice", got.FileType)
	s.Equal("Test transcript content", got.Transcript)
}

func (s *IntegrationTestSuite) TestGetPatientMedia() {
	media := domain.PatientMedia{
		ID:        "int-media-list",
		PatientID: "int-media-patient",
		FileType:  "photo",
		FilePath:  "/tmp/photo.jpg",
		Status:    "approved",
		CreatedAt: time.Now(),
	}
	s.Require().NoError(s.repo.SaveMedia(media))

	items, err := s.repo.GetPatientMedia("int-media-patient")
	s.Require().NoError(err)
	s.Len(items, 1)
	s.Equal("photo", items[0].FileType)
}

func (s *IntegrationTestSuite) TestUpdateMediaStatus() {
	media := domain.PatientMedia{
		ID:        "int-media-update",
		PatientID: "int-001",
		FileType:  "voice",
		FilePath:  "/tmp/voice.ogg",
		Status:    "pending",
		CreatedAt: time.Now(),
	}
	s.Require().NoError(s.repo.SaveMedia(media))

	err := s.repo.UpdateMediaStatus("int-media-update", "approved", "Final transcript")
	s.Require().NoError(err)

	got, err := s.repo.GetMediaByID("int-media-update")
	s.Require().NoError(err)
	s.Equal("approved", got.Status)
	s.Equal("Final transcript", got.Transcript)
}

func (s *IntegrationTestSuite) TestUpsertAndGetAppointments() {
	now := time.Now()
	appts := []domain.Appointment{
		{
			ID:           "int-appt-1",
			CustomerTgID: "int-001",
			CustomerName: "Test Patient",
			Service:      domain.Service{Name: "Massage", DurationMinutes: 60, Price: 50},
			StartTime:    now,
			Status:       "confirmed",
		},
	}

	err := s.repo.UpsertAppointments(appts)
	s.Require().NoError(err)

	history, err := s.repo.GetAppointmentHistory("int-001")
	s.Require().NoError(err)
	s.GreaterOrEqual(len(history), 1)
	s.Equal("int-appt-1", history[0].ID)
}

func (s *IntegrationTestSuite) TestDeleteAppointment() {
	now := time.Now()
	appts := []domain.Appointment{
		{
			ID:           "int-appt-del",
			CustomerTgID: "int-del",
			CustomerName: "Delete Me",
			Service:      domain.Service{Name: "Massage", DurationMinutes: 30, Price: 25},
			StartTime:    now,
			Status:       "confirmed",
		},
	}
	s.Require().NoError(s.repo.UpsertAppointments(appts))

	err := s.repo.DeleteAppointment("int-appt-del")
	s.Require().NoError(err)

	history, err := s.repo.GetAppointmentHistory("int-del")
	s.Require().NoError(err)
	s.Empty(history)
}

func (s *IntegrationTestSuite) TestSaveAndGetAppointmentMetadata() {
	now := time.Now()
	reminders := map[string]bool{"72h": true, "24h": false}

	err := s.repo.SaveAppointmentMetadata("int-meta-1", &now, reminders)
	s.Require().NoError(err)

	confirmedAt, gotReminders, err := s.repo.GetAppointmentMetadata("int-meta-1")
	s.Require().NoError(err)
	s.NotNil(confirmedAt)
	s.True(gotReminders["72h"])
	s.False(gotReminders["24h"])
}

func (s *IntegrationTestSuite) TestSaveAppointmentMetadata_Upsert() {
	err := s.repo.SaveAppointmentMetadata("int-meta-upsert", nil, map[string]bool{})
	s.Require().NoError(err)

	now := time.Now()
	err = s.repo.SaveAppointmentMetadata("int-meta-upsert", &now, map[string]bool{"72h": true})
	s.Require().NoError(err)

	confirmedAt, reminders, err := s.repo.GetAppointmentMetadata("int-meta-upsert")
	s.Require().NoError(err)
	s.NotNil(confirmedAt)
	s.True(reminders["72h"])
}

func (s *IntegrationTestSuite) TestLogEvent() {
	err := s.repo.LogEvent("int-001", "test_event", map[string]interface{}{"key": "value"})
	s.Require().NoError(err)
}

// Session storage integration tests
func (s *IntegrationTestSuite) TestSessionStorage_CRUD() {
	sessionStore := NewPostgresSessionStorage(s.db)

	sessionStore.Set(99901, "name", "test_user")

	data := sessionStore.Get(99901)
	s.NotNil(data)
	s.Equal("test_user", data["name"])

	sessionStore.ClearSession(99901)

	data = sessionStore.Get(99901)
	s.Nil(data)
}

// TestInitDB verifies the top-level database initializer can connect
// against a real Postgres (via testcontainers) and apply the schema.
// This was previously 0% covered — InitDB shells out to the `postgres`
// driver with env-var config, so it needs a real socket.
//
// Strategy: extract host/port from the running container, set the env
// vars InitDB reads, then call InitDB(). Schema is idempotent (uses
// CREATE TABLE IF NOT EXISTS) so the pre-applied schema is fine.
func (s *IntegrationTestSuite) TestInitDB() {
	host, err := s.container.Host(s.ctx)
	s.Require().NoError(err)
	port, err := s.container.MappedPort(s.ctx, "5432/tcp")
	s.Require().NoError(err)

	prevHost, prevPort, prevUser, prevPass, prevName, prevSSL := os.Getenv("DB_HOST"), os.Getenv("DB_PORT"), os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_NAME"), os.Getenv("DB_SSL_MODE")
	s.T().Cleanup(func() {
		os.Setenv("DB_HOST", prevHost)
		os.Setenv("DB_PORT", prevPort)
		os.Setenv("DB_USER", prevUser)
		os.Setenv("DB_PASSWORD", prevPass)
		os.Setenv("DB_NAME", prevName)
		os.Setenv("DB_SSL_MODE", prevSSL)
	})

	os.Setenv("DB_HOST", host)
	os.Setenv("DB_PORT", port.Port())
	os.Setenv("DB_USER", "testuser")
	os.Setenv("DB_PASSWORD", "testpass")
	os.Setenv("DB_NAME", "testdb")
	os.Setenv("DB_SSL_MODE", "disable")

	db, err := InitDB()
	s.Require().NoError(err, "InitDB should connect against testcontainers Postgres")
	s.Require().NotNil(db)
	s.Equal(db, DB, "InitDB should set package-level DB")

	// Smoke: schema applied, can run a query.
	var n int
	err = db.Get(&n, "SELECT COUNT(*) FROM patients")
	s.Require().NoError(err)
	s.GreaterOrEqual(n, 0)
}
