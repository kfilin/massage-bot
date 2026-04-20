package appointment

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/kfilin/massage-bot/internal/domain"
)

// mockRepo is a simple mock implementation of ports.AppointmentRepository
type mockRepo struct {
	appointments    map[string]*domain.Appointment
	shouldError     bool
	getFreeBusyFunc func(context.Context, time.Time, time.Time) ([]domain.TimeSlot, error)
}

func newMockRepo() *mockRepo {
	return &mockRepo{
		appointments: make(map[string]*domain.Appointment),
	}
}

func (m *mockRepo) Create(ctx context.Context, appt *domain.Appointment) (*domain.Appointment, error) {
	if m.shouldError {
		return nil, errors.New("mock error")
	}
	if appt.ID == "" {
		appt.ID = "generated-id"
	}
	m.appointments[appt.ID] = appt
	return appt, nil
}

func (m *mockRepo) FindByID(ctx context.Context, id string) (*domain.Appointment, error) {
	if m.shouldError {
		return nil, errors.New("mock error")
	}
	if appt, ok := m.appointments[id]; ok {
		return appt, nil
	}
	return nil, domain.ErrAppointmentNotFound
}

func (m *mockRepo) Delete(ctx context.Context, id string) error {
	if m.shouldError {
		return errors.New("mock error")
	}
	if _, ok := m.appointments[id]; ok {
		delete(m.appointments, id)
		return nil
	}
	return domain.ErrAppointmentNotFound
}

func (m *mockRepo) FindAll(ctx context.Context) ([]domain.Appointment, error) {
	if m.shouldError {
		return nil, errors.New("mock error")
	}
	var res []domain.Appointment
	for _, a := range m.appointments {
		res = append(res, *a)
	}
	return res, nil
}

func (m *mockRepo) FindEvents(ctx context.Context, start, end *time.Time) ([]domain.Appointment, error) {
	if m.shouldError {
		return nil, errors.New("mock error")
	}
	// Simplified mock: return all
	var res []domain.Appointment
	for _, a := range m.appointments {
		res = append(res, *a)
	}
	return res, nil
}

func (m *mockRepo) GetFreeBusy(ctx context.Context, start, end time.Time) ([]domain.TimeSlot, error) {
	if m.getFreeBusyFunc != nil {
		return m.getFreeBusyFunc(ctx, start, end)
	}
	return nil, nil
}

func (m *mockRepo) GetAccountInfo(ctx context.Context) (string, error) { return "mock@gmail.com", nil }
func (m *mockRepo) GetCalendarID() string                              { return "mock-cal-id" }
func (m *mockRepo) ListCalendars(ctx context.Context) ([]string, error) {
	return []string{"primary"}, nil
}

func TestService_FindByID(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo, nil)
	existingAppt := &domain.Appointment{ID: "123", CustomerName: "Alice"}
	repo.appointments["123"] = existingAppt

	tests := []struct {
		name    string
		id      string
		wantErr bool
	}{
		{"Existing", "123", false},
		{"Missing", "999", true},
		{"Empty ID", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := svc.FindByID(context.Background(), tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("FindByID() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && got.ID != tt.id {
				t.Errorf("FindByID() ID = %v, want %v", got.ID, tt.id)
			}
		})
	}
}

