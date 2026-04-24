package presentation

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/kfilin/massage-bot/internal/domain"
)

func TestBotPresenter_FormatAppointment(t *testing.T) {
	p := NewBotPresenter()
	appt := domain.Appointment{
		CustomerName: "John Doe",
		Service:      domain.Service{Name: "Relax Massage"},
		StartTime:    time.Date(2026, 2, 3, 10, 0, 0, 0, time.UTC),
		MeetLink:     "https://meet.google.com/abc",
	}

	t.Run("Admin view", func(t *testing.T) {
		got := p.FormatAppointment(appt, true)
		if !strings.Contains(got, "НОВАЯ ЗАПИСЬ") {
			t.Error("Expected admin header")
		}
		if !strings.Contains(got, "John Doe") {
			t.Error("Expected patient name")
		}
		if !strings.Contains(got, "Relax Massage") {
			t.Error("Expected service name")
		}
		if !strings.Contains(got, "10:00") {
			t.Error("Expected time")
		}
		if !strings.Contains(got, "https://meet.google.com/abc") {
			t.Error("Expected meet link")
		}
	})

	t.Run("Patient view", func(t *testing.T) {
		got := p.FormatAppointment(appt, false)
		if !strings.Contains(got, "ЗАПИСЬ ПОДТВЕРЖДЕНА") {
			t.Error("Expected patient header")
		}
		if !strings.Contains(got, "Приходите за 5 минут") {
			t.Error("Expected patient tip")
		}
	})
}

func TestBotPresenter_FormatDraftNotification(t *testing.T) {
	p := NewBotPresenter()
	got := p.FormatDraftNotification("Vera", "Something was said")
	if !strings.Contains(got, "ЧЕРНОВИК РАСШИФРОВКИ") {
		t.Error("Expected draft header")
	}
	if !strings.Contains(got, "Vera") {
		t.Error("Expected patient name")
	}
	if !strings.Contains(got, "Something was said") {
		t.Error("Expected transcript")
	}
}

func TestBotPresenter_FormatWelcome(t *testing.T) {
	p := NewBotPresenter()
	got := p.FormatWelcome("Kirill")
	if !strings.Contains(got, "Здравствуйте, Kirill!") {
		t.Error("Expected greeting")
	}
	if !strings.Contains(got, "Vera Massage Clinic") {
		t.Error("Expected clinic name")
	}
}

func TestBotPresenter_FormatPatientCard(t *testing.T) {
	p := NewBotPresenter()
	patient := domain.Patient{
		TelegramID:     "123",
		Name:           "John Smith",
		TotalVisits:    5,
		CurrentService: "Classic",
		TherapistNotes: "Needs more pressure",
	}

	got := p.FormatPatientCard(patient)
	if !strings.Contains(got, "КАРТА ПАЦИЕНТА #123") {
		t.Error("Expected card header")
	}
	if !strings.Contains(got, "John Smith") {
		t.Error("Expected name")
	}
	if !strings.Contains(got, "Needs more pressure") {
		t.Error("Expected notes")
	}
}

func TestWebPresenter_RenderCard(t *testing.T) {
	p, err := NewWebPresenter()
	if err != nil {
		t.Fatalf("Failed to create WebPresenter: %v", err)
	}

	data := struct {
		Title        string
		BotVersion   string
		Patient      domain.Patient
		RecentVisits []interface{}
		Drafts       []interface{}
		DocGroups    []interface{}
	}{
		Title:      "Med Card",
		BotVersion: "v1.0",
		Patient:    domain.Patient{Name: "Alice"},
	}

	var buf bytes.Buffer
	err = p.RenderCard(&buf, data)
	if err != nil {
		t.Fatalf("Failed to render card: %v", err)
	}

	got := buf.String()
	if !strings.Contains(got, "Alice") {
		t.Error("Expected patient name in HTML")
	}
	if !strings.Contains(got, "МЕДИЦИНСКАЯ КАРТА") {
		t.Error("Expected title in HTML")
	}
}

func TestWebPresenter_RenderSearch(t *testing.T) {
	p, err := NewWebPresenter()
	if err != nil {
		t.Fatalf("Failed to create WebPresenter: %v", err)
	}

	var buf bytes.Buffer
	err = p.RenderSearch(&buf, nil)
	if err != nil {
		t.Fatalf("Failed to render search: %v", err)
	}

	got := buf.String()
	if !strings.Contains(got, "Поиск пациентов") {
		t.Error("Expected search title in HTML")
	}
}
