package telegram

import (
	"context"
	"io"
	"testing"
	"time"

	"github.com/kfilin/massage-bot/internal/domain"
	"github.com/kfilin/massage-bot/internal/ports"
)

// Mock implementations for testing

// mockAppointmentService implements ports.AppointmentService
type mockAppointmentService struct {
	getAvailableServicesFunc       func(ctx context.Context) ([]domain.Service, error)
	getAvailableTimeSlotsFunc      func(ctx context.Context, date time.Time, durationMinutes int) ([]domain.TimeSlot, error)
	createAppointmentFunc          func(ctx context.Context, appointment *domain.Appointment) (*domain.Appointment, error)
	cancelAppointmentFunc          func(ctx context.Context, appointmentID string) error
	getUpcomingAppointmentsFunc    func(ctx context.Context, timeMin, timeMax time.Time) ([]domain.Appointment, error)
	getCustomerAppointmentsFunc    func(ctx context.Context, customerTgID string) ([]domain.Appointment, error)
	getCustomerHistoryFunc         func(ctx context.Context, customerTgID string) ([]domain.Appointment, error)
	getAllUpcomingAppointmentsFunc func(ctx context.Context) ([]domain.Appointment, error)
	findByIDFunc                   func(ctx context.Context, appointmentID string) (*domain.Appointment, error)
	getTotalUpcomingCountFunc      func(ctx context.Context) (int, error)
	getCalendarAccountInfoFunc     func(ctx context.Context) (string, error)
	getCalendarIDFunc              func() string
	listCalendarsFunc              func(ctx context.Context) ([]string, error)
}

func (m *mockAppointmentService) GetAvailableServices(ctx context.Context) ([]domain.Service, error) {
	if m.getAvailableServicesFunc != nil {
		return m.getAvailableServicesFunc(ctx)
	}
	return []domain.Service{}, nil
}

func (m *mockAppointmentService) GetAvailableTimeSlots(ctx context.Context, date time.Time, durationMinutes int) ([]domain.TimeSlot, error) {
	if m.getAvailableTimeSlotsFunc != nil {
		return m.getAvailableTimeSlotsFunc(ctx, date, durationMinutes)
	}
	return []domain.TimeSlot{}, nil
}

func (m *mockAppointmentService) CreateAppointment(ctx context.Context, appointment *domain.Appointment) (*domain.Appointment, error) {
	if m.createAppointmentFunc != nil {
		return m.createAppointmentFunc(ctx, appointment)
	}
	return appointment, nil
}

func (m *mockAppointmentService) CancelAppointment(ctx context.Context, appointmentID string) error {
	if m.cancelAppointmentFunc != nil {
		return m.cancelAppointmentFunc(ctx, appointmentID)
	}
	return nil
}

func (m *mockAppointmentService) GetUpcomingAppointments(ctx context.Context, timeMin, timeMax time.Time) ([]domain.Appointment, error) {
	if m.getUpcomingAppointmentsFunc != nil {
		return m.getUpcomingAppointmentsFunc(ctx, timeMin, timeMax)
	}
	return []domain.Appointment{}, nil
}

func (m *mockAppointmentService) GetCustomerAppointments(ctx context.Context, customerTgID string) ([]domain.Appointment, error) {
	if m.getCustomerAppointmentsFunc != nil {
		return m.getCustomerAppointmentsFunc(ctx, customerTgID)
	}
	return []domain.Appointment{}, nil
}

func (m *mockAppointmentService) GetCustomerHistory(ctx context.Context, customerTgID string) ([]domain.Appointment, error) {
	if m.getCustomerHistoryFunc != nil {
		return m.getCustomerHistoryFunc(ctx, customerTgID)
	}
	return []domain.Appointment{}, nil
}

func (m *mockAppointmentService) GetAllUpcomingAppointments(ctx context.Context) ([]domain.Appointment, error) {
	if m.getAllUpcomingAppointmentsFunc != nil {
		return m.getAllUpcomingAppointmentsFunc(ctx)
	}
	return []domain.Appointment{}, nil
}

func (m *mockAppointmentService) FindByID(ctx context.Context, appointmentID string) (*domain.Appointment, error) {
	if m.findByIDFunc != nil {
		return m.findByIDFunc(ctx, appointmentID)
	}
	return nil, nil
}

func (m *mockAppointmentService) GetTotalUpcomingCount(ctx context.Context) (int, error) {
	if m.getTotalUpcomingCountFunc != nil {
		return m.getTotalUpcomingCountFunc(ctx)
	}
	return 0, nil
}

func (m *mockAppointmentService) GetCalendarAccountInfo(ctx context.Context) (string, error) {
	if m.getCalendarAccountInfoFunc != nil {
		return m.getCalendarAccountInfoFunc(ctx)
	}
	return "test@example.com", nil
}

func (m *mockAppointmentService) GetCalendarID() string {
	if m.getCalendarIDFunc != nil {
		return m.getCalendarIDFunc()
	}
	return "test-calendar-id"
}

