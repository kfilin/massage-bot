package web

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
	patient            domain.Patient
	media              domain.PatientMedia
	getAllPatientsFunc func() ([]domain.Patient, error)
	getApptHistoryFunc func(id string) ([]domain.Appointment, error)
	getPatientMediaFunc func(id string) ([]domain.PatientMedia, error)
	updatePatientProfileFunc func(telegramID string, name string, notes string) error
}

func (m *mockRepo) GetPatient(id string) (domain.Patient, error) {
	if id == m.patient.TelegramID {
		return m.patient, nil
	}
	return domain.Patient{}, fmt.Errorf("not found")
}

func (m *mockRepo) GetAllPatients() ([]domain.Patient, error) {
	if m.getAllPatientsFunc != nil {
		return m.getAllPatientsFunc()
	}
	return nil, nil
}

func (m *mockRepo) GenerateHTMLRecord(p domain.Patient, h []domain.Appointment, isAdmin bool) string {
	return fmt.Sprintf("HTML_RECORD_FOR_%s_ADMIN_%v", p.Name, isAdmin)
}

func (m *mockRepo) GetAppointmentHistory(id string) ([]domain.Appointment, error) {
	if m.getApptHistoryFunc != nil {
		return m.getApptHistoryFunc(id)
	}
	return nil, nil
}

func (m *mockRepo) GetAppointmentHistoryPaginated(id string, limit, offset int) ([]domain.Appointment, bool, error) {
	if m.getApptHistoryFunc != nil {
		appts, err := m.getApptHistoryFunc(id)
		return appts, false, err
	}
	return nil, false, nil
}

func (m *mockRepo) GetPatientMedia(id string) ([]domain.PatientMedia, error) {
	if m.getPatientMediaFunc != nil {
		return m.getPatientMediaFunc(id)
	}
	return nil, nil
}

func (m *mockRepo) UpsertAppointments(a []domain.Appointment) error { return nil }

func (m *mockRepo) UpdateMediaStatus(id, status, transcript string) error { return nil }

func (m *mockRepo) SearchPatients(query string) ([]domain.Patient, error) {
	if query == "test_patient" {
		return []domain.Patient{{TelegramID: "999", Name: "Test Patient", TotalVisits: 5}}, nil
	}
	return []domain.Patient{}, nil
}

func (m *mockRepo) GenerateAdminSearchPage() string { return "ADMIN_SEARCH_PAGE" }

func (m *mockRepo) UpdatePatientProfile(telegramID string, name string, notes string) error {
	if m.updatePatientProfileFunc != nil {
		return m.updatePatientProfileFunc(telegramID, name, notes)
	}
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
	cancelError  error // if set, CancelAppointment returns this error
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
	if m.cancelError != nil {
		return m.cancelError
	}
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

func TestWebAppHandler_Unauthenticated(t *testing.T) {
	botToken := "123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11"
	repo := &mockRepo{}
	service := &mockApptService{}
	presenter, _ := presentation.NewWebPresenter()
	handler := NewWebAppHandler(repo, service, presenter, botToken, []string{}, "secret")

	// No auth at all -> should show loading page
	req, _ := http.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected 200 for unauthenticated (serves loading page), got %d", rr.Code)
	}
	body := rr.Body.String()
	if !strings.Contains(body, "Авторизация") && !strings.Contains(body, "initData") {
		t.Errorf("Expected auth loading page, got %q", body[:200])
	}
}

func TestWebAppHandler_AdminSearchPage(t *testing.T) {
	botToken := "123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11"
	adminID := "100"
	repo := &mockRepo{patient: domain.Patient{TelegramID: adminID, Name: "Admin"}}
	service := &mockApptService{}
	presenter, _ := presentation.NewWebPresenter()
	handler := NewWebAppHandler(repo, service, presenter, botToken, []string{adminID}, "secret")

	// Admin with no target ID -> search page
	initData := makeInitData(adminID, "Admin", botToken)
	req, _ := http.NewRequest("GET", "/?initData="+url.QueryEscape(initData), nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected 200 for admin search page, got %d", rr.Code)
	}
}

