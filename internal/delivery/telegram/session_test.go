package telegram

import (
	"testing"
)

func TestSession_SetAndGet(t *testing.T) {
	s := NewInMemorySessionStorage()

	s.Set(100, "key", "value")
	data := s.Get(100)

	if val, ok := data["key"]; !ok || val != "value" {
		t.Errorf("Expected 'value' for key 'key', got %v", val)
	}
}

func TestSession_GetNewUser_ReturnsNilOrEmpty(t *testing.T) {
	s := NewInMemorySessionStorage()
	data := s.Get(999)
	// Implementation returns nil for unknown user — safe to use with range
	if data != nil && len(data) != 0 {
		t.Errorf("Expected nil or empty session for unknown user, got %v", data)
	}
}

func TestSession_ClearSession(t *testing.T) {
	s := NewInMemorySessionStorage()
	s.Set(42, "foo", "bar")
	s.ClearSession(42)

	data := s.Get(42)
	if data != nil && len(data) != 0 {
		t.Errorf("Expected empty session after ClearSession, got %v", data)
	}
}

func TestSession_MultipleKeys(t *testing.T) {
	s := NewInMemorySessionStorage()
	s.Set(1, "a", 1)
	s.Set(1, "b", 2)
	s.Set(1, "c", "three")

	data := s.Get(1)
	if data["a"] != 1 || data["b"] != 2 || data["c"] != "three" {
		t.Errorf("Expected all 3 keys set, got %v", data)
	}
}

func TestSession_Overwrite(t *testing.T) {
	s := NewInMemorySessionStorage()
	s.Set(5, "x", "first")
	s.Set(5, "x", "second")

	data := s.Get(5)
	if data["x"] != "second" {
		t.Errorf("Expected overwritten value 'second', got %v", data["x"])
	}
}

func TestSession_IsolatedUsers(t *testing.T) {
	s := NewInMemorySessionStorage()
	s.Set(1, "key", "user1value")
	s.Set(2, "key", "user2value")

	d1 := s.Get(1)
	d2 := s.Get(2)

	if d1["key"] != "user1value" {
		t.Errorf("User 1 value wrong: %v", d1["key"])
	}
	if d2["key"] != "user2value" {
		t.Errorf("User 2 value wrong: %v", d2["key"])
	}
}

func TestSession_MultipleUsersSequential(t *testing.T) {
	s := NewInMemorySessionStorage()
	// Set keys for 10 different users sequentially and verify isolation
	for i := int64(0); i < 10; i++ {
		s.Set(i, "v", i)
	}
	for i := int64(0); i < 10; i++ {
		val := s.Get(i)["v"]
		if val != i {
			t.Errorf("User %d: expected value %d, got %v", i, i, val)
		}
	}
}
