package problem

import (
	"encoding/json"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"testing"
)

// mustPanic runs fn and returns the recovered panic value, failing the test if
// fn did not panic. The registry's entire contract is "fail at init", so almost
// every negative case here is a panic assertion.
func mustPanic(t *testing.T, fn func()) string {
	t.Helper()
	var got string
	func() {
		defer func() {
			r := recover()
			if r == nil {
				t.Fatal("expected panic, got none")
			}
			got, _ = r.(string)
		}()
		fn()
	}()
	return got
}

func TestRegister_HappyPath(t *testing.T) {
	c := Register("codetest", "validation.happy_path", 422)
	if c.String() != "validation.happy_path" {
		t.Fatalf("String() = %q", c.String())
	}
	status, ok := StatusFor(c)
	if !ok || status != 422 {
		t.Fatalf("StatusFor = (%d, %v), want (422, true)", status, ok)
	}
	got, ok := Lookup("validation.happy_path")
	if !ok || got != c {
		t.Fatalf("Lookup = (%v, %v), want the registered code", got, ok)
	}
}

// TestRegister_RejectsDuplicate is the cross-module redeclaration guard: two
// modules may not both own the same wire string. This is what stops the class of
// defect ADR 0089 names (tokens redeclaring the catalog's NOT_FOUND).
func TestRegister_RejectsDuplicate(t *testing.T) {
	Register("codetest", "validation.dup_guard", 422)
	msg := mustPanic(t, func() {
		Register("othermodule", "validation.dup_guard", 422)
	})
	if !strings.Contains(msg, "validation.dup_guard") ||
		!strings.Contains(msg, "codetest") ||
		!strings.Contains(msg, "othermodule") {
		t.Fatalf("panic message must name the code and BOTH modules, got: %s", msg)
	}
}

// A duplicate is a duplicate even when the second registration is identical in
// every field — same module, same status. Ownership is exclusive.
func TestRegister_RejectsIdenticalDuplicate(t *testing.T) {
	Register("codetest", "validation.dup_identical", 422)
	mustPanic(t, func() { Register("codetest", "validation.dup_identical", 422) })
}

func TestRegister_RejectsUnknownFamily(t *testing.T) {
	msg := mustPanic(t, func() { Register("codetest", "banana.not_a_family", 400) })
	if !strings.Contains(msg, "banana") {
		t.Fatalf("panic message must name the offending prefix, got: %s", msg)
	}
	// A SCREAMING_SNAKE legacy-shaped name has no dotted family at all.
	mustPanic(t, func() { Register("codetest", "VALIDATION_ERROR", 400) })
	// A dotted name whose first segment is a module name, not a family, is the
	// exact anti-pattern ADR 0089 forbids.
	mustPanic(t, func() { Register("codetest", "approval.something", 409) })
}

func TestRegister_RejectsMalformedName(t *testing.T) {
	for _, bad := range []string{
		"",                       // empty
		"validation.",            // empty tail segment
		".stale",                 // empty family segment
		"validation..stale",      // doubled dot
		"validation.Stale",       // not lower snake_case
		"validation.stale-thing", // hyphen
		"validation._stale",      // leading underscore
		"validation.stale_",      // trailing underscore
		"validation.stale__x",    // doubled underscore
		"validation",             // no member segment
	} {
		t.Run(bad, func(t *testing.T) {
			mustPanic(t, func() { Register("codetest", bad, 422) })
		})
	}
}

// Family default status is data in one place; declaring a contradicting status
// through Register is a panic, and the message must point at the opt-in.
func TestRegister_RejectsStatusContradictingFamily(t *testing.T) {
	msg := mustPanic(t, func() { Register("codetest", "validation.wrong_status", 409) })
	if !strings.Contains(msg, "RegisterWithStatus") {
		t.Fatalf("panic must point at the explicit opt-in, got: %s", msg)
	}
}

func TestRegister_RejectsStatusOutOfRange(t *testing.T) {
	mustPanic(t, func() { Register("codetest", "validation.status_200", 200) })
	mustPanic(t, func() { Register("codetest", "validation.status_600", 600) })
}

