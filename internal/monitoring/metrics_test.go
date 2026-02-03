package monitoring

import (
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

// TestMetricsRegistration verifies all metrics are properly registered
func TestMetricsRegistration(t *testing.T) {
	tests := []struct {
		name   string
		metric prometheus.Collector
	}{
		{"FreeBusyCacheHits", FreeBusyCacheHits},
		{"FreeBusyCacheMisses", FreeBusyCacheMisses},
		{"ApiRequestsTotal", ApiRequestsTotal},
		{"ApiLatency", ApiLatency},
		{"DbErrorsTotal", DbErrorsTotal},
		{"BotCommandsTotal", BotCommandsTotal},
		{"BookingsTotal", BookingsTotal},
		{"BookingLeadTimeDays", BookingLeadTimeDays},
		{"AppointmentTypeTotal", AppointmentTypeTotal},
		{"ClinicalNoteLength", ClinicalNoteLength},
		{"BookingCreationHour", BookingCreationHour},
		{"ServiceBookingsTotal", ServiceBookingsTotal},
		{"CancellationsTotal", CancellationsTotal},
		{"ActiveSessions", ActiveSessions},
		{"TokenExpiryDays", TokenExpiryDays},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.metric == nil {
				t.Errorf("%s is nil, should be registered", tt.name)
			}
		})
	}
}

// TestIncrementBooking tests the IncrementBooking helper function
func TestIncrementBooking(t *testing.T) {
	// Get initial value
	initialTotal := GetTotalBookings()

	// Increment booking
	serviceName := "test-massage"
	IncrementBooking(serviceName)

	// Verify internal counter incremented
	newTotal := GetTotalBookings()
	if newTotal != initialTotal+1 {
		t.Errorf("GetTotalBookings() = %d, want %d", newTotal, initialTotal+1)
	}

	// Verify Prometheus metric was incremented
	metric := &dto.Metric{}
	counter, err := BookingsTotal.GetMetricWithLabelValues(serviceName)
	if err != nil {
		t.Fatalf("Failed to get metric: %v", err)
	}
	if err := counter.Write(metric); err != nil {
		t.Fatalf("Failed to write metric: %v", err)
	}

	// The counter should have at least 1 (could be more if tests run multiple times)
	if metric.Counter.GetValue() < 1 {
		t.Errorf("BookingsTotal counter = %f, want >= 1", metric.Counter.GetValue())
	}
}

// TestUpdateTokenExpiry tests the UpdateTokenExpiry function
func TestUpdateTokenExpiry(t *testing.T) {
	tests := []struct {
		name string
		days float64
	}{
		{"30 days", 30.0},
		{"7 days", 7.0},
		{"0 days (expired)", 0.0},
		{"negative (already expired)", -5.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			UpdateTokenExpiry(tt.days)

			// Read the gauge value
			metric := &dto.Metric{}
			if err := TokenExpiryDays.Write(metric); err != nil {
				t.Fatalf("Failed to write metric: %v", err)
			}

			if metric.Gauge.GetValue() != tt.days {
				t.Errorf("TokenExpiryDays = %f, want %f", metric.Gauge.GetValue(), tt.days)
			}
		})
	}
}

// TestUpdateActiveSessions tests the UpdateActiveSessions function
func TestUpdateActiveSessions(t *testing.T) {
	tests := []struct {
		name  string
		count int
	}{
		{"zero sessions", 0},
		{"one session", 1},
		{"multiple sessions", 42},
		{"large number", 1000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			UpdateActiveSessions(tt.count)

			// Verify internal counter
			sessions := GetActiveSessions()
			if sessions != int64(tt.count) {
				t.Errorf("GetActiveSessions() = %d, want %d", sessions, tt.count)
			}

			// Verify Prometheus gauge
			metric := &dto.Metric{}
			if err := ActiveSessions.Write(metric); err != nil {
				t.Fatalf("Failed to write metric: %v", err)
			}

			if int(metric.Gauge.GetValue()) != tt.count {
				t.Errorf("ActiveSessions gauge = %f, want %d", metric.Gauge.GetValue(), tt.count)
			}
		})
	}
}

// TestGetTotalBookings tests the GetTotalBookings getter
func TestGetTotalBookings(t *testing.T) {
	// This is a simple getter, just verify it doesn't panic and returns a value
	total := GetTotalBookings()
	if total < 0 {
		t.Errorf("GetTotalBookings() = %d, should not be negative", total)
	}
}

// TestGetActiveSessions tests the GetActiveSessions getter
func TestGetActiveSessions(t *testing.T) {
	// Set a known value
	UpdateActiveSessions(123)

	sessions := GetActiveSessions()
	if sessions != 123 {
		t.Errorf("GetActiveSessions() = %d, want 123", sessions)
	}
}

