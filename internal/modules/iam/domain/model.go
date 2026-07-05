// Package domain holds the iam module's core identity and authorization
// contracts: the canonical Role and Capability registries (ADR 0022
// Principle-1 single source of truth for capabilities), the CapabilityScope
// classification, and the narrow cross-module read ports (UserDisplayNameReader,
// TenantUserReader, MfaUserReader, etc.) that let other modules resolve
// IAM-owned data without reaching across the module boundary into IAM's
// tables directly.
package domain

import (
	"fmt"
	"strings"

	"metaldocs/internal/platform/iamtypes"
)

// Role, its canonical constants, and the Role-only helpers (IsValidRole,
// IsAreaRole, AreaRoles) now live in internal/platform/iamtypes — a neutral
// platform package with no dependency on IAM's application/infrastructure
// layers or on the auth module. This alias keeps every existing
// `domain.Role` / `iamdomain.Role` call site across the codebase compiling
// unchanged (ARC-06: closes the auth<->iam bidirectional module dependency
// that existed solely because Role lived inside iam/domain). See
// iamtypes.Role godoc for the canonical-roles source and rationale.
type Role = iamtypes.Role

const (
	RoleApprover    = iamtypes.RoleApprover
	RoleAreaAdmin   = iamtypes.RoleAreaAdmin
	RoleAuthor      = iamtypes.RoleAuthor
	RoleEditor      = iamtypes.RoleEditor
	RoleQmsAdmin    = iamtypes.RoleQmsAdmin
	RoleSigner      = iamtypes.RoleSigner
	RoleSystemAdmin = iamtypes.RoleSystemAdmin
	RoleViewer      = iamtypes.RoleViewer
)

// IsValidRole reports whether role is one of the eight canonical roles.
// Forwards to iamtypes.IsValidRole; kept here so existing domain.IsValidRole
// call sites are unaffected.
func IsValidRole(role Role) bool {
	return iamtypes.IsValidRole(role)
}

// ParseRole normalizes raw (lowercase, trimmed) and validates it against the
// eight canonical roles. Returns ErrInvalidRole if raw is empty or unknown.
func ParseRole(raw string) (Role, error) {
	role := Role(strings.ToLower(strings.TrimSpace(raw)))
	if role == "" || !IsValidRole(role) {
		return "", ErrInvalidRole
	}
	return role, nil
}

// IsAreaRole reports whether role can be held as an area membership
// (user_process_areas). system_admin is intentionally excluded — it is
// tenant-wide. Forwards to iamtypes.IsAreaRole.
func IsAreaRole(role Role) bool {
	return iamtypes.IsAreaRole(role)
}

// AreaRoles returns the canonical area-assignable roles, sorted — for
// validation messages and diagnostics. Forwards to iamtypes.AreaRoles.
func AreaRoles() []Role {
	return iamtypes.AreaRoles()
}

// Capability is the ADR 0022 Principle-1 typed authorization unit — the
// enforced boundary is always a capability, never a role. Every Capability
// value MUST be one of the constants below (validCapabilities is the single
// source of truth); no capability may be referenced as a raw string anywhere
// else in the codebase.
type Capability string

