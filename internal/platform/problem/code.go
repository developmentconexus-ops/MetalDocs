package problem

import (
	"encoding/json"
	"fmt"
	"runtime"
	"sort"
	"strings"
	"sync"
)

// ---------------------------------------------------------------------------
// The closed Code type (ADR 0089 decision 1)
// ---------------------------------------------------------------------------

// Code is an opaque, registry-issued problem code — the machine-readable `code`
// member of an RFC 9457 problem+json body.
//
// WHY THIS IS A STRUCT AND NOT `type Code string`:
//
// The previous declaration was `type Code string`, carrying the comment "using a
// distinct type prevents arbitrary strings from being used as codes". That
// comment was FALSE, and its falseness is the defect ADR 0089 closes. Go
// implicitly converts an *untyped* string constant to any string-kinded named
// type, so `problem.New(409, "WHATEVER_I_WANT", …)` compiled silently. The audit
// on 2026-08-04 found 147 distinct codes in 3 competing conventions, 54 of them
// emitted as raw literals through exactly that hole, and 18 collision classes
// where two modules emitted different codes for the same condition.
//
// A struct with an unexported field has no untyped-constant conversion. There is
// no literal syntax for it outside this package, and composite-literal
// construction (`problem.Code{}`) cannot set `s`. Therefore the ONLY way to
// obtain a non-zero Code is Register / RegisterWithStatus / RegisterLegacy —
// which is what makes the registry the single source of truth rather than a
// convention someone can locally opt out of.
//
// If you are here because the closed type is inconvenient: that inconvenience is
// the mechanism. Do not reopen it (do not add an exported string conversion, a
// `Raw()` constructor, or `type Code = string`). Add a registration instead.
//
// Code stays comparable (`==`), usable as a map key (`map[Code]string`, used by
// the templates friendly-message table), and serializes to the byte-identical
// wire string the old string type produced.
type Code struct{ s string }

// String returns the wire representation of the code.
func (c Code) String() string { return c.s }

// IsZero reports whether c is the unregistered zero Code. A zero Code never
// reaches the wire through NewFor, which rejects it.
func (c Code) IsZero() bool { return c.s == "" }

// MarshalJSON emits the code as a plain JSON string, byte-identical to the
// pre-ADR-0089 `type Code string` encoding (`"code":"state.route_inactive"`).
// The wire contract is unchanged by the type change.
func (c Code) MarshalJSON() ([]byte, error) { return json.Marshal(c.s) }

// UnmarshalJSON is LENIENT by design. Tests, generated clients, and cross-build
// consumers decode Problem bodies produced by other binaries, which may carry
// codes this build has never registered; failing the decode would turn a
// vocabulary drift into an unparseable response. Callers that need registry
// membership must ask for it explicitly via Lookup.
func (c *Code) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	c.s = s
	return nil
}

// ---------------------------------------------------------------------------
// The closed family set (ADR 0089 decision 4)
// ---------------------------------------------------------------------------

// familyDefaults is the ONE place the semantic families and their default HTTP
// statuses are declared. Families are semantic, never module-named: the code is
// a wire contract telling the client what to do, and a family that names a Go
// module leaks internal topology onto a public contract and becomes a lie the
// moment the module moves (which literally happened when `approval` was
// extracted to a top-level module per ADR 0082, stranding `signoff.` and
// `sod.`).
//
// Adding a family here is an ADR-level decision, not a local one — the set being
// closed is what lets a client generate an exhaustive switch.
//
// Picking the family for a NEW code (annex §1.4, first match wins):
//
//  1. Could the caller have avoided it by sending a syntactically/structurally
//     different request?                                        -> request.
//  2. Does it parse, but a supplied value or a business rule over supplied
//     values rejects it, with no reference to stored state?     -> validation.
//  3. Is the problem WHO the caller is?                         -> auth.
//  4. Is the caller identified but lacks a capability?          -> permission.
//  5. Does the addressed subject not exist / is it invisible?   -> notfound.
//  6. Did a caller-SUPPLIED precondition stop holding?          -> precondition.
//  7. Race or uniqueness/idempotency collision, where retrying
//     against refreshed state could succeed?                    -> conflict.
//  8. Subject in the wrong lifecycle state, retry is futile?    -> state.
//  9. Is the caller being throttled?                            -> ratelimit.
//  10. Otherwise it is a server fault.                          -> internal.
//
// Tie-breaker between 7 and 8 (both 409): if a retry of the SAME request would
// succeed once a concurrent writer finishes, it is conflict.; if a DIFFERENT
// operation must change the subject first, it is state.
var familyDefaults = map[string]int{
	"request":      400,
	"validation":   422,
	"auth":         401,
	"permission":   403,
	"notfound":     404,
	"state":        409,
	"conflict":     409,
	"precondition": 412,
	"ratelimit":    429,
	"internal":     500,
}

