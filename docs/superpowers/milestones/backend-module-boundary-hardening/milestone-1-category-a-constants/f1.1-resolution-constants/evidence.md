# Feature F1.1 — Evidence

> **Milestone:** 1  ·  **Feature:** `f1.1-resolution-constants`  ·  **Closed:** 2026-06-20
> **Contract:** `spec.md` (consumer contract + Validation Gate this proves against).

## What was implemented

- `internal/modules/controlleddocuments/domain/resolution.go`: added import
  `templatesdomain "metaldocs/internal/modules/templates/domain"`; replaced the 3 bare status
  literals with references to the templates-owned constants:
  - `:42` (override) `!= "published"` → `!= string(templatesdomain.VersionStatusPublished)`
  - `:55` (default obsolete) `== "obsolete"` → `== string(templatesdomain.VersionStatusObsolete)`
  - `:58` (default not-published) `!= "published"` → `!= string(templatesdomain.VersionStatusPublished)`
- `internal/modules/controlleddocuments/domain/resolution_test.go`: added boundary regression guard
  `TestResolve_UsesTemplatesVocabulary` feeding the templates constants as inputs.
- **Producer matches consumer contract:** `Resolve`'s behavior, signature, return values, and error
  mapping are unchanged (D6 parity). `TemplateVersionCandidate.Status` stays `*string` — the field
  type and the ADR-0030 `GetTemplateVersionState` port signature are untouched (M2/HS-2 boundary
  respected). The constant values equal the prior literals
  (`templates/domain/version.go:14-15`: `VersionStatusPublished="published"`,
  `VersionStatusObsolete="obsolete"`), so the comparison semantics are identical.
- Commit: `<filled at commit>`.

## Verification

| Check | Command / action | Result (evidence) | Real vs fixture |
|-------|------------------|-------------------|-----------------|
| Baseline (pre-change) | `go build ./...`; `go test … -run TestResolve -v`; `go run ./tools/cilint ./...`; literal grep | build=0; **9** PASS; cilint=0; literals **present** at resolution.go:42,55,58 | real |
| TDD disposition | refactor (behavior-preserving) — characterization suite is the parity lock; 1 added regression guard wired to templates constants | guard `TestResolve_UsesTemplatesVocabulary` **PASS**; not RED-first (values equal pre/post) — labeled honestly | real |
| Targeted test (green-after) | `go test ./internal/modules/controlleddocuments/domain/ -run TestResolve -v` | **10/10 PASS** (9 existing unchanged + new guard); `ok …/domain 0.703s` | real |
| Static (build) | `go build ./...` | exit **0** | — |
| 0 bare status literals in resolution.go | `Select-String -Path …\resolution.go -Pattern '"published"','"obsolete"'` | **0 matches** | real |
| H-G guard unaffected | `go run ./tools/cilint ./...` | exit **0** (not an SQL site; debt ledger untouched) | real |
| Wider regression | `go test ./internal/modules/controlleddocuments/... ./internal/modules/templates/...` | all `ok`, exit **0** (application/delivery/domain/infrastructure + all templates pkgs) | real |

> All observable surface is in-memory domain logic exercised by the unit suite — no route, no DB, no
> provider. Docker Postgres (:5433) not required for this feature; no integration step was skipped
> (none applies). No fixture-as-real substitution.

## Acceptance vs spec Validation Gate

| Acceptance criterion (from spec.md) | Met? | Evidence |
|-------------------------------------|------|----------|
| Behavior unchanged — every resolution branch identical | yes | row "Targeted test" — 9 existing tests PASS unchanged from baseline |
| Regression guard: resolution agrees with templates vocabulary by reference | yes | row "Targeted test" — `TestResolve_UsesTemplatesVocabulary` PASS |
| Build clean | yes | row "Static (build)" — exit 0 |
| 0 bare status literals remain in resolution.go | yes | row "0 bare status literals" — 0 matches |
| H-G guard unaffected (cilint exit 0) | yes | row "H-G guard unaffected" — exit 0 |

## Review disposition

- **Spec-compliance review:** PASS. Diff is exactly the 3 literal swaps + import alias + 1 regression
  guard test — matches `spec.md` "What this feature implements" and touches only the two files named in
  the plan. No non-goal violated (no field retype, no port change, no SQL, no CD-local constants, no
  `CDStatus`/`api.gen.go` edits).
- **Code-quality review:** PASS. Import aliased (`templatesdomain`) to resolve the `domain`/`domain`
  package-name collision; `string(...)` conversion is the minimal parity-preserving form given
  `Status *string`. Consumes the owner's published vocabulary (drift-proof) rather than duplicating it
  — root-cause fix, not symptom patch.

## Bounded defers

| Defer | Why bounded | Written trigger / owner |
|-------|-------------|-------------------------|
| Retype `TemplateVersionCandidate.Status` `*string` → `*VersionStatus` (drop the `string(...)` conversion) | Would ripple into the ADR-0030 `GetTemplateVersionState` port return type — a cross-module contract change, out of M1's seam scope (HS-2). Current form is fully type-checked against the owner's constant; the only residual is the conversion call, not a literal. | Trigger: if/when M2 reshapes the templates state port. Owner: M2 backend agent. (Optional — not required by the mission; M1 done-bar is the literals, which are gone.) |
