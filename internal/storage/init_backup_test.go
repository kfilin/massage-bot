package storage

import (
	"archive/zip"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/kfilin/massage-bot/internal/domain"
)

// writePatientDir creates a patients/<id>/patient.json file in dataDir.
func writePatientDir(t *testing.T, dataDir, tgID, name string) {
	t.Helper()
	dir := filepath.Join(dataDir, "patients", tgID)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("mkdir patient dir: %v", err)
	}
	p := domain.Patient{
		TelegramID: tgID,
		Name:       name,
		FirstVisit: time.Now().Add(-24 * time.Hour),
		LastVisit:  time.Now(),
		TotalVisits: 3,
	}
	b, _ := json.Marshal(p)
	if err := os.WriteFile(filepath.Join(dir, "patient.json"), b, 0644); err != nil {
		t.Fatalf("write patient.json: %v", err)
	}
}

func TestMigrateJSONToPostgres_NoPatientsDir(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	repo := NewPostgresRepository(sqlx.NewDb(db, "sqlmock"), t.TempDir())

	// Missing patients dir -> nil, no error.
	if err := MigrateJSONToPostgres(repo, t.TempDir()); err != nil {
		t.Errorf("expected nil error when patients dir is missing, got %v", err)
	}
}

func TestMigrateJSONToPostgres_EmptyPatientsDir(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	repo := NewPostgresRepository(sqlx.NewDb(db, "sqlmock"), t.TempDir())

	dataDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dataDir, "patients"), 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := MigrateJSONToPostgres(repo, dataDir); err != nil {
		t.Errorf("expected nil error for empty patients dir, got %v", err)
	}
}

