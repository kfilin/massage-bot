package googlecalendar

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/kfilin/massage-bot/internal/domain" // Импортируем domain
	"github.com/kfilin/massage-bot/internal/ports"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/googleapi"
)

type adapter struct {
	client     *calendar.Service
	calendarID string
}

// NewAdapter creates a new Google Calendar adapter that implements ports.AppointmentRepository.
// It now accepts the Google Calendar ID from configuration.
func NewAdapter(client *calendar.Service, calendarID string) ports.AppointmentRepository {
	if calendarID == "" {
		log.Println("Warning: Google Calendar ID is empty, defaulting to 'primary'. Ensure GOOGLE_CALENDAR_ID is set in config.")
		calendarID = "primary" // Fallback to primary if empty, though config should handle this
	}
	return &adapter{
		client:     client,
		calendarID: calendarID,
	}
}

// Create creates a new appointment event in Google Calendar.
func (a *adapter) Create(ctx context.Context, appt *domain.Appointment) (*domain.Appointment, error) {
	if appt.StartTime.IsZero() || appt.EndTime.IsZero() {
		return nil, fmt.Errorf("appointment StartTime or EndTime is zero; ensure set by service layer")
	}

	description := appt.Notes
	if appt.CustomerTgID != "" {
		description = fmt.Sprintf("TGID:%s\n%s", appt.CustomerTgID, appt.Notes)
	}

	event := &calendar.Event{
		Summary:     fmt.Sprintf("%s - %s", appt.Service.Name, appt.CustomerName),
		Description: description,
		Start: &calendar.EventDateTime{
			DateTime: appt.StartTime.Format(time.RFC3339),
			TimeZone: appt.StartTime.Location().String(),
		},
		End: &calendar.EventDateTime{
			DateTime: appt.EndTime.Format(time.RFC3339),
			TimeZone: appt.EndTime.Location().String(),
		},
		// Add attendees if necessary, e.g., []string{appt.CustomerEmail}
		// Assuming for now it's a simple booking on the calendar.
	}

	createdEvent, err := a.client.Events.Insert(a.calendarID, event).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to create calendar event: %w", err)
	}

	appt.ID = createdEvent.Id // Store the Google Calendar event ID
	log.Printf("Google Calendar event created: %s (ID: %s)", createdEvent.Summary, createdEvent.Id)

	return appt, nil
}

// GetAvailableSlots fetches available time slots from Google Calendar.
// This is a placeholder and needs actual implementation for checking free/busy times.
func (a *adapter) GetAvailableSlots(ctx context.Context, date time.Time, durationMinutes int) ([]time.Time, error) {
	// --- Placeholder for actual free/busy query to Google Calendar ---
	// This is a complex logic involving Free/Busy API, checking existing events, and calculating gaps.
	// For now, we'll return some mock data or simply indicate that this needs implementation.

	// Example: Get existing events for the day to find busy slots
	timeMin := date.Format(time.RFC3339)
	timeMax := date.Add(24 * time.Hour).Format(time.RFC3339)

	events, err := a.client.Events.List(a.calendarID).
		ShowDeleted(false).
		SingleEvents(true).
		TimeMin(timeMin).
		TimeMax(timeMax).
		OrderBy("startTime").
		Do()
	if err != nil {
		return nil, fmt.Errorf("failed to list calendar events: %w", err)
	}

	// This part needs to be refined to genuinely calculate available slots
	// based on working hours, existing events, and slot duration.
	// For now, it's a very basic example assuming we *might* find a slot.
	availableSlots := []time.Time{}

	// Define typical working hours for the day in the correct timezone
	workStart := time.Date(date.Year(), date.Month(), date.Day(), domain.WorkDayStartHour, 0, 0, 0, domain.ApptTimeZone) // Используем domain.WorkDayStartHour и domain.ApptTimeZone
	workEnd := time.Date(date.Year(), date.Month(), date.Day(), domain.WorkDayEndHour, 0, 0, 0, domain.ApptTimeZone)     // Используем domain.WorkDayEndHour и domain.ApptTimeZone

	// Simple iteration (needs to consider existing events properly)
	currentTime := workStart
	for currentTime.Before(workEnd) {
		slotEnd := currentTime.Add(*domain.SlotDuration) // Используем domain.SlotDuration
		if slotEnd.After(workEnd) {
			break // Slot extends past working hours
		}

		isSlotBusy := false
		for _, event := range events.Items {
			eventStart, _ := time.Parse(time.RFC3339, event.Start.DateTime)
			eventEnd, _ := time.Parse(time.RFC3339, event.End.DateTime)

			// Check for overlap: [start, end)
			if currentTime.Before(eventEnd) && slotEnd.After(eventStart) {
				isSlotBusy = true
				break
			}
		}

		if !isSlotBusy {
			availableSlots = append(availableSlots, currentTime)
		}
		currentTime = slotEnd // Move to the next potential slot
	}

	log.Printf("Found %d available slots for %s", len(availableSlots), date.Format("2006-01-02"))
	return availableSlots, nil
}

