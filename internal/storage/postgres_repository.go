package storage

import (
	"encoding/json"
	"fmt"
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
	docs := r.listDocuments(p.TelegramID)
	// Replace obsidian links with simple text for PDF
	docs = strings.ReplaceAll(docs, "[[documents/", "")
	docs = strings.ReplaceAll(docs, "|", " - ")
	docs = strings.ReplaceAll(docs, "]]", "")

	// Handle separation of transcripts if new field is empty but they are in notes
	notes := p.TherapistNotes
	transcripts := p.VoiceTranscripts

	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body { font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif; line-height: 1.8; color: #2c3e50; max-width: 900px; margin: 0 auto; padding: 50px; background-color: #fff; }
        .document-wrapper { border: 1px solid #e1e8ed; padding: 40px; border-radius: 4px; box-shadow: 0 4px 6px rgba(0,0,0,0.05); }
        .header { text-align: left; border-bottom: 3px solid #3498db; padding-bottom: 20px; margin-bottom: 40px; }
        .header h1 { color: #2980b9; margin: 0; font-size: 2.2em; text-transform: uppercase; font-weight: 300; }
        .header-meta { color: #7f8c8d; font-size: 0.9em; margin-top: 5px; }
        
        .info-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 30px; margin-bottom: 40px; background: #f8f9fa; padding: 25px; border-left: 5px solid #3498db; }
        .info-item { display: flex; flex-direction: column; }
        .info-label { font-size: 0.75em; font-weight: bold; color: #7f8c8d; text-transform: uppercase; letter-spacing: 1px; }
        .info-value { font-size: 1.1em; color: #2c3e50; border-bottom: 1px solid #eee; padding-bottom: 2px; }

        .section { margin-bottom: 40px; }
        .section-title { font-size: 1.3em; color: #2980b9; border-bottom: 1px solid #3498db; padding-bottom: 10px; margin-bottom: 20px; font-weight: 600; }
        
        .clinical-notes { background: #fff; border: 1px solid #e1e8ed; padding: 20px; border-radius: 4px; white-space: pre-wrap; font-size: 1.05em; color: #34495e; }
        .transcript-content { background: #fdfdfd; border-left: 4px solid #95a5a6; padding: 15px; font-style: italic; font-size: 0.95em; color: #5d6d7e; white-space: pre-wrap; margin-top: 10px; }

        .docs-list { list-style: none; padding: 0; }
        .docs-list li { padding: 10px 15px; margin-bottom: 8px; background: #fbfcfc; border: 1px solid #ebedef; border-radius: 4px; font-size: 0.9em; color: #5d6d7e; display: flex; align-items: center; }
        .docs-list li::before { content: "📄"; margin-right: 10px; }

        .footer { margin-top: 60px; text-align: center; font-size: 0.8em; color: #bdc3c7; border-top: 1px solid #f0f3f4; padding-top: 30px; }
        .footer-tag { background: #3498db; color: white; padding: 2px 8px; border-radius: 10px; font-weight: bold; font-size: 0.8em; }
    </style>
</head>
<body>
    <div class="document-wrapper">
        <div class="header">
            <h1>Медицинская Карта</h1>
            <div class="header-meta">Клиническая запись • Конфиденциально</div>
        </div>

        <div class="info-grid">
            <div class="info-item">
                <span class="info-label">Пациент</span>
                <span class="info-value">%s</span>
            </div>
            <div class="info-item">
                <span class="info-label">ID Системы</span>
                <span class="info-value">%s</span>
            </div>
            <div class="info-item">
                <span class="info-label">Первый визит</span>
                <span class="info-value">%s</span>
            </div>
            <div class="info-item">
                <span class="info-label">Всего посещений</span>
                <span class="info-value">%d</span>
            </div>
            <div class="info-item">
                <span class="info-label">Последний визит</span>
                <span class="info-value">%s</span>
            </div>
            <div class="info-item">
                <span class="info-label">Текущая услуга</span>
                <span class="info-value">%s</span>
            </div>
        </div>

        <div class="section">
            <div class="section-title">🩺 Клинические Заметки</div>
            <div class="clinical-notes">%s</div>
        </div>

        %s

        <div class="section">
            <div class="section-title">📂 Прикрепленные файлы</div>
            <ul class="docs-list">
                %s
            </ul>
        </div>

        <div class="footer">
            <p>Сгенерировано автоматически <span class="footer-tag">Vera Bot</span> • %s</p>
            <p>Данный документ является электронной копией медицинской записи.</p>
        </div>
    </div>
</body>
</html>`,
		p.Name, p.TelegramID,
		p.FirstVisit.Format("02.01.2006"),
		p.TotalVisits,
		p.LastVisit.Format("02.01.2006"),
		p.CurrentService,
		notes,
		formatTranscriptsSection(transcripts),
		r.formatDocsForHTML(docs),
		time.Now().Format("02.01.2006 15:04"))
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

func (r *PostgresRepository) formatDocsForHTML(docs string) string {
	if docs == "Документов пока нет." {
		return "<li>Документов пока нет.</li>"
	}
	lines := strings.Split(strings.TrimSpace(docs), "\n")
	var htmlList string
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			htmlList += fmt.Sprintf("<li>%s</li>", strings.TrimPrefix(line, "- "))
		}
	}
	return htmlList
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