func (m *mockAppointmentService) ListCalendars(ctx context.Context) ([]string, error) {
	if m.listCalendarsFunc != nil {
		return m.listCalendarsFunc(ctx)
	}
	return []string{"calendar1", "calendar2"}, nil
}

// mockSessionStorage implements ports.SessionStorage
type mockSessionStorage struct {
	sessions map[int64]map[string]interface{}
}

func newMockSessionStorage() *mockSessionStorage {
	return &mockSessionStorage{
		sessions: make(map[int64]map[string]interface{}),
	}
}

func (m *mockSessionStorage) Set(userID int64, key string, value interface{}) {
	if m.sessions[userID] == nil {
		m.sessions[userID] = make(map[string]interface{})
	}
	m.sessions[userID][key] = value
}

func (m *mockSessionStorage) Get(userID int64) map[string]interface{} {
	if m.sessions[userID] == nil {
		return make(map[string]interface{})
	}
	return m.sessions[userID]
}

func (m *mockSessionStorage) ClearSession(userID int64) {
	delete(m.sessions, userID)
}

// mockRepository implements ports.Repository
type mockRepository struct {
	patients           map[string]domain.Patient
	bannedUsers        map[string]bool
	appointmentHistory map[string][]domain.Appointment
	isUserBannedFunc   func(telegramID string, username string) (bool, error)
	savePatientFunc    func(patient domain.Patient) error
	getPatientFunc     func(telegramID string) (domain.Patient, error)
}

func newMockRepository() *mockRepository {
	return &mockRepository{
		patients:           make(map[string]domain.Patient),
		bannedUsers:        make(map[string]bool),
		appointmentHistory: make(map[string][]domain.Appointment),
	}
}

func (m *mockRepository) SavePatient(patient domain.Patient) error {
	if m.savePatientFunc != nil {
		return m.savePatientFunc(patient)
	}
	m.patients[patient.TelegramID] = patient
	return nil
}

func (m *mockRepository) GetPatient(telegramID string) (domain.Patient, error) {
	if m.getPatientFunc != nil {
		return m.getPatientFunc(telegramID)
	}
	if patient, ok := m.patients[telegramID]; ok {
		return patient, nil
	}
	return domain.Patient{TelegramID: telegramID}, nil
}

func (m *mockRepository) IsUserBanned(telegramID string, username string) (bool, error) {
	if m.isUserBannedFunc != nil {
		return m.isUserBannedFunc(telegramID, username)
	}
	return m.bannedUsers[telegramID], nil
}

func (m *mockRepository) BanUser(telegramID string) error {
	m.bannedUsers[telegramID] = true
	return nil
}

func (m *mockRepository) UnbanUser(telegramID string) error {
	delete(m.bannedUsers, telegramID)
	return nil
}

func (m *mockRepository) LogEvent(patientID string, eventType string, details map[string]interface{}) error {
	return nil
}

func (m *mockRepository) GenerateHTMLRecord(patient domain.Patient, history []domain.Appointment) string {
	return "<html>Mock Record</html>"
}

func (m *mockRepository) SavePatientDocumentReader(telegramID string, filename string, category string, r io.Reader) (string, error) {
	return "/tmp/mock-file.pdf", nil
}

func (m *mockRepository) CreateBackup() (string, error) {
	return "/tmp/backup.zip", nil
}

func (m *mockRepository) SyncAll() error {
	return nil
}

func (m *mockRepository) MigrateFolderNames() error {
	return nil
}

func (m *mockRepository) GetAppointmentHistory(telegramID string) ([]domain.Appointment, error) {
	if history, ok := m.appointmentHistory[telegramID]; ok {
		return history, nil
	}
	return []domain.Appointment{}, nil
}

func (m *mockRepository) UpsertAppointments(appts []domain.Appointment) error {
	return nil
}

func (m *mockRepository) SaveAppointmentMetadata(apptID string, confirmedAt *time.Time, remindersSent map[string]bool) error {
	return nil
}

func (m *mockRepository) GetAppointmentMetadata(apptID string) (*time.Time, map[string]bool, error) {
	return nil, make(map[string]bool), nil
}

// mockTranscriptionService implements ports.TranscriptionService
type mockTranscriptionService struct{}

func (m *mockTranscriptionService) Transcribe(ctx context.Context, audio io.Reader, filename string) (string, error) {
	return "Mock transcription", nil
}

// TestMockSetup tests that our mocks implement the required interfaces
func TestMockSetup(t *testing.T) {
	// Verify mocks implement interfaces
	var _ ports.AppointmentService = (*mockAppointmentService)(nil)
	var _ ports.SessionStorage = (*mockSessionStorage)(nil)
	var _ ports.Repository = (*mockRepository)(nil)
	var _ ports.TranscriptionService = (*mockTranscriptionService)(nil)
}

