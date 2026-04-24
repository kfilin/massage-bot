package storage

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/kfilin/massage-bot/internal/logging"

	"github.com/jmoiron/sqlx"
	"github.com/kfilin/massage-bot/internal/domain"
	"github.com/kfilin/massage-bot/internal/monitoring"
	"github.com/kfilin/massage-bot/internal/ports"
)

var _ ports.Repository = (*PostgresRepository)(nil)

type PostgresRepository struct {
	db          *sqlx.DB
	dataDir     string
	BotVersion  string
	BotUsername string
}

func NewPostgresRepository(db *sqlx.DB, dataDir string) *PostgresRepository {
	if dataDir == "" {
		dataDir = "data"
	}
	patientsDir := filepath.Join(dataDir, "patients")
	if err := os.MkdirAll(patientsDir, 0755); err != nil {
		logging.Warnf("Warning: failed to create patients directory: %v", err)
	}
	return &PostgresRepository{
		db:      db,
		dataDir: dataDir,
	}
}

func (r *PostgresRepository) SavePatient(p domain.Patient) error {
	query := `
		INSERT INTO patients (
			telegram_id, name, first_visit, last_visit, total_visits, 
			health_status, therapist_notes, voice_transcripts, current_service
		) VALUES (
			:telegram_id, :name, :first_visit, :last_visit, :total_visits, 
			:health_status, :therapist_notes, :voice_transcripts, :current_service
		) ON CONFLICT (telegram_id) DO UPDATE SET
			name = EXCLUDED.name,
			last_visit = EXCLUDED.last_visit,
			total_visits = EXCLUDED.total_visits,
			health_status = EXCLUDED.health_status,
			therapist_notes = EXCLUDED.therapist_notes,
			voice_transcripts = EXCLUDED.voice_transcripts,
			current_service = EXCLUDED.current_service
	`
	logging.Debugf(": Saving patient record for ID: %s", p.TelegramID)
	_, err := r.db.NamedExec(query, p)
	if err != nil {
		monitoring.DbErrorsTotal.WithLabelValues("save_patient").Inc()
		return err
	}

	// Update clinical note length metric
	noteLen := len(p.TherapistNotes)
	monitoring.ClinicalNoteLength.Set(float64(noteLen))

	return nil
}

// UpdatePatientProfile updates specific fields of a patient profile (Name, Notes)
// This is safer than SavePatient for partial updates as it avoids overwriting other fields
func (r *PostgresRepository) UpdatePatientProfile(telegramID string, name string, notes string) error {
	query := `
		UPDATE patients 
		SET name = :name, therapist_notes = :therapist_notes
		WHERE telegram_id = :telegram_id
	`
	params := map[string]interface{}{
		"telegram_id":     telegramID,
		"name":            name,
		"therapist_notes": notes,
	}

	logging.Debugf(": Updating patient profile for ID: %s", telegramID)
	result, err := r.db.NamedExec(query, params)
	if err != nil {
		monitoring.DbErrorsTotal.WithLabelValues("update_patient_profile").Inc()
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("patient not found")
	}

	// Update clinical note length metric
	monitoring.ClinicalNoteLength.Set(float64(len(notes)))

	return nil
}

func (r *PostgresRepository) getPatientDir(p domain.Patient) string {
	patientsDir := filepath.Join(r.dataDir, "patients")
	// 1. Scan for any folder ending with (ID) - allows for manual renames in Obsidian
	entries, err := os.ReadDir(patientsDir)
	if err == nil {
		suffix := fmt.Sprintf("(%s)", p.TelegramID)
		for _, e := range entries {
			if e.IsDir() && strings.HasSuffix(e.Name(), suffix) {
				return filepath.Join(patientsDir, e.Name())
			}
		}
	}

	// 2. Default fallback if no existing folder found
	reg := regexp.MustCompile(`[<>:"/\\|?*]`)
	cleanName := reg.ReplaceAllString(p.Name, "_")
	folderName := fmt.Sprintf("%s (%s)", cleanName, p.TelegramID)
	return filepath.Join(patientsDir, folderName)
}

func (r *PostgresRepository) GetPatient(telegramID string) (domain.Patient, error) {
	logging.Debugf(": Fetching patient record for ID: %s", telegramID)
	var p domain.Patient
	err := r.db.Get(&p, "SELECT * FROM patients WHERE telegram_id = $1", telegramID)
	return p, err
}

func (r *PostgresRepository) IsUserBanned(telegramID string, username string) (bool, error) {
	var count int
	query := "SELECT count(*) FROM blacklist WHERE telegram_id = $1"
	params := []interface{}{telegramID}
	if username != "" {
		query += " OR username = $2 OR username = $3"
		params = append(params, username, "@"+username)
	}
	logging.Debugf(": Checking ban status for ID: %s, Username: %s", telegramID, username)
	err := r.db.Get(&count, query, params...)
	return count > 0, err
}