func TestRegisterWithStatus(t *testing.T) {
	c := RegisterWithStatus("codetest", "auth.locked_example", 403,
		"the account exists and the credentials are valid; the block is a policy state, not a challenge")
	status, ok := StatusFor(c)
	if !ok || status != 403 {
		t.Fatalf("StatusFor = (%d, %v), want (403, true)", status, ok)
	}

	// The reason is mandatory: an undocumented override is exactly the implicit
	// fallback the closed registry exists to eliminate.
	mustPanic(t, func() { RegisterWithStatus("codetest", "auth.no_reason", 403, "") })

	// Using the override to restate the family default hides the code among the
	// genuine exceptions, so it is rejected.
	msg := mustPanic(t, func() {
		RegisterWithStatus("codetest", "auth.redundant", 401, "no deviation at all")
	})
	if !strings.Contains(msg, "Register") {
		t.Fatalf("panic must point back at plain Register, got: %s", msg)
	}
}

// RegisterLegacy skips family validation (that is its whole purpose) but keeps
// the duplicate and status-range guards. It disappears with ADR 0089's rename
// step; if this test ever needs relaxing, delete the escape hatch instead.
func TestRegisterLegacy(t *testing.T) {
	c := RegisterLegacy("codetest", "LEGACY_SHAPED_CODE", 418)
	if c.String() != "LEGACY_SHAPED_CODE" {
		t.Fatalf("String() = %q", c.String())
	}
	if status, ok := StatusFor(c); !ok || status != 418 {
		t.Fatalf("StatusFor = (%d, %v)", status, ok)
	}
	mustPanic(t, func() { RegisterLegacy("othermodule", "LEGACY_SHAPED_CODE", 418) })
	mustPanic(t, func() { RegisterLegacy("codetest", "LEGACY_RANGE", 200) })
}

// Field codes live in their own namespace: a field code and a problem code may
// share a wire string without colliding.
func TestRegisterField_SeparateNamespace(t *testing.T) {
	pc := RegisterLegacy("codetest", "SHARED_SPELLING", 400)
	fc := RegisterField("codetest", "SHARED_SPELLING")
	if pc.String() != fc.String() {
		t.Fatalf("expected identical wire strings")
	}
	if _, ok := StatusFor(fc); !ok {
		// Same wire string, so the problem-code lookup resolves; the point is only
		// that registering the field code did not panic as a duplicate.
		t.Log("field code shares a wire string with a problem code, as expected")
	}
	mustPanic(t, func() { RegisterField("othermodule", "SHARED_SPELLING") })

	var found bool
	for _, r := range FieldRegistrations() {
		if r.Code == fc {
			found = true
		}
	}
	if !found {
		t.Fatal("field registration missing from FieldRegistrations()")
	}
}

// The generator (ADR 0089 step 14) reads Registrations() and must get a stable
// order, or every regeneration churns the emitted file.
func TestRegistrations_Deterministic(t *testing.T) {
	first := Registrations()
	if len(first) < len(familyDefaults) {
		t.Fatalf("registry looks empty: %d entries", len(first))
	}

	keys := make([]string, len(first))
	for i, r := range first {
		keys[i] = r.Code.String()
	}
	if !sort.StringsAreSorted(keys) {
		t.Fatalf("Registrations() is not sorted by wire string")
	}

	for i := 0; i < 5; i++ {
		next := Registrations()
		if len(next) != len(first) {
			t.Fatalf("length drift between calls")
		}
		for j := range next {
			if next[j] != first[j] {
				t.Fatalf("entry %d differs between calls: %+v vs %+v", j, next[j], first[j])
			}
		}
	}

	// The snapshot must be a copy: mutating it must not corrupt the registry.
	first[0].Module = "mutated"
	if Registrations()[0].Module == "mutated" {
		t.Fatal("Registrations() leaked the backing storage")
	}
}

func TestFamilies_ClosedSetAndDeterministic(t *testing.T) {
	want := map[string]int{
		"request": 400, "validation": 422, "auth": 401, "permission": 403,
		"notfound": 404, "state": 409, "conflict": 409, "precondition": 412,
		"ratelimit": 429, "internal": 500,
	}
	got := Families()
	if len(got) != len(want) {
		t.Fatalf("family count = %d, want %d (the set is closed)", len(got), len(want))
	}
	names := make([]string, len(got))
	for i, f := range got {
		names[i] = f.Name
		if want[f.Name] != f.DefaultStatus {
			t.Fatalf("family %q default = %d, want %d", f.Name, f.DefaultStatus, want[f.Name])
		}
	}
	if !sort.StringsAreSorted(names) {
		t.Fatalf("Families() is not sorted: %v", names)
	}
}

// Lookup is the ONLY string -> Code direction, and it is deliberately fallible.
func TestLookup_UnregisteredFails(t *testing.T) {
	if c, ok := Lookup("validation.definitely_never_registered"); ok {
		t.Fatalf("Lookup resolved an unregistered string: %v", c)
	}
	if c, ok := Lookup(""); ok {
		t.Fatalf("Lookup resolved the empty string: %v", c)
	}
}

