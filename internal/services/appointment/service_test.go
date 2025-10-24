package appointment_test

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/kfilin/massage-bot/internal/domain"
	apnt_svc "github.com/kfilin/massage-bot/internal/services/appointment" // Alias for constants
)

// --- Mocks and Helpers for Testing ---

// MockAppointmentRepository implements ports.AppointmentRepository for testing purposes.
type MockAppointmentRepository struct {
	Appointments  []domain.Appointment
	CreateFunc    func(ctx context.Context, appt *domain.Appointment) (*domain.Appointment, error)
	FindAllFunc   func(ctx context.Context) ([]domain.Appointment, error)
	FindByIDFunc  func(ctx context.Context, id string) (*domain.Appointment, error)
	DeleteFunc    func(ctx context.Context, id string) error
	CreateError   error
	FindAllError  error
	FindByIDError error
	DeleteError   error
}

func (m *MockAppointmentRepository) Create(ctx context.Context, appt *domain.Appointment) (*domain.Appointment, error) {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, appt)
	}
	if m.CreateError != nil {
		return nil, m.CreateError
	}
	// Simulate ID generation
	appt.ID = fmt.Sprintf("test-appt-%d", len(m.Appointments)+1)
	m.Appointments = append(m.Appointments, *appt)
	return appt, nil
}

func (m *MockAppointmentRepository) FindAll(ctx context.Context) ([]domain.Appointment, error) {
	if m.FindAllFunc != nil {
		return m.FindAllFunc(ctx)
	}
	return m.Appointments, m.FindAllError
}

func (m *MockAppointmentRepository) FindByID(ctx context.Context, id string) (*domain.Appointment, error) {
	if m.FindByIDFunc != nil {
		return m.FindByIDFunc(ctx, id)
	}
	if m.FindByIDError != nil {
		return nil, m.FindByIDError
	}
	for _, appt := range m.Appointments {
		if appt.ID == id {
			return &appt, nil
		}
	}
	return nil, domain.ErrAppointmentNotFound
}

func (m *MockAppointmentRepository) Delete(ctx context.Context, id string) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, id)
	}
	return m.DeleteError
}

// --- Test-specific Constants and Helpers ---
// This constant is used only in this test file to define a test service's duration.
const testServiceDurationMinutes = 60 // Example service duration for tests

// --- Tests for Service ---

func TestNewService(t *testing.T) {
	repo := &MockAppointmentRepository{}
	service := apnt_svc.NewService(repo)

	if service == nil {
		t.Errorf("NewService returned nil")
	}
	// You might want to check if the internal repository is correctly set if it's exported
	// (it's not, which is good encapsulation)
}

func TestGetAvailableServices(t *testing.T) {
	repo := &MockAppointmentRepository{}
	service := apnt_svc.NewService(repo)

	expectedServices := []domain.Service{
		{ID: "svc_classic", Name: "Классический массаж", DurationMinutes: 60, Price: 2500.00},
		{ID: "svc_relax", Name: "Расслабляющий массаж", DurationMinutes: 90, Price: 3500.00},
		{ID: "svc_sport", Name: "Спортивный массаж", DurationMinutes: 45, Price: 2000.00},
	}

	services, err := service.GetAvailableServices(context.Background())
	if err != nil {
		t.Fatalf("GetAvailableServices returned an error: %v", err)
	}

	if !reflect.DeepEqual(services, expectedServices) {
		t.Errorf("GetAvailableServices returned %+v, want %+v", services, expectedServices)
	}
}