func TestWebAppHandler_HMACAuth(t *testing.T) {
	botToken := "123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11"
	secret := "hmac_secret"
	patientID := "500"
	repo := &mockRepo{patient: domain.Patient{TelegramID: patientID, Name: "HMAC Patient"}}
	service := &mockApptService{}
	presenter, _ := presentation.NewWebPresenter()
	handler := NewWebAppHandler(repo, service, presenter, botToken, []string{}, secret)

	// Generate valid HMAC token
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(patientID))
	token := hex.EncodeToString(h.Sum(nil))

	req, _ := http.NewRequest("GET", "/?id="+patientID+"&token="+token, nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected 200 for HMAC auth, got %d. Body: %s", rr.Code, rr.Body.String())
	}
	if !strings.Contains(rr.Body.String(), "HMAC Patient") {
		t.Errorf("Expected patient name in response")
	}
}

func TestWebAppHandler_PatientNotFound_SelfHeal(t *testing.T) {
	botToken := "123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11"
	repo := &mockRepo{}
	service := &mockApptService{}
	presenter, _ := presentation.NewWebPresenter()
	handler := NewWebAppHandler(repo, service, presenter, botToken, []string{}, "secret")

	// Unknown patient -> self-heal path
	initData := makeInitData("777", "NewUser", botToken)
	req, _ := http.NewRequest("GET", "/?id=777&initData="+url.QueryEscape(initData), nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected 200 for self-heal, got %d. Body: %s", rr.Code, rr.Body.String())
	}
}

func TestDraftHandler_Discard(t *testing.T) {
	botToken := "123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11"
	adminID := "100"

	repo := &mockRepo{
		media: domain.PatientMedia{ID: "media2", PatientID: "200", Transcript: "Draft to discard"},
	}

	handler := NewDraftHandler(repo, botToken, []string{adminID}, "secret")

	initData := makeInitData(adminID, "Admin", botToken)
	body := map[string]string{
		"id":       "media2",
		"initData": initData,
	}
	jsonBody, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", "/api/drafts/media2/discard", bytes.NewBuffer(jsonBody))
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected 200 OK for discard, got %d. Body: %s", rr.Code, rr.Body.String())
	}
}

func TestDraftHandler_NonAdmin(t *testing.T) {
	botToken := "123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11"

	repo := &mockRepo{}
	handler := NewDraftHandler(repo, botToken, []string{"999"}, "secret")

	initData := makeInitData("100", "User", botToken)
	body := map[string]string{"id": "media1", "initData": initData}
	jsonBody, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", "/api/drafts/media1/approve", bytes.NewBuffer(jsonBody))
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("Expected 403 for non-admin, got %d", rr.Code)
	}
}

func TestDraftHandler_MissingParams(t *testing.T) {
	botToken := "123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11"
	adminID := "100"

	repo := &mockRepo{}
	handler := NewDraftHandler(repo, botToken, []string{adminID}, "secret")

	// Missing id
	initData := makeInitData(adminID, "Admin", botToken)
	body := map[string]string{"initData": initData}
	jsonBody, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", "/api/drafts//approve", bytes.NewBuffer(jsonBody))
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	// Handler may return 200 (media not found) or 400 depending on path
	if rr.Code == http.StatusInternalServerError {
		t.Errorf("Expected non-500 response, got %d", rr.Code)
	}
}

func TestSearchHandler_EmptyQuery(t *testing.T) {
	botToken := "123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11"
	adminID := "100"

	repo := &mockRepo{}
	handler := NewSearchHandler(repo, botToken, []string{adminID})

	// Empty query -> GetAllPatients
	initData := makeInitData(adminID, "Admin", botToken)
	req, _ := http.NewRequest("GET", "/api/search?q=", nil)
	req.Header.Set("X-Telegram-Init-Data", initData)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected 200 for empty query, got %d", rr.Code)
	}
}

