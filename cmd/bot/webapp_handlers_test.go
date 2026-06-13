package main

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/kfilin/massage-bot/internal/domain"
	"github.com/kfilin/massage-bot/internal/ports"
	"github.com/kfilin/massage-bot/internal/presentation"
)

// Minimal mocks
type mockRepo struct {
	ports.Repository // Embed interface
	patient          domain.Patient
	media            domain.PatientMedia
}

func (m *mockRepo) GetPatient(id string) (domain.Patient, error) {
	if id == m.patient.TelegramID {
		return m.patient, nil
	}
	return domain.Patient{}, fmt.Errorf("not found")
}

func (m *mockRepo) GetAllPatients() ([]domain.Patient, error) { return nil, nil }

func (m *mockRepo) GenerateHTMLRecord(p domain.Patient, h []domain.Appointment, isAdmin bool) string {
	return fmt.Sprintf("HTML_RECORD_FOR_%s_ADMIN_%v", p.Name, isAdmin)
}

func (m *mockRepo) GetAppointmentHistory(id string) ([]domain.Appointment, error) {
	return nil, nil // Return empty
}

func (m *mockRepo) UpsertAppointments(a []domain.Appointment) error { return nil }

func (m *mockRepo) GetPatientMedia(id string) ([]domain.PatientMedia, error) { return nil, nil }
func (m *mockRepo) UpdateMediaStatus(id, status, transcript string) error { return nil }

func (m *mockRepo) SearchPatients(query string) ([]domain.Patient, error) {
	if query == "test_patient" {
		return []domain.Patient{{TelegramID: "999", Name: "Test Patient", TotalVisits: 5}}, nil
	}
	return []domain.Patient{}, nil
}

func (m *mockRepo) GenerateAdminSearchPage() string { return "ADMIN_SEARCH_PAGE" }

func (m *mockRepo) UpdatePatientProfile(telegramID string, name string, notes string) error {
	if m.patient.TelegramID == telegramID {
		m.patient.Name = name
		m.patient.TherapistNotes = notes
	}
	return nil
}

func (m *mockRepo) GetMediaByID(mediaID string) (*domain.PatientMedia, error) {
	if m.media.ID == mediaID {
		return &m.media, nil
	}
	return nil, fmt.Errorf("media not found")
}

type mockApptService struct {
	ports.AppointmentService
	appointments map[string]domain.Appointment
}

func (m *mockApptService) GetCustomerHistory(ctx context.Context, id string) ([]domain.Appointment, error) {
	return []domain.Appointment{}, nil
}

func (m *mockApptService) FindByID(ctx context.Context, id string) (*domain.Appointment, error) {
	if appt, ok := m.appointments[id]; ok {
		return &appt, nil
	}
	return nil, fmt.Errorf("not found")
}

func (m *mockApptService) CancelAppointment(ctx context.Context, id string) error {
	if _, ok := m.appointments[id]; ok {
		delete(m.appointments, id)
		return nil
	}
	return fmt.Errorf("not found")
}

// signTestInitData mimics telegram's HMAC signature
func signTestInitData(data map[string]string, token string) string {
	var keys []string
	for k := range data {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var arr []string
	for _, k := range keys {
		arr = append(arr, k+"="+data[k])
	}
	checkString := strings.Join(arr, "\n")

	h1 := hmac.New(sha256.New, []byte("WebAppData"))
	h1.Write([]byte(token))
	secret := h1.Sum(nil)

	h2 := hmac.New(sha256.New, secret)
	h2.Write([]byte(checkString))
	return hex.EncodeToString(h2.Sum(nil))
}

func makeInitData(userID string, firstName string, token string) string {
	data := map[string]string{
		"query_id":  "QID",
		"user":      fmt.Sprintf(`{"id":%s,"first_name":"%s","last_name":"User"}`, userID, firstName),
		"auth_date": fmt.Sprintf("%d", time.Now().Unix()),
	}
	hash := signTestInitData(data, token)

	var parts []string
	for k, v := range data {
		parts = append(parts, k+"="+v)
	}
	parts = append(parts, "hash="+hash)
	return strings.Join(parts, "&")
}

func TestAdminViewPatient(t *testing.T) {
	adminID := "100"
	patientID := "200"
	botToken := "123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11"

	repo := &mockRepo{
		patient: domain.Patient{TelegramID: patientID, Name: "Target Patient"},
	}
	service := &mockApptService{}

	presenter, _ := presentation.NewWebPresenter()
	handler := NewWebAppHandler(repo, service, presenter, botToken, []string{adminID}, "secret")

	initData := makeInitData(adminID, "Admin", botToken)

	// Admin viewing patient
	req, _ := http.NewRequest("GET", "/?id="+patientID+"&initData="+url.QueryEscape(initData), nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected 200 OK, got %d. Body: %s", rr.Code, rr.Body.String())
	}

	body := rr.Body.String()
	if !strings.Contains(body, "Target Patient") {
		t.Errorf("Expected body to contain patient name, got %q", body)
	}
	if !strings.Contains(body, "МЕДИЦИНСКАЯ КАРТА") {
		t.Errorf("Expected body to contain card title, got %q", body)
	}
}

func TestHandleSearch(t *testing.T) {
	adminID := "100"
	botToken := "123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11"

	repo := &mockRepo{}

	// Create Search Handler
	handler := NewSearchHandler(repo, botToken, []string{adminID})

	// 1. Authorized Search
	initData := makeInitData(adminID, "Admin", botToken)
	req, _ := http.NewRequest("GET", "/api/search?q=test_patient", nil)
	req.Header.Set("X-Telegram-Init-Data", initData)

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected 200 OK for valid search, got %d", rr.Code)
	}

	// 2. Unauthorized (No initData)
	reqNoAuth, _ := http.NewRequest("GET", "/api/search?q=test_patient", nil)
	rrNoAuth := httptest.NewRecorder()
	handler.ServeHTTP(rrNoAuth, reqNoAuth)

	if rrNoAuth.Code != http.StatusUnauthorized {
		t.Errorf("Expected 401 Unauthorized for empty auth, got %d", rrNoAuth.Code)
	}
}

