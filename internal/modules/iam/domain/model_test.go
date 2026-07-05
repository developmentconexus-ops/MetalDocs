package domain

import (
	"testing"
)

func TestIsValidRole(t *testing.T) {
	if !IsValidRole(RoleEditor) {
		t.Fatal("RoleEditor must be valid")
	}
	if IsValidRole(Role("invalid")) {
		t.Fatal("unexpected valid result for invalid role")
	}
}

func TestIsAreaRole(t *testing.T) {
	// Every canonical role except system_admin is area-assignable.
	for _, r := range []Role{RoleApprover, RoleAreaAdmin, RoleAuthor, RoleEditor, RoleQmsAdmin, RoleSigner, RoleViewer} {
		if !IsAreaRole(r) {
			t.Fatalf("%q must be an area role", r)
		}
		if !IsValidRole(r) {
			t.Fatalf("area role %q must also be a valid role", r)
		}
	}
	// system_admin is tenant-wide, never an area membership.
	if IsAreaRole(RoleSystemAdmin) {
		t.Fatal("system_admin must NOT be an area role (tenant-wide tier-1)")
	}
	// Decommissioned phantom must not be an area role.
	if IsAreaRole(Role("reviewer")) {
		t.Fatal("reviewer is decommissioned and must not be an area role")
	}
	if IsAreaRole(Role("nonsense")) {
		t.Fatal("unknown role must not be an area role")
	}
}

func TestAreaRoleRegistrySize(t *testing.T) {
	// Locked count: 7 canonical area roles (the 8 canonical roles minus
	// system_admin). Mirrors the user_process_areas role CHECK. Bump this
	// deliberately when the area-role set changes — a silent add/remove is a bug.
	const wantAreaRoles = 7
	if got := len(AreaRoles()); got != wantAreaRoles {
		t.Fatalf("AreaRoles count = %d, want %d (update intentionally if the area-role set changed)", got, wantAreaRoles)
	}
}

func TestAreaRolesMatchesRegistryMinusSystemAdmin(t *testing.T) {
	// Locked count: 8 canonical roles total (mirrors TestAreaRoleRegistrySize's
	// 7 area roles = 8 - system_admin). validRoles is now unexported inside
	// iamtypes (ARC-06 move), so this asserts against the same canonical-role
	// count rather than reaching into the platform package's private map.
	const wantCanonicalRoles = 8
	got := AreaRoles()
	if len(got) != wantCanonicalRoles-1 {
		t.Fatalf("AreaRoles must equal canonical roles minus system_admin: got %d, want %d", len(got), wantCanonicalRoles-1)
	}
	for _, r := range got {
		if r == RoleSystemAdmin {
			t.Fatal("AreaRoles must not include system_admin")
		}
		if !IsValidRole(r) {
			t.Fatalf("AreaRoles entry %q is not a valid role", r)
		}
	}
	// Sorted, deterministic.
	for i := 1; i < len(got); i++ {
		if got[i-1] >= got[i] {
			t.Fatalf("AreaRoles must be sorted ascending: %v", got)
		}
	}
}

func TestCapabilityInvariants(t *testing.T) {
	if !IsValidCapability(CapMembershipManage) {
		t.Fatal("CapMembershipManage must be valid")
	}
	if IsValidCapability(Capability("not.real")) {
		t.Fatal("unexpected valid result for invalid capability")
	}
	if got := MustCapability("membership.manage"); got != CapMembershipManage {
		t.Fatalf("MustCapability = %q, want %q", got, CapMembershipManage)
	}
	defer func() {
		if recover() == nil {
			t.Fatal("MustCapability must panic for invalid value")
		}
	}()
	_ = MustCapability("invalid.capability")
}

// Locks the registry size so silent deletions or duplicate inserts surface
// as a test failure. Bump intentionally when caps are added/removed via ADR.
func TestCapabilityRegistrySize(t *testing.T) {
	const want = 35
	if got := len(validCapabilities); got != want {
		t.Fatalf("validCapabilities size = %d, want %d (bump only via ADR; current = 23 base + 4 ADR 0016 view caps + 1 ADR 0019 session.manage + ADR 0022 P1: -2 dead workflow.* +3 promoted doc.publish/doc.obsolete/template.archive + ADR 0022 P8: +4 registered phantoms + ADR 0022 P10: -4 redundant phantoms merged into document.edit/document.view doc.edit_draft/doc.reconstruct/doc.view_published/workflow.instance.cancel + ADR 0022 P12: doc.publish/doc.obsolete/doc.supersede renamed to document.* — count unchanged + ADR 0042 M2/F2.1c: +1 distribution.read + ADR 0043 M3/F3.1: +1 notification.read + SP-1: +2 token.view/token_dictionary.manage + ADR 0051: +1 template.manage + ADR 0069 M6/F6.2: +1 document.review)", got, want)
	}
	if got := len(AllCapabilities()); got != want {
		t.Fatalf("AllCapabilities() size = %d, want %d", got, want)
	}
}

func TestViewGradeCapabilitiesRegistered(t *testing.T) {
	viewCaps := []Capability{
		CapMetricsView,
		CapMembershipView,
		CapUserView,
		CapTaxonomyView,
	}
	wantValues := map[Capability]string{
		CapMetricsView:    "metrics.view",
		CapMembershipView: "membership.view",
		CapUserView:       "user.view",
		CapTaxonomyView:   "taxonomy.view",
	}
	for _, c := range viewCaps {
		if !IsValidCapability(c) {
			t.Errorf("view cap %q must be in validCapabilities", c)
		}
		if string(c) != wantValues[c] {
			t.Errorf("cap string drift: got %q, want %q", string(c), wantValues[c])
		}
	}
}
