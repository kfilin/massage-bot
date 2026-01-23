package storage

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"regexp"
	"time"

	"archive/zip"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"

	"github.com/jmoiron/sqlx"
	"github.com/kfilin/massage-bot/internal/domain"
	"github.com/kfilin/massage-bot/internal/ports"
	_ "github.com/lib/pq"
	"github.com/yuin/goldmark"
)

var _ ports.Repository = (*PostgresRepository)(nil)

type PostgresRepository struct {
	db          *sqlx.DB
	dataDir     string
	patientsDir string
	BotVersion  string
}

func NewPostgresRepository(db *sqlx.DB, dataDir string) *PostgresRepository {
	if dataDir == "" {
		dataDir = "data"
	}
	patientsDir := filepath.Join(dataDir, "patients")
	if err := os.MkdirAll(patientsDir, 0755); err != nil {
		fmt.Printf("Warning: failed to create patients directory: %v\n", err)
	}
	return &PostgresRepository{
		db:          db,
		dataDir:     dataDir,
		patientsDir: patientsDir,
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
	_, err := r.db.NamedExec(query, p)
	if err != nil {
		return err
	}

	// Mirror to Markdown file
	return r.saveToMarkdown(p)
}

func (r *PostgresRepository) getPatientDir(p domain.Patient) string {
	// 1. Scan for any folder ending with (ID) - allows for manual renames in Obsidian
	entries, err := os.ReadDir(r.patientsDir)
	if err == nil {
		suffix := fmt.Sprintf("(%s)", p.TelegramID)
		for _, e := range entries {
			if e.IsDir() && strings.HasSuffix(e.Name(), suffix) {
				return filepath.Join(r.patientsDir, e.Name())
			}
		}
	}

	// 2. Default fallback if no existing folder found
	reg := regexp.MustCompile(`[<>:"/\\|?*]`)
	cleanName := reg.ReplaceAllString(p.Name, "_")
	folderName := fmt.Sprintf("%s (%s)", cleanName, p.TelegramID)
	return filepath.Join(r.patientsDir, folderName)
}

func (r *PostgresRepository) saveToMarkdown(p domain.Patient) error {
	patientDir := r.getPatientDir(p)

	// Ensure old ID-only folder is migrated if it exists
	oldDir := filepath.Join(r.patientsDir, p.TelegramID)
	if _, err := os.Stat(oldDir); err == nil {
		os.Rename(oldDir, patientDir)
	}

	if err := os.MkdirAll(patientDir, 0755); err != nil {
		return fmt.Errorf("failed to create patient directory: %w", err)
	}

	filePath := filepath.Join(patientDir, fmt.Sprintf("%s.md", p.TelegramID))

	// Prevent duplication: if the notes already have the template markers,
	// we use them as is, otherwise we wrap with the full template structure.
	body := p.TherapistNotes
	if !strings.Contains(body, "## 📋 История болезни") {
		body = fmt.Sprintf(`## 📋 История болезни
%s

## 📝 Заметки терапевта
(Используйте этот раздел для ежедневных записей)

## 📁 Ссылки на документы
(Документы загружаются через бота и доступны в TWA)`, body)
	}

	// Create content from template
	content := fmt.Sprintf(`---
Name: %s
ID: %s
---

# 🩺 Медицинская карта: %s

%s

---
*Статистика (обновляется ботом):*
- Первый визит: %s
- Последний визит: %s
- Всего визитов: %d
- Услуга: %s
`, p.Name, p.TelegramID, p.Name, body,
		p.FirstVisit.Format("02.01.2006"),
		p.LastVisit.Format("02.01.2006"),
		p.TotalVisits, p.CurrentService)

	return os.WriteFile(filePath, []byte(content), 0644)
}

func (r *PostgresRepository) GetPatient(telegramID string) (domain.Patient, error) {
	var p domain.Patient
	err := r.db.Get(&p, "SELECT * FROM patients WHERE telegram_id = $1", telegramID)

	// If not found in DB, try to find in Markdown folder
	if err != nil {
		p.TelegramID = telegramID
		updated, errFile := r.syncFromFile(&p)
		if errFile == nil && updated {
			log.Printf("[Sync] Discovered patient %s from Markdown file after DB miss", telegramID)
			// Save to DB to establish record
			r.SavePatient(p)
			return p, nil
		}
		return p, err // Return original DB error if file also not found
	}

	// Sync from Markdown if file exists (picks up Vera's edits)
	updated, errFile := r.syncFromFile(&p)
	if errFile == nil && updated {
		// Save back to DB to keep analytics and TWA fast
		r.db.NamedExec(`UPDATE patients SET name = :name, therapist_notes = :therapist_notes WHERE telegram_id = :telegram_id`, p)
	}

	return p, nil
}

func (r *PostgresRepository) syncFromFile(p *domain.Patient) (bool, error) {
	// Try the descriptive folder first: data/patients/Name (ID)/ID.md
	filePath := filepath.Join(r.getPatientDir(*p), fmt.Sprintf("%s.md", p.TelegramID))
	info, err := os.Stat(filePath)
	if err != nil {
		// Try ID-only folder: data/patients/ID/ID.md
		filePath = filepath.Join(r.patientsDir, p.TelegramID, fmt.Sprintf("%s.md", p.TelegramID))
		info, err = os.Stat(filePath)
		if err != nil {
			// Fallback to legacy structure: data/patients/ID.md
			legacyPath := filepath.Join(r.patientsDir, fmt.Sprintf("%s.md", p.TelegramID))
			legacyInfo, errLegacy := os.Stat(legacyPath)
			if errLegacy != nil {
				return false, errLegacy // Neither exist
			}
			filePath = legacyPath
			info = legacyInfo
		}
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		return false, err
	}

	strContent := string(content)

	// 1. Extract name from frontmatter if possible
	name := p.Name
	if strings.Contains(strContent, "Name: ") {
		lines := strings.Split(strContent, "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "Name: ") {
				name = strings.TrimSpace(strings.TrimPrefix(line, "Name: "))
				break
			}
		}
	}

	// 2. Extract full notes body (after the title header and before the stats footer)
	bodyMarker := "# 🩺 Медицинская карта"
	statsMarker := "---" // Stats starts after the horizontal line

	var notes string
	headerIdx := strings.Index(strContent, bodyMarker)
	if headerIdx != -1 {
		// Find end of the title line
		lineEnd := strings.Index(strContent[headerIdx:], "\n")
		if lineEnd != -1 {
			bodyStart := headerIdx + lineEnd
			bodyEnd := strings.LastIndex(strContent, statsMarker)

			if bodyEnd > bodyStart {
				notes = strings.TrimSpace(strContent[bodyStart:bodyEnd])
			} else {
				notes = strings.TrimSpace(strContent[bodyStart:])
			}
		}
	}

	if notes == "" {
		// Fallback to old marker if title match failed
		noteMarker := "## 📋 История болезни"
		idx := strings.Index(strContent, noteMarker)
		if idx != -1 {
			notes = strings.TrimSpace(strContent[idx+len(noteMarker):])
			if nextIdx := strings.Index(notes, "---"); nextIdx != -1 {
				notes = strings.TrimSpace(notes[:nextIdx])
			}
		}
	}

	// Return true if something changed
	hasChanged := p.TherapistNotes != notes || p.Name != name
	if hasChanged {
		p.TherapistNotes = notes
		p.Name = name
		log.Printf("[Sync] Updated patient %s from Markdown file (Last Mod: %v)", p.TelegramID, info.ModTime())
		return true, nil
	}

	return false, nil
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

func (r *PostgresRepository) MigrateFolderNames() error {
	var patients []domain.Patient
	err := r.db.Select(&patients, "SELECT * FROM patients")
	if err != nil {
		return err
	}

	log.Printf("[Migration] Starting folder name migration for %d patients...", len(patients))
	for _, p := range patients {
		oldDir := filepath.Join(r.patientsDir, p.TelegramID)
		newDir := r.getPatientDir(p)

		if _, err := os.Stat(oldDir); err == nil {
			if oldDir == newDir {
				continue
			}
			err := os.Rename(oldDir, newDir)
			if err != nil {
				log.Printf("[Migration] Error renaming %s to %s: %v", oldDir, newDir, err)
			} else {
				log.Printf("[Migration] Migrated %s -> %s", p.TelegramID, filepath.Base(newDir))
			}
		}
	}
	return nil
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
	generateHMAC := func(id string, secret string) string {
		h := hmac.New(sha256.New, []byte(secret))
		h.Write([]byte(id))
		return hex.EncodeToString(h.Sum(nil))
	}

	type docItem struct {
		Name    string
		IsVoice bool
	}
	type templateData struct {
		ID                 string
		Token              string
		Name               string
		TelegramID         string
		TotalVisits        int
		GeneratedAt        string
		CurrentService     string
		BotVersion         string
		TherapistNotes     template.HTML
		VoiceTranscripts   template.HTML
		FirstVisit         string
		LastVisit          string
		FirstVisitLink     string
		LastVisitLink      string
		ShowFirstVisitLink bool
		ShowLastVisitLink  bool
		Documents          []docItem
		AppURL             string
	}

	// Helper to generate Google Calendar Link
	getCalLink := func(t time.Time, service string) string {
		start := t.Format("20060102T150405")
		end := t.Add(time.Hour).Format("20060102T150405")
		title := "Massage: " + service
		details := "Scheduled via Vera Massage Bot"
		return fmt.Sprintf(
			"https://www.google.com/calendar/render?action=TEMPLATE&text=%s&dates=%s/%s&details=%s",
			strings.ReplaceAll(title, " ", "+"),
			start, end,
			strings.ReplaceAll(details, " ", "+"),
		)
	}

	// Strip ALL emojis and special symbols for a clean clinical look
	re := regexp.MustCompile(`[\x{1F300}-\x{1FAD6}]|[\x{2600}-\x{26FF}]|[\x{2700}-\x{27BF}]|[\x{1F600}-\x{1F64F}]|[\x{1F680}-\x{1F6FF}]|[\x{1F1E6}-\x{1F1FF}]`)
	cleanNotes := re.ReplaceAllString(p.TherapistNotes, "")
	cleanTranscripts := re.ReplaceAllString(p.VoiceTranscripts, "")

	// Use the application's timezone for consistent display
	loc := domain.ApptTimeZone
	if loc == nil {
		loc = time.Local
	}

	// Convert Markdown to HTML
	var bufNotes bytes.Buffer
	if err := goldmark.Convert([]byte(cleanNotes), &bufNotes); err != nil {
		log.Printf("Markdown conversion error: %v", err)
		bufNotes.WriteString(cleanNotes) // Fallback to raw text
	}

	data := templateData{
		ID:                 p.TelegramID,
		Token:              generateHMAC(p.TelegramID, os.Getenv("WEBAPP_SECRET")),
		Name:               strings.ToUpper(p.Name),
		TelegramID:         p.TelegramID,
		TotalVisits:        p.TotalVisits,
		GeneratedAt:        time.Now().In(loc).Format("02.01.2006 15:04"),
		CurrentService:     p.CurrentService,
		BotVersion:         r.BotVersion,
		TherapistNotes:     template.HTML(bufNotes.String()),
		VoiceTranscripts:   template.HTML(strings.ReplaceAll(cleanTranscripts, "\n", "<br>")),
		FirstVisit:         p.FirstVisit.In(loc).Format("02.01.2006 15:04"),
		LastVisit:          p.LastVisit.In(loc).Format("02.01.2006 15:04"),
		FirstVisitLink:     getCalLink(p.FirstVisit, p.CurrentService),
		LastVisitLink:      getCalLink(p.LastVisit, p.CurrentService),
		ShowFirstVisitLink: p.FirstVisit.After(time.Now()),
		ShowLastVisitLink:  p.LastVisit.After(time.Now()),
		AppURL:             os.Getenv("WEBAPP_URL"),
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
	p, err := r.GetPatient(telegramID)
	if err != nil {
		return "Ошибка при загрузке данных пациента."
	}
	patientDir := r.getPatientDir(p)

	type fileInfo struct {
		name     string
		relPath  string
		modTime  time.Time
		category string
	}
	var infos []fileInfo

	// Walk recursively through patient folder
	filepath.Walk(patientDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}

		// Skip the main medical card .md file
		if info.Name() == fmt.Sprintf("%s.md", telegramID) {
			return nil
		}

		relPath, _ := filepath.Rel(patientDir, path)
		category := "docs"
		if strings.Contains(relPath, "images") {
			category = "🖼️"
		} else if strings.Contains(relPath, "messages") {
			category = "🎙️"
		} else if strings.Contains(relPath, "scans") {
			category = "🏥"
		}

		infos = append(infos, fileInfo{
			name:     info.Name(),
			relPath:  relPath,
			modTime:  info.ModTime(),
			category: category,
		})
		return nil
	})

	if len(infos) == 0 {
		return "Документов пока нет."
	}

	sort.Slice(infos, func(i, j int) bool {
		return infos[i].modTime.After(infos[j].modTime)
	})

	var list string
	for _, info := range infos {
		// Example format: - [22.01.2026] 🏥 scans/22.01.26/mri.pdf
		list += fmt.Sprintf("- [%s] %s [[%s|%s]]\n",
			info.modTime.Format("02.01.2006"),
			info.category,
			info.relPath,
			info.name)
	}
	return list
}

