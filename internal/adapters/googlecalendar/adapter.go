package googlecalendar

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/kfilin/massage-bot/internal/domain"
	"github.com/kfilin/massage-bot/internal/ports"
	"google.golang.org/api/calendar/v3"
)

type adapter struct {
	client     *calendar.Service
	calendarID string
}

// NewAdapter creates a new Google Calendar adapter that implements ports.AppointmentRepository.
func NewAdapter(client *calendar.Service) ports.AppointmentRepository {
	// TODO: Load calendarID from config/env var (e.g., os.Getenv("GOOGLE_CALENDAR_ID"))
	// For now, "primary" refers to the default calendar of the authenticated user.
	return &adapter{
		client:     client,
		calendarID: "primary",
	}
}

// Create creates a new appointment event in Google Calendar.
func (a *adapter) Create(ctx context.Context, appt *domain.Appointment) (*domain.Appointment, error) {
	// Ensure StartTime and EndTime are correctly set from the service layer before calling this.
	// The service layer (appointment.Service) should populate these based on appt.Time and appt.Duration.
	if appt.StartTime.IsZero() || appt.EndTime.IsZero() {
		return nil, fmt.Errorf("appointment StartTime or EndTime is zero; ensure set by service layer")
	}

	event := &calendar.Event{
		Summary: fmt.Sprintf("%s - %s", appt.Service.Name, appt.CustomerName),
		Description: fmt.Sprintf("Услуга: %s\nПродолжительность: %d мин\nКлиент Telegram ID: %s\nПримечания: %s",
			appt.Service.Name, appt.Duration, appt.CustomerTgID, appt.Notes),
		Start: &calendar.EventDateTime{
			DateTime: appt.StartTime.Format(time.RFC3339),
			TimeZone: appt.StartTime.Location().String(),
		},
		End: &calendar.EventDateTime{
			DateTime: appt.EndTime.Format(time.RFC3339),
			TimeZone: appt.EndTime.Location().String(),
		},
		// Optional: Add attendees if needed
		// Attendees: []*calendar.EventAttendee{{Email: "your.email@example.com"}},
		Source: &calendar.EventSource{
			Title: "Massage Bot",
			Url:   "https://t.me/YOUR_BOT_USERNAME", // Update with your actual bot username
		},
	}

	createdEvent, err := a.client.Events.Insert(a.calendarID, event).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to create calendar event: %w", err)
	}

	// Update the original appointment object with details from the created Google Calendar event
	appt.ID = createdEvent.Id
	appt.ClientID = createdEvent.Id // Google Calendar event ID
	// ClientName, Notes, etc. are already on appt, can be stored if desired
	log.Printf("Google Calendar Event created: %s, ID: %s", createdEvent.HtmlLink, createdEvent.Id)
	return appt, nil
}

// FindAll fetches appointments from Google Calendar within a specific time range.
func (a *adapter) FindAll(ctx context.Context) ([]domain.Appointment, error) {
	var appointments []domain.Appointment
	now := time.Now()
	// Fetch events for the next few months to accurately check for overlaps
	timeMin := now.Format(time.RFC3339)
	// Fetch up to 6 months in the future to cover potential bookings
	timeMax := now.Add(6 * 30 * 24 * time.Hour).Format(time.RFC3339)

	events, err := a.client.Events.List(a.calendarID).TimeMin(timeMin).TimeMax(timeMax).SingleEvents(true).OrderBy("startTime").Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve calendar events: %w", err)
	}

	for _, item := range events.Items {
		// Skip all-day events or malformed entries
		if item.Start == nil || item.End == nil || item.Start.DateTime == "" || item.End.DateTime == "" {
			continue
		}

		startTime, err := time.Parse(time.RFC3339, item.Start.DateTime)
		if err != nil {
			log.Printf("Error parsing start time %s from event %s: %v", item.Start.DateTime, item.Id, err)
			continue
		}
		endTime, err := time.Parse(time.RFC3339, item.End.DateTime)
		if err != nil {
			log.Printf("Error parsing end time %s from event %s: %v", item.End.DateTime, item.Id, err)
			continue
		}

		duration := int(endTime.Sub(startTime).Minutes())

		// Populate other fields from summary and description as best as possible
		// This parsing might need to be more robust depending on how event summaries/descriptions are formatted
		customerName := ""
		customerTgID := ""
		serviceName := item.Summary // Default summary as service name if no other info
		notes := item.Description

		// Basic attempt to parse customerTgID and customerName from Description
		// e.g., "Клиент Telegram ID: 12345"
		// And "Клиент: John Doe" (if you pass full name)
		// For now, let's just assign summary as service name and description as notes.
		// You would need custom logic to parse these from the notes string
		// For example:
		// if strings.Contains(notes, "Клиент Telegram ID:") {
		//     // Parse customerTgID
		// }

		appointments = append(appointments, domain.Appointment{
			ID:           item.Id,
			ClientID:     item.Id, // Google Calendar event ID
			StartTime:    startTime,
			EndTime:      endTime,
			Duration:     duration,
			Service:      domain.Service{Name: serviceName, DurationMinutes: duration}, // Populate basic service info
			CustomerName: customerName,                                                 // Will be empty unless parsed from description/summary
			CustomerTgID: customerTgID,                                                 // Will be empty unless parsed
			Notes:        notes,
		})
	}
	return appointments, nil
}

// FindByID fetches a single event from Google Calendar by its ID.
func (a *adapter) FindByID(ctx context.Context, id string) (*domain.Appointment, error) {
	event, err := a.client.Events.Get(a.calendarID, id).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve calendar event by ID %s: %w", id, err)
	}

	if event.Start == nil || event.End == nil || event.Start.DateTime == "" || event.End.DateTime == "" {
		return nil, fmt.Errorf("malformed event data for ID %s", id)
	}

	startTime, err := time.Parse(time.RFC3339, event.Start.DateTime)
	if err != nil {
		return nil, fmt.Errorf("error parsing start time for event ID %s: %v", id, err)
	}
	endTime, err := time.Parse(time.RFC3339, event.End.DateTime)
	if err != nil {
		return nil, fmt.Errorf("error parsing end time for event ID %s: %v", id, err)
	}

	duration := int(endTime.Sub(startTime).Minutes())

	// Populate other fields by parsing event.Summary and event.Description
	customerName := ""           // Placeholder
	customerTgID := ""           // Placeholder
	serviceName := event.Summary // Assuming summary contains service name
	notes := event.Description

	return &domain.Appointment{
		ID:           event.Id,
		ClientID:     event.Id,
		StartTime:    startTime,
		EndTime:      endTime,
		Duration:     duration,
		CustomerName: customerName, // You need to extract this from summary/description
		CustomerTgID: customerTgID, // You need to extract this
		Notes:        notes,
		Service:      domain.Service{Name: serviceName, DurationMinutes: duration},
	}, nil
}

// Delete deletes an event from Google Calendar by its ID.
func (a *adapter) Delete(ctx context.Context, id string) error {
	err := a.client.Events.Delete(a.calendarID, id).Context(ctx).Do()
	if err != nil {
		// CORRECTED: Return only the error, not nil, error
		return fmt.Errorf("failed to delete calendar event %s: %w", id, err)
	}
	log.Printf("Google Calendar Event %s deleted.", id)
	return nil
}
