package appointment

import (
	"context"
	"testing"
	"time"

	"github.com/kfilin/massage-bot/internal/domain"
)

type mockRepo struct {
	appointments []domain.Appointment
}

func (m *mockRepo) Create(ctx context.Context, appt *domain.Appointment) (*domain.Appointment, error) {
	return appt, nil
}
func (m *mockRepo) FindByID(ctx context.Context, id string) (*domain.Appointment, error) {
	return nil, nil
}
func (m *mockRepo) FindAll(ctx context.Context) ([]domain.Appointment, error) {
	return m.appointments, nil
}
func (m *mockRepo) Delete(ctx context.Context, id string) error {
	return nil
}
func (m *mockRepo) GetAccountInfo(ctx context.Context) (string, error) {
	return "mock@example.com", nil
}

func TestGetAvailableTimeSlots(t *testing.T) {
	repo := &mockRepo{
		appointments: []domain.Appointment{
			{
				StartTime: time.Date(2026, 1, 15, 10, 0, 0, 0, ApptTimeZone),
				EndTime:   time.Date(2026, 1, 15, 11, 0, 0, 0, ApptTimeZone),
			},
		},
	}
	s := NewService(repo)

	// Mock "now" to be before the test date
	s.NowFunc = func() time.Time {
		return time.Date(2026, 1, 1, 0, 0, 0, 0, ApptTimeZone)
	}

	date := time.Date(2026, 1, 15, 0, 0, 0, 0, ApptTimeZone)
	slots, err := s.GetAvailableTimeSlots(context.Background(), date, 60)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Work Day: 09:00 - 18:00 (9 slots of 60m)
	// One slot (10:00-11:00) is taken.
	// Expected: 8 slots.
	if len(slots) != 8 {
		t.Errorf("Expected 8 available slots, got %d", len(slots))
	}

	// Verify the taken slot is NOT in the list
	for _, slot := range slots {
		if slot.Start.Format("15:04") == "10:00" {
			t.Error("10:00 slot should be unavailable")
		}
	}
}

func TestCreateAppointmentValidation(t *testing.T) {
	repo := &mockRepo{}
	s := NewService(repo)
	s.NowFunc = func() time.Time {
		return time.Date(2026, 1, 1, 0, 0, 0, 0, ApptTimeZone)
	}

	// Test outside working hours (too early)
	appt := &domain.Appointment{
		Service:      domain.Service{ID: "1", Name: "Test"},
		StartTime:    time.Date(2026, 1, 15, 7, 0, 0, 0, ApptTimeZone),
		Duration:     60,
		CustomerName: "Test",
	}
	_, err := s.CreateAppointment(context.Background(), appt)
	if err == nil {
		t.Error("Should have failed for outside working hours (too early)")
	}

	// Test outside working hours (too late)
	appt.StartTime = time.Date(2026, 1, 15, 17, 30, 0, 0, ApptTimeZone) // Ends 18:30
	_, err = s.CreateAppointment(context.Background(), appt)
	if err == nil {
		t.Error("Should have failed for outside working hours (too late)")
	}
}
