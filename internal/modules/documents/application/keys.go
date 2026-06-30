package application

import (
	"encoding/hex"
	"fmt"
)

// documentRevisionKey builds the canonical tenant-scoped object key for a document
// revision's docx. Mirrors the layout the old DocumentPresigner.PresignRevisionPUT
// constructed internally.
func documentRevisionKey(tenantID, documentID, contentHash string) string {
	return fmt.Sprintf("tenants/%s/documents/%s/revisions/%s.docx", tenantID, documentID, contentHash)
}

// documentExportKey builds the canonical tenant-scoped object key for a document
// PDF export. The key format is identical to the inline fmt.Sprintf in
// export_service.go:77 — this function is the single authoritative definition.
func documentExportKey(tenantID, documentID string, compositeHash []byte) string {
	return fmt.Sprintf("tenants/%s/documents/%s/exports/%s.pdf", tenantID, documentID, hex.EncodeToString(compositeHash))
}
