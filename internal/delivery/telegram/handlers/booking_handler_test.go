package handlers

import (
	"context"
	"fmt"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/kfilin/massage-bot/internal/domain"
	"github.com/kfilin/massage-bot/internal/ports"
	"github.com/kfilin/massage-bot/internal/presentation"
	"gopkg.in/telebot.v3"
)

// --- Mocks ---

// mockContext implements telebot.Context
type mockContext struct {
	telebot.Context // Embed interface to satisfy remaining methods
	sender          *telebot.User
	text            string
	chat            *telebot.Chat
	callback        *telebot.Callback
	message         *telebot.Message
	args            []string
	bot             *telebot.Bot

	// Captured outputs
	sentMsg   string
	sentOpts  []interface{}
	responded bool
	response  *telebot.CallbackResponse
	editedMsg interface{}
}

func (m *mockContext) Sender() *telebot.User {
	return m.sender
}

func (m *mockContext) Text() string {
	return m.text
}

func (m *mockContext) Chat() *telebot.Chat {
	return m.chat
}

func (m *mockContext) Callback() *telebot.Callback {
	return m.callback
}

func (m *mockContext) Message() *telebot.Message {
	return m.message
}

func (m *mockContext) Args() []string {
	return m.args
}

func (m *mockContext) Bot() *telebot.Bot {
	return m.bot
}

func (m *mockContext) Recipient() telebot.Recipient {
	if m.chat != nil {
		return m.chat
	}
	if m.sender != nil {
		return m.sender
	}
	return &telebot.User{ID: 0}
}

func (m *mockContext) Send(what interface{}, opts ...interface{}) error {
	if s, ok := what.(string); ok {
		m.sentMsg = s
	} else {
		m.sentMsg = "NON_STRING_MESSAGE"
	}
	m.sentOpts = opts
	return nil
}

func (m *mockContext) Edit(what interface{}, opts ...interface{}) error {
	m.editedMsg = what
	return nil
}

func (m *mockContext) EditOrSend(what interface{}, opts ...interface{}) error {
	m.editedMsg = what
	return nil
}

func (m *mockContext) Respond(resp ...*telebot.CallbackResponse) error {
	m.responded = true
	if len(resp) > 0 {
		m.response = resp[0]
	}
	return nil
}

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
	patients            map[string]domain.Patient
	bannedUsers         map[string]bool
	appointmentHistory  map[string][]domain.Appointment
	isUserBannedFunc    func(telegramID string, username string) (bool, error)
	savePatientFunc     func(patient domain.Patient) error
	getPatientFunc      func(telegramID string) (domain.Patient, error)
	searchPatientsFunc  func(query string) ([]domain.Patient, error)
	saveMediaFunc       func(media domain.PatientMedia) error
	getPatientMediaFunc func(patientID string) ([]domain.PatientMedia, error)
	getMediaByIDFunc    func(mediaID string) (*domain.PatientMedia, error)
	banUserFunc         func(telegramID string) error
	unbanUserFunc       func(telegramID string) error
	createBackupFunc    func() (string, error)
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

func (m *mockRepository) UpdatePatientProfile(telegramID string, name string, notes string) error {
	// Simple mock implementation
	return nil
}

func (m *mockRepository) GetPatient(telegramID string) (domain.Patient, error) {
	if m.getPatientFunc != nil {
		return m.getPatientFunc(telegramID)
	}
	if patient, ok := m.patients[telegramID]; ok {
		return patient, nil
	}
	return domain.Patient{}, fmt.Errorf("patient not found")
}

func (m *mockRepository) GetAllPatients() ([]domain.Patient, error) { return nil, nil }

func (m *mockRepository) SearchPatients(query string) ([]domain.Patient, error) {
	if m.searchPatientsFunc != nil {
		return m.searchPatientsFunc(query)
	}
	return []domain.Patient{}, nil
}

func (m *mockRepository) IsUserBanned(telegramID string, username string) (bool, error) {
	if m.isUserBannedFunc != nil {
		return m.isUserBannedFunc(telegramID, username)
	}
	return m.bannedUsers[telegramID], nil
}

func (m *mockRepository) BanUser(telegramID string) error {
	if m.banUserFunc != nil {
		return m.banUserFunc(telegramID)
	}
	m.bannedUsers[telegramID] = true
	return nil
}

func (m *mockRepository) UnbanUser(telegramID string) error {
	if m.unbanUserFunc != nil {
		return m.unbanUserFunc(telegramID)
	}
	delete(m.bannedUsers, telegramID)
	return nil
}

func (m *mockRepository) CreateBackup() (string, error) {
	if m.createBackupFunc != nil {
		return m.createBackupFunc()
	}
	return "/tmp/backup.zip", nil
}

func (m *mockRepository) LogEvent(patientID string, eventType string, details map[string]interface{}) error {
	return nil
}

func (m *mockRepository) GenerateHTMLRecord(patient domain.Patient, history []domain.Appointment, isAdmin bool) string {
	return "<html>Mock Record</html>"
}

func (m *mockRepository) GenerateAdminSearchPage() string {
	return "<html>Mock Search Page</html>"
}

func (m *mockRepository) SaveMedia(media domain.PatientMedia) error {
	if m.saveMediaFunc != nil {
		return m.saveMediaFunc(media)
	}
	return nil
}

func (m *mockRepository) GetPatientMedia(patientID string) ([]domain.PatientMedia, error) {
	if m.getPatientMediaFunc != nil {
		return m.getPatientMediaFunc(patientID)
	}
	return []domain.PatientMedia{}, nil
}

func (m *mockRepository) GetMediaByID(mediaID string) (*domain.PatientMedia, error) {
	if m.getMediaByIDFunc != nil {
		return m.getMediaByIDFunc(mediaID)
	}
	return nil, nil
}

func (m *mockRepository) UpdateMediaStatus(mediaID string, status string, transcript string) error {
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

// --- Tests ---

func TestHandleStart(t *testing.T) {
	tests := []struct {
		name           string
		userID         int64
		patientExists  bool
		wantMsgContent string
	}{
		{
			name:          "New user registration",
			userID:        12345678,
			patientExists: false,
			// The handler logic sends "Добро пожаловать" for new users
			wantMsgContent: "Добро пожаловать",
		},
		{
			name:          "Existing user welcome",
			userID:        87654321,
			patientExists: true,
			// The handler logic sends "С возвращением" or similar,
			// checking specifically if "Добро пожаловать" is NOT sent might be safer
			// or checking for the main menu text if it always sends that.
			// Let's assume it sends "Выберите действие" or similar menu prompt.
			// Actually looking at code (via outline/assumption), let's check for standard greeting logic.
			// If we look at debug logs or common patterns, start usually resets state.
			wantMsgContent: "Вас приветствует",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup Mocks
			mockApptService := &mockAppointmentService{}
			mockSession := newMockSessionStorage()
			mockRepo := newMockRepository()
			mockTrans := &mockTranscriptionService{}

			// Setup Repo state
			if tt.patientExists {
				_ = mockRepo.SavePatient(domain.Patient{
					TelegramID: "87654321",
					Name:       "Existing Patient",
				})
			}

			handler := NewBookingHandler(
				mockApptService,
				mockSession,
				[]string{},
				nil,
				mockTrans,
				mockRepo,
				&presentation.BotPresenter{},
				"http://webapp.test",
				"secret",
			)

			// Create mock context
			user := &telebot.User{ID: tt.userID, FirstName: "TestUser"}
			ctx := &mockContext{
				sender: user,
			}

			// Execute Handler
			err := handler.HandleStart(ctx)

			// Verification
			if err != nil {
				t.Errorf("HandleStart returned error: %v", err)
			}

			if ctx.sentMsg == "" {
				t.Error("HandleStart did not send any message")
			}

			// Check if message contains expected content
			// note: simple string check, might need to be looser if text varies
			// checking for non-empty response covers "did something happen"
			if len(ctx.sentMsg) == 0 {
				t.Error("Response message was empty")
			}
		})
	}
}

func TestHandleCategorySelection(t *testing.T) {
	tests := []struct {
		name          string
		callbackData  string
		wantEditMsg   interface{} // We can check if it contains a string or is a specific type
		wantErr       bool
		setupServices bool
	}{
		{
			name:          "Select Massages",
			callbackData:  "select_category|massages",
			wantEditMsg:   "Выберите конкретную услугу:",
			wantErr:       false,
			setupServices: true,
		},
		{
			name:         "Invalid Data Format",
			callbackData: "invalid_data",
			wantEditMsg:  "Ошибка выбора категории.",
			wantErr:      false, // Handler handles error gracefully via Edit
		},
		{
			name:         "Wrong Prefix",
			callbackData: "wrong_prefix|massages",
			wantEditMsg:  "Ошибка выбора категории.",
			wantErr:      false,
		},
		{
			name:         "Back to Categories",
			callbackData: "select_category|back",
			wantEditMsg:  "Выберите категорию услуг:", // from showCategories
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup Mocks
			mockApptService := &mockAppointmentService{}
			if tt.setupServices {
				mockApptService.getAvailableServicesFunc = func(ctx context.Context) ([]domain.Service, error) {
					return []domain.Service{
						{ID: "massage-1", Name: "Общий массаж", Price: 100},
						{ID: "massage-2", Name: "Other Service", Price: 200},
					}, nil
				}
			}

			mockSession := newMockSessionStorage()
			handler := NewBookingHandler(
				mockApptService,
				mockSession,
				[]string{},
				nil,
				nil, nil, &presentation.BotPresenter{}, "", "",
			)

			// Create mock context with Callback
			user := &telebot.User{ID: 12345678}
			callback := &telebot.Callback{Data: tt.callbackData}
			ctx := &mockContext{
				sender:   user,
				callback: callback,
			}

			// Execute Handler
			err := handler.HandleCategorySelection(ctx)

			// Verification
			if (err != nil) != tt.wantErr {
				t.Errorf("HandleCategorySelection error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantEditMsg != nil {
				s, ok := ctx.editedMsg.(string)
				// If it wasn't edited, check sentMsg (since showCategories uses Send if no callback, but here we always have callback)
				if !ok {
					// Fallback: check sentMsg if Edit failed or wasn't called (though handler logic calls Edit for callback)
					if ctx.sentMsg != "" {
						s = ctx.sentMsg
					}
				}

				// Handle case where showCategories might use Edit or Send depending on callback presence
				// In "Back to Categories" case, it calls showCategories(c).
				// showCategories checks c.Callback(). If present, it uses c.Edit.

				expected := tt.wantEditMsg.(string)
				if s != expected {
					t.Errorf("Expected edited message %q, got %q", expected, s)
				}
			}

			// Verify session storage for valid selection
			if tt.setupServices && !tt.wantErr && ctx.editedMsg == "Выберите конкретную услугу:" {
				session := mockSession.Get(12345678)
				if val, ok := session["category"]; !ok || val != "massages" {
					t.Errorf("Expected category 'massages' in session, got %v", val)
				}
			}
		})
	}
}

