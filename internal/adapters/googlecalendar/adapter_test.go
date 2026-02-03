package googlecalendar

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/kfilin/massage-bot/internal/domain"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

func newTestAdapter(t *testing.T, handler http.Handler) *adapter {
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	ctx := context.Background()
	svc, err := calendar.NewService(ctx, option.WithEndpoint(server.URL), option.WithoutAuthentication())
	if err != nil {
		t.Fatalf("Failed to create calendar service: %v", err)
	}

	return &adapter{
		client:     svc,
		calendarID: "primary",
	}
}

// TestNewAdapter tests adapter creation
func TestNewAdapter(t *testing.T) {
	ctx := context.Background()
	svc, err := calendar.NewService(ctx, option.WithoutAuthentication())
	if err != nil {
		t.Fatalf("Failed to create calendar service: %v", err)
	}

	adapter := NewAdapter(svc, "test-calendar-id")

	if adapter == nil {
		t.Fatal("NewAdapter returned nil")
	}

	// Note: adapter type is unexported, so we can't directly test GetCalendarID here
	// The functionality is tested in TestAdapter_GetCalendarID with the test helper
}

// TestAdapter_GetCalendarID tests getting calendar ID
func TestAdapter_GetCalendarID(t *testing.T) {
	a := newTestAdapter(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	id := a.GetCalendarID()
	if id != "primary" {
		t.Errorf("GetCalendarID() = %s, want primary", id)
	}
}

// TestAdapter_Create tests creating appointments
func TestAdapter_Create(t *testing.T) {
	tests := []struct {
		name        string
		appt        *domain.Appointment
		mockHandler http.HandlerFunc
		wantErr     bool
		checkID     bool
	}{
		{
			name: "Success",
			appt: &domain.Appointment{
				ServiceID:    "massage-60",
				StartTime:    time.Now(),
				EndTime:      time.Now().Add(60 * time.Minute),
				CustomerName: "John Doe",
				CustomerTgID: "123456789",
				Status:       "confirmed",
			},
			mockHandler: func(w http.ResponseWriter, r *http.Request) {
				if r.Method != "POST" {
					http.Error(w, "Expected POST", http.StatusBadRequest)
					return
				}

				resp := &calendar.Event{
					Id:      "created-event-123",
					Summary: "Massage - John Doe",
					Start:   &calendar.EventDateTime{DateTime: time.Now().Format(time.RFC3339)},
					End:     &calendar.EventDateTime{DateTime: time.Now().Add(time.Hour).Format(time.RFC3339)},
					Status:  "confirmed",
				}
				json.NewEncoder(w).Encode(resp)
			},
			wantErr: false,
			checkID: true,
		},
		{
			name: "API Error",
			appt: &domain.Appointment{
				ServiceID:    "massage-60",
				StartTime:    time.Now(),
				EndTime:      time.Now().Add(60 * time.Minute),
				CustomerName: "Jane Doe",
			},
			mockHandler: func(w http.ResponseWriter, r *http.Request) {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			},
			wantErr: true,
			checkID: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := newTestAdapter(t, tt.mockHandler)
			got, err := a.Create(context.Background(), tt.appt)

			if (err != nil) != tt.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.checkID {
				if got.ID == "" {
					t.Error("Create() returned appointment with empty ID")
				}
			}
		})
	}
}

