package domain

import "errors"

// Sentinel errors for the appointment domain.
var (
	ErrInvalidAppointment    = errors.New("invalid appointment details provided")
	ErrAppointmentInPast     = errors.New("appointment time is in the past")           // Renamed from ErrInvalidAppointmentTime and consolidated
	ErrInvalidDuration       = errors.New("invalid appointment duration")              // Renamed from ErrDurationTooShort and consolidated
	ErrOutsideWorkingHours   = errors.New("appointment time is outside working hours") // Renamed from ErrOutsideBusinessHours and consolidated
	ErrSlotUnavailable       = errors.New("the chosen time slot is unavailable")       // Renamed from ErrSlotNotAvailable and consolidated
	ErrServiceNotFound       = errors.New("service not found")
	ErrAppointmentNotFound   = errors.New("appointment not found")
	ErrInvalidID             = errors.New("invalid ID provided")
	ErrCalendarEventNotFound = errors.New("calendar event not found")
	ErrUserBanned            = errors.New("user is banned")
)
