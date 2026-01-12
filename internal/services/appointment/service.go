package appointment

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/kfilin/massage-bot/internal/domain" // Import domain package to access its structs and errors
	"github.com/kfilin/massage-bot/internal/ports"
)

// Global constants for working hours and slot duration
const (
	WorkDayStartHour = 9  // 9 AM
	WorkDayEndHour   = 18 // 6 PM
)

var (
	SlotDuration *time.Duration // Duration of each booking slot (e.g., 30 minutes)
	ApptTimeZone *time.Location
	Err          error // This general Err variable might be leftover, consider if it's still needed.
)

func init() {
	var err error
	ApptTimeZone, err = time.LoadLocation("Europe/Istanbul")
	if err != nil {
		log.Fatalf("Failed to load timezone 'Europe/Istanbul': %v", err)
	}

	tempDuration := 60 * time.Minute // Default slot duration is now 60 minutes
	SlotDuration = &tempDuration
}

// Service implements ports.AppointmentService
type Service struct {
	repo ports.AppointmentRepository
	// NowFunc allows injecting a function to get the current time for testing
	NowFunc func() time.Time
}

// NewService creates a new appointment service
func NewService(repo ports.AppointmentRepository) *Service {
	return &Service{
		repo:    repo,
		NowFunc: time.Now, // Default to standard time.Now()
	}
}

// GetAvailableServices returns a predefined list of services.
func (s *Service) GetAvailableServices(ctx context.Context) ([]domain.Service, error) {
	services := []domain.Service{
		{
			ID:              "1",
			Name:            "Массаж Спина + Шея",
			DurationMinutes: 40,
			Price:           2000.00,
		},
		{
			ID:              "2",
			Name:            "Общий массаж",
			DurationMinutes: 60,
			Price:           2800.00,
		},
		{
			ID:              "3",
			Name:            "Лимфодренаж",
			DurationMinutes: 50,
			Price:           2400.00,
		},
		{
			ID:              "4",
			Name:            "Иглоукалывание",
			DurationMinutes: 30,
			Price:           1400.00,
		},
		{
			ID:              "5",
			Name:            "Консультация офлайн",
			DurationMinutes: 60,
			Price:           2000.00,
		},
		{
			ID:              "6",
			Name:            "Консультация онлайн",
			DurationMinutes: 45,
			Price:           1500.00,
		},
		{
			ID:              "7",
			Name:            "Реабилитационные программы",
			DurationMinutes: 60,
			Price:           13000.00,
			Description:     "от 13000 ₺ в месяц",
		},
	}
	log.Printf("DEBUG: GetAvailableServices returned %d services.", len(services))
	return services, nil
}

