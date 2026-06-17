package storage

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/kfilin/massage-bot/internal/domain"
)

func TestGetPatientDir_DefaultName(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	dataDir := t.TempDir()
	// No existing patient folder; function should build "<name> (<id>)"
	// with illegal chars sanitised.
	if err := os.MkdirAll(filepath.Join(dataDir, "patients"), 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	repo := NewPostgresRepository(sqlx.NewDb(db, "sqlmock"), dataDir)
	got := repo.getPatientDir(domain.Patient{
		TelegramID: "123",
		Name:       "Alice Bob",
	})
	want := filepath.Join(dataDir, "patients", "Alice Bob (123)")
	if got != want {
		t.Errorf("default path: got %q, want %q", got, want)
	}
}

func TestGetPatientDir_IllegalChars(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	dataDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dataDir, "patients"), 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	repo := NewPostgresRepository(sqlx.NewDb(db, "sqlmock"), dataDir)
	got := repo.getPatientDir(domain.Patient{
		TelegramID: "999",
		Name:       `Bob/Smith:Evil<>"|?*Name`,
	})
	// Each illegal char replaced with underscore.
	wantSuffix := "Bob_Smith_Evil______Name (999)"
	if !strings.HasSuffix(got, wantSuffix) {
		t.Errorf("sanitised path: got %q, want suffix %q", got, wantSuffix)
	}
}

func TestGetPatientDir_SuffixMatch(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	dataDir := t.TempDir()
	patientsDir := filepath.Join(dataDir, "patients")
	if err := os.MkdirAll(patientsDir, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	// Existing folder with "(id)" suffix but a different display name
	// (e.g. user manually renamed in Obsidian). Function should match
	// the suffix and return the existing folder.
	existing := filepath.Join(patientsDir, "Renamed In Obsidian (777)")
	if err := os.MkdirAll(existing, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	// Plant a marker file so we can confirm we found the right dir.
	if err := os.WriteFile(filepath.Join(existing, "marker.txt"), []byte("x"), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}

	repo := NewPostgresRepository(sqlx.NewDb(db, "sqlmock"), dataDir)
	got := repo.getPatientDir(domain.Patient{
		TelegramID: "777",
		Name:       "Original Name",
	})
	if got != existing {
		t.Errorf("suffix match: got %q, want %q", got, existing)
	}
}

func TestGetPatientDir_NoPatientsDir(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	// dataDir has no "patients" subdir at all. The function should fall
	// straight through to the default-naming branch.
	dataDir := t.TempDir()
	repo := NewPostgresRepository(sqlx.NewDb(db, "sqlmock"), dataDir)
	got := repo.getPatientDir(domain.Patient{
		TelegramID: "1",
		Name:       "Solo",
	})
	want := filepath.Join(dataDir, "patients", "Solo (1)")
	if got != want {
		t.Errorf("no patients dir: got %q, want %q", got, want)
	}
}

// Helper: a minimal valid patient.json for migration tests.
func writePatientJSON(t *testing.T, dir, name string) {
	t.Helper()
	p := domain.Patient{TelegramID: name, Name: name}
	b, _ := json.Marshal(p)
	if err := os.WriteFile(filepath.Join(dir, "patient.json"), b, 0644); err != nil {
		t.Fatalf("write: %v", err)
	}
}
