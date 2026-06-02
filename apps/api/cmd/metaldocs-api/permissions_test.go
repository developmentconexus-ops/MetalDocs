package main

import (
	"net/http"
	"strings"
	"testing"

	iamdelivery "metaldocs/internal/modules/iam/delivery/http"
	iamdomain "metaldocs/internal/modules/iam/domain"
)

func TestPermissionResolver(t *testing.T) {
	t.Parallel()

	resolver := newPermissionResolver()

	testCases := []struct {
		name           string
		method         string
		path           string
		wantCap        iamdomain.Capability
		wantVisibility iamdelivery.Visibility
	}{
		// --- Public routes (no session required) ---
		{name: "health live public", method: http.MethodGet, path: "/api/v1/health/live", wantCap: "", wantVisibility: iamdelivery.VisibilityPublic},
		{name: "health ready public", method: http.MethodGet, path: "/api/v1/health/ready", wantCap: "", wantVisibility: iamdelivery.VisibilityPublic},
		{name: "healthz public", method: http.MethodGet, path: "/healthz", wantCap: "", wantVisibility: iamdelivery.VisibilityPublic},
		{name: "auth login public", method: http.MethodPost, path: "/api/v1/auth/login", wantCap: "", wantVisibility: iamdelivery.VisibilityPublic},
		{name: "auth refresh public", method: http.MethodPost, path: "/api/v1/auth/refresh", wantCap: "", wantVisibility: iamdelivery.VisibilityPublic},
		{name: "feature flags public", method: http.MethodGet, path: "/api/v1/feature-flags", wantCap: "", wantVisibility: iamdelivery.VisibilityPublic},

		// --- Session-required (authenticated, no capability) ---
		{name: "auth me session required", method: http.MethodGet, path: "/api/v1/auth/me", wantCap: "", wantVisibility: iamdelivery.VisibilitySessionRequired},
		{name: "auth change password session required", method: http.MethodPost, path: "/api/v1/auth/change-password", wantCap: "", wantVisibility: iamdelivery.VisibilitySessionRequired},
		{name: "auth logout session required", method: http.MethodPost, path: "/api/v1/auth/logout", wantCap: "", wantVisibility: iamdelivery.VisibilitySessionRequired},

		// --- C2 fail-closed regressions: unmatched routes MUST default to SessionRequired, never Public ---
		{name: "unknown endpoint session required", method: http.MethodGet, path: "/api/v1/unknown", wantCap: "", wantVisibility: iamdelivery.VisibilitySessionRequired},
		{name: "unknown documents subpath session required", method: http.MethodDelete, path: "/api/v1/documents/d1/some-unmapped-action", wantCap: "", wantVisibility: iamdelivery.VisibilitySessionRequired},
		{name: "unknown templates subpath session required", method: http.MethodDelete, path: "/api/v1/templates/t1/unmapped", wantCap: "", wantVisibility: iamdelivery.VisibilitySessionRequired},
		{name: "iam users patch roles falls through to session required", method: http.MethodPatch, path: "/api/v1/iam/users/u-1/roles", wantCap: "", wantVisibility: iamdelivery.VisibilitySessionRequired},
		{name: "completely unknown root session required", method: http.MethodGet, path: "/random/path", wantCap: "", wantVisibility: iamdelivery.VisibilitySessionRequired},

		// --- Permission-guarded: documents ---
		{name: "documents list guarded", method: http.MethodGet, path: "/api/v1/documents", wantCap: iamdomain.CapDocumentView, wantVisibility: iamdelivery.VisibilityPermissionGuarded},
		{name: "documents create", method: http.MethodPost, path: "/api/v1/documents", wantCap: iamdomain.CapDocumentCreate, wantVisibility: iamdelivery.VisibilityPermissionGuarded},
		{name: "documents submit", method: http.MethodPost, path: "/api/v1/documents/d1/submit", wantCap: iamdomain.CapDocumentSubmit, wantVisibility: iamdelivery.VisibilityPermissionGuarded},
		{name: "documents signoff", method: http.MethodPost, path: "/api/v1/documents/d1/signoff", wantCap: iamdomain.CapDocumentSignoff, wantVisibility: iamdelivery.VisibilityPermissionGuarded},
		{name: "documents publish", method: http.MethodPost, path: "/api/v1/documents/d1/publish", wantCap: iamdomain.Capability("doc.publish"), wantVisibility: iamdelivery.VisibilityPermissionGuarded},
		{name: "documents schedule publish", method: http.MethodPost, path: "/api/v1/documents/d1/schedule-publish", wantCap: iamdomain.Capability("doc.publish"), wantVisibility: iamdelivery.VisibilityPermissionGuarded},
		{name: "documents supersede", method: http.MethodPost, path: "/api/v1/documents/d1/supersede", wantCap: iamdomain.Capability("doc.supersede"), wantVisibility: iamdelivery.VisibilityPermissionGuarded},
		{name: "documents obsolete", method: http.MethodPost, path: "/api/v1/documents/d1/obsolete", wantCap: iamdomain.Capability("doc.obsolete"), wantVisibility: iamdelivery.VisibilityPermissionGuarded},
		{name: "documents artifact metadata", method: http.MethodPost, path: "/api/v1/documents/d1/artifact-metadata", wantCap: iamdomain.CapDocumentEdit, wantVisibility: iamdelivery.VisibilityPermissionGuarded},
		{name: "documents finalize", method: http.MethodPost, path: "/api/v1/documents/d1/finalize", wantCap: iamdomain.CapDocumentSignoff, wantVisibility: iamdelivery.VisibilityPermissionGuarded},
		{name: "documents session force release", method: http.MethodPost, path: "/api/v1/documents/d1/session/force-release", wantCap: iamdomain.CapMembershipManage, wantVisibility: iamdelivery.VisibilityPermissionGuarded},
		{name: "documents session generic", method: http.MethodPost, path: "/api/v1/documents/d1/session/acquire", wantCap: iamdomain.CapDocumentEdit, wantVisibility: iamdelivery.VisibilityPermissionGuarded},
		{name: "documents autosave", method: http.MethodPost, path: "/api/v1/documents/d1/autosave/commit", wantCap: iamdomain.CapDocumentEdit, wantVisibility: iamdelivery.VisibilityPermissionGuarded},
		{name: "documents checkpoint restore", method: http.MethodPost, path: "/api/v1/documents/d1/checkpoints/c1/restore", wantCap: iamdomain.CapDocumentEdit, wantVisibility: iamdelivery.VisibilityPermissionGuarded},
		{name: "documents checkpoint create", method: http.MethodPost, path: "/api/v1/documents/d1/checkpoints", wantCap: iamdomain.CapDocumentEdit, wantVisibility: iamdelivery.VisibilityPermissionGuarded},
		{name: "documents placeholder update", method: http.MethodPut, path: "/api/v1/documents/d1/placeholders/p1", wantCap: iamdomain.CapDocumentEdit, wantVisibility: iamdelivery.VisibilityPermissionGuarded},
		{name: "documents patch", method: http.MethodPatch, path: "/api/v1/documents/d1", wantCap: iamdomain.CapDocumentEdit, wantVisibility: iamdelivery.VisibilityPermissionGuarded},

		// --- Permission-guarded: templates ---
		{name: "templates list", method: http.MethodGet, path: "/api/v1/templates", wantCap: iamdomain.CapTemplateView, wantVisibility: iamdelivery.VisibilityPermissionGuarded},
		{name: "templates create", method: http.MethodPost, path: "/api/v1/templates", wantCap: iamdomain.CapTemplateCreate, wantVisibility: iamdelivery.VisibilityPermissionGuarded},
		{name: "templates version create", method: http.MethodPost, path: "/api/v1/templates/t1/versions", wantCap: iamdomain.CapTemplateCreate, wantVisibility: iamdelivery.VisibilityPermissionGuarded},
		{name: "templates draft", method: http.MethodPut, path: "/api/v1/templates/t1/versions/1/draft", wantCap: iamdomain.CapTemplateEdit, wantVisibility: iamdelivery.VisibilityPermissionGuarded},
		{name: "templates schema", method: http.MethodPut, path: "/api/v1/templates/t1/versions/1/schema", wantCap: iamdomain.CapTemplateEdit, wantVisibility: iamdelivery.VisibilityPermissionGuarded},
		{name: "templates publish", method: http.MethodPost, path: "/api/v1/templates/t1/versions/1/publish", wantCap: iamdomain.CapTemplatePublish, wantVisibility: iamdelivery.VisibilityPermissionGuarded},
		{name: "templates docx upload url", method: http.MethodPost, path: "/api/v1/templates/t1/versions/1/docx-upload-url", wantCap: iamdomain.CapTemplateEdit, wantVisibility: iamdelivery.VisibilityPermissionGuarded},
		{name: "templates schema upload url", method: http.MethodPost, path: "/api/v1/templates/t1/versions/1/schema-upload-url", wantCap: iamdomain.CapTemplateEdit, wantVisibility: iamdelivery.VisibilityPermissionGuarded},
		{name: "templates autosave presign", method: http.MethodPost, path: "/api/v1/templates/t1/versions/1/autosave/presign", wantCap: iamdomain.CapTemplateEdit, wantVisibility: iamdelivery.VisibilityPermissionGuarded},
		{name: "templates autosave commit", method: http.MethodPost, path: "/api/v1/templates/t1/versions/1/autosave/commit", wantCap: iamdomain.CapTemplateEdit, wantVisibility: iamdelivery.VisibilityPermissionGuarded},
		{name: "templates submit", method: http.MethodPost, path: "/api/v1/templates/t1/versions/1/submit", wantCap: iamdomain.CapTemplateSubmit, wantVisibility: iamdelivery.VisibilityPermissionGuarded},
		{name: "templates review", method: http.MethodPost, path: "/api/v1/templates/t1/versions/1/review", wantCap: iamdomain.CapTemplateReview, wantVisibility: iamdelivery.VisibilityPermissionGuarded},
		{name: "templates approve", method: http.MethodPost, path: "/api/v1/templates/t1/versions/1/approve", wantCap: iamdomain.CapTemplateApprove, wantVisibility: iamdelivery.VisibilityPermissionGuarded},
		{name: "templates approval config", method: http.MethodPut, path: "/api/v1/templates/t1/approval-config", wantCap: iamdomain.CapTemplateEdit, wantVisibility: iamdelivery.VisibilityPermissionGuarded},
		{name: "templates archive", method: http.MethodPost, path: "/api/v1/templates/t1/archive", wantCap: iamdomain.Capability("template.archive"), wantVisibility: iamdelivery.VisibilityPermissionGuarded},

		// --- Permission-guarded: IAM users (F-001 split: GET=view, writes=manage) ---
		{name: "iam users list", method: http.MethodGet, path: "/api/v1/iam/users", wantCap: iamdomain.CapUserView, wantVisibility: iamdelivery.VisibilityPermissionGuarded},
		{name: "iam users create", method: http.MethodPost, path: "/api/v1/iam/users", wantCap: iamdomain.CapUserManage, wantVisibility: iamdelivery.VisibilityPermissionGuarded},
		{name: "iam users patch profile", method: http.MethodPatch, path: "/api/v1/iam/users/u-1", wantCap: iamdomain.CapUserManage, wantVisibility: iamdelivery.VisibilityPermissionGuarded},
		{name: "iam users post roles", method: http.MethodPost, path: "/api/v1/iam/users/u-1/roles", wantCap: iamdomain.CapUserManage, wantVisibility: iamdelivery.VisibilityPermissionGuarded},
		{name: "iam users put roles", method: http.MethodPut, path: "/api/v1/iam/users/u-1/roles", wantCap: iamdomain.CapUserManage, wantVisibility: iamdelivery.VisibilityPermissionGuarded},
		{name: "iam users reset password", method: http.MethodPost, path: "/api/v1/iam/users/u-1/reset-password", wantCap: iamdomain.CapUserManage, wantVisibility: iamdelivery.VisibilityPermissionGuarded},
		{name: "iam users unlock", method: http.MethodPost, path: "/api/v1/iam/users/u-1/unlock", wantCap: iamdomain.CapUserManage, wantVisibility: iamdelivery.VisibilityPermissionGuarded},
		{name: "iam admin overview", method: http.MethodGet, path: "/api/v1/iam/admin/overview", wantCap: iamdomain.CapUserView, wantVisibility: iamdelivery.VisibilityPermissionGuarded},

		// --- Permission-guarded: taxonomy (F-001 split: GET=view, writes=manage) + legacy aliases ---
		{name: "taxonomy families list", method: http.MethodGet, path: "/api/v1/taxonomy/families", wantCap: iamdomain.CapTaxonomyView, wantVisibility: iamdelivery.VisibilityPermissionGuarded},
		{name: "taxonomy families create", method: http.MethodPost, path: "/api/v1/taxonomy/families", wantCap: iamdomain.CapTaxonomyManage, wantVisibility: iamdelivery.VisibilityPermissionGuarded},
		{name: "taxonomy families patch", method: http.MethodPatch, path: "/api/v1/taxonomy/families/PROC", wantCap: iamdomain.CapTaxonomyManage, wantVisibility: iamdelivery.VisibilityPermissionGuarded},
		{name: "taxonomy areas list", method: http.MethodGet, path: "/api/v1/taxonomy/areas", wantCap: iamdomain.CapTaxonomyView, wantVisibility: iamdelivery.VisibilityPermissionGuarded},
		{name: "taxonomy areas patch", method: http.MethodPatch, path: "/api/v1/taxonomy/areas/QA", wantCap: iamdomain.CapTaxonomyManage, wantVisibility: iamdelivery.VisibilityPermissionGuarded},
		{name: "taxonomy profiles list", method: http.MethodGet, path: "/api/v1/taxonomy/profiles", wantCap: iamdomain.CapTaxonomyView, wantVisibility: iamdelivery.VisibilityPermissionGuarded},
		{name: "legacy document-profiles list", method: http.MethodGet, path: "/api/v1/document-profiles", wantCap: iamdomain.CapDocumentView, wantVisibility: iamdelivery.VisibilityPermissionGuarded},
		{name: "legacy process-areas create", method: http.MethodPost, path: "/api/v1/process-areas", wantCap: iamdomain.CapTaxonomyManage, wantVisibility: iamdelivery.VisibilityPermissionGuarded},
		{name: "legacy document-subjects delete", method: http.MethodDelete, path: "/api/v1/document-subjects/s-1", wantCap: iamdomain.CapTaxonomyManage, wantVisibility: iamdelivery.VisibilityPermissionGuarded},

		// --- Permission-guarded: controlled documents ---
		{name: "controlled documents list", method: http.MethodGet, path: "/api/v1/controlled-documents", wantCap: iamdomain.CapDocumentView, wantVisibility: iamdelivery.VisibilityPermissionGuarded},
		{name: "controlled documents create", method: http.MethodPost, path: "/api/v1/controlled-documents", wantCap: iamdomain.CapControlledDocumentCreate, wantVisibility: iamdelivery.VisibilityPermissionGuarded},
		{name: "controlled documents revisions create", method: http.MethodPost, path: "/api/v1/controlled-documents/cd-1/revisions", wantCap: iamdomain.CapDocumentEdit, wantVisibility: iamdelivery.VisibilityPermissionGuarded},
		{name: "controlled documents preview code", method: http.MethodGet, path: "/api/v1/controlled-documents/preview-code", wantCap: iamdomain.CapDocumentView, wantVisibility: iamdelivery.VisibilityPermissionGuarded},
		{name: "controlled documents obsolete", method: http.MethodPut, path: "/api/v1/controlled-documents/cd-1/obsolete", wantCap: iamdomain.CapControlledDocumentObsolete, wantVisibility: iamdelivery.VisibilityPermissionGuarded},
		{name: "controlled documents supersede", method: http.MethodPut, path: "/api/v1/controlled-documents/cd-1/supersede", wantCap: iamdomain.CapControlledDocumentSupersede, wantVisibility: iamdelivery.VisibilityPermissionGuarded},

		// --- Permission-guarded: workflow / search / notifications / access-policies ---
		{name: "workflow transitions", method: http.MethodPost, path: "/api/v1/workflow/documents/doc-1/transitions", wantCap: iamdomain.CapDocumentSubmit, wantVisibility: iamdelivery.VisibilityPermissionGuarded},
		{name: "workflow approvals", method: http.MethodGet, path: "/api/v1/workflow/documents/doc-1/approvals", wantCap: iamdomain.CapDocumentView, wantVisibility: iamdelivery.VisibilityPermissionGuarded},
		{name: "search documents", method: http.MethodGet, path: "/api/v1/search/documents", wantCap: iamdomain.CapDocumentView, wantVisibility: iamdelivery.VisibilityPermissionGuarded},
		{name: "notifications list", method: http.MethodGet, path: "/api/v1/notifications", wantCap: iamdomain.CapDocumentView, wantVisibility: iamdelivery.VisibilityPermissionGuarded},
		{name: "notification mark read", method: http.MethodPost, path: "/api/v1/notifications/n-1/read", wantCap: iamdomain.CapDocumentView, wantVisibility: iamdelivery.VisibilityPermissionGuarded},
		{name: "access policies get", method: http.MethodGet, path: "/api/v1/access-policies", wantCap: iamdomain.CapMembershipView, wantVisibility: iamdelivery.VisibilityPermissionGuarded},
		{name: "access policies put", method: http.MethodPut, path: "/api/v1/access-policies", wantCap: iamdomain.CapMembershipManage, wantVisibility: iamdelivery.VisibilityPermissionGuarded},

		// --- Permission-guarded: misc (F-001 split: area-memberships GET=view, writes=manage) ---
		{name: "iam area memberships list", method: http.MethodGet, path: "/api/v1/iam/area-memberships", wantCap: iamdomain.CapMembershipView, wantVisibility: iamdelivery.VisibilityPermissionGuarded},
		{name: "iam area memberships create", method: http.MethodPost, path: "/api/v1/iam/area-memberships", wantCap: iamdomain.CapMembershipManage, wantVisibility: iamdelivery.VisibilityPermissionGuarded},
		{name: "iam area memberships delete", method: http.MethodDelete, path: "/api/v1/iam/area-memberships/m-1", wantCap: iamdomain.CapMembershipManage, wantVisibility: iamdelivery.VisibilityPermissionGuarded},
		{name: "signed download", method: http.MethodGet, path: "/api/v1/signed", wantCap: iamdomain.CapTemplateView, wantVisibility: iamdelivery.VisibilityPermissionGuarded},
		{name: "approval get", method: http.MethodGet, path: "/api/v1/approval/instances/a-1", wantCap: iamdomain.CapDocumentView, wantVisibility: iamdelivery.VisibilityPermissionGuarded},
		{name: "approval post", method: http.MethodPost, path: "/api/v1/approval/instances/a-1/decisions", wantCap: iamdomain.CapDocumentSubmit, wantVisibility: iamdelivery.VisibilityPermissionGuarded},
		{name: "audit events", method: http.MethodGet, path: "/api/v1/audit/events", wantCap: iamdomain.CapAuditRead, wantVisibility: iamdelivery.VisibilityPermissionGuarded},
		{name: "metrics", method: http.MethodGet, path: "/api/v1/metrics", wantCap: iamdomain.CapMetricsView, wantVisibility: iamdelivery.VisibilityPermissionGuarded},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			gotCap, gotVisibility := resolver(tc.method, tc.path)
			if gotVisibility != tc.wantVisibility {
				t.Fatalf("visibility mismatch for %s %s: got %v want %v", tc.method, tc.path, gotVisibility, tc.wantVisibility)
			}
			if gotCap != tc.wantCap {
				t.Fatalf("capability mismatch for %s %s: got %q want %q", tc.method, tc.path, gotCap, tc.wantCap)
			}
		})
	}
}

