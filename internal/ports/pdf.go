package ports

import "context"

// PDFGenerator defines the interface for generating PDF documents from HTML.
type PDFGenerator interface {
	GeneratePDF(ctx context.Context, html string) ([]byte, error)
}
