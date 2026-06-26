package application

import "testing"

func TestTemplateDocxKey_IsTenantScoped(t *testing.T) {
	got := templateDocxKey("tenant-1", "tpl-9", 3)
	want := "tenants/tenant-1/templates/tpl-9/versions/3.docx"
	if got != want {
		t.Fatalf("templateDocxKey = %q, want %q", got, want)
	}
}
