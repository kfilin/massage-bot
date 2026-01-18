package storage

import (
	"testing"
	"time"

	"github.com/kfilin/massage-bot/internal/domain"
)

func TestSaveAndGetPatient(t *testing.T) {
	// Setup temporary data directory
	tmpDir := t.TempDir()
	DataDir = tmpDir

	patient := domain.Patient{
		TelegramID:     "12345",
		Name:           "Test User",
		FirstVisit:     time.Now(),
		LastVisit:      time.Now(),
		TotalVisits:    1,
		CurrentService: "Massage",
	}

	// Test Save
	err := SavePatient(patient)
	if err != nil {
		t.Fatalf("Failed to save patient: %v", err)
	}

	// Test Get
	retrieved, err := GetPatient("12345")
	if err != nil {
		t.Fatalf("Failed to get patient: %v", err)
	}

	if retrieved.Name != patient.Name {
		t.Errorf("Expected name %s, got %s", patient.Name, retrieved.Name)
	}
}

func TestBanSystem(t *testing.T) {
	tmpDir := t.TempDir()
	DataDir = tmpDir

	userID := "99999"

	// Should not be banned initially
	banned, err := IsUserBanned(userID, "")
	if err != nil {
		t.Errorf("Unexpected error checking ban: %v", err)
	}
	if banned {
		t.Error("User should not be banned initially")
	}

	// Ban user
	err = BanUser(userID)
	if err != nil {
		t.Fatalf("Failed to ban user: %v", err)
	}

	// Should be banned now
	banned, err = IsUserBanned(userID, "")
	if err != nil {
		t.Errorf("Unexpected error checking ban after banning: %v", err)
	}
	if !banned {
		t.Error("User should be banned")
	}

	// Unban user
	err = UnbanUser(userID)
	if err != nil {
		t.Fatalf("Failed to unban user: %v", err)
	}

	// Should not be banned again
	banned, err = IsUserBanned(userID, "")
	if err != nil {
		t.Errorf("Unexpected error checking ban after unbanning: %v", err)
	}
	if banned {
		t.Error("User should not be banned after unbanning")
	}

	// Test username banning
	username := "testuser"
	err = BanUser("@" + username)
	if err != nil {
		t.Fatalf("Failed to ban username: %v", err)
	}

	banned, err = IsUserBanned("diff_id", username)
	if err != nil {
		t.Errorf("Error checking username ban: %v", err)
	}
	if !banned {
		t.Error("Username should be banned")
	}
}
