# ADR 0089 — Problem codes are a registry, not a convention

> **Status:** Proposed 2026-08-04 (operator ratified the three governing rulings:
> single dotted semantic taxonomy; status bound to the code; full scope —
> registry + rename + generation + CI gate).
> **Supersedes:** the flat `SCREAMING_SNAKE` catalog in
> `internal/platform/problem/codes.go`, the module-local dotted taxonomies
> (`approval/http/errors.go`, `taxonomy/.../routes_profiles.go`,
> `tokens/.../handler.go`), and the opt-in allowlist guard
> `internal/platform/problem/codes_catalog_guard_test.go`.
> **Relates to:** [ADR 0035](0035-flat-body-envelope.md) (wire shape),
> the 2026-07-03 final architecture review (hand-synced enumerations = the
> repo's named meta-defect), and
> `docs/engineering/defect-class-catalog.md` (this ADR closes Classes 1, 2, 3,
> 10, 13, 19 for one vocabulary).
> **Annex (binding):**
> `docs/superpowers/analysis/2026-08-04-problem-code-registry-mapping.md` —
> §1 families, §2 the complete 155-row rename table, §3 the 26 status rulings,
> §4 registry shape. The annex is the executable detail; this ADR is the policy.

## Context

RFC 9457 `problem+json` is an enforced invariant here: every ≥400 response
carries the envelope, and `scripts/api-lint`'s `ENVELOPE-DRIFT` rule fails the
build otherwise. 15/15 modules comply.

The `code` **inside** that envelope is enforced by nothing. An audit on
2026-08-04 found:

- **155 distinct codes in 3 competing conventions** — 75 dotted, 75
  `SCREAMING_SNAKE`, 5 bare lower_snake. (The audit first counted 147; the 8
  `AREA_*` / `FAMILY_*` codes in `taxonomy/.../routes_areas.go` and
  `routes_families.go` never reached `scripts/dump-error-codes.go` at all — the
  regex scraper missed raw literals, which is itself an instance of the defect
  decision 6 removes.) `controlleddocuments/delivery/http/routes.go:501-577`
  emits all three styles from a single switch, via 26 raw string literals.
- **The type guard is fake.** `codes.go:7` declares `type Code string` with the
  comment *"prevents arbitrary strings from being used as codes"*. Untyped
  string constants convert implicitly, so `problem.New(409, "anything", …)`
  compiles. Every raw literal above is legal Go.
- **18 collision classes** where two modules emit *different codes for the same
  condition* — e.g. "no active approval route" is `state.approval_route_missing`
  in `controlleddocuments` and `APPROVAL_ROUTE_MISSING` in `templates`, while a
  comment at `routes.go:519-523` claims both surfaces are "one contract for the
  client".
- **Status is decided per call site**, so the same condition carries different
  statuses across modules: `UPLOAD_MISSING` is 410 in `documents` and 409 in
  `templates`; content-hash mismatch is 422 / 409 / 412 in three modules;
  ~25 iam/security sites emit **501 Not Implemented** paired with code
  `INTERNAL_ERROR`; `documents/.../export_handler.go:139` emits **502**.
- **The one guard that exists ratifies the drift.** `codes_catalog_guard_test.go`
  is an *opt-in allowlist*; `approval/http` (68 codes), the
  `controlleddocuments` switch, `taxonomy`, and `tokens` are all excluded, with a
  comment legitimizing the second taxonomy as intentional.
- **The spec names a taxonomy it does not encode.** `openapi.yaml:7190` types
  `code` as a free `string` described as *"Machine-readable code from canonical
  taxonomy"*. No enum. Clients cannot generate an exhaustive switch.
- **The FE bridge is manual and currently broken.** `scripts/dump-error-codes.go`
  regenerates `error-codes.generated.json`, is wired to **no** CI job, and is
  stale by exactly 3 codes — the three added by ADR 0087 on 2026-08-04. It also
  silently drops the 4 field-level codes, two of which genuinely reach the wire
  (`tokens/.../handler.go:201,205`). The FE coverage test compares the map to the
  *snapshot*, never to backend source, so a stale snapshot passes green forever.

The shape of the defect is not "someone picked the wrong style". It is that a
**cross-process vocabulary was governed by convention instead of by a
mechanism**. Every module solved it locally, correctly, and incompatibly.

## Decision

1. **`problem.Code` becomes unforgeable.** It is a struct with an unexported
   field, so a bare string literal is a **compile error**, not a lint finding:

   ```go
   type Code struct{ s string }
   func (c Code) String() string { return c.s }
   func (c Code) MarshalJSON() ([]byte, error) { return json.Marshal(c.s) }
   ```

   It stays comparable, usable as a map key, and serializes to the byte-identical
   wire string. Cost is ~12 `string(code)` → `code.String()` sites and
   `const` → `var`. The compiler replaces the allowlist guard, so
   `codes_catalog_guard_test.go` and its `guardedPackages` exclusions are
   **deleted** — the drift they legitimized becomes unrepresentable.

2. **One registry, one declaration site per code.**
   `problem.Register(module, code string, defaultStatus int) Code` is the only
   constructor. Registering the same code twice panics at init. Modules
   self-register via package-level `var`, so `platform` never imports a module.
   Discovery for the generator is a `cmd/problem-codes-dump` that blank-imports
   each registering package, guarded by an api-lint rule that AST-scans for
   `problem.Register(` and fails when a registering package is missing from the
   dumper. Chosen over a pure AST scan because it reports **runtime truth** and
   validates the status binding.

3. **Status is bound to the code at registration.** A per-call-site status
   becomes an explicit, documented override rather than a default, so "the
   caller's precondition failed" cannot be 409 in one module and 412 in another.
   The 26 status rulings are in annex §3; the five behaviour-visible ones are
   ratified here:
   - **content-hash mismatch → 412** everywhere. The caller *declared* a
     precondition and it failed (RFC 9110 §15.5.13). 422 would claim the payload
     is semantically bad, which is false — the payload is fine, the world moved.
   - **`UPLOAD_MISSING` → 409** everywhere. 410 Gone asserts the resource existed
     and was permanently removed; the upload never happened. 409 is the conflict
     with current resource state.
   - **`IDEMPOTENCY_KEY_REUSED` → 409.** The key is well-formed and meaningful;
     the fault is conflict with a prior request, not unprocessability.
   - **`FAMILY_NOT_FOUND` splits into two codes** (annex R-23 + R-26).
     It is emitted at `taxonomy/.../routes_families.go:129` (404), where the
     family **is** the request target, and at `routes_profiles.go:375` (409),
     where a `family_code` **referenced inside a profile body** does not resolve.
     Those are two conditions, and binding status to the code makes sharing one
     name impossible. They become `notfound.document_family` @404 and
     `validation.family_unknown` @422. Keeping a single code would have forced a
     status change on one of the two sites anyway, so the split is the only
     option that changes behaviour for a stated reason rather than by accident.
   - **The taxonomy area cluster moves 400 → 422** (annex R-25):
     `validation.area_parent_cycle` and `validation.area_code_immutable`. Both
     requests parse; a supplied value fails a business rule. This is not a new
     ruling but R-20's, which already moved the identical
     `PROFILE_CODE_IMMUTABLE` to 422 — leaving the area twin at 400 would
     recreate, inside one module, the same-condition-different-status defect this
     ADR exists to remove.

4. **One taxonomy: semantic dotted families, closed set.** Families are
   **semantic, never module-named**. The code is a wire contract telling the
   client *what to do*; the failing Go module is an implementation detail that
   leaks internal structure onto the wire and breaks when a module moves —
   which literally happened when `approval` was extracted to a top-level module
   (ADR 0082), stranding `signoff.` and `sod.`. Ten families:
   `request.` (400) · `validation.` (422) · `auth.` (401) · `permission.` (403) ·
   `notfound.` (404) · `state.` (409) · `precondition.` (412) · `conflict.` (409) ·
   `ratelimit.` (429) · `internal.` (500).

   `request.` is split out of `validation.` deliberately: 30 `validation.*` codes
   currently span four statuses, so the prefix carries no branchable signal. The
   split gives the client two distinct destinations — `request.*` is a client bug
   (log it), `validation.*` is a message for the user (render it). Every
   module-named prefix is re-homed per annex §1.3. There is **no `config.`
   family**: misconfiguration is indistinguishable from an internal fault on the
   wire and folds into `internal.`.

5. **Hard break, no compatibility layer.** All 67 `SCREAMING_SNAKE` codes are
   renamed, the 11 dead constants deleted, the 24 collision classes collapsed to
   one code each. No aliases, no dual-accept, no deprecation window — per the
   standing legacy-fallback-extermination rule. The product has not shipped;
   this break costs an afternoon now and a coordinated client migration later.
   Post-sweep total ≈ **124** registered codes.

6. **Every downstream representation is generated, and CI gates freshness.**
   The registry is the single source; the OpenAPI `Problem.code` enum, the FE
   `errorMessages` key set, and the wiki code table are **outputs**. A CI job
   regenerates and fails on diff. `scripts/dump-error-codes.go` is deleted and
   replaced by the dumper in (2), which also covers the field-level codes it
   silently dropped. The FE coverage test is re-pointed at the generated
   artifact's source of truth, so a stale snapshot can no longer pass green.

## Consequences

- A wrong or invented code stops being possible: it fails to compile.
- The same condition returns the same code and the same status from every
  module, which is what the `controlleddocuments` comment already promised.
- Clients can generate an exhaustive switch from the spec enum.
- The FE message map cannot silently go stale — the failure mode that shipped
  3 unmapped codes to users on 2026-08-04.
- Cost is a large mechanical diff across every HTTP-facing module. It is
  mechanical *because* annex §2 decides every row in advance; executing it
  without that table would be judgment-by-agent at 155 sites.
- One-way door acknowledged: after clients exist, this rename requires a
  versioned contract migration. That is the argument for doing it now.

## Rejected alternatives

- **Pick a style and sed the repo.** Relocates an unenforced convention instead
  of enforcing one, touches every call site twice (once to rename, once when the
  registry lands), and leaves the type hole that produced the drift.
- **Keep both conventions, register them as-is.** Zero wire breakage, but the
  semantic incoherence stays: the same condition keeps two codes and three
  statuses. It solves the mechanism and abandons the reason for it.
- **Module-named families** (`approval.`, `templates.`). Reads naturally to the
  server author and leaks module topology onto a public contract; ADR 0082's
  module extraction already proved those names outlive their modules.
- **Keep the allowlist guard, widen it over time.** Debt with no due date; the
  existing exclusions have already hardened into a documented second standard.
- **Alias old codes to new for one release.** Two live vocabularies, and the
  compatibility layer becomes permanent because nothing forces its removal.
