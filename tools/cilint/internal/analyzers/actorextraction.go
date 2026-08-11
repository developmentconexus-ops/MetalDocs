package analyzers

import (
	"go/ast"
	"go/token"
	"strings"
)

// actorExtractionAllow is the inline directive to suppress a finding on the
// offending line. A suppression must say WHY the site is exempt; the two rules
// below both encode a fail-closed invariant, so an unexplained opt-out is the
// defect this analyzer exists to catch.
const actorExtractionAllow = "//cilint:allow-actor-extraction"

// The import paths that make up the two forbidden shapes. Both rules resolve by
// PATH, never by the local qualifier a file happens to bind them to: an author
// who writes `import id "metaldocs/internal/modules/iam/domain"` and then calls
// `id.UserIDFromContext(ctx)` has written the same call, and a guard keyed to
// the spelling is a guard the author can opt out of by renaming.
const (
	iamDomainImportPath = "metaldocs/internal/modules/iam/domain"
	authnImportPath     = "metaldocs/internal/platform/authn"
)

// actorAccessorName is the low-level identity-storage accessor. It returns only
// a string, so absence is indistinguishable from a real actor at the call site
// — which is exactly why runtime code must not call it.
const actorAccessorName = "UserIDFromContext"

// authnPresenceFuncs are the canonical accessors whose SECOND result carries
// presence (a bool for UserIDFromContext, an error for RequireUserID).
// Discarding it converts "no authenticated principal" back into a silent "".
var authnPresenceFuncs = map[string]string{
	"UserIDFromContext": "presence bool",
	"RequireUserID":     "error",
}

// actorExtractionPlumbing is the ONE runtime file allowed to call the low-level
// iam/domain accessor: the canonical consumer API is implemented on top of it.
// The seam is a single file rather than a package prefix so that widening it is
// a visible, reviewable edit here instead of a new file quietly landing inside
// an allowed directory.
const actorExtractionPlumbing = "internal/platform/authn/context.go"

// ActorExtraction enforces the A3.3 property: absence of an authenticated actor
// must be an explicit decision, never a silent "" that keeps travelling into
// authorization, audit attribution, rate limiting, an application service, a
// domain mutation, or persistence.
//
// Two competing accessors exist, and that is the whole problem:
//
//   - iamdomain.UserIDFromContext(ctx) string        — low-level context
//     storage. Returns "" for "no actor", which is also a perfectly valid
//     string to pass along. Absence disappears at the call site.
//   - authn.UserIDFromContext(ctx) (string, bool)    — canonical consumer API.
//     Absence is a separate result the caller has to look at.
//
// Rule 1 (low-level consumer ban): outside actorExtractionPlumbing, runtime code
// may not call the low-level accessor at all. Making the fail-open accessor
// unreachable is stronger than asking every consumer to remember the check.
//
// Rule 2 (ignored presence result): runtime code may not discard the canonical
// accessor's second result. `actorID, _ := authn.UserIDFromContext(ctx)`
// compiles, yields "" on absence, and reintroduces exactly the fail-open shape
// Rule 1 removed — the ban would otherwise be one underscore wide.
//
// The rules deliberately match syntax, not data flow: every discard in this
// repository is a literal `_` in an assignment's second slot, and a general
// dataflow analysis would buy nothing the source actually needs.
//
// Test files are outside the walker (collectGoFiles skips _test.go): a test that
// CONSTRUCTS an auth context, or that asserts the storage primitive's own
// behavior, is not a runtime consumer. tools/ and scripts/ are outside
// inRuntimeTree for the same reason problemwriter excludes them — this file
// names the accessor as a constant, it does not extract an actor.
func ActorExtraction(files []string) []Finding {
	var out []Finding
	fset := token.NewFileSet()
	for _, path := range files {
		out = append(out, scanActorExtractionFile(fset, path)...)
	}
	return out
}

func scanActorExtractionFile(fset *token.FileSet, path string) []Finding {
	slashed := strings.ReplaceAll(path, "\\", "/")
	if !inRuntimeTree(slashed) || strings.HasSuffix(slashed, ".gen.go") {
		return nil
	}
	isPlumbing := strings.HasSuffix(slashed, actorExtractionPlumbing)

	_, raw := parseFile(fset, path)
	if raw == nil {
		return nil
	}
	f := raw.(*ast.File)
	src := readSource(path)
	aliases := importAliases(f)

	var out []Finding
	ast.Inspect(f, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.CallExpr:
			if isPlumbing {
				return true
			}
			if isSelector(node.Fun, aliases, iamDomainImportPath, actorAccessorName) {
				out = appendActorFinding(out, fset, src, path, node.Pos(),
					"calls "+iamDomainImportPath+"."+actorAccessorName+", the low-level identity-storage "+
						"accessor that returns \"\" for a missing actor (A3.3); runtime consumers must use "+
						"authn.UserIDFromContext (presence-aware) or authn.RequireUserID (fails closed) so "+
						"absence is an explicit decision instead of an empty string travelling downstream")
			}
		case *ast.AssignStmt:
			if msg, pos, ok := ignoredPresenceViolation(node, aliases); ok {
				out = appendActorFinding(out, fset, src, path, pos, msg)
			}
		}
		return true
	})
	return out
}

// ignoredPresenceViolation matches `x, _ := authn.UserIDFromContext(ctx)` and
// `x, _ = authn.RequireUserID(ctx)` — a two-value assignment from a single call
// to a canonical accessor whose second slot is the blank identifier.
func ignoredPresenceViolation(node *ast.AssignStmt, aliases map[string]string) (msg string, pos token.Pos, ok bool) {
	if len(node.Lhs) != 2 || len(node.Rhs) != 1 {
		return "", token.NoPos, false
	}
	blank, isIdent := node.Lhs[1].(*ast.Ident)
	if !isIdent || blank.Name != "_" {
		return "", token.NoPos, false
	}
	call, isCall := node.Rhs[0].(*ast.CallExpr)
	if !isCall {
		return "", token.NoPos, false
	}
	for name, result := range authnPresenceFuncs {
		if isSelector(call.Fun, aliases, authnImportPath, name) {
			return "discards the " + result + " of authn." + name + " (A3.3); the actor is then \"\" whenever " +
				"the request carries no authenticated principal, which is the fail-open shape the canonical " +
				"accessor exists to prevent — read the second result and fail explicitly, or use " +
				"authn.RequireUserID and propagate its error", node.Pos(), true
		}
	}
	return "", token.NoPos, false
}

func appendActorFinding(out []Finding, fset *token.FileSet, src, path string, pos token.Pos, msg string) []Finding {
	p := fset.Position(pos)
	if strings.Contains(getLine(src, p.Line), actorExtractionAllow) {
		return out
	}
	return append(out, Finding{
		Analyzer: "actorextraction",
		File:     path,
		Line:     p.Line,
		Message:  msg,
	})
}
