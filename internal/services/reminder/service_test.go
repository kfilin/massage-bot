package reminder

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/kfilin/massage-bot/internal/domain"
	"github.com/kfilin/massage-bot/internal/presentation"
	"gopkg.in/telebot.v3"
)

// --- Mocks ---

// mockBotSender captures Send calls without a real Telegram connection.
type mockBotSender struct {
	sentTo   []telebot.Recipient
	sentWhat []interface{}
	err      error // if non-nil, Send returns this error
}

func (m *mockBotSender) Send(to telebot.Recipient, what interface{}, opts ...interface{}) (*telebot.Message, error) {
	m.sentTo = append(m.sentTo, to)
	m.sentWhat = append(m.sentWhat, what)
	if m.err != nil {
		return nil, m.err
	}
	return &telebot.Message{}, nil
}

// mockApptService is a minimal AppointmentService stub for reminder tests.
type mockApptService struct {
	upcomingAppts []domain.Appointment
	err           error
}

func (m *mockApptService) GetUpcomingAppointments(ctx context.Context, timeMin, timeMax time.Time) ([]domain.Appointment, error) {
	return m.upcomingAppts, m.err
}
func (m *mockApptService) GetAvailableServices(ctx context.Context) ([]domain.Service, error) {
	return nil, nil
}
func (m *mockApptService) GetAvailableTimeSlots(ctx context.Context, date time.Time, dur int) ([]domain.TimeSlot, error) {
	return nil, nil
}
func (m *mockApptService) CreateAppointment(ctx context.Context, a *domain.Appointment) (*domain.Appointment, error) {
	return nil, nil
}
func (m *mockApptService) CancelAppointment(ctx context.Context, id string) error { return nil }
func (m *mockApptService) GetCustomerAppointments(ctx context.Context, id string) ([]domain.Appointment, error) {
	return nil, nil
}
func (m *mockApptService) GetCustomerHistory(ctx context.Context, id string) ([]domain.Appointment, error) {
	return nil, nil
}
func (m *mockApptService) GetAllUpcomingAppointments(ctx context.Context) ([]domain.Appointment, error) {
	return nil, nil
}
func (m *mockApptService) FindByID(ctx context.Context, id string) (*domain.Appointment, error) {
	return nil, nil
}
func (m *mockApptService) GetTotalUpcomingCount(ctx context.Context) (int, error) { return 0, nil }
func (m *mockApptService) GetCalendarAccountInfo(ctx context.Context) (string, error) {
	return "", nil
}
func (m *mockApptService) GetCalendarID() string                              { return "" }
func (m *mockApptService) ListCalendars(ctx context.Context) ([]string, error) { return nil, nil }

// mockReminderRepo covers the Repository methods used by reminder.Service.
type mockReminderRepo struct {
	metadataConfirmedAt *time.Time
	metadataReminderMap map[string]bool
	metadataErr         error
	savedMetadata       map[string]bool
}

func newMockReminderRepo() *mockReminderRepo {
	return &mockReminderRepo{
		metadataReminderMap: make(map[string]bool),
		savedMetadata:       make(map[string]bool),
	}
}

func (m *mockReminderRepo) GetAppointmentMetadata(apptID string) (*time.Time, map[string]bool, error) {
	return m.metadataConfirmedAt, m.metadataReminderMap, m.metadataErr
}

func (m *mockReminderRepo) SaveAppointmentMetadata(apptID string, confirmedAt *time.Time, remindersSent map[string]bool) error {
	for k, v := range remindersSent {
		m.savedMetadata[k] = v
	}
	return nil
}

