package analyzers

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
)

// actorExtractionAllow is the inline directive to suppress a finding on the
// offending line. A suppression must say WHY the site is exempt; the rules below
// all encode a fail-closed invariant, so an unexplained opt-out is the defect
// this analyzer exists to catch.
//
// The directive is only honoured when it is a REAL line comment that STARTS with
// this token and is followed by a non-empty rationale. Both halves are load
// bearing. A bare `//cilint:allow-actor-extraction` is the unexplained opt-out
// the paragraph above forbids, and a scan that matched the raw line text would
// also honour the token inside a string literal — a suppression mechanism that
// can be triggered by data is not a suppression mechanism.
const actorExtractionAllow = "//cilint:allow-actor-extraction"

// The import paths that make up the forbidden shapes. Every rule resolves by
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
// — which is exactly why runtime code must not reach it.
const actorAccessorName = "UserIDFromContext"

// authnPresenceFuncs are the canonical accessors whose SECOND result carries
// presence (a bool for UserIDFromContext, an error for RequireUserID).
// Discarding it converts "no authenticated principal" back into a silent "".
var authnPresenceFuncs = map[string]string{
	"UserIDFromContext": "presence bool",
	"RequireUserID":     "error",
}

// actorExtractionPlumbing is the ONE runtime file allowed to reference the
// low-level iam/domain accessor: the canonical consumer API is implemented on
// top of it. The seam is a single file rather than a package prefix so that
// widening it is a visible, reviewable edit here instead of a new file quietly
// landing inside an allowed directory.
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
// may not REFERENCE the low-level accessor at all. The ban is on the reference
// and not on the call because `extract := iamdomain.UserIDFromContext` followed
// by `extract(ctx)` is the same fail-open accessor reached one hop later, and a
// rule that only matched the call site would treat a variable as a laundering
// step. Making the fail-open accessor unreachable is stronger than asking every
// consumer to remember the check.
//
// Rule 2 (ignored presence result): runtime code may not discard the canonical
// accessor's second result. `actorID, _ := authn.UserIDFromContext(ctx)`
// compiles, yields "" on absence, and reintroduces exactly the fail-open shape
// Rule 1 removed — the ban would otherwise be one underscore wide. Both the
// assignment form and the `var actorID, _ = ...` declaration form are matched;
// they are different AST nodes for one indistinguishable defect.
//
// Rule 3 (dot-import ban): neither protected path may be dot-imported. A dot
// import puts `UserIDFromContext` into file scope as a bare identifier, where
// Rules 1 and 2 — which resolve a qualifier to an import path — cannot see it.
// Refusing the ImportSpec is a smaller and stronger property than trying to
// decide, without type information, whether an unqualified `UserIDFromContext`
// in some file is the forbidden one: there is nothing to infer, the escape is
// simply not available.
//
// The rules deliberately match syntax, not data flow: every discard in this
// repository is a literal `_` in a two-name binding, and a general dataflow
// analysis would buy nothing the source actually needs.
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

	f := parseActorFile(fset, path)
	if f == nil {
		return nil
	}
	aliases := importAliases(f)
	suppressed := actorSuppressedLines(fset, f)
	// #108 review round 1, finding 3: `extract := authn.RequireUserID` binds
	// a canonical accessor to a local, undetected by the reference rule below
	// (which only watches the low-level accessor) because the bind itself is
	// not the violation — the presence-aware function is the RIGHT one to
	// call. The violation is one hop later, when the alias is called and its
	// second result discarded: `actor, _ := extract(ctx)` has call.Fun as a
	// bare *ast.Ident, invisible to isSelector. Collecting these aliases
	// first lets ignoredPresenceViolation recognise a call through the alias
	// as a call through the qualified accessor it stands for.
	funcAliases := authnAccessorAliases(f, aliases)

	var out []Finding
	// Rule 3 runs off f.Imports rather than the walk: importAliases deliberately
	// drops dot imports (they bind no qualifier), so by the time the walker sees
	// anything the escape has already been made invisible to the other rules.
	for _, spec := range f.Imports {
		if msg, ok := dotImportViolation(spec); ok {
			out = appendActorFinding(out, fset, suppressed, path, spec.Pos(), msg)
		}
	}

	ast.Inspect(f, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.SelectorExpr:
			if isPlumbing {
				return true
			}
			if isSelector(node, aliases, iamDomainImportPath, actorAccessorName) {
				out = appendActorFinding(out, fset, suppressed, path, node.Pos(),
					"references "+iamDomainImportPath+"."+actorAccessorName+", the low-level identity-storage "+
						"accessor that returns \"\" for a missing actor (A3.3); runtime consumers must use "+
						"authn.UserIDFromContext (presence-aware) or authn.RequireUserID (fails closed) so "+
						"absence is an explicit decision instead of an empty string travelling downstream")
			}
		case *ast.AssignStmt:
			if msg, pos, ok := ignoredPresenceViolation(node.Lhs, node.Rhs, node.Pos(), aliases, funcAliases); ok {
				out = appendActorFinding(out, fset, suppressed, path, pos, msg)
			}
		case *ast.ValueSpec:
			if msg, pos, ok := ignoredPresenceViolation(identsAsExprs(node.Names), node.Values, node.Pos(), aliases, funcAliases); ok {
				out = appendActorFinding(out, fset, suppressed, path, pos, msg)
			}
		}
		return true
	})
	return out
}

