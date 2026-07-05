package domain

// CapabilityScope is the declared enforcement grade of a capability: whether it
// is checked tenant-wide or against a specific process area. ADR 0022 Phase 2
// makes scope an explicit, typed property of each capability instead of a
// per-call-site string literal ("tenant" vs a real areaCode passed to tier-2
// authz.Require). This is the declaration; runtime alignment (passing the real
// area for area-grade ops) lands in later ADR 0022 phases.
type CapabilityScope string

const (
	// ScopeTenant — checked tenant-wide. Tier-2 callers pass the literal
	// "tenant" sentinel deliberately; the area filter is intentionally OFF.
	ScopeTenant CapabilityScope = "tenant-grade"
	// ScopeArea — checked against the resource's process area. Tier-2 callers
	// pass the real areaCode so the grant stays least-privilege.
	ScopeArea CapabilityScope = "area-grade"
)

// capabilityScopes classifies every capability in validCapabilities.
//
// Area-grade = document / approval / controlled-document WRITE capabilities plus
// membership.manage (administration of an area-scoped resource). Everything else
// is tenant-grade: all *.view reads, the whole template.* family, taxonomy.manage,
// user.manage, route.manage, metrics.view, audit.read, session.manage.
//
// TestEveryCapabilityClassified asserts this map covers the registry exactly
// (no unclassified cap, no stale entry); TestAreaGradeCapabilitySet locks the
// area-grade set so a silent reclassification fails the build.
//
// Runtime alignment (ADR 0022 Phase 7, completed): the Phase 3 gap is CLOSED.
// document.create and the controlled_documents.* caps are classified area-grade
// here AND their tier-2 call sites now pass the resource's real areaCode (not
// the literal "tenant"): documents/repository/repository.go:145 (docArea) and
// controlleddocuments/application/service.go:307,353,526,661,753
// (ProcessAreaCode / locked areaCode). Declaration and runtime now agree.
var capabilityScopes = map[Capability]CapabilityScope{
	// --- Area-grade: writes against an area-scoped resource ---
	CapDocumentCreate:              ScopeArea,
	CapDocumentEdit:                ScopeArea,
	CapDocumentSubmit:              ScopeArea,
	CapDocumentSignoff:             ScopeArea,
	CapDocumentPublish:             ScopeArea,
	CapDocumentObsolete:            ScopeArea,
	CapDocumentSupersede:           ScopeArea,
	CapControlledDocumentCreate:    ScopeArea,
	CapControlledDocumentObsolete:  ScopeArea,
	CapControlledDocumentSupersede: ScopeArea,
	CapMembershipManage:            ScopeArea,

	// --- Tenant-grade: tenant-wide authority ---
	CapDocumentView: ScopeTenant,
	// document.review (M6 F6.2, ADR 0069) is tenant-grade: the mark-reviewed
	// act is a tenant-wide governance action, not an area-scoped WRITE like
	// document.edit/publish/obsolete. Tier-2 passes the "tenant" sentinel.
	CapDocumentReview:   ScopeTenant,
	CapTemplateView:     ScopeTenant,
	CapTemplateCreate:   ScopeTenant,
	CapTemplateEdit:     ScopeTenant,
	CapTemplateSubmit:   ScopeTenant,
	CapTemplateReview:   ScopeTenant,
	CapTemplateApprove:  ScopeTenant,
	CapTemplatePublish:  ScopeTenant,
	CapTemplateArchive:  ScopeTenant,
	CapTemplateManage:   ScopeTenant,
	CapTaxonomyView:     ScopeTenant,
	CapTaxonomyManage:   ScopeTenant,
	CapMembershipView:   ScopeTenant,
	CapRouteManage:      ScopeTenant,
	CapUserView:         ScopeTenant,
	CapUserManage:       ScopeTenant,
	CapMetricsView:      ScopeTenant,
	CapAuditRead:        ScopeTenant,
	CapSessionManage:    ScopeTenant,
	CapDistributionRead: ScopeTenant,
	// Tenant-grade in this enum (no area filter). Own-rows "self-scope" is
	// enforced by the SQL predicate (recipient_user_id = caller) in F3.2, not by
	// the area-vs-tenant grade declared here (ADR-0043).
	CapNotificationRead: ScopeTenant,
	// Token dictionary (SP-1) — tenant-wide, not area-scoped.
	CapTokenView:             ScopeTenant,
	CapTokenDictionaryManage: ScopeTenant,
	// tenant.onboard (M7 F7.2) — tenant-wide provisioning action, no area
	// exists yet for a tenant being created. Tier-2 passes the "tenant" sentinel.
	CapTenantOnboard: ScopeTenant,
	// tenant.export / tenant.erase (M7 F7.3, ADR 0070) — tenant-wide lifecycle
	// actions, no area filter applies.
	CapTenantExport: ScopeTenant,
	CapTenantErase:  ScopeTenant,
}

// ScopeOf returns the declared scope of a capability and whether it is
// classified. Unknown capabilities return ("", false).
func ScopeOf(cap Capability) (CapabilityScope, bool) {
	s, ok := capabilityScopes[cap]
	return s, ok
}

// IsAreaGrade reports whether cap is area-scoped (tier-2 must pass a real area).
func IsAreaGrade(cap Capability) bool {
	return capabilityScopes[cap] == ScopeArea
}