func TestPublicPathChecker_RespectsPublicAndPrivateBoundaries(t *testing.T) {
	t.Parallel()

	checker := newPublicPathChecker(newPermissionResolver())

	testCases := []struct {
		name   string
		method string
		path   string
		want   bool
	}{
		{name: "health ready public", method: http.MethodGet, path: "/api/v1/health/ready", want: true},
		{name: "auth login public", method: http.MethodPost, path: "/api/v1/auth/login", want: true},
		{name: "auth refresh public", method: http.MethodPost, path: "/api/v1/auth/refresh", want: true},
		{name: "feature flags public", method: http.MethodGet, path: "/api/v1/feature-flags", want: true},

		{name: "auth me not public", method: http.MethodGet, path: "/api/v1/auth/me", want: false},
		{name: "auth logout not public", method: http.MethodPost, path: "/api/v1/auth/logout", want: false},
		{name: "documents list not public", method: http.MethodGet, path: "/api/v1/documents", want: false},
		{name: "publish not public", method: http.MethodPost, path: "/api/v1/documents/d1/publish", want: false},
		{name: "artifact metadata not public", method: http.MethodPost, path: "/api/v1/documents/d1/artifact-metadata", want: false},

		// C2 fail-closed regressions: unmatched routes are NEVER public.
		{name: "unknown route not public", method: http.MethodGet, path: "/api/v1/unknown", want: false},
		{name: "unknown documents subpath not public", method: http.MethodDelete, path: "/api/v1/documents/d1/some-unmapped-action", want: false},
		{name: "completely unknown root not public", method: http.MethodGet, path: "/random/path", want: false},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := checker(tc.method, tc.path)
			if got != tc.want {
				t.Fatalf("checker(%s, %s) = %v, want %v", tc.method, tc.path, got, tc.want)
			}
		})
	}
}

