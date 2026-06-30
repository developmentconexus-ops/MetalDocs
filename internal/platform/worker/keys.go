package worker

import "fmt"

// workerPDFKey builds the canonical tenant-scoped object key for a rendered PDF
// output. This is the single authoritative definition for the layout that was
// previously an inline fmt.Sprintf in pdf_job_runner.go. Mirrors the keys.go
// packaging convention used by internal/modules/documents/application/keys.go.
func workerPDFKey(tenantID, revisionID string) string {
	return fmt.Sprintf("tenants/%s/revisions/%s/final.pdf", tenantID, revisionID)
}
