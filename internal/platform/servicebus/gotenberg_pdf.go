package servicebus

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
)

// PaperSize selects the output page geometry for PDF rendering.
type PaperSize string

const (
	PaperSizeA4     PaperSize = "A4"
	PaperSizeLetter PaperSize = "Letter"
)

// PDFRenderOpts carries optional page-geometry overrides. When nil/empty the
// PDF uses the docx's own page setup.
type PDFRenderOpts struct {
	PaperSize PaperSize `json:"paper_size,omitempty"`
	Landscape bool      `json:"landscape,omitempty"`
}

// ConvertPDFRequest is the input contract for PDF generation: read the docx at
// DocxKey from object storage and write the rendered PDF to OutputKey.
type ConvertPDFRequest struct {
	DocxKey    string         `json:"docx_key"`
	OutputKey  string         `json:"output_key"`
	RenderOpts *PDFRenderOpts `json:"render_opts,omitempty"`
}

// ConvertPDFResult is the outcome of a PDF generation.
type ConvertPDFResult struct {
	OutputKey   string `json:"output_key"`
	ContentHash string `json:"content_hash"`
	SizeBytes   int64  `json:"size_bytes"`
}

// docxToPDFConverter is the subset of the Gotenberg client this adapter needs.
// Satisfied by *render/gotenberg.Client. paperSize is "" / "A4" / "Letter".
type docxToPDFConverter interface {
	ConvertDocxToPDFWithOptions(ctx context.Context, docx []byte, paperSize string, landscape bool) ([]byte, error)
}

// pdfObjectStore is the minimal object-store surface the adapter needs.
// Satisfied by *storage/minio.Store.
type pdfObjectStore interface {
	Open(ctx context.Context, key string) (io.ReadCloser, error)
	Save(ctx context.Context, key string, content []byte) error
}

// GotenbergPDFClient implements the PDF-conversion contract (ConvertPDF) by
// reading the source docx from object storage, converting it to PDF directly
// via Gotenberg (LibreOffice route), and writing the PDF back to storage.
//
// It replaces the docx-renderer round-trip for PDF generation: the only live
// docgen job was a Gotenberg proxy with no eigenpal involvement, so the Go
// worker now calls Gotenberg directly. See wiki/decisions for the v1 decision.
type GotenbergPDFClient struct {
	store     pdfObjectStore
	converter docxToPDFConverter
}

func NewGotenbergPDFClient(store pdfObjectStore, converter docxToPDFConverter) *GotenbergPDFClient {
	return &GotenbergPDFClient{store: store, converter: converter}
}

func (c *GotenbergPDFClient) ConvertPDF(ctx context.Context, req ConvertPDFRequest) (ConvertPDFResult, error) {
	var zero ConvertPDFResult

	rc, err := c.store.Open(ctx, req.DocxKey)
	if err != nil {
		return zero, fmt.Errorf("gotenberg pdf: open docx %q: %w", req.DocxKey, err)
	}
	defer rc.Close()

	docx, err := io.ReadAll(rc)
	if err != nil {
		return zero, fmt.Errorf("gotenberg pdf: read docx %q: %w", req.DocxKey, err)
	}

	paperSize := ""
	landscape := false
	if req.RenderOpts != nil {
		paperSize = string(req.RenderOpts.PaperSize)
		landscape = req.RenderOpts.Landscape
	}

	pdf, err := c.converter.ConvertDocxToPDFWithOptions(ctx, docx, paperSize, landscape)
	if err != nil {
		return zero, fmt.Errorf("gotenberg pdf: convert %q: %w", req.DocxKey, err)
	}

	sum := sha256.Sum256(pdf)
	hash := hex.EncodeToString(sum[:])

	if err := c.store.Save(ctx, req.OutputKey, pdf); err != nil {
		return zero, fmt.Errorf("gotenberg pdf: save %q: %w", req.OutputKey, err)
	}

	return ConvertPDFResult{
		OutputKey:   req.OutputKey,
		ContentHash: hash,
		SizeBytes:   int64(len(pdf)),
	}, nil
}
