package appointment

import (
	"context"
	"fmt"
	"time"

	"github.com/kfilin/massage-bot/internal/domain"
	"github.com/kfilin/massage-bot/internal/logging"
)

// GetAvailableTimeSlots returns available time slots for a given date and duration.
// This logic was extracted from service.go to reduce complexity.
func (s *Service) GetAvailableTimeSlots(ctx context.Context, date time.Time, durationMinutes int) ([]domain.TimeSlot, error) {
	logging.Debugf("DEBUG: GetAvailableTimeSlots called for date: %s, duration: %d minutes.", date.Format("2006-01-02"), durationMinutes)

	if durationMinutes <= 0 {
		return nil, domain.ErrInvalidDuration
	}

	// Ensure the date is in the correct timezone for working hours logic
	dateInApptTimezone := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, ApptTimeZone)

	// Fetch busy intervals for the entire day
	timeMin := dateInApptTimezone
	timeMax := dateInApptTimezone.Add(24 * time.Hour)

	// Use cached FreeBusy if available (uses logic from service.go)
	busySlots, err := s.getFreeBusy(ctx, timeMin, timeMax)
	if err != nil {
		logging.Errorf("ERROR: Failed to fetch FreeBusy: %v", err)
		return nil, fmt.Errorf("failed to fetch available slots: %w", err)
	}

	var availableSlots []domain.TimeSlot

	// Iterate through the working day in 1-hour steps
	// Note: WorkDayStartHour and WorkDayEndHour are defined in service.go constants
	currentSlotStart := time.Date(dateInApptTimezone.Year(), dateInApptTimezone.Month(), dateInApptTimezone.Day(), domain.WorkDayStartHour, 0, 0, 0, ApptTimeZone)
	workDayEnd := time.Date(dateInApptTimezone.Year(), dateInApptTimezone.Month(), dateInApptTimezone.Day(), domain.WorkDayEndHour, 0, 0, 0, ApptTimeZone)

	// Default step interval - could be configurable
	stepInterval := 60 * time.Minute

	nowInApptTimezone := s.NowFunc().In(ApptTimeZone)

	// Loop until the potential slot end exceeds the workday end
	for currentSlotStart.Add(time.Duration(durationMinutes)*time.Minute).Before(workDayEnd) || currentSlotStart.Add(time.Duration(durationMinutes)*time.Minute).Equal(workDayEnd) {
		currentSlotEnd := currentSlotStart.Add(time.Duration(durationMinutes) * time.Minute)

		// Check if the slot is in the past
		if currentSlotStart.Before(nowInApptTimezone) {
			currentSlotStart = currentSlotStart.Add(stepInterval)
			continue
		}

		isAvailable := true
		for _, busy := range busySlots {
			// Check for overlap: [start, end)
			// Overlap logic: Start < BusyEnd AND End > BusyStart
			if currentSlotStart.Before(busy.End) && currentSlotEnd.After(busy.Start) {
				isAvailable = false
				break
			}
		}

		if isAvailable {
			availableSlots = append(availableSlots, domain.TimeSlot{
				Start: currentSlotStart,
				End:   currentSlotEnd,
			})
		}

		currentSlotStart = currentSlotStart.Add(stepInterval)
	}

	logging.Debugf("DEBUG: GetAvailableTimeSlots finished. Found %d available slots.", len(availableSlots))
	return availableSlots, nil
}