func TestSearchHandler_NonAdmin(t *testing.T) {
	botToken := "123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11"

	repo := &mockRepo{}
	handler := NewSearchHandler(repo, botToken, []string{"999"})

	initData := makeInitData("100", "User", botToken)
	req, _ := http.NewRequest("GET", "/api/search?q=test", nil)
	req.Header.Set("X-Telegram-Init-Data", initData)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("Expected 403 for non-admin, got %d", rr.Code)
	}
}

func TestSearchHandler_InvalidInitData(t *testing.T) {
	botToken := "123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11"

	repo := &mockRepo{}
	handler := NewSearchHandler(repo, botToken, []string{"100"})

	req, _ := http.NewRequest("GET", "/api/search?q=test", nil)
	req.Header.Set("X-Telegram-Init-Data", "garbage_data")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("Expected 401 for invalid auth, got %d", rr.Code)
	}
}

func TestWebAppHandler_WithAppointments(t *testing.T) {
	botToken := "123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11"
	patientID := "500"

	repo := &mockRepo{
		patient: domain.Patient{TelegramID: patientID, Name: "Patient With Appts"},
	}

	service := &mockApptService{}
	presenter, _ := presentation.NewWebPresenter()
	handler := NewWebAppHandler(repo, service, presenter, botToken, []string{}, "secret")

	initData := makeInitData(patientID, "Patient", botToken)
	req, _ := http.NewRequest("GET", "/?id="+patientID+"&initData="+url.QueryEscape(initData), nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d. Body: %s", rr.Code, rr.Body.String())
	}
	if !strings.Contains(rr.Body.String(), "Patient With Appts") {
		t.Error("Expected patient name in response")
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

	// 3. Method not allowed (GET)
	reqGet, _ := http.NewRequest("GET", "/cancel", nil)
	rrGet := httptest.NewRecorder()
	handler.ServeHTTP(rrGet, reqGet)
	if rrGet.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected 405 for GET, got %d", rrGet.Code)
	}

	// 4. Invalid JSON body
	reqBadJSON, _ := http.NewRequest("POST", "/cancel", bytes.NewBuffer([]byte("not json")))
	rrBadJSON := httptest.NewRecorder()
	handler.ServeHTTP(rrBadJSON, reqBadJSON)
	if rrBadJSON.Code != http.StatusBadRequest {
		t.Errorf("Expected 400 for bad JSON, got %d", rrBadJSON.Code)
	}

	// 5. Missing parameters
	bodyMissing := map[string]string{"initData": "data"}
	jsonMissing, _ := json.Marshal(bodyMissing)
	reqMissing, _ := http.NewRequest("POST", "/cancel", bytes.NewBuffer(jsonMissing))
	rrMissing := httptest.NewRecorder()
	handler.ServeHTTP(rrMissing, reqMissing)
	if rrMissing.Code != http.StatusBadRequest {
		t.Errorf("Expected 400 for missing params, got %d", rrMissing.Code)
	}

	// 6. Unauthorized (bad initData)
	bodyUnauthorized := map[string]string{"initData": "garbage", "apptId": apptID}
	jsonUnauthorized, _ := json.Marshal(bodyUnauthorized)
	reqUnauthorized, _ := http.NewRequest("POST", "/cancel", bytes.NewBuffer(jsonUnauthorized))
	rrUnauthorized := httptest.NewRecorder()
	handler.ServeHTTP(rrUnauthorized, reqUnauthorized)
	if rrUnauthorized.Code != http.StatusUnauthorized {
		t.Errorf("Expected 401 for bad auth, got %d", rrUnauthorized.Code)
	}

	// 7. Appointment not found
	bodyNotFound := map[string]string{"initData": initData, "apptId": "nonexistent"}
	jsonNotFound, _ := json.Marshal(bodyNotFound)
	reqNotFound, _ := http.NewRequest("POST", "/cancel", bytes.NewBuffer(jsonNotFound))
	rrNotFound := httptest.NewRecorder()
	handler.ServeHTTP(rrNotFound, reqNotFound)
	if rrNotFound.Code != http.StatusNotFound {
		t.Errorf("Expected 404 for not found, got %d", rrNotFound.Code)
	}

	// 8. Late cancellation blocked (< 72h, non-admin)
	service.appointments["late_appt"] = domain.Appointment{
		ID:           "late_appt",
		CustomerTgID: userID,
		StartTime:    time.Now().Add(24 * time.Hour), // < 72h
	}
	bodyLate := map[string]string{"initData": initData, "apptId": "late_appt"}
	jsonLate, _ := json.Marshal(bodyLate)
	reqLate, _ := http.NewRequest("POST", "/cancel", bytes.NewBuffer(jsonLate))
	rrLate := httptest.NewRecorder()
	handler.ServeHTTP(rrLate, reqLate)
	if rrLate.Code != http.StatusBadRequest {
		t.Errorf("Expected 400 for late cancel, got %d", rrLate.Code)
	}

	// 9. Admin bypasses < 72h restriction
	adminInitData := makeInitData("999", "Admin", botToken)
	bodyAdmin := map[string]string{"initData": adminInitData, "apptId": "late_appt"}
	jsonAdmin, _ := json.Marshal(bodyAdmin)
	reqAdmin, _ := http.NewRequest("POST", "/cancel", bytes.NewBuffer(jsonAdmin))
	rrAdmin := httptest.NewRecorder()
	handler.ServeHTTP(rrAdmin, reqAdmin)
	if rrAdmin.Code != http.StatusOK {
		t.Errorf("Expected 200 for admin cancel, got %d", rrAdmin.Code)
	}
}