// GetAvailableTimeSlots returns available time slots for a given date and duration.
func (s *Service) GetAvailableTimeSlots(ctx context.Context, date time.Time, durationMinutes int) ([]domain.TimeSlot, error) {
	log.Printf("DEBUG: GetAvailableTimeSlots called for date: %s, duration: %d minutes.", date.Format("2006-01-02"), durationMinutes)

	if durationMinutes <= 0 {
		log.Printf("ERROR: GetAvailableTimeSlots received invalid duration: %d", durationMinutes)
		return nil, domain.ErrInvalidDuration
	}

	// Ensure the date is in the correct timezone for working hours logic
	dateInApptTimezone := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, ApptTimeZone)

	var availableSlots []domain.TimeSlot

	// Iterate through the working day in SlotDuration increments
	// Start from WorkDayStartHour in the specified timezone
	currentSlotStart := time.Date(dateInApptTimezone.Year(), dateInApptTimezone.Month(), dateInApptTimezone.Day(), WorkDayStartHour, 0, 0, 0, ApptTimeZone)
	workDayEnd := time.Date(dateInApptTimezone.Year(), dateInApptTimezone.Month(), dateInApptTimezone.Day(), WorkDayEndHour, 0, 0, 0, ApptTimeZone)

	// Fetch all existing appointments for the given date
	// This assumes FindAll can filter by date, or you fetch all and filter in-memory.
	// For simplicity, let's assume FindAll returns all events and we filter them.
	// In a real app, you'd want a more efficient query to the repository.
	existingAppointments, err := s.repo.FindAll(ctx) // Fetch all events
	if err != nil {
		log.Printf("ERROR: Failed to fetch existing appointments: %v", err)
		// If it's a "not found" error, treat it as no existing appointments.
		if !errors.Is(err, domain.ErrAppointmentNotFound) {
			return nil, fmt.Errorf("failed to fetch existing appointments: %w", err)
		}
		existingAppointments = []domain.Appointment{} // Initialize as empty slice if not found
	}

	// Filter existing appointments for the specific date
	var appointmentsOnSelectedDate []domain.Appointment
	for _, appt := range existingAppointments {
		// Compare dates (ignoring time) in the same timezone
		if appt.StartTime.In(ApptTimeZone).Year() == dateInApptTimezone.Year() &&
			appt.StartTime.In(ApptTimeZone).Month() == dateInApptTimezone.Month() &&
			appt.StartTime.In(ApptTimeZone).Day() == dateInApptTimezone.Day() {
			appointmentsOnSelectedDate = append(appointmentsOnSelectedDate, appt)
		}
	}
	log.Printf("DEBUG: Found %d existing appointments on %s.", len(appointmentsOnSelectedDate), dateInApptTimezone.Format("2006-01-02"))

	// Iterate through potential slots
	for currentSlotStart.Add(time.Duration(durationMinutes)*time.Minute).Before(workDayEnd) || currentSlotStart.Add(time.Duration(durationMinutes)*time.Minute).Equal(workDayEnd) {
		currentSlotEnd := currentSlotStart.Add(time.Duration(durationMinutes) * time.Minute)

		// Check if the slot is in the past (using NowFunc for testability)
		nowInApptTimezone := s.NowFunc().In(ApptTimeZone)
		if currentSlotStart.Before(nowInApptTimezone) {
			log.Printf("DEBUG: Slot %s-%s is in the past, skipping.", currentSlotStart.Format("15:04"), currentSlotEnd.Format("15:04"))
			currentSlotStart = currentSlotEnd // Move to the next potential slot
			continue
		}

		isAvailable := true
		for _, existingAppt := range appointmentsOnSelectedDate {
			// Check for overlap
			// Slot starts before existing ends AND Slot ends after existing starts
			if currentSlotStart.Before(existingAppt.EndTime) && currentSlotEnd.After(existingAppt.StartTime) {
				log.Printf("DEBUG: Slot %s-%s overlaps with existing appointment %s-%s, skipping.",
					currentSlotStart.Format("15:04"), currentSlotEnd.Format("15:04"),
					existingAppt.StartTime.Format("15:04"), existingAppt.EndTime.Format("15:04"))
				isAvailable = false
				break
			}
		}

		if isAvailable {
			availableSlots = append(availableSlots, domain.TimeSlot{
				Start: currentSlotStart,
				End:   currentSlotEnd,
			})
			log.Printf("DEBUG: Added available slot: %s-%s", currentSlotStart.Format("15:04"), currentSlotEnd.Format("15:04"))
		}

		// Move to the next potential slot, which is the end of the current slot
		currentSlotStart = currentSlotEnd
	}

	log.Printf("DEBUG: GetAvailableTimeSlots finished. Found %d available slots.", len(availableSlots))
	return availableSlots, nil
}