func TestService_CancelAppointment(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo, nil)
	repo.appointments["valid"] = &domain.Appointment{ID: "valid"}

	tests := []struct {
		name    string
		id      string
		wantErr bool
	}{
		{"Success", "valid", false},
		{"Not Found", "missing", true},
		{"Invalid ID", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := svc.CancelAppointment(context.Background(), tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("CancelAppointment() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestService_Getters(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo, nil)
	ctx := context.Background()

	// Available Services
	services, err := svc.GetAvailableServices(ctx)
	if err != nil {
		t.Errorf("GetAvailableServices failed: %v", err)
	}
	if len(services) == 0 {
		t.Error("Expected services, got empty list")
	}

	// Customer History
	repo.appointments["1"] = &domain.Appointment{ID: "1", CustomerTgID: "user1"}
	repo.appointments["2"] = &domain.Appointment{ID: "2", CustomerTgID: "user2"}

	hist, err := svc.GetCustomerHistory(ctx, "user1")
	if err != nil {
		t.Errorf("GetCustomerHistory failed: %v", err)
	}
	if len(hist) != 1 || hist[0].ID != "1" {
		t.Errorf("Expected 1 appt for user1, got %v", hist)
	}

	// Customer Appointments (Upcoming)
	appts, err := svc.GetCustomerAppointments(ctx, "user2")
	if err != nil {
		t.Errorf("GetCustomerAppointments failed: %v", err)
	}
	if len(appts) != 1 || appts[0].ID != "2" {
		t.Errorf("Expected 1 appt for user2, got %v", appts)
	}

	// GetAllUpcomingAppointments
	// Currently mock FindAll returns all, GetAllUpcoming filters.
	// We didn't set date on mock appts, so they are Zero time (past).
	// Let's add a future appointment
	future := time.Now().Add(24 * time.Hour)
	repo.appointments["3"] = &domain.Appointment{ID: "3", StartTime: future, Status: "confirmed"}

	upcoming, err := svc.GetAllUpcomingAppointments(ctx)
	if err != nil {
		t.Errorf("GetAllUpcomingAppointments failed: %v", err)
	}
	if len(upcoming) != 1 || upcoming[0].ID != "3" {
		t.Errorf("Expected 1 upcoming appt, got %v", upcoming)
	}

	// GetUpcomingAppointments (wraps FindEvents)
	rangeAppts, err := svc.GetUpcomingAppointments(ctx, time.Now(), future.Add(time.Hour))
	if err != nil {
		t.Errorf("GetUpcomingAppointments failed: %v", err)
	}
	if len(rangeAppts) < 3 {
		// Mock FindEvents returns all (3)
		t.Errorf("Expected at least 3 appts, got %d", len(rangeAppts))
	}

	// Post-check Total count
	count, err := svc.GetTotalUpcomingCount(ctx)
	if err != nil {
		t.Errorf("GetTotalUpcomingCount failed: %v", err)
	}
	if count != 3 { // FindAll returns 3
		t.Errorf("Expected 3 total, got %d", count)
	}

	// Config Getters
	if svc.GetCalendarID() != "mock-cal-id" {
		t.Error("GetCalendarID failed")
	}
	info, _ := svc.GetCalendarAccountInfo(ctx)
	if info != "mock@gmail.com" {
		t.Error("GetCalendarAccountInfo failed")
	}
	calendars, _ := svc.ListCalendars(ctx)
	if len(calendars) != 1 {
		t.Error("ListCalendars failed")
	}
}

func TestService_CreateAppointment(t *testing.T) {
	repo := newMockRepo()
	svc := NewServiceWithMetrics(repo, nil, &NoOpCollector{})

	// Mock Now: 2023-10-25 10:00 (Wed)
	mockNow := time.Date(2023, 10, 25, 10, 0, 0, 0, time.UTC)
	svc.NowFunc = func() time.Time { return mockNow }

	// Fix timezone for tests
	domain.ApptTimeZone = time.UTC

	validService := domain.Service{ID: "1", Name: "Massage", DurationMinutes: 60}

	tests := []struct {
		name     string
		appt     *domain.Appointment
		mockBusy []domain.TimeSlot
		wantErr  error
	}{
		{
			name: "Success",
			appt: &domain.Appointment{
				Service:      validService,
				CustomerName: "Bob",
				StartTime:    mockNow.Add(24 * time.Hour), // Tomorrow 10:00
				Duration:     60,
			},
			wantErr: nil,
		},
		{
			name: "Past Time",
			appt: &domain.Appointment{
				Service:      validService,
				CustomerName: "Bob",
				StartTime:    mockNow.Add(-1 * time.Hour),
				Duration:     60,
			},
			wantErr: domain.ErrAppointmentInPast,
		},
		{
			name: "Outside Working Hours",
			appt: &domain.Appointment{
				Service:      validService,
				CustomerName: "Bob",
				StartTime:    time.Date(2023, 10, 26, 5, 0, 0, 0, time.UTC),
				Duration:     60,
			},
			wantErr: domain.ErrOutsideWorkingHours,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo.getFreeBusyFunc = func(ctx context.Context, start, end time.Time) ([]domain.TimeSlot, error) {
				return tt.mockBusy, nil
			}

			_, err := svc.CreateAppointment(context.Background(), tt.appt)
			if tt.wantErr != nil {
				if err == nil {
					t.Errorf("Expected error %v, got nil", tt.wantErr)
				}
			} else if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestService_CreateAppointment_NilAppointment(t *testing.T) {
	repo := newMockRepo()
	svc := NewServiceWithMetrics(repo, nil, &NoOpCollector{})
	_, err := svc.CreateAppointment(context.Background(), nil)
	if err == nil {
		t.Error("Expected error for nil appointment, got nil")
	}
}

func TestService_CreateAppointment_MissingFields(t *testing.T) {
	repo := newMockRepo()
	svc := NewServiceWithMetrics(repo, nil, &NoOpCollector{})
	domain.ApptTimeZone = time.UTC

	// Missing service ID
	_, err := svc.CreateAppointment(context.Background(), &domain.Appointment{
		StartTime:    time.Now().Add(24 * time.Hour),
		Duration:     60,
		CustomerName: "Alice",
	})
	if err == nil {
		t.Error("Expected error for missing service ID, got nil")
	}

	// Missing customer name
	_, err = svc.CreateAppointment(context.Background(), &domain.Appointment{
		Service:   domain.Service{ID: "1"},
		StartTime: time.Now().Add(24 * time.Hour),
		Duration:  60,
	})
	if err == nil {
		t.Error("Expected error for missing customer name, got nil")
	}

	// Zero duration
	_, err = svc.CreateAppointment(context.Background(), &domain.Appointment{
		Service:      domain.Service{ID: "1"},
		StartTime:    time.Now().Add(24 * time.Hour),
		Duration:     0,
		CustomerName: "Alice",
	})
	if err == nil {
		t.Error("Expected error for zero duration, got nil")
	}
}

func TestService_CreateAppointment_SlotConflict(t *testing.T) {
	repo := newMockRepo()
	svc := NewServiceWithMetrics(repo, nil, &NoOpCollector{})
	domain.ApptTimeZone = time.UTC
	mockNow := time.Date(2023, 10, 25, 10, 0, 0, 0, time.UTC)
	svc.NowFunc = func() time.Time { return mockNow }

	// Busy slot at 10:00–11:00 tomorrow
	tomorrow := time.Date(2023, 10, 26, 10, 0, 0, 0, time.UTC)
	busyStart := tomorrow
	busyEnd := tomorrow.Add(60 * time.Minute)
	repo.getFreeBusyFunc = func(ctx context.Context, start, end time.Time) ([]domain.TimeSlot, error) {
		return []domain.TimeSlot{{Start: busyStart, End: busyEnd}}, nil
	}

	_, err := svc.CreateAppointment(context.Background(), &domain.Appointment{
		Service:      domain.Service{ID: "1"},
		StartTime:    tomorrow,
		Duration:     60,
		CustomerName: "Bob",
	})
	if err != domain.ErrSlotUnavailable {
		t.Errorf("Expected ErrSlotUnavailable, got %v", err)
	}
}

func TestService_FreeBusy_CacheHit(t *testing.T) {
	callCount := 0
	repo := newMockRepo()
	repo.getFreeBusyFunc = func(ctx context.Context, start, end time.Time) ([]domain.TimeSlot, error) {
		callCount++
		return nil, nil
	}
	svc := NewServiceWithMetrics(repo, nil, &NoOpCollector{})

	start := time.Date(2023, 10, 25, 0, 0, 0, 0, time.UTC)
	end := start.Add(24 * time.Hour)

	// First call — cache miss → repo called
	_, _ = svc.getFreeBusy(context.Background(), start, end)
	// Second call same range — cache hit → repo NOT called again
	_, _ = svc.getFreeBusy(context.Background(), start, end)

	if callCount != 1 {
		t.Errorf("Expected repo.GetFreeBusy called once (cache hit on 2nd call), got %d", callCount)
	}
}

func TestService_InvalidateCache(t *testing.T) {
	callCount := 0
	repo := newMockRepo()
	repo.getFreeBusyFunc = func(ctx context.Context, start, end time.Time) ([]domain.TimeSlot, error) {
		callCount++
		return nil, nil
	}
	svc := NewServiceWithMetrics(repo, nil, &NoOpCollector{})

	start := time.Date(2023, 10, 25, 0, 0, 0, 0, time.UTC)
	end := start.Add(24 * time.Hour)

	_, _ = svc.getFreeBusy(context.Background(), start, end) // miss → callCount=1
	svc.invalidateCache()
	_, _ = svc.getFreeBusy(context.Background(), start, end) // miss again → callCount=2

	if callCount != 2 {
		t.Errorf("Expected 2 repo calls after cache invalidation, got %d", callCount)
	}
}

func TestService_GetCustomerAppointments_EmptyID(t *testing.T) {
	svc := NewService(newMockRepo(), nil)
	_, err := svc.GetCustomerAppointments(context.Background(), "")
	if err != domain.ErrInvalidID {
		t.Errorf("Expected ErrInvalidID, got %v", err)
	}
}

func TestService_GetCustomerHistory_EmptyID(t *testing.T) {
	svc := NewService(newMockRepo(), nil)
	_, err := svc.GetCustomerHistory(context.Background(), "")
	if err != domain.ErrInvalidID {
		t.Errorf("Expected ErrInvalidID, got %v", err)
	}
}

func TestService_GetAllUpcomingAppointments_FiltersCancelled(t *testing.T) {
	repo := newMockRepo()
	future := time.Now().Add(24 * time.Hour)
	repo.appointments["a1"] = &domain.Appointment{ID: "a1", Status: "confirmed", StartTime: future}
	repo.appointments["a2"] = &domain.Appointment{ID: "a2", Status: "cancelled", StartTime: future}

	svc := NewService(repo, nil)
	upcoming, err := svc.GetAllUpcomingAppointments(context.Background())
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(upcoming) != 1 || upcoming[0].ID != "a1" {
		t.Errorf("Expected only confirmed appointment, got %v", upcoming)
	}
}

func TestService_GetTotalUpcomingCount_RepoError(t *testing.T) {
	repo := newMockRepo()
	repo.shouldError = true
	svc := NewService(repo, nil)
	_, err := svc.GetTotalUpcomingCount(context.Background())
	if err == nil {
		t.Error("Expected error when repo fails, got nil")
	}
}
