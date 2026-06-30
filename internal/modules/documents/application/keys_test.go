package application

import (
	"encoding/hex"
	"testing"
)

func TestDocumentRevisionKey_IsTenantScoped(t *testing.T) {
	got := documentRevisionKey("tenant-1", "doc-7", "abc123")
	want := "tenants/tenant-1/documents/doc-7/revisions/abc123.docx"
	if got != want {
		t.Fatalf("documentRevisionKey = %q, want %q", got, want)
	}
}

// F-O5: documentExportKey must produce the identical key string that the inline
// fmt.Sprintf in export_service.go:77 produced before the refactor.
func TestDocumentExportKey_MatchesInlineFormat(t *testing.T) {
	compositeHash, _ := hex.DecodeString("deadbeef01020304")
	got := documentExportKey("tenant-1", "doc-7", compositeHash)
	want := "tenants/tenant-1/documents/doc-7/exports/" + hex.EncodeToString(compositeHash) + ".pdf"
	if got != want {
		t.Fatalf("documentExportKey = %q, want %q", got, want)
	}
}