// Families returns the closed family set with its default statuses, sorted by
// family name. Consumers: the generator (OpenAPI enum, wiki table) and tests.
func Families() []Family {
	out := make([]Family, 0, len(familyDefaults))
	for name, status := range familyDefaults {
		out = append(out, Family{Name: name, DefaultStatus: status})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

// Family is one entry of the closed semantic family set.
type Family struct {
	Name          string
	DefaultStatus int
}

// ---------------------------------------------------------------------------
// The registry (ADR 0089 decision 2 + 3)
// ---------------------------------------------------------------------------

// Registration is the registry record for one code.
type Registration struct {
	Code Code
	// Module is the owning bounded context, for the generated wiki table and the
	// drift lint ONLY. It is NEVER part of the wire contract — see the family
	// rationale above.
	Module string
	// Family is the code's prefix, validated against familyDefaults. Empty for
	// legacy registrations that predate the taxonomy (see RegisterLegacy) and for
	// field-level codes.
	Family string
	// DefaultStatus is the HTTP status NewFor binds to this code. ADR 0089
	// decision 3: status is bound to the code, so "the caller's precondition
	// failed" cannot be 409 in one module and 412 in another.
	DefaultStatus int
	// StatusOverrideReason is the documented justification when DefaultStatus
	// deviates from the family default. Empty when it matches.
	StatusOverrideReason string
	// Legacy marks a code registered through the temporary escape hatch.
	Legacy bool
	// DeclaredAt is the file:line of the registration call, for the drift report.
	DeclaredAt string
}

var (
	regMu sync.Mutex
	// registry holds Problem.code registrations, keyed by wire string.
	registry = map[string]Registration{}
	// fieldRegistry holds FieldError.code registrations. Field codes are a
	// SEPARATE namespace: FieldError.code describes an input field, not the
	// response, so it carries no HTTP status and must not collide-check against
	// Problem codes.
	fieldRegistry = map[string]Registration{}
)

// Register issues a Code whose default status is its family's default status.
//
// It PANICS on: an unknown or missing family prefix, a malformed name segment, a
// duplicate registration, or a status that contradicts the family default. A
// panic here fires during package initialisation, so it breaks every binary and
// every test at once — that is the point. A wrong code stops being a review
// finding and becomes a build break.
//
// The duplicate guard is also the cross-module redeclaration guard: two modules
// cannot independently declare the same wire string with different statuses
// (which `tokens` does today for NOT_FOUND / ALREADY_EXISTS).
func Register(module, code string, defaultStatus int) Code {
	family := mustValidate(module, code, defaultStatus)
	want := familyDefaults[family]
	if defaultStatus != want {
		panic(fmt.Sprintf(
			"problem.Register(%q, %q, %d): family %q defaults to %d. "+
				"If the deviation is deliberate, use RegisterWithStatus and state why.",
			module, code, defaultStatus, family, want))
	}
	return record(registry, Registration{
		Module:        module,
		Family:        family,
		DefaultStatus: defaultStatus,
		DeclaredAt:    callerSite(),
	}, code)
}

// RegisterWithStatus is the EXPLICIT opt-in for a code whose status legitimately
// deviates from its family default. The reason is mandatory and is carried into
// the generated documentation, so the deviation is reviewable rather than
// silent.
//
// It is legitimate when the family is semantically right but RFC 9110 assigns a
// narrower status to the specific condition. Ratified examples (annex §2/§3):
//
//   - auth.account_locked / auth.account_inactive at 403 rather than the auth.
//     default 401 — identity is PROVEN; the account is blocked (RFC 9110
//     §15.5.4). A 401 would wrongly invite the client to re-authenticate.
//   - precondition.if_match_required at 428 rather than 412 — no precondition
//     failed; the server refuses to act without one (RFC 6585 §3).
//   - internal.not_implemented at 501 and internal.upstream_failed at 502 —
//     server-fault family, specific 5xx.
//   - request.cursor_expired at 410, request.body_too_large at 413,
//     request.content_type_unsupported at 415, request.method_not_allowed at
//     405 — all "fix the request", each with its own dedicated 4xx.
//
// It is NOT legitimate as a way to keep a call site's historical status. If two
// modules disagree about a condition's status, that is the defect ADR 0089
// exists to remove; pick one and change the other.
func RegisterWithStatus(module, code string, status int, reason string) Code {
	family := mustValidate(module, code, status)
	if strings.TrimSpace(reason) == "" {
		panic(fmt.Sprintf(
			"problem.RegisterWithStatus(%q, %q, %d): reason is mandatory — "+
				"a status deviating from the family default must be justified in place",
			module, code, status))
	}
	if status == familyDefaults[family] {
		panic(fmt.Sprintf(
			"problem.RegisterWithStatus(%q, %q, %d): status equals the %q family default — use Register",
			module, code, status, family))
	}
	return record(registry, Registration{
		Module:               module,
		Family:               family,
		DefaultStatus:        status,
		StatusOverrideReason: reason,
		DeclaredAt:           callerSite(),
	}, code)
}

// RegisterLegacy registers a code that predates the ADR 0089 taxonomy, keeping
// its CURRENT wire string and CURRENT status without family validation.
//
// TEMPORARY — DELETE ME. This exists solely so the closed-type change (ADR 0089
// step 3) lands as one green, wire-identical commit, without smuggling the
// 147-row rename into it. The rename steps (annex §2, execution steps 5-9)
// convert every RegisterLegacy call to Register / RegisterWithStatus and then
// delete this function. If you are adding a NEW code and reached for this, you
// are doing it wrong: use Register.
//
// It still enforces the duplicate guard and the status range — those are not
// taxonomy rules, they are integrity rules.
func RegisterLegacy(module, code string, defaultStatus int) Code {
	mustValidateName(module, code)
	mustValidateStatus(module, code, defaultStatus)
	return record(registry, Registration{
		Module:        module,
		DefaultStatus: defaultStatus,
		Legacy:        true,
		DeclaredAt:    callerSite(),
	}, code)
}

// RegisterField issues a FIELD-level code for FieldError.code. Field codes live
// in their own namespace: they describe an input field rather than the response,
// so they carry no HTTP status and no family prefix. They are registered so the
// generator can enumerate them — the pre-0089 generator silently dropped all
// four, two of which genuinely reach the wire via tokens.
func RegisterField(module, code string) Code {
	mustValidateName(module, code)
	return record(fieldRegistry, Registration{
		Module:     module,
		DeclaredAt: callerSite(),
	}, code)
}

// record inserts reg (with its Code filled in from s) into dst, panicking on a
// duplicate. Shared by every Register* entry point so the duplicate guard has
// exactly one implementation.
func record(dst map[string]Registration, reg Registration, s string) Code {
	regMu.Lock()
	defer regMu.Unlock()

	if prev, dup := dst[s]; dup {
		panic(fmt.Sprintf(
			"problem: duplicate code %q registered by module %q at %s — already registered by module %q at %s. "+
				"One code, one declaration site: reference the existing registration instead of redeclaring it.",
			s, reg.Module, reg.DeclaredAt, prev.Module, prev.DeclaredAt))
	}
	reg.Code = Code{s: s}
	dst[s] = reg
	return reg.Code
}

// mustValidate runs the shared name/status checks and returns the family.
func mustValidate(module, code string, status int) string {
	mustValidateName(module, code)
	mustValidateStatus(module, code, status)
	return mustFamily(module, code)
}

func mustValidateName(module, code string) {
	if strings.TrimSpace(module) == "" {
		panic(fmt.Sprintf("problem: code %q registered with an empty module", code))
	}
	if code == "" {
		panic(fmt.Sprintf("problem: module %q registered an empty code", module))
	}
	if code != strings.TrimSpace(code) {
		panic(fmt.Sprintf("problem: module %q registered code %q with surrounding whitespace", module, code))
	}
}

func mustValidateStatus(module, code string, status int) {
	// Problem codes only ever accompany an error response. 100-399 would mean a
	// success or redirect carrying a problem body, which is nonsense.
	if status < 400 || status > 599 {
		panic(fmt.Sprintf("problem: module %q registered code %q with out-of-range status %d (want 400-599)", module, code, status))
	}
}

// mustFamily extracts and validates the semantic family prefix.
func mustFamily(module, code string) string {
	family, segment, ok := strings.Cut(code, ".")
	if !ok || family == "" || segment == "" {
		panic(fmt.Sprintf(
			"problem: module %q registered code %q — codes are `family.name_segment`; valid families: %s",
			module, code, strings.Join(familyNames(), " ")))
	}
	if _, known := familyDefaults[family]; !known {
		panic(fmt.Sprintf(
			"problem: module %q registered code %q with unknown family %q — the family set is closed: %s. "+
				"Families are semantic, never module-named (ADR 0089 decision 4).",
			module, code, family, strings.Join(familyNames(), " ")))
	}
	if !isSnakeSegment(segment) {
		panic(fmt.Sprintf(
			"problem: module %q registered code %q — the name segment must be lower snake_case, subject-first, "+
				"with no module name and no verb tense (state.document_not_published, not state.cannotPublishDocument)",
			module, code))
	}
	return family
}

// isSnakeSegment reports whether s is lower snake_case: [a-z][a-z0-9_]* with no
// leading/trailing/doubled underscore. Dots are rejected — the family is the
// only level of nesting, so `state.a.b` cannot exist.
func isSnakeSegment(s string) bool {
	if s == "" || s[0] < 'a' || s[0] > 'z' || s[len(s)-1] == '_' {
		return false
	}
	prevUnderscore := false
	for i := 0; i < len(s); i++ {
		ch := s[i]
		switch {
		case ch >= 'a' && ch <= 'z', ch >= '0' && ch <= '9':
			prevUnderscore = false
		case ch == '_':
			if prevUnderscore {
				return false
			}
			prevUnderscore = true
		default:
			return false
		}
	}
	return true
}

func familyNames() []string {
	names := make([]string, 0, len(familyDefaults))
	for name := range familyDefaults {
		names = append(names, name+".")
	}
	sort.Strings(names)
	return names
}

// callerSite returns the file:line of the Register* caller, skipping this file's
// own frames.
func callerSite() string {
	// skip 0 = callerSite, 1 = the Register* function, 2 = the registering site.
	if _, file, line, ok := runtime.Caller(2); ok {
		return fmt.Sprintf("%s:%d", trimRepoPrefix(file), line)
	}
	return "unknown"
}

// trimRepoPrefix shortens an absolute build path to a repo-relative-ish path so
// the generated artifacts are reproducible across machines.
func trimRepoPrefix(file string) string {
	file = strings.ReplaceAll(file, "\\", "/")
	for _, anchor := range []string{"/internal/", "/apps/", "/cmd/"} {
		if i := strings.LastIndex(file, anchor); i >= 0 {
			return strings.TrimPrefix(file[i:], "/")
		}
	}
	return file
}

// ---------------------------------------------------------------------------
// Accessors (for the generator, the drift lint, and tests)
// ---------------------------------------------------------------------------

// Registrations returns every registered Problem code, sorted by wire string.
// The order is deterministic so generated artifacts (OpenAPI enum, FE code map,
// wiki table) diff cleanly and a CI freshness gate is meaningful.
func Registrations() []Registration { return snapshot(registry) }

// FieldRegistrations returns every registered field-level code, sorted by wire
// string.
func FieldRegistrations() []Registration { return snapshot(fieldRegistry) }

func snapshot(src map[string]Registration) []Registration {
	regMu.Lock()
	defer regMu.Unlock()

	out := make([]Registration, 0, len(src))
	for _, reg := range src {
		out = append(out, reg)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Code.s < out[j].Code.s })
	return out
}

// Lookup resolves a wire string to its registered Code. It is the ONLY
// string->Code direction that exists, and it is deliberately fallible: a string
// that was never registered yields ok=false rather than a forged Code.
//
// Use it for decoding and for assertions in tests. Do NOT use it to manufacture
// a code at an emit site — that reopens the hole this type closes.
func Lookup(s string) (Code, bool) {
	regMu.Lock()
	defer regMu.Unlock()

	reg, ok := registry[s]
	return reg.Code, ok
}

// StatusFor returns the registered default status for c.
func StatusFor(c Code) (int, bool) {
	regMu.Lock()
	defer regMu.Unlock()

	reg, ok := registry[c.s]
	return reg.DefaultStatus, ok
}
