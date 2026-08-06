package main

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"testing"
	"time"

	approvalhttp "metaldocs/internal/modules/approval/http"
	auditdelivery "metaldocs/internal/modules/audit/delivery/http"
	auditdomain "metaldocs/internal/modules/audit/domain"
	authdelivery "metaldocs/internal/modules/auth/delivery/http"
	"metaldocs/internal/modules/controlleddocuments"
	cddhttp "metaldocs/internal/modules/controlleddocuments/delivery/http"
	distributionhttp "metaldocs/internal/modules/distribution/delivery/http"
	distributiondomain "metaldocs/internal/modules/distribution/domain"
	documents "metaldocs/internal/modules/documents"
	dhttp "metaldocs/internal/modules/documents/delivery/http"
	iamdelivery "metaldocs/internal/modules/iam/delivery/http"
	iamdomain "metaldocs/internal/modules/iam/domain"
	iampresence "metaldocs/internal/modules/iam/presence"
	notificationshttp "metaldocs/internal/modules/notifications/delivery/http"
	notificationsdomain "metaldocs/internal/modules/notifications/domain"
	searchdelivery "metaldocs/internal/modules/search/delivery/http"
	securitydelivery "metaldocs/internal/modules/security/delivery/http"
	"metaldocs/internal/modules/taxonomy"
	thttp "metaldocs/internal/modules/taxonomy/delivery/http"
	templateshttp "metaldocs/internal/modules/templates/delivery/http"
	"metaldocs/internal/modules/tokens"
	tokensapplication "metaldocs/internal/modules/tokens/application"
	tokenshttp "metaldocs/internal/modules/tokens/delivery/http"
	tokensdomain "metaldocs/internal/modules/tokens/domain"
	"metaldocs/internal/platform/config"
	"metaldocs/internal/platform/featureflags"
	"metaldocs/internal/platform/httprouter"
	"metaldocs/internal/platform/observability"
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
		// /healthz is deleted (operator ruling C): it has no routeRules entry
		// and falls through to the fail-closed default, so authn rejects it
		// with 401 before the mux (which no longer mounts it at all) is ever
		// reached — §10.3 row 3.
		{name: "healthz falls through to session required (deleted, not public)", method: http.MethodGet, path: "/healthz", wantCap: "", wantVisibility: iamdelivery.VisibilitySessionRequired},
		{name: "auth login public", method: http.MethodPost, path: "/api/v1/auth/login", wantCap: "", wantVisibility: iamdelivery.VisibilityPublic},
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
		{name: "documents submit", method: http.MethodPost, path: "/api/v1/documents/d1/submit", wantCap: iamdomain.CapDocumentSubmit, wantVisibility: iamdelivery.VisibilityPermissionGuarded},
		{name: "documents signoff", method: http.MethodPost, path: "/api/v1/documents/d1/signoff", wantCap: iamdomain.CapDocumentSignoff, wantVisibility: iamdelivery.VisibilityPermissionGuarded},
		{name: "documents obsolete", method: http.MethodPost, path: "/api/v1/documents/d1/obsolete", wantCap: iamdomain.CapDocumentObsolete, wantVisibility: iamdelivery.VisibilityPermissionGuarded},
		{name: "documents session force release", method: http.MethodPost, path: "/api/v1/documents/d1/session/force-release", wantCap: iamdomain.CapMembershipManage, wantVisibility: iamdelivery.VisibilityPermissionGuarded},
		{name: "documents session generic", method: http.MethodPost, path: "/api/v1/documents/d1/session/acquire", wantCap: iamdomain.CapDocumentEdit, wantVisibility: iamdelivery.VisibilityPermissionGuarded},
		{name: "documents autosave", method: http.MethodPost, path: "/api/v1/documents/d1/autosave/commit", wantCap: iamdomain.CapDocumentEdit, wantVisibility: iamdelivery.VisibilityPermissionGuarded},
		{name: "documents checkpoint restore", method: http.MethodPost, path: "/api/v1/documents/d1/checkpoints/c1/restore", wantCap: iamdomain.CapDocumentEdit, wantVisibility: iamdelivery.VisibilityPermissionGuarded},
		{name: "documents checkpoint create", method: http.MethodPost, path: "/api/v1/documents/d1/checkpoints", wantCap: iamdomain.CapDocumentEdit, wantVisibility: iamdelivery.VisibilityPermissionGuarded},
		{name: "documents placeholder update", method: http.MethodPut, path: "/api/v1/documents/d1/placeholders/p1", wantCap: iamdomain.CapDocumentEdit, wantVisibility: iamdelivery.VisibilityPermissionGuarded},
		{name: "documents patch", method: http.MethodPatch, path: "/api/v1/documents/d1", wantCap: iamdomain.CapDocumentEdit, wantVisibility: iamdelivery.VisibilityPermissionGuarded},

		// --- Permission-guarded: distribution (M2) ---
		// These GET routes MUST resolve to CapDistributionRead, NOT be shadowed by the
		// generic GET /api/v1/documents → CapDocumentView catch-all. Without these rows a
		// future reorder of routeRules could silently broaden distribution to document.view
		// (a privilege broadening) with a green suite. Mirrors the CapAuditRead/CapRouteManage
		// "more-specific cap must win over the generic block" precedent.
		{name: "distribution summary resolves to distribution.read", method: http.MethodGet, path: "/api/v1/documents/d1/distribution", wantCap: iamdomain.CapDistributionRead, wantVisibility: iamdelivery.VisibilityPermissionGuarded},
		{name: "distribution recipients resolves to distribution.read", method: http.MethodGet, path: "/api/v1/documents/d1/distribution/recipients", wantCap: iamdomain.CapDistributionRead, wantVisibility: iamdelivery.VisibilityPermissionGuarded},
		{name: "distribution coverage resolves to distribution.read", method: http.MethodGet, path: "/api/v1/documents/d1/distribution/coverage", wantCap: iamdomain.CapDistributionRead, wantVisibility: iamdelivery.VisibilityPermissionGuarded},

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
		{name: "templates submit-for-approval", method: http.MethodPost, path: "/api/v1/templates/t1/versions/1/submit-for-approval", wantCap: iamdomain.CapTemplateSubmit, wantVisibility: iamdelivery.VisibilityPermissionGuarded},
		{name: "templates signoff", method: http.MethodPost, path: "/api/v1/templates/t1/versions/1/signoff", wantCap: iamdomain.CapTemplateApprove, wantVisibility: iamdelivery.VisibilityPermissionGuarded},
		{name: "templates archive", method: http.MethodPost, path: "/api/v1/templates/t1/archive", wantCap: iamdomain.CapTemplateArchive, wantVisibility: iamdelivery.VisibilityPermissionGuarded},

		// --- Permission-guarded: IAM users (F-001 split: GET=view, writes=manage) ---
		{name: "iam users list", method: http.MethodGet, path: "/api/v1/iam/users", wantCap: iamdomain.CapUserView, wantVisibility: iamdelivery.VisibilityPermissionGuarded},
		{name: "iam users create", method: http.MethodPost, path: "/api/v1/iam/users", wantCap: iamdomain.CapUserManage, wantVisibility: iamdelivery.VisibilityPermissionGuarded},
		{name: "iam users patch profile", method: http.MethodPatch, path: "/api/v1/iam/users/u-1", wantCap: iamdomain.CapUserManage, wantVisibility: iamdelivery.VisibilityPermissionGuarded},
		{name: "iam users post roles", method: http.MethodPost, path: "/api/v1/iam/users/u-1/roles", wantCap: iamdomain.CapUserManage, wantVisibility: iamdelivery.VisibilityPermissionGuarded},
		{name: "iam users put roles", method: http.MethodPut, path: "/api/v1/iam/users/u-1/roles", wantCap: iamdomain.CapUserManage, wantVisibility: iamdelivery.VisibilityPermissionGuarded},
		{name: "iam users reset password", method: http.MethodPost, path: "/api/v1/iam/users/u-1/reset-password", wantCap: iamdomain.CapUserManage, wantVisibility: iamdelivery.VisibilityPermissionGuarded},
		{name: "iam users unlock", method: http.MethodPost, path: "/api/v1/iam/users/u-1/unlock", wantCap: iamdomain.CapUserManage, wantVisibility: iamdelivery.VisibilityPermissionGuarded},
		{name: "iam admin overview", method: http.MethodGet, path: "/api/v1/iam/admin/overview", wantCap: iamdomain.CapUserView, wantVisibility: iamdelivery.VisibilityPermissionGuarded},

		// --- Permission-guarded: taxonomy (F-001 split: GET=view, writes=manage) ---
		{name: "taxonomy families list", method: http.MethodGet, path: "/api/v1/taxonomy/families", wantCap: iamdomain.CapTaxonomyView, wantVisibility: iamdelivery.VisibilityPermissionGuarded},
		{name: "taxonomy families create", method: http.MethodPost, path: "/api/v1/taxonomy/families", wantCap: iamdomain.CapTaxonomyManage, wantVisibility: iamdelivery.VisibilityPermissionGuarded},
		{name: "taxonomy families patch", method: http.MethodPatch, path: "/api/v1/taxonomy/families/PROC", wantCap: iamdomain.CapTaxonomyManage, wantVisibility: iamdelivery.VisibilityPermissionGuarded},
		{name: "taxonomy areas list", method: http.MethodGet, path: "/api/v1/taxonomy/areas", wantCap: iamdomain.CapTaxonomyView, wantVisibility: iamdelivery.VisibilityPermissionGuarded},
		{name: "taxonomy areas patch", method: http.MethodPatch, path: "/api/v1/taxonomy/areas/QA", wantCap: iamdomain.CapTaxonomyManage, wantVisibility: iamdelivery.VisibilityPermissionGuarded},
		{name: "taxonomy profiles list", method: http.MethodGet, path: "/api/v1/taxonomy/profiles", wantCap: iamdomain.CapTaxonomyView, wantVisibility: iamdelivery.VisibilityPermissionGuarded},

		// --- Permission-guarded: controlled documents ---
		{name: "controlled documents list", method: http.MethodGet, path: "/api/v1/controlled-documents", wantCap: iamdomain.CapDocumentView, wantVisibility: iamdelivery.VisibilityPermissionGuarded},
		{name: "controlled documents create", method: http.MethodPost, path: "/api/v1/controlled-documents", wantCap: iamdomain.CapControlledDocumentCreate, wantVisibility: iamdelivery.VisibilityPermissionGuarded},
		{name: "controlled documents revisions create", method: http.MethodPost, path: "/api/v1/controlled-documents/cd-1/revisions", wantCap: iamdomain.CapDocumentEdit, wantVisibility: iamdelivery.VisibilityPermissionGuarded},
		// Create-flow GETs resolve to CapControlledDocumentCreate, not the
		// generic CapDocumentView prefix rule: creation-context's body IS the
		// caller's create authorization, and preview-code's tier-2
		// (PeekSeq → authz.Require) has always required create — the tier-1
		// row now agrees instead of admitting view-only callers (F-QA4-1).
		{name: "controlled documents creation context", method: http.MethodGet, path: "/api/v1/controlled-documents/creation-context", wantCap: iamdomain.CapControlledDocumentCreate, wantVisibility: iamdelivery.VisibilityPermissionGuarded},
		{name: "controlled documents preview code", method: http.MethodGet, path: "/api/v1/controlled-documents/preview-code", wantCap: iamdomain.CapControlledDocumentCreate, wantVisibility: iamdelivery.VisibilityPermissionGuarded},
		{name: "controlled documents obsolete", method: http.MethodPut, path: "/api/v1/controlled-documents/cd-1/obsolete", wantCap: iamdomain.CapControlledDocumentObsolete, wantVisibility: iamdelivery.VisibilityPermissionGuarded},
		{name: "controlled documents supersede", method: http.MethodPut, path: "/api/v1/controlled-documents/cd-1/supersede", wantCap: iamdomain.CapControlledDocumentSupersede, wantVisibility: iamdelivery.VisibilityPermissionGuarded},

		// --- Permission-guarded: search ---
		{name: "search documents", method: http.MethodGet, path: "/api/v1/search/documents", wantCap: iamdomain.CapDocumentView, wantVisibility: iamdelivery.VisibilityPermissionGuarded},

		// --- Permission-guarded: misc (F-001 split: area-memberships GET=view, writes=manage) ---
		{name: "iam area memberships list", method: http.MethodGet, path: "/api/v1/iam/area-memberships", wantCap: iamdomain.CapMembershipView, wantVisibility: iamdelivery.VisibilityPermissionGuarded},
		{name: "iam area memberships create", method: http.MethodPost, path: "/api/v1/iam/area-memberships", wantCap: iamdomain.CapMembershipManage, wantVisibility: iamdelivery.VisibilityPermissionGuarded},
		{name: "iam area memberships delete", method: http.MethodDelete, path: "/api/v1/iam/area-memberships/u-1/quality", wantCap: iamdomain.CapMembershipManage, wantVisibility: iamdelivery.VisibilityPermissionGuarded},
		{name: "signed download", method: http.MethodGet, path: "/api/v1/signed", wantCap: iamdomain.CapTemplateView, wantVisibility: iamdelivery.VisibilityPermissionGuarded},
		{name: "approval get instance", method: http.MethodGet, path: "/api/v1/approval/instances/a-1", wantCap: iamdomain.CapDocumentView, wantVisibility: iamdelivery.VisibilityPermissionGuarded},
		// M2b F3 (P1): generic /api/v1/approval/ prefix fallback deleted; these are
		// the real runtime routes (verified against internal/modules/approval/http/router.go),
		// each now explicit and matching its real tier-2 capability.
		{name: "approval signoff", method: http.MethodPost, path: "/api/v1/approval/instances/a-1/stages/s-1/signoffs", wantCap: iamdomain.CapDocumentSignoff, wantVisibility: iamdelivery.VisibilityPermissionGuarded},
		{name: "approval cancel", method: http.MethodPost, path: "/api/v1/approval/instances/a-1/cancel", wantCap: iamdomain.CapDocumentEdit, wantVisibility: iamdelivery.VisibilityPermissionGuarded},
		{name: "approval inbox", method: http.MethodGet, path: "/api/v1/approval/inbox", wantCap: iamdomain.CapDocumentView, wantVisibility: iamdelivery.VisibilityPermissionGuarded},
		// ADR 0022 Phase 11 (F4): approval-route CRUD gates on CapRouteManage at
		// tier-1, matching its tier-2 authz.Require(CapRouteManage). These rows MUST
		// precede the generic /api/v1/approval/ block, so they resolve to route.manage
		// and NOT the generic document.view (GET) / document.submit (writes).
		{name: "approval routes list", method: http.MethodGet, path: "/api/v1/approval/routes", wantCap: iamdomain.CapRouteManage, wantVisibility: iamdelivery.VisibilityPermissionGuarded},
		{name: "approval routes create", method: http.MethodPost, path: "/api/v1/approval/routes", wantCap: iamdomain.CapRouteManage, wantVisibility: iamdelivery.VisibilityPermissionGuarded},
		{name: "approval routes update", method: http.MethodPut, path: "/api/v1/approval/routes/r-1", wantCap: iamdomain.CapRouteManage, wantVisibility: iamdelivery.VisibilityPermissionGuarded},
		{name: "approval routes deactivate", method: http.MethodPost, path: "/api/v1/approval/routes/r-1/deactivate", wantCap: iamdomain.CapRouteManage, wantVisibility: iamdelivery.VisibilityPermissionGuarded},
		{name: "notifications list", method: http.MethodGet, path: "/api/v1/notifications", wantCap: iamdomain.CapNotificationRead, wantVisibility: iamdelivery.VisibilityPermissionGuarded},
		{name: "notifications unread-count", method: http.MethodGet, path: "/api/v1/notifications/unread-count", wantCap: iamdomain.CapNotificationRead, wantVisibility: iamdelivery.VisibilityPermissionGuarded},
		{name: "notifications mark-read", method: http.MethodPost, path: "/api/v1/notifications/3f2504e0-4f89-11d3-9a0c-0305e82c3301/read", wantCap: iamdomain.CapNotificationRead, wantVisibility: iamdelivery.VisibilityPermissionGuarded},
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
		{name: "feature flags public", method: http.MethodGet, path: "/api/v1/feature-flags", want: true},

		{name: "auth me not public", method: http.MethodGet, path: "/api/v1/auth/me", want: false},
		{name: "auth logout not public", method: http.MethodPost, path: "/api/v1/auth/logout", want: false},
		{name: "documents list not public", method: http.MethodGet, path: "/api/v1/documents", want: false},

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

