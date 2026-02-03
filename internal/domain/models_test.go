package domain

import (
	"testing"
	"time"
)

// TestSplitSummary tests the SplitSummary function with various input formats
func TestSplitSummary(t *testing.T) {
	tests := []struct {
		name        string
		summary     string
		wantLen     int
		wantService string
		wantName    string
	}{
		{
			name:        "Standard format with dash separator",
			summary:     "Massage - John Doe",
			wantLen:     2,
			wantService: "Massage",
			wantName:    "John Doe",
		},
		{
			name:        "Service only (no customer name)",
			summary:     "Massage",
			wantLen:     1,
			wantService: "Massage",
			wantName:    "",
		},
		{
			name:        "Multi-word service and name",
			summary:     "Full Body Massage - Jane Smith",
			wantLen:     2,
			wantService: "Full Body Massage",
			wantName:    "Jane Smith",
		},
		{
			name:        "Name with multiple dashes",
			summary:     "Massage - Jane - Doe",
			wantLen:     2,
			wantService: "Massage - Jane",
			wantName:    "Doe",
		},
		{
			name:        "Empty string",
			summary:     "",
			wantLen:     1,
			wantService: "",
			wantName:    "",
		},
		{
			name:        "Only dash separator",
			summary:     " - ",
			wantLen:     2,
			wantService: "",
			wantName:    "",
		},
		{
			name:        "Dash without spaces (should not split)",
			summary:     "Service-Name",
			wantLen:     1,
			wantService: "Service-Name",
			wantName:    "",
		},
		{
			name:        "Multiple dash separators",
			summary:     "Deep Tissue Massage - John - Doe - Smith",
			wantLen:     2,
			wantService: "Deep Tissue Massage - John - Doe",
			wantName:    "Smith",
		},
		{
			name:        "Cyrillic characters",
			summary:     "Массаж - Иван Иванов",
			wantLen:     2,
			wantService: "Массаж",
			wantName:    "Иван Иванов",
		},
		{
			name:        "Special characters in name",
			summary:     "Massage - O'Brien",
			wantLen:     2,
			wantService: "Massage",
			wantName:    "O'Brien",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SplitSummary(tt.summary)

			if len(result) != tt.wantLen {
				t.Errorf("SplitSummary() returned %d parts, want %d", len(result), tt.wantLen)
			}

			if len(result) > 0 && result[0] != tt.wantService {
				t.Errorf("SplitSummary() service = %q, want %q", result[0], tt.wantService)
			}

			if len(result) > 1 && result[1] != tt.wantName {
				t.Errorf("SplitSummary() name = %q, want %q", result[1], tt.wantName)
			}
		})
	}
}

// TestTimeConstants verifies that time-related constants are valid
func TestTimeConstants(t *testing.T) {
	t.Run("WorkDayStartHour is valid", func(t *testing.T) {
		if WorkDayStartHour < 0 || WorkDayStartHour > 23 {
			t.Errorf("WorkDayStartHour = %d, must be between 0 and 23", WorkDayStartHour)
		}
	})

	t.Run("WorkDayEndHour is valid", func(t *testing.T) {
		if WorkDayEndHour < 0 || WorkDayEndHour > 23 {
			t.Errorf("WorkDayEndHour = %d, must be between 0 and 23", WorkDayEndHour)
		}
	})

	t.Run("WorkDay hours are logical", func(t *testing.T) {
		if WorkDayStartHour >= WorkDayEndHour {
			t.Errorf("WorkDayStartHour (%d) must be less than WorkDayEndHour (%d)",
				WorkDayStartHour, WorkDayEndHour)
		}
	})

	t.Run("ApptTimeZone is initialized", func(t *testing.T) {
		if ApptTimeZone == nil {
			t.Fatal("ApptTimeZone is nil, should be initialized in init()")
		}
	})

	t.Run("ApptTimeZone is Europe/Istanbul", func(t *testing.T) {
		if ApptTimeZone.String() != "Europe/Istanbul" {
			t.Errorf("ApptTimeZone = %s, want Europe/Istanbul", ApptTimeZone.String())
		}
	})

	t.Run("SlotDuration is initialized", func(t *testing.T) {
		if SlotDuration == nil {
			t.Fatal("SlotDuration is nil, should be initialized in init()")
		}
	})

	t.Run("SlotDuration is 60 minutes", func(t *testing.T) {
		expected := 60 * time.Minute
		if *SlotDuration != expected {
			t.Errorf("SlotDuration = %v, want %v", *SlotDuration, expected)
		}
	})
}

// TestServiceStruct verifies the Service struct can be created and used
func TestServiceStruct(t *testing.T) {
	service := Service{
		ID:              "massage-60",
		Name:            "Classic Massage",
		DurationMinutes: 60,
		Price:           50.0,
		Description:     "A relaxing full-body massage",
	}

	if service.ID != "massage-60" {
		t.Errorf("Service.ID = %s, want massage-60", service.ID)
	}
	if service.DurationMinutes != 60 {
		t.Errorf("Service.DurationMinutes = %d, want 60", service.DurationMinutes)
	}
	if service.Price != 50.0 {
		t.Errorf("Service.Price = %.2f, want 50.00", service.Price)
	}
}

