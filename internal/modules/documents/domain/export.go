package domain

import "errors"

// Export is an immutable row in document_exports representing one cached PDF.
type Export struct {
	ID            string
	TenantID      string
	DocumentID    string
	RevisionID    string
	CompositeHash []byte // 32 bytes
	StorageKey    string
	SizeBytes     int64
	PaperSize     string
	Landscape     bool
	DocgenV2Ver   string
}

// ExportResult carries the export row and whether it was retrieved from cache.
type ExportResult struct {
	Export *Export
	Cached bool
}

// Export error sentinels: ErrExportGotenbergFailed signals the Gotenberg PDF
// conversion call failed, ErrExportDocxMissing signals the source docx body
// could not be located, and ErrInvalidExport signals the requested export
// parameters (e.g. paper size) are invalid.
var (
	ErrExportGotenbergFailed = errors.New("gotenberg_conversion_failed")
	ErrExportDocxMissing     = errors.New("docx_missing")
	ErrInvalidExport         = errors.New("invalid_export")
)

var allowedPaperSizes = map[string]struct{}{
	"A4":      {},
	"Letter":  {},
	"Legal":   {},
	"A3":      {},
	"Tabloid": {},
}

// NewExport constructs an Export after validating that paperSize is one of
// the allowed sizes, returning ErrInvalidExport otherwise.
func NewExport(tenantID, documentID, revisionID string, compositeHash []byte, storageKey string, sizeBytes int64, paperSize string, landscape bool, docgenV2Ver string) (*Export, error) {
	if _, ok := allowedPaperSizes[paperSize]; !ok {
		return nil, ErrInvalidExport
	}
	return &Export{
		TenantID:      tenantID,
		DocumentID:    documentID,
		RevisionID:    revisionID,
		CompositeHash: compositeHash,
		StorageKey:    storageKey,
		SizeBytes:     sizeBytes,
		PaperSize:     paperSize,
		Landscape:     landscape,
		DocgenV2Ver:   docgenV2Ver,
	}, nil
}