// recordingMux implements httprouter.Muxer over a real *http.ServeMux,
// recording every registered pattern before delegating. Used by
// TestRouteCoverage to observe exactly what buildRouter mounts, without
// needing a live server.
type recordingMux struct {
	mux      *http.ServeMux
	patterns []string
}

func newRecordingMux() *recordingMux {
	return &recordingMux{mux: http.NewServeMux()}
}

func (r *recordingMux) Handle(pattern string, handler http.Handler) {
	r.patterns = append(r.patterns, pattern)
	r.mux.Handle(pattern, handler)
}

func (r *recordingMux) HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	r.patterns = append(r.patterns, pattern)
	r.mux.HandleFunc(pattern, handler)
}

func (r *recordingMux) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.mux.ServeHTTP(w, req)
}

var _ httprouter.Muxer = (*recordingMux)(nil)

// --- Nil-safe stubs for constructing every routeHandlers member for
// registration purposes only. None of these methods are ever invoked by
// buildRouter — route registration binds method values, it never calls
// them — so returning zero values is safe and these stubs exist purely to
// satisfy constructors that panic on a nil interface dependency.

type stubTokenService struct{}

func (stubTokenService) Create(context.Context, tokensapplication.CreateCommand) (*tokensdomain.Entry, error) {
	return nil, nil
}
func (stubTokenService) Get(context.Context, string, string) (*tokensdomain.Entry, error) {
	return nil, nil
}
func (stubTokenService) List(context.Context, string) ([]tokensdomain.Entry, error) { return nil, nil }
func (stubTokenService) Update(context.Context, tokensapplication.UpdateCommand) (*tokensdomain.Entry, error) {
	return nil, nil
}
func (stubTokenService) Delete(context.Context, string, string, string) error { return nil }