func TestStatusFor_UnregisteredFails(t *testing.T) {
	if _, ok := StatusFor(Code{}); ok {
		t.Fatal("StatusFor resolved the zero Code")
	}
}

// TestCode_WireFormatUnchanged pins the byte-level wire output. The closed type
// is an internal representation change ONLY: every JSON byte a client sees must
// be identical to what `type Code string` produced.
func TestCode_WireFormatUnchanged(t *testing.T) {
	cases := []struct {
		code Code
		want string
	}{
		{CodeValidationError, `"VALIDATION_ERROR"`},
		{CodeNotFound, `"NOT_FOUND"`},
		{CodeInternalError, `"INTERNAL_ERROR"`},
		{Code{}, `""`},
	}
	for _, tc := range cases {
		b, err := json.Marshal(tc.code)
		if err != nil {
			t.Fatalf("marshal %v: %v", tc.code, err)
		}
		if string(b) != tc.want {
			t.Fatalf("marshal = %s, want %s", b, tc.want)
		}
	}

	// Whole-envelope check: the code field must serialize as a bare JSON string,
	// not as an object wrapping the unexported field.
	p := New(404, CodeNotFound, "not found").WithDetail("gone").WithInstance("/x")
	b, err := json.Marshal(p)
	if err != nil {
		t.Fatalf("marshal problem: %v", err)
	}
	if !strings.Contains(string(b), `"code":"NOT_FOUND"`) {
		t.Fatalf("envelope lost the flat code field: %s", b)
	}

	// Round-trip: decoding an envelope must rebuild an equal Code value, so
	// existing tests that decode into problem.Problem keep working.
	var back Problem
	if err := json.Unmarshal(b, &back); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if back.Code != CodeNotFound {
		t.Fatalf("round-trip code = %q, want %q", back.Code.String(), CodeNotFound.String())
	}
}

// Code must stay comparable and usable as a map key — several call sites switch
// on codes and the generator keys maps by them.
func TestCode_ComparableAndMapKey(t *testing.T) {
	m := map[Code]int{CodeNotFound: 1, CodeInternalError: 2}
	if m[CodeNotFound] != 1 || m[CodeInternalError] != 2 {
		t.Fatal("Code is not usable as a map key")
	}
	a, _ := Lookup("NOT_FOUND")
	if a != CodeNotFound {
		t.Fatal("equal wire strings must produce equal Code values")
	}
	if CodeNotFound == CodeInternalError {
		t.Fatal("distinct codes compared equal")
	}
}

func TestNewFor_UsesRegisteredStatus(t *testing.T) {
	c := Register("codetest", "notfound.newfor_example", 404)
	p := NewFor(c, "gone")
	if p.Status != 404 || p.Code != c {
		t.Fatalf("NewFor = %+v", p)
	}
	// An unregistered code has no status to bind, so NewFor refuses rather than
	// silently defaulting to 500 (no-fallback principle).
	mustPanic(t, func() { NewFor(Code{}, "x") })
}

// TestCode_BareStringLiteralDoesNotCompile is the load-bearing guarantee of ADR
// 0089: the OLD `type Code string` accepted any untyped string constant, so a
// raw literal compiled silently even though the type carried a comment claiming
// otherwise. A compile failure cannot be asserted from inside a compiling
// package, so this test shells out to the compiler against a package under
// testdata/ (excluded from ./... by the go tool) and asserts it FAILS with the
// expected type error.
func TestCode_BareStringLiteralDoesNotCompile(t *testing.T) {
	if testing.Short() {
		t.Skip("invokes the go toolchain")
	}
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go toolchain not on PATH")
	}

	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot locate this test file")
	}
	pkgDir := filepath.Join(filepath.Dir(thisFile), "testdata", "compilefail")

	out, err := exec.Command("go", "build", "./...").CombinedOutput()
	_ = out
	if err != nil {
		t.Fatalf("sanity: the main module must build before this check is meaningful: %v", err)
	}

	out, err = exec.Command("go", "vet", pkgDir).CombinedOutput()
	if err == nil {
		t.Fatalf("bare string literals still compile as problem.Code — the closed type has been reopened:\n%s", out)
	}
	got := string(out)
	for _, want := range []string{
		"untyped string constant",
		"problem.Code",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("compiler output missing %q; got:\n%s", want, got)
		}
	}
}