func (r *PostgresRepository) SavePatientDocumentReader(telegramID string, filename string, category string, reader io.Reader) (string, error) {
	p, err := r.GetPatient(telegramID)
	if err != nil {
		return "", fmt.Errorf("patient not found: %w", err)
	}
	// Base directory for the patient
	patientDir := r.getPatientDir(p)

	// Determine the specific subfolder based on the category
	var targetDir string
	switch strings.ToLower(category) {
	case "scans":
		// For scans, we use the requested date-based subfolder (e.g., 22.01.26)
		dateDir := time.Now().Format("02.01.06")
		targetDir = filepath.Join(patientDir, "scans", dateDir)
	case "images":
		targetDir = filepath.Join(patientDir, "images")
	case "messages":
		targetDir = filepath.Join(patientDir, "messages")
	default:
		// Fallback to a generic documents folder if category is unknown
		targetDir = filepath.Join(patientDir, "documents")
	}

	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create target directory %s: %w", targetDir, err)
	}

	filePath := filepath.Join(targetDir, filename)
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
func (r *PostgresRepository) SyncAll() error {
	// 1. Scan the patients directory for both new and legacy structures
	entries, err := os.ReadDir(r.patientsDir)
	if err != nil {
		return err
	}

	for _, e := range entries {
		name := e.Name()
		if e.IsDir() {
			// Extract ID from folder name if it's in "Name (ID)" format
			id := name
			if strings.Contains(name, "(") && strings.HasSuffix(name, ")") {
				start := strings.LastIndex(name, "(")
				id = name[start+1 : len(name)-1]
			}
			// Triggers file-to-db sync (and potential folder rename)
			p, err := r.GetPatient(id)
			if err != nil {
				continue
			}

			// 2. Migrate legacy document folders if they exist
			patientDir := r.getPatientDir(p)
			legacyDocDir := filepath.Join(patientDir, "documents")
			if _, err := os.Stat(legacyDocDir); err == nil {
				log.Printf("[Migration] Categorizing documents for %s...", p.Name)
				docEntries, _ := os.ReadDir(legacyDocDir)
				for _, de := range docEntries {
					if de.IsDir() {
						continue
					}
					oldPath := filepath.Join(legacyDocDir, de.Name())
					ext := strings.ToLower(filepath.Ext(de.Name()))

					category := "documents"
					switch ext {
					case ".pdf", ".doc", ".docx":
						category = "scans"
					case ".ogg", ".mp4", ".mov", ".mp3", ".wav":
						category = "messages"
					case ".jpg", ".jpeg", ".png", ".heic":
						category = "images"
					}

					info, _ := de.Info()
					destDir := filepath.Join(patientDir, category)
					if category == "scans" {
						dateStr := info.ModTime().Format("02.01.06")
						destDir = filepath.Join(destDir, dateStr)
					}

					if err := os.MkdirAll(destDir, 0755); err == nil {
						newPath := filepath.Join(destDir, de.Name())
						os.Rename(oldPath, newPath)
					}
				}
				os.Remove(legacyDocDir)
			}
		} else if strings.HasSuffix(name, ".md") {
			// Legacy file suspect: data/patients/{ID}.md
			actualID := strings.TrimSuffix(name, ".md")
			// Triggers sync AND saveToMarkdown, which will automatically migrate to new folder
			r.GetPatient(actualID)
		}
	}

	// 3. Sync from DB to Files (Generate missing files/folders for existing patients)
	var patients []domain.Patient
	errDB := r.db.Select(&patients, "SELECT * FROM patients")
	if errDB != nil {
		return errDB
	}

	log.Printf("[Sync] Checking %d database records for missing files...", len(patients))
	count := 0
	for _, p := range patients {
		patientDir := r.getPatientDir(p)
		filePath := filepath.Join(patientDir, fmt.Sprintf("%s.md", p.TelegramID))
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			if errSave := r.saveToMarkdown(p); errSave == nil {
				count++
			}
		}
	}
	log.Printf("[Sync] Bi-directional synchronization complete. Generated %d missing records.", count)

	return nil
}