type stubPresenceRepository struct{}

func (stubPresenceRepository) BumpLastSeen(context.Context, string, time.Time) error { return nil }
func (stubPresenceRepository) Snapshot(context.Context, string, time.Time) ([]iampresence.Item, error) {
	return nil, nil
}
func (stubPresenceRepository) TenantsWithRecentPresence(context.Context, time.Time) ([]string, error) {
	return nil, nil
}

type stubAuditQuerier struct{}

func (stubAuditQuerier) ListEvents(context.Context, auditdomain.ListEventsQuery) ([]auditdomain.Event, bool, error) {
	return nil, false, nil
}

type stubDistributionRepository struct{}

func (stubDistributionRepository) Summary(context.Context, string, string) (int, error) {
	return 0, nil
}
func (stubDistributionRepository) Recipients(context.Context, string, string, string, int) (distributiondomain.RecipientsPage, error) {
	return distributiondomain.RecipientsPage{}, nil
}
func (stubDistributionRepository) Coverage(context.Context, string, string) ([]distributiondomain.AreaCoverageRow, error) {
	return nil, nil
}

type stubNotificationsRepository struct{}

func (stubNotificationsRepository) List(context.Context, string, string, string, string, int) (notificationsdomain.NotificationsPage, error) {
	return notificationsdomain.NotificationsPage{}, nil
}
func (stubNotificationsRepository) UnreadCount(context.Context, string, string) (int, error) {
	return 0, nil
}
func (stubNotificationsRepository) MarkRead(context.Context, string, string, string) error {
	return nil
}
func (stubNotificationsRepository) MarkAllRead(context.Context, string, string) (int, error) {
	return 0, nil
}

