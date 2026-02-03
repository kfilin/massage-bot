package handlers

import (
	"strconv"
	"testing"
)

// TestAdminLogic replicates the exact logic used in the bot to check admin permissions
func TestAdminLogic(t *testing.T) {
	// 1. Simulate .env values
	envAdminID := "304528450"
	envAllowedIDs := []string{"304528450", "5331880756"}

	// 2. Simulate bot.go logic to build finalAdminIDs
	adminMap := make(map[string]bool)
	if envAdminID != "" {
		adminMap[envAdminID] = true
	}
	for _, id := range envAllowedIDs {
		if id != "" {
			adminMap[id] = true
		}
	}

	finalAdminIDs := make([]string, 0, len(adminMap))
	for id := range adminMap {
		finalAdminIDs = append(finalAdminIDs, id)
	}

	t.Logf("Final Admin IDs: %v", finalAdminIDs)

	// 3. Simulate Checking Logic in HandleBlock
	userID := int64(304528450) // User's ID
	userIDStr := strconv.FormatInt(userID, 10)

	isAdmin := false
	for _, id := range finalAdminIDs {
		if id == userIDStr {
			isAdmin = true
			break
		}
	}

	if !isAdmin {
		t.Fatalf("Logic Failed! User ID %d was NOT recognized as admin.", userID)
	} else {
		t.Logf("Success: User ID %d IS recognized as admin.", userID)
	}
}

func (m *mockRepository) DeleteAppointment(appointmentID string) error {
	return nil
}
