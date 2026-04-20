package transcription

import (
	"context"
	"io"
	"strings"
	"testing"
)

func TestNewGroqService(t *testing.T) {
	svc := NewGroqAdapter("fake-key")
	if svc == nil {
		t.Fatal("Expected non-nil service")
	}
}

func TestTranscribeVoice_NoKey(t *testing.T) {
	svc := NewGroqAdapter("")
	reader := strings.NewReader("fake audio data")
	
	_, err := svc.Transcribe(context.Background(), reader, "test.ogg")
	if err == nil || !strings.Contains(err.Error(), "API key not configured") {
		t.Errorf("Expected API key missing error, got: %v", err)
	}
}

func TestTranscribeVoice_NilReader(t *testing.T) {
	svc := NewGroqAdapter("fake-key")
	
	_, err := svc.Transcribe(context.Background(), nil, "test.ogg")
	if err == nil || !strings.Contains(err.Error(), "nil reader") {
		t.Errorf("Expected nil reader error, got: %v", err)
	}
}

// Full transcription test requires mocking the HTTP request, which can be done by
// refactoring groq.go to allow overriding the HTTP client or URL. For now, we test the
// guard clauses to improve basic coverage.

func TestSplitSummary_EdgeCases(t *testing.T) {
	// Not part of groq.go, but can be tested in domain/models_test.go
}

type errReader struct{}

func (errReader) Read(p []byte) (n int, err error) {
	return 0, io.ErrUnexpectedEOF
}

func TestTranscribeVoice_ReaderError(t *testing.T) {
	svc := NewGroqAdapter("fake-key")
	
	_, err := svc.Transcribe(context.Background(), errReader{}, "test.ogg")
	if err == nil {
		t.Errorf("Expected error reading from faulty reader, got nil")
	}
}
