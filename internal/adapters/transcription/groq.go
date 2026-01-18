package transcription

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"time"

	"github.com/kfilin/massage-bot/internal/ports"
)

type groqAdapter struct {
	apiKey string
	client *http.Client
	model  string
}

// NewGroqAdapter creates a new transcription service using Groq's Whisper API.
func NewGroqAdapter(apiKey string) ports.TranscriptionService {
	return &groqAdapter{
		apiKey: apiKey,
		client: &http.Client{
			Timeout: 60 * time.Second,
		},
		model: "whisper-large-v3",
	}
}

type groqResponse struct {
	Text string `json:"text"`
}

func (a *groqAdapter) Transcribe(ctx context.Context, audio io.Reader, filename string) (string, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add file part
	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return "", fmt.Errorf("failed to create form file: %w", err)
	}
	if _, err := io.Copy(part, audio); err != nil {
		return "", fmt.Errorf("failed to copy audio to form: %w", err)
	}

	// Add model part
	if err := writer.WriteField("model", a.model); err != nil {
		return "", fmt.Errorf("failed to write model field: %w", err)
	}

	if err := writer.Close(); err != nil {
		return "", fmt.Errorf("failed to close multipart writer: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.groq.com/openai/v1/audio/transcriptions", body)
	if err != nil {
		return "", fmt.Errorf("failed to create groq request: %w", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", a.apiKey))

	resp, err := a.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to call groq api: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("groq api returned error (status %d): %s", resp.StatusCode, string(respBody))
	}

	var groqResp groqResponse
	if err := json.NewDecoder(resp.Body).Decode(&groqResp); err != nil {
		return "", fmt.Errorf("failed to decode groq response: %w", err)
	}

	return groqResp.Text, nil
}