// The remaining Repository methods are not needed for reminder tests — stub them out.
func (m *mockReminderRepo) SavePatient(p domain.Patient) error                      { return nil }
func (m *mockReminderRepo) UpdatePatientProfile(id, name, notes string) error        { return nil }
func (m *mockReminderRepo) GetPatient(id string) (domain.Patient, error)             { return domain.Patient{}, nil }
func (m *mockReminderRepo) SearchPatients(q string) ([]domain.Patient, error)        { return nil, nil }
func (m *mockReminderRepo) IsUserBanned(tid, un string) (bool, error)                { return false, nil }
func (m *mockReminderRepo) BanUser(id string) error                                  { return nil }
func (m *mockReminderRepo) UnbanUser(id string) error                                { return nil }
func (m *mockReminderRepo) LogEvent(id, et string, d map[string]interface{}) error   { return nil }
func (m *mockReminderRepo) GenerateHTMLRecord(p domain.Patient, h []domain.Appointment, admin bool) string {
	return ""
}
func (m *mockReminderRepo) GenerateAdminSearchPage() string                          { return "" }
func (m *mockReminderRepo) CreateBackup() (string, error)                            { return "", nil }
func (m *mockReminderRepo) SaveMedia(media domain.PatientMedia) error                { return nil }
func (m *mockReminderRepo) GetPatientMedia(pid string) ([]domain.PatientMedia, error) {
	return nil, nil
}
func (m *mockReminderRepo) GetMediaByID(mid string) (*domain.PatientMedia, error) { return nil, nil }
func (m *mockReminderRepo) UpdateMediaStatus(mid, status, transcript string) error { return nil }
func (m *mockReminderRepo) GetAppointmentHistory(tid string) ([]domain.Appointment, error) {
	return nil, nil
}
func (m *mockReminderRepo) UpsertAppointments(appts []domain.Appointment) error { return nil }
func (m *mockReminderRepo) DeleteAppointment(id string) error                   { return nil }

// --- Tests ---

func TestScanAndSendReminders_NoAppointments(t *testing.T) {
	bot := &mockBotSender{}
	svc := NewService(&mockApptService{}, newMockReminderRepo(), bot, nil, presentation.NewBotPresenter())

	// Should not panic with an empty list
	svc.ScanAndSendReminders(context.Background())

	if len(bot.sentTo) != 0 {
		t.Errorf("expected no sends, got %d", len(bot.sentTo))
	}
}

func TestScanAndSendReminders_SkipsNoTgID(t *testing.T) {
	now := time.Now().In(domain.ApptTimeZone)
	appt := domain.Appointment{
		ID:           "appt-1",
		CustomerTgID: "", // no Telegram ID
		StartTime:    now.Add(72*time.Hour - 30*time.Minute),
		Status:       "confirmed",
	}

	bot := &mockBotSender{}
	svc := NewService(&mockApptService{upcomingAppts: []domain.Appointment{appt}}, newMockReminderRepo(), bot, nil, presentation.NewBotPresenter())
	svc.ScanAndSendReminders(context.Background())

	if len(bot.sentTo) != 0 {
		t.Errorf("expected no sends for appt without TgID, got %d", len(bot.sentTo))
	}
}

func TestScanAndSendReminders_SkipsCancelled(t *testing.T) {
	now := time.Now().In(domain.ApptTimeZone)
	appt := domain.Appointment{
		ID:           "appt-2",
		CustomerTgID: "123456",
		StartTime:    now.Add(72*time.Hour - 30*time.Minute),
		Status:       "cancelled",
	}

	bot := &mockBotSender{}
	svc := NewService(&mockApptService{upcomingAppts: []domain.Appointment{appt}}, newMockReminderRepo(), bot, nil, presentation.NewBotPresenter())
	svc.ScanAndSendReminders(context.Background())

	if len(bot.sentTo) != 0 {
		t.Errorf("expected no sends for cancelled appt, got %d", len(bot.sentTo))
	}
}

func TestScanAndSendReminders_Sends72hReminder(t *testing.T) {
	now := time.Now().In(domain.ApptTimeZone)
	// Appointment is exactly 71.5h away — inside the [71h, 72h] window
	appt := domain.Appointment{
		ID:           "appt-3",
		CustomerTgID: "123456",
		StartTime:    now.Add(71*time.Hour + 30*time.Minute),
		Status:       "confirmed",
		Service:      domain.Service{Name: "Массаж"},
	}

	bot := &mockBotSender{}
	repo := newMockReminderRepo()
	svc := NewService(&mockApptService{upcomingAppts: []domain.Appointment{appt}}, repo, bot, nil, presentation.NewBotPresenter())
	svc.ScanAndSendReminders(context.Background())

	if len(bot.sentTo) != 1 {
		t.Errorf("expected 1 send for 72h reminder, got %d", len(bot.sentTo))
	}
	if !repo.savedMetadata["72h"] {
		t.Error("expected '72h' to be marked as sent in metadata")
	}
}