func TestAdapter_FindByID(t *testing.T) {
	tests := []struct {
		name        string
		eventID     string
		mockHandler http.HandlerFunc
		wantErr     bool
		wantName    string // Check customer name or summary substring
	}{
		{
			name:    "Success",
			eventID: "event123",
			mockHandler: func(w http.ResponseWriter, r *http.Request) {
				if r.Method != "GET" {
					http.Error(w, "Expected GET", http.StatusBadRequest)
					return
				}
				// Verify URL path
				if r.URL.Path != "/calendars/primary/events/event123" {
					http.Error(w, "Wrong path", http.StatusBadRequest)
					return
				}

				resp := &calendar.Event{
					Id:      "event123",
					Summary: "Massage - John Doe",
					Start:   &calendar.EventDateTime{DateTime: time.Now().Format(time.RFC3339)},
					End:     &calendar.EventDateTime{DateTime: time.Now().Add(time.Hour).Format(time.RFC3339)},
					Status:  "confirmed",
				}
				json.NewEncoder(w).Encode(resp)
			},
			wantErr:  false,
			wantName: "John Doe",
		},
		{
			name:    "Not Found",
			eventID: "missing",
			mockHandler: func(w http.ResponseWriter, r *http.Request) {
				http.Error(w, "Not Found", http.StatusNotFound)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := newTestAdapter(t, tt.mockHandler)
			got, err := a.FindByID(context.Background(), tt.eventID)

			if (err != nil) != tt.wantErr {
				t.Errorf("FindByID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != nil {
				if got.CustomerName != tt.wantName {
					t.Errorf("FindByID() CustomerName = %v, want %v", got.CustomerName, tt.wantName)
				}
			}
		})
	}
}

func TestAdapter_Delete(t *testing.T) {
	tests := []struct {
		name        string
		eventID     string
		mockHandler http.HandlerFunc
		wantErr     bool
	}{
		{
			name:    "Success",
			eventID: "event123",
			mockHandler: func(w http.ResponseWriter, r *http.Request) {
				if r.Method != "DELETE" {
					http.Error(w, "Expected DELETE", http.StatusBadRequest)
					return
				}
				w.WriteHeader(http.StatusNoContent)
			},
			wantErr: false,
		},
		{
			name:    "Not Found",
			eventID: "missing",
			mockHandler: func(w http.ResponseWriter, r *http.Request) {
				http.Error(w, "Not Found", http.StatusNotFound)
			},
			wantErr: true, // Should return ErrAppointmentNotFound which is an error
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := newTestAdapter(t, tt.mockHandler)
			err := a.Delete(context.Background(), tt.eventID)

			if (err != nil) != tt.wantErr {
				t.Errorf("Delete() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr && tt.name == "Not Found" && err != domain.ErrAppointmentNotFound {
				t.Errorf("Delete() error = %v, want ErrAppointmentNotFound", err)
			}
		})
	}
}

// TestAdapter_FindAll tests fetching all events
func TestAdapter_FindAll(t *testing.T) {
	tests := []struct {
		name        string
		mockHandler http.HandlerFunc
		wantErr     bool
		wantCount   int
	}{
		{
			name: "Success with events",
			mockHandler: func(w http.ResponseWriter, r *http.Request) {
				if r.Method != "GET" {
					http.Error(w, "Expected GET", http.StatusBadRequest)
					return
				}

				resp := &calendar.Events{
					Items: []*calendar.Event{
						{
							Id:      "event1",
							Summary: "Massage - Alice",
							Start:   &calendar.EventDateTime{DateTime: time.Now().Format(time.RFC3339)},
							End:     &calendar.EventDateTime{DateTime: time.Now().Add(time.Hour).Format(time.RFC3339)},
							Status:  "confirmed",
						},
						{
							Id:      "event2",
							Summary: "Massage - Bob",
							Start:   &calendar.EventDateTime{DateTime: time.Now().Add(2 * time.Hour).Format(time.RFC3339)},
							End:     &calendar.EventDateTime{DateTime: time.Now().Add(3 * time.Hour).Format(time.RFC3339)},
							Status:  "confirmed",
						},
					},
				}
				json.NewEncoder(w).Encode(resp)
			},
			wantErr:   false,
			wantCount: 2,
		},
		{
			name: "Empty calendar",
			mockHandler: func(w http.ResponseWriter, r *http.Request) {
				resp := &calendar.Events{
					Items: []*calendar.Event{},
				}
				json.NewEncoder(w).Encode(resp)
			},
			wantErr:   false,
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := newTestAdapter(t, tt.mockHandler)
			got, err := a.FindAll(context.Background())

			if (err != nil) != tt.wantErr {
				t.Errorf("FindAll() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if len(got) != tt.wantCount {
				t.Errorf("FindAll() returned %d events, want %d", len(got), tt.wantCount)
			}
		})
	}
}

// TestAdapter_FindEvents tests fetching events with time range
func TestAdapter_FindEvents(t *testing.T) {
	now := time.Now()
	tomorrow := now.Add(24 * time.Hour)

	tests := []struct {
		name        string
		timeMin     *time.Time
		timeMax     *time.Time
		mockHandler http.HandlerFunc
		wantErr     bool
		wantCount   int
	}{
		{
			name:    "With time range",
			timeMin: &now,
			timeMax: &tomorrow,
			mockHandler: func(w http.ResponseWriter, r *http.Request) {
				// Verify query parameters
				if !strings.Contains(r.URL.RawQuery, "timeMin") {
					http.Error(w, "Missing timeMin", http.StatusBadRequest)
					return
				}

				resp := &calendar.Events{
					Items: []*calendar.Event{
						{
							Id:      "event1",
							Summary: "Massage - Charlie",
							Start:   &calendar.EventDateTime{DateTime: now.Add(time.Hour).Format(time.RFC3339)},
							End:     &calendar.EventDateTime{DateTime: now.Add(2 * time.Hour).Format(time.RFC3339)},
							Status:  "confirmed",
						},
					},
				}
				json.NewEncoder(w).Encode(resp)
			},
			wantErr:   false,
			wantCount: 1,
		},
		{
			name:    "No time range",
			timeMin: nil,
			timeMax: nil,
			mockHandler: func(w http.ResponseWriter, r *http.Request) {
				resp := &calendar.Events{
					Items: []*calendar.Event{},
				}
				json.NewEncoder(w).Encode(resp)
			},
			wantErr:   false,
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := newTestAdapter(t, tt.mockHandler)
			got, err := a.FindEvents(context.Background(), tt.timeMin, tt.timeMax)

			if (err != nil) != tt.wantErr {
				t.Errorf("FindEvents() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if len(got) != tt.wantCount {
				t.Errorf("FindEvents() returned %d events, want %d", len(got), tt.wantCount)
			}
		})
	}
}

// TestAdapter_GetFreeBusy tests fetching free/busy information
func TestAdapter_GetFreeBusy(t *testing.T) {
	now := time.Now()
	tomorrow := now.Add(24 * time.Hour)

	tests := []struct {
		name        string
		mockHandler http.HandlerFunc
		wantErr     bool
		wantSlots   int
	}{
		{
			name: "Success with busy slots",
			mockHandler: func(w http.ResponseWriter, r *http.Request) {
				if r.Method != "POST" {
					http.Error(w, "Expected POST", http.StatusBadRequest)
					return
				}

				resp := &calendar.FreeBusyResponse{
					Calendars: map[string]calendar.FreeBusyCalendar{
						"primary": calendar.FreeBusyCalendar{
							Busy: []*calendar.TimePeriod{
								{
									Start: now.Add(time.Hour).Format(time.RFC3339),
									End:   now.Add(2 * time.Hour).Format(time.RFC3339),
								},
								{
									Start: now.Add(3 * time.Hour).Format(time.RFC3339),
									End:   now.Add(4 * time.Hour).Format(time.RFC3339),
								},
							},
						},
					},
				}
				json.NewEncoder(w).Encode(resp)
			},
			wantErr:   false,
			wantSlots: 2,
		},
		{
			name: "No busy slots",
			mockHandler: func(w http.ResponseWriter, r *http.Request) {
				resp := &calendar.FreeBusyResponse{
					Calendars: map[string]calendar.FreeBusyCalendar{
						"primary": calendar.FreeBusyCalendar{
							Busy: []*calendar.TimePeriod{},
						},
					},
				}
				json.NewEncoder(w).Encode(resp)
			},
			wantErr:   false,
			wantSlots: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := newTestAdapter(t, tt.mockHandler)
			got, err := a.GetFreeBusy(context.Background(), now, tomorrow)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetFreeBusy() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if len(got) != tt.wantSlots {
				t.Errorf("GetFreeBusy() returned %d slots, want %d", len(got), tt.wantSlots)
			}
		})
	}
}

