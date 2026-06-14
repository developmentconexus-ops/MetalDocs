package test

import (
	"os"
	"strings"
)

// E2EEnabled reports whether the test-only E2E seed/reset/governance endpoints
// are active. METALDOCS_E2E must equal "1" after trimming surrounding
// whitespace; only the literal "1" enables them. Single source of truth shared
// by the composition root (which decides whether to MOUNT the routes) and the
// seed handlers (which re-check the flag on every request as defense in depth),
// so the two can never disagree on a value like " 1".
func E2EEnabled() bool {
	return strings.TrimSpace(os.Getenv("METALDOCS_E2E")) == "1"
}
