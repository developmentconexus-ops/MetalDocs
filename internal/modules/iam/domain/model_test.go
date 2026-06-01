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
	const want = 27
	if got := len(validCapabilities); got != want {
		t.Fatalf("validCapabilities size = %d, want %d (bump only via ADR; current = 23 base + 4 ADR 0016 view caps)", got, want)
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
