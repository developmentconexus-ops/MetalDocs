package main

import (
	"net/http"
	"testing"

	iamdomain "metaldocs/internal/modules/iam/domain"
)

func TestPermissionResolver(t *testing.T) {
	t.Parallel()

	resolver := newPermissionResolver()

	testCases := []struct {
		name      string
		method    string
		path      string
		wantCap   iamdomain.Capability
		wantGuard bool
	}{
		{name: "health live unguarded", method: http.MethodGet, path: "/api/v1/health/live", wantCap: "", wantGuard: false},
		{name: "auth login unguarded", method: http.MethodPost, path: "/api/v1/auth/login", wantCap: "", wantGuard: false},
		{name: "feature flags unguarded", method: http.MethodGet, path: "/api/v1/feature-flags", wantCap: "", wantGuard: false},
		{name: "unknown endpoint unguarded", method: http.MethodGet, path: "/api/v1/unknown", wantCap: "", wantGuard: false},
		{name: "v1 documents list guarded", method: http.MethodGet, path: "/api/v1/documents", wantCap: iamdomain.CapDocumentView, wantGuard: true},
		{name: "workflow transition", method: http.MethodPost, path: "/api/v1/workflow/documents/doc-1/transitions", wantCap: iamdomain.CapDocumentSubmit, wantGuard: true},
		{name: "iam users list", method: http.MethodGet, path: "/api/v1/iam/users", wantCap: iamdomain.CapUserManage, wantGuard: true},
		{name: "iam roles update", method: http.MethodPut, path: "/api/v1/iam/users/u-1/roles", wantCap: iamdomain.CapUserManage, wantGuard: true},
		{name: "template list", method: http.MethodGet, path: "/api/v1/templates", wantCap: iamdomain.CapTemplateView, wantGuard: true},
		{name: "template create", method: http.MethodPost, path: "/api/v1/templates", wantCap: iamdomain.CapTemplateCreate, wantGuard: true},
		{name: "v1 templates list", method: http.MethodGet, path: "/api/v1/templates", wantCap: iamdomain.CapTemplateView, wantGuard: true},
		{name: "v1 templates create", method: http.MethodPost, path: "/api/v1/templates", wantCap: iamdomain.CapTemplateCreate, wantGuard: true},
		{name: "v1 templates create version", method: http.MethodPost, path: "/api/v1/templates/t1/versions", wantCap: iamdomain.CapTemplateCreate, wantGuard: true},
		{name: "v1 templates version draft", method: http.MethodPut, path: "/api/v1/templates/t1/versions/1/draft", wantCap: iamdomain.CapTemplateEdit, wantGuard: true},
		{name: "v1 templates version schema", method: http.MethodPut, path: "/api/v1/templates/t1/versions/1/schema", wantCap: iamdomain.CapTemplateEdit, wantGuard: true},
		{name: "v1 templates publish", method: http.MethodPost, path: "/api/v1/templates/t1/versions/1/publish", wantCap: iamdomain.CapTemplatePublish, wantGuard: true},
		{name: "v1 docx-upload-url", method: http.MethodPost, path: "/api/v1/templates/t1/versions/1/docx-upload-url", wantCap: iamdomain.CapTemplateEdit, wantGuard: true},
		{name: "v1 schema-upload-url", method: http.MethodPost, path: "/api/v1/templates/t1/versions/1/schema-upload-url", wantCap: iamdomain.CapTemplateEdit, wantGuard: true},
		{name: "v1 autosave presign", method: http.MethodPost, path: "/api/v1/templates/t1/versions/1/autosave/presign", wantCap: iamdomain.CapTemplateEdit, wantGuard: true},
		{name: "v1 autosave commit", method: http.MethodPost, path: "/api/v1/templates/t1/versions/1/autosave/commit", wantCap: iamdomain.CapTemplateEdit, wantGuard: true},
		{name: "v1 templates approval config", method: http.MethodPut, path: "/api/v1/templates/t1/approval-config", wantCap: iamdomain.CapTemplateEdit, wantGuard: true},
		{name: "v1 signed download", method: http.MethodGet, path: "/api/v1/signed", wantCap: iamdomain.CapTemplateView, wantGuard: true},
		{name: "v1 template review", method: http.MethodPost, path: "/api/v1/templates/t1/versions/1/review", wantCap: iamdomain.CapTemplateReview, wantGuard: true},
		{name: "v1 template archive", method: http.MethodPost, path: "/api/v1/templates/t1/archive", wantCap: iamdomain.Capability("template.archive"), wantGuard: true},
		{name: "v1 doc submit", method: http.MethodPost, path: "/api/v1/documents/d1/submit", wantCap: iamdomain.CapDocumentSubmit, wantGuard: true},
		{name: "v1 doc signoff", method: http.MethodPost, path: "/api/v1/documents/d1/signoff", wantCap: iamdomain.CapDocumentSignoff, wantGuard: true},
		{name: "v1 taxonomy families list", method: http.MethodGet, path: "/api/v1/taxonomy/families", wantCap: iamdomain.CapDocumentView, wantGuard: true},
		{name: "v1 taxonomy families create", method: http.MethodPost, path: "/api/v1/taxonomy/families", wantCap: iamdomain.CapTaxonomyManage, wantGuard: true},
		{name: "v1 controlled-documents create", method: http.MethodPost, path: "/api/v1/controlled-documents", wantCap: iamdomain.CapRegistryCreate, wantGuard: true},
		{name: "v1 controlled-documents revisions create", method: http.MethodPost, path: "/api/v1/controlled-documents/cd-1/revisions", wantCap: iamdomain.CapDocumentEdit, wantGuard: true},
		{name: "v1 controlled-documents preview-code", method: http.MethodGet, path: "/api/v1/controlled-documents/preview-code", wantCap: iamdomain.CapDocumentView, wantGuard: true},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			gotCap, gotGuard := resolver(tc.method, tc.path)
			if gotGuard != tc.wantGuard {
				t.Fatalf("guard mismatch: got %v want %v", gotGuard, tc.wantGuard)
			}
			if gotCap != tc.wantCap {
				t.Fatalf("capability mismatch: got %q want %q", gotCap, tc.wantCap)
			}
		})
	}
}
func TestPermissionResolver_TaxonomyFamiliesPATCH_RequiresTaxonomyManage(t *testing.T) {
	resolver := newPermissionResolver()
	cap, ok := resolver(http.MethodPatch, "/api/v1/taxonomy/families/PROC")
	if !ok {
		t.Fatal("PATCH /taxonomy/families/{code}: resolver returned ok=false, want true")
	}
	if cap != iamdomain.CapTaxonomyManage {
		t.Fatalf("PATCH /taxonomy/families: cap = %v, want CapTaxonomyManage", cap)
	}
}
func TestPermissionResolver_TaxonomyAreasPATCH_RequiresTaxonomyManage(t *testing.T) {
	resolver := newPermissionResolver()
	cap, ok := resolver(http.MethodPatch, "/api/v1/taxonomy/areas/QA")
	if !ok {
		t.Fatal("PATCH /taxonomy/areas/{code}: resolver returned ok=false, want true")
	}
	if cap != iamdomain.CapTaxonomyManage {
		t.Fatalf("PATCH /taxonomy/areas: cap = %v, want CapTaxonomyManage", cap)
	}
}

func TestPermissionResolver_RegistryObsolete_RequiresRegistryObsoleteCap(t *testing.T) {
	r := newPermissionResolver()
	cap, ok := r(http.MethodPut, "/api/v1/controlled-documents/some-id/obsolete")
	if !ok {
		t.Fatal("PUT .../obsolete: resolver returned ok=false")
	}
	if cap != iamdomain.CapRegistryObsolete {
		t.Fatalf("PUT .../obsolete: cap = %v, want CapRegistryObsolete", cap)
	}
}

func TestPermissionResolver_RegistrySupersede_RequiresRegistrySupersedeCAP(t *testing.T) {
	r := newPermissionResolver()
	cap, ok := r(http.MethodPut, "/api/v1/controlled-documents/some-id/supersede")
	if !ok {
		t.Fatal("PUT .../supersede: resolver returned ok=false")
	}
	if cap != iamdomain.CapRegistrySupersede {
		t.Fatalf("PUT .../supersede: cap = %v, want CapRegistrySupersede", cap)
	}
}