func TestHandleCancel(t *testing.T) {
	botToken := "123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11"
	userID := "300"
	otherUserID := "400"
	apptID := "appt_1"

	// Setup Service with an appointment
	service := &mockApptService{
		appointments: map[string]domain.Appointment{
			apptID: {
				ID:           apptID,
				CustomerTgID: userID,
				CustomerName: "Test User",
				StartTime:    time.Now().Add(100 * time.Hour), // > 72h
				Service:      domain.Service{Name: "Massage"},
			},
		},
	}

	presenter := presentation.NewBotPresenter()
	handler := NewCancelHandler(service, botToken, []string{"999"}, presenter) // Admin 999

	// 1. Valid Cancel by Owner
	initData := makeInitData(userID, "User", botToken)
	body := map[string]string{
		"initData": initData,
		"apptId":   apptID,
	}
	jsonBody, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", "/cancel", bytes.NewBuffer(jsonBody))
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected 200 OK for valid cancel, got %d. Body: %s", rr.Code, rr.Body.String())
	}

	// 2. Cancel Someone Else's Appt (Forbidden)
	// Reset mock (simple way: recreate or re-add)
	service.appointments[apptID] = domain.Appointment{
		ID:           apptID,
		CustomerTgID: userID,
		StartTime:    time.Now().Add(100 * time.Hour),
	}

	otherInitData := makeInitData(otherUserID, "Hacker", botToken)
	bodyHacker := map[string]string{
		"initData": otherInitData,
		"apptId":   apptID,
	}
	jsonBodyHacker, _ := json.Marshal(bodyHacker)

	reqHacker, _ := http.NewRequest("POST", "/cancel", bytes.NewBuffer(jsonBodyHacker))
	rrHacker := httptest.NewRecorder()
	handler.ServeHTTP(rrHacker, reqHacker)

	if rrHacker.Code != http.StatusForbidden {
		t.Errorf("Expected 403 Forbidden, got %d", rrHacker.Code)
	}
}

func TestDraftHandler_Approve(t *testing.T) {
	botToken := "123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11"
	adminID := "100"

	repo := &mockRepo{
		patient: domain.Patient{TelegramID: "200", Name: "Test Patient", TherapistNotes: "Old notes"},
		media:   domain.PatientMedia{ID: "media1", PatientID: "200", Transcript: "Test transcript"},
	}

	handler := NewDraftHandler(repo, botToken, []string{adminID}, "secret")

	initData := makeInitData(adminID, "Admin", botToken)
	body := map[string]string{
		"id":       "media1",
		"initData": initData,
	}
	jsonBody, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", "/api/drafts/media1/approve", bytes.NewBuffer(jsonBody))
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected 200 OK, got %d. Body: %s", rr.Code, rr.Body.String())
	}
}

func TestDraftHandler_MethodNotAllowed(t *testing.T) {
	botToken := "123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11"
	handler := NewDraftHandler(&mockRepo{}, botToken, []string{"100"}, "secret")

	req, _ := http.NewRequest("GET", "/api/drafts", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected 405, got %d", rr.Code)
	}
}

func TestDraftHandler_Unauthorized(t *testing.T) {
	botToken := "123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11"
	handler := NewDraftHandler(&mockRepo{}, botToken, []string{"100"}, "secret")

	body := map[string]string{
		"id":       "media1",
		"initData": "invalid",
	}
	jsonBody, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", "/api/drafts/media1/approve", bytes.NewBuffer(jsonBody))
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("Expected 401, got %d", rr.Code)
	}
}

