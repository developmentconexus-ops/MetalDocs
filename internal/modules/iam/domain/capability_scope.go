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
// NOTE (runtime gap, ADR 0022 Phase 3): document.create and the
// controlled_documents.* caps are classified area-grade here (the declared
// target) but their tier-2 call sites still pass the literal "tenant" today.
// Phase 2 only declares scope; aligning those call sites to a real areaCode is
// later-phase work and out of scope for this change.
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
	CapDocumentView:     ScopeTenant,
	CapTemplateView:     ScopeTenant,
	CapTemplateCreate:   ScopeTenant,
	CapTemplateEdit:     ScopeTenant,
	CapTemplateSubmit:   ScopeTenant,
	CapTemplateReview:   ScopeTenant,
	CapTemplateApprove:  ScopeTenant,
	CapTemplatePublish:  ScopeTenant,
	CapTemplateArchive:  ScopeTenant,
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