// TestRouteCoverage walks every mux.Handle / RegisterRoutes family registered in
// main.go and asserts the path is matched by some routeRule — i.e. the resolver
// does NOT fall through to the SessionRequired default. This is the structural
// guard for C2: forgetting to add a new route to routeRules will fail this test
// loudly instead of silently demoting the route to "session-only" (or, pre-fix,
// to "public").
//
// Fixture list, not reflection: representative path per registered family.
// When adding a new RegisterRoutes call in main.go, add an entry here.
func TestRouteCoverage(t *testing.T) {
	t.Parallel()

	registeredRoutes := []struct {
		family string
		method string
		path   string
	}{
		// authHandler.RegisterRoutes (main.go:212)
		{"auth", http.MethodPost, "/api/v1/auth/login"},
		{"auth", http.MethodPost, "/api/v1/auth/refresh"},
		{"auth", http.MethodGet, "/api/v1/auth/me"},
		{"auth", http.MethodPost, "/api/v1/auth/change-password"},
		{"auth", http.MethodPost, "/api/v1/auth/logout"},

		// healthHandler.RegisterRoutes (main.go:213)
		{"health", http.MethodGet, "/api/v1/health/live"},
		{"health", http.MethodGet, "/api/v1/health/ready"},
		{"health", http.MethodGet, "/healthz"},

		// featureFlagsHandler.RegisterRoutes (main.go:214)
		{"feature-flags", http.MethodGet, "/api/v1/feature-flags"},

		// auditHandler.RegisterRoutes (main.go:215)
		{"audit", http.MethodGet, "/api/v1/audit/events"},

		// searchHandler.RegisterRoutes (main.go:216)
		{"search", http.MethodGet, "/api/v1/search/documents"},

		// iamAdminHandler.RegisterRoutes (main.go:217)
		{"iam-admin", http.MethodGet, "/api/v1/iam/users"},
		{"iam-admin", http.MethodPost, "/api/v1/iam/users"},
		{"iam-admin", http.MethodPatch, "/api/v1/iam/users/u-1"},
		{"iam-admin", http.MethodPost, "/api/v1/iam/users/u-1/roles"},
		{"iam-admin", http.MethodPut, "/api/v1/iam/users/u-1/roles"},
		{"iam-admin", http.MethodPost, "/api/v1/iam/users/u-1/reset-password"},
		{"iam-admin", http.MethodPost, "/api/v1/iam/users/u-1/unlock"},
		{"iam-admin", http.MethodGet, "/api/v1/iam/admin/overview"},

		// taxonomyModule.RegisterRoutes (main.go:224) + legacy aliases
		{"taxonomy", http.MethodGet, "/api/v1/taxonomy/families"},
		{"taxonomy", http.MethodPost, "/api/v1/taxonomy/families"},
		{"taxonomy", http.MethodPatch, "/api/v1/taxonomy/families/PROC"},
		{"taxonomy", http.MethodGet, "/api/v1/taxonomy/areas"},
		{"taxonomy", http.MethodPatch, "/api/v1/taxonomy/areas/QA"},
		{"taxonomy", http.MethodGet, "/api/v1/taxonomy/profiles"},
		{"taxonomy", http.MethodPost, "/api/v1/taxonomy/profiles"},
		{"taxonomy-legacy", http.MethodGet, "/api/v1/document-profiles"},
		{"taxonomy-legacy", http.MethodPost, "/api/v1/document-profiles"},
		{"taxonomy-legacy", http.MethodGet, "/api/v1/process-areas"},
		{"taxonomy-legacy", http.MethodPost, "/api/v1/process-areas"},
		{"taxonomy-legacy", http.MethodGet, "/api/v1/document-subjects"},
		{"taxonomy-legacy", http.MethodDelete, "/api/v1/document-subjects/s-1"},

		// controlledDocumentsModule.RegisterRoutes (main.go:231)
		{"controlled-documents", http.MethodGet, "/api/v1/controlled-documents"},
		{"controlled-documents", http.MethodPost, "/api/v1/controlled-documents"},
		{"controlled-documents", http.MethodPost, "/api/v1/controlled-documents/cd-1/revisions"},
		{"controlled-documents", http.MethodGet, "/api/v1/controlled-documents/preview-code"},
		{"controlled-documents", http.MethodPut, "/api/v1/controlled-documents/cd-1/obsolete"},
		{"controlled-documents", http.MethodPut, "/api/v1/controlled-documents/cd-1/supersede"},

		// iamdelivery.NewMembershipHandler(...).RegisterRoutes (main.go:238)
		{"area-memberships", http.MethodGet, "/api/v1/iam/area-memberships"},
		{"area-memberships", http.MethodPost, "/api/v1/iam/area-memberships"},
		{"area-memberships", http.MethodDelete, "/api/v1/iam/area-memberships/m-1"},

		// docMod.RegisterRoutes (main.go:379)
		{"documents", http.MethodGet, "/api/v1/documents"},
		{"documents", http.MethodPost, "/api/v1/documents"},
		{"documents", http.MethodPost, "/api/v1/documents/d1/submit"},
		{"documents", http.MethodPost, "/api/v1/documents/d1/signoff"},
		{"documents", http.MethodPost, "/api/v1/documents/d1/publish"},
		{"documents", http.MethodPost, "/api/v1/documents/d1/schedule-publish"},
		{"documents", http.MethodPost, "/api/v1/documents/d1/supersede"},
		{"documents", http.MethodPost, "/api/v1/documents/d1/obsolete"},
		{"documents", http.MethodPost, "/api/v1/documents/d1/finalize"},
		{"documents", http.MethodPost, "/api/v1/documents/d1/cancel"},
		{"documents", http.MethodPost, "/api/v1/documents/d1/reconstruct"},
		{"documents", http.MethodPost, "/api/v1/documents/d1/archive"},
		{"documents", http.MethodPost, "/api/v1/documents/d1/export/pdf"},
		{"documents", http.MethodPost, "/api/v1/documents/d1/artifact-metadata"},
		{"documents", http.MethodPost, "/api/v1/documents/d1/session/acquire"},
		{"documents", http.MethodPost, "/api/v1/documents/d1/session/force-release"},
		{"documents", http.MethodPost, "/api/v1/documents/d1/autosave/commit"},
		{"documents", http.MethodPost, "/api/v1/documents/d1/checkpoints"},
		{"documents", http.MethodPost, "/api/v1/documents/d1/checkpoints/c1/restore"},
		{"documents", http.MethodPut, "/api/v1/documents/d1/placeholders/p1"},
		{"documents", http.MethodPatch, "/api/v1/documents/d1"},

		// templateshttp.New(...).Register (main.go:393)
		{"templates", http.MethodGet, "/api/v1/templates"},
		{"templates", http.MethodPost, "/api/v1/templates"},
		{"templates", http.MethodPost, "/api/v1/templates/t1/versions"},
		{"templates", http.MethodPut, "/api/v1/templates/t1/versions/1/draft"},
		{"templates", http.MethodPut, "/api/v1/templates/t1/versions/1/schema"},
		{"templates", http.MethodPost, "/api/v1/templates/t1/versions/1/publish"},
		{"templates", http.MethodPost, "/api/v1/templates/t1/versions/1/docx-upload-url"},
		{"templates", http.MethodPost, "/api/v1/templates/t1/versions/1/schema-upload-url"},
		{"templates", http.MethodPost, "/api/v1/templates/t1/versions/1/autosave/presign"},
		{"templates", http.MethodPost, "/api/v1/templates/t1/versions/1/submit"},
		{"templates", http.MethodPost, "/api/v1/templates/t1/versions/1/review"},
		{"templates", http.MethodPost, "/api/v1/templates/t1/versions/1/approve"},
		{"templates", http.MethodPut, "/api/v1/templates/t1/approval-config"},
		{"templates", http.MethodPost, "/api/v1/templates/t1/archive"},
		{"templates", http.MethodGet, "/api/v1/signed"},

		// approvalHandler.RegisterRoutes (main.go:396)
		{"approval", http.MethodGet, "/api/v1/approval/instances/a-1"},
		{"approval", http.MethodPost, "/api/v1/approval/instances/a-1/decisions"},
		{"approval", http.MethodPut, "/api/v1/approval/instances/a-1"},
		{"approval", http.MethodDelete, "/api/v1/approval/instances/a-1"},
		{"workflow", http.MethodPost, "/api/v1/workflow/documents/doc-1/transitions"},
		{"workflow", http.MethodGet, "/api/v1/workflow/documents/doc-1/approvals"},
		{"notifications", http.MethodGet, "/api/v1/notifications"},
		{"notifications", http.MethodPost, "/api/v1/notifications/n-1/read"},
		{"access-policies", http.MethodGet, "/api/v1/access-policies"},
		{"access-policies", http.MethodPut, "/api/v1/access-policies"},

		// mux.Handle (main.go:447)
		{"metrics", http.MethodGet, "/api/v1/metrics"},
	}

	resolver := newPermissionResolver()

	for _, rt := range registeredRoutes {
		rt := rt
		t.Run(rt.family+" "+rt.method+" "+rt.path, func(t *testing.T) {
			t.Parallel()

			if !matchedByRule(rt.method, rt.path) {
				_, visibility := resolver(rt.method, rt.path)
				t.Fatalf("registered route %s %s (%s) fell through to default visibility=%v — add an entry to routeRules",
					rt.method, rt.path, rt.family, visibility)
			}
		})
	}
}

