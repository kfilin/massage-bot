package transcription

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
	"time"

	"github.com/kfilin/massage-bot/internal/logging"
	"github.com/kfilin/massage-bot/internal/monitoring"
)

// localAdapter sends audio to a self-hosted OpenAI-compatible Whisper server
// (faster-whisper-server or similar) running on the same Docker network.
type localAdapter struct {
	baseURL string
	client  *http.Client
}

// DefaultLocalURL is the default endpoint for the local Whisper server.
// Points to the faster-whisper-server container on the caddy-test-net bridge.
const DefaultLocalURL = "http://whisper:8000/v1/audio/transcriptions"

// NewLocalAdapter creates a transcription service pointing to a local
// self-hosted Whisper instance (OpenAI-compatible API).
func NewLocalAdapter(baseURL string) *localAdapter {
	if baseURL == "" {
		baseURL = DefaultLocalURL
	}
	return &localAdapter{
		baseURL: baseURL,
		client: &http.Client{
			Timeout: 120 * time.Second, // self-hosted can be slower than API
		},
	}
}

type localResponse struct {
	Text string `json:"text"`
}

// Transcribe implements ports.TranscriptionService.
// Sends the audio as a multipart/form-data POST to the local Whisper server.
// Mirrors the approach used in agentic-lab-2.0's connect_handler.go.
func (a *localAdapter) Transcribe(ctx context.Context, audio io.Reader, filename string) (string, error) {
	if audio == nil {
		return "", fmt.Errorf("nil reader")
	}

	// Read all audio data into memory first (same pattern as agentic-lab).
	// This avoids issues with streaming readers during multipart construction.
	audioData, err := io.ReadAll(audio)
	if err != nil {
		return "", fmt.Errorf("failed to read audio data: %w", err)
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add file part (matches agentic-lab: CreateFormFile + Write)
	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return "", fmt.Errorf("failed to create form file: %w", err)
	}
	if _, err := part.Write(audioData); err != nil {
		return "", fmt.Errorf("failed to write audio to form: %w", err)
	}

	// Add model — use the actual model name (matching agentic-lab)
	if err := writer.WriteField("model", "Systran/faster-whisper-small"); err != nil {
		return "", fmt.Errorf("failed to write model field: %w", err)
	}

	// Language is optional — the server defaults to auto-detect.
	// Don't force "ru" to match the server's default_language=None config.
	if err := writer.Close(); err != nil {
		return "", fmt.Errorf("failed to close multipart writer: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", a.baseURL, body)
	if err != nil {
		return "", fmt.Errorf("failed to create local whisper request: %w", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	start := time.Now()
	resp, err := a.client.Do(req)
	duration := time.Since(start).Seconds()

	status := "success"
	if err != nil {
		status = "error"
	} else if resp.StatusCode != http.StatusOK {
		status = "error_api"
	}

	monitoring.ApiRequestsTotal.WithLabelValues("whisper", "transcribe", status).Inc()
	monitoring.ApiLatency.WithLabelValues("whisper", "transcribe").Observe(duration)

	if err != nil {
		return "", fmt.Errorf("failed to call local whisper: %w", err)
	}
	defer resp.Body.Close()

	// Read full response body (matches agentic-lab pattern)
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read whisper response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		errMsg := fmt.Sprintf("local whisper returned error (status %d): %s", resp.StatusCode, string(respBody))
		logging.Error(errMsg)
		return "", fmt.Errorf("%s", errMsg)
	}

	var result struct {
		Text     string `json:"text"`
		Language string `json:"language,omitempty"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("failed to decode whisper response: %w", err)
	}

	text := strings.TrimSpace(result.Text)

	// Filter common hallucinations (especially "You" or "Thank you" on silence)
	lowerText := strings.ToLower(text)
	if len(text) < 20 { // Only filter short texts
		if strings.Contains(lowerText, "you") ||
			strings.Contains(lowerText, "thank you") ||
			strings.Contains(lowerText, "subscribe") ||
			strings.Contains(lowerText, "watching") ||
			strings.Contains(lowerText, "продолжение следует") {
			return "", nil // Return empty string to indicate silence/noise
		}
	}

	return text, nil
}