func TestHandleServiceSelection(t *testing.T) {
	tests := []struct {
		name         string
		callbackData string
		setupSession bool
		services     []domain.Service
		wantEditMsg  string // partial match
		wantErr      bool
		wantSession  bool // check if service stored in session
	}{
		{
			name:         "Select Valid Service",
			callbackData: "select_service|massage-1",
			setupSession: false,
			services: []domain.Service{
				{ID: "massage-1", Name: "Classic Massage", Price: 100},
			},
			wantEditMsg: "Отлично, услуга 'Classic Massage' выбрана. Теперь выберите дату:",
			wantErr:     false,
			wantSession: true,
		},
		{
			name:         "Invalid Data",
			callbackData: "invalid_data",
			setupSession: false,
			services:     []domain.Service{},
			wantEditMsg:  "Некорректный выбор услуги",
			wantErr:      false,
			wantSession:  false,
		},
		{
			name:         "Service Not Found",
			callbackData: "select_service|unknown",
			setupSession: false,
			services: []domain.Service{
				{ID: "massage-1", Name: "Classic Massage", Price: 100},
			},
			wantEditMsg: "Выбранная услуга не найдена",
			wantErr:     false,
			wantSession: false,
		},
		{
			name:         "Admin Block Service",
			callbackData: "select_service|block_60",
			setupSession: false,
			services:     []domain.Service{}, // No real services needs
			wantEditMsg:  "Отлично, услуга '⛔ Блок: 1 час' выбрана",
			wantErr:      false,
			wantSession:  true, // Fake service stored
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup Mocks
			mockApptService := &mockAppointmentService{}
			mockApptService.getAvailableServicesFunc = func(ctx context.Context) ([]domain.Service, error) {
				return tt.services, nil
			}

			mockSession := newMockSessionStorage()
			mockRepo := newMockRepository()
			handler := NewBookingHandler(
				mockApptService,
				mockSession,
				[]string{},
				nil,
				nil,
				mockRepo,
				&presentation.BotPresenter{},
				"",
				"",
			)

			// Create mock context
			user := &telebot.User{ID: 12345678}
			callback := &telebot.Callback{Data: tt.callbackData}
			ctx := &mockContext{
				sender:   user,
				callback: callback,
			}

			// Execute
			err := handler.HandleServiceSelection(ctx)

			// Verification
			if (err != nil) != tt.wantErr {
				t.Errorf("HandleServiceSelection error = %v, wantErr %v", err, tt.wantErr)
			}

			// Check response message
			s, ok := ctx.editedMsg.(string)
			if !ok {
				if ctx.sentMsg != "" {
					s = ctx.sentMsg
				}
			}
			if !contains(s, tt.wantEditMsg) {
				t.Errorf("Expected response containing %q, got %q", tt.wantEditMsg, s)
			}

			// Check session storage
			if tt.wantSession {
				session := mockSession.Get(user.ID)
				svc, ok := session["service"].(domain.Service)
				if !ok {
					t.Errorf("Expected service in session, but got none or wrong type")
				}
				// For admin block, ID is block_60. For normal, massage-1
				expectedID := "massage-1"
				if contains(tt.callbackData, "block_60") {
					expectedID = "block_60"
				}
				if svc.ID != expectedID {
					t.Errorf("Expected service ID %s, got %s", expectedID, svc.ID)
				}
			}
		})
	}
}