// parseActorFile parses with comments, which the shared parseFile does not. The
// suppression rule has to distinguish a real directive from the same characters
// inside a string literal, and only the scanner can tell them apart.
func parseActorFile(fset *token.FileSet, path string) *ast.File {
	f, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	if err != nil {
		return nil
	}
	return f
}

// dotImportViolation reports a dot import of either protected path.
func dotImportViolation(spec *ast.ImportSpec) (string, bool) {
	if spec.Name == nil || spec.Name.Name != "." {
		return "", false
	}
	path := strings.Trim(spec.Path.Value, `"`)
	if path != iamDomainImportPath && path != authnImportPath {
		return "", false
	}
	return "dot-imports " + path + " (A3.3); a dot import binds " + actorAccessorName + " as a bare " +
		"identifier with no qualifier to resolve, which is precisely the shape the low-level-accessor ban " +
		"and the ignored-presence ban cannot see — import it under a name so the actor rules apply", true
}

// identsAsExprs adapts a ValueSpec's names to the expression list the shared
// discard matcher works on. `var actorID, _ = authn.UserIDFromContext(ctx)` is a
// declaration, not an assignment, and reaches the walker as a different node
// carrying the identical defect.
func identsAsExprs(names []*ast.Ident) []ast.Expr {
	out := make([]ast.Expr, 0, len(names))
	for _, n := range names {
		out = append(out, n)
	}
	return out
}

// ignoredPresenceViolation matches a two-name binding fed by one call to a
// canonical accessor whose second slot is the blank identifier — covering both
// `x, _ := authn.UserIDFromContext(ctx)` and
// `var x, _ = authn.RequireUserID(ctx)`. funcAliases resolves the same
// violation one hop later, when the call runs through a local that was bound
// to the bare accessor first (`extract := authn.RequireUserID; x, _ :=
// extract(ctx)`) — see authnAccessorAliases.
func ignoredPresenceViolation(lhs, rhs []ast.Expr, pos token.Pos, aliases, funcAliases map[string]string) (msg string, out token.Pos, ok bool) {
	if len(lhs) != 2 || len(rhs) != 1 {
		return "", token.NoPos, false
	}
	blank, isIdent := lhs[1].(*ast.Ident)
	if !isIdent || blank.Name != "_" {
		return "", token.NoPos, false
	}
	call, isCall := rhs[0].(*ast.CallExpr)
	if !isCall {
		return "", token.NoPos, false
	}
	name, matched := resolvePresenceFuncName(call.Fun, aliases, funcAliases)
	if !matched {
		return "", token.NoPos, false
	}
	result := authnPresenceFuncs[name]
	return "discards the " + result + " of authn." + name + " (A3.3); the actor is then \"\" whenever " +
		"the request carries no authenticated principal, which is the fail-open shape the canonical " +
		"accessor exists to prevent — read the second result and fail explicitly, or use " +
		"authn.RequireUserID and propagate its error", pos, true
}