func TestScanAndSendReminders_Sends24hReminder(t *testing.T) {
	now := time.Now().In(domain.ApptTimeZone)
	// Appointment is exactly 23.5h away — inside the [23h, 24h] window
	appt := domain.Appointment{
		ID:           "appt-4",
		CustomerTgID: "123456",
		StartTime:    now.Add(23*time.Hour + 30*time.Minute),
		Status:       "confirmed",
		Service:      domain.Service{Name: "Лимфодренаж"},
	}

	bot := &mockBotSender{}
	repo := newMockReminderRepo()
	svc := NewService(&mockApptService{upcomingAppts: []domain.Appointment{appt}}, repo, bot, nil, presentation.NewBotPresenter())
	svc.ScanAndSendReminders(context.Background())

	if len(bot.sentTo) != 1 {
		t.Errorf("expected 1 send for 24h reminder, got %d", len(bot.sentTo))
	}
	if !repo.savedMetadata["24h"] {
		t.Error("expected '24h' to be marked as sent in metadata")
	}
}

func TestScanAndSendReminders_RepoError(t *testing.T) {
	bot := &mockBotSender{}
	svc := NewService(&mockApptService{err: errors.New("db error")}, newMockReminderRepo(), bot, nil, presentation.NewBotPresenter())

	// Should not panic — just log and return
	svc.ScanAndSendReminders(context.Background())

	if len(bot.sentTo) != 0 {
		t.Error("expected no sends when repo returns error")
	}
}

func TestSendReminder_AlreadySent_Skip(t *testing.T) {
	now := time.Now().In(domain.ApptTimeZone)
	appt := domain.Appointment{
		ID:           "appt-5",
		CustomerTgID: "123456",
		StartTime:    now.Add(71*time.Hour + 30*time.Minute),
		Status:       "confirmed",
		Service:      domain.Service{Name: "Иглоукалывание"},
	}

	bot := &mockBotSender{}
	repo := newMockReminderRepo()
	// Mark 72h as already sent
	repo.metadataReminderMap["72h"] = true

	svc := NewService(&mockApptService{upcomingAppts: []domain.Appointment{appt}}, repo, bot, nil, presentation.NewBotPresenter())
	svc.ScanAndSendReminders(context.Background())

	// Should be skipped
	if len(bot.sentTo) != 0 {
		t.Errorf("expected no sends for already-sent reminder, got %d", len(bot.sentTo))
	}
}

func TestSendReminder_Confirmed24h_Skip(t *testing.T) {
	now := time.Now().In(domain.ApptTimeZone)
	appt := domain.Appointment{
		ID:           "appt-6",
		CustomerTgID: "123456",
		StartTime:    now.Add(23*time.Hour + 30*time.Minute),
		Status:       "confirmed",
		Service:      domain.Service{Name: "Консультация"},
	}

	bot := &mockBotSender{}
	repo := newMockReminderRepo()
	// Mark as confirmed
	confirmedTime := now.Add(-1 * time.Hour)
	repo.metadataConfirmedAt = &confirmedTime

	svc := NewService(&mockApptService{upcomingAppts: []domain.Appointment{appt}}, repo, bot, nil, presentation.NewBotPresenter())
	svc.ScanAndSendReminders(context.Background())

	// 24h reminder should be skipped when appointment is already confirmed
	if len(bot.sentTo) != 0 {
		t.Errorf("expected no 24h send for confirmed appt, got %d", len(bot.sentTo))
	}
}

func TestSendReminder_BotSendError_DoesNotPanic(t *testing.T) {
	now := time.Now().In(domain.ApptTimeZone)
	appt := domain.Appointment{
		ID:           "appt-7",
		CustomerTgID: "123456",
		StartTime:    now.Add(71*time.Hour + 30*time.Minute),
		Status:       "confirmed",
		Service:      domain.Service{Name: "Общий массаж"},
	}

	bot := &mockBotSender{err: errors.New("telegram unavailable")}
	svc := NewService(&mockApptService{upcomingAppts: []domain.Appointment{appt}}, newMockReminderRepo(), bot, nil, presentation.NewBotPresenter())

	// Should not panic even if bot.Send fails
	svc.ScanAndSendReminders(context.Background())
}
