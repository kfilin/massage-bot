package storage

import (
	"encoding/json"
	"fmt"
	"html/template"
	"regexp"
	"time"

	"archive/zip"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/kfilin/massage-bot/internal/domain"
	"github.com/kfilin/massage-bot/internal/ports"
	_ "github.com/lib/pq"
)

var _ ports.Repository = (*PostgresRepository)(nil)

type PostgresRepository struct {
	db      *sqlx.DB
	dataDir string
}

func NewPostgresRepository(db *sqlx.DB, dataDir string) *PostgresRepository {
	if dataDir == "" {
		dataDir = "data"
	}
	return &PostgresRepository{db: db, dataDir: dataDir}
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
	_, err := r.db.NamedExec(query, p)
	return err
}

func (r *PostgresRepository) GetPatient(telegramID string) (domain.Patient, error) {
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
	detailsJSON, err := json.Marshal(details)
	if err != nil {
		return err
	}

	_, err = r.db.Exec(
		"INSERT INTO analytics_events (patient_id, event_type, details) VALUES ($1, $2, $3)",
		patientID, eventType, detailsJSON,
	)
	return err
}

func (r *PostgresRepository) GenerateHTMLRecord(p domain.Patient) string {
	type docItem struct {
		Name    string
		IsVoice bool
	}
	type templateData struct {
		Name             string
		TelegramID       string
		TotalVisits      int
		GeneratedAt      string
		CurrentService   string
		TherapistNotes   string
		VoiceTranscripts template.HTML
		FirstVisit       string
		LastVisit        string
		Documents        []docItem
	}

	// Strip ALL emojis and special symbols for a clean clinical look
	re := regexp.MustCompile(`[\x{1F300}-\x{1FAD6}]|[\x{2600}-\x{26FF}]|[\x{2700}-\x{27BF}]|[\x{1F600}-\x{1F64F}]|[\x{1F680}-\x{1F6FF}]|[\x{1F1E6}-\x{1F1FF}]`)
	cleanNotes := re.ReplaceAllString(p.TherapistNotes, "")
	cleanTranscripts := re.ReplaceAllString(p.VoiceTranscripts, "")

	data := templateData{
		Name:             strings.ToUpper(p.Name),
		TelegramID:       p.TelegramID,
		TotalVisits:      p.TotalVisits,
		GeneratedAt:      time.Now().Format("02.01.2006 15:04"),
		CurrentService:   p.CurrentService,
		TherapistNotes:   cleanNotes,
		VoiceTranscripts: template.HTML(strings.ReplaceAll(cleanTranscripts, "\n", "<br>")),
		FirstVisit:       p.FirstVisit.Format("02.01.2006 15:04"),
		LastVisit:        p.LastVisit.Format("02.01.2006 15:04"),
	}

	// Parse documents
	docList := r.listDocuments(p.TelegramID)
	if docList != "Документов пока нет." {
		lines := strings.Split(strings.TrimSpace(docList), "\n")
		for _, line := range lines {
			if strings.TrimSpace(line) == "" {
				continue
			}
			cleanLine := strings.TrimPrefix(line, "- ")
			// Extract name from Obsidian format [date] [[path|name]]
			name := cleanLine
			if strings.Contains(cleanLine, "|") {
				name = cleanLine[strings.Index(cleanLine, "|")+1 : strings.Index(cleanLine, "]]")]
			}

			isVoice := strings.Contains(strings.ToLower(name), ".ogg") || strings.Contains(strings.ToLower(name), ".wav")
			data.Documents = append(data.Documents, docItem{Name: name, IsVoice: isVoice})
		}
	}

	tmpl, err := template.New("medical_record").Parse(medicalRecordTemplate)
	if err != nil {
		return fmt.Sprintf("Error parsing template: %v", err)
	}

	var buf strings.Builder
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Sprintf("Error executing template: %v", err)
	}

	return buf.String()
}

func (r *PostgresRepository) listDocuments(telegramID string) string {
	docDir := filepath.Join(r.dataDir, "patients", telegramID, "documents")
	entries, err := os.ReadDir(docDir)
	if err != nil || len(entries) == 0 {
		return "Документов пока нет."
	}

	type fileInfo struct {
		name    string
		modTime time.Time
	}
	var infos []fileInfo
	for _, e := range entries {
		if !e.IsDir() {
			fi, err := e.Info()
			if err == nil {
				infos = append(infos, fileInfo{name: e.Name(), modTime: fi.ModTime()})
			}
		}
	}

	sort.Slice(infos, func(i, j int) bool {
		return infos[i].modTime.After(infos[j].modTime)
	})

	var list string
	for _, info := range infos {
		list += fmt.Sprintf("- [%s] [[documents/%s|%s]]\n", info.modTime.Format("02.01.2006 15:04"), info.name, info.name)
	}
	return list
}

func (r *PostgresRepository) SavePatientPDF(telegramID string, pdfBytes []byte) error {
	patientDir := filepath.Join(r.dataDir, "patients", telegramID)
	if err := os.MkdirAll(patientDir, 0755); err != nil {
		return fmt.Errorf("failed to create patient directory: %w", err)
	}

	pdfPath := filepath.Join(patientDir, "medical_record.pdf")
	return os.WriteFile(pdfPath, pdfBytes, 0644)
}

func (r *PostgresRepository) SavePatientDocumentReader(telegramID string, filename string, reader io.Reader) (string, error) {
	docDir := filepath.Join(r.dataDir, "patients", telegramID, "documents")
	if err := os.MkdirAll(docDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create documents directory: %w", err)
	}

	filePath := filepath.Join(docDir, filename)
	f, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to create document file: %w", err)
	}
	defer f.Close()

	if _, err := io.Copy(f, reader); err != nil {
		return "", fmt.Errorf("failed to save document data: %w", err)
	}

	return filePath, nil
}

func (r *PostgresRepository) CreateBackup() (string, error) {
	backupDir := filepath.Join(r.dataDir, "backups")
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create backup directory: %w", err)
	}

	timestamp := time.Now().Format("20060102_150405")
	backupPath := filepath.Join(backupDir, fmt.Sprintf("backup_%s.zip", timestamp))

	newZipFile, err := os.Create(backupPath)
	if err != nil {
		return "", err
	}
	defer newZipFile.Close()

	zipWriter := zip.NewWriter(newZipFile)
	defer zipWriter.Close()

	patientsPath := filepath.Join(r.dataDir, "patients")
	err = filepath.Walk(patientsPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(r.dataDir, path)
		if err != nil {
			return err
		}

		zipFile, err := zipWriter.Create(relPath)
		if err != nil {
			return err
		}

		fsFile, err := os.Open(path)
		if err != nil {
			return err
		}
		defer fsFile.Close()

		_, err = io.Copy(zipFile, fsFile)
		return err
	})

	return backupPath, err
}
