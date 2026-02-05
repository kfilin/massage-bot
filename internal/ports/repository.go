package ports

import (
	"io"
	"time"

	"github.com/kfilin/massage-bot/internal/domain"
)

type Repository interface {
	SavePatient(patient domain.Patient) error
	GetPatient(telegramID string) (domain.Patient, error)
	SearchPatients(query string) ([]domain.Patient, error)
	IsUserBanned(telegramID string, username string) (bool, error)
	BanUser(telegramID string) error
	UnbanUser(telegramID string) error

	// Analytics
	LogEvent(patientID string, eventType string, details map[string]interface{}) error

	// Clinical Records & Documents
	GenerateHTMLRecord(patient domain.Patient, history []domain.Appointment) string
	GenerateAdminSearchPage() string
	SavePatientDocumentReader(telegramID string, filename string, category string, r io.Reader) (string, error)
	CreateBackup() (string, error)
	SyncAll() error
	MigrateFolderNames() error
	GetAppointmentHistory(telegramID string) ([]domain.Appointment, error)
	UpsertAppointments(appts []domain.Appointment) error
	DeleteAppointment(appointmentID string) error

	// Appointment Metadata (Reminders/Confirmations)
	SaveAppointmentMetadata(apptID string, confirmedAt *time.Time, remindersSent map[string]bool) error
	GetAppointmentMetadata(apptID string) (confirmedAt *time.Time, remindersSent map[string]bool, err error)
}
