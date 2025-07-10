package domain

import "errors"

// Sentinel errors for the appointment domain.
var (
	ErrInvalidAppointmentTime = errors.New("appointment time is in the past or invalid")
	ErrSlotNotAvailable       = errors.New("requested time slot is not available")
	ErrDurationTooShort       = errors.New("appointment duration is invalid")
	ErrOutsideBusinessHours   = errors.New("appointment is outside business hours")
)
