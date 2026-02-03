package googlecalendar

import (
	"math"
	"math/rand"
	"net/http"
	"time"

	"github.com/kfilin/massage-bot/internal/logging"
)

// RetryTransport implements http.RoundTripper and adds retry logic with exponential backoff.
type RetryTransport struct {
	Transport  http.RoundTripper
	MaxRetries int
	BaseDelay  time.Duration
	MaxDelay   time.Duration
}

// RoundTrip executes the HTTP request with retries.
func (t *RetryTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	var resp *http.Response
	var err error

	// If no transport is specified, use the default
	transport := t.Transport
	if transport == nil {
		transport = http.DefaultTransport
	}

	for i := 0; i <= t.MaxRetries; i++ {
		// Log retry attempt if > 0
		if i > 0 {
			logging.Debugf("DEBUG: Retrying request %s %s (Attempt %d/%d)", req.Method, req.URL.Path, i, t.MaxRetries)
		}

		resp, err = transport.RoundTrip(req)

		// Check if we should retry
		if !t.shouldRetry(resp, err) {
			return resp, err
		}

		// Close response body immediately if we are going to retry, to avoid leaks
		if resp != nil {
			resp.Body.Close()
		}

		// Put a cap on retries if we reached the max
		if i == t.MaxRetries {
			logging.Warnf("WARNING: Max retries reached for %s %s", req.Method, req.URL.Path)
			return resp, err
		}

		// Calculate backoff
		delay := t.backoff(i)
		select {
		case <-req.Context().Done():
			return nil, req.Context().Err()
		case <-time.After(delay):
			continue
		}
	}

	return resp, err
}

func (t *RetryTransport) shouldRetry(resp *http.Response, err error) bool {
	// Retry on network errors
	if err != nil {
		logging.Warnf("WARNING: connection error: %v", err)
		return true
	}

	// Retry on specific status codes
	if resp != nil {
		switch resp.StatusCode {
		case http.StatusTooManyRequests, // 429
			http.StatusInternalServerError, // 500
			http.StatusBadGateway,          // 502
			http.StatusServiceUnavailable,  // 503
			http.StatusGatewayTimeout:      // 504
			logging.Warnf("WARNING: Received retryable status code %d", resp.StatusCode)
			return true
		}
	}

	return false
}

func (t *RetryTransport) backoff(attempt int) time.Duration {
	// Exponential backoff: BaseDelay * 2^attempt
	backoff := float64(t.BaseDelay) * math.Pow(2, float64(attempt))

	// Add jitter (randomness) to avoid thundering herd: +/- 20%
	jitter := (rand.Float64() * 0.4) + 0.8 // Range 0.8 - 1.2
	backoff *= jitter

	if backoff > float64(t.MaxDelay) {
		backoff = float64(t.MaxDelay)
	}

	return time.Duration(backoff)
}