func (r *PostgresRepository) BanUser(telegramID string) error {
	_, err := r.db.Exec("INSERT INTO blacklist (telegram_id) VALUES ($1) ON CONFLICT DO NOTHING", telegramID)
	return err
}

func (r *PostgresRepository) UnbanUser(telegramID string) error {
	_, err := r.db.Exec("DELETE FROM blacklist WHERE telegram_id = $1", telegramID)
	return err
}

func (r *PostgresRepository) LogEvent(patientID string, eventType string, details map[string]interface{}) error {
	logging.Debugf(": Logging analytics event: %s for patient: %s", eventType, patientID)
	detailsJSON, err := json.Marshal(details)
	if err != nil {
		return fmt.Errorf("failed to marshal event details: %w", err)
	}
	_, err = r.db.Exec("INSERT INTO analytics_events (patient_id, event_type, details) VALUES ($1, $2, $3)", patientID, eventType, detailsJSON)
	return err
}

// GetAppointmentHistory retrieves all appointments for a patient from the database
func (r *PostgresRepository) GetAppointmentHistory(telegramID string) ([]domain.Appointment, error) {
	var appts []domain.Appointment

	// Denormalized fetch (no JOINs needed as we store service details)
	query := `
		SELECT id, customer_id, service_id, start_time, status, customer_name,
		       service_name as "service.name", service_duration as "service.duration", service_price as "service.price"
		FROM appointments
		WHERE customer_id = $1
		ORDER BY start_time DESC
	`

	err := r.db.Select(&appts, query, telegramID)
	if err != nil {
		return nil, err
	}

	return appts, nil
}

func (r *PostgresRepository) SaveMedia(media domain.PatientMedia) error {
	if media.Status == "" {
		media.Status = "approved"
	}
	query := `
		INSERT INTO patient_media (id, patient_id, file_type, file_path, telegram_file_id, transcript, status, created_at)
		VALUES (:id, :patient_id, :file_type, :file_path, :telegram_file_id, :transcript, :status, :created_at)
		ON CONFLICT (id) DO UPDATE SET
			transcript = EXCLUDED.transcript,
			status = EXCLUDED.status
	`
	_, err := r.db.NamedExec(query, media)
	if err != nil {
		return fmt.Errorf("failed to save media: %w", err)
	}
	return nil
}

func (r *PostgresRepository) UpdateMediaStatus(mediaID string, status string, transcript string) error {
	query := `UPDATE patient_media SET status = $1, transcript = $2 WHERE id = $3`
	_, err := r.db.Exec(query, status, transcript, mediaID)
	return err
}

func (r *PostgresRepository) GetPatientMedia(patientID string) ([]domain.PatientMedia, error) {
	var media []domain.PatientMedia
	err := r.db.Select(&media, "SELECT * FROM patient_media WHERE patient_id = $1 ORDER BY created_at DESC", patientID)
	if err != nil {
		return nil, err
	}
	return media, nil
}

// GetMediaByID retrieves a single media record by ID
func (r *PostgresRepository) GetMediaByID(mediaID string) (*domain.PatientMedia, error) {
	var media domain.PatientMedia
	err := r.db.Get(&media, "SELECT * FROM patient_media WHERE id = $1", mediaID)
	if err != nil {
		return nil, err
	}
	return &media, nil
}

// UpsertAppointments batch inserts or updates appointments in the database
func (r *PostgresRepository) UpsertAppointments(appts []domain.Appointment) error {
	if len(appts) == 0 {
		return nil
	}

	query := `
		INSERT INTO appointments (id, customer_id, service_id, start_time, status, customer_name, 
		                          service_name, service_duration, service_price, created_at, updated_at)
		VALUES (:id, :customer_id, :service_id, :start_time, :status, :customer_name, 
		        :service.name, :service.duration, :service.price, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		ON CONFLICT (id) DO UPDATE SET
			customer_id = EXCLUDED.customer_id,
			service_id = EXCLUDED.service_id,
			start_time = EXCLUDED.start_time,
			status = EXCLUDED.status,
			customer_name = EXCLUDED.customer_name,
			service_name = EXCLUDED.service_name,
			service_duration = EXCLUDED.service_duration,
			service_price = EXCLUDED.service_price,
			updated_at = CURRENT_TIMESTAMP;
	`

	// Create a named prepared statement for better performance (optional, direct NamedExec is fine for now)
	_, err := r.db.NamedExec(query, appts)
	if err != nil {
		return fmt.Errorf("failed to batch upsert appointments: %w", err)
	}
	return nil
}