func TestUpdatePatientHandler_Success(t *testing.T) {
	botToken := "123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11"
	adminID := "100"

	repo := &mockRepo{
		patient: domain.Patient{TelegramID: "200", Name: "Old Name"},
	}

	handler := NewUpdatePatientHandler(repo, botToken, []string{adminID})

	initData := makeInitData(adminID, "Admin", botToken)
	body := map[string]string{
		"initData": initData,
		"id":       "200",
		"name":     "New Name",
		"notes":    "Updated notes",
	}
	jsonBody, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", "/api/patients/update", bytes.NewBuffer(jsonBody))
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected 200 OK, got %d. Body: %s", rr.Code, rr.Body.String())
	}

	var resp map[string]string
	json.NewDecoder(rr.Body).Decode(&resp)
	if resp["status"] != "ok" {
		t.Errorf("Expected status ok, got %s", resp["status"])
	}
}

func TestUpdatePatientHandler_MethodNotAllowed(t *testing.T) {
	botToken := "123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11"
	handler := NewUpdatePatientHandler(&mockRepo{}, botToken, []string{"100"})

	req, _ := http.NewRequest("GET", "/api/patients/update", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected 405, got %d", rr.Code)
	}
}

func TestUpdatePatientHandler_MissingFields(t *testing.T) {
	botToken := "123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11"
	handler := NewUpdatePatientHandler(&mockRepo{}, botToken, []string{"100"})

	body := map[string]string{
		"initData": "some_data",
		// Missing id
	}
	jsonBody, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", "/api/patients/update", bytes.NewBuffer(jsonBody))
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected 400, got %d", rr.Code)
	}
}

func TestUpdatePatientHandler_NotesTooLong(t *testing.T) {
	botToken := "123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11"
	handler := NewUpdatePatientHandler(&mockRepo{}, botToken, []string{"100"})

	longNotes := strings.Repeat("x", 60000) // > 50KB
	body := map[string]string{
		"initData": "some_data",
		"id":       "200",
		"notes":    longNotes,
	}
	jsonBody, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", "/api/patients/update", bytes.NewBuffer(jsonBody))
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected 400 for notes too long, got %d", rr.Code)
	}
}

func TestUpdatePatientHandler_NonAdmin(t *testing.T) {
	botToken := "123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11"
	repo := &mockRepo{}

	handler := NewUpdatePatientHandler(repo, botToken, []string{"999"}) // Admin is 999

	initData := makeInitData("100", "User", botToken) // User 100 is not admin
	body := map[string]string{
		"initData": initData,
		"id":       "200",
		"name":     "New Name",
	}
	jsonBody, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", "/api/patients/update", bytes.NewBuffer(jsonBody))
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("Expected 403, got %d", rr.Code)
	}
}

type mockTranscriptionService struct{}

func (m *mockTranscriptionService) Transcribe(ctx context.Context, audio io.Reader, filename string) (string, error) {
	return "Test transcription result", nil
}

func TestTranscribeHandler_Success(t *testing.T) {
	botToken := "123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11"

	transService := &mockTranscriptionService{}
	handler := NewTranscribeHandler(transService, botToken)

	// Create multipart form
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// Add initData field
	initData := makeInitData("100", "User", botToken)
	_ = writer.WriteField("initData", initData)

	// Add voice file
	part, _ := writer.CreateFormFile("voice", "voice.ogg")
	part.Write([]byte("fake audio data"))

	writer.Close()

	req, _ := http.NewRequest("POST", "/api/transcribe", &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected 200 OK, got %d. Body: %s", rr.Code, rr.Body.String())
	}

	var resp map[string]string
	json.NewDecoder(rr.Body).Decode(&resp)
	if resp["status"] != "ok" {
		t.Errorf("Expected status ok, got %s", resp["status"])
	}
	if resp["text"] != "Test transcription result" {
		t.Errorf("Expected transcription result, got %s", resp["text"])
	}
}

func TestTranscribeHandler_MethodNotAllowed(t *testing.T) {
	botToken := "123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11"
	handler := NewTranscribeHandler(&mockTranscriptionService{}, botToken)

	req, _ := http.NewRequest("GET", "/api/transcribe", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected 405, got %d", rr.Code)
	}
}

func TestTranscribeHandler_MissingAuth(t *testing.T) {
	botToken := "123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11"
	handler := NewTranscribeHandler(&mockTranscriptionService{}, botToken)

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	part, _ := writer.CreateFormFile("voice", "voice.ogg")
	part.Write([]byte("fake audio data"))
	writer.Close()

	req, _ := http.NewRequest("POST", "/api/transcribe", &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("Expected 401, got %d", rr.Code)
	}
}

func TestTranscribeHandler_MissingFile(t *testing.T) {
	botToken := "123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11"
	handler := NewTranscribeHandler(&mockTranscriptionService{}, botToken)

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	initData := makeInitData("100", "User", botToken)
	_ = writer.WriteField("initData", initData)
	writer.Close()

	req, _ := http.NewRequest("POST", "/api/transcribe", &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected 400, got %d", rr.Code)
	}
}
