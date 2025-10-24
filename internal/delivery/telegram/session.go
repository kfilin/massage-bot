package telegram

import (
	"github.com/kfilin/massage-bot/internal/ports"
)

// InMemorySessionStorage is a simple in-memory implementation of SessionStorage for development.
// NOTE: This will lose all session data if the bot restarts.
// For production, consider using a persistent store like Redis or a database.
type InMemorySessionStorage struct {
	sessions map[int64]map[string]interface{} // userID -> {key -> value}
}

// NewInMemorySessionStorage creates a new in-memory session storage.
func NewInMemorySessionStorage() ports.SessionStorage { // Implements ports.SessionStorage
	return &InMemorySessionStorage{
		sessions: make(map[int64]map[string]interface{}),
	}
}

// Set stores a value in the session for a given user.
func (s *InMemorySessionStorage) Set(userID int64, key string, value interface{}) {
	if s.sessions[userID] == nil {
		s.sessions[userID] = make(map[string]interface{})
	}
	s.sessions[userID][key] = value
}

// Get retrieves the session data for a given user.
func (s *InMemorySessionStorage) Get(userID int64) map[string]interface{} {
	return s.sessions[userID]
}

// ClearSession clears all session data for a given user.
func (s *InMemorySessionStorage) ClearSession(userID int64) {
	delete(s.sessions, userID)
}
