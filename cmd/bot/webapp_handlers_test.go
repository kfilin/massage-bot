package main

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/kfilin/massage-bot/internal/domain"
	"github.com/kfilin/massage-bot/internal/ports"
)

// Minimal mocks
type mockRepo struct {
	ports.Repository // Embed interface
	patient          domain.Patient
}

func (m *mockRepo) GetPatient(id string) (domain.Patient, error) {
	if id == m.patient.TelegramID {
		return m.patient, nil
	}
	return domain.Patient{}, fmt.Errorf("not found")
}

func (m *mockRepo) GenerateHTMLRecord(p domain.Patient, h []domain.Appointment, isAdmin bool) string {
	return fmt.Sprintf("HTML_RECORD_FOR_%s_ADMIN_%v", p.Name, isAdmin)
}

func (m *mockRepo) GetAppointmentHistory(id string) ([]domain.Appointment, error) {
	return nil, nil // Return empty
}

func (m *mockRepo) UpsertAppointments(a []domain.Appointment) error { return nil }

func (m *mockRepo) GenerateAdminSearchPage() string { return "ADMIN_SEARCH_PAGE" }

type mockApptService struct {
	ports.AppointmentService
}

func (m *mockApptService) GetCustomerHistory(ctx context.Context, id string) ([]domain.Appointment, error) {
	return []domain.Appointment{}, nil
}

func TestAdminViewPatient(t *testing.T) {
	adminID := "100"
	patientID := "200"
	botToken := "123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11"

	repo := &mockRepo{
		patient: domain.Patient{TelegramID: patientID, Name: "Target Patient"},
	}
	service := &mockApptService{}

	handler := NewWebAppHandler(repo, service, botToken, []string{adminID}, "secret")

	// Prepare data map
	data := map[string]string{
		"query_id":  "QID",
		"user":      fmt.Sprintf(`{"id":%s,"first_name":"Admin","last_name":"User"}`, adminID),
		"auth_date": fmt.Sprintf("%d", time.Now().Unix()),
	}

	// Calculate hash
	hash := signInitData(data, botToken)

	// Build full initData string
	var parts []string
	for k, v := range data {
		parts = append(parts, fmt.Sprintf("%s=%s", k, v))
	}
	parts = append(parts, "hash="+hash)
	initData := strings.Join(parts, "&")

	// Request: Admin trying to view Patient
	// URL: /?id=200&initData=...
	req, _ := http.NewRequest("GET", "/?id="+patientID+"&initData="+url.QueryEscape(initData), nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected 200 OK, got %d. Body: %s", rr.Code, rr.Body.String())
	}

	body := rr.Body.String()
	expected := "HTML_RECORD_FOR_Target Patient_ADMIN_true"
	if !strings.Contains(body, expected) {
		t.Errorf("Expected body to contain %q, got %q", expected, body)
	}
}

func TestNormalUserViewSelf(t *testing.T) {
	userID := "300"
	botToken := "123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11"

	repo := &mockRepo{
		patient: domain.Patient{TelegramID: userID, Name: "Self User"},
	}
	service := &mockApptService{}

	handler := NewWebAppHandler(repo, service, botToken, []string{"999"}, "secret") // Admin is 999

	data := map[string]string{
		"query_id":  "QID",
		"user":      fmt.Sprintf(`{"id":%s,"first_name":"User","last_name":"Self"}`, userID),
		"auth_date": fmt.Sprintf("%d", time.Now().Unix()),
	}
	hash := signInitData(data, botToken)
	var parts []string
	for k, v := range data {
		parts = append(parts, fmt.Sprintf("%s=%s", k, v))
	}
	parts = append(parts, "hash="+hash)
	initData := strings.Join(parts, "&")

	// User viewing themselves (id param matches or empty)
	req, _ := http.NewRequest("GET", "/?initData="+url.QueryEscape(initData), nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected 200 OK, got %d", rr.Code)
	}

	body := rr.Body.String()
	expected := "HTML_RECORD_FOR_Self User_ADMIN_false"
	if !strings.Contains(body, expected) {
		t.Errorf("Expected body to contain %q, got %q", expected, body)
	}
}
