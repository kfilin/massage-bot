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

func TestNewLocalAdapter(t *testing.T) {
	svc := NewLocalAdapter("http://whisper:8000/v1")
	if svc == nil {
		t.Fatal("Expected non-nil service")
	}
	if svc.baseURL != "http://whisper:8000/v1" {
		t.Errorf("Expected baseURL http://whisper:8000/v1, got %s", svc.baseURL)
	}
}

func TestNewLocalAdapter_DefaultURL(t *testing.T) {
	svc := NewLocalAdapter("")
	if svc.baseURL != DefaultLocalURL {
		t.Errorf("Expected default URL %s, got %s", DefaultLocalURL, svc.baseURL)
	}
}

func TestTranscribeLocal_NilReader(t *testing.T) {
	svc := NewLocalAdapter("fake-url")

	_, err := svc.Transcribe(context.Background(), nil, "test.ogg")
	if err == nil || !strings.Contains(err.Error(), "nil reader") {
		t.Errorf("Expected nil reader error, got: %v", err)
	}
}

type errReader struct{}

func (errReader) Read(p []byte) (n int, err error) {
	return 0, io.ErrUnexpectedEOF
}

func TestTranscribeLocal_ReaderError(t *testing.T) {
	svc := NewLocalAdapter("fake-url")

	_, err := svc.Transcribe(context.Background(), errReader{}, "test.ogg")
	if err == nil {
		t.Errorf("Expected error reading from faulty reader, got nil")
	}
}

func TestTranscribeLocal_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check Content-Type is multipart/form-data
		ct := r.Header.Get("Content-Type")
		if !strings.HasPrefix(ct, "multipart/form-data") {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"text": "Привет, мир"})
	}))
	defer server.Close()

	svc := NewLocalAdapter(server.URL)

	text, err := svc.Transcribe(context.Background(), strings.NewReader("audio data"), "test.ogg")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if text != "Привет, мир" {
		t.Errorf("Expected 'Привет, мир', got '%s'", text)
	}
}

func TestTranscribeLocal_ContextCancelled(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	svc := NewLocalAdapter(server.URL)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := svc.Transcribe(ctx, strings.NewReader("audio"), "test.ogg")
	if err == nil {
		t.Error("Expected error from cancelled context")
	}
}

func TestTranscribeLocal_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("model not loaded"))
	}))
	defer server.Close()

	svc := NewLocalAdapter(server.URL)

	_, err := svc.Transcribe(context.Background(), strings.NewReader("audio data"), "test.ogg")
	if err == nil || !strings.Contains(err.Error(), "500") {
		t.Errorf("Expected HTTP 500 error, got: %v", err)
	}
}

func TestTranscribeLocal_HallucinationFilter(t *testing.T) {
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
				_ = json.NewEncoder(w).Encode(map[string]string{"text": tt.apiText})
			}))
			defer server.Close()

			svc := NewLocalAdapter(server.URL)

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