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
		{name: "v1 documents list now unguarded", method: http.MethodGet, path: "/api/v1/documents", wantCap: "", wantGuard: false},
		{name: "workflow transition", method: http.MethodPost, path: "/api/v1/workflow/documents/doc-1/transitions", wantCap: iamdomain.CapDocumentSubmit, wantGuard: true},
		{name: "iam users list", method: http.MethodGet, path: "/api/v1/iam/users", wantCap: iamdomain.CapUserManage, wantGuard: true},
		{name: "iam roles update", method: http.MethodPut, path: "/api/v1/iam/users/u-1/roles", wantCap: iamdomain.CapUserManage, wantGuard: true},
		{name: "template list", method: http.MethodGet, path: "/api/v1/templates", wantCap: iamdomain.CapTemplateView, wantGuard: true},
		{name: "template create", method: http.MethodPost, path: "/api/v1/templates", wantCap: iamdomain.CapTemplateView, wantGuard: true},
		{name: "v2 templates list", method: http.MethodGet, path: "/api/v2/templates", wantCap: iamdomain.CapTemplateView, wantGuard: true},
		{name: "v2 templates create", method: http.MethodPost, path: "/api/v2/templates", wantCap: iamdomain.CapTemplateEdit, wantGuard: true},
		{name: "v2 templates version draft", method: http.MethodPut, path: "/api/v2/templates/t1/versions/1/draft", wantCap: iamdomain.CapTemplateEdit, wantGuard: true},
		{name: "v2 templates publish", method: http.MethodPost, path: "/api/v2/templates/t1/versions/1/publish", wantCap: iamdomain.CapTemplatePublish, wantGuard: true},
		{name: "v2 docx-upload-url", method: http.MethodPost, path: "/api/v2/templates/t1/versions/1/docx-upload-url", wantCap: iamdomain.CapTemplateEdit, wantGuard: true},
		{name: "v2 schema-upload-url", method: http.MethodPost, path: "/api/v2/templates/t1/versions/1/schema-upload-url", wantCap: iamdomain.CapTemplateEdit, wantGuard: true},
		{name: "v2 signed download", method: http.MethodGet, path: "/api/v2/signed", wantCap: iamdomain.CapTemplateView, wantGuard: true},
		{name: "v2 doc submit", method: http.MethodPost, path: "/api/v2/documents/d1/submit", wantCap: iamdomain.CapDocumentSubmit, wantGuard: true},
		{name: "v2 doc signoff", method: http.MethodPost, path: "/api/v2/documents/d1/signoff", wantCap: iamdomain.CapDocumentSignoff, wantGuard: true},
		{name: "v2 taxonomy families list", method: http.MethodGet, path: "/api/v2/taxonomy/families", wantCap: iamdomain.CapDocumentView, wantGuard: true},
		{name: "v2 taxonomy families create", method: http.MethodPost, path: "/api/v2/taxonomy/families", wantCap: iamdomain.CapTaxonomyManage, wantGuard: true},
		{name: "v2 controlled-documents create", method: http.MethodPost, path: "/api/v2/controlled-documents", wantCap: iamdomain.CapRegistryCreate, wantGuard: true},
		{name: "v2 controlled-documents revisions create", method: http.MethodPost, path: "/api/v2/controlled-documents/cd-1/revisions", wantCap: iamdomain.CapDocEdit, wantGuard: true},
		{name: "v2 controlled-documents preview-code", method: http.MethodGet, path: "/api/v2/controlled-documents/preview-code", wantCap: iamdomain.CapDocumentView, wantGuard: true},
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