func TestMigrateJSONToPostgres_HappyPath(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	repo := NewPostgresRepository(sqlx.NewDb(db, "sqlmock"), t.TempDir())

	dataDir := t.TempDir()
	writePatientDir(t, dataDir, "tg-100", "Alice")
	writePatientDir(t, dataDir, "tg-200", "Bob")
	// Directories without patient.json should be skipped.
	if err := os.MkdirAll(filepath.Join(dataDir, "patients", "nojson"), 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	// Expect two INSERT calls.
	for i := 0; i < 2; i++ {
		mock.ExpectExec("INSERT INTO patients").
			WillReturnResult(sqlmock.NewResult(1, 1))
	}

	if err := MigrateJSONToPostgres(repo, dataDir); err != nil {
		t.Errorf("MigrateJSONToPostgres failed: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestMigrateJSONToPostgres_MalformedJSON(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	repo := NewPostgresRepository(sqlx.NewDb(db, "sqlmock"), t.TempDir())

	dataDir := t.TempDir()
	dir := filepath.Join(dataDir, "patients", "tg-bad")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	// Write invalid JSON
	if err := os.WriteFile(filepath.Join(dir, "patient.json"), []byte("{not json"), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}

	// Malformed JSON is logged+skipped, no SQL expected.
	if err := MigrateJSONToPostgres(repo, dataDir); err != nil {
		t.Errorf("expected nil error (malformed is skipped), got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("expected no SQL, got: %v", err)
	}
}

func TestMigrateJSONToPostgres_SaveError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	repo := NewPostgresRepository(sqlx.NewDb(db, "sqlmock"), t.TempDir())

	dataDir := t.TempDir()
	writePatientDir(t, dataDir, "tg-err", "Charlie")
	writePatientDir(t, dataDir, "tg-ok", "Dana")

	// First save fails, second succeeds. Both are attempted.
	mock.ExpectExec("INSERT INTO patients").
		WillReturnError(fmt.Errorf("db down"))
	mock.ExpectExec("INSERT INTO patients").
		WillReturnResult(sqlmock.NewResult(1, 1))

	if err := MigrateJSONToPostgres(repo, dataDir); err != nil {
		t.Errorf("expected nil (errors are logged+skipped), got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

// --- CreateBackup tests ---------------------------------------------------

// makeFakePgDump writes a shell script that mimics pg_dump: writes a
// minimal SQL file to the path given by -f and exits 0. Returns the
// directory the script lives in, which the caller prepends to PATH.
func makeFakePgDump(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	// The fake script must be executable.
	script := `#!/bin/sh
# Fake pg_dump for tests. Parse -f <path> and write a stub SQL file.
out=""
for arg in "$@"; do
  case "$prev" in
    -f) out="$arg" ;;
  esac
  prev="$arg"
done
if [ -z "$out" ]; then
  echo "fake pg_dump: -f not specified" >&2
  exit 2
fi
mkdir -p "$(dirname "$out")"
printf '-- fake pg_dump\nCREATE TABLE patients(id int);\n' > "$out"
exit 0
`
	path := filepath.Join(dir, "pg_dump")
	if err := os.WriteFile(path, []byte(script), 0755); err != nil {
		t.Fatalf("write fake pg_dump: %v", err)
	}
	return dir
}

// withEnv sets env vars for the duration of a test, restoring old values.
func withEnv(t *testing.T, kv map[string]string) {
	t.Helper()
	old := make(map[string]string)
	for k, v := range kv {
		old[k] = os.Getenv(k)
		os.Setenv(k, v)
	}
	t.Cleanup(func() {
		for k, v := range old {
			if v == "" {
				os.Unsetenv(k)
			} else {
				os.Setenv(k, v)
			}
		}
	})
}

func TestCreateBackup_Success(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	dataDir := t.TempDir()
	// Create a few patients files to be included in the backup.
	writePatientDir(t, dataDir, "tg-1", "Alice")
	writePatientDir(t, dataDir, "tg-2", "Bob")

	binDir := makeFakePgDump(t)
	origPath := os.Getenv("PATH")
	os.Setenv("PATH", binDir+string(os.PathListSeparator)+origPath)
	t.Cleanup(func() { os.Setenv("PATH", origPath) })

	withEnv(t, map[string]string{
		"DB_HOST":     "localhost",
		"DB_PORT":     "5432",
		"DB_USER":     "user",
		"DB_PASSWORD": "pw",
		"DB_NAME":     "db",
	})

	repo := NewPostgresRepository(sqlx.NewDb(db, "sqlmock"), dataDir)
	zipPath, err := repo.CreateBackup()
	if err != nil {
		t.Fatalf("CreateBackup failed: %v", err)
	}
	if !strings.HasPrefix(zipPath, dataDir) {
		t.Errorf("zip path %q should be inside dataDir %q", zipPath, dataDir)
	}
	if _, err := os.Stat(zipPath); err != nil {
		t.Errorf("zip file missing: %v", err)
	}

	// Inspect the zip: must contain db_dump.sql and the two patient dirs.
	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		t.Fatalf("open zip: %v", err)
	}
	defer reader.Close()
	names := make(map[string]bool)
	for _, f := range reader.File {
		names[f.Name] = true
	}
	if !names["db_dump.sql"] {
		t.Errorf("zip missing db_dump.sql; got %v", names)
	}
	if !names["patients/tg-1/patient.json"] {
		t.Errorf("zip missing patients/tg-1/patient.json; got %v", names)
	}
	if !names["patients/tg-2/patient.json"] {
		t.Errorf("zip missing patients/tg-2/patient.json; got %v", names)
	}

	// temp_backups dir should be removed after success.
	if _, err := os.Stat(filepath.Join(dataDir, "temp_backups")); !os.IsNotExist(err) {
		t.Errorf("temp_backups should be removed after success, stat err=%v", err)
	}
}

func TestCreateBackup_PgDumpFailure(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	// Fake pg_dump that always fails.
	dir := t.TempDir()
	script := "#!/bin/sh\necho 'fake pg_dump: connection refused' >&2\nexit 1\n"
	if err := os.WriteFile(filepath.Join(dir, "pg_dump"), []byte(script), 0755); err != nil {
		t.Fatalf("write: %v", err)
	}
	origPath := os.Getenv("PATH")
	os.Setenv("PATH", dir+string(os.PathListSeparator)+origPath)
	t.Cleanup(func() { os.Setenv("PATH", origPath) })

	withEnv(t, map[string]string{
		"DB_HOST": "localhost", "DB_PORT": "5432",
		"DB_USER": "u", "DB_PASSWORD": "p", "DB_NAME": "d",
	})

	repo := NewPostgresRepository(sqlx.NewDb(db, "sqlmock"), t.TempDir())
	_, err = repo.CreateBackup()
	if err == nil {
		t.Error("CreateBackup should fail when pg_dump fails")
	}
	if !strings.Contains(err.Error(), "pg_dump failed") {
		t.Errorf("expected pg_dump failure in error, got: %v", err)
	}
}

// TestCreateBackup_NonExistentPgDump ensures we fail cleanly when pg_dump
// is not on PATH at all (rather than hanging or producing a partial file).
func TestCreateBackup_NonExistentPgDump(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	// Empty PATH.
	origPath := os.Getenv("PATH")
	os.Setenv("PATH", t.TempDir())
	t.Cleanup(func() { os.Setenv("PATH", origPath) })

	withEnv(t, map[string]string{
		"DB_HOST": "localhost", "DB_PORT": "5432",
		"DB_USER": "u", "DB_PASSWORD": "p", "DB_NAME": "d",
	})

	repo := NewPostgresRepository(sqlx.NewDb(db, "sqlmock"), t.TempDir())
	_, err = repo.CreateBackup()
	if err == nil {
		t.Error("CreateBackup should fail when pg_dump is not on PATH")
	}
}

// TestLogEvent_NilDetails ensures the nil-details path is covered.
func TestLogEvent_NilDetails(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectExec("INSERT INTO analytics_events").
		WithArgs("p1", "evt", []byte("null")).
		WillReturnResult(sqlmock.NewResult(1, 1))

	repo := NewPostgresRepository(sqlx.NewDb(db, "sqlmock"), t.TempDir())
	if err := repo.LogEvent("p1", "evt", nil); err != nil {
		t.Errorf("LogEvent with nil details failed: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled: %v", err)
	}
}

// TestSaveAppointmentMetadata_NilReminders covers the nil map branch.
func TestSaveAppointmentMetadata_NilReminders(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectExec("INSERT INTO appointment_meta").
		WithArgs("appt-1", sqlmock.AnyArg(), []byte("null")).
		WillReturnResult(sqlmock.NewResult(1, 1))

	repo := NewPostgresRepository(sqlx.NewDb(db, "sqlmock"), t.TempDir())
	if err := repo.SaveAppointmentMetadata("appt-1", nil, nil); err != nil {
		t.Errorf("SaveAppointmentMetadata with nil reminders failed: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled: %v", err)
	}
}

// TestGetPatient_NotFound_Error covers the sql.ErrNoRows path explicitly
// even though the integration test does too — guards against regressions
// if a future refactor drops the error mapping.
func TestGetPatient_NotFound_ReturnsErrNoRows(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery("SELECT .* FROM patients WHERE telegram_id").
		WithArgs("missing").
		WillReturnError(sql.ErrNoRows)

	repo := NewPostgresRepository(sqlx.NewDb(db, "sqlmock"), t.TempDir())
	_, err = repo.GetPatient("missing")
	if err != sql.ErrNoRows {
		t.Errorf("expected sql.ErrNoRows, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled: %v", err)
	}
}
