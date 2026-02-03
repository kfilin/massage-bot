package domain

import (
	"errors"
	"testing"
)

// TestDomainErrors verifies that all domain errors are properly defined
func TestDomainErrors(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		wantText string
	}{
		{
			name:     "ErrInvalidAppointment",
			err:      ErrInvalidAppointment,
			wantText: "invalid appointment details provided",
		},
		{
			name:     "ErrAppointmentInPast",
			err:      ErrAppointmentInPast,
			wantText: "appointment time is in the past",
		},
		{
			name:     "ErrInvalidDuration",
			err:      ErrInvalidDuration,
			wantText: "invalid appointment duration",
		},
		{
			name:     "ErrOutsideWorkingHours",
			err:      ErrOutsideWorkingHours,
			wantText: "appointment time is outside working hours",
		},
		{
			name:     "ErrSlotUnavailable",
			err:      ErrSlotUnavailable,
			wantText: "the chosen time slot is unavailable",
		},
		{
			name:     "ErrServiceNotFound",
			err:      ErrServiceNotFound,
			wantText: "service not found",
		},
		{
			name:     "ErrAppointmentNotFound",
			err:      ErrAppointmentNotFound,
			wantText: "appointment not found",
		},
		{
			name:     "ErrInvalidID",
			err:      ErrInvalidID,
			wantText: "invalid ID provided",
		},
		{
			name:     "ErrCalendarEventNotFound",
			err:      ErrCalendarEventNotFound,
			wantText: "calendar event not found",
		},
		{
			name:     "ErrUserBanned",
			err:      ErrUserBanned,
			wantText: "user is banned",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err == nil {
				t.Fatalf("%s is nil", tt.name)
			}

			if tt.err.Error() != tt.wantText {
				t.Errorf("%s.Error() = %q, want %q", tt.name, tt.err.Error(), tt.wantText)
			}
		})
	}
}

// TestErrorComparison verifies that errors can be compared using errors.Is
func TestErrorComparison(t *testing.T) {
	tests := []struct {
		name      string
		err       error
		target    error
		wantMatch bool
	}{
		{
			name:      "Same error matches",
			err:       ErrAppointmentNotFound,
			target:    ErrAppointmentNotFound,
			wantMatch: true,
		},
		{
			name:      "Different errors don't match",
			err:       ErrAppointmentNotFound,
			target:    ErrServiceNotFound,
			wantMatch: false,
		},
		{
			name:      "Wrapped error matches",
			err:       errors.New("wrapped: " + ErrUserBanned.Error()),
			target:    ErrUserBanned,
			wantMatch: false, // String wrapping doesn't work with errors.Is
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			match := errors.Is(tt.err, tt.target)
			if match != tt.wantMatch {
				t.Errorf("errors.Is(%v, %v) = %v, want %v", tt.err, tt.target, match, tt.wantMatch)
			}
		})
	}
}

// TestErrorUniqueness verifies that all errors are unique
func TestErrorUniqueness(t *testing.T) {
	allErrors := []error{
		ErrInvalidAppointment,
		ErrAppointmentInPast,
		ErrInvalidDuration,
		ErrOutsideWorkingHours,
		ErrSlotUnavailable,
		ErrServiceNotFound,
		ErrAppointmentNotFound,
		ErrInvalidID,
		ErrCalendarEventNotFound,
		ErrUserBanned,
	}

	// Check that all error messages are unique
	seen := make(map[string]bool)
	for _, err := range allErrors {
		msg := err.Error()
		if seen[msg] {
			t.Errorf("Duplicate error message found: %q", msg)
		}
		seen[msg] = true
	}

	if len(seen) != len(allErrors) {
		t.Errorf("Expected %d unique errors, got %d", len(allErrors), len(seen))
	}
}