// matchedByRule reports whether some routeRule explicitly matches (method, path).
// Distinguishes "matched as session-required" (intentional) from "unmatched →
// session-required default" (forgotten registration). Used by TestRouteCoverage.
func matchedByRule(method, path string) bool {
	for _, rule := range routeRules {
		if rule.matches(method, path) {
			return true
		}
	}
	return false
}

// TestPermissionsTable_NoMethodlessWriteShadowing locks the F-001 authoring
// invariants. Fails if any rule:
//   (a) has empty method AND a Manage/Submit/write-grade cap, OR
//   (b) has empty method on a prefix where some OTHER rule declares a write
//       verb (so the methodless row would shadow per-verb intent).
//
// See wiki/concepts/authz-tiers.md §Tier-1 rule authoring rules.
func TestPermissionsTable_NoMethodlessWriteShadowing(t *testing.T) {
	t.Parallel()

	writeMethods := map[string]bool{
		http.MethodPost:   true,
		http.MethodPut:    true,
		http.MethodPatch:  true,
		http.MethodDelete: true,
	}
	isWriteCap := func(c iamdomain.Capability) bool {
		s := string(c)
		return strings.Contains(s, ".manage") ||
			strings.Contains(s, ".submit") ||
			strings.Contains(s, ".create") ||
			strings.Contains(s, ".edit") ||
			strings.Contains(s, ".approve") ||
			strings.Contains(s, ".publish") ||
			strings.Contains(s, ".signoff") ||
			strings.Contains(s, ".obsolete") ||
			strings.Contains(s, ".supersede") ||
			strings.Contains(s, ".review")
	}

	// (a) methodless row carrying a write-grade cap.
	for i, r := range routeRules {
		if r.method == "" && r.capability != "" && isWriteCap(r.capability) {
			t.Errorf("routeRules[%d]: methodless rule cannot carry write-grade cap %q (prefix=%q, exact=%q). Split into per-verb rows.",
				i, r.capability, r.pathPrefix, r.pathExact)
		}
	}

	// (b) methodless prefix overlapping a path declared with a write verb elsewhere.
	for i, mr := range routeRules {
		if mr.method != "" || mr.pathPrefix == "" {
			continue
		}
		for j, other := range routeRules {
			if i == j || !writeMethods[other.method] {
				continue
			}
			overlap := false
			if other.pathPrefix != "" && strings.HasPrefix(other.pathPrefix, mr.pathPrefix) {
				overlap = true
			}
			if other.pathExact != "" && strings.HasPrefix(other.pathExact, mr.pathPrefix) {
				overlap = true
			}
			if overlap {
				t.Errorf("routeRules[%d]: methodless prefix %q shadows write-verb rule routeRules[%d] (%s %s%s). Replace methodless row with per-verb rows.",
					i, mr.pathPrefix, j, other.method, other.pathPrefix, other.pathExact)
			}
		}
	}
}

