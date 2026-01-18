package storage

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/kfilin/massage-bot/internal/domain"
)

var DataDir string

func init() {
	DataDir = os.Getenv("DATA_DIR")
	if DataDir == "" {
		DataDir = "data"
	}
}

func SavePatient(patient domain.Patient) error {
	// Create patient directory
	patientDir := filepath.Join(DataDir, "patients", patient.TelegramID)
	if err := os.MkdirAll(patientDir, 0755); err != nil {
		return fmt.Errorf("failed to create patient directory: %w", err)
	}

	// 1. Save JSON (for bot internal use)
	jsonPath := filepath.Join(patientDir, "patient.json")
	jsonData, err := json.MarshalIndent(patient, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal patient data: %w", err)
	}

	if err := os.WriteFile(jsonPath, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write JSON file: %w", err)
	}

	return nil
}

// SavePatientPDF saves the generated medical record PDF to the patient's directory.
func SavePatientPDF(telegramID string, pdfBytes []byte) error {
	patientDir := filepath.Join(DataDir, "patients", telegramID)
	if err := os.MkdirAll(patientDir, 0755); err != nil {
		return fmt.Errorf("failed to create patient directory: %w", err)
	}

	pdfPath := filepath.Join(patientDir, "medical_record.pdf")
	if err := os.WriteFile(pdfPath, pdfBytes, 0644); err != nil {
		return fmt.Errorf("failed to write PDF file: %w", err)
	}

	log.Printf("DEBUG: Saved patient PDF to %s (%d bytes)", pdfPath, len(pdfBytes))
	return nil
}

func SavePatientDocumentReader(telegramID string, filename string, r io.Reader) (string, error) {
	docDir := filepath.Join(DataDir, "patients", telegramID, "documents")
	if err := os.MkdirAll(docDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create documents directory: %w", err)
	}

	filePath := filepath.Join(docDir, filename)
	f, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to create document file: %w", err)
	}
	defer f.Close()

	if _, err := io.Copy(f, r); err != nil {
		return "", fmt.Errorf("failed to save document data: %w", err)
	}

	// Update the patient record to include this new document
	patient, err := GetPatient(telegramID)
	if err == nil {
		// Just re-save to trigger markdown regeneration
		SavePatient(patient)
	}

	return filePath, nil
}

func SavePatientDocument(telegramID string, filename string, data []byte) (string, error) {
	docDir := filepath.Join(DataDir, "patients", telegramID, "documents")
	if err := os.MkdirAll(docDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create documents directory: %w", err)
	}

	filePath := filepath.Join(docDir, filename)
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return "", fmt.Errorf("failed to write document file: %w", err)
	}

	// Update the patient record to include this new document
	patient, err := GetPatient(telegramID)
	if err == nil {
		// Just re-save to trigger markdown regeneration
		SavePatient(patient)
	}

	return filePath, nil
}

func listDocuments(telegramID string) string {
	docDir := filepath.Join(DataDir, "patients", telegramID, "documents")
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

	// Sort by ModTime (newest first)
	sort.Slice(infos, func(i, j int) bool {
		return infos[i].modTime.After(infos[j].modTime)
	})

	var list string
	for _, info := range infos {
		// Obsidian format with time prefix for clinical precision
		list += fmt.Sprintf("- [%s] [[documents/%s|%s]]\n", info.modTime.Format("02.01.2006 15:04"), info.name, info.name)
	}
	return list
}

func GenerateHTMLRecord(p domain.Patient) string {
	docs := listDocuments(p.TelegramID)
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
		formatDocsForHTML(docs),
		time.Now().Format("02.01.2006 15:04"))
}

func formatTranscriptsSection(transcripts string) string {
	if transcripts == "" {
		return ""
	}
	return fmt.Sprintf(`
        <div class="section">
            <div class="section-title">🎙 Автоматические расшифровки</div>
            <div class="transcript-content">%s</div>
        </div>`, transcripts)
}

func formatDocsForHTML(docs string) string {
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

func GetPatient(telegramID string) (domain.Patient, error) {
	jsonPath := filepath.Join(DataDir, "patients", telegramID, "patient.json")

	data, err := os.ReadFile(jsonPath)
	if err != nil {
		return domain.Patient{}, fmt.Errorf("patient not found: %w", err)
	}

	var patient domain.Patient
	if err := json.Unmarshal(data, &patient); err != nil {
		return domain.Patient{}, fmt.Errorf("failed to parse patient data: %w", err)
	}

	return patient, nil
}

// BanUser adds a telegram ID to the blacklist
func BanUser(telegramID string) error {
	path := filepath.Join(DataDir, "blacklist.txt")
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	if isBanned, _ := IsUserBanned(telegramID, ""); isBanned {
		return nil
	}

	_, err = f.WriteString(telegramID + "\n")
	return err
}

// UnbanUser removes a telegram ID from the blacklist
func UnbanUser(telegramID string) error {
	path := filepath.Join(DataDir, "blacklist.txt")
	content, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	lines := strings.Split(string(content), "\n")
	var newLines []string
	for _, line := range lines {
		if strings.TrimSpace(line) != telegramID && strings.TrimSpace(line) != "" {
			newLines = append(newLines, line)
		}
	}

	return os.WriteFile(path, []byte(strings.Join(newLines, "\n")+"\n"), 0644)
}

// IsUserBanned checks if a telegram ID or username is in the blacklist
func IsUserBanned(telegramID string, username string) (bool, error) {
	path := filepath.Join(DataDir, "blacklist.txt")
	content, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}

	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		// Match numeric ID or @username
		if trimmed == telegramID || (username != "" && (trimmed == username || trimmed == "@"+username)) {
			return true, nil
		}
	}
	return false, nil
}

// CreateBackup creates a zip file of the entire patient data directory
func CreateBackup() (string, error) {
	backupDir := filepath.Join(DataDir, "backups")
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

	// Walk through the patient data directory
	patientsPath := filepath.Join(DataDir, "patients")
	err = filepath.Walk(patientsPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(DataDir, path)
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

	if err != nil {
		return "", fmt.Errorf("failed to walk and zip data: %w", err)
	}

	return backupPath, nil
}
