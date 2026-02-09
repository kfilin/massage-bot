package storage

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"html"
	"html/template"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
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
	query := `
		INSERT INTO patient_media (id, patient_id, file_type, file_path, telegram_file_id, created_at)
		VALUES (:id, :patient_id, :file_type, :file_path, :telegram_file_id, :created_at)
		ON CONFLICT (id) DO NOTHING
	`
	_, err := r.db.NamedExec(query, media)
	if err != nil {
		return fmt.Errorf("failed to save media: %w", err)
	}
	return nil
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

func (r *PostgresRepository) mdToHTML(md string) template.HTML {
	if md == "" {
		return template.HTML("")
	}

	// 1. Try to unescape if it was accidentally saved as double-escaped HTML
	// This helps with older records that might have been saved incorrectly.
	h := html.UnescapeString(md)

	// 2. Simple Markdown logic (order matters)
	// Headers
	reH3 := regexp.MustCompile(`(?m)^### (.*)$`)
	h = reH3.ReplaceAllString(h, "<h3>$1</h3>")
	reH2 := regexp.MustCompile(`(?m)^## (.*)$`)
	h = reH2.ReplaceAllString(h, "<h2>$1</h2>")
	reH1 := regexp.MustCompile(`(?m)^# (.*)$`)
	h = reH1.ReplaceAllString(h, "<h1>$1</h1>")

	// Bold
	reBold := regexp.MustCompile(`\*\*(.*?)\*\*`)
	h = reBold.ReplaceAllString(h, "<strong>$1</strong>")

	// Lists
	reList := regexp.MustCompile(`(?m)^[*-] (.*)$`)
	h = reList.ReplaceAllString(h, "• $1")

	// Line breaks: ONLY if there are no HTML tags already doing line breaks
	// If the text looks like plain text (\n present, no <h2> or <br>), convert \n to <br>
	if !strings.Contains(h, "<h") && !strings.Contains(h, "<br") && !strings.Contains(h, "<p") {
		h = strings.ReplaceAll(h, "\n", "<br>")
	} else {
		// If it has HTML, we still might want to preserve single \n as <br>?
		// But let's be careful not to double up.
		// For now, let's just do it if it's not looking like a full HTML doc.
		h = strings.ReplaceAll(h, "\n", "<br>")
	}

	return template.HTML(h)
}

