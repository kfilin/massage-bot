package googlecalendar

import (
	"bytes"
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
