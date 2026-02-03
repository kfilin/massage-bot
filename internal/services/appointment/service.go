package appointment

import (
	"context"
	"errors"
	"fmt"
	"time"

	"sync"

	"github.com/kfilin/massage-bot/internal/domain" // Import domain package to access its structs and errors
	"github.com/kfilin/massage-bot/internal/logging"
	"github.com/kfilin/massage-bot/internal/ports"
)

var (
	SlotDuration *time.Duration // Duration of each booking slot (e.g., 30 minutes)
)

func init() {
	tempDuration := 60 * time.Minute // Default slot duration is 60 minutes
	SlotDuration = &tempDuration
}

// Service implements ports.AppointmentService
type Service struct {
	repo ports.AppointmentRepository
	// NowFunc allows injecting a function to get the current time for testing
	NowFunc func() time.Time
	mu      sync.Mutex

	// Cache for FreeBusy results
	fbCacheMu sync.RWMutex
	fbCache   map[string]freeBusyEntry

	metrics MetricsCollector
}

type freeBusyEntry struct {
	slots     []domain.TimeSlot
	expiresAt time.Time
}

const cacheTTL = 2 * time.Minute

// NewService creates a new appointment service with default dependencies.
func NewService(repo ports.AppointmentRepository) *Service {
	return &Service{
		repo:    repo,
		NowFunc: time.Now, // Default to standard time.Now()
		fbCache: make(map[string]freeBusyEntry),
		metrics: NewPrometheusCollector(), // Default to Prometheus
	}
}

// NewServiceWithMetrics creates a new appointment service with a custom metrics collector.
func NewServiceWithMetrics(repo ports.AppointmentRepository, metrics MetricsCollector) *Service {
	return &Service{
		repo:    repo,
		NowFunc: time.Now,
		fbCache: make(map[string]freeBusyEntry),
		metrics: metrics,
	}
}

// getFreeBusy retrieves busy slots from cache or repository
func (s *Service) getFreeBusy(ctx context.Context, timeMin, timeMax time.Time) ([]domain.TimeSlot, error) {
	// Create a unique cache key based on the time range
	// Since we typically query for full days, Format("2006-01-02") is sufficient if timeMin is start of day
	// But to be safe for arbitrary ranges, we can use a more precise key
	key := fmt.Sprintf("%s-%s", timeMin.Format(time.RFC3339), timeMax.Format(time.RFC3339))

	s.fbCacheMu.RLock()
	entry, found := s.fbCache[key]
	s.fbCacheMu.RUnlock()

	if found && time.Now().Before(entry.expiresAt) {
		s.metrics.RecordFreeBusyCacheHit()
		logging.Debugf("DEBUG: FreeBusy cache HIT for %s", key)
		return entry.slots, nil
	}

	s.metrics.RecordFreeBusyCacheMiss()
	logging.Debugf("DEBUG: FreeBusy cache MISS for %s", key)

	// Fetch from repo
	slots, err := s.repo.GetFreeBusy(ctx, timeMin, timeMax)
	if err != nil {
		return nil, err
	}

	// Cache the result
	s.fbCacheMu.Lock()
	s.fbCache[key] = freeBusyEntry{
		slots:     slots,
		expiresAt: time.Now().Add(cacheTTL),
	}
	s.fbCacheMu.Unlock()

	return slots, nil
}

// invalidateCache clears the FreeBusy cache.
// Should be called when appointments are created or cancelled.
func (s *Service) invalidateCache() {
	s.fbCacheMu.Lock()
	s.fbCache = make(map[string]freeBusyEntry)
	s.fbCacheMu.Unlock()
	logging.Debug("DEBUG: FreeBusy cache invalidated.")
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
	logging.Debugf("DEBUG: GetAvailableServices returned %d services.", len(services))
	return services, nil
}

