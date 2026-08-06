package main

import (
	"net/http"
	"sort"
	"testing"

	"metaldocs/internal/platform/httprouter"
)

// TestGeneratedKeysMatchCodegenPatterns proves the generator's key formula
// and oapi-codegen's HandlerWithOptions registration formula produce the same
// bytes for every operation that goes through codegen. It is exhaustive by
// construction: it walks httpSurface, not a sampled list.
//
// The boot assertion (surface.go's assertSurface, wired for real in
// TestRealPublishersSatisfyTheAssertion) proves the same thing on every real
// boot; this test proves it in CI without a database, iterating the real
// `publishers` list — buildTestRouteHandlers (permissions_test.go) and
// buildPublishers (publishers.go) are production's own construction/mount
// path, reused here rather than re-implemented, so this test observes
// exactly what main() mounts.
func TestGeneratedKeysMatchCodegenPatterns(t *testing.T) {
	mux := http.NewServeMux()
	pubs := buildPublishers(buildTestRouteHandlers()) // nil services are safe at mount time

	mounted := map[string]bool{}
	for _, p := range pubs {
		rec := httprouter.NewRecorder(mux)
		p.Mount(rec)
		for _, pattern := range rec.Patterns() {
			mounted[pattern] = true
		}
	}

	var missing []string
	for key := range httpSurface {
		if !mounted[key] {
			missing = append(missing, key)
		}
	}
	sort.Strings(missing)
	if len(missing) > 0 {
		t.Fatalf("%d declared keys were not registered by codegen with the same bytes:\n%v", len(missing), missing)
	}
}