func (r *PostgresRepository) GenerateHTMLRecord(p domain.Patient, history []domain.Appointment, isAdmin bool) string {
	type docGroup struct {
		Name   string
		Count  int
		Latest string
		Files  []domain.PatientMedia
	}
	type visitInfo struct {
		Date    string
		Service string
	}
	type futureInfo struct {
		ID        string
		Date      string
		Service   string
		CanCancel bool
	}
	type templateData struct {
		Name               string
		TelegramID         string
		TotalVisits        int
		GeneratedAt        string
		CurrentService     string
		BotVersion         string
		TherapistNotes     template.HTML
		RawNotes           string
		VoiceTranscripts   template.HTML
		FirstVisit         time.Time
		LastVisit          time.Time
		FirstVisitLink     string
		NextVisitLink      string // Renamed from LastVisitLink for clarity in countdown
		ShowFirstVisitLink bool
		ShowNextVisitLink  bool // Renamed from ShowLastVisitLink
		DocGroups          []docGroup
		RecentVisits       []visitInfo
		FutureAppointments []futureInfo
		NextApptUnix       int64
		IsAdmin            bool
		BotUsername        string
		Media              []domain.PatientMedia
	}

	getCalLink := func(t time.Time, service string) string {
		start := t.Format("20060102T150405")
		end := t.Add(time.Hour).Format("20060102T150405")
		title := "Массаж: " + service
		return fmt.Sprintf("https://www.google.com/calendar/render?action=TEMPLATE&text=%s&dates=%s/%s", strings.ReplaceAll(title, " ", "+"), start, end)
	}

	// We'll keep emojis for now as they are part of modern UI
	// re := regexp.MustCompile(`[\x{1F300}-\x{1FAD6}]|[\x{2600}-\x{26FF}]|[\x{2700}-\x{27BF}]|[\x{1F600}-\x{1F64F}]|[\x{1F680}-\x{1F6FF}]|[\x{1F1E6}-\x{1F1FF}]`)
	// cleanNotes := re.ReplaceAllString(p.TherapistNotes, "")
	cleanNotes := p.TherapistNotes

	mediaList, errMedia := r.GetPatientMedia(p.TelegramID)
	if errMedia != nil {
		logging.Warnf("Failed to fetch media for patient %s: %v", p.TelegramID, errMedia)
	}

	data := templateData{
		Name:               strings.ToUpper(p.Name),
		TelegramID:         p.TelegramID,
		TotalVisits:        p.TotalVisits,
		GeneratedAt:        time.Now().Format("02.01.2006 15:04"),
		CurrentService:     p.CurrentService,
		BotVersion:         r.BotVersion,
		TherapistNotes:     r.mdToHTML(cleanNotes),
		RawNotes:           p.TherapistNotes,
		VoiceTranscripts:   template.HTML(strings.ReplaceAll(template.HTMLEscapeString(p.VoiceTranscripts), "\n", "<br>")),
		FirstVisit:         p.FirstVisit,
		LastVisit:          p.LastVisit,
		FirstVisitLink:     getCalLink(p.FirstVisit, p.CurrentService),
		ShowFirstVisitLink: p.FirstVisit.After(time.Now()),
		IsAdmin:            isAdmin,

		BotUsername: r.BotUsername,
		Media:       mediaList,
	}

	if r.BotUsername == "" {
		logging.Warn("WARNING: BotUsername is empty in PostgresRepository during HTML generation! TWA links will be broken.")
	} else {
		logging.Debugf("GenerateHTMLRecord using BotUsername: %s", r.BotUsername)
	}

	// Identify Future Appointments and Next Appointment for Countdown
	now := time.Now().In(domain.ApptTimeZone)
	var futureAppts []domain.Appointment
	for _, a := range history {
		if a.Status != "cancelled" && a.StartTime.After(now) && !strings.Contains(strings.ToLower(a.Service.Name), "block") {
			futureAppts = append(futureAppts, a)
		}
	}
	sort.Slice(futureAppts, func(i, j int) bool {
		return futureAppts[i].StartTime.Before(futureAppts[j].StartTime)
	})

	if len(futureAppts) > 0 {
		next := futureAppts[0]
		data.NextApptUnix = next.StartTime.Unix()
		data.NextVisitLink = getCalLink(next.StartTime, next.Service.Name)
		data.ShowNextVisitLink = true

		for _, a := range futureAppts {
			data.FutureAppointments = append(data.FutureAppointments, futureInfo{
				ID:        a.ID,
				Date:      a.StartTime.In(domain.ApptTimeZone).Format("02.01.2006 15:04"),
				Service:   a.Service.Name,
				CanCancel: isAdmin || a.StartTime.Sub(now) > 72*time.Hour,
			})
		}
	}

	// Populate Recent Visits (only confirmed, non-block)
	var confirmedRecents []domain.Appointment
	for _, a := range history {
		if a.Status != "cancelled" && !strings.Contains(strings.ToLower(a.Service.Name), "block") && !strings.Contains(strings.ToLower(a.CustomerName), "admin block") {
			confirmedRecents = append(confirmedRecents, a)
		}
	}
	// Sort by date descending
	sort.Slice(confirmedRecents, func(i, j int) bool {
		return confirmedRecents[i].StartTime.After(confirmedRecents[j].StartTime)
	})

	// Take last 5
	limit := 5
	if len(confirmedRecents) < limit {
		limit = len(confirmedRecents)
	}
	for i := 0; i < limit; i++ {
		data.RecentVisits = append(data.RecentVisits, visitInfo{
			Date:    confirmedRecents[i].StartTime.Format("02.01.2006"),
			Service: confirmedRecents[i].Service.Name,
		})
	}

	// Grouping Logic - FROM DB NOW
	groups := make(map[string]*docGroup)
	initGroup := func(name string) {
		groups[name] = &docGroup{Name: name, Count: 0}
	}
	initGroup("Снимки")            // Scans
	initGroup("Фотографии")        // Photos
	initGroup("Видео")             // Videos
	initGroup("Голосовые заметки") // Voice Messages
	initGroup("Тексты")            // Texts
	initGroup("Прочее")            // Others

	for _, m := range mediaList {
		var targetGroup *docGroup
		switch m.FileType {
		case "scan":
			targetGroup = groups["Снимки"]
		case "photo", "image":
			targetGroup = groups["Фотографии"]
		case "voice", "audio":
			targetGroup = groups["Голосовые заметки"]
		case "video":
			targetGroup = groups["Видео"]
		case "document", "text":
			targetGroup = groups["Тексты"]
		default:
			targetGroup = groups["Прочее"]
		}

		if targetGroup != nil {
			targetGroup.Count++
			targetGroup.Files = append(targetGroup.Files, m)
			modTime := m.CreatedAt.Format("02.01.2006 15:04")
			if targetGroup.Latest == "" || m.CreatedAt.After(parseTime(targetGroup.Latest)) {
				targetGroup.Latest = modTime
			}
		}
	}

	// Add only populated groups
	order := []string{"Снимки", "Фотографии", "Видео", "Голосовые заметки", "Тексты", "Прочее"}
	for _, name := range order {
		if g := groups[name]; g != nil && g.Count > 0 {
			data.DocGroups = append(data.DocGroups, *g)
		}
	}

	var buf bytes.Buffer
	tmpl, errTmpl := template.New("medicalRecord").Parse(medicalRecordTemplate)
	if errTmpl != nil {
		logging.Errorf(": Failed to parse medical record template: %v", errTmpl)
		return "Error generating record."
	}
	errTmpl = tmpl.Execute(&buf, data)
	if errTmpl != nil {
		logging.Errorf(": Failed to execute medical record template: %v", errTmpl)
		return "Error generating record."
	}
	return buf.String()
}

func parseTime(s string) time.Time {
	t, _ := time.Parse("02.01.2006 15:04", s)
	return t
}

func (r *PostgresRepository) GenerateAdminSearchPage() string {
	if r.BotUsername == "" {
		logging.Warn("WARNING: BotUsername is empty in PostgresRepository! TWA links will be broken.")
	} else {
		logging.Debugf("GenerateAdminSearchPage using BotUsername: %s", r.BotUsername)
	}
	return strings.ReplaceAll(adminSearchTemplate, "{{.BotUsername}}", r.BotUsername)
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