func TestGetAvailableTimeSlots(t *testing.T) {
	ctx := context.Background()
	loc, _ := time.LoadLocation("Europe/Moscow") // Use a specific location for deterministic tests

	tests := []struct {
		name             string
		date             time.Time
		durationMinutes  int
		existingAppts    []domain.Appointment
		mockFindAllError error
		expectedSlots    []domain.TimeSlot
		expectedError    error
	}{
		{
			name:            "empty day, full slots",
			date:            time.Date(2025, time.July, 9, 0, 0, 0, 0, loc),
			durationMinutes: testServiceDurationMinutes, // Using test-specific constant
			existingAppts:   []domain.Appointment{},
			expectedSlots:   generateExpectedSlots(time.Date(2025, time.July, 9, 0, 0, 0, 0, loc), testServiceDurationMinutes, loc),
			expectedError:   nil,
		},
		{
			name:            "single existing appointment, some slots blocked",
			date:            time.Date(2025, time.July, 10, 0, 0, 0, 0, loc),
			durationMinutes: testServiceDurationMinutes,
			existingAppts: []domain.Appointment{
				{
					StartTime: time.Date(2025, time.July, 10, 10, 0, 0, 0, loc),
					EndTime:   time.Date(2025, time.July, 10, 11, 0, 0, 0, loc), // 10:00 - 11:00
				},
			},
			expectedSlots: func() []domain.TimeSlot {
				allSlots := generateExpectedSlots(time.Date(2025, time.July, 10, 0, 0, 0, 0, loc), testServiceDurationMinutes, loc)
				var filtered []domain.TimeSlot
				for _, slot := range allSlots {
					// Filter out slots that overlap with 10:00-11:00
					// A 60-min slot starting at 9:00 ends at 10:00 (OK)
					// A 60-min slot starting at 9:15 ends at 10:15 (Overlap)
					// A 60-min slot starting at 9:30 ends at 10:30 (Overlap)
					// A 60-min slot starting at 9:45 ends at 10:45 (Overlap)
					// A 60-min slot starting at 10:00 ends at 11:00 (Overlap)
					// A 60-min slot starting at 10:15 ends at 11:15 (Overlap)
					// A 60-min slot starting at 10:30 ends at 11:30 (Overlap)
					// A 60-min slot starting at 10:45 ends at 11:45 (Overlap)
					// A 60-min slot starting at 11:00 ends at 12:00 (OK)

					// Check if current slot's start or end overlaps
					// `currentSlotStart.Before(appt.EndTime) && currentSlotEnd.After(appt.StartTime)`
					apptStart := time.Date(2025, time.July, 10, 10, 0, 0, 0, loc)
					apptEnd := time.Date(2025, time.July, 10, 11, 0, 0, 0, loc)

					if !(slot.Start.Before(apptEnd) && slot.End.After(apptStart)) {
						filtered = append(filtered, slot)
					}
				}
				return filtered
			}(),
			expectedError: nil,
		},
		{
			name:            "zero duration",
			date:            time.Date(2025, time.July, 11, 0, 0, 0, 0, loc),
			durationMinutes: 0,
			existingAppts:   []domain.Appointment{},
			expectedSlots:   nil,
			expectedError:   domain.ErrInvalidDuration,
		},
		{
			name:             "repo error",
			date:             time.Date(2025, time.July, 12, 0, 0, 0, 0, loc),
			durationMinutes:  testServiceDurationMinutes,
			mockFindAllError: errors.New("database connection failed"),
			expectedSlots:    generateExpectedSlots(time.Date(2025, time.July, 12, 0, 0, 0, 0, loc), testServiceDurationMinutes, loc), // Fallback if repo errors are ignored in service
			expectedError:    nil,                                                                                                     // Service logs warning but continues if repo errors
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &MockAppointmentRepository{
				Appointments: tt.existingAppts,
				FindAllError: tt.mockFindAllError,
			}
			service := apnt_svc.NewService(repo)

			slots, err := service.GetAvailableTimeSlots(ctx, tt.date, tt.durationMinutes)

			if !errors.Is(err, tt.expectedError) {
				t.Errorf("expected error %v, got %v", tt.expectedError, err)
			}

			// Sort both slices before deep comparison for reliable results
			sortTimeSlots(slots)
			sortTimeSlots(tt.expectedSlots)

			if !reflect.DeepEqual(slots, tt.expectedSlots) {
				t.Errorf("GetAvailableTimeSlots mismatch for date %s, duration %d:\nGot:  %v\nWant: %v",
					tt.date.Format("2006-01-02"), tt.durationMinutes, formatTimeSlots(slots), formatTimeSlots(tt.expectedSlots))
			}
		})
	}

	t.Run("should not return slots in the past", func(t *testing.T) {
		pastDate := time.Date(2023, time.January, 1, 0, 0, 0, 0, loc)
		repo := &MockAppointmentRepository{}
		service := apnt_svc.NewService(repo)

		// Set a specific "now" for this test (note: newWithClock doesn't mock actual time.Now() in service)
		// This test would only pass if `time.Now()` itself was mocked for the service.
		// As `service.go` uses `time.Now()` directly, this test's `newWithClock` is non-functional for true mocking.
		// It will pass because pastDate is always before current time.Now().
		// defer newWithClock(time.Date(2025, time.July, 1, 10, 0, 0, 0, loc))()

		slots, err := service.GetAvailableTimeSlots(ctx, pastDate, testServiceDurationMinutes)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(slots) > 0 {
			t.Errorf("expected no slots for past date, got %d", len(slots))
		}
	})

	t.Run("should not return slots outside working hours", func(t *testing.T) {
		date := time.Date(2025, time.July, 9, 0, 0, 0, 0, loc)
		repo := &MockAppointmentRepository{}
		service := apnt_svc.NewService(repo)

		// This defer newWithClock is illustrative but won't mock `time.Now()` inside the service.
		// defer newWithClock(time.Date(2025, time.July, 9, 10, 0, 0, 0, loc))()

		slots, err := service.GetAvailableTimeSlots(ctx, date, testServiceDurationMinutes)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Check if any returned slot starts before WorkStartHour or ends after WorkEndHour
		for _, slot := range slots {
			// CORRECTED: Use constants from the imported package
			if slot.Start.Hour() < apnt_svc.WorkDayStartHour || slot.End.Hour() > apnt_svc.WorkDayEndHour {
			if !(slot.End.Hour() == apnt_svc.WorkDayEndHour && slot.End.Minute() == 0) { // Allow ending exactly at WorkEndHour
				t.Errorf("slot %s-%s is outside working hours %d:00-%d:00",
					slot.Start.Format("15:04"), slot.End.Format("15:04"), apnt_svc.WorkDayStartHour, apnt_svc.WorkDayEndHour)
				}
			}
		}
	})
}

