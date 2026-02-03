package logging

import (
	"os"
	"strings"
	"sync"
	"testing"
)

func TestRedactPII(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "No PII",
			input: "Hello World",
			want:  "Hello World",
		},
		{
			name:  "Telegram ID",
			input: "User 123456789 connected",
			want:  "User [REDACTED_ID] connected",
		},
		{
			name:  "Short number (safe)",
			input: "User 123 connected",
			want:  "User 123 connected",
		},
		{
			name:  "Multiple IDs",
			input: "User 123456789 and 987654321",
			want:  "User [REDACTED_ID] and [REDACTED_ID]",
		},
		{
			name:  "Email Address",
			input: "Contact test@example.com for support",
			want:  "Contact [REDACTED_EMAIL] for support",
		},
		{
			name:  "Phone Number",
			input: "Call 5312345678",
			want:  "Call [REDACTED_ID]", // Collides with number regex, which is acceptable for safe defaults
		},
		{
			name:  "Multiple emails",
			input: "Send to user@example.com and admin@test.org",
			want:  "Send to [REDACTED_EMAIL] and [REDACTED_EMAIL]",
		},
		{
			name:  "Mixed PII",
			input: "User 987654321 email: test@example.com phone: 1234567890",
			want:  "User [REDACTED_ID] email: [REDACTED_EMAIL] phone: [REDACTED_ID]",
		},
		{
			name:  "Empty string",
			input: "",
			want:  "",
		},
		{
			name:  "Only numbers (short)",
			input: "12345678",
			want:  "12345678",
		},
		{
			name:  "Only numbers (long)",
			input: "123456789",
			want:  "[REDACTED_ID]",
		},
		{
			name:  "Email with plus addressing",
			input: "user+tag@example.com",
			want:  "[REDACTED_EMAIL]",
		},
		{
			name:  "Email with subdomain",
			input: "admin@mail.example.com",
			want:  "[REDACTED_EMAIL]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := RedactPII(tt.input); got != tt.want {
				t.Errorf("RedactPII() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestInit verifies logger initialization
func TestInit(t *testing.T) {
	t.Run("Production mode", func(t *testing.T) {
		once = sync.Once{}
		logger = nil

		Init(false)

		if logger == nil {
			t.Fatal("Logger is nil after Init(false)")
		}

		// Verify we can log without panic
		logger.Info("test message")
	})

	t.Run("Debug mode", func(t *testing.T) {
		once = sync.Once{}
		logger = nil

		Init(true)

		if logger == nil {
			t.Fatal("Logger is nil after Init(true)")
		}

		// Verify we can log without panic
		logger.Debug("debug message")
	})
}

// TestGet verifies Get() initializes logger if not already done
func TestGet(t *testing.T) {
	once = sync.Once{}
	logger = nil

	// Get should initialize with defaults
	l := Get()
	if l == nil {
		t.Fatal("Get() returned nil logger")
	}

	// Second call should return same logger
	l2 := Get()
	if l2 != l {
		t.Error("Get() returned different logger instance on second call")
	}
}

// TestGetWithEnvVar tests Get() respects LOG_LEVEL environment variable
func TestGetWithEnvVar(t *testing.T) {
	originalLogLevel := os.Getenv("LOG_LEVEL")
	defer os.Setenv("LOG_LEVEL", originalLogLevel)

	t.Run("With DEBUG env var", func(t *testing.T) {
		once = sync.Once{}
		logger = nil
		os.Setenv("LOG_LEVEL", "DEBUG")

		l := Get()
		if l == nil {
			t.Fatal("Get() returned nil with DEBUG env var")
		}
	})

	t.Run("Without DEBUG env var", func(t *testing.T) {
		once = sync.Once{}
		logger = nil
		os.Setenv("LOG_LEVEL", "INFO")

		l := Get()
		if l == nil {
			t.Fatal("Get() returned nil with INFO env var")
		}
	})
}

// TestWrapperFunctions tests all convenience wrapper functions don't panic
func TestWrapperFunctions(t *testing.T) {
	once = sync.Once{}
	logger = nil
	Init(false)

	tests := []struct {
		name    string
		logFunc func()
	}{
		{"Info", func() { Info("info message") }},
		{"Infof", func() { Infof("formatted %s", "message") }},
		{"Debug", func() { Debug("debug message") }},
		{"Debugf", func() { Debugf("debug %d", 123) }},
		{"Warn", func() { Warn("warning message") }},
		{"Warnf", func() { Warnf("warning %s", "formatted") }},
		{"Error", func() { Error("error message") }},
		{"Errorf", func() { Errorf("error %v", "formatted") }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("%s panicked: %v", tt.name, r)
				}
			}()
			tt.logFunc()
		})
	}
}

// TestConcurrentAccess tests thread-safety of logger access
func TestConcurrentAccess(t *testing.T) {
	once = sync.Once{}
	logger = nil

	const goroutines = 10
	const iterations = 10

	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				l := Get()
				if l == nil {
					t.Errorf("Got nil logger in goroutine %d", id)
				}
				// Just verify logging doesn't panic
				Infof("Goroutine %d iteration %d", id, j)
			}
		}(i)
	}

	wg.Wait()
}

// TestRedactArgsWithNonStringTypes tests redactArgs with various types
func TestRedactArgsWithNonStringTypes(t *testing.T) {
	args := []interface{}{
		"test@example.com",
		123,
		true,
		nil,
		struct{ Name string }{Name: "test"},
	}

	redacted := redactArgs(args)

	if len(redacted) != len(args) {
		t.Errorf("redactArgs changed slice length: got %d, want %d", len(redacted), len(args))
	}

	// First arg (string) should be redacted
	if !strings.Contains(redacted[0].(string), "[REDACTED_EMAIL]") {
		t.Errorf("String arg not redacted: %v", redacted[0])
	}

	// Other args should pass through unchanged
	if redacted[1] != 123 {
		t.Errorf("Int arg changed: got %v, want 123", redacted[1])
	}
	if redacted[2] != true {
		t.Errorf("Bool arg changed: got %v, want true", redacted[2])
	}
}

// BenchmarkRedactPII benchmarks the PII redaction function
func BenchmarkRedactPII(b *testing.B) {
	input := "User 123456789 email test@example.com contacted support"
	for i := 0; i < b.N; i++ {
		RedactPII(input)
	}
}

// BenchmarkLogging benchmarks logging performance
func BenchmarkLogging(b *testing.B) {
	once = sync.Once{}
	logger = nil
	Init(false)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Infof("Benchmark message %d", i)
	}
}