// TestTimeSlotStruct verifies the TimeSlot struct
func TestTimeSlotStruct(t *testing.T) {
	start := time.Date(2026, 2, 3, 10, 0, 0, 0, ApptTimeZone)
	end := start.Add(60 * time.Minute)

	slot := TimeSlot{
		Start: start,
		End:   end,
	}

	if !slot.Start.Equal(start) {
		t.Errorf("TimeSlot.Start = %v, want %v", slot.Start, start)
	}
	if !slot.End.Equal(end) {
		t.Errorf("TimeSlot.End = %v, want %v", slot.End, end)
	}

	duration := slot.End.Sub(slot.Start)
	if duration != 60*time.Minute {
		t.Errorf("TimeSlot duration = %v, want 60m", duration)
	}
}

// TestAppointmentStruct verifies the Appointment struct
func TestAppointmentStruct(t *testing.T) {
	now := time.Now().In(ApptTimeZone)
	confirmedAt := now.Add(-24 * time.Hour)

	appt := Appointment{
		ID:              "event-123",
		ServiceID:       "massage-60",
		Time:            now,
		Duration:        60,
		StartTime:       now,
		EndTime:         now.Add(60 * time.Minute),
		CustomerName:    "John Doe",
		CustomerTgID:    "123456789",
		CalendarEventID: "gcal-event-123",
		Status:          "confirmed",
		ConfirmedAt:     &confirmedAt,
		RemindersSent:   map[string]bool{"72h": true, "24h": false},
	}

	if appt.ID != "event-123" {
		t.Errorf("Appointment.ID = %s, want event-123", appt.ID)
	}
	if appt.Duration != 60 {
		t.Errorf("Appointment.Duration = %d, want 60", appt.Duration)
	}
	if appt.Status != "confirmed" {
		t.Errorf("Appointment.Status = %s, want confirmed", appt.Status)
	}
	if appt.ConfirmedAt == nil {
		t.Error("Appointment.ConfirmedAt is nil, want non-nil")
	}
	if len(appt.RemindersSent) != 2 {
		t.Errorf("Appointment.RemindersSent has %d entries, want 2", len(appt.RemindersSent))
	}
	if !appt.RemindersSent["72h"] {
		t.Error("Appointment.RemindersSent[72h] = false, want true")
	}
}

// TestPatientStruct verifies the Patient struct
func TestPatientStruct(t *testing.T) {
	firstVisit := time.Date(2025, 1, 1, 10, 0, 0, 0, ApptTimeZone)
	lastVisit := time.Date(2026, 2, 1, 14, 0, 0, 0, ApptTimeZone)

	patient := Patient{
		TelegramID:       "987654321",
		Name:             "Jane Smith",
		FirstVisit:       firstVisit,
		LastVisit:        lastVisit,
		TotalVisits:      5,
		HealthStatus:     "Good",
		TherapistNotes:   "Prefers deep tissue massage",
		VoiceTranscripts: "Patient mentioned lower back pain",
		CurrentService:   "Deep Tissue Massage",
	}

	if patient.TelegramID != "987654321" {
		t.Errorf("Patient.TelegramID = %s, want 987654321", patient.TelegramID)
	}
	if patient.TotalVisits != 5 {
		t.Errorf("Patient.TotalVisits = %d, want 5", patient.TotalVisits)
	}
	if patient.HealthStatus != "Good" {
		t.Errorf("Patient.HealthStatus = %s, want Good", patient.HealthStatus)
	}
}

// TestAnalyticsEventStruct verifies the AnalyticsEvent struct
func TestAnalyticsEventStruct(t *testing.T) {
	now := time.Now()
	event := AnalyticsEvent{
		ID:        1,
		PatientID: "123456789",
		EventType: "appointment_booked",
		Details: map[string]interface{}{
			"service_id": "massage-60",
			"duration":   60,
		},
		CreatedAt: now,
	}

	if event.ID != 1 {
		t.Errorf("AnalyticsEvent.ID = %d, want 1", event.ID)
	}
	if event.EventType != "appointment_booked" {
		t.Errorf("AnalyticsEvent.EventType = %s, want appointment_booked", event.EventType)
	}
	if len(event.Details) != 2 {
		t.Errorf("AnalyticsEvent.Details has %d entries, want 2", len(event.Details))
	}
}

// BenchmarkSplitSummary benchmarks the SplitSummary function
func BenchmarkSplitSummary(b *testing.B) {
	summary := "Deep Tissue Massage - John Doe Smith"
	for i := 0; i < b.N; i++ {
		SplitSummary(summary)
	}
}
