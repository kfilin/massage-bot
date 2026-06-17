package appointment

import (
	"testing"
)

// TestPrometheusCollector_RecordAppointmentCreated exercises the
// happy path of the metrics collector. The Prometheus metric values
// aren't asserted (they live in the global monitoring package) — we
// just verify the calls don't panic and that the collector satisfies
// the MetricsCollector interface.
func TestPrometheusCollector_RecordAppointmentCreated(t *testing.T) {
	c := NewPrometheusCollector()
	c.RecordAppointmentCreated("massage-classic", 5.5)
	c.RecordAppointmentCreated("", 0) // edge: empty service name
	c.RecordAppointmentCreated("massage-deep", -1.0) // negative lead time
}

func TestPrometheusCollector_RecordAppointmentCancelled(t *testing.T) {
	c := NewPrometheusCollector()
	c.RecordAppointmentCancelled()
}

func TestPrometheusCollector_RecordFreeBusyCacheHit(t *testing.T) {
	c := NewPrometheusCollector()
	c.RecordFreeBusyCacheHit()
}

func TestPrometheusCollector_RecordFreeBusyCacheMiss(t *testing.T) {
	c := NewPrometheusCollector()
	c.RecordFreeBusyCacheMiss()
}

// TestNoOpCollector_AllMethods ensures the no-op collector satisfies
// the interface and can be called without panicking.
func TestNoOpCollector_AllMethods(t *testing.T) {
	var _ MetricsCollector = (*NoOpCollector)(nil)
	n := &NoOpCollector{}
	n.RecordAppointmentCreated("svc", 1.0)
	n.RecordAppointmentCancelled()
	n.RecordFreeBusyCacheHit()
	n.RecordFreeBusyCacheMiss()
}

// TestPrometheusCollector_ImplementsInterface is a compile-time
// guarantee that PrometheusCollector satisfies MetricsCollector.
func TestPrometheusCollector_ImplementsInterface(t *testing.T) {
	var _ MetricsCollector = (*PrometheusCollector)(nil)
}