// CreateAppointment handles the creation of a new appointment.
func (s *Service) CreateAppointment(ctx context.Context, appt *domain.Appointment) (*domain.Appointment, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if appt == nil {
		logging.Error("ERROR: CreateAppointment - Appointment is nil")
		return nil, domain.ErrInvalidAppointment
	}

	logging.Debugf("DEBUG: CreateAppointment called for service '%s' at %s", appt.Service.Name, appt.StartTime.Format("2006-01-02 15:04"))

	if appt.Service.ID == "" || appt.StartTime.IsZero() || appt.Duration <= 0 || appt.CustomerName == "" {
		logging.Errorf("ERROR: CreateAppointment - Invalid appointment details: %+v", appt)
		return nil, domain.ErrInvalidAppointment
	}

	// Ensure times are in the correct timezone for validation
	loc := domain.ApptTimeZone
	if loc == nil {
		logging.Warn("WARNING: ApptTimeZone is nil during appointment creation validation, defaulting to Local time.")
		loc = time.Local
	}

	appt.StartTime = appt.StartTime.In(loc)
	appt.EndTime = appt.StartTime.Add(time.Duration(appt.Duration) * time.Minute)

	// 1. Validate against past time
	nowInLoc := s.NowFunc().In(loc)
	if appt.StartTime.Before(nowInLoc) {
		logging.Errorf("ERROR: Appointment time %s is in the past (now: %s)", appt.StartTime.Format("15:04"), nowInLoc.Format("15:04"))
		return nil, domain.ErrAppointmentInPast
	}

	// 2. Validate against working hours
	startHour := appt.StartTime.Hour()
	endHour := appt.EndTime.Hour()
	endMinute := appt.EndTime.Minute()

	// If end time is exactly on the hour (e.g., 18:00 for a 17:00-18:00 appointment), it's still within.
	// If it's past the hour (e.g., 18:01), it's outside.
	if startHour < domain.WorkDayStartHour || startHour >= domain.WorkDayEndHour ||
		(endHour > domain.WorkDayEndHour || (endHour == domain.WorkDayEndHour && endMinute > 0)) {
		logging.Errorf("ERROR: Appointment time %s-%s is outside working hours %d:00-%d:00",
			appt.StartTime.Format("15:04"), appt.EndTime.Format("15:04"), domain.WorkDayStartHour, domain.WorkDayEndHour)
		return nil, domain.ErrOutsideWorkingHours
	}

	// 3. Check for slot availability (re-check to prevent race conditions or double bookings)
	// Fetch busy intervals for the day
	dayStart := time.Date(appt.StartTime.Year(), appt.StartTime.Month(), appt.StartTime.Day(), 0, 0, 0, 0, loc)
	dayEnd := dayStart.Add(24 * time.Hour)

	busySlots, err := s.getFreeBusy(ctx, dayStart, dayEnd)
	if err != nil {
		logging.Errorf("ERROR: Failed to fetch FreeBusy for overlapping check: %v", err)
		return nil, fmt.Errorf("failed to verify slot availability: %w", err)
	}

	for _, busy := range busySlots {
		// Check for overlap: [start, end)
		if appt.StartTime.Before(busy.End) && appt.EndTime.After(busy.Start) {
			logging.Errorf("ERROR: New appointment %s-%s overlaps with busy interval %s-%s",
				appt.StartTime.Format("15:04"), appt.EndTime.Format("15:04"),
				busy.Start.Format("15:04"), busy.End.Format("15:04"))
			return nil, domain.ErrSlotUnavailable
		}
	}
	logging.Debug("DEBUG: Appointment slot is available.")

	// 4. Persist the appointment (e.g., in Google Calendar)
	createdAppt, err := s.repo.Create(ctx, appt)
	if err != nil {
		logging.Errorf("ERROR: Failed to create appointment in repository: %v", err)
		return nil, fmt.Errorf("failed to create appointment in repository: %w", err)
	}
	logging.Debugf("DEBUG: Appointment successfully created in repository with ID: %s", createdAppt.ID)

	// Record metrics
	leadTimeDays := time.Until(createdAppt.StartTime).Hours() / 24
	if leadTimeDays < 0 {
		leadTimeDays = 0
	}
	s.metrics.RecordAppointmentCreated(createdAppt.Service.Name, leadTimeDays)

	// Invalidate cache to prevent stale availability
	s.invalidateCache()

	return createdAppt, nil
}

