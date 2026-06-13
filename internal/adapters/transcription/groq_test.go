package transcription

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
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

func TestTranscribe_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-key" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(groqResponse{Text: "Привет, мир"})
	}))
	defer server.Close()

	svc := &groqAdapter{
		apiKey:  "test-key",
		client:  server.Client(),
		model:   "whisper-large-v3",
		baseURL: server.URL,
	}

	text, err := svc.Transcribe(context.Background(), strings.NewReader("audio data"), "test.ogg")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if text != "Привет, мир" {
		t.Errorf("Expected 'Привет, мир', got '%s'", text)
	}
}

func TestTranscribe_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal error"))
	}))
	defer server.Close()

	svc := &groqAdapter{
		apiKey:  "test-key",
		client:  server.Client(),
		model:   "whisper-large-v3",
		baseURL: server.URL,
	}

	_, err := svc.Transcribe(context.Background(), strings.NewReader("audio data"), "test.ogg")
	if err == nil || !strings.Contains(err.Error(), "status 500") {
		t.Errorf("Expected API error with status 500, got: %v", err)
	}
}

func TestTranscribe_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("not json"))
	}))
	defer server.Close()

	svc := &groqAdapter{
		apiKey:  "test-key",
		client:  server.Client(),
		model:   "whisper-large-v3",
		baseURL: server.URL,
	}

	_, err := svc.Transcribe(context.Background(), strings.NewReader("audio data"), "test.ogg")
	if err == nil || !strings.Contains(err.Error(), "decode") {
		t.Errorf("Expected decode error, got: %v", err)
	}
}

func TestTranscribe_HallucinationFilter(t *testing.T) {
	tests := []struct {
		name     string
		apiText  string
		expected string
	}{
		{"Short 'You'", "You", ""},
		{"Short 'Thank you'", "Thank you", ""},
		{"Short 'Subscribe'", "Subscribe", ""},
		{"Long text kept", "This is a long transcription that contains the word you but is over twenty characters", "This is a long transcription that contains the word you but is over twenty characters"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				json.NewEncoder(w).Encode(groqResponse{Text: tt.apiText})
			}))
			defer server.Close()

			svc := &groqAdapter{
				apiKey:  "test-key",
				client:  server.Client(),
				model:   "whisper-large-v3",
				baseURL: server.URL,
			}

			text, err := svc.Transcribe(context.Background(), strings.NewReader("audio"), "test.ogg")
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if text != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, text)
			}
		})
	}
}

func TestTranscribe_ContextCancelled(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	svc := &groqAdapter{
		apiKey:  "test-key",
		client:  server.Client(),
		model:   "whisper-large-v3",
		baseURL: server.URL,
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := svc.Transcribe(ctx, strings.NewReader("audio"), "test.ogg")
	if err == nil {
		t.Error("Expected error from cancelled context")
	}
}
