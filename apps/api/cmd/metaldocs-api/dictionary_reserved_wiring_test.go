package main

// TestDictionaryReservedWiring proves that the resolver registry built in
// buildTemplatesModule (resolvers.NewRegistry + RegisterBuiltins) contains the
// native token names that the D5 guard checks. If someone drops the registry
// argument from templatesapp.New again, s.resolvers becomes nil, the guard is
// silently disabled, and this test would not catch that — but the test DOES
// prove the backing registry is non-empty and holds the keys expected by D5,
// so any regression that changes the builtin set is immediately visible.
//
// Module-boundary note: this test lives in package main (the composition root)
// where render/resolvers is already imported. Templates non-test code never
// imports render — the registry is only passed across the boundary here.

import (
	"testing"

	"metaldocs/internal/modules/render/resolvers"
)

func TestDictionaryReservedWiringRegistryContainsBuiltins(t *testing.T) {
	t.Parallel()

	// Reproduce exactly the pattern used in buildTemplatesModule.
	reg := resolvers.NewRegistry()
	resolvers.RegisterBuiltins(reg)
	known := reg.Known()

	if len(known) == 0 {
		t.Fatal("D5 guard: builtins registry is empty; templates reserved-name guard would be silently disabled in production")
	}

	// These are the 8 static builtins defined by RegisterBuiltins. If any is
	// absent the D5 guard won't fire for that name.
	required := []string{
		"author",
		"doc_code",
		"doc_title",
		"revision_number",
		"effective_date",
		"approvers",
		"controlled_by_area",
		"approval_date",
	}
	for _, name := range required {
		if _, ok := known[name]; !ok {
			t.Errorf("D5 guard: builtin %q missing from registry; dictionary placeholder named %q would not be rejected at schema-save time", name, name)
		}
	}
}