func TestHandleCancel_ServiceError(t *testing.T) {
	botToken := "123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11"
	userID := "300"

	service := &mockApptService{
		appointments: map[string]domain.Appointment{
			"appt_1": {
				ID:           "appt_1",
				CustomerTgID: userID,
				StartTime:    time.Now().Add(100 * time.Hour),
				Service:      domain.Service{Name: "Massage"},
			},
		},
		cancelError: fmt.Errorf("service unavailable"),
	}

	presenter := presentation.NewBotPresenter()
	handler := NewCancelHandler(service, botToken, []string{}, presenter)

	initData := makeInitData(userID, "User", botToken)
	body := map[string]string{"initData": initData, "apptId": "appt_1"}
	jsonBody, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", "/cancel", bytes.NewBuffer(jsonBody))
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("Expected 500 for cancellation service error, got %d. Body: %s", rr.Code, rr.Body.String())
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

type mockTranscriptionService struct{
	transcribeFunc func(ctx context.Context, audio io.Reader, filename string) (string, error)
}

func (m *mockTranscriptionService) Transcribe(ctx context.Context, audio io.Reader, filename string) (string, error) {
	if m.transcribeFunc != nil {
		return m.transcribeFunc(ctx, audio, filename)
	}
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

func TestTranscribeHandler_InvalidInitData(t *testing.T) {
	botToken := "123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11"
	handler := NewTranscribeHandler(&mockTranscriptionService{}, botToken)

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	_ = writer.WriteField("initData", "garbage_data")
	part, _ := writer.CreateFormFile("voice", "voice.ogg")
	part.Write([]byte("fake audio data"))
	writer.Close()

	req, _ := http.NewRequest("POST", "/api/transcribe", &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("Expected 401 for invalid auth, got %d", rr.Code)
	}
}

func TestTranscribeHandler_NonMultipart(t *testing.T) {
	botToken := "123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11"
	handler := NewTranscribeHandler(&mockTranscriptionService{}, botToken)

	req, _ := http.NewRequest("POST", "/api/transcribe", bytes.NewBuffer([]byte("plain text body")))
	req.Header.Set("Content-Type", "text/plain")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected 400 for non-multipart, got %d", rr.Code)
	}
}

func TestWebAppHandler_WithAppointmentsAndMedia(t *testing.T) {
	botToken := "123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11"
	patientID := "600"
	now := time.Now()

	repo := &mockRepo{
		patient: domain.Patient{TelegramID: patientID, Name: "Patient With Data"},
		getApptHistoryFunc: func(id string) ([]domain.Appointment, error) {
			return []domain.Appointment{
				{
					ID: "appt1", CustomerTgID: id, CustomerName: "Patient With Data",
					StartTime: now.AddDate(0, 0, -7), EndTime: now.AddDate(0, 0, -7).Add(1 * time.Hour),
					Status: "confirmed", Service: domain.Service{Name: "Massage"},
				},
				{
					ID: "appt2", CustomerTgID: id, CustomerName: "Patient With Data",
					StartTime: now.AddDate(0, 0, -1), EndTime: now.AddDate(0, 0, -1).Add(1 * time.Hour),
					Status: "cancelled", Service: domain.Service{Name: "Massage"},
				},
				{
					ID: "appt3", CustomerTgID: id, CustomerName: "Patient With Data",
					StartTime: now.AddDate(0, 0, -14), EndTime: now.AddDate(0, 0, -14).Add(1 * time.Hour),
					Status: "confirmed", Service: domain.Service{Name: "Block 60"},
				},
			}, nil
		},
		getPatientMediaFunc: func(id string) ([]domain.PatientMedia, error) {
			return []domain.PatientMedia{
				{ID: "m1", PatientID: id, FileType: "photo", FilePath: "media/"+id+"/photo.jpg", CreatedAt: now},
				{ID: "m2", PatientID: id, FileType: "scan", FilePath: "media/"+id+"/scan.pdf", CreatedAt: now},
				{ID: "m3", PatientID: id, FileType: "voice", FilePath: "media/"+id+"/voice.ogg", CreatedAt: now, Status: "pending", Transcript: "Test draft"},
				{ID: "m4", PatientID: id, FileType: "video", FilePath: "media/"+id+"/video.mp4", CreatedAt: now},
			}, nil
		},
	}

	service := &mockApptService{}
	presenter, _ := presentation.NewWebPresenter()
	handler := NewWebAppHandler(repo, service, presenter, botToken, []string{}, "secret")

	initData := makeInitData(patientID, "Patient", botToken)
	req, _ := http.NewRequest("GET", "/?id="+patientID+"&initData="+url.QueryEscape(initData), nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d. Body: %s", rr.Code, rr.Body.String())
	}
	if !strings.Contains(rr.Body.String(), "Patient With Data") {
		t.Error("Expected patient name in response")
	}
}

func TestWebAppHandler_InitDataFailHMACFallback(t *testing.T) {
	botToken := "123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11"
	secret := "fallback_secret"
	patientID := "700"

	repo := &mockRepo{
		patient: domain.Patient{TelegramID: patientID, Name: "HMAC Fallback Patient"},
	}
	service := &mockApptService{}
	presenter, _ := presentation.NewWebPresenter()
	handler := NewWebAppHandler(repo, service, presenter, botToken, []string{}, secret)

	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(patientID))
	token := hex.EncodeToString(h.Sum(nil))

	req, _ := http.NewRequest("GET", "/?id="+patientID+"&token="+token+"&initData=invalid_init_data", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected 200 for HMAC fallback, got %d. Body: %s", rr.Code, rr.Body.String())
	}
	if !strings.Contains(rr.Body.String(), "HMAC Fallback Patient") {
		t.Error("Expected patient name in response via HMAC fallback")
	}
}

