package application

import "testing"

func TestDocumentRevisionKey_IsTenantScoped(t *testing.T) {
	got := documentRevisionKey("tenant-1", "doc-7", "abc123")
	want := "tenants/tenant-1/documents/doc-7/revisions/abc123.docx"
	if got != want {
		t.Fatalf("documentRevisionKey = %q, want %q", got, want)
	}
}
