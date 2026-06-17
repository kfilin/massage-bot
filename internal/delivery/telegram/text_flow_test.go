package telegram

import (
	"io"
	"strings"
	"sync"
	"testing"
	"time"

	"gopkg.in/telebot.v3"

	"github.com/kfilin/massage-bot/internal/domain"
)

// =====================================================================
// Mock BotAPI
// =====================================================================

// mockBotAPI implements BotAPI for testing. It records every Send call
// so tests can assert on the recipient, payload, and options. All
// methods are safe for concurrent use.
type mockBotAPI struct {
	mu sync.Mutex

	sentMessages []sentRecord

	// Optional behaviour overrides
	sendErr  error
	fileRead io.ReadCloser
	fileErr  error
	copyErr  error
}

type sentRecord struct {
	to   telebot.Recipient
	what interface{}
	opts []interface{}
}

func (m *mockBotAPI) Send(to telebot.Recipient, what interface{}, opts ...interface{}) (*telebot.Message, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.sendErr != nil {
		return nil, m.sendErr
	}
	optsCopy := make([]interface{}, len(opts))
	copy(optsCopy, opts)
	m.sentMessages = append(m.sentMessages, sentRecord{to: to, what: what, opts: optsCopy})
	return &telebot.Message{}, nil
}

func (m *mockBotAPI) Copy(_ telebot.Recipient, _ telebot.Editable, _ ...interface{}) (*telebot.Message, error) {
	if m.copyErr != nil {
		return nil, m.copyErr
	}
	return &telebot.Message{}, nil
}

func (m *mockBotAPI) Delete(_ telebot.Editable) error { return nil }

func (m *mockBotAPI) EditReplyMarkup(_ telebot.Editable, _ *telebot.ReplyMarkup) (*telebot.Message, error) {
	return &telebot.Message{}, nil
}

func (m *mockBotAPI) File(_ *telebot.File) (io.ReadCloser, error) {
	if m.fileErr != nil {
		return nil, m.fileErr
	}
	if m.fileRead != nil {
		return m.fileRead, nil
	}
	return io.NopCloser(strings.NewReader("")), nil
}

func (m *mockBotAPI) Raw(_ string, _ interface{}) ([]byte, error) {
	return nil, nil
}

// =====================================================================
// Mock Repository (minimal — only methods used by text_flow)
// =====================================================================

type tfMockRepo struct {
	mu sync.Mutex

	patients       map[string]domain.Patient
	getPatientErr  error
	savePatientErr error
}

func (m *tfMockRepo) GetPatient(tgID string) (domain.Patient, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.getPatientErr != nil {
		return domain.Patient{}, m.getPatientErr
	}
	if p, ok := m.patients[tgID]; ok {
		return p, nil
	}
	return domain.Patient{}, &notFoundError{tgID: tgID}
}

func (m *tfMockRepo) SavePatient(p domain.Patient) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.savePatientErr != nil {
		return m.savePatientErr
	}
	if m.patients == nil {
		m.patients = make(map[string]domain.Patient)
	}
	m.patients[p.TelegramID] = p
	return nil
}

func (m *tfMockRepo) GetAllPatients() ([]domain.Patient, error) {
	return nil, nil
}
func (m *tfMockRepo) SearchPatients(_ string) ([]domain.Patient, error) {
	return nil, nil
}
func (m *tfMockRepo) IsUserBanned(_ string, _ string) (bool, error) {
	return false, nil
}
func (m *tfMockRepo) BanUser(_ string) error   { return nil }
func (m *tfMockRepo) UnbanUser(_ string) error { return nil }
func (m *tfMockRepo) UpdatePatientProfile(_ string, _ string, _ string) error {
	return nil
}
func (m *tfMockRepo) LogEvent(_ string, _ string, _ map[string]interface{}) error {
	return nil
}
func (m *tfMockRepo) SaveMedia(_ domain.PatientMedia) error { return nil }
func (m *tfMockRepo) GetPatientMedia(_ string) ([]domain.PatientMedia, error) {
	return nil, nil
}
func (m *tfMockRepo) GetMediaByID(_ string) (*domain.PatientMedia, error) {
	return nil, nil
}
func (m *tfMockRepo) UpdateMediaStatus(_ string, _ string, _ string) error {
	return nil
}
func (m *tfMockRepo) CreateBackup() (string, error) { return "", nil }
func (m *tfMockRepo) GetAppointmentHistory(_ string) ([]domain.Appointment, error) {
	return nil, nil
}
func (m *tfMockRepo) UpsertAppointments(_ []domain.Appointment) error { return nil }
func (m *tfMockRepo) DeleteAppointment(_ string) error                { return nil }
func (m *tfMockRepo) SaveAppointmentMetadata(_ string, _ *time.Time, _ map[string]bool) error {
	return nil
}
func (m *tfMockRepo) GetAppointmentMetadata(_ string) (*time.Time, map[string]bool, error) {
	return nil, nil, nil
}

