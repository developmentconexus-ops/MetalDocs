package main

import (
	"net/http"
	"strings"

	authdelivery "metaldocs/internal/modules/auth/delivery/http"
	iamdelivery "metaldocs/internal/modules/iam/delivery/http"
	iamdomain "metaldocs/internal/modules/iam/domain"
)

// routeRule matches an HTTP request to a capability + visibility classification.
// All non-empty fields are AND-ed; an empty field is ignored.
//
//	method     — exact HTTP method match (empty = any method)
//	pathExact  — full path equality match
//	pathPrefix — strings.HasPrefix
//	pathSuffix — strings.HasSuffix
//	contains   — strings.Contains
//	notSuffix  — strings.HasSuffix that DISQUALIFIES the rule (used for negative guards)
//
// Rules are scanned in order; the first match wins. Any route not matched by
// any rule falls through to VisibilitySessionRequired — never public.
type routeRule struct {
	method     string
	pathExact  string
	pathPrefix string
	pathSuffix string
	contains   string
	notSuffix  string
	capability iamdomain.Capability
	visibility iamdelivery.Visibility
}

func (r routeRule) matches(method, path string) bool {
	if r.method != "" && r.method != method {
		return false
	}
	if r.pathExact != "" && r.pathExact != path {
		return false
	}
	if r.pathPrefix != "" && !strings.HasPrefix(path, r.pathPrefix) {
		return false
	}
	if r.pathSuffix != "" && !strings.HasSuffix(path, r.pathSuffix) {
		return false
	}
	if r.contains != "" && !strings.Contains(path, r.contains) {
		return false
	}
	if r.notSuffix != "" && strings.HasSuffix(path, r.notSuffix) {
		return false
	}
	return true
}