// buildTestRouteHandlers constructs one instance of every routeHandlers
// member with nil/stub inner dependencies — safe because buildRouter's Mount
// calls only bind method values to mux patterns, they never invoke a
// service. This exercises every mount buildRouter performs, including ones
// main() sometimes skips at runtime (e.g. the no-SQLDB boot path), so
// TestRouteCoverage gets maximal true coverage of what a route table
// omission would actually miss.
//
// presence is no longer its own routeHandlers field (Task 15a, Ruling 2):
// presenceHandler is still constructed here and passed into
// iamdelivery.NewRouter, which folds the hand-mounted stream route into its
// own Mount.
func buildTestRouteHandlers() routeHandlers {
	presenceRepo := stubPresenceRepository{}
	presenceHub := iampresence.NewHub(presenceRepo, nil)
	presenceHandler := iampresence.NewHandler(presenceHub, presenceRepo, nil)

	iamRouter := iamdelivery.NewRouter(
		iamdelivery.NewAdminHandler(nil, nil),
		iamdelivery.NewPeopleHandler(nil, nil, nil),
		iamdelivery.NewMembershipHandler(nil, nil),
		iamdelivery.NewRolesCapsHandler(nil),
		iamdelivery.NewSessionsHandler(nil),
		iamdelivery.NewObservabilityHandler(nil),
		presenceHandler,
	)

	docMod := &documents.Module{Handler: dhttp.NewHandler(nil)}
	docMod.WithRateLimit(nil, func(*http.Request) string { return "" })

	return routeHandlers{
		auth:          authdelivery.NewHandler(nil),
		health:        observability.NewHealthHandler(nil),
		observability: observability.NewMetricsHandler(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {})),
		featureFlags:  featureflags.NewHandler(config.FeatureFlagsConfig{}),
		audit:         auditdelivery.NewHandler(stubAuditQuerier{}),
		search:        searchdelivery.NewHandler(nil),
		security:      securitydelivery.NewHandler(nil),
		taxonomy:      &taxonomy.Module{Handler: thttp.NewHandler(nil, nil, nil, nil)},
		tokens:        &tokens.Module{Handler: tokenshttp.NewHandler(stubTokenService{})},
		controlledDocuments: &controlleddocuments.Module{
			Handler: cddhttp.NewHandler(nil, nil),
		},
		iamRouter:     iamRouter,
		documents:     docMod,
		templates:     templateshttp.New(nil, func(*http.Request, string, string, string) error { return nil }, nil),
		approval:      approvalhttp.NewHandler(nil, nil, nil, nil),
		distribution:  distributionhttp.NewHandler(stubDistributionRepository{}),
		notifications: notificationshttp.NewHandler(stubNotificationsRepository{}),
	}
}

