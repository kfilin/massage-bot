package monitoring

import (
	"sync/atomic"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// Business metrics
	BookingsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "vera_bookings_total",
			Help: "Total number of bookings made",
		},
		[]string{"service"},
	)

	// System metrics
	ActiveSessions = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "vera_active_sessions",
			Help: "Number of active user sessions",
		},
	)

	// Token expiry warning
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
