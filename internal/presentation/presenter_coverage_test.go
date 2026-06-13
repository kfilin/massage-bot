package presentation

import (
	"strings"
	"testing"
	"time"

	"github.com/kfilin/massage-bot/internal/domain"
)

// --- FormatCancellation ---

func TestBotPresenter_FormatCancellation_AdminView(t *testing.T) {
	p := NewBotPresenter()
	appt := &domain.Appointment{
		CustomerName: "Иван Петров",
		StartTime:    time.Date(2026, 3, 15, 14, 30, 0, 0, time.UTC),
	}

	got := p.FormatCancellation(appt, true)

	checks := []string{"ОТМЕНЕНА", "Иван Петров", "15.03.2026", "14:30"}
	for _, want := range checks {
		if !strings.Contains(got, want) {
			t.Errorf("FormatCancellation(admin) missing %q in:\n%s", want, got)
		}
	}
	// Admin view should NOT have the "use /start" tip
	if strings.Contains(got, "/start") {
		t.Error("Admin view should not contain /start tip")
	}
}

func TestBotPresenter_FormatCancellation_PatientView(t *testing.T) {
	p := NewBotPresenter()
	appt := &domain.Appointment{
		CustomerName: "Анна Сидорова",
		StartTime:    time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC),
	}

	got := p.FormatCancellation(appt, false)

	if !strings.Contains(got, "ВАША ЗАПИСЬ ОТМЕНЕНА") {
		t.Error("Expected patient cancellation header")
	}
	if !strings.Contains(got, "/start") {
		t.Error("Patient view should contain /start tip")
	}
}

// --- FormatNotification ---

func TestBotPresenter_FormatNotification_Basic(t *testing.T) {
	p := NewBotPresenter()

	details := map[string]string{
		"Пациент": "Иван",
		"Услуга":  "Классический массаж",
	}

	got := p.FormatNotification("Блокировка", details)

	if !strings.Contains(got, "БЛОКИРОВКА") {
		t.Error("Expected uppercased header")
	}
	if !strings.Contains(got, "Пациент") || !strings.Contains(got, "Иван") {
		t.Error("Expected detail key-value in output")
	}
}

func TestBotPresenter_FormatNotification_SortedKeys(t *testing.T) {
	p := NewBotPresenter()

	details := map[string]string{
		"Zebra":  "last",
		"Alpha":  "first",
		"Middle": "mid",
	}

	got := p.FormatNotification("Test", details)

	// Keys should be sorted alphabetically
	alphaIdx := strings.Index(got, "Alpha")
	middleIdx := strings.Index(got, "Middle")
	zebraIdx := strings.Index(got, "Zebra")

	if alphaIdx >= middleIdx || middleIdx >= zebraIdx {
		t.Errorf("Keys not sorted: Alpha@%d, Middle@%d, Zebra@%d", alphaIdx, middleIdx, zebraIdx)
	}
}

func TestBotPresenter_FormatNotification_EmptyDetails(t *testing.T) {
	p := NewBotPresenter()
	got := p.FormatNotification("Пусто", map[string]string{})

	if !strings.Contains(got, "ПУСТО") {
		t.Error("Expected header even with empty details")
	}
}

// --- FormatBookingSummary ---

func TestBotPresenter_FormatBookingSummary_WithPrice(t *testing.T) {
	p := NewBotPresenter()
	date := time.Date(2026, 7, 20, 11, 0, 0, 0, time.UTC)

	got := p.FormatBookingSummary(
		"Подтверждение",
		"Мария Иванова",
		"Глубокий массаж",
		date,
		90,
		2500,
	)

	checks := []struct {
		label, substr string
	}{
		{"title", "ПОДТВЕРЖДЕНИЕ"},
		{"patient", "Мария Иванова"},
		{"service", "Глубокий массаж"},
		{"date", "20.07.2026"},
		{"time", "11:00"},
		{"duration", "90 мин"},
		{"price", "2500"},
		{"currency", "₺"},
		{"confirm prompt", "Всё верно"},
	}

	for _, c := range checks {
		if !strings.Contains(got, c.substr) {
			t.Errorf("FormatBookingSummary missing [%s]: %q in:\n%s", c.label, c.substr, got)
		}
	}
}

func TestBotPresenter_FormatBookingSummary_ZeroPrice(t *testing.T) {
	p := NewBotPresenter()
	date := time.Date(2026, 7, 20, 11, 0, 0, 0, time.UTC)

	got := p.FormatBookingSummary("Тест", "Пациент", "Услуга", date, 60, 0)

	if strings.Contains(got, "Цена") {
		t.Error("Zero price should not render price line")
	}
}

// --- FormatAppointment with Duration ---

func TestBotPresenter_FormatAppointment_WithDuration(t *testing.T) {
	p := NewBotPresenter()
	appt := &domain.Appointment{
		CustomerName: "Тест Тестов",
		Service:      domain.Service{Name: "Тест"},
		StartTime:    time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC),
		Duration:     45,
	}

	got := p.FormatAppointment(appt, false)

	if !strings.Contains(got, "45 мин") {
		t.Error("Expected duration in output")
	}
}

func TestBotPresenter_FormatAppointment_ZeroDuration(t *testing.T) {
	p := NewBotPresenter()
	appt := &domain.Appointment{
		CustomerName: "Тест",
		Service:      domain.Service{Name: "Тест"},
		StartTime:    time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC),
		Duration:     0,
	}

	got := p.FormatAppointment(appt, false)

	if strings.Contains(got, "Длительность") {
		t.Error("Zero duration should not render duration line")
	}
}

func TestBotPresenter_FormatAppointment_NoMeetLink(t *testing.T) {
	p := NewBotPresenter()
	appt := &domain.Appointment{
		CustomerName: "Тест",
		Service:      domain.Service{Name: "Тест"},
		StartTime:    time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC),
		MeetLink:     "",
	}

	got := p.FormatAppointment(appt, true)

	if strings.Contains(got, "Meet") {
		t.Error("Empty MeetLink should not render Meet line")
	}
}

// --- FormatPatientCard edge cases ---

func TestBotPresenter_FormatPatientCard_EmptyNotes(t *testing.T) {
	p := NewBotPresenter()
	patient := domain.Patient{
		TelegramID:     "999",
		Name:           "Пустой Пациент",
		TotalVisits:    0,
		CurrentService: "",
		TherapistNotes: "",
	}

	got := p.FormatPatientCard(patient)

	if !strings.Contains(got, "Записей пока нет") {
		t.Error("Empty notes should show placeholder")
	}
}

func TestBotPresenter_FormatPatientCard_LongNotes(t *testing.T) {
	p := NewBotPresenter()
	// Notes longer than 500 chars should be truncated
	longNotes := strings.Repeat("А", 600)
	patient := domain.Patient{
		TelegramID:     "888",
		Name:           "Длинный Пациент",
		TotalVisits:    10,
		CurrentService: "Классический",
		TherapistNotes: longNotes,
	}

	got := p.FormatPatientCard(patient)

	if !strings.Contains(got, "...") {
		t.Error("Long notes should be truncated with ...")
	}
	// Verify truncation at 497 + "..." = 500 chars
	if strings.Contains(got, strings.Repeat("А", 500)) {
		t.Error("Notes should be truncated to 500 chars")
	}
}
