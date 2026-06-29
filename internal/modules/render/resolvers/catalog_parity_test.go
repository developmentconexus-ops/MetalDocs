package resolvers_test

import (
	"testing"

	"metaldocs/internal/modules/render/domain"
	"metaldocs/internal/modules/render/resolvers"
)

// Every registered resolver MUST have a catalog descriptor, and every catalog
// descriptor MUST have a registered resolver. This makes catalog<->resolver
// drift impossible at the source (the defect that hid approval_date pre-2026-06-29).
func TestCatalogResolverParity(t *testing.T) {
	reg := resolvers.NewRegistry()
	resolvers.RegisterBuiltins(reg)
	registered := reg.Known() // map[key]version

	catalog := make(map[string]bool)
	for _, e := range domain.ComputedCatalog() {
		catalog[e.Key] = true
	}

	for key := range registered {
		if !catalog[key] {
			t.Errorf("resolver %q has no render/domain catalog descriptor", key)
		}
	}
	for key := range catalog {
		if _, ok := registered[key]; !ok {
			t.Errorf("catalog token %q has no registered resolver", key)
		}
	}
}
