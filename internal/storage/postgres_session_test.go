package storage

import (
	"encoding/json"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
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
