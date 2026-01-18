package pdf

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"time"

	"github.com/kfilin/massage-bot/internal/ports"
)

type adapter struct {
	url         string
	internalKey string
	client      *http.Client
}

// NewAdapter creates a new PDF generator adapter using bentopdf.
func NewAdapter(url, internalKey string) ports.PDFGenerator {
	return &adapter{
		url:         url,
		internalKey: internalKey,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

type pdfRequest struct {
	HTML    string                 `json:"html"`
	Options map[string]interface{} `json:"options"`
}

func (a *adapter) GeneratePDF(ctx context.Context, htmlContent string) ([]byte, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Stirling-PDF expects the HTML as a file part
	part, err := writer.CreateFormFile("fileInput", "medical_record.html")
	if err != nil {
		return nil, fmt.Errorf("failed to create form file: %w", err)
	}
	if _, err := part.Write([]byte(htmlContent)); err != nil {
		return nil, fmt.Errorf("failed to write html to form: %w", err)
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("failed to close multipart writer: %w", err)
	}

	apiURL := fmt.Sprintf("%s/api/v1/convert/html/pdf", a.url)
	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create pdf request: %w", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	if a.internalKey != "" {
		req.Header.Set("X-API-KEY", a.internalKey)
	}

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call Stirling-PDF: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Stirling-PDF returned error (status %d): %s", resp.StatusCode, string(respBody))
	}

	return io.ReadAll(resp.Body)
}
