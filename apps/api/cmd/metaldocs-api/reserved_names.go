package main

import "metaldocs/internal/modules/render/resolvers"

// reservedNamesFromRegistry adapts the render resolver registry's known native
// token keys to the tokens.ReservedNames consumer port. Built at the composition
// root so the tokens module never imports render (SP-2 §5.1, §11). The 8 builtins
// are static, so a dedicated registry instance here is cheap and authoritative.
type reservedNamesFromRegistry struct {
	known map[string]int
}

func newReservedNamesFromRegistry() reservedNamesFromRegistry {
	reg := resolvers.NewRegistry()
	resolvers.RegisterBuiltins(reg)
	return reservedNamesFromRegistry{known: reg.Known()}
}

func (r reservedNamesFromRegistry) IsReserved(name string) bool {
	_, ok := r.known[name]
	return ok
}