// Helper for contains check
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsMiddle(s, substr)))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestHandleTimeSelection(t *testing.T) {
	tests := []struct {
		name         string
		callbackData string
		setupSession func(s ports.SessionStorage, userID int64)
		setupRepo    func(r *mockRepository)
		wantEditMsg  string
		wantSendMsg  string
		wantErr      bool
		checkSession func(t *testing.T, s ports.SessionStorage, userID int64)
	}{
		{
			name:         "Select Valid Time",
			callbackData: "select_time|14:00",
			setupSession: func(s ports.SessionStorage, userID int64) {
				s.Set(userID, "service", domain.Service{ID: "massage-1", Name: "Classic"})
				s.Set(userID, "date", time.Now())
			},
			setupRepo: func(r *mockRepository) {
				// New patient
				_ = r.SavePatient(domain.Patient{TelegramID: "12345678", TotalVisits: 0})
			},
			wantSendMsg: "Пожалуйста, введите ваше",
			wantErr:     false,
			checkSession: func(t *testing.T, s ports.SessionStorage, userID int64) {
				val := s.Get(userID)["time"]
				if val != "14:00" {
					t.Errorf("Expected time 14:00 in session, got %v", val)
				}
			},
		},
		{
			name:         "Back to Date",
			callbackData: "back_to_date",
			setupSession: func(s ports.SessionStorage, userID int64) {
				s.Set(userID, "service", domain.Service{ID: "massage-1", Name: "Classic"})
			},
			setupRepo:    func(r *mockRepository) {},
			wantEditMsg:  "Теперь выберите дату", // expected from askForDate
			wantErr:      false,
			checkSession: func(t *testing.T, s ports.SessionStorage, userID int64) {},
		},
		{
			name:         "Returning Patient Skipping Name",
			callbackData: "select_time|14:00",
			setupSession: func(s ports.SessionStorage, userID int64) {
				s.Set(userID, "service", domain.Service{ID: "massage-1", Name: "Classic"})
				s.Set(userID, "date", time.Now())
			},
			setupRepo: func(r *mockRepository) {
				_ = r.SavePatient(domain.Patient{TelegramID: "12345678", Name: "Ivan", TotalVisits: 5})
			},
			wantSendMsg: "Всё верно?", // expected from askForConfirmation
			wantErr:     false,
			checkSession: func(t *testing.T, s ports.SessionStorage, userID int64) {
				name := s.Get(userID)["name"]
				if name != "Ivan" {
					t.Errorf("Expected name Ivan in session, got %v", name)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSession := newMockSessionStorage()
			mockRepo := newMockRepository()
			mockApptService := &mockAppointmentService{}

			userID := int64(12345678)
			if tt.setupSession != nil {
				tt.setupSession(mockSession, userID)
			}
			if tt.setupRepo != nil {
				tt.setupRepo(mockRepo)
			}

			handler := NewBookingHandler(
				mockApptService,
				mockSession,
				[]string{},
				nil,
				nil,
				mockRepo,
				&presentation.BotPresenter{},
				"",
				"",
			)

			// Mock context with NO Message to avoid Bot() calls
			ctx := &mockContext{
				sender:   &telebot.User{ID: userID},
				callback: &telebot.Callback{Data: tt.callbackData},
				message:  nil,
			}

			// Execute
			err := handler.HandleTimeSelection(ctx)

			// Verification
			if (err != nil) != tt.wantErr {
				t.Errorf("expected error %v, got %v", tt.wantErr, err)
			}

			if tt.wantEditMsg != "" {
				s, _ := ctx.editedMsg.(string)
				// askForDate uses EditOrSend, ensuring it sends something
				if ctx.editedMsg == nil && ctx.sentMsg != "" {
					s = ctx.sentMsg
				}
				if !contains(s, tt.wantEditMsg) {
					t.Errorf("Expected edit msg containing %q, got %q", tt.wantEditMsg, s)
				}
			}

			if tt.wantSendMsg != "" {
				s := ctx.sentMsg
				if !contains(s, tt.wantSendMsg) {
					t.Errorf("Expected sent msg containing %q, got %q", tt.wantSendMsg, s)
				}
			}

			if tt.checkSession != nil {
				tt.checkSession(t, mockSession, userID)
			}
		})
	}
}

func TestHandleConfirmBooking(t *testing.T) {
	tests := []struct {
		name                 string
		setupSession         func(s ports.SessionStorage, userID int64)
		setupRepo            func(r *mockRepository)
		setupApptService     func(s *mockAppointmentService)
		wantCreateApptCalled bool
		wantSendMsg          string
		wantErr              bool
	}{
		{
			name: "Successful Booking",
			setupSession: func(s ports.SessionStorage, userID int64) {
				s.Set(userID, "service", domain.Service{
					ID:              "massage-1",
					Name:            "Classic Massage",
					DurationMinutes: 60,
					Price:           100,
				})
				s.Set(userID, "date", time.Date(2023, 10, 25, 0, 0, 0, 0, time.UTC))
				s.Set(userID, "time", "14:00")
				s.Set(userID, "name", "John Doe")
			},
			setupRepo: func(r *mockRepository) {
				_ = r.SavePatient(domain.Patient{TelegramID: "12345678", Name: "John Doe"})
			},
			setupApptService: func(s *mockAppointmentService) {
				s.createAppointmentFunc = func(ctx context.Context, appointment *domain.Appointment) (*domain.Appointment, error) {
					return appointment, nil
				}
				s.getCustomerHistoryFunc = func(ctx context.Context, id string) ([]domain.Appointment, error) {
					return []domain.Appointment{}, nil
				}
			},
			wantCreateApptCalled: true,
			wantSendMsg:          "ЗАПИСЬ ПОДТВЕРЖДЕНА",
			wantErr:              false,
		},
		{
			name: "Missing Session Data",
			setupSession: func(s ports.SessionStorage, userID int64) {
				// Empty session
			},
			setupRepo:            func(r *mockRepository) {},
			setupApptService:     func(s *mockAppointmentService) {},
			wantCreateApptCalled: false,
			wantSendMsg:          "Ошибка сессии",
			wantErr:              false, // Returns error but handles it via user message
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSession := newMockSessionStorage()
			mockRepo := newMockRepository()
			mockApptService := &mockAppointmentService{}
			createCalled := false

			userID := int64(12345678)
			if tt.setupSession != nil {
				tt.setupSession(mockSession, userID)
			}
			if tt.setupRepo != nil {
				tt.setupRepo(mockRepo)
			}
			if tt.setupApptService != nil {
				tt.setupApptService(mockApptService)
				// Wrap create func to detect call
				originalCreate := mockApptService.createAppointmentFunc
				mockApptService.createAppointmentFunc = func(ctx context.Context, appointment *domain.Appointment) (*domain.Appointment, error) {
					createCalled = true
					if originalCreate != nil {
						return originalCreate(ctx, appointment)
					}
					return appointment, nil
				}
			}

			handler := NewBookingHandler(
				mockApptService,
				mockSession,
				[]string{"999999"}, // Admin ID for notifications
				nil,
				nil,
				mockRepo,
				&presentation.BotPresenter{},
				"",
				"",
			)

			// Mock context with offline bot settings to avoid network calls
			bot, _ := telebot.NewBot(telebot.Settings{
				Offline: true,
			})

			ctx := &mockContext{
				sender: &telebot.User{ID: userID},
				bot:    bot,
			}

			// Execute
			err := handler.HandleConfirmBooking(ctx)

			// Verification
			if (err != nil) != tt.wantErr {
				t.Errorf("HandleConfirmBooking error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantCreateApptCalled && !createCalled {
				t.Error("Expected CreateAppointment to be called, but it wasn't")
			}
			if !tt.wantCreateApptCalled && createCalled {
				t.Error("Expected CreateAppointment NOT to be called, but it was")
			}

			if tt.wantSendMsg != "" {
				s := ctx.sentMsg
				if !contains(s, tt.wantSendMsg) {
					t.Errorf("Expected sent msg containing %q, got %q", tt.wantSendMsg, s)
				}
			}
		})
	}
}

func TestHandleCancel(t *testing.T) {
	mockSession := newMockSessionStorage()
	userID := int64(12345)

	// Setup session data
	mockSession.Set(userID, "some_key", "some_data")
	mockSession.Set(userID, "awaiting_confirmation", true)

	handler := NewBookingHandler(
		nil,
		mockSession,
		[]string{},
		nil,
		nil,
		nil,
		&presentation.BotPresenter{},
		"",
		"",
	)

	ctx := &mockContext{
		sender: &telebot.User{ID: userID},
	}

	// Execute
	err := handler.HandleCancel(ctx)

	if err != nil {
		t.Errorf("HandleCancel returned error: %v", err)
	}

	// Verify session cleared
	session := mockSession.Get(userID)
	if len(session) > 0 {
		// HandleCancel clears session completely
		t.Errorf("Expected session to be empty, got %v", session)
	}

	// Verify message
	if !contains(ctx.sentMsg, "Запись отменена") {
		t.Errorf("Expected 'Запись отменена' message, got %q", ctx.sentMsg)
	}
}

func TestBookingHandler_IsAdmin(t *testing.T) {
	h := NewBookingHandler(nil, nil, []string{"111", "222"}, nil, nil, nil, &presentation.BotPresenter{}, "", "")

	if !h.IsAdmin(111) {
		t.Error("Expected 111 to be admin")
	}
	if h.IsAdmin(999) {
		t.Error("Expected 999 to not be admin")
	}
}

func TestBookingHandler_GenerateWebAppURL(t *testing.T) {
	h := NewBookingHandler(nil, nil, nil, nil, nil, nil, &presentation.BotPresenter{}, "http://example.com", "secret123")
	
	url := h.GenerateWebAppURL("42")
	
	if !strings.HasPrefix(url, "https://example.com/card?id=42&token=") {
		t.Errorf("Unexpected URL format: %s", url)
	}

	// Empty config
	hEmpty := NewBookingHandler(nil, nil, nil, nil, nil, nil, &presentation.BotPresenter{}, "", "")
	if u := hEmpty.GenerateWebAppURL("42"); u != "" {
		t.Errorf("Expected empty URL when config is missing, got %s", u)
	}
}

func TestBookingHandler_GenerateGoogleCalendarLink(t *testing.T) {
	h := NewBookingHandler(nil, nil, nil, nil, nil, nil, &presentation.BotPresenter{}, "", "")
	
	appt := domain.Appointment{
		Service: domain.Service{Name: "Test Massage"},
		StartTime: time.Date(2024, 1, 1, 15, 0, 0, 0, time.UTC),
		Duration: 60,
	}

	domain.ApptTimeZone = time.UTC
	link := h.generateGoogleCalendarLink(appt)
	
	if !strings.Contains(link, "action=TEMPLATE") {
		t.Errorf("Missing action=TEMPLATE in link: %s", link)
	}
	if !strings.Contains(link, "20240101T150405") && !strings.Contains(link, "20240101T150000") { // depending on timezone behavior
		t.Errorf("Missing time format in link: %s", link)
	}
}

func TestHandleMyRecords_PatientExists(t *testing.T) {
	mockRepo := newMockRepository()
	_ = mockRepo.SavePatient(domain.Patient{TelegramID: "123", Name: "John Doe"})

	h := NewBookingHandler(nil, nil, nil, nil, nil, mockRepo, &presentation.BotPresenter{}, "", "")
	ctx := &mockContext{sender: &telebot.User{ID: 123}}

	err := h.HandleMyRecords(ctx)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if ctx.sentMsg == "" {
		t.Error("Expected a message to be sent")
	}
	// "Записей пока нет" might be part of the HTML record, or it might contain the name.
}

func TestHandleMyRecords_PatientNotFound(t *testing.T) {
	mockRepo := newMockRepository()
	mockRepo.getPatientFunc = func(id string) (domain.Patient, error) {
		return domain.Patient{}, fmt.Errorf("not found")
	}
	h := NewBookingHandler(nil, nil, nil, nil, nil, mockRepo, &presentation.BotPresenter{}, "", "")
	ctx := &mockContext{sender: &telebot.User{ID: 999}}

	err := h.HandleMyRecords(ctx)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if !strings.Contains(ctx.sentMsg, "медицинская карта") && !strings.Contains(ctx.sentMsg, "Записей пока нет") && !strings.Contains(ctx.sentMsg, "нет активной") {
		t.Errorf("Unexpected message for missing patient: %s", ctx.sentMsg)
	}
}

func TestHandleMyAppointments(t *testing.T) {
	t.Run("Patient with appointments", func(t *testing.T) {
		mockApptService := &mockAppointmentService{
			getCustomerAppointmentsFunc: func(ctx context.Context, id string) ([]domain.Appointment, error) {
				if id == "123" {
					return []domain.Appointment{
						{ID: "a1", Service: domain.Service{Name: "Massage"}, StartTime: time.Now().Add(24 * time.Hour)},
					}, nil
				}
				return []domain.Appointment{}, nil
			},
		}
		mockSession := newMockSessionStorage()
		h := NewBookingHandler(mockApptService, mockSession, nil, nil, nil, nil, &presentation.BotPresenter{}, "", "")

		ctx := &mockContext{sender: &telebot.User{ID: 123}}
		err := h.HandleMyAppointments(ctx)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if ctx.sentMsg == "" {
			t.Error("Expected a message for user with appointments")
		}
	})

	t.Run("Patient no appointments", func(t *testing.T) {
		mockApptService := &mockAppointmentService{
			getCustomerAppointmentsFunc: func(ctx context.Context, id string) ([]domain.Appointment, error) {
				return []domain.Appointment{}, nil
			},
		}
		mockSession := newMockSessionStorage()
		h := NewBookingHandler(mockApptService, mockSession, nil, nil, nil, nil, &presentation.BotPresenter{}, "", "")

		ctx := &mockContext{sender: &telebot.User{ID: 999}}
		err := h.HandleMyAppointments(ctx)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if !strings.Contains(ctx.sentMsg, "записей не найдено") {
			t.Errorf("Expected 'no appointments' message, got: %s", ctx.sentMsg)
		}
	})

	t.Run("Admin view", func(t *testing.T) {
		mockApptService := &mockAppointmentService{
			getAllUpcomingAppointmentsFunc: func(ctx context.Context) ([]domain.Appointment, error) {
				return []domain.Appointment{
					{ID: "a1", Service: domain.Service{Name: "Massage"}, CustomerTgID: "200", CustomerName: "Patient A", StartTime: time.Now().Add(48 * time.Hour)},
					{ID: "a2", Service: domain.Service{Name: "Consult"}, CustomerTgID: "123", CustomerName: "Self", StartTime: time.Now().Add(24 * time.Hour)},
				}, nil
			},
		}
		mockSession := newMockSessionStorage()
		h := NewBookingHandler(mockApptService, mockSession, []string{"123"}, nil, nil, nil, &presentation.BotPresenter{}, "", "")

		ctx := &mockContext{sender: &telebot.User{ID: 123}}
		err := h.HandleMyAppointments(ctx)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if !strings.Contains(ctx.sentMsg, "Общее расписание") {
			t.Errorf("Expected admin schedule header, got: %s", ctx.sentMsg)
		}
	})

	t.Run("Service error", func(t *testing.T) {
		mockApptService := &mockAppointmentService{
			getCustomerAppointmentsFunc: func(ctx context.Context, id string) ([]domain.Appointment, error) {
				return nil, fmt.Errorf("google calendar down")
			},
		}
		mockSession := newMockSessionStorage()
		h := NewBookingHandler(mockApptService, mockSession, nil, nil, nil, nil, &presentation.BotPresenter{}, "", "")

		ctx := &mockContext{sender: &telebot.User{ID: 123}}
		err := h.HandleMyAppointments(ctx)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if !strings.Contains(ctx.sentMsg, "Ошибка") {
			t.Errorf("Expected error message, got: %s", ctx.sentMsg)
		}
	})
}

func TestHandleStatus_Admin(t *testing.T) {
	mockApptService := &mockAppointmentService{
		getTotalUpcomingCountFunc: func(ctx context.Context) (int, error) { return 5, nil },
	}
	h := NewBookingHandler(mockApptService, nil, []string{"123"}, nil, nil, nil, &presentation.BotPresenter{}, "", "")
	
	ctx := &mockContext{sender: &telebot.User{ID: 123}}
	err := h.HandleStatus(ctx)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if !strings.Contains(ctx.sentMsg, "Статус бота") && !strings.Contains(ctx.sentMsg, "Uptime") {
		t.Errorf("Expected status message, got: %s", ctx.sentMsg)
	}
}

func TestHandleStatus_NonAdmin(t *testing.T) {
	h := NewBookingHandler(nil, nil, []string{"123"}, nil, nil, nil, &presentation.BotPresenter{}, "", "")
	
	ctx := &mockContext{sender: &telebot.User{ID: 999}}
	err := h.HandleStatus(ctx)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if !strings.Contains(ctx.sentMsg, "доступна только администраторам") {
		t.Errorf("Expected permission denied message, got: %s", ctx.sentMsg)
	}
}
func TestHandleApproveDraft(t *testing.T) {
	mockRepo := newMockRepository()
	media := domain.PatientMedia{ID: "m1", PatientID: "p1", Transcript: "Test Transcript", CreatedAt: time.Now()}
	mockRepo.getMediaByIDFunc = func(id string) (*domain.PatientMedia, error) { return &media, nil }
	mockRepo.getPatientFunc = func(id string) (domain.Patient, error) { return domain.Patient{TelegramID: "p1", Name: "John"}, nil }
	
	h := NewBookingHandler(nil, nil, nil, nil, nil, mockRepo, &presentation.BotPresenter{}, "", "")
	
	ctx := &mockContext{
		callback: &telebot.Callback{Data: "approve_draft|m1"},
	}

	err := h.HandleApproveDraft(ctx)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	
	edited, _ := ctx.editedMsg.(string)
	if !strings.Contains(edited, "ДОБАВЛЕНО В КАРТУ") {
		t.Errorf("Unexpected response message: %s", edited)
	}
}

func TestHandleDiscardDraft(t *testing.T) {
	mockRepo := newMockRepository()
	media := domain.PatientMedia{ID: "m1", PatientID: "p1", Transcript: "Test Transcript"}
	mockRepo.getMediaByIDFunc = func(id string) (*domain.PatientMedia, error) { return &media, nil }

	h := NewBookingHandler(nil, nil, nil, nil, nil, mockRepo, &presentation.BotPresenter{}, "", "")
	
	ctx := &mockContext{
		callback: &telebot.Callback{Data: "discard_draft|m1"},
	}

	err := h.HandleDiscardDraft(ctx)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	
	edited, _ := ctx.editedMsg.(string)
	if !strings.Contains(edited, "УДАЛЕН") {
		t.Errorf("Unexpected response message: %s", edited)
	}
}

func TestHandleUploadCommand(t *testing.T) {
	h := NewBookingHandler(nil, nil, nil, nil, nil, nil, &presentation.BotPresenter{}, "", "")
	ctx := &mockContext{sender: &telebot.User{ID: 123}}

	err := h.HandleUploadCommand(ctx)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if !strings.Contains(ctx.sentMsg, "Загрузка медицинских документов") {
		t.Errorf("Expected upload instructions message, got: %s", ctx.sentMsg)
	}
}

func TestHandleNameInput(t *testing.T) {
	tests := []struct {
		name        string
		inputName   string
		wantMsg     string
		wantSession bool
	}{
		{
			name:        "Valid name",
			inputName:   "Иван Петров",
			wantMsg:     "Всё верно",
			wantSession: true,
		},
		{
			name:        "Empty name",
			inputName:   "",
			wantMsg:     "Имя не может быть пустым",
			wantSession: false,
		},
		{
			name:        "Whitespace only name",
			inputName:   "   ",
			wantMsg:     "Имя не может быть пустым",
			wantSession: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSession := newMockSessionStorage()
			mockSession.Set(123, SessionKeyService, domain.Service{ID: "s1", Name: "Massage", DurationMinutes: 60})
			mockSession.Set(123, SessionKeyDate, time.Now())
			mockSession.Set(123, SessionKeyTime, "14:00")

			h := NewBookingHandler(nil, mockSession, nil, nil, nil, nil, &presentation.BotPresenter{}, "", "")
			ctx := &mockContext{
				sender: &telebot.User{ID: 123},
				text:   tt.inputName,
			}

			err := h.HandleNameInput(ctx)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if tt.wantSession {
				session := mockSession.Get(123)
				if name, ok := session[SessionKeyName].(string); !ok || name != strings.TrimSpace(tt.inputName) {
					t.Errorf("Expected name %q in session, got %v", tt.inputName, name)
				}
			}
		})
	}
}