// resolvePresenceFuncName identifies which canonical presence-aware accessor
// fun calls, directly (`authn.RequireUserID(ctx)`, resolved by import path via
// isSelector) or through a local alias of the bare function
// (`extract(ctx)` where `extract := authn.RequireUserID` elsewhere in the
// file, resolved via funcAliases). Both shapes name the identical function; a
// rule that recognised only the first is a rule one assignment statement
// turns off.
func resolvePresenceFuncName(fun ast.Expr, aliases, funcAliases map[string]string) (string, bool) {
	for name := range authnPresenceFuncs {
		if isSelector(fun, aliases, authnImportPath, name) {
			return name, true
		}
	}
	if ident, ok := fun.(*ast.Ident); ok {
		if name, ok := funcAliases[ident.Name]; ok {
			return name, true
		}
	}
	return "", false
}

// authnAccessorAliases scans f for local bindings whose right-hand side is a
// BARE reference to one of the canonical presence-aware accessors — not a
// call, the function value itself, e.g. `extract := authn.RequireUserID` —
// and returns a map from the local identifier to the accessor's canonical
// name ("RequireUserID" / "UserIDFromContext"). It runs as a first pass over
// the whole file, before the discard check, so the alias is known regardless
// of where in the file it was declared relative to its use.
func authnAccessorAliases(f *ast.File, aliases map[string]string) map[string]string {
	out := map[string]string{}
	ast.Inspect(f, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.AssignStmt:
			collectAuthnAccessorAlias(node.Lhs, node.Rhs, aliases, out)
		case *ast.ValueSpec:
			collectAuthnAccessorAlias(identsAsExprs(node.Names), node.Values, aliases, out)
		}
		return true
	})
	return out
}

// collectAuthnAccessorAlias records lhs[0] -> canonical name when the binding
// is exactly one name fed by one bare SelectorExpr naming a canonical
// accessor (no call involved — `extract := authn.RequireUserID`, not
// `id := authn.RequireUserID(ctx)`).
func collectAuthnAccessorAlias(lhs, rhs []ast.Expr, aliases map[string]string, out map[string]string) {
	if len(lhs) != 1 || len(rhs) != 1 {
		return
	}
	ident, isIdent := lhs[0].(*ast.Ident)
	if !isIdent || ident.Name == "_" {
		return
	}
	sel, isSel := rhs[0].(*ast.SelectorExpr)
	if !isSel {
		return
	}
	for name := range authnPresenceFuncs {
		if isSelector(sel, aliases, authnImportPath, name) {
			out[ident.Name] = name
			return
		}
	}
}

// actorSuppressedLines collects the lines carrying a valid suppression: a real
// line comment starting with the directive and followed by a non-empty
// rationale. Reading f.Comments rather than the raw source is what makes the
// directive undefeatable by a string literal that happens to contain it, and
// what makes the rationale a requirement the analyzer can actually check.
func actorSuppressedLines(fset *token.FileSet, f *ast.File) map[int]bool {
	out := map[int]bool{}
	for _, group := range f.Comments {
		for _, c := range group.List {
			if _, ok := actorAllowRationale(c.Text); ok {
				out[fset.Position(c.Pos()).Line] = true
			}
		}
	}
	return out
}

// actorAllowRationale returns the rationale trailing a well-formed directive.
// The comment must START with the directive: a sentence that merely mentions the
// token in prose is discussing the mechanism, not invoking it.
func actorAllowRationale(commentText string) (string, bool) {
	if !strings.HasPrefix(commentText, actorExtractionAllow) {
		return "", false
	}
	rationale := strings.TrimSpace(commentText[len(actorExtractionAllow):])
	return rationale, rationale != ""
}

func appendActorFinding(out []Finding, fset *token.FileSet, suppressed map[int]bool, path string, pos token.Pos, msg string) []Finding {
	p := fset.Position(pos)
	if suppressed[p.Line] {
		return out
	}
	return append(out, Finding{
		Analyzer: "actorextraction",
		File:     path,
		Line:     p.Line,
		Message:  msg,
	})
}
