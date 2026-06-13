package storage

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/kfilin/massage-bot/internal/domain"
	"github.com/kfilin/massage-bot/internal/monitoring"
)

func TestPostgresSessionStorage_Metrics(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")

	// 1. Test existing sessions loading
	rows := sqlmock.NewRows([]string{"user_id", "data"}).
		AddRow(100, []byte("{}")).
		AddRow(101, []byte("{}"))

	mock.ExpectQuery("SELECT user_id, data FROM sessions").WillReturnRows(rows)

	repo := NewPostgresSessionStorage(sqlxDB)

	// Check initial count
	if val := monitoring.GetActiveSessions(); val != 2 {
		t.Errorf("Initial active sessions = %d, want 2", val)
	}

	// 2. Test Set (New User)
	userID := int64(102)
	data, _ := json.Marshal(map[string]interface{}{"foo": "bar"})

	mock.ExpectExec("INSERT INTO sessions").
		WithArgs(userID, data).
		WillReturnResult(sqlmock.NewResult(1, 1))

	repo.Set(userID, "foo", "bar")

	if val := monitoring.GetActiveSessions(); val != 3 {
		t.Errorf("Active sessions after Set(new) = %d, want 3", val)
	}

	// 3. Test Set (Existing User)
	data2, _ := json.Marshal(map[string]interface{}{"foo": "baz"})
	mock.ExpectExec("INSERT INTO sessions").
		WithArgs(userID, data2).
		WillReturnResult(sqlmock.NewResult(1, 1))

	repo.Set(userID, "foo", "baz")

	if val := monitoring.GetActiveSessions(); val != 3 {
		t.Errorf("Active sessions after Set(existing) = %d, want 3", val)
	}

	// 4. Test ClearSession (Existing User)
	mock.ExpectExec("DELETE FROM sessions").
		WithArgs(userID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	repo.ClearSession(userID)

	if val := monitoring.GetActiveSessions(); val != 2 {
		t.Errorf("Active sessions after ClearSession = %d, want 2", val)
	}

	// 5. Test ClearSession (Non-Existing User)
	mock.ExpectExec("DELETE FROM sessions").
		WithArgs(int64(999)).
		WillReturnResult(sqlmock.NewResult(0, 0))

	repo.ClearSession(999)

	if val := monitoring.GetActiveSessions(); val != 2 {
		t.Errorf("Active sessions after ClearSession(non-active) = %d, want 2", val)
	}
}

func TestSessionStorage_Get(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")

	rows := sqlmock.NewRows([]string{"user_id", "data"}).
		AddRow(100, []byte(`{"key1":"val1","key2":42}`))

	mock.ExpectQuery("SELECT user_id, data FROM sessions").WillReturnRows(rows)

	repo := NewPostgresSessionStorage(sqlxDB)

	// Get existing session
	session := repo.Get(100)
	if session == nil {
		t.Fatal("Expected session for user 100, got nil")
	}
	if session["key1"] != "val1" {
		t.Errorf("Expected key1=val1, got %v", session["key1"])
	}

	// Get non-existing session
	session = repo.Get(999)
	if session != nil {
		t.Errorf("Expected nil for non-existing user, got %v", session)
	}
}

func TestFixSessionData(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := &PostgresSessionStorage{
		db:       sqlxDB,
		sessions: make(map[int64]map[string]interface{}),
	}

	t.Run("Service map converted to domain.Service", func(t *testing.T) {
		data := map[string]interface{}{
			"service": map[string]interface{}{
				"id":       "massage-1",
				"name":     "Classic",
				"duration": 60.0,
				"price":    100.0,
			},
		}
		fixed := repo.fixSessionData(data)
		svc, ok := fixed["service"].(domain.Service)
		if !ok {
			t.Fatalf("Expected domain.Service, got %T", fixed["service"])
		}
		if svc.ID != "massage-1" || svc.Name != "Classic" {
			t.Errorf("Unexpected service: %+v", svc)
		}
	})

	t.Run("Service non-map passthrough", func(t *testing.T) {
		data := map[string]interface{}{
			"service": "already-a-string",
		}
		fixed := repo.fixSessionData(data)
		if fixed["service"] != "already-a-string" {
			t.Errorf("Expected string passthrough, got %v", fixed["service"])
		}
	})

	t.Run("Date string parsed to time.Time", func(t *testing.T) {
		now := time.Now().Truncate(time.Second)
		data := map[string]interface{}{
			"date": now.Format(time.RFC3339),
		}
		fixed := repo.fixSessionData(data)
		parsed, ok := fixed["date"].(time.Time)
		if !ok {
			t.Fatalf("Expected time.Time, got %T", fixed["date"])
		}
		if !parsed.Equal(now) {
			t.Errorf("Expected %v, got %v", now, parsed)
		}
	})

	t.Run("Date invalid string passthrough", func(t *testing.T) {
		data := map[string]interface{}{
			"date": "not-a-date",
		}
		fixed := repo.fixSessionData(data)
		if fixed["date"] != "not-a-date" {
			t.Errorf("Expected string passthrough for invalid date, got %v", fixed["date"])
		}
	})

	t.Run("Date non-string passthrough", func(t *testing.T) {
		data := map[string]interface{}{
			"date": 42,
		}
		fixed := repo.fixSessionData(data)
		if fixed["date"] != 42 {
			t.Errorf("Expected int passthrough, got %v", fixed["date"])
		}
	})

	t.Run("Unknown key passthrough", func(t *testing.T) {
		data := map[string]interface{}{
			"custom_key": "custom_value",
		}
		fixed := repo.fixSessionData(data)
		if fixed["custom_key"] != "custom_value" {
			t.Errorf("Expected passthrough, got %v", fixed["custom_key"])
		}
	})
}

func TestLoadAllSessions_DBError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")

	mock.ExpectQuery("SELECT user_id, data FROM sessions").
		WillReturnError(fmt.Errorf("connection refused"))

	// Should not panic, just log error
	repo := NewPostgresSessionStorage(sqlxDB)

	// Sessions should be empty
	if val := monitoring.GetActiveSessions(); val != 0 {
		t.Errorf("Expected 0 active sessions on DB error, got %d", val)
	}

	// Verify we can still use the storage
	repo.Set(1, "key", "value")
	session := repo.Get(1)
	if session["key"] != "value" {
		t.Error("Storage should still work after load error")
	}
}

func TestLoadAllSessions_MalformedJSON(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")

	rows := sqlmock.NewRows([]string{"user_id", "data"}).
		AddRow(100, []byte(`{"valid": "json"}`)).
		AddRow(101, []byte(`{broken json`)). // malformed
		AddRow(102, []byte(`{"also": "valid"}`))

	mock.ExpectQuery("SELECT user_id, data FROM sessions").WillReturnRows(rows)

	repo := NewPostgresSessionStorage(sqlxDB)

	// Valid sessions should be loaded, malformed one skipped
	session100 := repo.Get(100)
	if session100 == nil || session100["valid"] != "json" {
		t.Error("Expected valid session for user 100")
	}

	session101 := repo.Get(101)
	if session101 != nil {
		t.Errorf("Expected nil for malformed session user 101, got %v", session101)
	}

	session102 := repo.Get(102)
	if session102 == nil || session102["also"] != "valid" {
		t.Error("Expected valid session for user 102")
	}
}
