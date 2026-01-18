package ports

import (
	"context"
	"io"
)

// TranscriptionService defines the interface for transcribing audio to text.
type TranscriptionService interface {
	Transcribe(ctx context.Context, audio io.Reader, filename string) (string, error)
}
