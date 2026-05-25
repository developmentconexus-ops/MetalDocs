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
