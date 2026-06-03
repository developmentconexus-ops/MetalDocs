package domain

import (
	"strings"
	"testing"
)

// TestCanonicalRolesShape locks the 8-role canonical catalogue: codes,
// categories, and that every Role const surfaces as a canonical entry.
func TestCanonicalRolesShape(t *testing.T) {
	roles := CanonicalRoles()
	if got, want := len(roles), 8; got != want {
		t.Fatalf("CanonicalRoles() size = %d, want %d", got, want)
	}

	wantCategory := map[Role]RoleCategory{
		RoleSystemAdmin: RoleCategoryTenant,
		RoleApprover:    RoleCategoryTenant,
		RoleAuthor:      RoleCategoryTenant,
		RoleEditor:      RoleCategoryTenant,
		RoleViewer:      RoleCategoryTenant,
		RoleSigner:      RoleCategoryArea,
		RoleAreaAdmin:   RoleCategoryArea,
		RoleQmsAdmin:    RoleCategoryArea,
	}
	seen := make(map[Role]struct{}, len(roles))
	for _, role := range roles {
		seen[role.Code] = struct{}{}
		if !IsValidRole(role.Code) {
			t.Errorf("role %q in canonical catalogue but not in validRoles", role.Code)
		}
		if got, want := role.Category, wantCategory[role.Code]; got != want {
			t.Errorf("role %q category = %q, want %q", role.Code, got, want)
		}
		if strings.TrimSpace(role.Label) == "" {
			t.Errorf("role %q has empty label", role.Code)
		}
		if strings.TrimSpace(role.Description) == "" {
			t.Errorf("role %q has empty description", role.Code)
		}
	}
	for code := range validRoles {
		if _, ok := seen[code]; !ok {
			t.Errorf("validRoles contains %q but CanonicalRoles() does not", code)
		}
	}
}

// TestCanonicalRolesReturnsCopy verifies mutation of the returned slice
// does not corrupt the package-internal source of truth.
func TestCanonicalRolesReturnsCopy(t *testing.T) {
	first := CanonicalRoles()
	first[0].Label = "MUTATED"
	second := CanonicalRoles()
	if second[0].Label == "MUTATED" {
		t.Fatalf("CanonicalRoles() leaked internal slice; second call sees mutation")
	}
}

// TestCapabilityCatalogShape locks one descriptor per Capability const,
// category derivation from the code prefix, and pt-BR description presence.
func TestCapabilityCatalogShape(t *testing.T) {
	caps := CapabilityCatalog()
	if got, want := len(caps), len(validCapabilities); got != want {
		t.Fatalf("CapabilityCatalog() size = %d, want %d (one descriptor per Capability const)", got, want)
	}

	seen := make(map[Capability]struct{}, len(caps))
	for _, descriptor := range caps {
		if _, dup := seen[descriptor.Code]; dup {
			t.Errorf("duplicate capability in catalogue: %q", descriptor.Code)
		}
		seen[descriptor.Code] = struct{}{}

		if !IsValidCapability(descriptor.Code) {
			t.Errorf("capability %q in catalogue but not in validCapabilities", descriptor.Code)
		}
		if strings.TrimSpace(descriptor.Description) == "" {
			t.Errorf("capability %q has empty description", descriptor.Code)
		}
		if strings.TrimSpace(descriptor.Category) == "" {
			t.Errorf("capability %q has empty category", descriptor.Code)
		}

		code := string(descriptor.Code)
		if i := strings.IndexByte(code, '.'); i > 0 {
			if got, want := descriptor.Category, code[:i]; got != want {
				t.Errorf("capability %q category = %q, want %q (derived from prefix)", descriptor.Code, got, want)
			}
		}
	}
	for cap := range validCapabilities {
		if _, ok := seen[cap]; !ok {
			t.Errorf("validCapabilities contains %q but CapabilityCatalog() does not", cap)
		}
	}
}

// TestCapabilityCatalogSortedByCode asserts deterministic output ordering.
func TestCapabilityCatalogSortedByCode(t *testing.T) {
	caps := CapabilityCatalog()
	for i := 1; i < len(caps); i++ {
		if caps[i-1].Code >= caps[i].Code {
			t.Fatalf("CapabilityCatalog() not sorted at index %d: %q >= %q", i, caps[i-1].Code, caps[i].Code)
		}
	}
}

func TestCapabilityCategoryDerivation(t *testing.T) {
	tests := map[Capability]string{
		"document.view":              "document",
		"audit.read":                 "audit",
		"session.manage":             "session",
		"controlled_documents.create": "controlled_documents",
		"general":                    "general",
	}
	for cap, want := range tests {
		if got := capabilityCategory(cap); got != want {
			t.Errorf("capabilityCategory(%q) = %q, want %q", cap, got, want)
		}
	}
}
