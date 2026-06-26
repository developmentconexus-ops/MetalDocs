package application

import "fmt"

// documentRevisionKey builds the canonical tenant-scoped object key for a document
// revision's docx. Mirrors the layout the old DocumentPresigner.PresignRevisionPUT
// constructed internally.
func documentRevisionKey(tenantID, documentID, contentHash string) string {
	return fmt.Sprintf("tenants/%s/documents/%s/revisions/%s.docx", tenantID, documentID, contentHash)
}
