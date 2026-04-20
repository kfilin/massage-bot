package appointment

import (
	"context"
	"testing"
	"time"

	"github.com/kfilin/massage-bot/internal/domain"
)

// slotTestRepo satisfies ports.AppointmentRepository for slot engine tests.
// It reuses the mockRepo defined in service_test.go (same package).

func TestGetAvailableTimeSlots_InvalidDuration(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo, nil)
	svc.NowFunc = func() time.Time { return time.Now() }

	_, err := svc.GetAvailableTimeSlots(context.Background(), time.Now(), 0)
	if err != domain.ErrInvalidDuration {
		t.Errorf("Expected ErrInvalidDuration for 0 duration, got %v", err)
	}

	_, err = svc.GetAvailableTimeSlots(context.Background(), time.Now(), -10)
	if err != domain.ErrInvalidDuration {
		t.Errorf("Expected ErrInvalidDuration for negative duration, got %v", err)
	}
}

func TestGetAvailableTimeSlots_AllFree(t *testing.T) {
	repo := newMockRepo()
	// No busy slots
	repo.getFreeBusyFunc = func(ctx context.Context, start, end time.Time) ([]domain.TimeSlot, error) {
		return nil, nil
	}
	svc := NewService(repo, nil)

	// Set "now" to yesterday so none of today's slots are in the past
	domain.ApptTimeZone = time.UTC
	yesterday := time.Now().UTC().Add(-24 * time.Hour)
	svc.NowFunc = func() time.Time { return yesterday }

	date := time.Now().UTC()
	slots, err := svc.GetAvailableTimeSlots(context.Background(), date, 60)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Working hours 09:00–18:00 with 60-min slots = 9 possible slots
	if len(slots) != 9 {
		t.Errorf("Expected 9 free slots (60-min), got %d", len(slots))
	}
}

func TestGetAvailableTimeSlots_BusyBlocking(t *testing.T) {
	domain.ApptTimeZone = time.UTC
	date := time.Now().UTC()
	// Busy from 10:00 to 11:00
	dayBase := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
	busyStart := dayBase.Add(10 * time.Hour)
	busyEnd := dayBase.Add(11 * time.Hour)

	repo := newMockRepo()
	repo.getFreeBusyFunc = func(ctx context.Context, start, end time.Time) ([]domain.TimeSlot, error) {
		return []domain.TimeSlot{{Start: busyStart, End: busyEnd}}, nil
	}
	svc := NewService(repo, nil)
	yesterday := time.Now().UTC().Add(-24 * time.Hour)
	svc.NowFunc = func() time.Time { return yesterday }

	slots, err := svc.GetAvailableTimeSlots(context.Background(), date, 60)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// 9 total - 1 busy = 8 available
	if len(slots) != 8 {
		t.Errorf("Expected 8 free slots with one busy, got %d", len(slots))
	}

	// Verify 10:00 slot is not present
	for _, s := range slots {
		if s.Start.Hour() == 10 {
			t.Error("10:00 slot should be busy but appeared in available slots")
		}
	}
}

func TestGetAvailableTimeSlots_PastSlotsExcluded(t *testing.T) {
	domain.ApptTimeZone = time.UTC
	date := time.Now().UTC()
	dayBase := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)

	repo := newMockRepo()
	repo.getFreeBusyFunc = func(ctx context.Context, start, end time.Time) ([]domain.TimeSlot, error) {
		return nil, nil
	}
	svc := NewService(repo, nil)
	// Set "now" to 13:00 today — slots at 9,10,11,12 should be excluded
	svc.NowFunc = func() time.Time { return dayBase.Add(13 * time.Hour) }

	slots, err := svc.GetAvailableTimeSlots(context.Background(), date, 60)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Slots 13:00–18:00 = 5 available (13,14,15,16,17)
	if len(slots) != 5 {
		t.Errorf("Expected 5 future slots after 13:00, got %d", len(slots))
	}
	for _, s := range slots {
		if s.Start.Hour() < 13 {
			t.Errorf("Slot at %d:00 should have been excluded as past", s.Start.Hour())
		}
	}
}

func TestGetAvailableTimeSlots_AllBusy(t *testing.T) {
	domain.ApptTimeZone = time.UTC
	date := time.Now().UTC()
	dayBase := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
	// Block the entire workday
	busyStart := dayBase.Add(time.Duration(domain.WorkDayStartHour) * time.Hour)
	busyEnd := dayBase.Add(time.Duration(domain.WorkDayEndHour) * time.Hour)

	repo := newMockRepo()
	repo.getFreeBusyFunc = func(ctx context.Context, start, end time.Time) ([]domain.TimeSlot, error) {
		return []domain.TimeSlot{{Start: busyStart, End: busyEnd}}, nil
	}
	svc := NewService(repo, nil)
	yesterday := time.Now().UTC().Add(-24 * time.Hour)
	svc.NowFunc = func() time.Time { return yesterday }

	slots, err := svc.GetAvailableTimeSlots(context.Background(), date, 60)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(slots) != 0 {
		t.Errorf("Expected 0 slots when entire day is busy, got %d", len(slots))
	}
}

func TestGetAvailableTimeSlots_40MinDuration(t *testing.T) {
	domain.ApptTimeZone = time.UTC
	date := time.Now().UTC()

	repo := newMockRepo()
	repo.getFreeBusyFunc = func(ctx context.Context, start, end time.Time) ([]domain.TimeSlot, error) {
		return nil, nil
	}
	svc := NewService(repo, nil)
	yesterday := time.Now().UTC().Add(-24 * time.Hour)
	svc.NowFunc = func() time.Time { return yesterday }

	// 40 min service, 60 min step: same 9 starting slots (step is still 60 min),
	// but each 40 min slot ends by 17:40 at latest — all 9 fit
	slots, err := svc.GetAvailableTimeSlots(context.Background(), date, 40)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	// Last possible start: 17:00 (17:00 + 40m = 17:40 ≤ 18:00) → 9 slots
	if len(slots) != 9 {
		t.Errorf("Expected 9 slots for 40-min service, got %d", len(slots))
	}
}
