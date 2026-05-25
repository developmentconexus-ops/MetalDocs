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
// capability. New routes MUST be added here; the unmatched default is
// VisibilitySessionRequired, so a forgotten entry produces a 401 (loud), not a
// silent public route (catastrophic).
var routeRules = []routeRule{
	// ---- Public (no session required) -----------------------------------
	{method: http.MethodGet, pathPrefix: "/api/v1/health/", visibility: iamdelivery.VisibilityPublic},
	{pathExact: "/healthz", visibility: iamdelivery.VisibilityPublic},
	{method: http.MethodPost, pathExact: "/api/v1/auth/login", visibility: iamdelivery.VisibilityPublic},
	{method: http.MethodPost, pathExact: "/api/v1/auth/refresh", visibility: iamdelivery.VisibilityPublic},
	{method: http.MethodGet, pathExact: "/api/v1/feature-flags", visibility: iamdelivery.VisibilityPublic},

	// ---- Session-required (authenticated, no specific capability) -------
	{method: http.MethodGet, pathExact: "/api/v1/auth/me", visibility: iamdelivery.VisibilitySessionRequired},
	{method: http.MethodPost, pathExact: "/api/v1/auth/change-password", visibility: iamdelivery.VisibilitySessionRequired},
	{method: http.MethodPost, pathExact: "/api/v1/auth/logout", visibility: iamdelivery.VisibilitySessionRequired},

	// ---- Permission-guarded ---------------------------------------------
	// Metrics (H5 left intact per scope — separate finding).
	{pathExact: "/api/v1/metrics", capability: iamdomain.CapUserManage, visibility: iamdelivery.VisibilityPermissionGuarded},

	// Search / notifications.
	{method: http.MethodGet, pathExact: "/api/v1/search/documents", capability: iamdomain.CapDocumentView, visibility: iamdelivery.VisibilityPermissionGuarded},
	{method: http.MethodGet, pathExact: "/api/v1/notifications", capability: iamdomain.CapDocumentView, visibility: iamdelivery.VisibilityPermissionGuarded},
	{method: http.MethodPost, pathPrefix: "/api/v1/notifications/", pathSuffix: "/read", capability: iamdomain.CapDocumentView, visibility: iamdelivery.VisibilityPermissionGuarded},

	// Access policies.
	{method: http.MethodGet, pathExact: "/api/v1/access-policies", capability: iamdomain.CapMembershipManage, visibility: iamdelivery.VisibilityPermissionGuarded},
	{method: http.MethodPut, pathExact: "/api/v1/access-policies", capability: iamdomain.CapMembershipManage, visibility: iamdelivery.VisibilityPermissionGuarded},

	// Workflow.
	{method: http.MethodPost, pathPrefix: "/api/v1/workflow/documents/", pathSuffix: "/transitions", capability: iamdomain.CapDocumentSubmit, visibility: iamdelivery.VisibilityPermissionGuarded},
	{method: http.MethodGet, pathPrefix: "/api/v1/workflow/documents/", pathSuffix: "/approvals", capability: iamdomain.CapDocumentView, visibility: iamdelivery.VisibilityPermissionGuarded},

	// IAM users.
	{method: http.MethodPost, pathExact: "/api/v1/iam/users", capability: iamdomain.CapUserManage, visibility: iamdelivery.VisibilityPermissionGuarded},
	{method: http.MethodGet, pathExact: "/api/v1/iam/users", capability: iamdomain.CapUserManage, visibility: iamdelivery.VisibilityPermissionGuarded},
	{method: http.MethodPatch, pathPrefix: "/api/v1/iam/users/", notSuffix: "/roles", capability: iamdomain.CapUserManage, visibility: iamdelivery.VisibilityPermissionGuarded},
	{method: http.MethodPost, pathPrefix: "/api/v1/iam/users/", pathSuffix: "/roles", capability: iamdomain.CapUserManage, visibility: iamdelivery.VisibilityPermissionGuarded},
	{method: http.MethodPut, pathPrefix: "/api/v1/iam/users/", pathSuffix: "/roles", capability: iamdomain.CapUserManage, visibility: iamdelivery.VisibilityPermissionGuarded},
	{method: http.MethodPost, pathPrefix: "/api/v1/iam/users/", pathSuffix: "/reset-password", capability: iamdomain.CapUserManage, visibility: iamdelivery.VisibilityPermissionGuarded},
	{method: http.MethodPost, pathPrefix: "/api/v1/iam/users/", pathSuffix: "/unlock", capability: iamdomain.CapUserManage, visibility: iamdelivery.VisibilityPermissionGuarded},
	{method: http.MethodGet, pathExact: "/api/v1/iam/admin/overview", capability: iamdomain.CapUserManage, visibility: iamdelivery.VisibilityPermissionGuarded},

	// Legacy taxonomy aliases.
	{method: http.MethodGet, pathPrefix: "/api/v1/document-profiles", capability: iamdomain.CapDocumentView, visibility: iamdelivery.VisibilityPermissionGuarded},
	{method: http.MethodPost, pathPrefix: "/api/v1/document-profiles", capability: iamdomain.CapTaxonomyManage, visibility: iamdelivery.VisibilityPermissionGuarded},
	{method: http.MethodPut, pathPrefix: "/api/v1/document-profiles", capability: iamdomain.CapTaxonomyManage, visibility: iamdelivery.VisibilityPermissionGuarded},
	{method: http.MethodDelete, pathPrefix: "/api/v1/document-profiles", capability: iamdomain.CapTaxonomyManage, visibility: iamdelivery.VisibilityPermissionGuarded},

	{method: http.MethodGet, pathPrefix: "/api/v1/process-areas", capability: iamdomain.CapDocumentView, visibility: iamdelivery.VisibilityPermissionGuarded},
	{method: http.MethodPost, pathPrefix: "/api/v1/process-areas", capability: iamdomain.CapTaxonomyManage, visibility: iamdelivery.VisibilityPermissionGuarded},
	{method: http.MethodPut, pathPrefix: "/api/v1/process-areas", capability: iamdomain.CapTaxonomyManage, visibility: iamdelivery.VisibilityPermissionGuarded},
	{method: http.MethodDelete, pathPrefix: "/api/v1/process-areas", capability: iamdomain.CapTaxonomyManage, visibility: iamdelivery.VisibilityPermissionGuarded},

	{method: http.MethodGet, pathPrefix: "/api/v1/document-subjects", capability: iamdomain.CapDocumentView, visibility: iamdelivery.VisibilityPermissionGuarded},
	{method: http.MethodPost, pathPrefix: "/api/v1/document-subjects", capability: iamdomain.CapTaxonomyManage, visibility: iamdelivery.VisibilityPermissionGuarded},
	{method: http.MethodPut, pathPrefix: "/api/v1/document-subjects", capability: iamdomain.CapTaxonomyManage, visibility: iamdelivery.VisibilityPermissionGuarded},
	{method: http.MethodDelete, pathPrefix: "/api/v1/document-subjects", capability: iamdomain.CapTaxonomyManage, visibility: iamdelivery.VisibilityPermissionGuarded},

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
	{method: http.MethodPost, pathPrefix: "/api/v1/templates", pathSuffix: "/archive", capability: iamdomain.Capability("template.archive"), visibility: iamdelivery.VisibilityPermissionGuarded},

	// Documents — order preserves the original switch semantics (more specific suffixes/contains first).
	{method: http.MethodGet, pathPrefix: "/api/v1/documents", capability: iamdomain.CapDocumentView, visibility: iamdelivery.VisibilityPermissionGuarded},
	{method: http.MethodPost, pathExact: "/api/v1/documents", capability: iamdomain.CapDocumentCreate, visibility: iamdelivery.VisibilityPermissionGuarded},
	{method: http.MethodPost, pathPrefix: "/api/v1/documents", pathSuffix: "/finalize", capability: iamdomain.CapDocumentSignoff, visibility: iamdelivery.VisibilityPermissionGuarded},
	{method: http.MethodPost, pathPrefix: "/api/v1/documents", pathSuffix: "/archive", capability: iamdomain.CapDocumentEdit, visibility: iamdelivery.VisibilityPermissionGuarded},
	{method: http.MethodPost, pathPrefix: "/api/v1/documents", contains: "/session/force-release", capability: iamdomain.CapMembershipManage, visibility: iamdelivery.VisibilityPermissionGuarded},
	{method: http.MethodPost, pathPrefix: "/api/v1/documents", contains: "/session/", capability: iamdomain.CapDocumentEdit, visibility: iamdelivery.VisibilityPermissionGuarded},
	{method: http.MethodPost, pathPrefix: "/api/v1/documents", contains: "/autosave/", capability: iamdomain.CapDocumentEdit, visibility: iamdelivery.VisibilityPermissionGuarded},
	{method: http.MethodPost, pathPrefix: "/api/v1/documents", pathSuffix: "/artifact-metadata", capability: iamdomain.CapDocumentEdit, visibility: iamdelivery.VisibilityPermissionGuarded},
	{method: http.MethodPost, pathPrefix: "/api/v1/documents", contains: "/checkpoints/", pathSuffix: "/restore", capability: iamdomain.CapDocumentEdit, visibility: iamdelivery.VisibilityPermissionGuarded},
	{method: http.MethodPost, pathPrefix: "/api/v1/documents", contains: "/checkpoints", capability: iamdomain.CapDocumentEdit, visibility: iamdelivery.VisibilityPermissionGuarded},
	{method: http.MethodPost, pathPrefix: "/api/v1/documents", pathSuffix: "/export/pdf", capability: iamdomain.CapDocumentView, visibility: iamdelivery.VisibilityPermissionGuarded},
	{method: http.MethodPut, pathPrefix: "/api/v1/documents", contains: "/placeholders/", capability: iamdomain.CapDocumentEdit, visibility: iamdelivery.VisibilityPermissionGuarded},
	{method: http.MethodPatch, pathPrefix: "/api/v1/documents", capability: iamdomain.CapDocumentEdit, visibility: iamdelivery.VisibilityPermissionGuarded},
	{method: http.MethodPost, pathPrefix: "/api/v1/documents", pathSuffix: "/submit", capability: iamdomain.CapDocumentSubmit, visibility: iamdelivery.VisibilityPermissionGuarded},
	{method: http.MethodPost, pathPrefix: "/api/v1/documents", pathSuffix: "/signoff", capability: iamdomain.CapDocumentSignoff, visibility: iamdelivery.VisibilityPermissionGuarded},
	{method: http.MethodPost, pathPrefix: "/api/v1/documents", pathSuffix: "/publish", capability: iamdomain.Capability("doc.publish"), visibility: iamdelivery.VisibilityPermissionGuarded},
	{method: http.MethodPost, pathPrefix: "/api/v1/documents", pathSuffix: "/schedule-publish", capability: iamdomain.Capability("doc.publish"), visibility: iamdelivery.VisibilityPermissionGuarded},
	{method: http.MethodPost, pathPrefix: "/api/v1/documents", pathSuffix: "/supersede", capability: iamdomain.Capability("doc.supersede"), visibility: iamdelivery.VisibilityPermissionGuarded},
	{method: http.MethodPost, pathPrefix: "/api/v1/documents", pathSuffix: "/obsolete", capability: iamdomain.Capability("doc.obsolete"), visibility: iamdelivery.VisibilityPermissionGuarded},
	{method: http.MethodPost, pathPrefix: "/api/v1/documents", pathSuffix: "/cancel", capability: iamdomain.CapDocumentEdit, visibility: iamdelivery.VisibilityPermissionGuarded},
	{method: http.MethodPost, pathPrefix: "/api/v1/documents", pathSuffix: "/reconstruct", capability: iamdomain.CapDocumentEdit, visibility: iamdelivery.VisibilityPermissionGuarded},

	// Taxonomy profiles / areas / families.
	{method: http.MethodGet, pathPrefix: "/api/v1/taxonomy/profiles", capability: iamdomain.CapTaxonomyManage, visibility: iamdelivery.VisibilityPermissionGuarded},
	{method: http.MethodPost, pathPrefix: "/api/v1/taxonomy/profiles", capability: iamdomain.CapTaxonomyManage, visibility: iamdelivery.VisibilityPermissionGuarded},
	{method: http.MethodPatch, pathPrefix: "/api/v1/taxonomy/profiles", capability: iamdomain.CapTaxonomyManage, visibility: iamdelivery.VisibilityPermissionGuarded},
	{method: http.MethodPut, pathPrefix: "/api/v1/taxonomy/profiles", capability: iamdomain.CapTaxonomyManage, visibility: iamdelivery.VisibilityPermissionGuarded},
	{method: http.MethodDelete, pathPrefix: "/api/v1/taxonomy/profiles", capability: iamdomain.CapTaxonomyManage, visibility: iamdelivery.VisibilityPermissionGuarded},

	{method: http.MethodGet, pathPrefix: "/api/v1/taxonomy/areas", capability: iamdomain.CapTaxonomyManage, visibility: iamdelivery.VisibilityPermissionGuarded},
	{method: http.MethodPost, pathPrefix: "/api/v1/taxonomy/areas", capability: iamdomain.CapTaxonomyManage, visibility: iamdelivery.VisibilityPermissionGuarded},
	{method: http.MethodPatch, pathPrefix: "/api/v1/taxonomy/areas", capability: iamdomain.CapTaxonomyManage, visibility: iamdelivery.VisibilityPermissionGuarded},
	{method: http.MethodPut, pathPrefix: "/api/v1/taxonomy/areas", capability: iamdomain.CapTaxonomyManage, visibility: iamdelivery.VisibilityPermissionGuarded},
	{method: http.MethodDelete, pathPrefix: "/api/v1/taxonomy/areas", capability: iamdomain.CapTaxonomyManage, visibility: iamdelivery.VisibilityPermissionGuarded},

	{method: http.MethodGet, pathPrefix: "/api/v1/taxonomy/families", capability: iamdomain.CapTaxonomyManage, visibility: iamdelivery.VisibilityPermissionGuarded},
	{method: http.MethodPost, pathPrefix: "/api/v1/taxonomy/families", capability: iamdomain.CapTaxonomyManage, visibility: iamdelivery.VisibilityPermissionGuarded},
	{method: http.MethodPatch, pathPrefix: "/api/v1/taxonomy/families", capability: iamdomain.CapTaxonomyManage, visibility: iamdelivery.VisibilityPermissionGuarded},
	{method: http.MethodPut, pathPrefix: "/api/v1/taxonomy/families", capability: iamdomain.CapTaxonomyManage, visibility: iamdelivery.VisibilityPermissionGuarded},
	{method: http.MethodDelete, pathPrefix: "/api/v1/taxonomy/families", capability: iamdomain.CapTaxonomyManage, visibility: iamdelivery.VisibilityPermissionGuarded},

	// Controlled documents.
	{method: http.MethodGet, pathPrefix: "/api/v1/controlled-documents", capability: iamdomain.CapDocumentView, visibility: iamdelivery.VisibilityPermissionGuarded},
	{method: http.MethodPost, pathExact: "/api/v1/controlled-documents", capability: iamdomain.CapControlledDocumentCreate, visibility: iamdelivery.VisibilityPermissionGuarded},
	{method: http.MethodPost, pathPrefix: "/api/v1/controlled-documents", pathSuffix: "/revisions", capability: iamdomain.CapDocumentEdit, visibility: iamdelivery.VisibilityPermissionGuarded},
	{method: http.MethodPut, pathPrefix: "/api/v1/controlled-documents", pathSuffix: "/obsolete", capability: iamdomain.CapControlledDocumentObsolete, visibility: iamdelivery.VisibilityPermissionGuarded},
	{method: http.MethodPut, pathPrefix: "/api/v1/controlled-documents", pathSuffix: "/supersede", capability: iamdomain.CapControlledDocumentSupersede, visibility: iamdelivery.VisibilityPermissionGuarded},

	// IAM area memberships — any method.
	{pathPrefix: "/api/v1/iam/area-memberships", capability: iamdomain.CapMembershipManage, visibility: iamdelivery.VisibilityPermissionGuarded},

	// Signed-URL relay.
	{method: http.MethodGet, pathExact: "/api/v1/signed", capability: iamdomain.CapTemplateView, visibility: iamdelivery.VisibilityPermissionGuarded},

	// Approval (legacy mount).
	{method: http.MethodGet, pathPrefix: "/api/v1/approval/", capability: iamdomain.CapDocumentView, visibility: iamdelivery.VisibilityPermissionGuarded},
	{method: http.MethodPost, pathPrefix: "/api/v1/approval/", capability: iamdomain.CapDocumentSubmit, visibility: iamdelivery.VisibilityPermissionGuarded},
	{method: http.MethodPut, pathPrefix: "/api/v1/approval/", capability: iamdomain.CapDocumentSubmit, visibility: iamdelivery.VisibilityPermissionGuarded},
	{method: http.MethodDelete, pathPrefix: "/api/v1/approval/", capability: iamdomain.CapDocumentSubmit, visibility: iamdelivery.VisibilityPermissionGuarded},

	// Audit.
	{method: http.MethodGet, pathExact: "/api/v1/audit/events", capability: iamdomain.CapAuditRead, visibility: iamdelivery.VisibilityPermissionGuarded},
}

func newPermissionResolver() iamdelivery.PermissionResolver {
	return func(method, path string) (iamdomain.Capability, iamdelivery.Visibility) {
		for _, rule := range routeRules {
			if rule.matches(method, path) {
				return rule.capability, rule.visibility
			}
		}
		// Fail-closed default. Any route not enumerated above demands at
		// least a session — never silently public.
		if strings.HasPrefix(path, "/api/v1/documents") {
			return iamdomain.CapDocumentView, iamdelivery.VisibilitySessionRequired
		}
		if strings.HasPrefix(path, "/api/v1/templates") {
			return iamdomain.CapTemplateView, iamdelivery.VisibilitySessionRequired
		}
		return "", iamdelivery.VisibilitySessionRequired
	}
}

func newPublicPathChecker(resolver iamdelivery.PermissionResolver) authdelivery.PublicPathChecker {
	return func(method, path string) bool {
		_, visibility := resolver(method, path)
		return visibility == iamdelivery.VisibilityPublic
	}
}
