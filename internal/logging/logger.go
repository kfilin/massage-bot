package logging

import (
	"os"
	"regexp"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	logger *zap.SugaredLogger
	once   sync.Once
	// Regex to identify 9+ digit numbers (likely Telegram IDs or Phones)
	// We use a simple \d{9,} to catch both for now, as it covers most phone numbers and IDs
	// For emails, we add a basic regex
	piiNumbersRegex = regexp.MustCompile(`\b\d{9,}\b`)
	piiEmailRegex   = regexp.MustCompile(`[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`)
)

// Init initializes the global logger.
// If debug is true, it uses development config (human-readable, debug level).
// Otherwise, it uses production config (JSON, info level).
func Init(debug bool) {
	once.Do(func() {
		initLogger(debug)
	})
}

// initLogger contains the actual logger initialization logic.
// This is called via once.Do from either Init() or Get().
func initLogger(debug bool) {
	var config zap.Config
	if debug {
		config = zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	} else {
		config = zap.NewProductionConfig()
		config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	}

	l, err := config.Build()
	if err != nil {
		panic(err)
	}

	// We use SugaredLogger for printf-style formatting compatibility
	logger = l.Sugar()
}

// Get returns the global logger instance. Initializes with defaults if not already done.
// This is thread-safe due to sync.Once ensuring single initialization.
func Get() *zap.SugaredLogger {
	// If Init() was already called, once.Do is a no-op and we just return logger.
	// If Init() was NOT called, this triggers initialization with default settings.
	once.Do(func() {
		initLogger(os.Getenv("LOG_LEVEL") == "DEBUG")
	})
	return logger
}

// RedactPII replaces sensitive matches with "[REDACTED]".
func RedactPII(s string) string {
	s = piiNumbersRegex.ReplaceAllString(s, "[REDACTED_ID]")
	s = piiEmailRegex.ReplaceAllString(s, "[REDACTED_EMAIL]")
	return s
}

// Wrapper functions for convenience and redaction

func Info(args ...interface{}) {
	Get().Info(args...)
}

func Infof(template string, args ...interface{}) {
	Get().Infof(template, redactArgs(args)...)
}

func Error(args ...interface{}) {
	Get().Error(args...)
}

func Errorf(template string, args ...interface{}) {
	Get().Errorf(template, redactArgs(args)...)
}

func Debug(args ...interface{}) {
	Get().Debug(args...)
}

func Debugf(template string, args ...interface{}) {
	Get().Debugf(template, redactArgs(args)...)
}

func Warn(args ...interface{}) {
	Get().Warn(args...)
}

func Warnf(template string, args ...interface{}) {
	Get().Warnf(template, redactArgs(args)...)
}

func Fatal(args ...interface{}) {
	Get().Fatal(args...)
}

func Fatalf(template string, args ...interface{}) {
	Get().Fatalf(template, redactArgs(args)...)
}

// redactArgs applies RedactPII to string arguments
func redactArgs(args []interface{}) []interface{} {
	redacted := make([]interface{}, len(args))
	for i, arg := range args {
		if s, ok := arg.(string); ok {
			redacted[i] = RedactPII(s)
		} else {
			redacted[i] = arg
		}
	}
	return redacted
}
