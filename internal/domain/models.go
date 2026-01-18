package domain

import (
	"log"
	"time"
)

// Service represents a massage service offered.
type Service struct {
	ID              string  `json:"id"` // Unique identifier for the service
	Name            string  `json:"name"`
	DurationMinutes int     `json:"duration_minutes"`
	Price           float64 `json:"price"`
	Description     string  `json:"description,omitempty"`
}

// TimeSlot represents an available time slot for an appointment.
type TimeSlot struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// Appointment represents a booked appointment.
type Appointment struct {
	ID        string    `json:"id"`         // Unique identifier for the appointment (e.g., Google Calendar event ID)
	ServiceID string    `json:"service_id"` // ID of the booked service
	Service   Service   `json:"service"`    // Details of the booked service
	Time      time.Time `json:"time"`       // The primary start time of the appointment (used for initial booking)
	Duration  int       `json:"duration"`   // Duration in minutes

	// Fields derived from Time and Duration, used by calendar adapters
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`

	// Client/Customer related information
	ClientID     string `json:"client_id"`      // Can be the same as ID, or a separate client-specific ID
	ClientName   string `json:"client_name"`    // Full name of the client (from Telegram or input)
	CustomerName string `json:"customer_name"`  // Client's name from Telegram (e.g., FirstName LastName)
	CustomerTgID string `json:"customer_tg_id"` // Telegram User ID

	Notes           string `json:"notes"`               // Any additional notes for the appointment
	CalendarEventID string `json:"calendar_event_id"`   // ID из Google Calendar или другого репозитория
	MeetLink        string `json:"meet_link,omitempty"` // Google Meet link for online consultations
}

// --- Константы и глобальные переменные для временных слотов и рабочего дня ---
const (
	WorkDayStartHour = 9  // 9 AM
	WorkDayEndHour   = 18 // 6 PM
)

var (
	SlotDuration *time.Duration
	ApptTimeZone *time.Location
)

func init() {
	var err error
	// Используем часовой пояс для Турции (Fethiye, Muğla)
	ApptTimeZone, err = time.LoadLocation("Europe/Istanbul")
	if err != nil {
		log.Fatalf("Failed to load timezone 'Europe/Istanbul': %v", err)
	}

	tempDuration := 60 * time.Minute // Длительность слота по умолчанию 60 минут
	SlotDuration = &tempDuration
}

// --- Конец секции констант ---

// Patient represents a patient/client record
type Patient struct {
	TelegramID       string    `json:"telegram_id"`
	Name             string    `json:"name"`
	FirstVisit       time.Time `json:"first_visit"`
	LastVisit        time.Time `json:"last_visit"`
	TotalVisits      int       `json:"total_visits"`
	HealthStatus     string    `json:"health_status"`
	TherapistNotes   string    `json:"therapist_notes,omitempty"`
	VoiceTranscripts string    `json:"voice_transcripts,omitempty"`
	CurrentService   string    `json:"current_service,omitempty"`
}

// SplitSummary splits a calendar event summary into [Service Name, Customer Name]
func SplitSummary(summary string) []string {
	// Standard format: "Service Name - Customer Name"
	// We'll use " - " as the delimiter
	parts := []string{summary}
	for i := len(summary) - 3; i >= 0; i-- {
		if summary[i:i+3] == " - " {
			parts = []string{summary[:i], summary[i+3:]}
			break
		}
	}
	return parts
}