// routeRules is the single authoritative table for route visibility +
// capability. This intentionally couples startup route registration and authz
// classification until the app has generated route metadata. New routes MUST be
// added here; the unmatched default is VisibilitySessionRequired, so a
// forgotten entry produces a 401 (loud), not a silent public route
// (catastrophic).
//
// F-001 read/write split truth table (audit row -> intended caps):
//
//	metrics           GET   /api/v1/metrics              -> CapMetricsView
//	iam/users         GET   /api/v1/iam/users            -> CapUserView
//	iam/users         POST/PATCH/PUT (writes)            -> CapUserManage
//	iam/admin/overview GET  /api/v1/iam/admin/overview   -> CapUserView
//	taxonomy/profiles GET                                -> CapTaxonomyView
//	taxonomy/profiles POST/PATCH/PUT/DELETE              -> CapTaxonomyManage
//	taxonomy/areas    GET                                -> CapTaxonomyView
//	taxonomy/areas    POST/PATCH/PUT/DELETE              -> CapTaxonomyManage
//	taxonomy/families GET                                -> CapTaxonomyView
//	taxonomy/families POST/PATCH/PUT/DELETE              -> CapTaxonomyManage
//	iam/area-memberships GET                             -> CapMembershipView
//	iam/area-memberships POST/PUT/PATCH/DELETE           -> CapMembershipManage
//
// Authoring rule: every writable prefix MUST have an explicit GET row with a
// View-grade cap; methodless rows on writable prefixes are forbidden.
// TestPermissionsTable_NoMethodlessWriteShadowing enforces this in CI.
var routeRules = []routeRule{
	// ---- Public (no session required) -----------------------------------
	{method: http.MethodGet, pathPrefix: "/api/v1/health/", visibility: iamdelivery.VisibilityPublic},
	{pathExact: "/healthz", visibility: iamdelivery.VisibilityPublic},
	{method: http.MethodPost, pathExact: "/api/v1/auth/login", visibility: iamdelivery.VisibilityPublic},
	{method: http.MethodGet, pathExact: "/api/v1/feature-flags", visibility: iamdelivery.VisibilityPublic},

	// ---- Session-required (authenticated, no specific capability) -------
	{method: http.MethodGet, pathExact: "/api/v1/auth/me", visibility: iamdelivery.VisibilitySessionRequired},
	{method: http.MethodPost, pathExact: "/api/v1/auth/change-password", visibility: iamdelivery.VisibilitySessionRequired},
	{method: http.MethodPost, pathExact: "/api/v1/auth/logout", visibility: iamdelivery.VisibilitySessionRequired},

	// ---- Permission-guarded ---------------------------------------------
	// Metrics — JSON scrape endpoint; read-only.
	{method: http.MethodGet, pathExact: "/api/v1/metrics", capability: iamdomain.CapMetricsView, visibility: iamdelivery.VisibilityPermissionGuarded},

	// Search.
	{method: http.MethodGet, pathExact: "/api/v1/search/documents", capability: iamdomain.CapDocumentView, visibility: iamdelivery.VisibilityPermissionGuarded},

	// Tenant onboarding (M7 F7.2): provisioning a new tenant is a tenant-wide
	// operator action, gated on tenant.onboard — never a role check (ADR 0022).
	{method: http.MethodPost, pathExact: "/api/v1/tenants", capability: iamdomain.CapTenantOnboard, visibility: iamdelivery.VisibilityPermissionGuarded},

	// Tenant export/erase (M7 F7.3, ADR 0070): parameterized {tenantId} lifecycle
	// actions. tenant.erase is seeded system_admin-ONLY (no other role) per spec.
	{method: http.MethodPost, pathPrefix: "/api/v1/tenants/", pathSuffix: "/export", capability: iamdomain.CapTenantExport, visibility: iamdelivery.VisibilityPermissionGuarded},
	{method: http.MethodPost, pathPrefix: "/api/v1/tenants/", pathSuffix: "/erase", capability: iamdomain.CapTenantErase, visibility: iamdelivery.VisibilityPermissionGuarded},

	// IAM users. F-001 split: GET view, writes manage.
	{method: http.MethodGet, pathExact: "/api/v1/iam/users", capability: iamdomain.CapUserView, visibility: iamdelivery.VisibilityPermissionGuarded},
	{method: http.MethodPost, pathExact: "/api/v1/iam/users", capability: iamdomain.CapUserManage, visibility: iamdelivery.VisibilityPermissionGuarded},
	// PR-4 PeopleHandler routes — exact path match first so the PATCH /users/
	// prefix rule below cannot shadow them. Without these explicit entries the
	// fallback returns VisibilitySessionRequired and any authenticated user
	// (Viewer included) can invite / bulk-mutate / list memberships.
	{method: http.MethodPost, pathExact: "/api/v1/iam/users/invite", capability: iamdomain.CapUserManage, visibility: iamdelivery.VisibilityPermissionGuarded},
	{method: http.MethodPost, pathExact: "/api/v1/iam/users/bulk", capability: iamdomain.CapUserManage, visibility: iamdelivery.VisibilityPermissionGuarded},
	{method: http.MethodGet, pathPrefix: "/api/v1/iam/users/", pathSuffix: "/memberships", capability: iamdomain.CapMembershipView, visibility: iamdelivery.VisibilityPermissionGuarded},
	{method: http.MethodPatch, pathPrefix: "/api/v1/iam/users/", notSuffix: "/roles", capability: iamdomain.CapUserManage, visibility: iamdelivery.VisibilityPermissionGuarded},
	{method: http.MethodPost, pathPrefix: "/api/v1/iam/users/", pathSuffix: "/roles", capability: iamdomain.CapUserManage, visibility: iamdelivery.VisibilityPermissionGuarded},
	{method: http.MethodPut, pathPrefix: "/api/v1/iam/users/", pathSuffix: "/roles", capability: iamdomain.CapUserManage, visibility: iamdelivery.VisibilityPermissionGuarded},
	{method: http.MethodPost, pathPrefix: "/api/v1/iam/users/", pathSuffix: "/reset-password", capability: iamdomain.CapUserManage, visibility: iamdelivery.VisibilityPermissionGuarded},
	{method: http.MethodPost, pathPrefix: "/api/v1/iam/users/", pathSuffix: "/unlock", capability: iamdomain.CapUserManage, visibility: iamdelivery.VisibilityPermissionGuarded},
	{method: http.MethodGet, pathExact: "/api/v1/iam/admin/overview", capability: iamdomain.CapUserView, visibility: iamdelivery.VisibilityPermissionGuarded},

	// IAM Admin Center Usage + KPI cards (PR-8). Tenant-scoped read-only
	// observability — guarded by CapMetricsView (separate from the Overview
	// composition which retains CapUserView so a viewer can see the tab).
	{method: http.MethodGet, pathExact: "/api/v1/iam/usage", capability: iamdomain.CapMetricsView, visibility: iamdelivery.VisibilityPermissionGuarded},
	{method: http.MethodGet, pathExact: "/api/v1/iam/kpi", capability: iamdomain.CapMetricsView, visibility: iamdelivery.VisibilityPermissionGuarded},

	// Templates — GET first, then exact POST, then sub-route POST/PUTs.
	{method: http.MethodGet, pathPrefix: "/api/v1/templates", capability: iamdomain.CapTemplateView, visibility: iamdelivery.VisibilityPermissionGuarded},
	{method: http.MethodPost, pathExact: "/api/v1/templates", capability: iamdomain.CapTemplateCreate, visibility: iamdelivery.VisibilityPermissionGuarded},
	{method: http.MethodPost, pathPrefix: "/api/v1/templates", pathSuffix: "/versions", capability: iamdomain.CapTemplateCreate, visibility: iamdelivery.VisibilityPermissionGuarded},
	{method: http.MethodPut, pathPrefix: "/api/v1/templates", pathSuffix: "/draft", capability: iamdomain.CapTemplateEdit, visibility: iamdelivery.VisibilityPermissionGuarded},
	{method: http.MethodPut, pathPrefix: "/api/v1/templates", pathSuffix: "/schema", capability: iamdomain.CapTemplateEdit, visibility: iamdelivery.VisibilityPermissionGuarded},
	{method: http.MethodPost, pathPrefix: "/api/v1/templates", pathSuffix: "/publish", capability: iamdomain.CapTemplatePublish, visibility: iamdelivery.VisibilityPermissionGuarded},
	{method: http.MethodPost, pathPrefix: "/api/v1/templates", pathSuffix: "/docx-upload-url", capability: iamdomain.CapTemplateEdit, visibility: iamdelivery.VisibilityPermissionGuarded},
	{method: http.MethodPost, pathPrefix: "/api/v1/templates", pathSuffix: "/schema-upload-url", capability: iamdomain.CapTemplateEdit, visibility: iamdelivery.VisibilityPermissionGuarded},
	{method: http.MethodPost, pathPrefix: "/api/v1/templates", contains: "/autosave/", capability: iamdomain.CapTemplateEdit, visibility: iamdelivery.VisibilityPermissionGuarded},
	{method: http.MethodPost, pathPrefix: "/api/v1/templates", pathSuffix: "/submit", capability: iamdomain.CapTemplateSubmit, visibility: iamdelivery.VisibilityPermissionGuarded},
	{method: http.MethodPost, pathPrefix: "/api/v1/templates", pathSuffix: "/review", capability: iamdomain.CapTemplateReview, visibility: iamdelivery.VisibilityPermissionGuarded},
	{method: http.MethodPost, pathPrefix: "/api/v1/templates", pathSuffix: "/approve", capability: iamdomain.CapTemplateApprove, visibility: iamdelivery.VisibilityPermissionGuarded},
	{method: http.MethodPut, pathPrefix: "/api/v1/templates", pathSuffix: "/approval-config", capability: iamdomain.CapTemplateEdit, visibility: iamdelivery.VisibilityPermissionGuarded},
	{method: http.MethodPost, pathPrefix: "/api/v1/templates", pathSuffix: "/archive", capability: iamdomain.CapTemplateArchive, visibility: iamdelivery.VisibilityPermissionGuarded},

	// Distribution — GET-only endpoints; must precede the generic documents GET catch-all
	// so CapDistributionRead is enforced at tier-1 (mirrors CapAuditRead precedent).
	// Three suffix variants in specificity order (most specific first).
	{method: http.MethodGet, pathPrefix: "/api/v1/documents", pathSuffix: "/distribution/recipients", capability: iamdomain.CapDistributionRead, visibility: iamdelivery.VisibilityPermissionGuarded},
	{method: http.MethodGet, pathPrefix: "/api/v1/documents", pathSuffix: "/distribution/coverage", capability: iamdomain.CapDistributionRead, visibility: iamdelivery.VisibilityPermissionGuarded},
	{method: http.MethodGet, pathPrefix: "/api/v1/documents", pathSuffix: "/distribution", capability: iamdomain.CapDistributionRead, visibility: iamdelivery.VisibilityPermissionGuarded},

	// Documents — order preserves the original switch semantics (more specific suffixes/contains first).
	{method: http.MethodGet, pathPrefix: "/api/v1/documents", capability: iamdomain.CapDocumentView, visibility: iamdelivery.VisibilityPermissionGuarded},
	// CON-01 (grade-A simplification, DEC-01): finalize is a deprecated convenience
	// wrapper over the same draft->under_review transition as /submit — it must gate
	// on the same tier-1 capability as submit, not signoff (finalize never records a
	// signoff; the prior CapDocumentSignoff gate contradicted capability semantics).
	{method: http.MethodPost, pathPrefix: "/api/v1/documents", pathSuffix: "/finalize", capability: iamdomain.CapDocumentSubmit, visibility: iamdelivery.VisibilityPermissionGuarded},
	{method: http.MethodPost, pathPrefix: "/api/v1/documents", pathSuffix: "/archive", capability: iamdomain.CapDocumentEdit, visibility: iamdelivery.VisibilityPermissionGuarded},
	{method: http.MethodPost, pathPrefix: "/api/v1/documents", pathSuffix: "/duplicate", capability: iamdomain.CapDocumentCreate, visibility: iamdelivery.VisibilityPermissionGuarded},
	{method: http.MethodPost, pathPrefix: "/api/v1/documents", pathSuffix: "/comments", capability: iamdomain.CapDocumentEdit, visibility: iamdelivery.VisibilityPermissionGuarded},
	{method: http.MethodPost, pathPrefix: "/api/v1/documents", contains: "/session/force-release", capability: iamdomain.CapMembershipManage, visibility: iamdelivery.VisibilityPermissionGuarded},
	{method: http.MethodPost, pathPrefix: "/api/v1/documents", contains: "/session/", capability: iamdomain.CapDocumentEdit, visibility: iamdelivery.VisibilityPermissionGuarded},
	{method: http.MethodPost, pathPrefix: "/api/v1/documents", contains: "/autosave/", capability: iamdomain.CapDocumentEdit, visibility: iamdelivery.VisibilityPermissionGuarded},
	{method: http.MethodPost, pathPrefix: "/api/v1/documents", contains: "/checkpoints/", pathSuffix: "/restore", capability: iamdomain.CapDocumentEdit, visibility: iamdelivery.VisibilityPermissionGuarded},
	{method: http.MethodPost, pathPrefix: "/api/v1/documents", contains: "/checkpoints", capability: iamdomain.CapDocumentEdit, visibility: iamdelivery.VisibilityPermissionGuarded},
	{method: http.MethodPost, pathPrefix: "/api/v1/documents", pathSuffix: "/export/pdf", capability: iamdomain.CapDocumentView, visibility: iamdelivery.VisibilityPermissionGuarded},
	{method: http.MethodPut, pathPrefix: "/api/v1/documents", contains: "/placeholders/", capability: iamdomain.CapDocumentEdit, visibility: iamdelivery.VisibilityPermissionGuarded},
	{method: http.MethodPatch, pathPrefix: "/api/v1/documents", capability: iamdomain.CapDocumentEdit, visibility: iamdelivery.VisibilityPermissionGuarded},
	{method: http.MethodDelete, pathPrefix: "/api/v1/documents", contains: "/comments/", capability: iamdomain.CapDocumentEdit, visibility: iamdelivery.VisibilityPermissionGuarded},
	{method: http.MethodPost, pathPrefix: "/api/v1/documents", pathSuffix: "/submit", capability: iamdomain.CapDocumentSubmit, visibility: iamdelivery.VisibilityPermissionGuarded},
	{method: http.MethodPost, pathPrefix: "/api/v1/documents", pathSuffix: "/signoff", capability: iamdomain.CapDocumentSignoff, visibility: iamdelivery.VisibilityPermissionGuarded},
	{method: http.MethodPost, pathPrefix: "/api/v1/documents", pathSuffix: "/publish", capability: iamdomain.CapDocumentPublish, visibility: iamdelivery.VisibilityPermissionGuarded},
	{method: http.MethodPost, pathPrefix: "/api/v1/documents", pathSuffix: "/schedule-publish", capability: iamdomain.CapDocumentPublish, visibility: iamdelivery.VisibilityPermissionGuarded},
	// M6 F6.2 (markDocumentReviewed): the eQMS mark-reviewed governance action —
	// tier-1 gates document.review, tier-2 authz.Require lands in the T5 service.
	{method: http.MethodPost, pathPrefix: "/api/v1/documents", pathSuffix: "/review", capability: iamdomain.CapDocumentReview, visibility: iamdelivery.VisibilityPermissionGuarded},
	{method: http.MethodPost, pathPrefix: "/api/v1/documents", pathSuffix: "/supersede", capability: iamdomain.CapDocumentSupersede, visibility: iamdelivery.VisibilityPermissionGuarded},
	{method: http.MethodPost, pathPrefix: "/api/v1/documents", pathSuffix: "/obsolete", capability: iamdomain.CapDocumentObsolete, visibility: iamdelivery.VisibilityPermissionGuarded},
	{method: http.MethodPost, pathPrefix: "/api/v1/documents", pathSuffix: "/cancel", capability: iamdomain.CapDocumentEdit, visibility: iamdelivery.VisibilityPermissionGuarded},
	{method: http.MethodPost, pathPrefix: "/api/v1/documents", pathSuffix: "/reconstruct", capability: iamdomain.CapDocumentEdit, visibility: iamdelivery.VisibilityPermissionGuarded},

	// Taxonomy profiles / areas / families. F-001 split: GET view, writes manage.
	{method: http.MethodGet, pathPrefix: "/api/v1/taxonomy/profiles", capability: iamdomain.CapTaxonomyView, visibility: iamdelivery.VisibilityPermissionGuarded},
	{method: http.MethodPost, pathPrefix: "/api/v1/taxonomy/profiles", capability: iamdomain.CapTaxonomyManage, visibility: iamdelivery.VisibilityPermissionGuarded},
	{method: http.MethodPatch, pathPrefix: "/api/v1/taxonomy/profiles", capability: iamdomain.CapTaxonomyManage, visibility: iamdelivery.VisibilityPermissionGuarded},
	{method: http.MethodPut, pathPrefix: "/api/v1/taxonomy/profiles", capability: iamdomain.CapTaxonomyManage, visibility: iamdelivery.VisibilityPermissionGuarded},
	{method: http.MethodDelete, pathPrefix: "/api/v1/taxonomy/profiles", capability: iamdomain.CapTaxonomyManage, visibility: iamdelivery.VisibilityPermissionGuarded},

	{method: http.MethodGet, pathPrefix: "/api/v1/taxonomy/areas", capability: iamdomain.CapTaxonomyView, visibility: iamdelivery.VisibilityPermissionGuarded},
	{method: http.MethodPost, pathPrefix: "/api/v1/taxonomy/areas", capability: iamdomain.CapTaxonomyManage, visibility: iamdelivery.VisibilityPermissionGuarded},
	{method: http.MethodPatch, pathPrefix: "/api/v1/taxonomy/areas", capability: iamdomain.CapTaxonomyManage, visibility: iamdelivery.VisibilityPermissionGuarded},
	{method: http.MethodPut, pathPrefix: "/api/v1/taxonomy/areas", capability: iamdomain.CapTaxonomyManage, visibility: iamdelivery.VisibilityPermissionGuarded},
	{method: http.MethodDelete, pathPrefix: "/api/v1/taxonomy/areas", capability: iamdomain.CapTaxonomyManage, visibility: iamdelivery.VisibilityPermissionGuarded},

	{method: http.MethodGet, pathPrefix: "/api/v1/taxonomy/families", capability: iamdomain.CapTaxonomyView, visibility: iamdelivery.VisibilityPermissionGuarded},
	{method: http.MethodPost, pathPrefix: "/api/v1/taxonomy/families", capability: iamdomain.CapTaxonomyManage, visibility: iamdelivery.VisibilityPermissionGuarded},
	{method: http.MethodPatch, pathPrefix: "/api/v1/taxonomy/families", capability: iamdomain.CapTaxonomyManage, visibility: iamdelivery.VisibilityPermissionGuarded},
	{method: http.MethodPut, pathPrefix: "/api/v1/taxonomy/families", capability: iamdomain.CapTaxonomyManage, visibility: iamdelivery.VisibilityPermissionGuarded},
	{method: http.MethodDelete, pathPrefix: "/api/v1/taxonomy/families", capability: iamdomain.CapTaxonomyManage, visibility: iamdelivery.VisibilityPermissionGuarded},

	// Token dictionary (SP-1). F-001 split: GET view, writes manage.
	{method: http.MethodGet, pathPrefix: "/api/v1/tokens", capability: iamdomain.CapTokenView, visibility: iamdelivery.VisibilityPermissionGuarded},
	{method: http.MethodPost, pathExact: "/api/v1/tokens", capability: iamdomain.CapTokenDictionaryManage, visibility: iamdelivery.VisibilityPermissionGuarded},
	{method: http.MethodPut, pathPrefix: "/api/v1/tokens", capability: iamdomain.CapTokenDictionaryManage, visibility: iamdelivery.VisibilityPermissionGuarded},
	{method: http.MethodDelete, pathPrefix: "/api/v1/tokens", capability: iamdomain.CapTokenDictionaryManage, visibility: iamdelivery.VisibilityPermissionGuarded},

	// Controlled documents.
	{method: http.MethodGet, pathPrefix: "/api/v1/controlled-documents", capability: iamdomain.CapDocumentView, visibility: iamdelivery.VisibilityPermissionGuarded},
	{method: http.MethodPost, pathExact: "/api/v1/controlled-documents", capability: iamdomain.CapControlledDocumentCreate, visibility: iamdelivery.VisibilityPermissionGuarded},
	{method: http.MethodPost, pathPrefix: "/api/v1/controlled-documents", pathSuffix: "/revisions", capability: iamdomain.CapDocumentEdit, visibility: iamdelivery.VisibilityPermissionGuarded},
	{method: http.MethodPut, pathPrefix: "/api/v1/controlled-documents", pathSuffix: "/obsolete", capability: iamdomain.CapControlledDocumentObsolete, visibility: iamdelivery.VisibilityPermissionGuarded},
	{method: http.MethodPut, pathPrefix: "/api/v1/controlled-documents", pathSuffix: "/supersede", capability: iamdomain.CapControlledDocumentSupersede, visibility: iamdelivery.VisibilityPermissionGuarded},

	// IAM Admin Center "Roles & Capabilities" tab (PR-5). Read-only matrix
	// endpoints; all gated on CapMembershipView.
	{method: http.MethodGet, pathExact: "/api/v1/iam/roles", capability: iamdomain.CapMembershipView, visibility: iamdelivery.VisibilityPermissionGuarded},
	{method: http.MethodGet, pathExact: "/api/v1/iam/capabilities", capability: iamdomain.CapMembershipView, visibility: iamdelivery.VisibilityPermissionGuarded},
	{method: http.MethodGet, pathExact: "/api/v1/iam/role-capabilities", capability: iamdomain.CapMembershipView, visibility: iamdelivery.VisibilityPermissionGuarded},

	// IAM area memberships (PR-1). ADR 0016 view/manage split: GET view (the
	// directory result is area-scoped at the data layer via
	// AreaMembershipService.DirectoryScope — ADR 0022 Phase 4; the former
	// canManageMembershipTarget handler role gate was removed in Phase 3),
	// POST + DELETE manage. PUT/PATCH are intentionally omitted — the API surface
	// is grant + revoke only; role changes go via revoke-then-grant.
	{method: http.MethodGet, pathExact: "/api/v1/iam/area-memberships", capability: iamdomain.CapMembershipView, visibility: iamdelivery.VisibilityPermissionGuarded},
	{method: http.MethodPost, pathExact: "/api/v1/iam/area-memberships", capability: iamdomain.CapMembershipManage, visibility: iamdelivery.VisibilityPermissionGuarded},
	// F-DELETE-SHAPE (Phase E2): revoke moved to /iam/area-memberships/{user_id}/{area_code};
	// the trailing slash on the prefix keeps it from matching the GET/POST collection path.
	{method: http.MethodDelete, pathPrefix: "/api/v1/iam/area-memberships/", capability: iamdomain.CapMembershipManage, visibility: iamdelivery.VisibilityPermissionGuarded},

	// Signed-URL relay.
	{method: http.MethodGet, pathExact: "/api/v1/signed", capability: iamdomain.CapTemplateView, visibility: iamdelivery.VisibilityPermissionGuarded},

	// Approval route administration (ADR 0022 Phase 11, F4). These MUST precede
	// the generic /api/v1/approval/ block below so the route-admin CRUD ops gate on
	// CapRouteManage at tier-1 — matching their tier-2 authz.Require(CapRouteManage)
	// in approval/application/route_admin_service.go. Without these explicit rows
	// the generic prefix would gate them on document.view/document.submit (tier-1)
	// while tier-2 demands route.manage — the F4 cross-tier divergence. route.manage
	// is tenant-grade, so all four ops (incl. the GET list, whose tier-2 List check
	// is also route.manage) map to it.
	// POST covers both create (/approval/routes) and the F-DELETE-SHAPE (Phase E2)
	// deactivate action (/approval/routes/{id}/deactivate); no DELETE under this prefix anymore.
	{method: http.MethodGet, pathPrefix: "/api/v1/approval/routes", capability: iamdomain.CapRouteManage, visibility: iamdelivery.VisibilityPermissionGuarded},
	{method: http.MethodPost, pathPrefix: "/api/v1/approval/routes", capability: iamdomain.CapRouteManage, visibility: iamdelivery.VisibilityPermissionGuarded},
	{method: http.MethodPut, pathPrefix: "/api/v1/approval/routes", capability: iamdomain.CapRouteManage, visibility: iamdelivery.VisibilityPermissionGuarded},

	// Approval (legacy mount).
	{method: http.MethodGet, pathPrefix: "/api/v1/approval/", capability: iamdomain.CapDocumentView, visibility: iamdelivery.VisibilityPermissionGuarded},
	{method: http.MethodPost, pathPrefix: "/api/v1/approval/", capability: iamdomain.CapDocumentSubmit, visibility: iamdelivery.VisibilityPermissionGuarded},
	{method: http.MethodPut, pathPrefix: "/api/v1/approval/", capability: iamdomain.CapDocumentSubmit, visibility: iamdelivery.VisibilityPermissionGuarded},
	{method: http.MethodDelete, pathPrefix: "/api/v1/approval/", capability: iamdomain.CapDocumentSubmit, visibility: iamdelivery.VisibilityPermissionGuarded},

	// Notifications — self-scoped read surface (M3/F3.2, F3.4). CapNotificationRead at tier-1
	// (mirrors CapAuditRead). Most specific first: read-all (exact) and unread-count (exact)
	// and /read (suffix) precede the bare collection GET.
	{method: http.MethodPost, pathExact: "/api/v1/notifications/read-all", capability: iamdomain.CapNotificationRead, visibility: iamdelivery.VisibilityPermissionGuarded},
	{method: http.MethodGet, pathExact: "/api/v1/notifications/unread-count", capability: iamdomain.CapNotificationRead, visibility: iamdelivery.VisibilityPermissionGuarded},
	{method: http.MethodPost, pathPrefix: "/api/v1/notifications/", pathSuffix: "/read", capability: iamdomain.CapNotificationRead, visibility: iamdelivery.VisibilityPermissionGuarded},
	{method: http.MethodGet, pathExact: "/api/v1/notifications", capability: iamdomain.CapNotificationRead, visibility: iamdelivery.VisibilityPermissionGuarded},

	// Audit.
	{method: http.MethodGet, pathExact: "/api/v1/audit/events", capability: iamdomain.CapAuditRead, visibility: iamdelivery.VisibilityPermissionGuarded},
	{method: http.MethodPost, pathExact: "/api/v1/audit/events/export", capability: iamdomain.CapAuditRead, visibility: iamdelivery.VisibilityPermissionGuarded},
	{method: http.MethodGet, pathPrefix: "/api/v1/audit/events/export/", capability: iamdomain.CapAuditRead, visibility: iamdelivery.VisibilityPermissionGuarded},

	// Sessions & Security (PR-7). Read endpoints fold under CapUserView; the
	// destructive force-logout requires CapSessionManage (PR-2 cap).
	{method: http.MethodGet, pathExact: "/api/v1/auth/sessions", capability: iamdomain.CapUserView, visibility: iamdelivery.VisibilityPermissionGuarded},
	{method: http.MethodDelete, pathPrefix: "/api/v1/auth/sessions/", capability: iamdomain.CapSessionManage, visibility: iamdelivery.VisibilityPermissionGuarded},
	{method: http.MethodGet, pathExact: "/api/v1/security/mfa-coverage", capability: iamdomain.CapUserView, visibility: iamdelivery.VisibilityPermissionGuarded},
	{method: http.MethodGet, pathExact: "/api/v1/security/lockouts", capability: iamdomain.CapUserView, visibility: iamdelivery.VisibilityPermissionGuarded},
	{method: http.MethodGet, pathExact: "/api/v1/security/signals", capability: iamdomain.CapUserView, visibility: iamdelivery.VisibilityPermissionGuarded},

	// Presence (PR-9). Snapshot fallback + WebSocket stream are both
	// tenant-scoped observability; both fold under CapUserView.
	{method: http.MethodGet, pathExact: "/api/v1/iam/presence/snapshot", capability: iamdomain.CapUserView, visibility: iamdelivery.VisibilityPermissionGuarded},
	{method: http.MethodGet, pathExact: "/api/v1/iam/presence/stream", capability: iamdomain.CapUserView, visibility: iamdelivery.VisibilityPermissionGuarded},
}

func newPermissionResolver() iamdelivery.PermissionResolver {
	return func(method, path string) (iamdomain.Capability, iamdelivery.Visibility) {
		if capability, visibility, ok := resolveRoutePermission(method, path); ok {
			return capability, visibility
		}
		// Fail-closed default. Any route not enumerated above demands at
		// least a session — never silently public.
		return "", iamdelivery.VisibilitySessionRequired
	}
}

func resolveRoutePermission(method, path string) (iamdomain.Capability, iamdelivery.Visibility, bool) {
	for _, rule := range routeRules {
		if rule.matches(method, path) {
			return rule.capability, rule.visibility, true
		}
	}
	return "", iamdelivery.VisibilityPermissionGuarded, false
}

func newPublicPathChecker(resolver iamdelivery.PermissionResolver) authdelivery.PublicPathChecker {
	return func(method, path string) bool {
		_, visibility := resolver(method, path)
		return visibility == iamdelivery.VisibilityPublic
	}
}
