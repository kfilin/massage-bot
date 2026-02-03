package appointment

import (
	"fmt"
	"time"

	"github.com/kfilin/massage-bot/internal/monitoring"
)

// MetricsCollector defines the interface for recording business metrics.
// This abstraction allows for easier testing and swapping of monitoring backends.
type MetricsCollector interface {
	RecordAppointmentCreated(serviceName string, leadTimeDays float64)
	RecordAppointmentCancelled()
	RecordFreeBusyCacheHit()
	RecordFreeBusyCacheMiss()
}

// PrometheusCollector implements MetricsCollector using the global monitoring package.
type PrometheusCollector struct{}

// NewPrometheusCollector creates a new metrics collector.
func NewPrometheusCollector() *PrometheusCollector {
	return &PrometheusCollector{}
}

func (p *PrometheusCollector) RecordAppointmentCreated(serviceName string, leadTimeDays float64) {
	monitoring.BookingLeadTimeDays.Observe(leadTimeDays)
	monitoring.ServiceBookingsTotal.WithLabelValues(serviceName).Inc()
	monitoring.BookingCreationHour.WithLabelValues(fmt.Sprintf("%02d", time.Now().Hour())).Inc()
}

func (p *PrometheusCollector) RecordAppointmentCancelled() {
	monitoring.CancellationsTotal.WithLabelValues("unknown").Inc()
}

func (p *PrometheusCollector) RecordFreeBusyCacheHit() {
	monitoring.FreeBusyCacheHits.Inc()
}

func (p *PrometheusCollector) RecordFreeBusyCacheMiss() {
	monitoring.FreeBusyCacheMisses.Inc()
}

// NoOpCollector for testing
type NoOpCollector struct{}

func (n *NoOpCollector) RecordAppointmentCreated(serviceName string, leadTimeDays float64) {}
func (n *NoOpCollector) RecordAppointmentCancelled()                                       {}
func (n *NoOpCollector) RecordFreeBusyCacheHit()                                           {}
func (n *NoOpCollector) RecordFreeBusyCacheMiss()                                          {}