const (
	CapDocumentView      Capability = "document.view"
	CapDocumentCreate    Capability = "document.create"
	CapDocumentEdit      Capability = "document.edit"
	CapDocumentSubmit    Capability = "document.submit"
	CapDocumentSignoff   Capability = "document.signoff"
	CapDocumentPublish   Capability = "document.publish"
	CapDocumentObsolete  Capability = "document.obsolete"
	CapDocumentSupersede Capability = "document.supersede"
	// CapDocumentReview gates the eQMS periodic-review mark-reviewed workflow
	// (M6 F6.2): recording a review completion on a live published revision
	// (sets last_reviewed_at + the next review_due_at). Tenant-grade; enforced
	// tier-2 in-tx via authz.Require and backstopped by the documents/UPDATE DB
	// tripwire arm. ADR 0069.
	CapDocumentReview Capability = "document.review"

	CapTemplateView    Capability = "template.view"
	CapTemplateCreate  Capability = "template.create"
	CapTemplateEdit    Capability = "template.edit"
	CapTemplateSubmit  Capability = "template.submit"
	CapTemplateReview  Capability = "template.review"
	CapTemplateApprove Capability = "template.approve"
	CapTemplatePublish Capability = "template.publish"
	CapTemplateArchive Capability = "template.archive"
	// CapTemplateManage governs elevated template governance operations that
	// require operator authority: editing a published template's approval config.
	// Checked tier-2 in-tx via authz.Require. ADR 0051 / REQ-AUTHZ-5.
	CapTemplateManage Capability = "template.manage"

	CapControlledDocumentCreate    Capability = "controlled_documents.create"
	CapControlledDocumentObsolete  Capability = "controlled_documents.obsolete"
	CapControlledDocumentSupersede Capability = "controlled_documents.supersede"
	CapTaxonomyView                Capability = "taxonomy.view"
	CapTaxonomyManage              Capability = "taxonomy.manage"
	CapMembershipView              Capability = "membership.view"
	CapMembershipManage            Capability = "membership.manage"
	CapRouteManage                 Capability = "route.manage"
	CapUserView                    Capability = "user.view"
	CapUserManage                  Capability = "user.manage"
	CapMetricsView                 Capability = "metrics.view"
	CapAuditRead                   Capability = "audit.read"
	CapSessionManage               Capability = "session.manage"
	CapDistributionRead            Capability = "distribution.read"
	CapNotificationRead            Capability = "notification.read"

	// Token dictionary (SP-1). Both ScopeTenant — tenant-wide, not area-scoped.
	CapTokenView             Capability = "token.view"
	CapTokenDictionaryManage Capability = "token_dictionary.manage"

	// CapTenantOnboard gates provisioning a new tenant (M7 F7.2): a tenant-wide,
	// operator-grade action performed before any area/membership exists for the
	// tenant being created. Tenant-grade; enforced tier-2 in-tx via
	// authz.Require and backstopped by the tenants/INSERT DB tripwire arm.
	CapTenantOnboard Capability = "tenant.onboard"

	// CapTenantExport gates requesting a full tenant data export (M7 F7.3):
	// a tenant-wide, operator-grade action producing a self-evidencing export
	// artifact. Tenant-grade; enforced tier-2 in-tx via authz.Require and
	// backstopped by the tenant_lifecycle_jobs/INSERT DB tripwire arm.
	CapTenantExport Capability = "tenant.export"

	// CapTenantErase gates requesting irreversible tenant erasure /
	// crypto-shred (M7 F7.3). Seeded to system_admin ONLY. Tenant-grade;
	// enforced tier-2 in-tx via authz.Require and backstopped by the
	// tenant_lifecycle_jobs/INSERT DB tripwire arm.
	CapTenantErase Capability = "tenant.erase"
)

// ADR 0022 Phase 10 (F2) — the four caps Phase 8 minted (doc.edit_draft,
// doc.reconstruct, workflow.instance.cancel, doc.view_published) were redundant:
// grant sets and HTTP surfaces identical to document.edit / document.view, zero
// authorization differentiation. Merged back into the canonical caps; their
// tier-2 call sites now Require document.edit / document.view. This reverses
// Phase 8's registry over-modeling, NOT its security fix (the surfaces stay
// enforced, now by the canonical cap).

var validCapabilities = map[Capability]struct{}{
	CapDocumentView:                {},
	CapDocumentCreate:              {},
	CapDocumentEdit:                {},
	CapDocumentSubmit:              {},
	CapDocumentSignoff:             {},
	CapDocumentPublish:             {},
	CapDocumentObsolete:            {},
	CapDocumentSupersede:           {},
	CapDocumentReview:              {},
	CapTemplateView:                {},
	CapTemplateCreate:              {},
	CapTemplateEdit:                {},
	CapTemplateSubmit:              {},
	CapTemplateReview:              {},
	CapTemplateApprove:             {},
	CapTemplatePublish:             {},
	CapTemplateArchive:             {},
	CapTemplateManage:              {},
	CapControlledDocumentCreate:    {},
	CapControlledDocumentObsolete:  {},
	CapControlledDocumentSupersede: {},
	CapTaxonomyView:                {},
	CapTaxonomyManage:              {},
	CapMembershipView:              {},
	CapMembershipManage:            {},
	CapRouteManage:                 {},
	CapUserView:                    {},
	CapUserManage:                  {},
	CapMetricsView:                 {},
	CapAuditRead:                   {},
	CapSessionManage:               {},
	CapDistributionRead:            {},
	CapNotificationRead:            {},
	CapTokenView:                   {},
	CapTokenDictionaryManage:       {},
	CapTenantOnboard:               {},
	CapTenantExport:                {},
	CapTenantErase:                 {},
}

// IsValidCapability reports whether cap is one of the registered constants in
// validCapabilities.
func IsValidCapability(cap Capability) bool {
	_, ok := validCapabilities[cap]
	return ok
}

// AllCapabilities returns every known capability. Used by callers that need
// to express a "holds everything" set (e.g. system_admin bypass).
func AllCapabilities() []Capability {
	out := make([]Capability, 0, len(validCapabilities))
	for cap := range validCapabilities {
		out = append(out, cap)
	}
	return out
}

// MustCapability converts raw to a Capability, panicking if it is not one of
// the registered constants. Intended for compile-time-known literals (test
// setup, static registrations) — never for values derived from external input.
func MustCapability(raw string) Capability {
	capability := Capability(raw)
	if !IsValidCapability(capability) {
		panic(fmt.Sprintf("invalid capability: %q", raw))
	}
	return capability
}