type notFoundError struct{ tgID string }

func (e *notFoundError) Error() string { return "patient not found: " + e.tgID }

// =====================================================================
// Mock SessionStorage
// =====================================================================

type tfMockSessionStorage struct {
	mu       sync.Mutex
	sessions map[int64]map[string]interface{}
}

func newTFMockSessionStorage() *tfMockSessionStorage {
	return &tfMockSessionStorage{sessions: make(map[int64]map[string]interface{})}
}

func (m *tfMockSessionStorage) Set(userID int64, key string, value interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.sessions[userID] == nil {
		m.sessions[userID] = make(map[string]interface{})
	}
	if value == nil {
		delete(m.sessions[userID], key)
	} else {
		m.sessions[userID][key] = value
	}
}

func (m *tfMockSessionStorage) Get(userID int64) map[string]interface{} {
	m.mu.Lock()
	defer m.mu.Unlock()
	if s, ok := m.sessions[userID]; ok {
		return s
	}
	return nil
}

func (m *tfMockSessionStorage) ClearSession(userID int64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.sessions, userID)
}

// sessionKeyAdminReplyingTo is the key used to look up the patient ID
// being replied to. Must match handlers.SessionKeyAdminReplyingTo.
// We duplicate the constant here to avoid a cyclic import.
const sessionKeyAdminReplyingTo = "admin_replying_to"

// =====================================================================
// Mock telebot.Context
// =====================================================================

// mockContext is the local mock for telebot.Context used by text-flow
// tests. It embeds telebot.Context (nil) to satisfy the interface's
// method set; any method we don't override will panic if called.
// The Bot() override is required because text_flow functions do not
// call it — but Go's type system still requires the method to match
// the interface signature, so we return a nil *telebot.Bot. Callers
// that actually invoke c.Bot() will panic; the text-flow tests don't.
type mockContext struct {
	telebot.Context
	sender *telebot.User
	text   string

	sentMessages []string
	editedMsg    string
}

func (m *mockContext) Sender() *telebot.User { return m.sender }
func (m *mockContext) Text() string          { return m.text }
func (m *mockContext) Bot() *telebot.Bot     { return nil }
func (m *mockContext) Recipient() telebot.Recipient {
	if m.sender != nil {
		return m.sender
	}
	return &telebot.User{}
}
func (m *mockContext) Send(what interface{}, _ ...interface{}) error {
	if s, ok := what.(string); ok {
		m.sentMessages = append(m.sentMessages, s)
	}
	return nil
}
func (m *mockContext) Edit(what interface{}, _ ...interface{}) error {
	if s, ok := what.(string); ok {
		m.editedMsg = s
	}
	return nil
}
func (m *mockContext) Respond(_ ...*telebot.CallbackResponse) error { return nil }

// =====================================================================
// Tests
// =====================================================================

// TestHandleAdminReply_EmptyReplyingTo exercises the "no patient to
// reply to" branch: session lacks SessionKeyAdminReplyingTo, function
// should send an error message and return nil.
func TestHandleAdminReply_EmptyReplyingTo(t *testing.T) {
	repo := &tfMockRepo{}
	sess := newMockSessionStorage()
	bot := &mockBotAPI{}
	ctx := &mockContext{sender: &telebot.User{ID: 1001}}

	if err := handleAdminReply(ctx, bot, repo, sess, 1001, "hello"); err != nil {
		t.Fatalf("handleAdminReply: %v", err)
	}
	if len(ctx.sentMessages) != 1 {
		t.Errorf("expected 1 sent message, got %d", len(ctx.sentMessages))
	}
	if len(bot.sentMessages) != 0 {
		t.Errorf("expected bot to not send anything when no replying-to ID, got %d", len(bot.sentMessages))
	}
}

