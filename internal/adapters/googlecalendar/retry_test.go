package googlecalendar

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// MockTransport allows us to simulate responses and errors
type MockTransport struct {
	RoundTripFunc func(req *http.Request) (*http.Response, error)
}

func (m *MockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return m.RoundTripFunc(req)
}

func TestRetryTransport_Success(t *testing.T) {
	mockT := &MockTransport{
		RoundTripFunc: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewBufferString("OK")),
			}, nil
		},
	}

	retryT := &RetryTransport{
		Transport:  mockT,
		MaxRetries: 3,
		BaseDelay:  time.Millisecond,
		MaxDelay:   time.Second,
	}

	client := &http.Client{Transport: retryT}
	resp, err := client.Get("http://example.com")

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestRetryTransport_Recover(t *testing.T) {
	attempts := 0
	mockT := &MockTransport{
		RoundTripFunc: func(req *http.Request) (*http.Response, error) {
			attempts++
			if attempts < 3 {
				return &http.Response{
					StatusCode: http.StatusServiceUnavailable, // 503 -> Retry
					Body:       io.NopCloser(bytes.NewBufferString("Fail")),
				}, nil
			}
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewBufferString("OK")),
			}, nil
		},
	}

	retryT := &RetryTransport{
		Transport:  mockT,
		MaxRetries: 3,
		BaseDelay:  time.Millisecond,
		MaxDelay:   time.Second,
	}

	client := &http.Client{Transport: retryT}
	resp, err := client.Get("http://example.com")

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, 3, attempts) // Initial + 2 retries
}

func TestRetryTransport_FailAfterMaxRetries(t *testing.T) {
	attempts := 0
	mockT := &MockTransport{
		RoundTripFunc: func(req *http.Request) (*http.Response, error) {
			attempts++
			return &http.Response{
				StatusCode: http.StatusGatewayTimeout, // 504 -> Retry
				Body:       io.NopCloser(bytes.NewBufferString("Fail")),
			}, nil
		},
	}

	retryT := &RetryTransport{
		Transport:  mockT,
		MaxRetries: 2,
		BaseDelay:  time.Millisecond,
		MaxDelay:   time.Second,
	}

	client := &http.Client{Transport: retryT}
	resp, err := client.Get("http://example.com")

	assert.NoError(t, err)
	assert.Equal(t, http.StatusGatewayTimeout, resp.StatusCode)
	assert.Equal(t, 3, attempts) // Initial + 2 retries
}

func TestRetryTransport_NetworkError(t *testing.T) {
	attempts := 0
	mockT := &MockTransport{
		RoundTripFunc: func(req *http.Request) (*http.Response, error) {
			attempts++
			return nil, errors.New("connection reset")
		},
	}

	retryT := &RetryTransport{
		Transport:  mockT,
		MaxRetries: 1,
		BaseDelay:  time.Millisecond,
		MaxDelay:   time.Second,
	}

	client := &http.Client{Transport: retryT}
	_, err := client.Get("http://example.com")

	assert.Error(t, err)
	assert.Equal(t, 2, attempts) // Initial + 1 retry
}

func TestRetryTransport_NilTransport(t *testing.T) {
	retryT := &RetryTransport{
		Transport:  nil,
		MaxRetries: 1,
		BaseDelay:  time.Millisecond,
		MaxDelay:   time.Second,
	}

	client := &http.Client{Transport: retryT}
	resp, err := client.Get("http://example.com")

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestRetryTransport_ContextCancelled(t *testing.T) {
	mockT := &MockTransport{
		RoundTripFunc: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusServiceUnavailable,
				Body:       io.NopCloser(bytes.NewBufferString("Fail")),
			}, nil
		},
	}

	retryT := &RetryTransport{
		Transport:  mockT,
		MaxRetries: 5,
		BaseDelay:  1 * time.Millisecond,
		MaxDelay:   1 * time.Second,
	}

	ctx, cancel := context.WithCancel(context.Background())
	req, _ := http.NewRequestWithContext(ctx, "GET", "http://example.com", nil)
	cancel()

	_, err := retryT.RoundTrip(req)

	assert.Error(t, err)
	assert.Equal(t, context.Canceled, err)
}

func TestBackoff_MaxDelayCap(t *testing.T) {
	rt := &RetryTransport{
		BaseDelay: 100 * time.Millisecond,
		MaxDelay:  200 * time.Millisecond,
	}

	delay := rt.backoff(10) // 100ms * 2^10 = ~102s, should be capped at 200ms
	assert.LessOrEqual(t, delay, 200*time.Millisecond)
	assert.GreaterOrEqual(t, delay, 160*time.Millisecond) // Allow for jitter range
}

func TestRetryTransport_NonRetryableStatus(t *testing.T) {
	mockT := &MockTransport{
		RoundTripFunc: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusBadRequest,
				Body:       io.NopCloser(bytes.NewBufferString("Bad Request")),
			}, nil
		},
	}

	retryT := &RetryTransport{
		Transport:  mockT,
		MaxRetries: 3,
		BaseDelay:  time.Millisecond,
		MaxDelay:   time.Second,
	}

	client := &http.Client{Transport: retryT}
	resp, err := client.Get("http://example.com")

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}
