package analyzers_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"metaldocs/tools/cilint/internal/analyzers"
)

// actorFixture writes a Go source file at the given relative path inside a temp
// directory so the analyzer's path-based filtering (runtime tree, plumbing seam)
// behaves as it does against the real repo.
func actorFixture(t *testing.T, relPath, src string) string {
	t.Helper()
	dir := t.TempDir()
	full := filepath.Join(dir, filepath.FromSlash(relPath))
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(full, []byte(src), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	return full
}

func actorFindings(t *testing.T, relPath, src string) []analyzers.Finding {
	t.Helper()
	return analyzers.ActorExtraction([]string{actorFixture(t, relPath, src)})
}

func requireActorFinding(t *testing.T, findings []analyzers.Finding, want string) {
	t.Helper()
	for _, f := range findings {
		if strings.Contains(f.Message, want) {
			return
		}
	}
	t.Fatalf("expected a finding containing %q, got %+v", want, findings)
}

// ─── E1: the ban is on the REFERENCE, not on the call ────────────────────────

// A rule that only matched CallExpr treats a variable as a laundering step: the
// author binds the accessor to a local and calls it one line later, and the same
// fail-open "" reaches the same consumer through a guard that reported nothing.
func TestActorExtraction_Positive_IndirectReference(t *testing.T) {
	src := `package application

import (
	"context"

	iamdomain "metaldocs/internal/modules/iam/domain"
)

func charge(ctx context.Context) string {
	extract := iamdomain.UserIDFromContext
	return extract(ctx)
}
`
	findings := actorFindings(t, "internal/modules/billing/application/svc.go", src)
	requireActorFinding(t, findings, "the low-level identity-storage accessor")
}

// The same indirection under an import alias. Alias-independence is a property of
// the reference rule too, not only of the call rule it replaced.
func TestActorExtraction_Positive_IndirectReference_Aliased(t *testing.T) {
	src := `package application

import (
	"context"

	id "metaldocs/internal/modules/iam/domain"
)

var extract = id.UserIDFromContext

func charge(ctx context.Context) string { return extract(ctx) }
`
	findings := actorFindings(t, "internal/modules/billing/application/svc.go", src)
	requireActorFinding(t, findings, "the low-level identity-storage accessor")
}

// The plumbing seam still holds: authn's own implementation is built ON the
// low-level accessor, so it is the one file allowed to name it.
func TestActorExtraction_Negative_PlumbingSeam(t *testing.T) {
	src := `package authn

import (
	"context"

	iamdomain "metaldocs/internal/modules/iam/domain"
)

func UserIDFromContext(ctx context.Context) (string, bool) {
	raw := iamdomain.UserIDFromContext(ctx)
	return raw, raw != ""
}
`
	findings := actorFindings(t, "internal/platform/authn/context.go", src)
	if len(findings) != 0 {
		t.Fatalf("the plumbing seam may reference the low-level accessor, got %+v", findings)
	}
}

// ─── #108 review round 1, finding 3: alias of the CANONICAL accessor ─────────

// E1 taught Rule 1 to follow a reference through a local alias of the
// low-level accessor. This is the same hole on Rule 2, on the OTHER accessor:
// `extract := authn.RequireUserID` binds the fail-closed function to a local,
// and `actor, _ := extract(ctx)` discards its error exactly as
// `actor, _ := authn.RequireUserID(ctx)` would. ignoredPresenceViolation
// resolves call.Fun with isSelector, which requires a *ast.SelectorExpr —
// `extract(ctx)` has call.Fun as a bare *ast.Ident, so the direct-call check
// never fires and the discard ships clean. A guard that only matches
// `authn.RequireUserID(ctx)` verbatim is a guard a one-line local assignment
// turns off.
func TestActorExtraction_Positive_IndirectCanonicalAccessorDiscard_Error(t *testing.T) {
	src := `package application

import (
	"context"

	"metaldocs/internal/platform/authn"
)

func chargeViaAliasedAccessor(ctx context.Context) string {
	extract := authn.RequireUserID
	actor, _ := extract(ctx)
	return actor
}
`
	findings := actorFindings(t, "internal/modules/billing/application/svc.go", src)
	requireActorFinding(t, findings, "discards the error of authn.RequireUserID")
}

// The bool-returning sibling is the same shape one accessor over: aliasing
// UserIDFromContext and discarding its presence bool through the alias.
func TestActorExtraction_Positive_IndirectCanonicalAccessorDiscard_PresenceBool(t *testing.T) {
	src := `package application

import (
	"context"

	"metaldocs/internal/platform/authn"
)

func settleViaAliasedAccessor(ctx context.Context) string {
	probe := authn.UserIDFromContext
	actor, _ := probe(ctx)
	return actor
}
`
	findings := actorFindings(t, "internal/modules/billing/application/svc.go", src)
	requireActorFinding(t, findings, "discards the presence bool of authn.UserIDFromContext")
}

// ─── E2: dot imports ─────────────────────────────────────────────────────────

// A dot import puts UserIDFromContext into file scope with no qualifier to
// resolve, which is exactly what Rules 1 and 2 read. Refusing the ImportSpec is
// smaller and stronger than inferring bare identifiers.
func TestActorExtraction_Positive_DotImportIAMDomain(t *testing.T) {
	src := `package application

import (
	"context"

	. "metaldocs/internal/modules/iam/domain"
)

func charge(ctx context.Context) string { return UserIDFromContext(ctx) }
`
	findings := actorFindings(t, "internal/modules/billing/application/svc.go", src)
	requireActorFinding(t, findings, "dot-imports metaldocs/internal/modules/iam/domain")
}

func TestActorExtraction_Positive_DotImportAuthn(t *testing.T) {
	src := `package application

import (
	"context"

	. "metaldocs/internal/platform/authn"
)

func charge(ctx context.Context) string {
	actorID, _ := UserIDFromContext(ctx)
	return actorID
}
`
	findings := actorFindings(t, "internal/modules/billing/application/svc.go", src)
	requireActorFinding(t, findings, "dot-imports metaldocs/internal/platform/authn")
}

// A named import of the same package is the supported shape and must stay quiet
// on its own — the ban is on the dot, not on the dependency.
func TestActorExtraction_Negative_NamedImportIsNotADotImport(t *testing.T) {
	src := `package application

import (
	"context"

	"metaldocs/internal/platform/authn"
)

func charge(ctx context.Context) (string, error) { return authn.RequireUserID(ctx) }
`
	findings := actorFindings(t, "internal/modules/billing/application/svc.go", src)
	if len(findings) != 0 {
		t.Fatalf("a named import of a protected path is the supported shape, got %+v", findings)
	}
}

// ─── E3: the discard is a declaration too ────────────────────────────────────

// `var actorID, _ = ...` is a ValueSpec, not an AssignStmt. Same defect, same
// resulting "", different node — an AssignStmt-only matcher waves it through.
func TestActorExtraction_Positive_VarDeclarationDiscard(t *testing.T) {
	src := `package application

import (
	"context"

	"metaldocs/internal/platform/authn"
)

func settle(ctx context.Context) string {
	var actorID, _ = authn.UserIDFromContext(ctx)
	return actorID
}

func void(ctx context.Context) string {
	var actorID, _ = authn.RequireUserID(ctx)
	return actorID
}
`
	findings := actorFindings(t, "internal/modules/billing/application/svc.go", src)
	requireActorFinding(t, findings, "discards the presence bool of authn.UserIDFromContext")
	requireActorFinding(t, findings, "discards the error of authn.RequireUserID")
}

// The assignment form the rule started with, kept under test so widening to
// declarations cannot quietly drop it.
func TestActorExtraction_Positive_AssignmentDiscard(t *testing.T) {
	src := `package application

import (
	"context"

	"metaldocs/internal/platform/authn"
)

func settle(ctx context.Context) string {
	actorID, _ := authn.UserIDFromContext(ctx)
	return actorID
}
`
	findings := actorFindings(t, "internal/modules/billing/application/svc.go", src)
	requireActorFinding(t, findings, "discards the presence bool of authn.UserIDFromContext")
}

// ─── E4: suppression must be a real directive with a real reason ─────────────

// (1) A bare directive does not suppress. The analyzer's own contract is that an
// exemption from a fail-closed invariant has to say why; an unexplained opt-out
// is the defect, not the exception.
func TestActorExtraction_Suppression_BareDirectiveDoesNotSuppress(t *testing.T) {
	src := `package application

import (
	"context"

	iamdomain "metaldocs/internal/modules/iam/domain"
)

func charge(ctx context.Context) string {
	return iamdomain.UserIDFromContext(ctx) //cilint:allow-actor-extraction
}
`
	findings := actorFindings(t, "internal/modules/billing/application/svc.go", src)
	requireActorFinding(t, findings, "the low-level identity-storage accessor")
}

// (2) The token inside a string literal does not suppress. A raw-line scan
// cannot tell code from data, which makes the suppression mechanism reachable by
// anything that merely prints the directive.
func TestActorExtraction_Suppression_StringLiteralDoesNotSuppress(t *testing.T) {
	src := `package application

import (
	"context"

	iamdomain "metaldocs/internal/modules/iam/domain"
)

func charge(ctx context.Context) string {
	return iamdomain.UserIDFromContext(ctx) + "//cilint:allow-actor-extraction not a directive"
}
`
	findings := actorFindings(t, "internal/modules/billing/application/svc.go", src)
	requireActorFinding(t, findings, "the low-level identity-storage accessor")
}

// A comment that merely discusses the directive in prose is documentation, not
// an invocation: the directive has to START the comment.
func TestActorExtraction_Suppression_ProseMentionDoesNotSuppress(t *testing.T) {
	src := `package application

import (
	"context"

	iamdomain "metaldocs/internal/modules/iam/domain"
)

func charge(ctx context.Context) string {
	return iamdomain.UserIDFromContext(ctx) // see //cilint:allow-actor-extraction in the guard docs
}
`
	findings := actorFindings(t, "internal/modules/billing/application/svc.go", src)
	requireActorFinding(t, findings, "the low-level identity-storage accessor")
}

// (3) A real directive carrying a rationale does suppress — the escape hatch is
// tightened, not removed.
func TestActorExtraction_Suppression_DirectiveWithRationaleSuppresses(t *testing.T) {
	src := `package application

import (
	"context"

	iamdomain "metaldocs/internal/modules/iam/domain"
)

func charge(ctx context.Context) string {
	return iamdomain.UserIDFromContext(ctx) //cilint:allow-actor-extraction identity-storage probe: asserts the primitive itself, no consumer downstream
}
`
	findings := actorFindings(t, "internal/modules/billing/application/svc.go", src)
	if len(findings) != 0 {
		t.Fatalf("a directive with a rationale must suppress, got %+v", findings)
	}
}

// The rationale requirement applies to every rule the directive can suppress,
// not only to the one it was first written for.
func TestActorExtraction_Suppression_AppliesToDiscardRule(t *testing.T) {
	src := `package application

import (
	"context"

	"metaldocs/internal/platform/authn"
)

func settle(ctx context.Context) string {
	actorID, _ := authn.UserIDFromContext(ctx) //cilint:allow-actor-extraction anonymous telemetry counter: absence is bucketed, never attributed
	return actorID
}
`
	findings := actorFindings(t, "internal/modules/billing/application/svc.go", src)
	if len(findings) != 0 {
		t.Fatalf("a directive with a rationale must suppress the discard rule too, got %+v", findings)
	}
}

// ─── KNOWN GAPS (#108 review round 1): documented, not desired ───────────────
//
// A cold reviewer broke Rule 2 with five bypasses, all producing zero
// findings. The ruling on that finding was explicit: do NOT strengthen the
// analyzer to chase them — Rule 2 defends a dataflow property with syntax
// matching, and hardening it re-implements reachability analysis in a
// bespoke tool while entrenching the exact guard the typed-actor follow-up
// (an `authn.Actor` minted only by middleware and threaded to services as an
// explicit parameter) is meant to delete.
//
// These five tests exist so the boundary is executable rather than prose:
// anyone tightening Rule 2 later sees exactly where the line sits today, and
// this file fails loudly the moment a change makes one of these shapes fire
// — at which point it stops being a known gap and becomes a real finding to
// report, not silently drop.
//
// Each case asserts ZERO findings. None of these fixtures may ever be added
// to scripts/testdata/guard-fixtures/arch-lint/: that directory is the
// negative-fixture spine consumed by `go run ./tools/verify -guard-fixtures
// -only arch-lint`, which requires every fixture there to produce a
// non-zero exit. A zero-finding fixture placed there breaks the harness.

// actorFixtureMulti writes multiple source files into ONE temp directory —
// i.e. one package on disk — and returns their full paths. Needed only by
// the cross-file KnownGap below, where the whole point is that the two
// files must coexist in the same package for the gap to be observable.
func actorFixtureMulti(t *testing.T, files map[string]string) []string {
	t.Helper()
	dir := t.TempDir()
	paths := make([]string, 0, len(files))
	for relPath, src := range files {
		full := filepath.Join(dir, filepath.FromSlash(relPath))
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
		if err := os.WriteFile(full, []byte(src), 0o644); err != nil {
			t.Fatalf("write fixture: %v", err)
		}
		paths = append(paths, full)
	}
	return paths
}

// KnownGap 1: a chained alias. authnAccessorAliases only records a local
// bound DIRECTLY to a bare SelectorExpr (`a := authn.RequireUserID`); `b :=
// a` binds b to an *ast.Ident, not a SelectorExpr, so collectAuthnAccessorAlias
// never records "b". `b(ctx)` then resolves through neither isSelector (b is
// not a qualified reference) nor funcAliases (b was never added). Closed by
// the typed-actor follow-up, not by teaching this rule to chase transitive
// aliases.
func TestActorExtraction_KnownGap_ChainedAlias(t *testing.T) {
	src := `package application

import (
	"context"

	"metaldocs/internal/platform/authn"
)

func settle(ctx context.Context) string {
	a := authn.RequireUserID
	b := a
	actor, _ := b(ctx)
	return actor
}
`
	findings := actorFindings(t, "internal/modules/billing/application/svc.go", src)
	if len(findings) != 0 {
		t.Fatalf("documented known gap regressed to a real finding (report this, do not just update the assertion): %+v", findings)
	}
}

// KnownGap 2: a struct-field alias. `e.Fn(ctx)` reaches isSelector as a
// SelectorExpr, but sel.Sel.Name is "Fn", not "RequireUserID"/
// "UserIDFromContext" — isSelector never gets past the name check, and
// funcAliases only ever indexes bare identifiers, never struct fields.
func TestActorExtraction_KnownGap_StructFieldAlias(t *testing.T) {
	src := `package application

import (
	"context"

	"metaldocs/internal/platform/authn"
)

type Extractor struct {
	Fn func(context.Context) (string, error)
}

func settle(ctx context.Context) string {
	e := Extractor{Fn: authn.RequireUserID}
	actor, _ := e.Fn(ctx)
	return actor
}
`
	findings := actorFindings(t, "internal/modules/billing/application/svc.go", src)
	if len(findings) != 0 {
		t.Fatalf("documented known gap regressed to a real finding (report this, do not just update the assertion): %+v", findings)
	}
}

// KnownGap 3: function-parameter indirection. The accessor is passed as a
// `func(context.Context) (string, error)` argument and called through the
// parameter name inside the callee. funcAliases is built per-file from
// AssignStmt/ValueSpec bindings only; a function parameter is neither, so
// the callee's `get(ctx)` never resolves to a canonical accessor no matter
// what was passed at the call site.
func TestActorExtraction_KnownGap_FunctionParameterIndirection(t *testing.T) {
	src := `package application

import (
	"context"

	"metaldocs/internal/platform/authn"
)

func callAccessor(ctx context.Context, get func(context.Context) (string, error)) string {
	actor, _ := get(ctx)
	return actor
}

func settle(ctx context.Context) string {
	return callAccessor(ctx, authn.RequireUserID)
}
`
	findings := actorFindings(t, "internal/modules/billing/application/svc.go", src)
	if len(findings) != 0 {
		t.Fatalf("documented known gap regressed to a real finding (report this, do not just update the assertion): %+v", findings)
	}
}

// KnownGap 4: cross-file, same package. authnAccessorAliases runs once per
// *ast.File inside scanActorExtractionFile; ActorExtraction calls it
// separately for every path in the file list, so an alias table built while
// scanning one file is never consulted while scanning a sibling file in the
// same package — even though Go itself resolves the reference package-wide.
func TestActorExtraction_KnownGap_CrossFileAlias(t *testing.T) {
	paths := actorFixtureMulti(t, map[string]string{
		"internal/modules/billing/application/alias.go": `package application

import "metaldocs/internal/platform/authn"

var Extract = authn.RequireUserID
`,
		"internal/modules/billing/application/svc.go": `package application

import "context"

func settle(ctx context.Context) string {
	actor, _ := Extract(ctx)
	return actor
}
`,
	})
	findings := analyzers.ActorExtraction(paths)
	if len(findings) != 0 {
		t.Fatalf("documented known gap regressed to a real finding (report this, do not just update the assertion): %+v", findings)
	}
}

// KnownGap 5: a local wrapper function. `wrap` calls the canonical accessor
// and returns its result untouched, but the FuncDecl itself is never fed
// into authnAccessorAliases (which only watches AssignStmt/ValueSpec), so
// `wrap` is not in funcAliases and `actor, _ := wrap(ctx)` resolves to
// nothing the rule recognises — the discard inside wrap's own body is fine
// (it returns both values), and the discard at the call site is invisible.
func TestActorExtraction_KnownGap_LocalWrapperFunction(t *testing.T) {
	src := `package application

import (
	"context"

	"metaldocs/internal/platform/authn"
)

func wrap(ctx context.Context) (string, error) {
	return authn.RequireUserID(ctx)
}

func settle(ctx context.Context) string {
	actor, _ := wrap(ctx)
	return actor
}
`
	findings := actorFindings(t, "internal/modules/billing/application/svc.go", src)
	if len(findings) != 0 {
		t.Fatalf("documented known gap regressed to a real finding (report this, do not just update the assertion): %+v", findings)
	}
}