// TestHandleAdminReply_DeliversToPatient verifies the happy path:
// message reaches the patient, therapist notes are appended, session
// is cleared.
func TestHandleAdminReply_DeliversToPatient(t *testing.T) {
	repo := &tfMockRepo{
		patients: map[string]domain.Patient{
			"500": {TelegramID: "500", Name: "Patient Five", TherapistNotes: "Initial notes."},
		},
	}
	sess := newMockSessionStorage()
	sess.Set(1001, sessionKeyAdminReplyingTo, "500")

	bot := &mockBotAPI{}
	ctx := &mockContext{sender: &telebot.User{ID: 1001}}

	if err := handleAdminReply(ctx, bot, repo, sess, 1001, "Привет, как дела?"); err != nil {
		t.Fatalf("handleAdminReply: %v", err)
	}

	// Bot should have sent one message to the patient.
	if len(bot.sentMessages) != 1 {
		t.Fatalf("expected 1 bot message, got %d", len(bot.sentMessages))
	}
	if u, ok := bot.sentMessages[0].to.(*telebot.User); !ok || u.ID != 500 {
		t.Errorf("message recipient: got %T %+v, want *telebot.User ID=500", bot.sentMessages[0].to, bot.sentMessages[0].to)
	}
	msg, _ := bot.sentMessages[0].what.(string)
	if msg == "" || !contains(msg, "Привет, как дела?") {
		t.Errorf("message body missing or wrong: %q", msg)
	}

	// Therapist notes should have been updated.
	got, err := repo.GetPatient("500")
	if err != nil {
		t.Fatalf("GetPatient: %v", err)
	}
	if got.TherapistNotes == "Initial notes." {
		t.Error("TherapistNotes should have been appended to")
	}
	if !contains(got.TherapistNotes, "Привет, как дела?") {
		t.Errorf("TherapistNotes should contain the reply text, got: %q", got.TherapistNotes)
	}

	// Session key should be cleared.
	if v := sess.Get(1001)[sessionKeyAdminReplyingTo]; v != nil {
		t.Errorf("session key should be cleared, got %v", v)
	}
}

// TestHandleAdminReply_SendFailure verifies error handling when the
// bot's Send fails. Function should return nil (caller doesn't bubble
// errors) and report the failure to the admin via c.Send.
func TestHandleAdminReply_SendFailure(t *testing.T) {
	repo := &tfMockRepo{
		patients: map[string]domain.Patient{"700": {TelegramID: "700"}},
	}
	sess := newMockSessionStorage()
	sess.Set(1001, sessionKeyAdminReplyingTo, "700")

	bot := &mockBotAPI{sendErr: io.ErrUnexpectedEOF}
	ctx := &mockContext{sender: &telebot.User{ID: 1001}}

	if err := handleAdminReply(ctx, bot, repo, sess, 1001, "msg"); err != nil {
		t.Errorf("handleAdminReply should swallow send errors, got %v", err)
	}
	// Should have sent an error message back to the admin (via ctx.Send).
	if len(ctx.sentMessages) == 0 {
		t.Error("expected an error message sent to admin via ctx.Send")
	}
}

// TestHandleAdminReply_PatientMissing verifies that when the patient
// record is missing, the message is still delivered to the chat (the
// chat ID was provided by the session), but therapist notes are not
// updated.
func TestHandleAdminReply_PatientMissing(t *testing.T) {
	repo := &tfMockRepo{patients: map[string]domain.Patient{}} // empty
	sess := newMockSessionStorage()
	sess.Set(1001, sessionKeyAdminReplyingTo, "9999")

	bot := &mockBotAPI{}
	ctx := &mockContext{sender: &telebot.User{ID: 1001}}

	if err := handleAdminReply(ctx, bot, repo, sess, 1001, "ping"); err != nil {
		t.Fatalf("handleAdminReply: %v", err)
	}
	if len(bot.sentMessages) != 1 {
		t.Errorf("expected 1 bot message even if patient missing, got %d", len(bot.sentMessages))
	}
}

