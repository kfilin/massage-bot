package ports

import (
	"context"
	"time"

	"github.com/kfilin/massage-bot/internal/domain"
)

// AppointmentService defines the interface for managing appointments (business logic layer).
type AppointmentService interface {
	GetAvailableServices(ctx context.Context) ([]domain.Service, error)
	GetAvailableTimeSlots(ctx context.Context, date time.Time, durationMinutes int) ([]domain.TimeSlot, error)
	CreateAppointment(ctx context.Context, appointment *domain.Appointment) (*domain.Appointment, error)
	CancelAppointment(ctx context.Context, appointmentID string) error
	GetCustomerAppointments(ctx context.Context, customerTgID string) ([]domain.Appointment, error)
	FindByID(ctx context.Context, appointmentID string) (*domain.Appointment, error)
}

// AppointmentRepository defines the interface for data persistence (e.g., Google Calendar).
// This is implemented by the Google Calendar adapter.
type AppointmentRepository interface {
	Create(ctx context.Context, appt *domain.Appointment) (*domain.Appointment, error)
	FindAll(ctx context.Context) ([]domain.Appointment, error) // For fetching all existing events
	FindByID(ctx context.Context, id string) (*domain.Appointment, error)
	Delete(ctx context.Context, id string) error
}

// SessionStorage defines the interface for managing user sessions (e.g., in-memory or Redis).
type SessionStorage interface {
	Set(userID int64, key string, value interface{})
	Get(userID int64) map[string]interface{}
	ClearSession(userID int64)
}