func TestCreateAppointment(t *testing.T) {
	ctx := context.Background()
	loc, _ := time.LoadLocation("Europe/Moscow")

	sampleService := domain.Service{ID: "svc1", Name: "Test Service", DurationMinutes: 60, Price: 1000}

	tests := []struct {
		name             string
		appointment      *domain.Appointment
		existingAppts    []domain.Appointment
		mockCreateError  error
		mockFindAllError error
		expectedError    error
	}{
		{
			name: "successful creation",
			appointment: &domain.Appointment{
				Service:  sampleService,
				Time:     time.Date(2025, 7, 15, 10, 0, 0, 0, loc),
				Duration: sampleService.DurationMinutes,
			},
			existingAppts: nil,
			expectedError: nil,
		},
		{
			name: "appointment in past",
			appointment: &domain.Appointment{
				Service:  sampleService,
				Time:     time.Date(2023, 1, 1, 10, 0, 0, 0, loc), // In the past
				Duration: sampleService.DurationMinutes,
			},
			existingAppts: nil,
			expectedError: domain.ErrAppointmentInPast,
		},
		{
			name: "outside working hours (start too early)",
			appointment: &domain.Appointment{
				Service:  sampleService,
				Time:     time.Date(2025, 7, 15, 8, 0, 0, 0, loc), // Before WorkStartHour
				Duration: sampleService.DurationMinutes,
			},
			existingAppts: nil,
			expectedError: domain.ErrOutsideWorkingHours,
		},
		{
			name: "outside working hours (ends too late)",
			appointment: &domain.Appointment{
				Service:  sampleService,
				Time:     time.Date(2025, 7, 15, 17, 30, 0, 0, loc), // Ends at 18:30 (after WorkEndHour)
				Duration: sampleService.DurationMinutes,
			},
			existingAppts: nil,
			expectedError: domain.ErrOutsideWorkingHours,
		},
		{
			name: "slot unavailable (exact overlap)",
			appointment: &domain.Appointment{
				Service:  sampleService,
				Time:     time.Date(2025, 7, 16, 10, 0, 0, 0, loc),
				Duration: sampleService.DurationMinutes,
			},
			existingAppts: []domain.Appointment{
				{
					StartTime: time.Date(2025, 7, 16, 10, 0, 0, 0, loc),
					EndTime:   time.Date(2025, 7, 16, 11, 0, 0, 0, loc),
				},
			},
			expectedError: domain.ErrSlotUnavailable,
		},
		{
			name: "slot unavailable (partial overlap, new starts during existing)",
			appointment: &domain.Appointment{
				Service:  sampleService,
				Time:     time.Date(2025, 7, 16, 10, 30, 0, 0, loc),
				Duration: sampleService.DurationMinutes,
			},
			existingAppts: []domain.Appointment{
				{
					StartTime: time.Date(2025, 7, 16, 10, 0, 0, 0, loc),
					EndTime:   time.Date(2025, 7, 16, 11, 0, 0, 0, loc),
				},
			},
			expectedError: domain.ErrSlotUnavailable,
		},
		{
			name: "repository create error",
			appointment: &domain.Appointment{
				Service:  sampleService,
				Time:     time.Date(2025, 7, 17, 10, 0, 0, 0, loc),
				Duration: sampleService.DurationMinutes,
			},
			mockCreateError: errors.New("db write failed"),
			expectedError:   errors.New("failed to create appointment in repository: db write failed"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &MockAppointmentRepository{
				Appointments: tt.existingAppts,
				CreateError:  tt.mockCreateError,
				FindAllError: tt.mockFindAllError,
			}
			service := apnt_svc.NewService(repo)

			// The newWithClock stub doesn't change `time.Now()` for service methods.
			// defer newWithClock(time.Date(2025, 7, 1, 9, 0, 0, 0, loc))()

			_, err := service.CreateAppointment(ctx, tt.appointment)

			if !errors.Is(err, tt.expectedError) && err == nil && tt.expectedError == nil {
				// OK
			} else if errors.Is(err, tt.expectedError) {
				// OK
			} else if err != nil && tt.expectedError != nil && err.Error() == tt.expectedError.Error() {
				// OK (for non-wrapped errors)
			} else {
				t.Errorf("expected error %v, got %v", tt.expectedError, err)
			}
		})
	}
}

