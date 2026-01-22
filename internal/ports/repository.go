package ports

import (
	"io"

	"github.com/kfilin/massage-bot/internal/domain"
)

type Repository interface {
	SavePatient(patient domain.Patient) error
	GetPatient(telegramID string) (domain.Patient, error)
	IsUserBanned(telegramID string, username string) (bool, error)
	BanUser(telegramID string) error
	UnbanUser(telegramID string) error

	// Analytics
	LogEvent(patientID string, eventType string, details map[string]interface{}) error

	// Clinical Records & Documents
	GenerateHTMLRecord(patient domain.Patient) string
	SavePatientDocumentReader(telegramID string, filename string, category string, r io.Reader) (string, error)
	CreateBackup() (string, error)
	SyncAll() error
	MigrateFolderNames() error
}