// CreateAppointment handles the creation of a new appointment.
func (s *Service) CreateAppointment(ctx context.Context, appt *domain.Appointment) (*domain.Appointment, error) {
	log.Printf("DEBUG: CreateAppointment called for service '%s' at %s", appt.Service.Name, appt.StartTime.Format("2006-01-02 15:04"))

	if appt == nil || appt.Service.ID == "" || appt.StartTime.IsZero() || appt.Duration <= 0 || appt.CustomerName == "" {
		log.Printf("ERROR: CreateAppointment - Invalid appointment details: %+v", appt)
		return nil, domain.ErrInvalidAppointment
	}

	// Ensure times are in the correct timezone for validation
	loc := ApptTimeZone
	if loc == nil {
		log.Println("WARNING: ApptTimeZone is nil during appointment creation validation, defaulting to Local time.")
		loc = time.Local
	}

	appt.StartTime = appt.StartTime.In(loc)
	appt.EndTime = appt.StartTime.Add(time.Duration(appt.Duration) * time.Minute)

	// 1. Validate against past time
	nowInLoc := s.NowFunc().In(loc)
	if appt.StartTime.Before(nowInLoc) {
		log.Printf("ERROR: Appointment time %s is in the past (now: %s)", appt.StartTime.Format("15:04"), nowInLoc.Format("15:04"))
		return nil, domain.ErrAppointmentInPast
	}

	// 2. Validate against working hours
	startHour := appt.StartTime.Hour()
	endHour := appt.EndTime.Hour()
	endMinute := appt.EndTime.Minute()

	// If end time is exactly on the hour (e.g., 18:00 for a 17:00-18:00 appointment), it's still within.
	// If it's past the hour (e.g., 18:01), it's outside.
	if startHour < WorkDayStartHour || startHour >= WorkDayEndHour ||
		(endHour > WorkDayEndHour || (endHour == WorkDayEndHour && endMinute > 0)) {
		log.Printf("ERROR: Appointment time %s-%s is outside working hours %d:00-%d:00",
			appt.StartTime.Format("15:04"), appt.EndTime.Format("15:04"), WorkDayStartHour, WorkDayEndHour)
		return nil, domain.ErrOutsideWorkingHours
	}

	// 3. Check for slot availability (re-check to prevent race conditions or double bookings)
	// Fetch all existing appointments for the specific date
	existingAppointments, err := s.repo.FindAll(ctx) // Fetch all events
	if err != nil {
		log.Printf("ERROR: Failed to fetch existing appointments for availability check: %v", err)
		if !errors.Is(err, domain.ErrAppointmentNotFound) {
			return nil, fmt.Errorf("failed to fetch existing appointments for availability check: %w", err)
		}
		existingAppointments = []domain.Appointment{}
	}

	for _, existingAppt := range existingAppointments {
		// Compare dates (ignoring time) in the same timezone
		if existingAppt.StartTime.In(loc).Year() == appt.StartTime.Year() &&
			existingAppt.StartTime.In(loc).Month() == appt.StartTime.Month() &&
			existingAppt.StartTime.In(loc).Day() == appt.StartTime.Day() {
			// Check for overlap with the new appointment
			if appt.StartTime.Before(existingAppt.EndTime) && appt.EndTime.After(existingAppt.StartTime) {
				log.Printf("ERROR: New appointment %s-%s overlaps with existing appointment %s-%s",
					appt.StartTime.Format("15:04"), appt.EndTime.Format("15:04"),
					existingAppt.StartTime.Format("15:04"), existingAppt.EndTime.Format("15:04"))
				return nil, domain.ErrSlotUnavailable
			}
		}
	}
	log.Printf("DEBUG: Appointment slot is available.")

	// 4. Persist the appointment (e.g., in Google Calendar)
	createdAppt, err := s.repo.Create(ctx, appt)
	if err != nil {
		log.Printf("ERROR: Failed to create appointment in repository: %v", err)
		return nil, fmt.Errorf("failed to create appointment in repository: %w", err)
	}
	log.Printf("DEBUG: Appointment successfully created in repository with ID: %s", createdAppt.ID)

	return createdAppt, nil
}

// CancelAppointment cancels an appointment by ID.
func (s *Service) CancelAppointment(ctx context.Context, appointmentID string) error {
	log.Printf("DEBUG: CancelAppointment called for ID: %s", appointmentID)
	if appointmentID == "" {
		return domain.ErrInvalidID
	}

	err := s.repo.Delete(ctx, appointmentID)
	if err != nil {
		if errors.Is(err, domain.ErrAppointmentNotFound) {
			log.Printf("WARNING: Attempted to cancel non-existent appointment ID: %s", appointmentID)
			return err
		}
		log.Printf("ERROR: Failed to delete appointment %s in repository: %v", appointmentID, err)
		return fmt.Errorf("failed to delete appointment in repository: %w", err)
	}
	log.Printf("DEBUG: Appointment %s successfully cancelled.", appointmentID)
	return nil
}

// FindByID retrieves an appointment by its ID.
func (s *Service) FindByID(ctx context.Context, id string) (*domain.Appointment, error) {
	log.Printf("DEBUG: FindByID called for ID: %s", id)
	if id == "" {
		return nil, domain.ErrInvalidID
	}
	appt, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, domain.ErrAppointmentNotFound) {
			log.Printf("WARNING: Appointment with ID %s not found.", id)
			return nil, err
		}
		log.Printf("ERROR: Failed to find appointment %s in repository: %v", id, err)
		return nil, fmt.Errorf("failed to find appointment in repository: %w", err)
	}
	log.Printf("DEBUG: Found appointment with ID %s.", id)
	return appt, nil
}

// GetCustomerAppointments returns all upcoming appointments for a specific customer.
func (s *Service) GetCustomerAppointments(ctx context.Context, customerTgID string) ([]domain.Appointment, error) {
	log.Printf("DEBUG: GetCustomerAppointments called for customer TGID: %s", customerTgID)
	if customerTgID == "" {
		return nil, domain.ErrInvalidID
	}

	allAppts, err := s.repo.FindAll(ctx)
	if err != nil {
		if errors.Is(err, domain.ErrAppointmentNotFound) {
			return []domain.Appointment{}, nil
		}
		return nil, fmt.Errorf("failed to fetch appointments for customer: %w", err)
	}

	var customerAppts []domain.Appointment
	for _, appt := range allAppts {
		if appt.CustomerTgID == customerTgID {
			customerAppts = append(customerAppts, appt)
		}
	}

	log.Printf("DEBUG: Found %d appointments for customer %s", len(customerAppts), customerTgID)
	return customerAppts, nil
}