// routeHandlerFields returns every routeHandlers field that names a route
// family, derived from the struct's own field names via reflection — NOT a
// hand-typed list. routeFamily.name is a string literal in router.go's mount
// table; without this cross-check it could be misspelled or a whole entry
// omitted, silently reintroducing the hand-sync-drift class this test exists
// to eliminate. Anchoring the expected set to reflect.Type field names means
// it cannot drift: it comes from routeHandlers' actual fields at compile
// time, not from a second, independently-typed enumeration.
//
// Deliberately TYPE-driven, never value-driven: it takes no routeHandlers
// instance and inspects no field's value. An earlier version skipped fields
// that were nil, which let a newly added family launder itself out of the
// expected set simply by being forgotten in buildTestRouteHandlers too — the
// exact omission this guard exists to catch. Mount is total (§4): every field
// returned here MUST have an entry in router.go's routeFamilies table, with
// no exemption set to subtract against.
//
// Every routeHandlers field is a route family (Task 15a): the former
// documentsRateLimit/documentsUserID config params moved onto *documents.
// Module as constructor fields (WithRateLimit, wired in main.go before the
// module is placed in routeHandlers), so there is no longer a non-family
// field to exclude.
func routeHandlerFields() map[string]bool {
	fields := map[string]bool{}
	t := reflect.TypeOf(routeHandlers{})
	for i := 0; i < t.NumField(); i++ {
		fields[t.Field(i).Name] = true
	}
	return fields
}