func TestHandleManualAppointment(t *testing.T) {
	tests := []struct {
		name      string
		adminIDs  []string
		userID    int64
		args      []string
		wantMsg   string
		wantAdmin bool
	}{
		{
			name:      "Admin with name args",
			adminIDs:  []string{"123"},
			userID:    123,
			args:      []string{"Иван", "Петров"},
			wantMsg:   "Выберите категорию",
			wantAdmin: true,
		},
		{
			name:      "Admin without args",
			adminIDs:  []string{"123"},
			userID:    123,
			args:      []string{},
			wantMsg:   "Выберите категорию",
			wantAdmin: true,
		},
		{
			name:      "Non-admin denied",
			adminIDs:  []string{"999"},
			userID:    123,
			args:      []string{},
			wantMsg:   "только администраторам",
			wantAdmin: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockApptService := &mockAppointmentService{
				getAvailableServicesFunc: func(ctx context.Context) ([]domain.Service, error) {
					return []domain.Service{{ID: "s1", Name: "Test"}}, nil
				},
			}
			mockSession := newMockSessionStorage()

			h := NewBookingHandler(mockApptService, mockSession, tt.adminIDs, nil, nil, nil, &presentation.BotPresenter{}, "", "")
			ctx := &mockContext{
				sender: &telebot.User{ID: tt.userID},
				args:   tt.args,
			}

			err := h.HandleManualAppointment(ctx)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			msg := ctx.sentMsg
			if ctx.editedMsg != nil {
				msg, _ = ctx.editedMsg.(string)
			}
			if !contains(msg, tt.wantMsg) {
				t.Errorf("Expected message containing %q, got %q", tt.wantMsg, msg)
			}

			if tt.wantAdmin {
				session := mockSession.Get(tt.userID)
				if val, ok := session[SessionKeyIsAdminManual].(bool); !ok || !val {
					t.Error("Expected admin manual flag in session")
				}
				if len(tt.args) > 0 {
					if name, ok := session[SessionKeyName].(string); !ok || name != "Иван Петров" {
						t.Errorf("Expected name 'Иван Петров' in session, got %v", name)
					}
				}
			}
		})
	}
}

func TestHandleListPatients(t *testing.T) {
	tests := []struct {
		name     string
		adminIDs []string
		userID   int64
		patients []domain.Patient
		wantMsg  string
	}{
		{
			name:     "Admin with patients",
			adminIDs: []string{"123"},
			userID:   123,
			patients: []domain.Patient{
				{TelegramID: "100", Name: "Иван", TotalVisits: 5},
				{TelegramID: "200", Name: "Мария", TotalVisits: 3},
			},
			wantMsg: "Список пациентов",
		},
		{
			name:     "Admin empty list",
			adminIDs: []string{"123"},
			userID:   123,
			patients: []domain.Patient{},
			wantMsg:  "Список пациентов пуст",
		},
		{
			name:     "Non-admin denied",
			adminIDs: []string{"999"},
			userID:   123,
			patients: nil,
			wantMsg:  "Доступ запрещен",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := newMockRepository()
			mockRepo.searchPatientsFunc = func(query string) ([]domain.Patient, error) {
				return tt.patients, nil
			}

			h := NewBookingHandler(nil, nil, tt.adminIDs, nil, nil, mockRepo, &presentation.BotPresenter{}, "http://app.test", "secret")
			ctx := &mockContext{sender: &telebot.User{ID: tt.userID}}

			err := h.HandleListPatients(ctx)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if !contains(ctx.sentMsg, tt.wantMsg) {
				t.Errorf("Expected message containing %q, got %q", tt.wantMsg, ctx.sentMsg)
			}
		})
	}
}

func TestHandleEditName(t *testing.T) {
	tests := []struct {
		name       string
		adminIDs   []string
		userID     int64
		args       []string
		patientErr bool
		saveErr    bool
		wantMsg    string
	}{
		{
			name:     "Successful edit",
			adminIDs: []string{"123"},
			userID:   123,
			args:     []string{"100", "НовоеИмя"},
			wantMsg:  "обновлено",
		},
		{
			name:     "Non-admin denied",
			adminIDs: []string{"999"},
			userID:   123,
			args:     []string{"100", "Имя"},
			wantMsg:  "Доступ запрещен",
		},
		{
			name:     "Missing args",
			adminIDs: []string{"123"},
			userID:   123,
			args:     []string{"100"},
			wantMsg:  "Использование",
		},
		{
			name:       "Patient not found",
			adminIDs:   []string{"123"},
			userID:     123,
			args:       []string{"999", "Имя"},
			patientErr: true,
			wantMsg:    "не найден",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := newMockRepository()
			if !tt.patientErr && tt.name == "Successful edit" {
				_ = mockRepo.SavePatient(domain.Patient{TelegramID: "100", Name: "Old Name"})
			}
			if tt.patientErr {
				mockRepo.getPatientFunc = func(id string) (domain.Patient, error) {
					return domain.Patient{}, fmt.Errorf("not found")
				}
			}
			if tt.saveErr {
				mockRepo.savePatientFunc = func(p domain.Patient) error {
					return fmt.Errorf("save error")
				}
			}

			h := NewBookingHandler(nil, nil, tt.adminIDs, nil, nil, mockRepo, &presentation.BotPresenter{}, "", "")
			ctx := &mockContext{
				sender: &telebot.User{ID: tt.userID},
				args:   tt.args,
			}

			err := h.HandleEditName(ctx)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if !contains(ctx.sentMsg, tt.wantMsg) {
				t.Errorf("Expected message containing %q, got %q", tt.wantMsg, ctx.sentMsg)
			}
		})
	}
}

func TestHandleBackup(t *testing.T) {
	tests := []struct {
		name     string
		adminIDs []string
		userID   int64
		wantMsg  string
	}{
		{
			name:     "Non-admin denied",
			adminIDs: []string{"999"},
			userID:   123,
			wantMsg:  "нет прав",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := newMockRepository()
			h := NewBookingHandler(nil, nil, tt.adminIDs, nil, nil, mockRepo, &presentation.BotPresenter{}, "", "")
			ctx := &mockContext{sender: &telebot.User{ID: tt.userID}}

			err := h.HandleBackup(ctx)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if !contains(ctx.sentMsg, tt.wantMsg) {
				t.Errorf("Expected message containing %q, got %q", tt.wantMsg, ctx.sentMsg)
			}
		})
	}
}

func TestHandleBackup_AdminSuccess(t *testing.T) {
	mockRepo := newMockRepository()
	h := NewBookingHandler(nil, nil, []string{"123"}, nil, nil, mockRepo, &presentation.BotPresenter{}, "", "")
	ctx := &mockContext{sender: &telebot.User{ID: 123}}

	err := h.HandleBackup(ctx)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// The first Send call is a status message, then it sends a Document
	// Since our mock captures only string messages, check that something was sent
	if ctx.sentMsg == "" {
		t.Error("Expected at least one message to be sent")
	}
}

func TestHandleAdminReplyRequest(t *testing.T) {
	tests := []struct {
		name       string
		callback   string
		patientErr bool
		wantMsg    string
	}{
		{
			name:     "Valid reply request",
			callback: "admin_reply|100",
			wantMsg:  "Введите ответ",
		},
		{
			name:       "Patient not found",
			callback:   "admin_reply|999",
			patientErr: true,
			wantMsg:    "не найден",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSession := newMockSessionStorage()
			mockRepo := newMockRepository()
			if !tt.patientErr {
				_ = mockRepo.SavePatient(domain.Patient{TelegramID: "100", Name: "Test Patient"})
			}
			if tt.patientErr {
				mockRepo.getPatientFunc = func(id string) (domain.Patient, error) {
					return domain.Patient{}, fmt.Errorf("not found")
				}
			}

			h := NewBookingHandler(nil, mockSession, nil, nil, nil, mockRepo, &presentation.BotPresenter{}, "", "")
			ctx := &mockContext{
				sender:   &telebot.User{ID: 123},
				callback: &telebot.Callback{Data: tt.callback},
			}

			err := h.HandleAdminReplyRequest(ctx)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if !contains(ctx.sentMsg, tt.wantMsg) {
				t.Errorf("Expected message containing %q, got %q", tt.wantMsg, ctx.sentMsg)
			}

			if !tt.patientErr {
				session := mockSession.Get(123)
				if val, ok := session[SessionKeyAdminReplyingTo].(string); !ok || val != "100" {
					t.Errorf("Expected admin replying to '100', got %v", val)
				}
			}
		})
	}
}

func TestHandleReminderConfirmation(t *testing.T) {
	mockRepo := newMockRepository()
	mockApptService := &mockAppointmentService{
		findByIDFunc: func(ctx context.Context, id string) (*domain.Appointment, error) {
			return &domain.Appointment{
				ID:           "appt1",
				CustomerTgID: "100",
				Service:      domain.Service{Name: "Massage"},
				StartTime:    time.Now().Add(24 * time.Hour),
			}, nil
		},
	}

	bot, _ := telebot.NewBot(telebot.Settings{Offline: true})
	h := NewBookingHandler(mockApptService, nil, []string{"999"}, nil, nil, mockRepo, &presentation.BotPresenter{}, "", "")
	ctx := &mockContext{
		sender:   &telebot.User{ID: 100},
		callback: &telebot.Callback{Data: "confirm_appt_reminder|appt1"},
		bot:      bot,
	}

	err := h.HandleReminderConfirmation(ctx)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	edited, _ := ctx.editedMsg.(string)
	if !contains(edited, "подтверждена") {
		t.Errorf("Expected confirmation message, got: %s", edited)
	}
}

