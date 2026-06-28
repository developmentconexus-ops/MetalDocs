package domain

import (
	"fmt"
	"sort"
	"strings"
)

type Role string

// Canonical tenant roles (8). PR-5 will move this catalog to a dedicated file
// (`canonical_roles.go`) when the Roles & Capabilities matrix lands; the three
// area-only roles (signer / area_admin / qms_admin) are added here in PR-4 so
// People-tab invite validation can accept area-membership.role values that
// match the OpenAPI UserRole enum. Source: wiki/modules/iam.md §canonical-roles.
const (
	RoleApprover    Role = "approver"
	RoleAreaAdmin   Role = "area_admin"
	RoleAuthor      Role = "author"
	RoleEditor      Role = "editor"
	RoleQmsAdmin    Role = "qms_admin"
	RoleSigner      Role = "signer"
	RoleSystemAdmin Role = "system_admin"
	RoleViewer      Role = "viewer"
)

var validRoles = map[Role]struct{}{
	RoleApprover:    {},
	RoleAreaAdmin:   {},
	RoleAuthor:      {},
	RoleEditor:      {},
	RoleQmsAdmin:    {},
	RoleSigner:      {},
	RoleSystemAdmin: {},
	RoleViewer:      {},
}

func IsValidRole(role Role) bool {
	_, ok := validRoles[role]
	return ok
}

func ParseRole(raw string) (Role, error) {
	role := Role(strings.ToLower(strings.TrimSpace(raw)))
	if role == "" || !IsValidRole(role) {
		return "", ErrInvalidRole
	}
	return role, nil
}

// areaRoles is the subset of canonical roles assignable as an AREA membership
// (a public.user_process_areas row). It is every canonical role EXCEPT
// system_admin: system_admin is a tenant-wide tier-1 role that bypasses tier-2
// and is never an area membership. This set is the Go single source of truth for
// area-assignable roles and mirrors the user_process_areas role CHECK constraint
// in the DB. Consumers that resolve actors by an area role (e.g. the approval
// module's stage required_role, joined against user_process_areas.role) bind
// against this set so a role no user can ever hold cannot be configured
// (ADR 0022 — role strings are bound to the registry, never free text).
var areaRoles = map[Role]struct{}{
	RoleApprover:  {},
	RoleAreaAdmin: {},
	RoleAuthor:    {},
	RoleEditor:    {},
	RoleQmsAdmin:  {},
	RoleSigner:    {},
	RoleViewer:    {},
}

// IsAreaRole reports whether role can be held as an area membership
// (user_process_areas). system_admin is intentionally excluded — it is tenant-wide.
func IsAreaRole(role Role) bool {
	_, ok := areaRoles[role]
	return ok
}

// AreaRoles returns the canonical area-assignable roles, sorted — for validation
// messages and diagnostics.
func AreaRoles() []Role {
	out := make([]Role, 0, len(areaRoles))
	for r := range areaRoles {
		out = append(out, r)
	}
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}

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

	CapTemplateView    Capability = "template.view"
	CapTemplateCreate  Capability = "template.create"
	CapTemplateEdit    Capability = "template.edit"
	CapTemplateSubmit  Capability = "template.submit"
	CapTemplateReview  Capability = "template.review"
	CapTemplateApprove Capability = "template.approve"
	CapTemplatePublish Capability = "template.publish"
	CapTemplateArchive Capability = "template.archive"

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
	CapTemplateView:                {},
	CapTemplateCreate:              {},
	CapTemplateEdit:                {},
	CapTemplateSubmit:              {},
	CapTemplateReview:              {},
	CapTemplateApprove:             {},
	CapTemplatePublish:             {},
	CapTemplateArchive:             {},
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
}

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

func MustCapability(raw string) Capability {
	capability := Capability(raw)
	if !IsValidCapability(capability) {
		panic(fmt.Sprintf("invalid capability: %q", raw))
	}
	return capability
}