// FindAll fetches all events (appointments) from Google Calendar.
// Placeholder implementation for now.
func (a *adapter) FindAll(ctx context.Context) ([]domain.Appointment, error) {
	events, err := a.client.Events.List(a.calendarID).
		ShowDeleted(false).
		SingleEvents(true).
		TimeMin(time.Now().Format(time.RFC3339)). // From now
		OrderBy("startTime").
		Do()
	if err != nil {
		return nil, fmt.Errorf("failed to list all calendar events: %w", err)
	}

	var appointments []domain.Appointment
	for _, event := range events.Items {
		appt, err := eventToAppointment(event)
		if err != nil {
			log.Printf("Warning: failed to convert event %s to appointment: %v", event.Id, err)
			continue
		}
		appointments = append(appointments, *appt)
	}
	return appointments, nil
}

// FindByID retrieves an event from Google Calendar by its ID.
func (a *adapter) FindByID(ctx context.Context, id string) (*domain.Appointment, error) {
	event, err := a.client.Events.Get(a.calendarID, id).Context(ctx).Do()
	if err != nil {
		if isNotFound(err) {
			return nil, domain.ErrAppointmentNotFound
		}
		return nil, fmt.Errorf("failed to get calendar event by ID %s: %w", id, err)
	}

	appt, err := eventToAppointment(event)
	if err != nil {
		return nil, fmt.Errorf("failed to convert event to appointment: %w", err)
	}
	return appt, nil
}

// Delete deletes an event from Google Calendar by its ID.
func (a *adapter) Delete(ctx context.Context, id string) error {
	err := a.client.Events.Delete(a.calendarID, id).Context(ctx).Do()
	if err != nil {
		if isNotFound(err) {
			return domain.ErrAppointmentNotFound
		}
		return fmt.Errorf("failed to delete calendar event: %w", err)
	}
	return nil
}

// Helper to check if an error indicates "not found"
func isNotFound(err error) bool {
	// Google API errors are often of type *googleapi.Error
	// Check the status code. 404 is typically "not found"
	if gErr, ok := err.(*googleapi.Error); ok {
		return gErr.Code == http.StatusNotFound
	}
	return false
}

// eventToAppointment converts a Google Calendar Event to a domain.Appointment.
func eventToAppointment(event *calendar.Event) (*domain.Appointment, error) {
	if event == nil || event.Id == "" || event.Start == nil || event.End == nil {
		return nil, fmt.Errorf("malformed event data for ID %s", event.Id)
	}

	startTime, err := time.Parse(time.RFC3339, event.Start.DateTime)
	if err != nil {
		return nil, fmt.Errorf("error parsing start time for event ID %s: %v", event.Id, err)
	}
	endTime, err := time.Parse(time.RFC3339, event.End.DateTime)
	if err != nil {
		return nil, fmt.Errorf("error parsing end time for event ID %s: %v", event.Id, err)
	}

	duration := int(endTime.Sub(startTime).Minutes())

	// Populate other fields by parsing event.Summary and event.Description
	customerTgID := ""
	notes := event.Description

	// Extract TGID if present in description
	// Format: TGID:123456789\nNotes
	if len(event.Description) > 5 && event.Description[:5] == "TGID:" {
		var extractedID string
		var remainingNotes string
		n, _ := fmt.Sscanf(event.Description, "TGID:%s", &extractedID)
		if n > 0 {
			// Find the newline to get the rest of the notes
			for i, char := range event.Description {
				if char == '\n' {
					customerTgID = extractedID
					remainingNotes = event.Description[i+1:]
					break
				}
			}
			if customerTgID == "" { // No newline found
				customerTgID = extractedID
				remainingNotes = ""
			}
			notes = remainingNotes
		}
	}

	customerName := "" // Still placeholder, usually derived from summary
	if len(event.Summary) > 0 {
		// Summary format: "Service Name - Customer Name"
		// This is a simple heuristic
		parts := domain.SplitSummary(event.Summary)
		if len(parts) >= 2 {
			customerName = parts[1]
		}
	}

	serviceName := event.Summary
	if parts := domain.SplitSummary(event.Summary); len(parts) >= 1 {
		serviceName = parts[0]
	}

	return &domain.Appointment{
		ID:           event.Id,
		ClientID:     event.Id, // Assuming ClientID is the same as Google Event ID
		StartTime:    startTime,
		EndTime:      endTime,
		Duration:     duration,
		CustomerName: customerName,
		CustomerTgID: customerTgID,
		Notes:        notes,
		Service:      domain.Service{Name: serviceName, DurationMinutes: duration}, // Populate service details
	}, nil
}