// TestRouteCoverage builds a real routeHandlers instance and mounts every
// family in router.go's routeFamilies table onto a recordingMux, then asserts
// every recorded pattern resolves to a routeRules entry — and separately that
// buildRouter (the exact function main() calls) registers that same pattern
// set. This replaces a hand-maintained fixture list of representative routes
// (which could silently drift from what main() actually mounts) with a
// structural guard: a handler that's registered here but has no matching
// routeRules entry is a build-passing-but-red-test event, not a live
// discovery.
//
// Method-qualified patterns ("GET /path", the oapi-codegen convention) are
// checked strictly via matchedByRule(method, path). Bare legacy patterns
// (no method prefix — auth, health, featureFlags, search, security; each
// still hand-registered and internally switching on r.Method, EXCEPT health,
// which has no method switch at all, see health.go) only support a looser
// check: some routeRules entry matches the path for some method. That's the
// strongest sound claim without inspecting each handler's internal dispatch
// — TestPermissionsTable_NoMethodlessWriteShadowing and friends guard
// routeRules' own shape, not this gap; they don't observe a registered
// pattern at all. (presence is NOT in this set: its stream endpoint
// registers method-qualified via presence.MountStream, called from
// iamRouter's Mount, so it is strict-checked.)
//
// TRANSITIONAL LOCAL MAXIMUM (CLAUDE.md "Global Maximum, Not Local
// Maximum"): the global-maximum structure is finishing the CON-07/ARC-02
// codegen migration for those five handlers, so every one of their patterns
// becomes method-qualified by construction and this loose check is deleted
// in favor of the strict one used for every other family. That migration is
// NOT YET SCHEDULED to a milestone — flagging that gap to the operator is
// itself part of shipping this local maximum honestly, rather than picking
// a milestone unilaterally.
func TestRouteCoverage(t *testing.T) {
	rec := newRecordingMux()
	h := buildTestRouteHandlers()

	// Walk router.go's mount table directly — buildRouter's entire body is a
	// loop over it, so this is the same registration production performs,
	// with per-family visibility production doesn't need. Three assertions
	// the aggregate pattern count below cannot make:
	//   - a family whose name isn't a routeHandlers field (typo in the table),
	//   - a duplicate table entry,
	//   - a family registering zero patterns (e.g. a miswired stub here),
	//     which would otherwise hide behind every other family registering
	//     fine.
	fields := routeHandlerFields()
	seen := map[string]bool{}
	for _, f := range routeFamilies(h) {
		if !fields[f.name] {
			t.Errorf("routeFamilies returned %q, which is not a routeHandlers field — typo in router.go's mount table?", f.name)
		}
		if seen[f.name] {
			t.Errorf("routeFamilies returned %q twice — duplicate entry in router.go's mount table", f.name)
		}
		seen[f.name] = true

		before := len(rec.patterns)
		f.register(rec)
		if len(rec.patterns) == before {
			t.Errorf("route family %q registered zero patterns — its handler construction in buildTestRouteHandlers (or its RegisterRoutes) is broken", f.name)
		}
	}
	for family := range fields {
		if !seen[family] {
			t.Errorf("routeHandlers field %q has no entry in router.go's routeFamilies mount table — it is constructed but never mounted", family)
		}
	}

	if len(rec.patterns) == 0 {
		t.Fatal("routeFamilies registered zero patterns onto recordingMux — the mount table or its handler construction is broken")
	}

	// Fidelity: buildRouter is production's sole mount path (main.go), and its
	// body is supposed to be nothing but a loop over routeFamilies. Assert
	// that rather than trusting it — otherwise a statement added to
	// buildRouter outside the loop, or a loop that stopped registering, would
	// be invisible to this test even though the walk above stayed green.
	viaBuildRouter := newRecordingMux()
	buildRouter(viaBuildRouter, buildTestRouteHandlers())
	if !reflect.DeepEqual(viaBuildRouter.patterns, rec.patterns) {
		t.Errorf("buildRouter's registered patterns diverge from a direct walk of routeFamilies — buildRouter must be exactly a loop over the mount table\nbuildRouter:   %v\nrouteFamilies: %v", viaBuildRouter.patterns, rec.patterns)
	}

	resolver := newPermissionResolver()

	for _, pattern := range rec.patterns {
		pattern := pattern
		t.Run(pattern, func(t *testing.T) {
			if method, path, ok := strings.Cut(pattern, " "); ok && isHTTPMethod(method) {
				if !matchedByRule(method, path) {
					_, visibility := resolver(method, path)
					t.Fatalf("registered route %s %s fell through to default visibility=%v — add an entry to routeRules",
						method, path, visibility)
				}
				return
			}

			// Bare legacy pattern: assert some rule matches this path for some
			// method (loose check — see doc comment above).
			path := pattern
			matched := false
			for _, m := range []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete} {
				if matchedByRule(m, path) {
					matched = true
					break
				}
			}
			if !matched {
				t.Fatalf("registered legacy route %q is matched by no routeRules entry for any method — add an entry to routeRules", path)
			}
		})
	}
}

