package domain

import "testing"

// TestEveryCapabilityClassified is the ADR 0022 Phase 2 binding: scope is a
// declared property of EVERY capability. Fails if a registry cap is missing a
// classification, if capabilityScopes carries a stale entry for a removed cap,
// or if any entry holds an unknown grade. A new cap added to validCapabilities
// without a scope entry turns into a red build.
func TestEveryCapabilityClassified(t *testing.T) {
	for cap := range validCapabilities {
		if _, ok := capabilityScopes[cap]; !ok {
			t.Errorf("capability %q is in validCapabilities but not classified in capabilityScopes (add it as ScopeTenant or ScopeArea)", cap)
		}
	}
	for cap := range capabilityScopes {
		if !IsValidCapability(cap) {
			t.Errorf("capabilityScopes classifies %q which is not in validCapabilities (remove the stale entry)", cap)
		}
	}
	for cap, s := range capabilityScopes {
		if s != ScopeTenant && s != ScopeArea {
			t.Errorf("capability %q has unknown scope %q (want ScopeTenant or ScopeArea)", cap, s)
		}
	}
}

// TestAreaGradeCapabilitySet locks the exact area-grade set so a silent
// reclassification (e.g. flipping a tenant-grade cap to area-grade or vice
// versa) fails the build. Area-grade = document/approval/controlled-document
// write caps + membership.manage.
func TestAreaGradeCapabilitySet(t *testing.T) {
	wantArea := map[Capability]struct{}{
		CapDocumentCreate:              {},
		CapDocumentEdit:                {},
		CapDocumentSubmit:              {},
		CapDocumentSignoff:             {},
		CapDocumentObsolete:            {},
		CapDocumentSupersede:           {},
		CapControlledDocumentCreate:    {},
		CapControlledDocumentObsolete:  {},
		CapControlledDocumentSupersede: {},
		CapMembershipManage:            {},
		// ADR 0022 Phase 10 (F2): the phantom area-grade caps doc.edit_draft /
		// doc.reconstruct / workflow.instance.cancel were merged into the canonical
		// CapDocumentEdit (already in this set) — area-grade set 14 → 11.
		// M2b F3: approval.review acts on a stage in a specific process area — 11 → 12.
		CapApprovalReview: {},
	}
	for cap := range validCapabilities {
		_, want := wantArea[cap]
		if got := IsAreaGrade(cap); got != want {
			t.Errorf("capability %q: IsAreaGrade = %v, want %v", cap, got, want)
		}
	}
	if got, want := countScope(ScopeArea), len(wantArea); got != want {
		t.Fatalf("area-grade count = %d, want %d", got, want)
	}
}

// TestScopeOf covers the lookup contract: classified caps report their grade,
// unknown caps report (_, false).
func TestScopeOf(t *testing.T) {
	if s, ok := ScopeOf(CapMembershipManage); !ok || s != ScopeArea {
		t.Fatalf("ScopeOf(membership.manage) = (%q, %v), want (area-grade, true)", s, ok)
	}
	if s, ok := ScopeOf(CapUserManage); !ok || s != ScopeTenant {
		t.Fatalf("ScopeOf(user.manage) = (%q, %v), want (tenant-grade, true)", s, ok)
	}
	if _, ok := ScopeOf(Capability("not.real")); ok {
		t.Fatal("ScopeOf(not.real) reported classified; want false")
	}
}

func countScope(want CapabilityScope) int {
	n := 0
	for _, s := range capabilityScopes {
		if s == want {
			n++
		}
	}
	return n
}
