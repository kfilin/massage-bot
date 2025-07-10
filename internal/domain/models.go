package domain

import (
	"fmt"
	"time"
)

// Service represents a massage service offered.
type Service struct {
	ID              string  `json:"id"` // Unique identifier for the service
	Name            string  `json:"name"`
	DurationMinutes int     `json:"duration_minutes"`
	Price           float64 `json:"price"`
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

	Notes string `json:"notes"` // Any additional notes for the appointment
}

// Errors
var (
	ErrAppointmentInPast   = fmt.Errorf("appointment time is in the past")
	ErrInvalidDuration     = fmt.Errorf("invalid appointment duration")
	ErrOutsideWorkingHours = fmt.Errorf("appointment time is outside working hours")
	ErrSlotUnavailable     = fmt.Errorf("the chosen time slot is unavailable")
	ErrServiceNotFound     = fmt.Errorf("service not found")
	ErrAppointmentNotFound = fmt.Errorf("appointment not found")
)