func TestUpdatePatientHandler_InvalidJSON(t *testing.T) {
	botToken := "123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11"
	handler := NewUpdatePatientHandler(&mockRepo{}, botToken, []string{"100"})

	req, _ := http.NewRequest("POST", "/api/patients/update", bytes.NewBuffer([]byte("not json")))
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected 400 for invalid JSON, got %d", rr.Code)
	}
}

func TestUpdatePatientHandler_UnauthorizedAuth(t *testing.T) {
	botToken := "123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11"
	handler := NewUpdatePatientHandler(&mockRepo{}, botToken, []string{"100"})

	body := map[string]string{
		"initData": "garbage_auth_data",
		"id":       "200",
		"name":     "New Name",
	}
	jsonBody, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", "/api/patients/update", bytes.NewBuffer(jsonBody))
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("Expected 401 for unauthorized, got %d", rr.Code)
	}
}

func TestUpdatePatientHandler_EmptyBody(t *testing.T) {
	botToken := "123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11"
	handler := NewUpdatePatientHandler(&mockRepo{}, botToken, []string{"100"})

	req, _ := http.NewRequest("POST", "/api/patients/update", bytes.NewBuffer([]byte("{}")))
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected 400 for empty body, got %d", rr.Code)
	}
}

func TestTranscribeHandler_TranscriptionError(t *testing.T) {
	botToken := "123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11"
	transService := &mockTranscriptionService{
		transcribeFunc: func(ctx context.Context, audio io.Reader, filename string) (string, error) {
			return "", fmt.Errorf("AI service down")
		},
	}
	handler := NewTranscribeHandler(transService, botToken)

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	initData := makeInitData("100", "User", botToken)
	_ = writer.WriteField("initData", initData)
	part, _ := writer.CreateFormFile("voice", "voice.ogg")
	part.Write([]byte("fake audio data"))
	writer.Close()

	req, _ := http.NewRequest("POST", "/api/transcribe", &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("Expected 500 for transcription error, got %d. Body: %s", rr.Code, rr.Body.String())
	}
}