// TestSessionStorage tests the mock session storage
func TestSessionStorage(t *testing.T) {
	storage := newMockSessionStorage()
	userID := int64(123456789)

	// Test Set and Get
	storage.Set(userID, "test_key", "test_value")
	session := storage.Get(userID)

	if val, ok := session["test_key"]; !ok || val != "test_value" {
		t.Errorf("Expected test_value, got %v", val)
	}

	// Test ClearSession
	storage.ClearSession(userID)
	session = storage.Get(userID)

	if len(session) != 0 {
		t.Errorf("Expected empty session after clear, got %d items", len(session))
	}
}

// TestMockRepository tests the mock repository
func TestMockRepository(t *testing.T) {
	repo := newMockRepository()

	// Test SavePatient and GetPatient
	patient := domain.Patient{
		TelegramID:  "123456789",
		Name:        "Test Patient",
		TotalVisits: 5,
	}

	err := repo.SavePatient(patient)
	if err != nil {
		t.Errorf("SavePatient failed: %v", err)
	}

	retrieved, err := repo.GetPatient("123456789")
	if err != nil {
		t.Errorf("GetPatient failed: %v", err)
	}

	if retrieved.Name != patient.Name {
		t.Errorf("Expected name %s, got %s", patient.Name, retrieved.Name)
	}

	// Test BanUser and IsUserBanned
	err = repo.BanUser("123456789")
	if err != nil {
		t.Errorf("BanUser failed: %v", err)
	}

	banned, err := repo.IsUserBanned("123456789", "")
	if err != nil {
		t.Errorf("IsUserBanned failed: %v", err)
	}

	if !banned {
		t.Error("Expected user to be banned")
	}

	// Test UnbanUser
	err = repo.UnbanUser("123456789")
	if err != nil {
		t.Errorf("UnbanUser failed: %v", err)
	}

	banned, err = repo.IsUserBanned("123456789", "")
	if err != nil {
		t.Errorf("IsUserBanned failed: %v", err)
	}

	if banned {
		t.Error("Expected user to be unbanned")
	}
}

// TestMockAppointmentService tests the mock appointment service
func TestMockAppointmentService(t *testing.T) {
	service := &mockAppointmentService{
		getAvailableServicesFunc: func(ctx context.Context) ([]domain.Service, error) {
			return []domain.Service{
				{
					ID:              "massage-60",
					Name:            "Classic Massage",
					DurationMinutes: 60,
					Price:           50.0,
				},
			}, nil
		},
	}

	ctx := context.Background()
	services, err := service.GetAvailableServices(ctx)
	if err != nil {
		t.Errorf("GetAvailableServices failed: %v", err)
	}

	if len(services) != 1 {
		t.Errorf("Expected 1 service, got %d", len(services))
	}

	if services[0].Name != "Classic Massage" {
		t.Errorf("Expected Classic Massage, got %s", services[0].Name)
	}
}

// TestBotConfiguration tests bot configuration validation
// Note: We can't easily test StartBot without a real Telegram token,
// but we can test the configuration logic
func TestBotConfiguration(t *testing.T) {
	tests := []struct {
		name               string
		adminTelegramID    string
		allowedTelegramIDs []string
		wantAdminCount     int
	}{
		{
			name:               "Single admin",
			adminTelegramID:    "123456789",
			allowedTelegramIDs: []string{},
			wantAdminCount:     1,
		},
		{
			name:               "Admin with allowed IDs",
			adminTelegramID:    "123456789",
			allowedTelegramIDs: []string{"987654321", "111222333"},
			wantAdminCount:     3,
		},
		{
			name:               "Duplicate IDs filtered",
			adminTelegramID:    "123456789",
			allowedTelegramIDs: []string{"123456789", "987654321"},
			wantAdminCount:     2,
		},
		{
			name:               "Empty admin ID",
			adminTelegramID:    "",
			allowedTelegramIDs: []string{"987654321"},
			wantAdminCount:     1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate the admin ID deduplication logic from StartBot
			adminMap := make(map[string]bool)
			if tt.adminTelegramID != "" {
				adminMap[tt.adminTelegramID] = true
			}
			for _, id := range tt.allowedTelegramIDs {
				if id != "" {
					adminMap[id] = true
				}
			}

			finalAdminIDs := make([]string, 0, len(adminMap))
			for id := range adminMap {
				finalAdminIDs = append(finalAdminIDs, id)
			}

			if len(finalAdminIDs) != tt.wantAdminCount {
				t.Errorf("Expected %d admin IDs, got %d", tt.wantAdminCount, len(finalAdminIDs))
			}
		})
	}
}

// TestContextCancellation tests that context cancellation is handled
func TestContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	// Create a channel to signal when context is done
	done := make(chan bool)

	go func() {
		<-ctx.Done()
		done <- true
	}()

	// Cancel the context
	cancel()

	// Wait for done signal with timeout
	select {
	case <-done:
		// Success
	case <-time.After(1 * time.Second):
		t.Error("Context cancellation not detected within timeout")
	}
}
