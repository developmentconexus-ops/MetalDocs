package main

import (
	"net/http"
	"strings"

	authdelivery "metaldocs/internal/modules/auth/delivery/http"
	iamdelivery "metaldocs/internal/modules/iam/delivery/http"
	iamdomain "metaldocs/internal/modules/iam/domain"
)

func newPermissionResolver() iamdelivery.PermissionResolver {
	return func(method, path string) (iamdomain.Capability, bool) {
		if path == "/api/v1/health/live" || path == "/api/v1/health/ready" {
			return "", false
		}
		if strings.HasPrefix(path, "/api/v1/auth/") {
			return "", false
		}
		if method == http.MethodGet && path == "/api/v1/feature-flags" {
			return "", false
		}

		if path == "/api/v1/metrics" {
			return iamdomain.CapUserManage, true
		}
		if method == http.MethodGet && path == "/api/v1/search/documents" {
			return iamdomain.CapDocumentView, true
		}
		if method == http.MethodGet && path == "/api/v1/notifications" {
			return iamdomain.CapDocumentView, true
		}
		if method == http.MethodPost && strings.HasPrefix(path, "/api/v1/notifications/") && strings.HasSuffix(path, "/read") {
			return iamdomain.CapDocumentView, true
		}
		if (method == http.MethodGet || method == http.MethodPut) && path == "/api/v1/access-policies" {
			return iamdomain.CapMembershipManage, true
		}
		if method == http.MethodPost && strings.HasPrefix(path, "/api/v1/workflow/documents/") && strings.HasSuffix(path, "/transitions") {
			return iamdomain.CapDocumentSubmit, true
		}
		if method == http.MethodGet && strings.HasPrefix(path, "/api/v1/workflow/documents/") && strings.HasSuffix(path, "/approvals") {
			return iamdomain.CapDocumentView, true
		}
		if method == http.MethodPost && path == "/api/v1/iam/users" {
			return iamdomain.CapUserManage, true
		}
		if method == http.MethodGet && path == "/api/v1/iam/users" {
			return iamdomain.CapUserManage, true
		}
		if method == http.MethodPatch && strings.HasPrefix(path, "/api/v1/iam/users/") && !strings.HasSuffix(path, "/roles") {
			return iamdomain.CapUserManage, true
		}
		if method == http.MethodPost && strings.HasPrefix(path, "/api/v1/iam/users/") && strings.HasSuffix(path, "/roles") {
			return iamdomain.CapUserManage, true
		}
		if method == http.MethodPut && strings.HasPrefix(path, "/api/v1/iam/users/") && strings.HasSuffix(path, "/roles") {
			return iamdomain.CapUserManage, true
		}
		if method == http.MethodPost && strings.HasPrefix(path, "/api/v1/iam/users/") && strings.HasSuffix(path, "/reset-password") {
			return iamdomain.CapUserManage, true
		}
		if method == http.MethodPost && strings.HasPrefix(path, "/api/v1/iam/users/") && strings.HasSuffix(path, "/unlock") {
			return iamdomain.CapUserManage, true
		}
		if method == http.MethodGet && path == "/api/v1/iam/admin/overview" {
			return iamdomain.CapUserManage, true
		}
		if strings.HasPrefix(path, "/api/v1/document-profiles") {
			switch method {
			case http.MethodGet:
				return iamdomain.CapDocumentView, true
			case http.MethodPost, http.MethodPut, http.MethodDelete:
				return iamdomain.CapTaxonomyManage, true
			}
		}
		if strings.HasPrefix(path, "/api/v1/process-areas") {
			switch method {
			case http.MethodGet:
				return iamdomain.CapDocumentView, true
			case http.MethodPost, http.MethodPut, http.MethodDelete:
				return iamdomain.CapTaxonomyManage, true
			}
		}
		if strings.HasPrefix(path, "/api/v1/document-subjects") {
			switch method {
			case http.MethodGet:
				return iamdomain.CapDocumentView, true
			case http.MethodPost, http.MethodPut, http.MethodDelete:
				return iamdomain.CapTaxonomyManage, true
			}
		}
		if strings.HasPrefix(path, "/api/v1/templates") {
			switch {
			case method == http.MethodGet:
				return iamdomain.CapTemplateView, true
			case method == http.MethodPost && path == "/api/v1/templates":
				return iamdomain.CapTemplateCreate, true
			case method == http.MethodPost && strings.HasSuffix(path, "/versions"):
				return iamdomain.CapTemplateCreate, true
			case method == http.MethodPut && strings.HasSuffix(path, "/draft"):
				return iamdomain.CapTemplateEdit, true
			case method == http.MethodPut && strings.HasSuffix(path, "/schema"):
				return iamdomain.CapTemplateEdit, true
			case method == http.MethodPost && strings.HasSuffix(path, "/publish"):
				return iamdomain.CapTemplatePublish, true
			case method == http.MethodPost && strings.HasSuffix(path, "/docx-upload-url"):
				return iamdomain.CapTemplateEdit, true
			case method == http.MethodPost && strings.HasSuffix(path, "/schema-upload-url"):
				return iamdomain.CapTemplateEdit, true
			case method == http.MethodPost && strings.Contains(path, "/autosave/"):
				return iamdomain.CapTemplateEdit, true
			case method == http.MethodPost && strings.HasSuffix(path, "/submit"):
				return iamdomain.CapTemplateSubmit, true
			case method == http.MethodPost && strings.HasSuffix(path, "/review"):
				return iamdomain.CapTemplateReview, true
			case method == http.MethodPost && strings.HasSuffix(path, "/approve"):
				return iamdomain.CapTemplateApprove, true
			case method == http.MethodPut && strings.HasSuffix(path, "/approval-config"):
				return iamdomain.CapTemplateEdit, true
			case method == http.MethodPost && strings.HasSuffix(path, "/archive"):
				return iamdomain.Capability("template.archive"), true
			}
		}
		if strings.HasPrefix(path, "/api/v1/documents") {
			switch {
			case method == http.MethodGet:
				return iamdomain.CapDocumentView, true
			case method == http.MethodPost && path == "/api/v1/documents":
				return iamdomain.CapDocumentCreate, true
			case method == http.MethodPost && strings.HasSuffix(path, "/finalize"):
				return iamdomain.CapDocumentSignoff, true
			case method == http.MethodPost && strings.HasSuffix(path, "/archive"):
				return iamdomain.CapDocumentEdit, true
			case method == http.MethodPost && strings.Contains(path, "/session/force-release"):
				return iamdomain.CapMembershipManage, true
			case method == http.MethodPost && strings.Contains(path, "/session/"):
				return iamdomain.CapDocumentEdit, true
			case method == http.MethodPost && strings.Contains(path, "/autosave/"):
				return iamdomain.CapDocumentEdit, true
			case method == http.MethodPost && strings.HasSuffix(path, "/artifact-metadata"):
				return iamdomain.CapDocumentEdit, true
			case method == http.MethodPost && strings.Contains(path, "/checkpoints/") && strings.HasSuffix(path, "/restore"):
				return iamdomain.CapDocumentEdit, true
			case method == http.MethodPost && strings.Contains(path, "/checkpoints"):
				return iamdomain.CapDocumentEdit, true
			case method == http.MethodPost && strings.HasSuffix(path, "/export/pdf"):
				return iamdomain.CapDocumentView, true
			case method == http.MethodPut && strings.Contains(path, "/placeholders/"):
				return iamdomain.CapDocumentEdit, true
			case method == http.MethodPatch:
				return iamdomain.CapDocumentEdit, true
			case method == http.MethodPost && strings.HasSuffix(path, "/submit"):
				return iamdomain.CapDocumentSubmit, true
			case method == http.MethodPost && strings.HasSuffix(path, "/signoff"):
				return iamdomain.CapDocumentSignoff, true
			case method == http.MethodPost && strings.HasSuffix(path, "/publish"):
				return iamdomain.Capability("doc.publish"), true
			case method == http.MethodPost && strings.HasSuffix(path, "/schedule-publish"):
				return iamdomain.Capability("doc.publish"), true
			case method == http.MethodPost && strings.HasSuffix(path, "/supersede"):
				return iamdomain.Capability("doc.supersede"), true
			case method == http.MethodPost && strings.HasSuffix(path, "/obsolete"):
				return iamdomain.Capability("doc.obsolete"), true
			case method == http.MethodPost && strings.HasSuffix(path, "/cancel"):
				return iamdomain.CapDocumentEdit, true
			case method == http.MethodGet && strings.HasSuffix(path, "/approval-instance"):
				return iamdomain.CapDocumentView, true
			case method == http.MethodPost && strings.HasSuffix(path, "/reconstruct"):
				return iamdomain.CapDocumentEdit, true
			}
		}
		if strings.HasPrefix(path, "/api/v1/taxonomy/profiles") {
			switch method {
			case http.MethodGet:
				return iamdomain.CapDocumentView, true
			case http.MethodPost, http.MethodPatch, http.MethodPut, http.MethodDelete:
				return iamdomain.CapTaxonomyManage, true
			}
		}
		if strings.HasPrefix(path, "/api/v1/taxonomy/areas") {
			switch method {
			case http.MethodGet:
				return iamdomain.CapDocumentView, true
			case http.MethodPost, http.MethodPatch, http.MethodPut, http.MethodDelete:
				return iamdomain.CapTaxonomyManage, true
			}
		}
		if strings.HasPrefix(path, "/api/v1/taxonomy/families") {
			switch method {
			case http.MethodGet:
				return iamdomain.CapDocumentView, true
			case http.MethodPost, http.MethodPatch, http.MethodPut, http.MethodDelete:
				return iamdomain.CapTaxonomyManage, true
			}
		}
		if strings.HasPrefix(path, "/api/v1/controlled-documents") {
			switch {
			case method == http.MethodGet:
				return iamdomain.CapDocumentView, true
			case method == http.MethodPost && path == "/api/v1/controlled-documents":
				return iamdomain.CapControlledDocumentCreate, true
			case method == http.MethodPost && strings.HasSuffix(path, "/revisions"):
				return iamdomain.CapDocumentEdit, true
			case method == http.MethodPut && strings.HasSuffix(path, "/obsolete"):
				return iamdomain.CapControlledDocumentObsolete, true
			case method == http.MethodPut && strings.HasSuffix(path, "/supersede"):
				return iamdomain.CapControlledDocumentSupersede, true
			}
		}
		if strings.HasPrefix(path, "/api/v1/iam/area-memberships") {
			return iamdomain.CapMembershipManage, true
		}
		if method == http.MethodGet && path == "/api/v1/signed" {
			return iamdomain.CapTemplateView, true
		}
		if strings.HasPrefix(path, "/api/v1/approval/") {
			switch method {
			case http.MethodGet:
				return iamdomain.CapDocumentView, true
			case http.MethodPost, http.MethodPut, http.MethodDelete:
				return iamdomain.CapDocumentSubmit, true
			}
		}
		if method == http.MethodGet && path == "/api/v1/audit/events" {
			return iamdomain.CapAuditRead, true
		}

		return "", false
	}
}

func newPublicPathChecker(resolver iamdelivery.PermissionResolver) authdelivery.PublicPathChecker {
	return func(method, path string) bool {
		if requiresSessionButNoPermission(method, path) {
			return false
		}
		_, guarded := resolver(method, path)
		return !guarded
	}
}

func requiresSessionButNoPermission(method, path string) bool {
	if method == http.MethodGet && path == "/api/v1/auth/me" {
		return true
	}
	if method == http.MethodPost && path == "/api/v1/auth/change-password" {
		return true
	}
	if method == http.MethodPost && path == "/api/v1/auth/logout" {
		return true
	}
	return false
}
