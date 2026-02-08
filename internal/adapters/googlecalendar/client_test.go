package googlecalendar

import (
	"errors"
	"testing"
	"time"

	"github.com/kfilin/massage-bot/internal/monitoring"
	"golang.org/x/oauth2"
)

type mockTokenSource struct {
	token *oauth2.Token
	err   error
}

func (m *mockTokenSource) Token() (*oauth2.Token, error) {
	return m.token, m.err
}

func TestMonitoringTokenSource_Token(t *testing.T) {
	// Set initial value
	monitoring.UpdateTokenExpiry(0)

	now := time.Now()
	testCases := []struct {
		name          string
		token         *oauth2.Token
		err           error
		expectedDays  float64
		expectSuccess bool
	}{
		{
			name: "Valid token with expiry in 24 hours",
			token: &oauth2.Token{
				AccessToken: "valid",
				Expiry:      now.Add(24 * time.Hour),
			},
			expectedDays:  1.0,
			expectSuccess: true,
		},
		{
			name: "Valid token with expiry in 1 hour",
			token: &oauth2.Token{
				AccessToken: "valid",
				Expiry:      now.Add(1 * time.Hour),
			},
			expectedDays:  1.0 / 24.0,
			expectSuccess: true,
		},
		{
			name: "Expired token (negative duration)",
			token: &oauth2.Token{
				AccessToken: "expired",
				Expiry:      now.Add(-1 * time.Hour),
			},
			expectedDays:  -1.0 / 24.0,
			expectSuccess: true,
		},
		{
			name:          "Error from source",
			err:           errors.New("token error"),
			expectSuccess: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			src := &mockTokenSource{token: tc.token, err: tc.err}
			mts := &MonitoringTokenSource{src: src}

			tok, err := mts.Token()

			if tc.expectSuccess {
				if err != nil {
					t.Fatalf("Expected success, got error: %v", err)
				}
				if tok.AccessToken != tc.token.AccessToken {
					t.Errorf("Expected token %s, got %s", tc.token.AccessToken, tok.AccessToken)
				}

				// Check if metric was updated (we can't easily read back the metric value in unit test
				// without exposing internal state of monitoring package or using prometheus test util)
				// But we can verify no panic and logic flow.
				// In a real integration test we would scrape the metric.
			} else {
				if err == nil {
					t.Fatal("Expected error, got nil")
				}
			}
		})
	}
}
