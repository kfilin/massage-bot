package storage

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/kfilin/massage-bot/internal/domain"
	"github.com/kfilin/massage-bot/internal/ports"
)

func MigrateJSONToPostgres(repo ports.Repository, dataDir string) error {
	patientsPath := filepath.Join(dataDir, "patients")
	entries, err := os.ReadDir(patientsPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // Nothing to migrate
		}
		return err
	}

	count := 0
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		patientID := entry.Name()
		jsonPath := filepath.Join(patientsPath, patientID, "patient.json")

		data, err := os.ReadFile(jsonPath)
		if err != nil {
			log.Printf("WARNING: Could not read patient.json for %s: %v", patientID, err)
			continue
		}

		var patient domain.Patient
		if err := json.Unmarshal(data, &patient); err != nil {
			log.Printf("WARNING: Could not parse patient.json for %s: %v", patientID, err)
			continue
		}

		// Save to Postgres
		if err := repo.SavePatient(patient); err != nil {
			log.Printf("ERROR: Could not migrate patient %s to Postgres: %v", patientID, err)
			continue
		}
		count++
	}

	// Migrate Blacklist
	blacklistPath := filepath.Join(dataDir, "blacklist.txt")
	content, err := os.ReadFile(blacklistPath)
	if err == nil {
		lines := strings.Split(string(content), "\n")
		for _, line := range lines {
			id := strings.TrimSpace(line)
			if id == "" {
				continue
			}
			if err := repo.BanUser(id); err != nil {
				log.Printf("WARNING: Could not migrate banned user %s: %v", id, err)
			}
		}
	}

	log.Printf("Migration completed: %d patients migrated.", count)
	return nil
}