func TestTranscribeHandler_FileFallback(t *testing.T) {
	botToken := "123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11"
	handler := NewTranscribeHandler(&mockTranscriptionService{}, botToken)

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	initData := makeInitData("100", "User", botToken)
	_ = writer.WriteField("initData", initData)
	part, _ := writer.CreateFormFile("file", "voice.ogg")
	part.Write([]byte("fake audio data"))
	writer.Close()

	req, _ := http.NewRequest("POST", "/api/transcribe", &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected 200 for file fallback, got %d. Body: %s", rr.Code, rr.Body.String())
	}
	var resp map[string]string
	json.NewDecoder(rr.Body).Decode(&resp)
	if resp["status"] != "ok" {
		t.Errorf("Expected status ok, got %s", resp["status"])
	}
}

func TestUpdatePatientHandler_UpdateError(t *testing.T) {
	botToken := "123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11"
	repo := &mockRepo{
		updatePatientProfileFunc: func(telegramID string, name string, notes string) error {
			return fmt.Errorf("db error")
		},
	}
	handler := NewUpdatePatientHandler(repo, botToken, []string{"100"})

	initData := makeInitData("100", "Admin", botToken)
	body := map[string]string{
		"initData": initData,
		"id":       "200",
		"name":     "Updated Name",
		"notes":    "Updated notes",
	}
	jsonBody, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", "/api/patients/update", bytes.NewBuffer(jsonBody))
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("Expected 500 for update error, got %d. Body: %s", rr.Code, rr.Body.String())
	}
}

func TestWebAppHandler_HMACAdminSearchPage(t *testing.T) {
	botToken := "123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11"
	secret := "hmac_admin_secret"
	adminID := "100"

	repo := &mockRepo{}
	service := &mockApptService{}
	presenter, _ := presentation.NewWebPresenter()
	handler := NewWebAppHandler(repo, service, presenter, botToken, []string{adminID}, secret)

	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(adminID))
	token := hex.EncodeToString(h.Sum(nil))

	req, _ := http.NewRequest("GET", "/?id="+adminID+"&token="+token, nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected 200 for HMAC admin search page, got %d. Body: %s", rr.Code, rr.Body.String())
	}
	if !strings.Contains(rr.Body.String(), "Поиск пациентов") {
		t.Errorf("Expected admin search page, got: %s", rr.Body.String())
	}
}

