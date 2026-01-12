package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/kfilin/massage-bot/internal/domain"
)

var DataDir = "data"

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

*Создано Vera Massage Bot • %s*`,
		p.Name,
		p.TelegramID,
		p.FirstVisit.Format("02.01.2006"),
		p.LastVisit.Format("02.01.2006"),
		p.TotalVisits,
		p.CurrentService,
		p.TherapistNotes,
		time.Now().Format("02.01.2006"))
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