// TestPermissionResolver_PeopleHandlerRoutes asserts the PR-4 People-tab
// routes resolve to VisibilityPermissionGuarded with the documented cap.
// Without explicit rules they would fall through to VisibilitySessionRequired
// — the long-term guard against future route additions slipping the gate.
func TestPermissionResolver_PeopleHandlerRoutes(t *testing.T) {
	t.Parallel()

	resolver := newPermissionResolver()
	cases := []struct {
		method string
		path   string
		cap    iamdomain.Capability
	}{
		{http.MethodGet, "/api/v1/iam/users", iamdomain.CapUserView},
		{http.MethodPost, "/api/v1/iam/users/invite", iamdomain.CapUserManage},
		{http.MethodPost, "/api/v1/iam/users/bulk", iamdomain.CapUserManage},
		{http.MethodPatch, "/api/v1/iam/users/u-1", iamdomain.CapUserManage},
		{http.MethodPost, "/api/v1/iam/users/u-1/reset-password", iamdomain.CapUserManage},
		{http.MethodPost, "/api/v1/iam/users/u-1/unlock", iamdomain.CapUserManage},
		{http.MethodGet, "/api/v1/iam/users/u-1/memberships", iamdomain.CapMembershipView},
	}
	for _, tc := range cases {
		gotCap, gotVis := resolver(tc.method, tc.path)
		if gotVis != iamdelivery.VisibilityPermissionGuarded {
			t.Errorf("%s %s: visibility=%v want PermissionGuarded", tc.method, tc.path, gotVis)
		}
		if gotCap != tc.cap {
			t.Errorf("%s %s: cap=%q want %q", tc.method, tc.path, gotCap, tc.cap)
		}
	}
}
