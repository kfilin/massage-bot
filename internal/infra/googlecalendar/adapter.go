package googlecalendar

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/kfilin/massage-bot/internal/domain"
	"google.golang.org/api/calendar/v3"
)

type Adapter struct {
	service  *calendar.Service
	timezone string
}

func NewAdapter(service *calendar.Service) *Adapter {
	return &Adapter{
		service:  service,
		timezone: "Europe/Moscow",
	}
}

func (a *Adapter) Create(ctx context.Context, appt domain.Appointment) error {
	event := &calendar.Event{
		Summary:     fmt.Sprintf("%s - %s", appt.Service.Name, appt.User.FirstName),
		Description: a.formatDescription(appt.User),
		Start:       a.toEventTime(appt.StartTime),
		End:         a.toEventTime(appt.EndTime),
		Reminders: &calendar.EventReminders{
			UseDefault: false,
			Overrides: []*calendar.EventReminder{
				{Method: "popup", Minutes: 60},
			},
		},
	}

	_, err := a.service.Events.
		Insert("primary", event).
		Context(ctx).
		Do()
	return err
}

func (a *Adapter) GetByTimeRange(ctx context.Context, start, end time.Time) ([]domain.Appointment, error) {
	events, err := a.service.Events.
		List("primary").
		TimeMin(start.Format(time.RFC3339)).
		TimeMax(end.Format(time.RFC3339)).
		Context(ctx).
		Do()

	if err != nil {
		return nil, fmt.Errorf("calendar query failed: %w", err)
	}

	var appointments []domain.Appointment
	for _, event := range events.Items {
		appt, err := a.toDomainAppointment(event)
		if err != nil {
			continue // Skip malformed events
		}
		appointments = append(appointments, appt)
	}
	return appointments, nil
}

func (a *Adapter) UpdateStatus(ctx context.Context, id string, status string) error {
	// Implementation using extendedProperties if needed
	return fmt.Errorf("not implemented")
}

// Helpers
func (a *Adapter) formatDescription(user domain.User) string {
	return fmt.Sprintf(
		"Client: %s %s\nTelegram: @%s\nUser ID: %d",
		user.FirstName,
		user.LastName,
		user.Username,
		user.ID,
	)
}

func (a *Adapter) toEventTime(t time.Time) *calendar.EventDateTime {
	return &calendar.EventDateTime{
		DateTime: t.Format(time.RFC3339),
		TimeZone: a.timezone,
	}
}

func (a *Adapter) toDomainAppointment(event *calendar.Event) (domain.Appointment, error) {
	start, err := time.Parse(time.RFC3339, event.Start.DateTime)
	if err != nil {
		return domain.Appointment{}, err
	}

	end, err := time.Parse(time.RFC3339, event.End.DateTime)
	if err != nil {
		return domain.Appointment{}, err
	}

	return domain.Appointment{
		ID:        event.Id,
		User:      a.extractUser(event),
		StartTime: start,
		EndTime:   end,
		Status:    "confirmed",
	}, nil
}

func (a *Adapter) extractUser(event *calendar.Event) domain.User {
	return domain.User{
		ID:        extractUserID(event.Description),
		FirstName: extractFirstName(event.Summary),
		Username:  extractUsername(event.Description),
	}
}

func extractUserID(desc string) int64 {
	prefix := "User ID: "
	if idx := strings.Index(desc, prefix); idx != -1 {
		var id int64
		fmt.Sscanf(desc[idx+len(prefix):], "%d", &id)
		return id
	}
	return 0
}

func extractFirstName(summary string) string {
	parts := strings.Split(summary, " - ")
	if len(parts) > 1 {
		return parts[1]
	}
	return ""
}

func extractUsername(desc string) string {
	prefix := "Telegram: @"
	if idx := strings.Index(desc, prefix); idx != -1 {
		end := strings.Index(desc[idx:], "\n")
		if end == -1 {
			return desc[idx+len(prefix):]
		}
		return desc[idx+len(prefix) : idx+end]
	}
	return ""
}
