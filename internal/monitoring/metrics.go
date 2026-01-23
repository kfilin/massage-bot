package monitoring

import (
	"sync/atomic"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// --- Technical Metrics ---

	ApiRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "vera_api_requests_total",
			Help: "Total number of external API requests",
		},
		[]string{"provider", "operation", "status"},
	)

	ApiLatency = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "vera_api_latency_seconds",
			Help:    "Latency of external API requests in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"provider", "operation"},
	)

	DbErrorsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "vera_db_errors_total",
			Help: "Total number of database errors",
		},
		[]string{"operation"},
	)

	BotCommandsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "vera_bot_commands_total",
			Help: "Total number of bot commands processed",
		},
		[]string{"command"},
	)

	// --- Business & Patient Behavior Metrics ---

	BookingsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "vera_bookings_total",
			Help: "Total number of bookings made",
		},
		[]string{"service"},
	)

	BookingLeadTimeDays = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "vera_booking_lead_time_days",
			Help:    "How far in advance (days) appointments are booked",
			Buckets: []float64{0, 1, 2, 3, 5, 7, 14, 30},
		},
	)

	AppointmentTypeTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "vera_appointment_type_total",
			Help: "Count of bookings by patient type",
		},
		[]string{"type"}, // first_visit, returning
	)

	ClinicalNoteLength = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "vera_clinical_note_length_chars",
			Help: "Average length of therapist notes in characters",
		},
	)

	BookingCreationHour = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "vera_booking_creation_hour_total",
			Help: "Houly distribution of booking creations",
		},
		[]string{"hour"},
	)

	ServiceBookingsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "vera_service_bookings_total",
			Help: "Total bookings broken down by service type",
		},
		[]string{"service_name"},
	)

	CancellationsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "vera_cancellations_total",
			Help: "Total number of appointment cancellations",
		},
		[]string{"service_name"},
	)

	// --- System Status ---

	ActiveSessions = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "vera_active_sessions",
			Help: "Number of active user sessions",
		},
	)

	TokenExpiryDays = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "vera_token_expiry_days",
			Help: "Days until OAuth token expiry",
		},
	)

	// Internal counters for status display
	StartTime      = time.Now()
	totalBookings  int64
	activeSessions int64
)

// Helper functions
func IncrementBooking(serviceName string) {
	BookingsTotal.WithLabelValues(serviceName).Inc()
	atomic.AddInt64(&totalBookings, 1)
}

func UpdateTokenExpiry(days float64) {
	TokenExpiryDays.Set(days)
}

func UpdateActiveSessions(count int) {
	ActiveSessions.Set(float64(count))
	atomic.StoreInt64(&activeSessions, int64(count))
}

// Getter functions for status command
func GetTotalBookings() int64 {
	return atomic.LoadInt64(&totalBookings)
}

func GetActiveSessions() int64 {
	return atomic.LoadInt64(&activeSessions)
}
