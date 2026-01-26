package appointment

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"sync"

	"github.com/kfilin/massage-bot/internal/domain" // Import domain package to access its structs and errors
	"github.com/kfilin/massage-bot/internal/monitoring"
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

	tempDuration := 60 * time.Minute // Default slot duration is 60 minutes
	SlotDuration = &tempDuration
}

// Service implements ports.AppointmentService
type Service struct {
	repo ports.AppointmentRepository
	// NowFunc allows injecting a function to get the current time for testing
	NowFunc func() time.Time
	mu      sync.Mutex

	// Cache for FindAll results
	cacheMu      sync.RWMutex
	cachedEvents []domain.Appointment
	cacheExpires time.Time
}

const cacheTTL = 2 * time.Minute

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
		return nil, domain.ErrInvalidDuration
	}

	// Ensure the date is in the correct timezone for working hours logic
	dateInApptTimezone := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, ApptTimeZone)

	// Fetch busy intervals for the entire day
	timeMin := dateInApptTimezone
	timeMax := dateInApptTimezone.Add(24 * time.Hour)

	busySlots, err := s.repo.GetFreeBusy(ctx, timeMin, timeMax)
	if err != nil {
		log.Printf("ERROR: Failed to fetch FreeBusy: %v", err)
		return nil, fmt.Errorf("failed to fetch available slots: %w", err)
	}

	var availableSlots []domain.TimeSlot

	// Iterate through the working day in 1-hour steps
	currentSlotStart := time.Date(dateInApptTimezone.Year(), dateInApptTimezone.Month(), dateInApptTimezone.Day(), WorkDayStartHour, 0, 0, 0, ApptTimeZone)
	workDayEnd := time.Date(dateInApptTimezone.Year(), dateInApptTimezone.Month(), dateInApptTimezone.Day(), WorkDayEndHour, 0, 0, 0, ApptTimeZone)
	stepInterval := 60 * time.Minute

	nowInApptTimezone := s.NowFunc().In(ApptTimeZone)

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

	log.Printf("DEBUG: GetAvailableTimeSlots finished. Found %d available slots.", len(availableSlots))
	return availableSlots, nil
}

// CreateAppointment handles the creation of a new appointment.
func (s *Service) CreateAppointment(ctx context.Context, appt *domain.Appointment) (*domain.Appointment, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

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
	// Fetch busy intervals for the day
	dayStart := time.Date(appt.StartTime.Year(), appt.StartTime.Month(), appt.StartTime.Day(), 0, 0, 0, 0, loc)
	dayEnd := dayStart.Add(24 * time.Hour)

	busySlots, err := s.repo.GetFreeBusy(ctx, dayStart, dayEnd)
	if err != nil {
		log.Printf("ERROR: Failed to fetch FreeBusy for overlapping check: %v", err)
		return nil, fmt.Errorf("failed to verify slot availability: %w", err)
	}

	for _, busy := range busySlots {
		// Check for overlap: [start, end)
		if appt.StartTime.Before(busy.End) && appt.EndTime.After(busy.Start) {
			log.Printf("ERROR: New appointment %s-%s overlaps with busy interval %s-%s",
				appt.StartTime.Format("15:04"), appt.EndTime.Format("15:04"),
				busy.Start.Format("15:04"), busy.End.Format("15:04"))
			return nil, domain.ErrSlotUnavailable
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

	// Record metrics
	leadTimeDays := time.Until(createdAppt.StartTime).Hours() / 24
	if leadTimeDays < 0 {
		leadTimeDays = 0
	}
	monitoring.BookingLeadTimeDays.Observe(leadTimeDays)
	monitoring.ServiceBookingsTotal.WithLabelValues(createdAppt.Service.Name).Inc()
	monitoring.BookingCreationHour.WithLabelValues(fmt.Sprintf("%02d", time.Now().Hour())).Inc()

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

	// Record cancellation metric if we can find the appt info (even from cache/repo)
	// For simplicity, we just increment without service name if we can't find it easily
	monitoring.CancellationsTotal.WithLabelValues("unknown").Inc()

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

// GetCustomerAppointments returns all upcoming appointments (from -24h) for a specific customer.
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

// GetCustomerHistory returns ALL appointments (past and future) for a specific customer.
func (s *Service) GetCustomerHistory(ctx context.Context, customerTgID string) ([]domain.Appointment, error) {
	log.Printf("DEBUG: GetCustomerHistory called for customer TGID: %s", customerTgID)
	if customerTgID == "" {
		return nil, domain.ErrInvalidID
	}

	// Fetch without time limits (nil, nil)
	allAppts, err := s.repo.FindEvents(ctx, nil, nil)
	if err != nil {
		if errors.Is(err, domain.ErrAppointmentNotFound) {
			return []domain.Appointment{}, nil
		}
		return nil, fmt.Errorf("failed to fetch history for customer: %w", err)
	}

	var customerAppts []domain.Appointment
	for _, appt := range allAppts {
		if appt.CustomerTgID == customerTgID {
			customerAppts = append(customerAppts, appt)
		}
	}

	log.Printf("DEBUG: Found %d history events for customer %s", len(customerAppts), customerTgID)
	return customerAppts, nil
}

// GetUpcomingAppointments returns all appointments within a specific time range.
func (s *Service) GetUpcomingAppointments(ctx context.Context, timeMin, timeMax time.Time) ([]domain.Appointment, error) {
	log.Printf("DEBUG: GetUpcomingAppointments called for range %s - %s", timeMin.Format("02.01 15:04"), timeMax.Format("02.01 15:04"))
	return s.repo.FindEvents(ctx, &timeMin, &timeMax)
}

// GetTotalUpcomingCount returns the total number of upcoming appointments for all customers.
func (s *Service) GetTotalUpcomingCount(ctx context.Context) (int, error) {
	allAppts, err := s.repo.FindAll(ctx)
	if err != nil {
		if errors.Is(err, domain.ErrAppointmentNotFound) {
			return 0, nil
		}
		return 0, fmt.Errorf("failed to fetch all appointments: %w", err)
	}
	return len(allAppts), nil
}

// GetCalendarAccountInfo returns the email address or summary of the connected Google Calendar.
func (s *Service) GetCalendarAccountInfo(ctx context.Context) (string, error) {
	return s.repo.GetAccountInfo(ctx)
}

func (s *Service) GetCalendarID() string {
	return s.repo.GetCalendarID()
}

func (s *Service) ListCalendars(ctx context.Context) ([]string, error) {
	return s.repo.ListCalendars(ctx)
}