// --- History pagination (BACKLOG #21) ---

// countVisitCards returns the number of <div class="card"> occurrences in
// the history section of the rendered HTML. The full page has cards in
// other sections (notes, files), so we look for the history-specific
// service-name "Massage Pagination Test" to disambiguate.
func countHistoryCards(body string) int {
	// Each history card contains the service name we set in the test
	return strings.Count(body, "Massage Pagination Test")
}

// makePaginationHandler builds a handler that returns N appointments for
// the given patient ID. Used by the pagination tests below.
func makePaginationHandler(t *testing.T, n int) (http.HandlerFunc, string, string, *mockRepo) {
	t.Helper()
	adminID := "100"
	patientID := "200"
	botToken := "123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11"

	appts := make([]domain.Appointment, n)
	now := time.Now()
	for i := 0; i < n; i++ {
		appts[i] = domain.Appointment{
			ID:           fmt.Sprintf("appt-%d", i),
			ClientID:     patientID,
			CustomerTgID: patientID,
			StartTime:    now.Add(-time.Duration(i) * time.Hour),
			Status:       "confirmed",
			Service:      domain.Service{Name: "Massage Pagination Test", DurationMinutes: 60, Price: 5000},
		}
	}
	repo := &mockRepo{
		patient: domain.Patient{TelegramID: patientID, Name: "Target Patient"},
		getApptHistoryFunc: func(id string) ([]domain.Appointment, error) {
			return appts, nil
		},
	}
	service := &mockApptService{}
	presenter, _ := presentation.NewWebPresenter()
	handler := NewWebAppHandler(repo, service, presenter, botToken, []string{adminID}, "secret")
	return handler, adminID, patientID, repo
}

func TestPagination_DefaultLimit30(t *testing.T) {
	handler, adminID, patientID, _ := makePaginationHandler(t, 50)

	initData := makeInitData(adminID, "Admin", "123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11")
	req, _ := http.NewRequest("GET", "/?id="+patientID+"&initData="+url.QueryEscape(initData), nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("Expected 200, got %d. Body: %s", rr.Code, rr.Body.String())
	}
	body := rr.Body.String()
	if got := countHistoryCards(body); got != 30 {
		t.Errorf("Expected 30 history cards (default limit), got %d", got)
	}
	if !strings.Contains(body, "Показать ещё") {
		t.Errorf("Expected 'Show more' button when more pages exist, got body: %s", body)
	}
}

func TestPagination_ExplicitLimit(t *testing.T) {
	handler, adminID, patientID, _ := makePaginationHandler(t, 50)

	initData := makeInitData(adminID, "Admin", "123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11")
	req, _ := http.NewRequest("GET", "/?id="+patientID+"&initData="+url.QueryEscape(initData)+"&limit=10", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("Expected 200, got %d", rr.Code)
	}
	if got := countHistoryCards(rr.Body.String()); got != 10 {
		t.Errorf("Expected 10 history cards, got %d", got)
	}
	if !strings.Contains(rr.Body.String(), "Показать ещё") {
		t.Errorf("Expected 'Show more' button when more pages exist")
	}
}

func TestPagination_LastPageNoButton(t *testing.T) {
	handler, adminID, patientID, _ := makePaginationHandler(t, 50)

	initData := makeInitData(adminID, "Admin", "123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11")
	// offset=45, limit=10 → last 5 visits, no more
	req, _ := http.NewRequest("GET", "/?id="+patientID+"&initData="+url.QueryEscape(initData)+"&offset=45&limit=10", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("Expected 200, got %d", rr.Code)
	}
	if got := countHistoryCards(rr.Body.String()); got != 5 {
		t.Errorf("Expected 5 history cards on last page, got %d", got)
	}
	if strings.Contains(rr.Body.String(), "Показать ещё") {
		t.Errorf("Expected NO 'Show more' button on last page, but found one")
	}
}

