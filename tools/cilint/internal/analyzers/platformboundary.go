package analyzers

import (
	"go/ast"
	"go/token"
	"path/filepath"
	"strconv"
	"strings"
)

const modulesImportPrefix = "metaldocs/internal/modules"

// platformBoundaryAllowed is the FROZEN per-package disposition for platform
// packages that still import internal/modules (Wave 1.10 baseline, F-06a
// lock-in / REQ-TOP-2). Each entry must carry a removal trigger; new platform
// packages, and new module imports in packages not listed here, are
// CI-blocked. Do NOT add entries without an architecture decision — the
// target state is an empty map.
var platformBoundaryAllowed = map[string]string{
	// Composition-adjacent: builds module dependency graphs for the three
	// binaries. Disposition: composition-root concern; revisit when Wave 2
	// boundary extractions (F-06b/c/d) shrink it.
	"internal/platform/bootstrap": "composition-root wiring",
	// Loads auth/session config consumed by the auth module. Disposition:
	// Wave 2 boundary review (composition-adjacent per Wave 0 round-2).
	"internal/platform/authn": "composition-adjacent config",
	// Render pipeline coupling. Disposition: Wave 2 F-06 extraction family.
	"internal/platform/docgenv2": "pending Wave 2 boundary extraction",
	// Presigner reads module domain types. Disposition: Wave 2 F-06 family.
	"internal/platform/objectstore": "pending Wave 2 boundary extraction",
}

// PlatformBoundary flags imports of internal/modules from internal/platform
// packages (REQ-TOP-2: platform is module-agnostic infrastructure; the legal
// dependency direction is modules -> platform). The fix for a violation is
// dependency inversion: inject a callback or define a port, as done for
// platform/observability in Wave 0.6.
func PlatformBoundary(files []string) []Finding {
	var out []Finding
	fset := token.NewFileSet()

	for _, path := range files {
		slashed := filepath.ToSlash(path)
		idx := strings.Index(slashed, "internal/platform/")
		if idx < 0 {
			continue
		}
		pkgDir := filepath.ToSlash(filepath.Dir(slashed[idx:]))
		if allowedPlatformPackage(pkgDir) {
			continue
		}
		_, raw := parseFile(fset, path)
		if raw == nil {
			continue
		}
		f := raw.(*ast.File)

		for _, imp := range f.Imports {
			val, err := strconv.Unquote(imp.Path.Value)
			if err != nil || !strings.HasPrefix(val, modulesImportPrefix) {
				continue
			}
			pos := fset.Position(imp.Pos())
			out = append(out, Finding{
				Analyzer: "platformboundary",
				File:     path,
				Line:     pos.Line,
				Message:  "internal/platform must not import internal/modules (" + val + "); invert the dependency with a callback or port (REQ-TOP-2)",
			})
		}
	}
	return out
}

// allowedPlatformPackage reports whether pkgDir (slash-separated, rooted at
// internal/platform/...) is — or is nested under — a frozen-baseline package.
func allowedPlatformPackage(pkgDir string) bool {
	for allowed := range platformBoundaryAllowed {
		if pkgDir == allowed || strings.HasPrefix(pkgDir, allowed+"/") {
			return true
		}
	}
	return false
}
