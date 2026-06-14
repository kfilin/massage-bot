package googlecalendar

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/kfilin/massage-bot/internal/monitoring"
	"golang.org/x/oauth2"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
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
			name: "Token with zero expiry",
			token: &oauth2.Token{
				AccessToken: "no-expiry",
				Expiry:      time.Time{},
			},
			expectedDays:  0,
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
			} else {
				if err == nil {
					t.Fatal("Expected error, got nil")
				}
			}
		})
	}
}

func TestTokenFromFile(t *testing.T) {
	t.Run("File does not exist", func(t *testing.T) {
		_, err := tokenFromFile("/nonexistent/path/token.json")
		if err == nil {
			t.Error("tokenFromFile() should return error for nonexistent file")
		}
	})
}

func TestNewAdapter_EmptyCalendarID(t *testing.T) {
	ctx := context.Background()
	svc, err := calendar.NewService(ctx, option.WithoutAuthentication())
	if err != nil {
		t.Fatalf("Failed to create calendar service: %v", err)
	}
	a := NewAdapter(svc, "")
	if a == nil {
		t.Fatal("NewAdapter() returned nil")
	}
	casted, ok := a.(*adapter)
	if !ok {
		t.Fatal("NewAdapter() did not return *adapter")
	}
	if casted.calendarID != "primary" {
		t.Errorf("NewAdapter() calendarID = %s, want primary", casted.calendarID)
	}
}
