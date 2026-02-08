package storage

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/kfilin/massage-bot/internal/logging"
	"github.com/kfilin/massage-bot/internal/monitoring"

	"github.com/jmoiron/sqlx"
	"github.com/kfilin/massage-bot/internal/domain"
	"github.com/kfilin/massage-bot/internal/ports"
)

type PostgresSessionStorage struct {
	db       *sqlx.DB
	sessions map[int64]map[string]interface{}
	mu       sync.RWMutex
}

func NewPostgresSessionStorage(db *sqlx.DB) ports.SessionStorage {
	s := &PostgresSessionStorage{
		db:       db,
		sessions: make(map[int64]map[string]interface{}),
	}
	s.loadAllSessions()
	monitoring.UpdateActiveSessions(len(s.sessions))
	return s
}

func (s *PostgresSessionStorage) loadAllSessions() {
	var rows []struct {
		UserID int64  `db:"user_id"`
		Data   []byte `db:"data"`
	}
	err := s.db.Select(&rows, "SELECT user_id, data FROM sessions")
	if err != nil {
		logging.Errorf(": Failed to load sessions from DB: %v", err)
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for _, row := range rows {
		var data map[string]interface{}
		if err := json.Unmarshal(row.Data, &data); err != nil {
			logging.Warnf("ING: Failed to unmarshal session for user %d: %v", row.UserID, err)
			continue
		}
		s.sessions[row.UserID] = s.fixSessionData(data)
	}
	logging.Infof("Loaded %d sessions from database", len(s.sessions))
}

func (s *PostgresSessionStorage) fixSessionData(data map[string]interface{}) map[string]interface{} {
	fixed := make(map[string]interface{})
	for k, v := range data {
		switch k {
		case "service": // handlers.SessionKeyService
			if m, ok := v.(map[string]interface{}); ok {
				var svc domain.Service
				b, _ := json.Marshal(m)
				if err := json.Unmarshal(b, &svc); err != nil {
					logging.Errorf(": Failed to unmarshal session service data: %v", err)
				}
				fixed[k] = svc
			} else {
				fixed[k] = v
			}
		case "date": // handlers.SessionKeyDate
			if str, ok := v.(string); ok {
				if t, err := time.Parse(time.RFC3339, str); err == nil {
					fixed[k] = t
				} else {
					fixed[k] = v
				}
			} else {
				fixed[k] = v
			}
		default:
			fixed[k] = v
		}
	}
	return fixed
}

func (s *PostgresSessionStorage) Set(userID int64, key string, value interface{}) {
	s.mu.Lock()
	isNew := s.sessions[userID] == nil
	if isNew {
		s.sessions[userID] = make(map[string]interface{})
	}
	s.sessions[userID][key] = value
	currentCount := len(s.sessions)
	// Serialize while under lock
	data, err := json.Marshal(s.sessions[userID])
	s.mu.Unlock()

	if isNew {
		monitoring.UpdateActiveSessions(currentCount)
	}

	if err != nil {
		logging.Errorf("Failed to marshal session for user %d: %v", userID, err)
		return
	}

	_, err = s.db.Exec(`
		INSERT INTO sessions (user_id, data, updated_at) 
		VALUES ($1, $2, CURRENT_TIMESTAMP)
		ON CONFLICT (user_id) DO UPDATE SET data = EXCLUDED.data, updated_at = CURRENT_TIMESTAMP
	`, userID, data)
	if err != nil {
		logging.Errorf("Failed to save session to DB for user %d: %v", userID, err)
	}
}

func (s *PostgresSessionStorage) Get(userID int64) map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.sessions[userID]
}

func (s *PostgresSessionStorage) ClearSession(userID int64) {
	s.mu.Lock()
	_, exists := s.sessions[userID]
	delete(s.sessions, userID)
	currentCount := len(s.sessions)
	s.mu.Unlock()

	if exists {
		monitoring.UpdateActiveSessions(currentCount)
	}

	_, err := s.db.Exec("DELETE FROM sessions WHERE user_id = $1", userID)
	if err != nil {
		logging.Errorf("Failed to delete session from DB for user %d: %v", userID, err)
	}
}