// TestStartTime verifies StartTime is initialized
func TestStartTime(t *testing.T) {
	if StartTime.IsZero() {
		t.Error("StartTime is zero, should be initialized in init()")
	}

	// StartTime should be in the past (or very recent)
	if time.Since(StartTime) < 0 {
		t.Error("StartTime is in the future, should be in the past")
	}

	// StartTime should be reasonable (not more than 1 hour ago for a test run)
	if time.Since(StartTime) > 1*time.Hour {
		t.Logf("Warning: StartTime is %v ago, which seems old for a test run", time.Since(StartTime))
	}
}

// TestCounterVecLabels tests that CounterVec metrics accept expected labels
func TestCounterVecLabels(t *testing.T) {
	tests := []struct {
		name   string
		metric *prometheus.CounterVec
		labels []string
	}{
		{
			name:   "ApiRequestsTotal",
			metric: ApiRequestsTotal,
			labels: []string{"google_calendar", "create_event", "success"},
		},
		{
			name:   "DbErrorsTotal",
			metric: DbErrorsTotal,
			labels: []string{"save_patient"},
		},
		{
			name:   "BotCommandsTotal",
			metric: BotCommandsTotal,
			labels: []string{"/start"},
		},
		{
			name:   "BookingsTotal",
			metric: BookingsTotal,
			labels: []string{"massage-60"},
		},
		{
			name:   "AppointmentTypeTotal",
			metric: AppointmentTypeTotal,
			labels: []string{"first_visit"},
		},
		{
			name:   "BookingCreationHour",
			metric: BookingCreationHour,
			labels: []string{"14"},
		},
		{
			name:   "ServiceBookingsTotal",
			metric: ServiceBookingsTotal,
			labels: []string{"Deep Tissue Massage"},
		},
		{
			name:   "CancellationsTotal",
			metric: CancellationsTotal,
			labels: []string{"Classic Massage"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Try to get metric with labels - should not panic
			counter, err := tt.metric.GetMetricWithLabelValues(tt.labels...)
			if err != nil {
				t.Errorf("Failed to get metric with labels %v: %v", tt.labels, err)
			}
			if counter == nil {
				t.Error("Got nil counter")
			}
		})
	}
}

// TestHistogramVecLabels tests that HistogramVec metrics accept expected labels
func TestHistogramVecLabels(t *testing.T) {
	tests := []struct {
		name   string
		labels []string
	}{
		{
			name:   "Google Calendar API",
			labels: []string{"google_calendar", "list_events"},
		},
		{
			name:   "Telegram API",
			labels: []string{"telegram", "send_message"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Try to get histogram with labels - should not panic
			observer, err := ApiLatency.GetMetricWithLabelValues(tt.labels...)
			if err != nil {
				t.Errorf("Failed to get histogram with labels %v: %v", tt.labels, err)
			}
			if observer == nil {
				t.Error("Got nil observer")
			}

			// Try to observe a value
			observer.Observe(0.123)
		})
	}
}

// TestBookingLeadTimeBuckets tests the histogram buckets for booking lead time
func TestBookingLeadTimeBuckets(t *testing.T) {
	// Test various lead times
	testValues := []float64{0, 0.5, 1, 2, 3, 5, 7, 14, 30, 60}

	for _, days := range testValues {
		BookingLeadTimeDays.Observe(days)
	}

	// Verify the histogram was updated (no panic)
	metric := &dto.Metric{}
	if err := BookingLeadTimeDays.Write(metric); err != nil {
		t.Fatalf("Failed to write histogram metric: %v", err)
	}

	// Should have recorded all observations
	if metric.Histogram.GetSampleCount() < uint64(len(testValues)) {
		t.Errorf("Histogram sample count = %d, want >= %d",
			metric.Histogram.GetSampleCount(), len(testValues))
	}
}

// TestClinicalNoteLengthGauge tests the gauge for clinical note length
func TestClinicalNoteLengthGauge(t *testing.T) {
	testLengths := []float64{0, 50, 100, 500, 1000, 5000}

	for _, length := range testLengths {
		ClinicalNoteLength.Set(length)

		metric := &dto.Metric{}
		if err := ClinicalNoteLength.Write(metric); err != nil {
			t.Fatalf("Failed to write gauge metric: %v", err)
		}

		if metric.Gauge.GetValue() != length {
			t.Errorf("ClinicalNoteLength = %f, want %f", metric.Gauge.GetValue(), length)
		}
	}
}

// TestConcurrentMetricUpdates tests thread-safety of metric updates
func TestConcurrentMetricUpdates(t *testing.T) {
	const goroutines = 10
	const iterations = 100

	done := make(chan bool, goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			for j := 0; j < iterations; j++ {
				IncrementBooking("concurrent-test")
				UpdateActiveSessions(j)
				UpdateTokenExpiry(float64(j))
			}
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < goroutines; i++ {
		<-done
	}

	// Verify no panics occurred and metrics are accessible
	_ = GetTotalBookings()
	_ = GetActiveSessions()
}
