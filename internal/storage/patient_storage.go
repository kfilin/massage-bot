package storage

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
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

	// 2. Save Markdown record (for patients to download)
	mdPath := filepath.Join(patientDir, "medical_record.md")
	mdContent := generateMarkdownRecord(patient)

	if err := os.WriteFile(mdPath, []byte(mdContent), 0644); err != nil {
		return fmt.Errorf("failed to write markdown file: %w", err)
	}

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

func generateMarkdownRecord(p domain.Patient) string {
	return fmt.Sprintf(`# Медицинская карта: %s

**Telegram ID:** %s  
**Первое посещение:** %s  
**Последний визит:** %s  
**Всего посещений:** %d  
**Текущая услуга:** %s

## Заметки терапевта
%s

## Как открыть этот файл
1. **Рекомендуем Obsidian** (бесплатно) — это мощный инструмент для ведения заметок, который превратит вашу медицинскую карту в удобную базу данных. Он доступен для **всех ваших устройств**:
   - 💻 **Компьютер:** Windows, macOS, Linux
   - 📱 **Мобильный:** Скачайте в App Store или Google Play
   *Скачайте на [obsidian.md/download](https://obsidian.md/download)*
2. **Или любой текстовый редактор** (Блокнот, TextEdit).

## Прикрепленные документы
%s

*Создано Vera Massage Bot • %s*`,
		p.Name,
		p.TelegramID,
		p.FirstVisit.Format("02.01.2006"),
		p.LastVisit.Format("02.01.2006"),
		p.TotalVisits,
		p.CurrentService,
		p.TherapistNotes,
		listDocuments(p.TelegramID),
		time.Now().Format("02.01.2006"))
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

func GetPatientMarkdownFile(telegramID string) (string, error) {
	mdPath := filepath.Join(DataDir, "patients", telegramID, "medical_record.md")

	if _, err := os.Stat(mdPath); err != nil {
		return "", fmt.Errorf("medical record not found: %w", err)
	}

	return mdPath, nil
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