func TestPagination_PartialRender(t *testing.T) {
	handler, adminID, patientID, _ := makePaginationHandler(t, 50)

	initData := makeInitData(adminID, "Admin", "123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11")
	// ?partial=history returns just the cards + possible show-more button
	req, _ := http.NewRequest("GET", "/?id="+patientID+"&initData="+url.QueryEscape(initData)+"&offset=20&limit=10&partial=history", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("Expected 200, got %d", rr.Code)
	}
	body := rr.Body.String()
	if got := countHistoryCards(body); got != 10 {
		t.Errorf("Expected 10 history cards in partial render, got %d", got)
	}
	// Partial render must NOT include the full page chrome
	if strings.Contains(body, "<!DOCTYPE html>") {
		t.Errorf("Partial render leaked full page chrome")
	}
	if !strings.Contains(body, "МЕДИЦИНСКАЯ КАРТА") && strings.Contains(body, "<!DOCTYPE") {
		// ignore — partial should not have card title
	}
	// Partial render at offset=20, limit=10 with 50 total → still has more
	if !strings.Contains(body, "Показать ещё") {
		t.Errorf("Expected 'Show more' button in partial render when more pages exist")
	}
}

func TestPagination_LimitCap(t *testing.T) {
	handler, adminID, patientID, _ := makePaginationHandler(t, 50)

	initData := makeInitData(adminID, "Admin", "123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11")
	// limit=500 must be capped at 100
	req, _ := http.NewRequest("GET", "/?id="+patientID+"&initData="+url.QueryEscape(initData)+"&limit=500", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("Expected 200, got %d", rr.Code)
	}
	// 50 visits total, cap is 100, so we render all 50, no button
	if got := countHistoryCards(rr.Body.String()); got != 50 {
		t.Errorf("Expected 50 history cards (limit capped at 100, all rendered), got %d", got)
	}
	if strings.Contains(rr.Body.String(), "Показать ещё") {
		t.Errorf("Expected NO 'Show more' button when all visits fit in one page")
	}
}

func TestWebAppHandler_SelfHealNoName(t *testing.T) {
	botToken := "123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11"
	secret := "hmac_secret"
	repo := &mockRepo{}
	service := &mockApptService{}
	presenter, _ := presentation.NewWebPresenter()
	handler := NewWebAppHandler(repo, service, presenter, botToken, []string{}, secret)

	// Use HMAC auth (no name in payload) to trigger the `name == ""` fallback to "Пациент"
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte("777"))
	token := hex.EncodeToString(h.Sum(nil))

	req, _ := http.NewRequest("GET", "/?id=777&token="+token, nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected 200 for self-heal without name, got %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "Пациент") {
		t.Errorf("Expected fallback name 'Пациент' in response")
	}
}

func TestWebAppHandler_DBHistoryError(t *testing.T) {
	botToken := "123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11"
	patientID := "800"

	repo := &mockRepo{
		patient: domain.Patient{TelegramID: patientID, Name: "DB Error Patient"},
		getApptHistoryFunc: func(id string) ([]domain.Appointment, error) {
			return nil, fmt.Errorf("database connection lost")
		},
	}
	service := &mockApptService{}
	presenter, _ := presentation.NewWebPresenter()
	handler := NewWebAppHandler(repo, service, presenter, botToken, []string{}, "secret")

	initData := makeInitData(patientID, "Patient", botToken)
	req, _ := http.NewRequest("GET", "/?id="+patientID+"&initData="+url.QueryEscape(initData), nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected 200 even with DB error, got %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "DB Error Patient") {
		t.Errorf("Expected patient name in response")
	}
}