// TestForwardPatientMessageToAdmins_NoWebAppURL covers the standard
// "patient sent free text" path: notify admins, log to Med-Card, no
// WebApp link button.
func TestForwardPatientMessageToAdmins_NoWebAppURL(t *testing.T) {
	repo := &tfMockRepo{
		patients: map[string]domain.Patient{
			"800": {TelegramID: "800", Name: "Tester", TherapistNotes: ""},
		},
	}
	bot := &mockBotAPI{}
	ctx := &mockContext{
		sender: &telebot.User{ID: 800, FirstName: "Test", LastName: "User"},
	}

	// bh is no longer used; function takes (webAppURL, generateCardURL) directly
	if err := forwardPatientMessageToAdmins(ctx, bot, repo, "", nil, []string{"1001", "1002"}, "Hello Vera!"); err != nil {
		t.Fatalf("forwardPatientMessageToAdmins: %v", err)
	}

	// Acknowledgement to patient.
	if len(ctx.sentMessages) == 0 {
		t.Error("expected acknowledgement to patient via ctx.Send")
	}

	// One notification per admin.
	if len(bot.sentMessages) != 2 {
		t.Fatalf("expected 2 admin notifications, got %d", len(bot.sentMessages))
	}
	for i, rec := range bot.sentMessages {
		u, ok := rec.to.(*telebot.User)
		if !ok {
			t.Errorf("admin %d: recipient is %T, want *telebot.User", i, rec.to)
			continue
		}
		if u.ID != int64(1001+i) {
			t.Errorf("admin %d: ID=%d, want %d", i, u.ID, 1001+i)
		}
		if !contains(rec.what.(string), "Hello Vera!") {
			t.Errorf("admin %d: notification missing message text", i)
		}
	}

	// Med-Card should have the patient message appended.
	got, _ := repo.GetPatient("800")
	if !contains(got.TherapistNotes, "Hello Vera!") {
		t.Errorf("TherapistNotes should contain patient message, got %q", got.TherapistNotes)
	}
}

// TestForwardPatientMessageToAdmins_NoPatientsRecord: patient has no
// record yet. Function should still send notifications and ack the
// patient.
func TestForwardPatientMessageToAdmins_NoPatientsRecord(t *testing.T) {
	repo := &tfMockRepo{} // empty
	bot := &mockBotAPI{}
	ctx := &mockContext{sender: &telebot.User{ID: 801}}
	// bh no longer used
	if err := forwardPatientMessageToAdmins(ctx, bot, repo, "", nil, []string{"1001"}, "hi"); err != nil {
		t.Fatalf("forwardPatientMessageToAdmins: %v", err)
	}
	if len(bot.sentMessages) != 1 {
		t.Errorf("expected 1 admin notification, got %d", len(bot.sentMessages))
	}
}

// TestForwardPatientMessageToAdmins_WithWebAppURL covers the branch
// where a non-empty webAppURL is provided: the notification should
// include a generated card link.
func TestForwardPatientMessageToAdmins_WithWebAppURL(t *testing.T) {
	repo := &tfMockRepo{
		patients: map[string]domain.Patient{
			"800": {TelegramID: "800", Name: "Tester"},
		},
	}
	bot := &mockBotAPI{}
	ctx := &mockContext{sender: &telebot.User{ID: 800, FirstName: "Test"}}

	generateURL := func(tgID string) string {
		return "https://vera.massage/app?tg=" + tgID
	}

	if err := forwardPatientMessageToAdmins(
		ctx, bot, repo,
		"https://vera.massage/app",
		generateURL,
		[]string{"1001"},
		"ping",
	); err != nil {
		t.Fatalf("forwardPatientMessageToAdmins: %v", err)
	}

	if len(bot.sentMessages) != 1 {
		t.Fatalf("expected 1 admin notification, got %d", len(bot.sentMessages))
	}
	body, _ := bot.sentMessages[0].what.(string)
	if !contains(body, "https://vera.massage/app?tg=800") {
		t.Errorf("notification should contain generated card URL, got: %q", body)
	}
	if !contains(body, "Открыть мед-карту") {
		t.Errorf("notification should contain the card link label, got: %q", body)
	}
}

// --- helpers -------------------------------------------------------------

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || (len(sub) > 0 && indexOf(s, sub) >= 0))
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}

// mockBookingHandler is no longer needed; tests pass
// (webAppURL, generateCardURL) directly to forwardPatientMessageToAdmins.
type mockBookingHandler struct {
	WebAppURL string
}
