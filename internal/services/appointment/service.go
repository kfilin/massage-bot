package appointment

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/kfilin/massage-bot/internal/domain"
	"github.com/kfilin/massage-bot/internal/ports"
)

const (
	// Working hours
	WorkStartHour = 9
	WorkEndHour   = 18
	// Slot duration (default if not specified by service)
	DefaultSlotCheckInterval = 15 * time.Minute // Check for slots every 15 minutes
)

type Service struct {
	repo ports.AppointmentRepository // Dependency on the repository interface
}

func NewService(repo ports.AppointmentRepository) *Service {
	return &Service{repo: repo}
}

// GetAvailableServices returns a predefined list of services.
// In a real application, this would typically fetch data from a database.
func (s *Service) GetAvailableServices(ctx context.Context) ([]domain.Service, error) {
	return []domain.Service{
		{ID: "svc_classic", Name: "Классический массаж", DurationMinutes: 60, Price: 2500.00},
		{ID: "svc_relax", Name: "Расслабляющий массаж", DurationMinutes: 90, Price: 3500.00},
		{ID: "svc_sport", Name: "Спортивный массаж", DurationMinutes: 45, Price: 2000.00},
	}, nil
}

// GetAvailableTimeSlots finds available slots for a given date and service duration.
func (s *Service) GetAvailableTimeSlots(ctx context.Context, date time.Time, durationMinutes int) ([]domain.TimeSlot, error) {
	if durationMinutes <= 0 {
		return nil, domain.ErrInvalidDuration
	}
	slotDuration := time.Duration(durationMinutes) * time.Minute

	dayStart := time.Date(date.Year(), date.Month(), date.Day(), WorkStartHour, 0, 0, 0, date.Location())
	dayEnd := time.Date(date.Year(), date.Month(), date.Day(), WorkEndHour, 0, 0, 0, date.Location())

	// Fetch existing appointments for the selected day from the repository
	existingAppointments, err := s.repo.FindAll(ctx) // This should fetch events for a relevant period (e.g., next 30 days)
	if err != nil {
		log.Printf("Warning: Failed to fetch existing appointments for overlap check: %v. Proceeding assuming no overlaps.", err)
		existingAppointments = []domain.Appointment{} // Fallback to empty if repo fails
	}

	var availableSlots []domain.TimeSlot
	for currentSlotStart := dayStart; currentSlotStart.Add(slotDuration).Before(dayEnd.Add(1 * time.Minute)); currentSlotStart = currentSlotStart.Add(DefaultSlotCheckInterval) {
		currentSlotEnd := currentSlotStart.Add(slotDuration)

		// Ensure the entire slot ends within working hours
		if currentSlotEnd.After(dayEnd.Add(1 * time.Minute)) { // Allow slots ending exactly at WorkEndHour
			continue
		}

		// Skip if the slot is in the past
		if currentSlotStart.Before(time.Now()) {
			continue
		}

		isOverlap := false
		for _, appt := range existingAppointments {
			// Check if appt is for the same day
			if appt.StartTime.Year() == date.Year() && appt.StartTime.Month() == date.Month() && appt.StartTime.Day() == date.Day() {
				// Overlap condition: (StartA < EndB) AND (EndA > StartB)
				if currentSlotStart.Before(appt.EndTime) && currentSlotEnd.After(appt.StartTime) {
					isOverlap = true
					break
				}
			}
		}

		if !isOverlap {
			availableSlots = append(availableSlots, domain.TimeSlot{Start: currentSlotStart, End: currentSlotEnd})
		}
	}

	return availableSlots, nil
}

// CreateAppointment validates and then creates a new appointment using the repository.
func (s *Service) CreateAppointment(ctx context.Context, appointment *domain.Appointment) (*domain.Appointment, error) {
	if appointment == nil || appointment.Service.ID == "" || appointment.Duration <= 0 || appointment.Time.IsZero() {
		return nil, fmt.Errorf("invalid appointment data")
	}

	// Set StartTime and EndTime based on Time and Duration for internal consistency
	appointment.StartTime = appointment.Time
	appointment.EndTime = appointment.Time.Add(time.Duration(appointment.Duration) * time.Minute)

	// Basic validation
	if appointment.StartTime.Before(time.Now()) {
		return nil, domain.ErrAppointmentInPast
	}

	// Check if the slot is within working hours
	dayStart := time.Date(appointment.StartTime.Year(), appointment.StartTime.Month(), appointment.StartTime.Day(), WorkStartHour, 0, 0, 0, appointment.StartTime.Location())
	dayEnd := time.Date(appointment.StartTime.Year(), appointment.StartTime.Month(), appointment.StartTime.Day(), WorkEndHour, 0, 0, 0, appointment.StartTime.Location())

	if appointment.StartTime.Before(dayStart) || appointment.EndTime.After(dayEnd) {
		return nil, domain.ErrOutsideWorkingHours
	}

	// Check for overlaps with existing appointments before creating
	existingAppointments, err := s.repo.FindAll(ctx) // Re-fetch to ensure no concurrent bookings
	if err != nil {
		log.Printf("Warning: Could not fetch existing appointments for overlap check during creation: %v", err)
		// Decide if you want to proceed without overlap check or return an error.
		// For robustness, returning an error or retrying would be better.
	} else {
		for _, appt := range existingAppointments {
			// Check if appt is for the same day and overlaps
			if appt.StartTime.Year() == appointment.StartTime.Year() &&
				appt.StartTime.Month() == appointment.StartTime.Month() &&
				appt.StartTime.Day() == appointment.StartTime.Day() &&
				appointment.StartTime.Before(appt.EndTime) && appointment.EndTime.After(appt.StartTime) {
				return nil, domain.ErrSlotUnavailable
			}
		}
	}

	// Delegate to the repository to actually create the appointment (e.g., in Google Calendar)
	createdApp, err := s.repo.Create(ctx, appointment)
	if err != nil {
		return nil, fmt.Errorf("failed to create appointment in repository: %w", err)
	}

	return createdApp, nil
}

// CancelAppointment cancels an existing appointment by its ID using the repository.
func (s *Service) CancelAppointment(ctx context.Context, appointmentID string) error {
	if appointmentID == "" {
		return fmt.Errorf("appointment ID cannot be empty")
	}
	err := s.repo.Delete(ctx, appointmentID)
	if err != nil {
		return fmt.Errorf("failed to cancel appointment in repository: %w", err)
	}
	return nil
}