func isHTTPMethod(s string) bool {
	switch s {
	case http.MethodGet, http.MethodHead, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodConnect, http.MethodOptions, http.MethodTrace:
		return true
	default:
		return false
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
//
//	(a) has empty method AND a Manage/Submit/write-grade cap, OR
//	(b) has empty method on a prefix where some OTHER rule declares a write
//	    verb (so the methodless row would shadow per-verb intent).
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

// TestPermissionResolver_AreaMembershipRoutes pins the PR-1 contract: GET
// resolves to membership.view, POST + DELETE to membership.manage, and PUT +
// PATCH are NOT on the surface (the API offers grant + revoke only — role
// changes go via revoke-then-grant). Without this lock test the rules at
// permissions.go:225-233 could regress to a path-prefix entry that silently
// admits PUT/PATCH.
func TestPermissionResolver_AreaMembershipRoutes(t *testing.T) {
	t.Parallel()

	resolver := newPermissionResolver()
	cases := []struct {
		method string
		path   string
		cap    iamdomain.Capability
	}{
		{http.MethodGet, "/api/v1/iam/area-memberships", iamdomain.CapMembershipView},
		{http.MethodPost, "/api/v1/iam/area-memberships", iamdomain.CapMembershipManage},
		{http.MethodDelete, "/api/v1/iam/area-memberships/u-1/quality", iamdomain.CapMembershipManage},
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

	// PUT + PATCH must NOT be permission-guarded on this path — they're not
	// part of the surface (the API offers grant + revoke only). A future
	// regression that resolves them to *any* permission-guarded cap (manage,
	// view, or anything else) silently expands the attack surface; the only
	// safe verdict is fall-through to VisibilitySessionRequired so the
	// ServeMux 405 handles them.
	for _, method := range []string{http.MethodPut, http.MethodPatch} {
		gotCap, gotVis := resolver(method, "/api/v1/iam/area-memberships")
		if gotVis == iamdelivery.VisibilityPermissionGuarded {
			t.Errorf("%s /api/v1/iam/area-memberships: visibility=PermissionGuarded cap=%q — PUT/PATCH must fall through to VisibilitySessionRequired", method, gotCap)
		}
	}
}

// TestDocumentsRoutesCapabilityMapping pins the ADR 0022 Phase 9 documents
// tier-1 contract: every documents route resolves to a capability +
// VisibilityPermissionGuarded via permissions.go alone. This is the
// production-defect-fixed proof — the per-verb /duplicate, /comments POST and
// /comments/ DELETE rows close the gaps that previously fell through to the
// (now-deleted) handler role gate.
func TestDocumentsRoutesCapabilityMapping(t *testing.T) {
	t.Parallel()

	resolver := newPermissionResolver()
	cases := []struct {
		method string
		path   string
		cap    iamdomain.Capability
	}{
		{http.MethodGet, "/api/v1/documents", iamdomain.CapDocumentView},
		{http.MethodPost, "/api/v1/documents/d1/session/acquire", iamdomain.CapDocumentEdit},
		{http.MethodPost, "/api/v1/documents/d1/session/force-release", iamdomain.CapMembershipManage},
		{http.MethodPost, "/api/v1/documents/d1/duplicate", iamdomain.CapDocumentCreate},
		{http.MethodPost, "/api/v1/documents/d1/comments", iamdomain.CapDocumentEdit},
		{http.MethodDelete, "/api/v1/documents/d1/comments/5", iamdomain.CapDocumentEdit},
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

// TestEveryRouteCapInRegistry binds the Tier-1 route table to the Go capability
// registry (single source of truth, ADR 0022 Phase 1). Every non-empty
// routeRules[i].capability MUST be a typed const present in validCapabilities.
// A raw iamdomain.Capability("typo") row — which compiles clean — fails here.
func TestEveryRouteCapInRegistry(t *testing.T) {
	t.Parallel()

	for i, r := range routeRules {
		if r.capability == "" {
			continue // public / session-required rows carry no cap
		}
		if !iamdomain.IsValidCapability(r.capability) {
			t.Errorf("routeRules[%d]: capability %q is not in the registry (validCapabilities). Promote it to a typed const in internal/modules/iam/domain/model.go (path prefix=%q exact=%q suffix=%q).",
				i, r.capability, r.pathPrefix, r.pathExact, r.pathSuffix)
		}
	}
}

// seededCaps reads the canonical role_capabilities seed and returns the set of
// distinct capability values granted to ≥1 role. The seed is the DB source of
// truth bound here so a registry cap that is never seeded (and never bypass-only
// by design) becomes a red build instead of a silent always-403 capability.
func seededCaps(t *testing.T) map[iamdomain.Capability]struct{} {
	t.Helper()

	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed; cannot locate seed file")
	}
	// apps/api/cmd/metaldocs-api/ → repo root is four levels up.
	seedPath := filepath.Join(filepath.Dir(thisFile), "..", "..", "..", "..",
		"db", "reference-data", "0001_product_reference_data.sql")
	raw, err := os.ReadFile(seedPath)
	if err != nil {
		t.Fatalf("read seed %s: %v", seedPath, err)
	}

	// Match: INSERT INTO metaldocs.role_capabilities (...) VALUES ('role', 'cap', ...
	re := regexp.MustCompile(`(?i)INSERT INTO\s+metaldocs\.role_capabilities[^V]*VALUES\s*\(\s*'[^']*'\s*,\s*'([^']+)'`)
	matches := re.FindAllStringSubmatch(string(raw), -1)
	if len(matches) == 0 {
		t.Fatalf("no role_capabilities INSERT rows parsed from %s", seedPath)
	}
	set := make(map[iamdomain.Capability]struct{}, len(matches))
	for _, m := range matches {
		set[iamdomain.Capability(m[1])] = struct{}{}
	}
	return set
}

// seededRoleCaps reads the canonical role_capabilities seed and returns a
// role -> set-of-capabilities map. Used by the deny-by-default seed guard.
func seededRoleCaps(t *testing.T) map[string]map[iamdomain.Capability]struct{} {
	t.Helper()

	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed; cannot locate seed file")
	}
	seedPath := filepath.Join(filepath.Dir(thisFile), "..", "..", "..", "..",
		"db", "reference-data", "0001_product_reference_data.sql")
	raw, err := os.ReadFile(seedPath)
	if err != nil {
		t.Fatalf("read seed %s: %v", seedPath, err)
	}

	// Match: INSERT INTO metaldocs.role_capabilities (...) VALUES ('role', 'cap', ...
	re := regexp.MustCompile(`(?i)INSERT INTO\s+metaldocs\.role_capabilities[^V]*VALUES\s*\(\s*'([^']*)'\s*,\s*'([^']+)'`)
	matches := re.FindAllStringSubmatch(string(raw), -1)
	if len(matches) == 0 {
		t.Fatalf("no role_capabilities INSERT rows parsed from %s", seedPath)
	}
	out := map[string]map[iamdomain.Capability]struct{}{}
	for _, m := range matches {
		role := m[1]
		if out[role] == nil {
			out[role] = map[iamdomain.Capability]struct{}{}
		}
		out[role][iamdomain.Capability(m[2])] = struct{}{}
	}
	return out
}

// TestDenyByDefault_ViewerHoldsNoAreaGradeCapability (ADR 0022 Phase 11 F8) is the
// seed half of the deny-by-default guard. The least-privileged real role, viewer,
// must never be seeded an area-grade (write / area-scoped) capability — only
// tenant-grade reads. A future seed that accidentally grants viewer an area-grade
// cap (e.g. document.edit) is a red build. Complements the authz-layer deny matrix
// (TestDenyByDefault_EveryCapabilityDeniesUnprivilegedActor) which locks the
// enforcement code; this locks the grant data.
func TestDenyByDefault_ViewerHoldsNoAreaGradeCapability(t *testing.T) {
	t.Parallel()

	roleCaps := seededRoleCaps(t)
	viewerCaps, ok := roleCaps["viewer"]
	if !ok {
		t.Fatal("viewer role not present in role_capabilities seed")
	}
	for cap := range viewerCaps {
		if iamdomain.IsAreaGrade(cap) {
			t.Errorf("viewer is seeded area-grade capability %q; the least-privileged role must hold only tenant-grade reads (deny-by-default, ADR 0022 Phase 11 F8)", cap)
		}
	}
}

// TestEveryCapSeededOrDeferred asserts every registry capability is either
// seeded to ≥1 role OR is an explicitly-deferred write cap enforced today only
// via the system_admin tier-2 bypass (routed in permissions.go, granted to no
// tenant role yet). ADR 0022 Phase 1. A new registry cap that is neither seeded
// nor allow-listed fails — preventing a capability that nobody (except
// system_admin) can ever satisfy from slipping in unnoticed.
func TestEveryCapSeededOrDeferred(t *testing.T) {
	t.Parallel()

	// Routed-but-unseeded write caps enforced only by the system_admin bypass.
	// ADR 0022 Phase 1 deferred doc.publish/doc.obsolete/doc.supersede/template.archive
	// here; Phase 2 (migration 0225 + reference-data mirror) seeds all four to
	// real tenant roles. A future routed-but-unseeded cap goes here with a
	// documented deferral. Mirrors deferredCaps in scripts/api-lint/registry_rules.go.
	//
	// distribution.read — minted in mission frontend-screen-completion M2/F2.1c
	// (ADR-0042). Deliberately NOT seeded to any tenant role by the agent —
	// the operator grants it to roles separately.
	deferred := map[iamdomain.Capability]struct{}{
		iamdomain.CapDistributionRead: {},
		iamdomain.CapNotificationRead: {},
	}

	seeded := seededCaps(t)

	caps := iamdomain.AllCapabilities()
	sort.Slice(caps, func(i, j int) bool { return caps[i] < caps[j] }) // map order is random

	for _, c := range caps {
		_, isSeeded := seeded[c]
		_, isDeferred := deferred[c]
		if !isSeeded && !isDeferred {
			t.Errorf("capability %q is in the registry but seeded to no role and not in the deferred allow-list. Seed it in db/reference-data/0001_product_reference_data.sql or document the deferral.", c)
		}
	}
}

// TestTier1Tier2CapabilityCoherence_F4Sites pins the ADR 0022 Phase 11 (F4)
// coherence fix for the two tier-1/tier-2 capability sites that were
// historically divergent (see ADR 0022 §349-351, "Amendment — Phase 1-7 final
// review" open follow-ups). This is a NARROW regression lock on exactly these
// two named routes — it must NOT be broadened into a blanket "tier-1 == tier-2
// for all /approval/ routes" assertion: the generic /approval/ prefix
// (document.submit at tier-1) legitimately diverges from the fine-grained
// tier-2 checks inside individual approval operations (e.g. document.signoff)
// by design (coarse route gate vs. fine in-tx business capability), and that
// divergence is NOT a defect.
//
// Tier-2 truth for each site (hand-verified against source, not re-derived):
//   - force-release: internal/modules/documents/repository/repository.go
//     ForceReleaseSession/ForceReleaseSessionTx call
//     authz.Require(ctx, tx, string(iamdomain.CapMembershipManage), docArea).
//   - approval-route management: internal/modules/approval/application/
//     route_admin_service.go calls
//     authz.Require(ctx, tx, string(iamdomain.CapRouteManage), "tenant").
//
// This test only pins the tier-1 side (this package cannot import the
// tier-2 packages without an import cycle risk); the tier-2 capability
// constants above are asserted by literal reference to the same typed
// iamdomain consts the tier-2 call sites use, so any rename of either const
// is a compile error here, and any future rename of ONLY the tier-1 route
// row's capability (leaving tier-2 unchanged) fails this test at runtime.
func TestTier1Tier2CapabilityCoherence_F4Sites(t *testing.T) {
	t.Parallel()

	resolver := newPermissionResolver()

	// Tier-2 truth, pinned by direct reference to the same typed consts the
	// tier-2 call sites use (repository.go / route_admin_service.go).
	const tier2ForceRelease = iamdomain.CapMembershipManage
	const tier2ApprovalRouteManage = iamdomain.CapRouteManage

	cases := []struct {
		name      string
		method    string
		path      string
		wantTier2 iamdomain.Capability
	}{
		{
			name:      "force-release session: tier-1 must equal tier-2 CapMembershipManage",
			method:    http.MethodPost,
			path:      "/api/v1/documents/d1/session/force-release",
			wantTier2: tier2ForceRelease,
		},
		{
			name:      "approval routes create: tier-1 must equal tier-2 CapRouteManage",
			method:    http.MethodPost,
			path:      "/api/v1/approval/routes",
			wantTier2: tier2ApprovalRouteManage,
		},
	}

	for _, tc := range cases {
		gotCap, gotVis := resolver(tc.method, tc.path)
		if gotVis != iamdelivery.VisibilityPermissionGuarded {
			t.Errorf("%s: %s %s: visibility=%v want PermissionGuarded", tc.name, tc.method, tc.path, gotVis)
		}
		if gotCap != tc.wantTier2 {
			t.Errorf("%s: %s %s: tier-1 cap=%q diverges from tier-2 cap=%q (ADR 0022 F4 regression)", tc.name, tc.method, tc.path, gotCap, tc.wantTier2)
		}
	}
}