// CancelAppointment cancels an appointment by ID.
func (s *Service) CancelAppointment(ctx context.Context, appointmentID string) error {
	logging.Debugf("DEBUG: CancelAppointment called for ID: %s", appointmentID)
	if appointmentID == "" {
		return domain.ErrInvalidID
	}

	err := s.repo.Delete(ctx, appointmentID)
	if err != nil {
		if errors.Is(err, domain.ErrAppointmentNotFound) {
			logging.Warnf("WARNING: Attempted to cancel non-existent appointment ID: %s", appointmentID)
			return err
		}
		logging.Errorf("ERROR: Failed to delete appointment %s in repository: %v", appointmentID, err)
		return fmt.Errorf("failed to delete appointment in repository: %w", err)
	}
	logging.Debugf("DEBUG: Appointment %s successfully cancelled.", appointmentID)

	// Record cancellation metric
	s.metrics.RecordAppointmentCancelled()

	// Invalidate cache as a slot just freed up
	s.invalidateCache()

	return nil
}

// FindByID retrieves an appointment by its ID.
func (s *Service) FindByID(ctx context.Context, id string) (*domain.Appointment, error) {
	logging.Debugf("DEBUG: FindByID called for ID: %s", id)
	if id == "" {
		return nil, domain.ErrInvalidID
	}
	appt, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, domain.ErrAppointmentNotFound) {
			logging.Warnf("WARNING: Appointment with ID %s not found.", id)
			return nil, err
		}
		logging.Errorf("ERROR: Failed to find appointment %s in repository: %v", id, err)
		return nil, fmt.Errorf("failed to find appointment in repository: %w", err)
	}
	logging.Debugf("DEBUG: Found appointment with ID %s.", id)
	return appt, nil
}

// GetCustomerAppointments returns all upcoming appointments (from -24h) for a specific customer.
func (s *Service) GetCustomerAppointments(ctx context.Context, customerTgID string) ([]domain.Appointment, error) {
	logging.Debugf("DEBUG: GetCustomerAppointments called for customer TGID: %s", customerTgID)
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

	logging.Debugf("DEBUG: Found %d appointments for customer %s", len(customerAppts), customerTgID)
	return customerAppts, nil
}

// GetAllUpcomingAppointments returns all upcoming appointments (from -24h) for ALL customers.
func (s *Service) GetAllUpcomingAppointments(ctx context.Context) ([]domain.Appointment, error) {
	logging.Debug("DEBUG: GetAllUpcomingAppointments called")

	allAppts, err := s.repo.FindAll(ctx)
	if err != nil {
		if errors.Is(err, domain.ErrAppointmentNotFound) {
			return []domain.Appointment{}, nil
		}
		return nil, fmt.Errorf("failed to fetch all appointments: %w", err)
	}

	now := time.Now().In(domain.ApptTimeZone)
	cutoff := now.Add(-24 * time.Hour)

	var upcomingAppts []domain.Appointment
	for _, appt := range allAppts {
		// Filter: Not cancelled AND (start time in future OR within last 24h)
		if appt.Status != "cancelled" && appt.StartTime.After(cutoff) {
			upcomingAppts = append(upcomingAppts, appt)
		}
	}

	logging.Debugf("DEBUG: Found %d upcoming appointments in total", len(upcomingAppts))
	return upcomingAppts, nil
}

// GetCustomerHistory returns ALL appointments (past and future) for a specific customer.
func (s *Service) GetCustomerHistory(ctx context.Context, customerTgID string) ([]domain.Appointment, error) {
	logging.Debugf("DEBUG: GetCustomerHistory called for customer TGID: %s", customerTgID)
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

	logging.Debugf("DEBUG: Found %d history events for customer %s", len(customerAppts), customerTgID)
	return customerAppts, nil
}

// GetUpcomingAppointments returns all appointments within a specific time range.
func (s *Service) GetUpcomingAppointments(ctx context.Context, timeMin, timeMax time.Time) ([]domain.Appointment, error) {
	logging.Debugf("DEBUG: GetUpcomingAppointments called for range %s - %s", timeMin.Format("02.01 15:04"), timeMax.Format("02.01 15:04"))
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