// TestAdapter_GetAccountInfo tests getting account information
func TestAdapter_GetAccountInfo(t *testing.T) {
	tests := []struct {
		name        string
		mockHandler http.HandlerFunc
		wantErr     bool
		wantEmail   string
	}{
		{
			name: "Success",
			mockHandler: func(w http.ResponseWriter, r *http.Request) {
				if r.Method != "GET" {
					http.Error(w, "Expected GET", http.StatusBadRequest)
					return
				}

				resp := &calendar.Calendar{
					Id:      "primary",
					Summary: "Test Calendar",
				}
				json.NewEncoder(w).Encode(resp)
			},
			wantErr:   false,
			wantEmail: "Test Calendar",
		},
		{
			name: "API Error",
			mockHandler: func(w http.ResponseWriter, r *http.Request) {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := newTestAdapter(t, tt.mockHandler)
			got, err := a.GetAccountInfo(context.Background())

			if (err != nil) != tt.wantErr {
				t.Errorf("GetAccountInfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && got != tt.wantEmail {
				// Note: GetAccountInfo returns calendar Summary, not ID
				// For the test, we're checking it returns the expected value
				t.Errorf("GetAccountInfo() = %s, want %s", got, tt.wantEmail)
			}
		})
	}
}

// TestAdapter_ListCalendars tests listing calendars
func TestAdapter_ListCalendars(t *testing.T) {
	tests := []struct {
		name        string
		mockHandler http.HandlerFunc
		wantErr     bool
		wantCount   int
	}{
		{
			name: "Success with multiple calendars",
			mockHandler: func(w http.ResponseWriter, r *http.Request) {
				resp := &calendar.CalendarList{
					Items: []*calendar.CalendarListEntry{
						{Id: "calendar1@example.com", Summary: "Calendar 1"},
						{Id: "calendar2@example.com", Summary: "Calendar 2"},
					},
				}
				json.NewEncoder(w).Encode(resp)
			},
			wantErr:   false,
			wantCount: 2,
		},
		{
			name: "Empty list",
			mockHandler: func(w http.ResponseWriter, r *http.Request) {
				resp := &calendar.CalendarList{
					Items: []*calendar.CalendarListEntry{},
				}
				json.NewEncoder(w).Encode(resp)
			},
			wantErr:   false,
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := newTestAdapter(t, tt.mockHandler)
			got, err := a.ListCalendars(context.Background())

			if (err != nil) != tt.wantErr {
				t.Errorf("ListCalendars() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if len(got) != tt.wantCount {
				t.Errorf("ListCalendars() returned %d calendars, want %d", len(got), tt.wantCount)
			}
		})
	}
}

// TestIsNotFound tests the isNotFound helper
func TestIsNotFound(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "Nil error",
			err:  nil,
			want: false,
		},
		{
			name: "Non-404 error",
			err:  domain.ErrInvalidAppointment,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isNotFound(tt.err); got != tt.want {
				t.Errorf("isNotFound() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestEventToAppointment tests event conversion
func TestEventToAppointment(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name    string
		event   *calendar.Event
		wantErr bool
		checkFn func(*testing.T, *domain.Appointment)
	}{
		{
			name: "Valid event with summary",
			event: &calendar.Event{
				Id:      "event123",
				Summary: "Massage - John Doe",
				Start:   &calendar.EventDateTime{DateTime: now.Format(time.RFC3339)},
				End:     &calendar.EventDateTime{DateTime: now.Add(time.Hour).Format(time.RFC3339)},
				Status:  "confirmed",
			},
			wantErr: false,
			checkFn: func(t *testing.T, appt *domain.Appointment) {
				if appt.ID != "event123" {
					t.Errorf("ID = %s, want event123", appt.ID)
				}
				if appt.CustomerName != "John Doe" {
					t.Errorf("CustomerName = %s, want John Doe", appt.CustomerName)
				}
				if appt.Status != "confirmed" {
					t.Errorf("Status = %s, want confirmed", appt.Status)
				}
			},
		},
		{
			name: "Event with all-day date",
			event: &calendar.Event{
				Id:      "event456",
				Summary: "All Day Event",
				Start:   &calendar.EventDateTime{Date: "2026-02-03"},
				End:     &calendar.EventDateTime{Date: "2026-02-04"},
				Status:  "confirmed",
			},
			wantErr: false,
			checkFn: func(t *testing.T, appt *domain.Appointment) {
				if appt.ID != "event456" {
					t.Errorf("ID = %s, want event456", appt.ID)
				}
			},
		},
		{
			name: "Cancelled event",
			event: &calendar.Event{
				Id:      "event789",
				Summary: "Cancelled Event",
				Start:   &calendar.EventDateTime{DateTime: now.Format(time.RFC3339)},
				End:     &calendar.EventDateTime{DateTime: now.Add(time.Hour).Format(time.RFC3339)},
				Status:  "cancelled",
			},
			wantErr: false,
			checkFn: func(t *testing.T, appt *domain.Appointment) {
				if appt.Status != "cancelled" {
					t.Errorf("Status = %s, want cancelled", appt.Status)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := eventToAppointment(tt.event)

			if (err != nil) != tt.wantErr {
				t.Errorf("eventToAppointment() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.checkFn != nil {
				tt.checkFn(t, got)
			}
		})
	}
}