func TestCancelAppointment(t *testing.T) {
	ctx := context.Background()
	repo := &MockAppointmentRepository{}
	service := apnt_svc.NewService(repo)

	t.Run("successful cancellation", func(t *testing.T) {
		repo.DeleteFunc = func(ctx context.Context, id string) error {
			if id != "existing-id" {
				return domain.ErrAppointmentNotFound
			}
			return nil
		}
		err := service.CancelAppointment(ctx, "existing-id")
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})

	t.Run("cancellation with empty ID", func(t *testing.T) {
		err := service.CancelAppointment(ctx, "")
		if err == nil {
			t.Error("expected error for empty ID, got nil")
		}
	})

	t.Run("repository delete error", func(t *testing.T) {
		repo.DeleteError = errors.New("delete failed")
		err := service.CancelAppointment(ctx, "some-id")
		if err == nil {
			t.Error("expected error from repository, got nil")
		}
		if err != nil && err.Error() != "failed to cancel appointment in repository: delete failed" {
			t.Errorf("unexpected error message: %v", err)
		}
	})
}

// --- Helper functions for tests ---

// generateExpectedSlots creates a slice of expected time slots for a given day.
func generateExpectedSlots(date time.Time, slotDurationMinutes int, loc *time.Location) []domain.TimeSlot {
	var slots []domain.TimeSlot
	// CORRECTED: Use constants from the imported package
	dayStart := time.Date(date.Year(), date.Month(), date.Day(), apnt_svc.WorkDayStartHour, 0, 0, 0, loc)
	dayEnd := time.Date(date.Year(), date.Month(), date.Day(), apnt_svc.WorkDayEndHour, 0, 0, 0, loc)

	interval := *apnt_svc.SlotDuration // Use the slot duration from the service package
	serviceDuration := time.Duration(slotDurationMinutes) * time.Minute

	for current := dayStart; current.Add(serviceDuration).Before(dayEnd.Add(1 * time.Minute)); current = current.Add(interval) {
		if current.Before(time.Now()) { // Skip past slots, similar to service logic
			continue
		}
		slots = append(slots, domain.TimeSlot{Start: current, End: current.Add(serviceDuration)})
	}
	return slots
}

// sortTimeSlots sorts a slice of TimeSlot by their Start time.
func sortTimeSlots(slots []domain.TimeSlot) {
	// A simple bubble sort for small slices, replace with sort.Slice for larger ones
	for i := 0; i < len(slots)-1; i++ {
		for j := 0; j < len(slots)-i-1; j++ {
			if slots[j].Start.After(slots[j+1].Start) {
				slots[j], slots[j+1] = slots[j+1], slots[j]
			}
		}
	}
}

// formatTimeSlots for better debug output
func formatTimeSlots(slots []domain.TimeSlot) []string {
	var s []string
	for _, slot := range slots {
		s = append(s, fmt.Sprintf("%s-%s", slot.Start.Format("15:04"), slot.End.Format("15:04")))
	}
	return s
}