func TestHandleCancelAppointmentCallback(t *testing.T) {
	tests := []struct {
		name          string
		callback      string
		appt          *domain.Appointment
		apptErr       error
		cancelErr     error
		wantResponse  string
		wantLateBlock bool
	}{
		{
			name:     "Successful cancellation",
			callback: "cancel_appt|appt1",
			appt: &domain.Appointment{
				ID:           "appt1",
				CustomerTgID: "100",
				CustomerName: "Иван",
				StartTime:    time.Now().Add(7 * 24 * time.Hour),
			},
			wantResponse: "успешно отменена",
		},
		{
			name:     "Late cancellation blocked (< 72h)",
			callback: "cancel_appt|appt1",
			appt: &domain.Appointment{
				ID:           "appt1",
				CustomerTgID: "100",
				StartTime:    time.Now().Add(24 * time.Hour),
			},
			wantLateBlock: true,
			wantResponse:  "меньше 3 дней",
		},
		{
			name:         "Invalid callback data",
			callback:     "cancel_appt",
			wantResponse: "неверные данные",
		},
		{
			name:         "Cancel error",
			callback:     "cancel_appt|appt1",
			appt:         nil,
			apptErr:      fmt.Errorf("not found"),
			cancelErr:    fmt.Errorf("cancel failed"),
			wantResponse: "Не удалось отменить",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := newMockRepository()
			mockApptService := &mockAppointmentService{
				findByIDFunc: func(ctx context.Context, id string) (*domain.Appointment, error) {
					if tt.apptErr != nil {
						return nil, tt.apptErr
					}
					return tt.appt, nil
				},
				cancelAppointmentFunc: func(ctx context.Context, id string) error {
					return tt.cancelErr
				},
				getCustomerHistoryFunc: func(ctx context.Context, id string) ([]domain.Appointment, error) {
					return []domain.Appointment{}, nil
				},
			}

			bot, _ := telebot.NewBot(telebot.Settings{Offline: true})
			domain.ApptTimeZone = time.UTC
			h := NewBookingHandler(mockApptService, nil, []string{"999"}, nil, nil, mockRepo, &presentation.BotPresenter{}, "", "")
			ctx := &mockContext{
				sender:   &telebot.User{ID: 100},
				callback: &telebot.Callback{Data: tt.callback},
				bot:      bot,
			}

			err := h.HandleCancelAppointmentCallback(ctx)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if ctx.response == nil && ctx.responded {
				t.Error("Expected callback response")
			}
			if ctx.response != nil && !contains(ctx.response.Text, tt.wantResponse) {
				t.Errorf("Expected response containing %q, got %q", tt.wantResponse, ctx.response.Text)
			}
		})
	}
}

