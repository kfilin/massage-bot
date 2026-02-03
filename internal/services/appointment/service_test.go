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
	svc := NewService(repo)
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
	svc := NewService(repo)
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
	svc := NewService(repo)
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
	svc := NewServiceWithMetrics(repo, &NoOpCollector{})

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
