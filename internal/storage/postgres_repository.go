package storage

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/kfilin/massage-bot/internal/domain"
	"github.com/kfilin/massage-bot/internal/monitoring"
	"github.com/kfilin/massage-bot/internal/ports"
)

var _ ports.Repository = (*PostgresRepository)(nil)

type PostgresRepository struct {
	db         *sqlx.DB
	dataDir    string
	BotVersion string
}

func NewPostgresRepository(db *sqlx.DB, dataDir string) *PostgresRepository {
	if dataDir == "" {
		dataDir = "data"
	}
	patientsDir := filepath.Join(dataDir, "patients")
	if err := os.MkdirAll(patientsDir, 0755); err != nil {
		log.Printf("Warning: failed to create patients directory: %v", err)
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
	_, err := r.db.NamedExec(query, p)
	if err != nil {
		monitoring.DbErrorsTotal.WithLabelValues("save_patient").Inc()
		return err
	}

	// Update clinical note length metric
	noteLen := len(p.TherapistNotes)
	monitoring.ClinicalNoteLength.Set(float64(noteLen))

	// Mirror to Markdown file
	return r.saveToMarkdown(p)
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

func (r *PostgresRepository) saveToMarkdown(p domain.Patient) error {
	patientDir := r.getPatientDir(p)

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
(Используйте этот раздел для ежедневных записей)`, body)
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

	// Sync from Markdown if file exists (picks up edits)
	updated, errFile := r.syncFromFile(&p)
	if errFile == nil && updated {
		// Save back to DB to keep analytics and TWA fast
		r.db.NamedExec(`UPDATE patients SET name = :name, therapist_notes = :therapist_notes WHERE telegram_id = :telegram_id`, p)
	}

	return p, nil
}

func (r *PostgresRepository) syncFromFile(p *domain.Patient) (bool, error) {
	patientDir := r.getPatientDir(*p)
	filePath := filepath.Join(patientDir, fmt.Sprintf("%s.md", p.TelegramID))
	info, err := os.Stat(filePath)
	if err != nil {
		return false, err
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

	// 2. Extract full notes body
	bodyMarker := "# 🩺 Медицинская карта"
	statsMarker := "---"

	var notes string
	headerIdx := strings.Index(strContent, bodyMarker)
	if headerIdx != -1 {
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

func (r *PostgresRepository) LogEvent(patientID string, eventType string, details map[string]interface{}) error {
	detailsJSON, _ := json.Marshal(details)
	_, err := r.db.Exec("INSERT INTO analytics_events (patient_id, event_type, details) VALUES ($1, $2, $3)", patientID, eventType, detailsJSON)
	return err
}

func (r *PostgresRepository) mdToHTML(md string) template.HTML {
	if md == "" {
		return template.HTML("")
	}
	// 1. Basic escaping and line breaks
	h := template.HTMLEscapeString(md)

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

	// Line breaks (convert remaining \n to <br>)
	h = strings.ReplaceAll(h, "\n", "<br>")

	return template.HTML(h)
}

func (r *PostgresRepository) GenerateHTMLRecord(p domain.Patient, history []domain.Appointment) string {
	type docGroup struct {
		Name   string
		Count  int
		Latest string
	}
	type visitInfo struct {
		Date    string
		Service string
	}
	type templateData struct {
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
		DocGroups          []docGroup
		RecentVisits       []visitInfo
	}

	getCalLink := func(t time.Time, service string) string {
		start := t.Format("20060102T150405")
		end := t.Add(time.Hour).Format("20060102T150405")
		title := "Massage: " + service
		return fmt.Sprintf("https://www.google.com/calendar/render?action=TEMPLATE&text=%s&dates=%s/%s", strings.ReplaceAll(title, " ", "+"), start, end)
	}

	// We'll keep emojis for now as they are part of modern UI
	// re := regexp.MustCompile(`[\x{1F300}-\x{1FAD6}]|[\x{2600}-\x{26FF}]|[\x{2700}-\x{27BF}]|[\x{1F600}-\x{1F64F}]|[\x{1F680}-\x{1F6FF}]|[\x{1F1E6}-\x{1F1FF}]`)
	// cleanNotes := re.ReplaceAllString(p.TherapistNotes, "")
	cleanNotes := p.TherapistNotes

	data := templateData{
		Name:               strings.ToUpper(p.Name),
		TelegramID:         p.TelegramID,
		TotalVisits:        p.TotalVisits,
		GeneratedAt:        time.Now().Format("02.01.2006 15:04"),
		CurrentService:     p.CurrentService,
		BotVersion:         r.BotVersion,
		TherapistNotes:     r.mdToHTML(cleanNotes),
		VoiceTranscripts:   template.HTML(strings.ReplaceAll(template.HTMLEscapeString(p.VoiceTranscripts), "\n", "<br>")),
		FirstVisit:         p.FirstVisit.Format("02.01.2006 15:04"),
		LastVisit:          p.LastVisit.Format("02.01.2006 15:04"),
		FirstVisitLink:     getCalLink(p.FirstVisit, p.CurrentService),
		LastVisitLink:      getCalLink(p.LastVisit, p.CurrentService),
		ShowFirstVisitLink: p.FirstVisit.After(time.Now()),
		ShowLastVisitLink:  p.LastVisit.After(time.Now()),
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

	// Grouping Logic
	groups := make(map[string]*docGroup)
	initGroup := func(name string) {
		groups[name] = &docGroup{Name: name, Count: 0}
	}
	initGroup("Scans")
	initGroup("Photos")
	initGroup("Videos")
	initGroup("Voice Messages")
	initGroup("Texts")
	initGroup("Others")

	patientDir := r.getPatientDir(p)
	filepath.Walk(patientDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		if info.Name() == fmt.Sprintf("%s.md", p.TelegramID) {
			return nil
		}

		relPath, _ := filepath.Rel(patientDir, path)
		folder := ""
		if parts := strings.Split(relPath, string(os.PathSeparator)); len(parts) > 0 {
			folder = parts[0]
		}

		var targetGroup *docGroup
		ext := strings.ToLower(filepath.Ext(path))

		switch folder {
		case "scans":
			targetGroup = groups["Scans"]
		case "images":
			targetGroup = groups["Photos"]
		case "messages":
			targetGroup = groups["Voice Messages"]
		case "documents":
			if ext == ".pdf" || ext == ".doc" || ext == ".docx" || ext == ".txt" {
				targetGroup = groups["Texts"]
			} else {
				targetGroup = groups["Others"]
			}
		default:
			// Fallback by extension if outside standard folders
			if ext == ".jpg" || ext == ".png" || ext == ".jpeg" {
				targetGroup = groups["Photos"]
			} else if ext == ".mp4" || ext == ".mov" || ext == ".avi" {
				targetGroup = groups["Videos"]
			} else if ext == ".pdf" || ext == ".doc" || ext == ".docx" {
				targetGroup = groups["Texts"]
			} else {
				targetGroup = groups["Others"]
			}
		}

		if targetGroup != nil {
			targetGroup.Count++
			modTime := info.ModTime().Format("02.01.2006 15:04")
			if targetGroup.Latest == "" || info.ModTime().After(parseTime(targetGroup.Latest)) {
				targetGroup.Latest = modTime
			}
		}
		return nil
	})

	// Add only populated groups
	order := []string{"Scans", "Photos", "Videos", "Voice Messages", "Texts", "Others"}
	for _, name := range order {
		if g := groups[name]; g != nil && g.Count > 0 {
			data.DocGroups = append(data.DocGroups, *g)
		}
	}

	var buf bytes.Buffer
	tmpl, errTmpl := template.New("medicalRecord").Parse(medicalRecordTemplate)
	if errTmpl != nil {
		log.Printf("ERROR: Failed to parse medical record template: %v", errTmpl)
		return "Error generating record."
	}
	errTmpl = tmpl.Execute(&buf, data)
	if errTmpl != nil {
		log.Printf("ERROR: Failed to execute medical record template: %v", errTmpl)
		return "Error generating record."
	}
	return buf.String()
}

func parseTime(s string) time.Time {
	t, _ := time.Parse("02.01.2006 15:04", s)
	return t
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
	filepath.Walk(patientDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
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
		infos = append(infos, fileInfo{name: info.Name(), relPath: relPath, modTime: info.ModTime(), category: category})
		return nil
	})

	if len(infos) == 0 {
		return "Документов пока нет."
	}
	sort.Slice(infos, func(i, j int) bool { return infos[i].modTime.After(infos[j].modTime) })
	var list string
	for _, info := range infos {
		list += fmt.Sprintf("- [%s] %s [[%s|%s]]\n", info.modTime.Format("02.01.2006"), info.category, info.relPath, info.name)
	}
	return list
}

func (r *PostgresRepository) SavePatientDocumentReader(telegramID string, filename string, category string, reader io.Reader) (string, error) {
	p, err := r.GetPatient(telegramID)
	if err != nil {
		return "", err
	}
	patientDir := r.getPatientDir(p)
	var targetDir string
	switch strings.ToLower(category) {
	case "scans":
		targetDir = filepath.Join(patientDir, "scans", time.Now().Format("02.01.06"))
	case "images":
		targetDir = filepath.Join(patientDir, "images")
	case "messages":
		targetDir = filepath.Join(patientDir, "messages")
	default:
		targetDir = filepath.Join(patientDir, "documents")
	}
	os.MkdirAll(targetDir, 0755)
	filePath := filepath.Join(targetDir, filename)
	f, _ := os.Create(filePath)
	defer f.Close()
	io.Copy(f, reader)
	return filePath, nil
}

func (r *PostgresRepository) MigrateFolderNames() error {
	var patients []domain.Patient
	r.db.Select(&patients, "SELECT * FROM patients")
	for _, p := range patients {
		oldDir := filepath.Join(r.dataDir, "patients", p.TelegramID)
		newDir := r.getPatientDir(p)
		if _, err := os.Stat(oldDir); err == nil && oldDir != newDir {
			os.Rename(oldDir, newDir)
		}
	}
	return nil
}

func (r *PostgresRepository) SyncAll() error {
	patientsDir := filepath.Join(r.dataDir, "patients")
	entries, _ := os.ReadDir(patientsDir)
	for _, e := range entries {
		if e.IsDir() {
			id := e.Name()
			if strings.Contains(id, "(") {
				id = id[strings.LastIndex(id, "(")+1 : len(id)-1]
			}
			r.GetPatient(id)
		}
	}
	var patients []domain.Patient
	r.db.Select(&patients, "SELECT * FROM patients")
	for _, p := range patients {
		// --- PERMANENT SCRUB of legacy boilerplate (Clinical Edition Cleanup) ---
		if strings.Contains(p.TherapistNotes, "## 📁 Ссылки на документы") {
			parts := strings.Split(p.TherapistNotes, "## 📁 Ссылки на документы")
			p.TherapistNotes = strings.TrimSpace(parts[0])
			r.SavePatient(p) // Persists cleaned notes to DB and mirrors to Markdown
			log.Printf("[Cleanup] Permanently scrubbed legacy section for patient %s", p.TelegramID)
		}

		filePath := filepath.Join(r.getPatientDir(p), fmt.Sprintf("%s.md", p.TelegramID))
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			r.saveToMarkdown(p)
		}
	}
	return nil
}

func (r *PostgresRepository) CreateBackup() (string, error) { return "", nil } // Simplified for now
