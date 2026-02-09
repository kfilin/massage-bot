package ports

import (
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
	UpdatePatientProfile(telegramID string, name string, notes string) error

	// Analytics
	LogEvent(patientID string, eventType string, details map[string]interface{}) error

	// Clinical Records & Documents
	// Clinical Records & Documents
	GenerateHTMLRecord(patient domain.Patient, history []domain.Appointment, isAdmin bool) string
	GenerateAdminSearchPage() string
	SaveMedia(media domain.PatientMedia) error
	GetPatientMedia(patientID string) ([]domain.PatientMedia, error)
	GetMediaByID(mediaID string) (*domain.PatientMedia, error)
	CreateBackup() (string, error)
	GetAppointmentHistory(telegramID string) ([]domain.Appointment, error)
	UpsertAppointments(appts []domain.Appointment) error
	DeleteAppointment(appointmentID string) error

	// Appointment Metadata (Reminders/Confirmations)
	SaveAppointmentMetadata(apptID string, confirmedAt *time.Time, remindersSent map[string]bool) error
	GetAppointmentMetadata(apptID string) (confirmedAt *time.Time, remindersSent map[string]bool, err error)
}
