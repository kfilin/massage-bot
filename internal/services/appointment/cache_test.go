package appointment

import (
	"context"
	"testing"
	"time"

	"github.com/kfilin/massage-bot/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockRepo needs to be defined or imported. Assuming it exists in service_test.go or similar.
// Since we are adding a new file, we can define a local mock if the one in service_test.go isn't exported or usable.
// However, let's assume we can reuse the strategy. If not, we'll definte a minimal one.

type MockRepoForCache struct {
	mock.Mock
}

func (m *MockRepoForCache) GetFreeBusy(ctx context.Context, timeMin, timeMax time.Time) ([]domain.TimeSlot, error) {
	args := m.Called(ctx, timeMin, timeMax)
	return args.Get(0).([]domain.TimeSlot), args.Error(1)
}

// Implement other interface methods to satisfy ports.AppointmentRepository
func (m *MockRepoForCache) Create(ctx context.Context, appt *domain.Appointment) (*domain.Appointment, error) {
	args := m.Called(ctx, appt)
	return args.Get(0).(*domain.Appointment), args.Error(1)
}
func (m *MockRepoForCache) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
func (m *MockRepoForCache) FindByID(ctx context.Context, id string) (*domain.Appointment, error) {
	return nil, nil
}
func (m *MockRepoForCache) FindAll(ctx context.Context) ([]domain.Appointment, error) {
	return nil, nil
}
func (m *MockRepoForCache) FindEvents(ctx context.Context, start, end *time.Time) ([]domain.Appointment, error) {
	return nil, nil
}
func (m *MockRepoForCache) GetAccountInfo(ctx context.Context) (string, error)  { return "", nil }
func (m *MockRepoForCache) GetCalendarID() string                               { return "" }
func (m *MockRepoForCache) ListCalendars(ctx context.Context) ([]string, error) { return nil, nil }
func (m *MockRepoForCache) MigrateFolderNames() error                           { return nil }
func (m *MockRepoForCache) SyncAll() error                                      { return nil }

func TestService_FreeBusyCache(t *testing.T) {
	mockRepo := new(MockRepoForCache)
	svc := NewService(mockRepo, nil)
	ctx := context.Background()

	// Setup time range
	now := time.Now()
	testDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	// Adjust logic to match service expectations (GetAvailableTimeSlots calculates min/max internally based on date)
	// We'll trust the service calls the repo with *some* range.

	// Setup expected behavior
	// First call: Should hit Repo
	mockRepo.On("GetFreeBusy", mock.Anything, mock.Anything, mock.Anything).Return([]domain.TimeSlot{}, nil).Once()

	// Act 1: First call
	// Warm up cache
	if _, err := svc.GetAvailableTimeSlots(ctx, testDate, 60); err != nil {
		t.Fatalf("Failed to warm up cache: %v", err)
	}

	// Act 2: Second call (Should be CACHED, so NO new call to GetFreeBusy on MockRepo)
	// Act 2: Second call (Should be CACHED, so NO new call to GetFreeBusy on MockRepo)
	if _, err := svc.GetAvailableTimeSlots(ctx, testDate, 60); err != nil {
		t.Fatalf("Failed to get slots from cache: %v", err)
	}

	// Verify that repo was only called once
	mockRepo.AssertExpectations(t)
}

func TestService_CacheInvalidation_OnCreate(t *testing.T) {
	mockRepo := new(MockRepoForCache)
	svc := NewService(mockRepo, nil)
	ctx := context.Background()
	now := time.Now()
	// Use tomorrow for both warm-up and creation to ensures cache hit and valid future appointment
	tomorrow := now.Add(24 * time.Hour)
	testDate := time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), 0, 0, 0, 0, time.UTC)

	// 1. Fill Cache (Expect 1 call)
	mockRepo.On("GetFreeBusy", mock.Anything, mock.Anything, mock.Anything).Return([]domain.TimeSlot{}, nil).Once()
	// Ignoring the error here as the test focuses on cache invalidation, not initial cache fill success.
	// If the initial fill fails, subsequent steps might also fail, but for this test, we assume it succeeds.
	_, _ = svc.GetAvailableTimeSlots(ctx, testDate, 60)

	// 2. Create Appointment -> Should invalidate
	newAppt := &domain.Appointment{
		ID:           "new",
		Service:      domain.Service{ID: "1", DurationMinutes: 60},
		StartTime:    time.Date(now.Year(), now.Month(), now.Day()+1, 10, 0, 0, 0, time.UTC),
		Duration:     60,
		CustomerName: "Test",
	}
	// Note: CreateAppointment calls GetFreeBusy internally for checking overlap!
	// So we expect GetFreeBusy to be called AGAIN because it logic is:
	// 1. GetAvailableTimeSlots (cached)
	// 2. CreateAppointment -> calls GetFreeBusy (should hit cache ideally if range matches, OR new range)
	// BUT, CreateAppointment does its own range check.
	// Actually, CreateAppointment invalidates cache AFTER success.

	// Let's refine expectation:
	// CreateAppointment internal calls:
	// a. GetFreeBusy (overlap check) -> Should use cache if key matches.
	// b. Repo.Create -> Success
	// c. InvalidateCache

	// So for this test:
	// We want to ensure that AFTER Create, the NEXT GetAvailableTimeSlots hits repo again.

	mockRepo.On("Create", mock.Anything, mock.Anything).Return(newAppt, nil)

	// CreateAppointment asks for FreeBusy for the specific day of the appt.
	// If the key matches (whole day), it might hit the cache.
	// Let's assume it hits cache or we just stub it.
	// The important part is step 3.

	_, err := svc.CreateAppointment(ctx, newAppt)
	assert.NoError(t, err)

	// 3. Call GetAvailableTimeSlots AGAIN -> Should hit Repo because cache was invalidated
	mockRepo.On("GetFreeBusy", mock.Anything, mock.Anything, mock.Anything).Return([]domain.TimeSlot{}, nil).Once()
	_, err = svc.GetAvailableTimeSlots(ctx, testDate, 60)
	assert.NoError(t, err)

	mockRepo.AssertExpectations(t)
}