func (r *PostgresRepository) CreateBackup() (string, error) {
	logging.Debugf(": Starting database backup creation tool...")
	timestamp := time.Now().Format("20060102_150405")
	backupDir := filepath.Join(r.dataDir, "temp_backups")
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create backup directory: %w", err)
	}

	sqlFile := filepath.Join(backupDir, fmt.Sprintf("db_dump_%s.sql", timestamp))
	zipFile := filepath.Join(r.dataDir, fmt.Sprintf("backup_%s.zip", timestamp))

	// 1. Perform Database Dump
	// Map our DB environment variables to pg_dump expected ones
	cmd := exec.Command("pg_dump", "-f", sqlFile)
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("PGDATABASE=%s", os.Getenv("DB_NAME")),
		fmt.Sprintf("PGUSER=%s", os.Getenv("DB_USER")),
		fmt.Sprintf("PGPASSWORD=%s", os.Getenv("DB_PASSWORD")),
		fmt.Sprintf("PGHOST=%s", os.Getenv("DB_HOST")),
		fmt.Sprintf("PGPORT=%s", os.Getenv("DB_PORT")),
	)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("pg_dump failed: %v (stderr: %s)", err, stderr.String())
	}

	// 2. Create ZIP archive
	newZipFile, err := os.Create(zipFile)
	if err != nil {
		return "", err
	}
	defer newZipFile.Close()

	zipWriter := zip.NewWriter(newZipFile)
	defer zipWriter.Close()

	// Add SQL dump to ZIP
	if err := r.addFileToZip(zipWriter, sqlFile, "db_dump.sql"); err != nil {
		return "", err
	}

	// Add patients directory to ZIP
	patientsDir := filepath.Join(r.dataDir, "patients")
	err = filepath.Walk(patientsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		relPath, _ := filepath.Rel(r.dataDir, path)
		return r.addFileToZip(zipWriter, path, relPath)
	})
	if err != nil {
		return "", err
	}

	// 3. Cleanup temp files
	os.RemoveAll(backupDir)

	return zipFile, nil
}

func (r *PostgresRepository) addFileToZip(zipWriter *zip.Writer, filePath string, zipPath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	writer, err := zipWriter.Create(zipPath)
	if err != nil {
		return err
	}

	_, err = io.Copy(writer, file)
	return err
}

func (r *PostgresRepository) SaveAppointmentMetadata(apptID string, confirmedAt *time.Time, remindersSent map[string]bool) error {
	remindersSentJSON, err := json.Marshal(remindersSent)
	if err != nil {
		return fmt.Errorf("failed to marshal reminders_sent: %w", err)
	}

	query := `
		INSERT INTO appointment_metadata (appointment_id, confirmed_at, reminders_sent, updated_at)
		VALUES ($1, $2, $3, CURRENT_TIMESTAMP)
		ON CONFLICT (appointment_id) DO UPDATE SET
			confirmed_at = EXCLUDED.confirmed_at,
			reminders_sent = EXCLUDED.reminders_sent,
			updated_at = CURRENT_TIMESTAMP
	`
	_, err = r.db.Exec(query, apptID, confirmedAt, remindersSentJSON)
	if err != nil {
		monitoring.DbErrorsTotal.WithLabelValues("save_appointment_metadata").Inc()
		return fmt.Errorf("failed to save appointment metadata: %w", err)
	}
	return nil
}

func (r *PostgresRepository) GetAppointmentMetadata(apptID string) (*time.Time, map[string]bool, error) {
	var row struct {
		ConfirmedAt   *time.Time `db:"confirmed_at"`
		RemindersSent []byte     `db:"reminders_sent"`
	}

	query := "SELECT confirmed_at, reminders_sent FROM appointment_metadata WHERE appointment_id = $1"
	err := r.db.Get(&row, query, apptID)
	if err != nil {
		return nil, nil, err
	}

	remindersSent := make(map[string]bool)
	if len(row.RemindersSent) > 0 {
		if err := json.Unmarshal(row.RemindersSent, &remindersSent); err != nil {
			return nil, nil, fmt.Errorf("failed to unmarshal reminders_sent: %w", err)
		}
	}

	return row.ConfirmedAt, remindersSent, nil
}

// DeleteAppointment deletes an appointment from the database by ID
func (r *PostgresRepository) DeleteAppointment(appointmentID string) error {
	query := `DELETE FROM appointments WHERE id = $1`
	result, err := r.db.Exec(query, appointmentID)
	if err != nil {
		return fmt.Errorf("failed to delete appointment from database: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		logging.Debugf("DeleteAppointment: No rows deleted for appointment ID %s (may not exist in DB)", appointmentID)
	} else {
		logging.Debugf("DeleteAppointment: Successfully deleted appointment ID %s from database", appointmentID)
	}

	return nil
}

// SearchPatients finds patients by name or Telegram ID matching the query
func (r *PostgresRepository) SearchPatients(query string) ([]domain.Patient, error) {
	var patients []domain.Patient
	// Case-insensitive search on name, or exact match on telegram_id
	sqlQuery := `
		SELECT * FROM patients 
		WHERE name ILIKE $1 OR telegram_id = $2
		ORDER BY name ASC
		LIMIT 20
	`
	// Match anything containing the query string for name
	searchPattern := "%" + query + "%"

	err := r.db.Select(&patients, sqlQuery, searchPattern, query)
	if err != nil {
		return nil, fmt.Errorf("failed to search patients: %w", err)
	}
	return patients, nil
}