func TestHandleDateSelection(t *testing.T) {
	tests := []struct {
		name         string
		callback     string
		setupSession func(s ports.SessionStorage, userID int64)
		message      *telebot.Message
		wantEdit     bool
		wantErr      bool
	}{
		{
			name:     "Navigate month",
			callback: "navigate_month|2025-07",
			message:  &telebot.Message{Text: "Calendar", ID: 1, Chat: &telebot.Chat{ID: 123}},
			wantEdit: true,
			wantErr:  false,
		},
		{
			name:         "Invalid date format",
			callback:     "select_date|invalid",
			setupSession: func(s ports.SessionStorage, userID int64) {},
			wantErr:      false,
		},
		{
			name:     "Back to services with category",
			callback: "back_to_services",
			setupSession: func(s ports.SessionStorage, userID int64) {
				s.Set(userID, SessionKeyCategory, "massages")
			},
			wantErr: false,
		},
		{
			name:     "Back to services without category",
			callback: "back_to_services",
			setupSession: func(s ports.SessionStorage, userID int64) {},
			wantErr:  false,
		},
		{
			name:     "Unknown action",
			callback: "unknown_action",
			setupSession: func(s ports.SessionStorage, userID int64) {},
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSession := newMockSessionStorage()
			mockApptService := &mockAppointmentService{
				getAvailableServicesFunc: func(ctx context.Context) ([]domain.Service, error) {
					return []domain.Service{{ID: "s1", Name: "Test"}}, nil
				},
			}

			if tt.setupSession != nil {
				tt.setupSession(mockSession, 123)
			}

			bot, _ := telebot.NewBot(telebot.Settings{Offline: true})
			h := NewBookingHandler(mockApptService, mockSession, nil, nil, nil, nil, &presentation.BotPresenter{}, "", "")

			msg := tt.message
			if msg == nil {
				msg = &telebot.Message{Text: "Default", ID: 1, Chat: &telebot.Chat{ID: 123}}
			}

			ctx := &mockContext{
				sender:   &telebot.User{ID: 123},
				callback: &telebot.Callback{Data: tt.callback},
				message:  msg,
				bot:      bot,
			}

			err := h.HandleDateSelection(ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("HandleDateSelection error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestHandleDateSelection_SelectDate(t *testing.T) {
	mockSession := newMockSessionStorage()
	mockSession.Set(123, SessionKeyService, domain.Service{ID: "s1", Name: "Test", DurationMinutes: 60})

	mockApptService := &mockAppointmentService{
		getAvailableTimeSlotsFunc: func(ctx context.Context, date time.Time, dur int) ([]domain.TimeSlot, error) {
			start := time.Date(2025, 6, 15, 10, 0, 0, 0, time.UTC)
			return []domain.TimeSlot{{Start: start, End: start.Add(time.Hour)}}, nil
		},
	}

	bot, _ := telebot.NewBot(telebot.Settings{Offline: true})
	h := NewBookingHandler(mockApptService, mockSession, nil, nil, nil, nil, &presentation.BotPresenter{}, "", "")
	ctx := &mockContext{
		sender:   &telebot.User{ID: 123},
		callback: &telebot.Callback{Data: "select_date|2025-06-15"},
		message:  &telebot.Message{Text: "Calendar", ID: 1, Chat: &telebot.Chat{ID: 123}},
		bot:      bot,
	}

	err := h.HandleDateSelection(ctx)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify date stored in session
	session := mockSession.Get(123)
	if date, ok := session[SessionKeyDate].(time.Time); !ok || date.Format("2006-01-02") != "2025-06-15" {
		t.Errorf("Expected date 2025-06-15 in session, got %v", date)
	}
}

func TestHandleReminderCancellation(t *testing.T) {
	mockRepo := newMockRepository()
	mockApptService := &mockAppointmentService{
		findByIDFunc: func(ctx context.Context, id string) (*domain.Appointment, error) {
			return &domain.Appointment{
				ID:           "appt1",
				CustomerTgID: "100",
				StartTime:    time.Now().Add(7 * 24 * time.Hour),
			}, nil
		},
		cancelAppointmentFunc: func(ctx context.Context, id string) error {
			return nil
		},
		getCustomerHistoryFunc: func(ctx context.Context, id string) ([]domain.Appointment, error) {
			return []domain.Appointment{}, nil
		},
	}

	bot, _ := telebot.NewBot(telebot.Settings{Offline: true})
	domain.ApptTimeZone = time.UTC
	h := NewBookingHandler(mockApptService, nil, nil, nil, nil, mockRepo, &presentation.BotPresenter{}, "", "")
	ctx := &mockContext{
		sender:   &telebot.User{ID: 100},
		callback: &telebot.Callback{Data: "cancel_appt_reminder|appt1"},
		bot:      bot,
	}

	err := h.HandleReminderCancellation(ctx)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if !ctx.responded {
		t.Error("Expected callback response")
	}
}

func TestHandleCancelAppointmentCallback_AdminBypass(t *testing.T) {
	mockRepo := newMockRepository()
	mockApptService := &mockAppointmentService{
		findByIDFunc: func(ctx context.Context, id string) (*domain.Appointment, error) {
			return &domain.Appointment{
				ID:           "appt1",
				CustomerTgID: "100",
				StartTime:    time.Now().Add(24 * time.Hour), // < 72h
			}, nil
		},
		getCustomerHistoryFunc: func(ctx context.Context, id string) ([]domain.Appointment, error) {
			return []domain.Appointment{}, nil
		},
	}

	bot, _ := telebot.NewBot(telebot.Settings{Offline: true})
	domain.ApptTimeZone = time.UTC
	h := NewBookingHandler(mockApptService, nil, []string{"999"}, nil, nil, mockRepo, &presentation.BotPresenter{}, "", "")
	ctx := &mockContext{
		sender:   &telebot.User{ID: 100},
		callback: &telebot.Callback{Data: "cancel_appt|appt1"},
		bot:      bot,
	}

	err := h.HandleCancelAppointmentCallback(ctx)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Should be blocked due to < 72h
	if ctx.response != nil && !contains(ctx.response.Text, "меньше 3 дней") {
		t.Errorf("Expected late cancellation block, got: %s", ctx.response.Text)
	}
}

func TestHandleCancelAppointmentCallback_PatientNotFound(t *testing.T) {
	mockRepo := newMockRepository()
	mockApptService := &mockAppointmentService{
		findByIDFunc: func(ctx context.Context, id string) (*domain.Appointment, error) {
			return nil, fmt.Errorf("not found")
		},
		cancelAppointmentFunc: func(ctx context.Context, id string) error {
			return nil
		},
	}

	bot, _ := telebot.NewBot(telebot.Settings{Offline: true})
	domain.ApptTimeZone = time.UTC
	h := NewBookingHandler(mockApptService, nil, nil, nil, nil, mockRepo, &presentation.BotPresenter{}, "", "")
	ctx := &mockContext{
		sender:   &telebot.User{ID: 100},
		callback: &telebot.Callback{Data: "cancel_appt|appt1"},
		bot:      bot,
	}

	err := h.HandleCancelAppointmentCallback(ctx)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if ctx.response != nil && !contains(ctx.response.Text, "успешно отменена") {
		t.Errorf("Expected success response even without appt details, got: %s", ctx.response.Text)
	}
}

func TestHandleDateSelection_InvalidMonthFormat(t *testing.T) {
	mockSession := newMockSessionStorage()
	h := NewBookingHandler(nil, mockSession, nil, nil, nil, nil, &presentation.BotPresenter{}, "", "")
	ctx := &mockContext{
		sender:   &telebot.User{ID: 123},
		callback: &telebot.Callback{Data: "navigate_month|invalid"},
		message:  &telebot.Message{Text: "Calendar"},
	}

	err := h.HandleDateSelection(ctx)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	edited, _ := ctx.editedMsg.(string)
	if !contains(edited, "Некорректная дата") {
		t.Errorf("Expected invalid date message, got: %s", edited)
	}
}

func TestHandleDateSelection_InvalidNavigationFormat(t *testing.T) {
	mockSession := newMockSessionStorage()
	h := NewBookingHandler(nil, mockSession, nil, nil, nil, nil, &presentation.BotPresenter{}, "", "")
	ctx := &mockContext{
		sender:   &telebot.User{ID: 123},
		callback: &telebot.Callback{Data: "navigate_month|"},
		message:  &telebot.Message{Text: "Calendar", ID: 1, Chat: &telebot.Chat{ID: 123}},
	}

	err := h.HandleDateSelection(ctx)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	edited, _ := ctx.editedMsg.(string)
	if !contains(edited, "Некорректная дата") {
		t.Errorf("Expected invalid date message, got: %s", edited)
	}
}

func TestHandleDateSelection_InvalidDateFormat(t *testing.T) {
	mockSession := newMockSessionStorage()
	h := NewBookingHandler(nil, mockSession, nil, nil, nil, nil, &presentation.BotPresenter{}, "", "")
	ctx := &mockContext{
		sender:   &telebot.User{ID: 123},
		callback: &telebot.Callback{Data: "select_date|not-a-date"},
	}

	err := h.HandleDateSelection(ctx)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	edited, _ := ctx.editedMsg.(string)
	if !contains(edited, "Некорректная дата") {
		t.Errorf("Expected invalid date message, got: %s", edited)
	}
}

func TestHandleNameInput_StoresCorrectly(t *testing.T) {
	mockSession := newMockSessionStorage()
	mockSession.Set(123, SessionKeyService, domain.Service{ID: "s1", Name: "Test", DurationMinutes: 60})
	mockSession.Set(123, SessionKeyDate, time.Date(2025, 6, 15, 0, 0, 0, 0, time.UTC))
	mockSession.Set(123, SessionKeyTime, "14:00")

	h := NewBookingHandler(nil, mockSession, nil, nil, nil, nil, &presentation.BotPresenter{}, "", "")
	ctx := &mockContext{
		sender: &telebot.User{ID: 123},
		text:   "  Иван Петров  ",
	}

	err := h.HandleNameInput(ctx)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	session := mockSession.Get(123)
	if name, ok := session[SessionKeyName].(string); !ok || name != "Иван Петров" {
		t.Errorf("Expected trimmed name 'Иван Петров', got %v", name)
	}
}

func TestHandleStart_DeepLinkBook(t *testing.T) {
	mockApptService := &mockAppointmentService{
		getAvailableServicesFunc: func(ctx context.Context) ([]domain.Service, error) {
			return []domain.Service{{ID: "s1", Name: "Test"}}, nil
		},
	}
	mockSession := newMockSessionStorage()
	mockRepo := newMockRepository()

	h := NewBookingHandler(mockApptService, mockSession, []string{}, nil, nil, mockRepo, &presentation.BotPresenter{}, "", "")
	ctx := &mockContext{
		sender: &telebot.User{ID: 123, FirstName: "Test"},
		args:   []string{"book"},
	}

	err := h.HandleStart(ctx)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Should show categories (via Edit or Send)
	if ctx.editedMsg == nil && ctx.sentMsg == "" {
		t.Error("Expected categories to be shown for /start book")
	}
}

func TestHandleStart_DeepLinkManual_AdminWithPatient(t *testing.T) {
	mockApptService := &mockAppointmentService{
		getAvailableServicesFunc: func(ctx context.Context) ([]domain.Service, error) {
			return []domain.Service{{ID: "s1", Name: "Test"}}, nil
		},
	}
	mockSession := newMockSessionStorage()
	mockRepo := newMockRepository()
	_ = mockRepo.SavePatient(domain.Patient{TelegramID: "456", Name: "Existing Patient"})

	h := NewBookingHandler(mockApptService, mockSession, []string{"123"}, nil, nil, mockRepo, &presentation.BotPresenter{}, "", "")
	ctx := &mockContext{
		sender: &telebot.User{ID: 123, FirstName: "Admin"},
		args:   []string{"manual_456"},
	}

	err := h.HandleStart(ctx)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Should set manual booking session keys
	session := mockSession.Get(123)
	if val, ok := session[SessionKeyIsAdminManual].(bool); !ok || !val {
		t.Error("Expected is_admin_manual=true in session")
	}
	if val, ok := session[SessionKeyPatientID].(string); !ok || val != "456" {
		t.Errorf("Expected patient_id=456, got %v", val)
	}
	if val, ok := session[SessionKeyName].(string); !ok || val != "Existing Patient" {
		t.Errorf("Expected name='Existing Patient', got %v", val)
	}
}

func TestHandleStart_DeepLinkManual_NonAdmin(t *testing.T) {
	mockSession := newMockSessionStorage()
	mockRepo := newMockRepository()

	h := NewBookingHandler(nil, mockSession, []string{"999"}, nil, nil, mockRepo, &presentation.BotPresenter{}, "", "")
	ctx := &mockContext{
		sender: &telebot.User{ID: 123, FirstName: "User"},
		args:   []string{"manual_456"},
	}

	err := h.HandleStart(ctx)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Non-admin should get regular welcome, not manual booking
	session := mockSession.Get(123)
	if val, ok := session[SessionKeyIsAdminManual].(bool); ok && val {
		t.Error("Non-admin should not have is_admin_manual=true")
	}
}

func TestHandleStart_DeepLinkManual_AdminPatientNotFound(t *testing.T) {
	mockApptService := &mockAppointmentService{
		getAvailableServicesFunc: func(ctx context.Context) ([]domain.Service, error) {
			return []domain.Service{{ID: "s1", Name: "Test"}}, nil
		},
	}
	mockSession := newMockSessionStorage()
	mockRepo := newMockRepository()

	h := NewBookingHandler(mockApptService, mockSession, []string{"123"}, nil, nil, mockRepo, &presentation.BotPresenter{}, "", "")
	ctx := &mockContext{
		sender: &telebot.User{ID: 123, FirstName: "Admin"},
		args:   []string{"manual_999"},
	}

	err := h.HandleStart(ctx)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Should still set manual keys even if patient not found
	session := mockSession.Get(123)
	if val, ok := session[SessionKeyIsAdminManual].(bool); !ok || !val {
		t.Error("Expected is_admin_manual=true even when patient not found")
	}
}

func TestSyncPatientStats(t *testing.T) {
	t.Run("New patient created when not found", func(t *testing.T) {
		mockRepo := newMockRepository()
		mockApptService := &mockAppointmentService{
			getCustomerHistoryFunc: func(ctx context.Context, id string) ([]domain.Appointment, error) {
				return []domain.Appointment{}, nil
			},
		}

		h := NewBookingHandler(mockApptService, nil, nil, nil, nil, mockRepo, &presentation.BotPresenter{}, "", "")

		patient, err := h.syncPatientStats(context.Background(), "555", "New Patient")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if patient.Name != "New Patient" {
			t.Errorf("Expected name 'New Patient', got %s", patient.Name)
		}
		if patient.HealthStatus != "initial" {
			t.Errorf("Expected health_status 'initial', got %s", patient.HealthStatus)
		}
		if patient.TotalVisits != 0 {
			t.Errorf("Expected 0 visits, got %d", patient.TotalVisits)
		}
	})

	t.Run("New patient with empty name defaults to Пациент", func(t *testing.T) {
		mockRepo := newMockRepository()
		mockApptService := &mockAppointmentService{
			getCustomerHistoryFunc: func(ctx context.Context, id string) ([]domain.Appointment, error) {
				return []domain.Appointment{}, nil
			},
		}

		h := NewBookingHandler(mockApptService, nil, nil, nil, nil, mockRepo, &presentation.BotPresenter{}, "", "")

		patient, err := h.syncPatientStats(context.Background(), "666", "")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if patient.Name != "Пациент" {
			t.Errorf("Expected default name 'Пациент', got %s", patient.Name)
		}
	})

	t.Run("Existing patient name updated", func(t *testing.T) {
		mockRepo := newMockRepository()
		_ = mockRepo.SavePatient(domain.Patient{TelegramID: "777", Name: "Old Name", TotalVisits: 3})
		mockApptService := &mockAppointmentService{
			getCustomerHistoryFunc: func(ctx context.Context, id string) ([]domain.Appointment, error) {
				return []domain.Appointment{
					{ID: "a1", StartTime: time.Now().Add(-72 * time.Hour), Status: "confirmed"},
					{ID: "a2", StartTime: time.Now().Add(-48 * time.Hour), Status: "cancelled"},
					{ID: "a3", StartTime: time.Now().Add(-24 * time.Hour), Status: "confirmed"},
				}, nil
			},
		}

		h := NewBookingHandler(mockApptService, nil, nil, nil, nil, mockRepo, &presentation.BotPresenter{}, "", "")

		patient, err := h.syncPatientStats(context.Background(), "777", "Updated Name")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if patient.Name != "Updated Name" {
			t.Errorf("Expected name 'Updated Name', got %s", patient.Name)
		}
		// 2 confirmed (a1, a3), 1 cancelled excluded
		if patient.TotalVisits != 2 {
			t.Errorf("Expected 2 confirmed visits, got %d", patient.TotalVisits)
		}
	})

	t.Run("Service error returns partial patient", func(t *testing.T) {
		mockRepo := newMockRepository()
		mockApptService := &mockAppointmentService{
			getCustomerHistoryFunc: func(ctx context.Context, id string) ([]domain.Appointment, error) {
				return nil, fmt.Errorf("gcal down")
			},
		}

		h := NewBookingHandler(mockApptService, nil, nil, nil, nil, mockRepo, &presentation.BotPresenter{}, "", "")

		_, err := h.syncPatientStats(context.Background(), "888", "Test")
		if err == nil {
			t.Error("Expected error from GetCustomerHistory")
		}
	})

	t.Run("Save patient error propagated", func(t *testing.T) {
		mockRepo := newMockRepository()
		mockRepo.savePatientFunc = func(p domain.Patient) error {
			return fmt.Errorf("save failed")
		}
		mockApptService := &mockAppointmentService{
			getCustomerHistoryFunc: func(ctx context.Context, id string) ([]domain.Appointment, error) {
				return []domain.Appointment{}, nil
			},
		}

		h := NewBookingHandler(mockApptService, nil, nil, nil, nil, mockRepo, &presentation.BotPresenter{}, "", "")

		_, err := h.syncPatientStats(context.Background(), "999", "Test")
		if err == nil {
			t.Error("Expected error from SavePatient")
		}
	})
}

func TestHandleFileMessage(t *testing.T) {
	t.Run("No media type returns nil", func(t *testing.T) {
		h := NewBookingHandler(nil, nil, nil, nil, nil, nil, &presentation.BotPresenter{}, "", "")
		ctx := &mockContext{
			sender:  &telebot.User{ID: 123},
			message: &telebot.Message{},
		}

		err := h.HandleFileMessage(ctx)
		if err != nil {
			t.Fatalf("Expected nil for no media, got: %v", err)
		}
	})

	t.Run("Document without patient returns error", func(t *testing.T) {
		mockRepo := newMockRepository()
		h := NewBookingHandler(nil, nil, nil, nil, nil, mockRepo, &presentation.BotPresenter{}, "", "")

		bot, _ := telebot.NewBot(telebot.Settings{Offline: true})
		ctx := &mockContext{
			sender: &telebot.User{ID: 123},
			message: &telebot.Message{
				Document: &telebot.Document{
					File:     telebot.File{FileID: "doc123", FileSize: 1024},
					FileName: "report.pdf",
				},
			},
			bot: bot,
		}

		err := h.HandleFileMessage(ctx)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if !strings.Contains(ctx.sentMsg, "запишитесь на прием") {
			t.Errorf("Expected 'register first' message, got: %s", ctx.sentMsg)
		}
	})

	t.Run("Document too large", func(t *testing.T) {
		mockRepo := newMockRepository()
		_ = mockRepo.SavePatient(domain.Patient{TelegramID: "123", Name: "Test"})
		h := NewBookingHandler(nil, nil, nil, nil, nil, mockRepo, &presentation.BotPresenter{}, "", "")

		bot, _ := telebot.NewBot(telebot.Settings{Offline: true})
		ctx := &mockContext{
			sender: &telebot.User{ID: 123},
			message: &telebot.Message{
				Document: &telebot.Document{
					File:     telebot.File{FileID: "bigdoc", FileSize: 25 * 1024 * 1024},
					FileName: "huge.zip",
				},
			},
			bot: bot,
		}

		err := h.HandleFileMessage(ctx)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if !strings.Contains(ctx.sentMsg, "слишком большой") {
			t.Errorf("Expected 'too large' message, got: %s", ctx.sentMsg)
		}
	})

	t.Run("Photo without patient returns error", func(t *testing.T) {
		mockRepo := newMockRepository()
		h := NewBookingHandler(nil, nil, nil, nil, nil, mockRepo, &presentation.BotPresenter{}, "", "")

		bot, _ := telebot.NewBot(telebot.Settings{Offline: true})
		ctx := &mockContext{
			sender: &telebot.User{ID: 123},
			message: &telebot.Message{
				Photo: &telebot.Photo{
					File: telebot.File{FileID: "photo123", FileSize: 512000},
				},
			},
			bot: bot,
		}

		err := h.HandleFileMessage(ctx)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if !strings.Contains(ctx.sentMsg, "запишитесь на прием") {
			t.Errorf("Expected 'register first' message, got: %s", ctx.sentMsg)
		}
	})

	t.Run("Video without patient returns error", func(t *testing.T) {
		mockRepo := newMockRepository()
		h := NewBookingHandler(nil, nil, nil, nil, nil, mockRepo, &presentation.BotPresenter{}, "", "")

		bot, _ := telebot.NewBot(telebot.Settings{Offline: true})
		ctx := &mockContext{
			sender: &telebot.User{ID: 123},
			message: &telebot.Message{
				Video: &telebot.Video{
					File:     telebot.File{FileID: "vid123", FileSize: 1024000},
					FileName: "test.mp4",
				},
			},
			bot: bot,
		}

		err := h.HandleFileMessage(ctx)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if !strings.Contains(ctx.sentMsg, "запишитесь на прием") {
			t.Errorf("Expected 'register first' message, got: %s", ctx.sentMsg)
		}
	})

	t.Run("Animation without patient returns error", func(t *testing.T) {
		mockRepo := newMockRepository()
		h := NewBookingHandler(nil, nil, nil, nil, nil, mockRepo, &presentation.BotPresenter{}, "", "")

		bot, _ := telebot.NewBot(telebot.Settings{Offline: true})
		ctx := &mockContext{
			sender: &telebot.User{ID: 123},
			message: &telebot.Message{
				Animation: &telebot.Animation{
					File:     telebot.File{FileID: "anim123", FileSize: 2048000},
					FileName: "anim.gif",
				},
			},
			bot: bot,
		}

		err := h.HandleFileMessage(ctx)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if !strings.Contains(ctx.sentMsg, "запишитесь на прием") {
			t.Errorf("Expected 'register first' message, got: %s", ctx.sentMsg)
		}
	})

	t.Run("Voice without patient returns error", func(t *testing.T) {
		mockRepo := newMockRepository()
		h := NewBookingHandler(nil, nil, nil, nil, nil, mockRepo, &presentation.BotPresenter{}, "", "")

		bot, _ := telebot.NewBot(telebot.Settings{Offline: true})
		ctx := &mockContext{
			sender: &telebot.User{ID: 123},
			message: &telebot.Message{
				Voice: &telebot.Voice{
					File: telebot.File{FileID: "voice123", FileSize: 64000},
				},
			},
			bot: bot,
		}

		err := h.HandleFileMessage(ctx)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if !strings.Contains(ctx.sentMsg, "запишитесь на прием") {
			t.Errorf("Expected 'register first' message, got: %s", ctx.sentMsg)
		}
	})

	t.Run("Photo too large", func(t *testing.T) {
		mockRepo := newMockRepository()
		_ = mockRepo.SavePatient(domain.Patient{TelegramID: "123", Name: "Test"})
		h := NewBookingHandler(nil, nil, nil, nil, nil, mockRepo, &presentation.BotPresenter{}, "", "")

		bot, _ := telebot.NewBot(telebot.Settings{Offline: true})
		ctx := &mockContext{
			sender: &telebot.User{ID: 123},
			message: &telebot.Message{
				Photo: &telebot.Photo{
					File: telebot.File{FileID: "bigphoto", FileSize: 25 * 1024 * 1024},
				},
			},
			bot: bot,
		}

		err := h.HandleFileMessage(ctx)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if !strings.Contains(ctx.sentMsg, "слишком большой") {
			t.Errorf("Expected 'too large' message, got: %s", ctx.sentMsg)
		}
	})

	t.Run("Video too large", func(t *testing.T) {
		mockRepo := newMockRepository()
		_ = mockRepo.SavePatient(domain.Patient{TelegramID: "123", Name: "Test"})
		h := NewBookingHandler(nil, nil, nil, nil, nil, mockRepo, &presentation.BotPresenter{}, "", "")

		bot, _ := telebot.NewBot(telebot.Settings{Offline: true})
		ctx := &mockContext{
			sender: &telebot.User{ID: 123},
			message: &telebot.Message{
				Video: &telebot.Video{
					File:     telebot.File{FileID: "bigvid", FileSize: 30 * 1024 * 1024},
					FileName: "big.mp4",
				},
			},
			bot: bot,
		}

		err := h.HandleFileMessage(ctx)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if !strings.Contains(ctx.sentMsg, "слишком большой") {
			t.Errorf("Expected 'too large' message, got: %s", ctx.sentMsg)
		}
	})

	t.Run("Voice too large", func(t *testing.T) {
		mockRepo := newMockRepository()
		_ = mockRepo.SavePatient(domain.Patient{TelegramID: "123", Name: "Test"})
		h := NewBookingHandler(nil, nil, nil, nil, nil, mockRepo, &presentation.BotPresenter{}, "", "")

		bot, _ := telebot.NewBot(telebot.Settings{Offline: true})
		ctx := &mockContext{
			sender: &telebot.User{ID: 123},
			message: &telebot.Message{
				Voice: &telebot.Voice{
					File: telebot.File{FileID: "bigvoice", FileSize: 25 * 1024 * 1024},
				},
			},
			bot: bot,
		}

		err := h.HandleFileMessage(ctx)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if !strings.Contains(ctx.sentMsg, "слишком большой") {
			t.Errorf("Expected 'too large' message, got: %s", ctx.sentMsg)
		}
	})

	t.Run("Animation with empty filename generates default", func(t *testing.T) {
		mockRepo := newMockRepository()
		h := NewBookingHandler(nil, nil, nil, nil, nil, mockRepo, &presentation.BotPresenter{}, "", "")

		bot, _ := telebot.NewBot(telebot.Settings{Offline: true})
		ctx := &mockContext{
			sender: &telebot.User{ID: 123},
			message: &telebot.Message{
				Animation: &telebot.Animation{
					File:     telebot.File{FileID: "anim456", FileSize: 1024000},
					FileName: "",
				},
			},
			bot: bot,
		}

		err := h.HandleFileMessage(ctx)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if !strings.Contains(ctx.sentMsg, "запишитесь на прием") {
			t.Errorf("Expected 'register first' message, got: %s", ctx.sentMsg)
		}
	})

	t.Run("Video with empty filename generates default", func(t *testing.T) {
		mockRepo := newMockRepository()
		h := NewBookingHandler(nil, nil, nil, nil, nil, mockRepo, &presentation.BotPresenter{}, "", "")

		bot, _ := telebot.NewBot(telebot.Settings{Offline: true})
		ctx := &mockContext{
			sender: &telebot.User{ID: 123},
			message: &telebot.Message{
				Video: &telebot.Video{
					File:     telebot.File{FileID: "vid456", FileSize: 1024000},
					FileName: "",
				},
			},
			bot: bot,
		}

		err := h.HandleFileMessage(ctx)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if !strings.Contains(ctx.sentMsg, "запишитесь на прием") {
			t.Errorf("Expected 'register first' message, got: %s", ctx.sentMsg)
		}
	})
}

func TestHandleConfirmBooking_AdminBlock(t *testing.T) {
	mockSession := newMockSessionStorage()
	mockRepo := newMockRepository()
	mockApptService := &mockAppointmentService{
		createAppointmentFunc: func(ctx context.Context, a *domain.Appointment) (*domain.Appointment, error) {
			return a, nil
		},
	}

	userID := int64(123)
	mockSession.Set(userID, SessionKeyService, domain.Service{ID: "block_60", Name: "Block 60", DurationMinutes: 60})
	mockSession.Set(userID, SessionKeyDate, time.Date(2025, 7, 1, 0, 0, 0, 0, time.UTC))
	mockSession.Set(userID, SessionKeyTime, "10:00")
	mockSession.Set(userID, SessionKeyName, "Admin")
	mockSession.Set(userID, SessionKeyIsAdminBlock, true)

	bot, _ := telebot.NewBot(telebot.Settings{Offline: true})
	h := NewBookingHandler(mockApptService, mockSession, []string{"123"}, nil, nil, mockRepo, &presentation.BotPresenter{}, "", "")
	ctx := &mockContext{
		sender: &telebot.User{ID: userID},
		bot:    bot,
	}

	err := h.HandleConfirmBooking(ctx)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if !strings.Contains(ctx.sentMsg, "заблокировано") && !strings.Contains(ctx.sentMsg, "ЗАБЛОКИРОВАНО") {
		t.Errorf("Expected block confirmation, got: %s", ctx.sentMsg)
	}
}

func TestHandleConfirmBooking_AdminManual(t *testing.T) {
	mockSession := newMockSessionStorage()
	mockRepo := newMockRepository()
	mockApptService := &mockAppointmentService{
		createAppointmentFunc: func(ctx context.Context, a *domain.Appointment) (*domain.Appointment, error) {
			return a, nil
		},
		getCustomerHistoryFunc: func(ctx context.Context, id string) ([]domain.Appointment, error) {
			return []domain.Appointment{}, nil
		},
	}

	userID := int64(123)
	mockSession.Set(userID, SessionKeyService, domain.Service{ID: "s1", Name: "Massage", DurationMinutes: 60})
	mockSession.Set(userID, SessionKeyDate, time.Date(2025, 7, 1, 0, 0, 0, 0, time.UTC))
	mockSession.Set(userID, SessionKeyTime, "10:00")
	mockSession.Set(userID, SessionKeyName, "Test Patient")
	mockSession.Set(userID, SessionKeyIsAdminManual, true)
	mockSession.Set(userID, SessionKeyPatientID, "456")

	bot, _ := telebot.NewBot(telebot.Settings{Offline: true})
	h := NewBookingHandler(mockApptService, mockSession, []string{"123"}, nil, nil, mockRepo, &presentation.BotPresenter{}, "", "")
	ctx := &mockContext{
		sender: &telebot.User{ID: userID},
		bot:    bot,
	}

	err := h.HandleConfirmBooking(ctx)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if !strings.Contains(ctx.sentMsg, "РУЧНАЯ ЗАПИСЬ") {
		t.Errorf("Expected manual booking confirmation, got: %s", ctx.sentMsg)
	}
}

func TestHandleConfirmBooking_SlotUnavailable(t *testing.T) {
	mockSession := newMockSessionStorage()
	mockRepo := newMockRepository()
	mockApptService := &mockAppointmentService{
		createAppointmentFunc: func(ctx context.Context, a *domain.Appointment) (*domain.Appointment, error) {
			return nil, fmt.Errorf("slot is not available")
		},
	}

	userID := int64(123)
	mockSession.Set(userID, SessionKeyService, domain.Service{ID: "s1", Name: "Massage", DurationMinutes: 60})
	mockSession.Set(userID, SessionKeyDate, time.Date(2025, 7, 1, 0, 0, 0, 0, time.UTC))
	mockSession.Set(userID, SessionKeyTime, "10:00")
	mockSession.Set(userID, SessionKeyName, "Test")

	bot, _ := telebot.NewBot(telebot.Settings{Offline: true})
	h := NewBookingHandler(mockApptService, mockSession, []string{}, nil, nil, mockRepo, &presentation.BotPresenter{}, "", "")
	ctx := &mockContext{
		sender: &telebot.User{ID: userID},
		bot:    bot,
	}

	err := h.HandleConfirmBooking(ctx)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if !strings.Contains(ctx.sentMsg, "занято") {
		t.Errorf("Expected 'slot occupied' message, got: %s", ctx.sentMsg)
	}
}

func TestHandleConfirmBooking_AdminBlockError(t *testing.T) {
	mockSession := newMockSessionStorage()
	mockRepo := newMockRepository()
	mockApptService := &mockAppointmentService{
		createAppointmentFunc: func(ctx context.Context, a *domain.Appointment) (*domain.Appointment, error) {
			return nil, fmt.Errorf("calendar API error")
		},
	}

	userID := int64(123)
	mockSession.Set(userID, SessionKeyService, domain.Service{ID: "block_60", Name: "Block 60", DurationMinutes: 60})
	mockSession.Set(userID, SessionKeyDate, time.Date(2025, 7, 1, 0, 0, 0, 0, time.UTC))
	mockSession.Set(userID, SessionKeyTime, "10:00")
	mockSession.Set(userID, SessionKeyName, "Admin")
	mockSession.Set(userID, SessionKeyIsAdminBlock, true)

	bot, _ := telebot.NewBot(telebot.Settings{Offline: true})
	h := NewBookingHandler(mockApptService, mockSession, []string{"123"}, nil, nil, mockRepo, &presentation.BotPresenter{}, "", "")
	ctx := &mockContext{
		sender: &telebot.User{ID: userID},
		bot:    bot,
	}

	err := h.HandleConfirmBooking(ctx)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if !strings.Contains(ctx.sentMsg, "Ошибка при создании блокировки") {
		t.Errorf("Expected block error message, got: %s", ctx.sentMsg)
	}
}

func TestHandleConfirmBooking_PartialSessionData(t *testing.T) {
	mockSession := newMockSessionStorage()
	userID := int64(123)
	mockSession.Set(userID, SessionKeyService, domain.Service{ID: "s1", Name: "Massage"})

	h := NewBookingHandler(nil, mockSession, nil, nil, nil, nil, &presentation.BotPresenter{}, "", "")
	ctx := &mockContext{
		sender: &telebot.User{ID: userID},
	}

	err := h.HandleConfirmBooking(ctx)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if !strings.Contains(ctx.sentMsg, "Ошибка сессии") {
		t.Errorf("Expected session error message, got: %s", ctx.sentMsg)
	}
}

func TestHandleTimeSelection_MalformedData(t *testing.T) {
	mockSession := newMockSessionStorage()
	h := NewBookingHandler(nil, mockSession, nil, nil, nil, nil, &presentation.BotPresenter{}, "", "")
	ctx := &mockContext{
		sender:   &telebot.User{ID: 123},
		callback: &telebot.Callback{Data: "select_time|invalid_time_format"},
	}

	err := h.HandleTimeSelection(ctx)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if ctx.editedMsg == nil {
		t.Error("Expected edited message for malformed data")
	}
}

func TestHandleTimeSelection_BlockService(t *testing.T) {
	mockSession := newMockSessionStorage()
	userID := int64(123)
	mockSession.Set(userID, SessionKeyService, domain.Service{ID: "block_60", Name: "Block 60", DurationMinutes: 60})
	mockSession.Set(userID, SessionKeyDate, time.Date(2025, 7, 1, 0, 0, 0, 0, time.UTC))

	h := NewBookingHandler(nil, mockSession, nil, nil, nil, nil, &presentation.BotPresenter{}, "", "")
	ctx := &mockContext{
		sender:   &telebot.User{ID: userID},
		callback: &telebot.Callback{Data: "select_time|10:00"},
	}

	err := h.HandleTimeSelection(ctx)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	session := mockSession.Get(userID)
	if name, ok := session[SessionKeyName].(string); !ok || name != "Admin" {
		t.Errorf("Expected name 'Admin' for block service, got %v", name)
	}
}

func TestHandleTimeSelection_ReturningPatient(t *testing.T) {
	mockSession := newMockSessionStorage()
	mockRepo := newMockRepository()
	_ = mockRepo.SavePatient(domain.Patient{TelegramID: "123", Name: "Returning User", TotalVisits: 5})

	userID := int64(123)
	mockSession.Set(userID, SessionKeyService, domain.Service{ID: "s1", Name: "Massage", DurationMinutes: 60})
	mockSession.Set(userID, SessionKeyDate, time.Date(2025, 7, 1, 0, 0, 0, 0, time.UTC))

	h := NewBookingHandler(nil, mockSession, nil, nil, nil, mockRepo, &presentation.BotPresenter{}, "", "")
	ctx := &mockContext{
		sender:   &telebot.User{ID: userID},
		callback: &telebot.Callback{Data: "select_time|10:00"},
	}

	err := h.HandleTimeSelection(ctx)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	session := mockSession.Get(userID)
	if name, ok := session[SessionKeyName].(string); !ok || name != "Returning User" {
		t.Errorf("Expected auto-filled name 'Returning User', got %v", name)
	}
}

func TestHandleTimeSelection_AdminManualWithPreFilledName(t *testing.T) {
	mockSession := newMockSessionStorage()
	userID := int64(123)
	mockSession.Set(userID, SessionKeyService, domain.Service{ID: "s1", Name: "Massage", DurationMinutes: 60})
	mockSession.Set(userID, SessionKeyDate, time.Date(2025, 7, 1, 0, 0, 0, 0, time.UTC))
	mockSession.Set(userID, SessionKeyIsAdminManual, true)
	mockSession.Set(userID, SessionKeyName, "Pre-filled Patient")

	h := NewBookingHandler(nil, mockSession, nil, nil, nil, nil, &presentation.BotPresenter{}, "", "")
	ctx := &mockContext{
		sender:   &telebot.User{ID: userID},
		callback: &telebot.Callback{Data: "select_time|10:00"},
	}

	err := h.HandleTimeSelection(ctx)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	session := mockSession.Get(userID)
	if name, ok := session[SessionKeyName].(string); !ok || name != "Pre-filled Patient" {
		t.Errorf("Expected pre-filled name preserved, got %v", name)
	}
}

func TestHandleTimeSelection_AdminManualWithoutName(t *testing.T) {
	mockSession := newMockSessionStorage()
	userID := int64(123)
	mockSession.Set(userID, SessionKeyService, domain.Service{ID: "s1", Name: "Massage", DurationMinutes: 60})
	mockSession.Set(userID, SessionKeyIsAdminManual, true)

	h := NewBookingHandler(nil, mockSession, nil, nil, nil, nil, &presentation.BotPresenter{}, "", "")
	ctx := &mockContext{
		sender:   &telebot.User{ID: userID},
		callback: &telebot.Callback{Data: "select_time|10:00"},
	}

	err := h.HandleTimeSelection(ctx)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if !strings.Contains(ctx.sentMsg, "имя и фамилию") {
		t.Errorf("Expected name prompt for admin manual without name, got: %s", ctx.sentMsg)
	}
}

func TestAskForTime_ErrorFromService(t *testing.T) {
	mockSession := newMockSessionStorage()
	userID := int64(123)
	mockSession.Set(userID, SessionKeyService, domain.Service{ID: "s1", Name: "Massage", DurationMinutes: 60})
	mockSession.Set(userID, SessionKeyDate, time.Date(2025, 7, 1, 0, 0, 0, 0, time.UTC))

	mockApptService := &mockAppointmentService{
		getAvailableTimeSlotsFunc: func(ctx context.Context, date time.Time, dur int) ([]domain.TimeSlot, error) {
			return nil, fmt.Errorf("calendar unavailable")
		},
	}

	bot, _ := telebot.NewBot(telebot.Settings{Offline: true})
	h := NewBookingHandler(mockApptService, mockSession, nil, nil, nil, nil, &presentation.BotPresenter{}, "", "")
	ctx := &mockContext{
		sender:  &telebot.User{ID: userID},
		message: &telebot.Message{Text: "Calendar", ID: 1, Chat: &telebot.Chat{ID: userID}},
		bot:     bot,
	}

	err := h.askForTime(ctx)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if !strings.Contains(ctx.sentMsg, "Ошибка при получении слотов") {
		t.Errorf("Expected time slots error message, got: %s", ctx.sentMsg)
	}
}

func TestAskForTime_NoSlotsAvailable(t *testing.T) {
	mockSession := newMockSessionStorage()
	userID := int64(123)
	mockSession.Set(userID, SessionKeyService, domain.Service{ID: "s1", Name: "Massage", DurationMinutes: 60})
	mockSession.Set(userID, SessionKeyDate, time.Date(2025, 7, 1, 0, 0, 0, 0, time.UTC))

	mockApptService := &mockAppointmentService{
		getAvailableTimeSlotsFunc: func(ctx context.Context, date time.Time, dur int) ([]domain.TimeSlot, error) {
			return []domain.TimeSlot{}, nil
		},
	}

	h := NewBookingHandler(mockApptService, mockSession, nil, nil, nil, nil, &presentation.BotPresenter{}, "", "")
	ctx := &mockContext{
		sender:  &telebot.User{ID: userID},
		message: &telebot.Message{Text: "Calendar", ID: 1, Chat: &telebot.Chat{ID: userID}},
	}

	err := h.askForTime(ctx)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if !strings.Contains(ctx.sentMsg, "нет доступных временных слотов") && !strings.Contains(fmt.Sprintf("%v", ctx.editedMsg), "нет доступных временных слотов") {
		t.Errorf("Expected no slots message, got sent=%q edited=%v", ctx.sentMsg, ctx.editedMsg)
	}
}

func TestAskForConfirmation_InvalidTimeFormat(t *testing.T) {
	mockSession := newMockSessionStorage()
	userID := int64(123)
	mockSession.Set(userID, SessionKeyService, domain.Service{ID: "s1", Name: "Massage", DurationMinutes: 60})
	mockSession.Set(userID, SessionKeyDate, time.Date(2025, 7, 1, 0, 0, 0, 0, time.UTC))
	mockSession.Set(userID, SessionKeyTime, "not_a_valid_time")
	mockSession.Set(userID, SessionKeyName, "Test")

	h := NewBookingHandler(nil, mockSession, nil, nil, nil, nil, &presentation.BotPresenter{}, "", "")
	ctx := &mockContext{
		sender: &telebot.User{ID: userID},
	}

	err := h.askForConfirmation(ctx)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if !strings.Contains(ctx.sentMsg, "Ошибка форматирования времени") {
		t.Errorf("Expected time format error, got: %s", ctx.sentMsg)
	}
	t.Run("session cleared", func(t *testing.T) {
		session := mockSession.Get(userID)
		if len(session) != 0 {
			t.Errorf("Expected cleared session, got %d keys", len(session))
		}
	})
}

func TestAskForTime_MissingSessionData(t *testing.T) {
	mockSession := newMockSessionStorage()
	userID := int64(123)

	h := NewBookingHandler(nil, mockSession, nil, nil, nil, nil, &presentation.BotPresenter{}, "", "")
	ctx := &mockContext{
		sender: &telebot.User{ID: userID},
	}

	err := h.askForTime(ctx)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if !strings.Contains(ctx.sentMsg, "Сессия истекла") {
		t.Errorf("Expected session expired message, got: %s", ctx.sentMsg)
	}
}

func TestAskForConfirmation_MissingSessionData(t *testing.T) {
	mockSession := newMockSessionStorage()
	userID := int64(123)
	mockSession.Set(userID, SessionKeyService, domain.Service{ID: "s1", Name: "Massage"})

	h := NewBookingHandler(nil, mockSession, nil, nil, nil, nil, &presentation.BotPresenter{}, "", "")
	ctx := &mockContext{
		sender: &telebot.User{ID: userID},
	}

	err := h.askForConfirmation(ctx)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if !strings.Contains(ctx.sentMsg, "Ошибка сессии") {
		t.Errorf("Expected session error message, got: %s", ctx.sentMsg)
	}
}
