# Feature F1.1 — Spec: resolution.go typed status constants

> **Milestone:** 1 — Category A: typed status constants  ·  **Folder:** `f1.1-resolution-constants`
> **Status:** Approved (pre-code)
> **Approved before code:** 2026-06-20 / leandrotca (operator-directed via mission §7 M1 F1.1; contract read from consumer source + ADR-0030/0039)

> This is the feature's **contract**, written and approved **before any code**. It defines what the
> feature must do and how it will be proven — not how it will be built (that is `plan.md`).

## Interview record (fail-closed gate)

The feature is mechanical and its contract is **explicit in existing source** (the mission row, the
templates owner package, and the consumer's current comparison). The ambiguities that *did* exist
were design-fork questions; resolved by reading source, recorded below.

| # | Question | Answer |
|---|----------|--------|
| 1 | Whose vocabulary do the `"published"`/`"obsolete"` literals belong to — CD's own status enum or another module's? | **templates.** `resolution.go` compares `TemplateVersionCandidate.Status`, which is populated from `s.tplCheck.GetTemplateVersionState(...)` (`service.go:221,232`) — the ADR-0030 templates-owned port returning the raw *template version* status. The vocabulary owner is `templates/domain` (`VersionStatus`, `version.go:8-15`). CD's own enum is the **separate** `CDStatus{Active,Obsolete,Superseded}` (`controlled_document.go:10-15`) — must **not** be confused with it. |
| 2 | Reference the **owner's** published constants, or define CD-local copies? | **Owner's.** Mission F1.1 directs `templates/domain VersionStatus*`. CD-local copies would re-duplicate the foreign vocabulary — the exact drift this milestone removes (a CD-local `"published"` const can silently diverge from templates'). Importing the owner's constant is drift-proof and matches the ADR-0030/0039 "consume the owner's published contract" principle. |
| 3 | Retype `TemplateVersionCandidate.Status` from `*string` to `*VersionStatus`, or keep `*string` and convert at the comparison? | **Keep `*string`; convert at comparison.** Retyping the field changes the struct shape and ripples into the port return type (`GetTemplateVersionState` returns `(*string, …)`, `service.go:29`/ADR-0030) — a cross-module contract change that is **M2 port territory and an HS-2 boundary** for M1. M1 stays minimal: `*candidate.Status != string(templatesdomain.VersionStatusPublished)`. Behavior identical. |
| 4 | Will importing `templates/domain` into CD `domain` create an import cycle or name collision? | **No cycle** (`templates/domain` imports nothing from `controlleddocuments`, verified). **Name collision yes** — both packages are named `domain`; resolve with an import alias (`templatesdomain`). |
| 5 | Are A1–A3 the only Category-A literal-coupling sites (HS-6 check)? | **Yes.** Tree grep of `internal/modules/controlleddocuments` for `"published"`/`"obsolete"` (non-test) returns only: resolution.go:42,55,58 (these), `controlled_document.go:14` (CD's own `CDStatusObsolete` — owner-defining, not coupling), `api.gen.go:64,103` (generated CD enum — own). HS-6 does not trip. |

## Consumer contract (FIRST — before any producer)

There is **no new producer** in this feature — it is a refactor of an existing consumer
(`Resolve`) to reference an already-published vocabulary. The "consumer contract" is the behavioral
contract the callers of `Resolve` rely on, which must be preserved exactly.

- **Consumer(s):** `controlleddocuments/application/service.go:227,354` (`Resolve(TemplateResolutionInput{…})`
  on the manual + auto create paths) and `domain/integration_test.go:33`, `domain/resolution_test.go`.
- **Contract (must be byte-identical post-change):** `Resolve` returns
  - override path: `Status == nil` → `ErrOverrideTemplateDeleted`; `Status != "published"` →
    `ErrOverrideNotPublished`; profile mismatch → `ErrTemplateProfileMismatch`; else `{ID, "override"}`.
  - default path: `nil` → `ErrProfileHasNoDefaultTemplate`; `== "obsolete"` → `ErrDefaultObsolete`;
    `!= "published"` → `ErrProfileHasNoDefaultTemplate`; else `{ID, "default"}`.
  The status comparison values are the templates `VersionStatus` wire values `"published"`/`"obsolete"`.
- **Source of truth for the vocabulary:** `internal/modules/templates/domain/version.go:14-15`
  (`VersionStatusPublished = "published"`, `VersionStatusObsolete = "obsolete"`) — the owner. ADR 0030
  (templates owns template-version state reads); ADR 0039 §"Worked classification" row 3 (this site is
  a literal coupling fixed via typed constants, out of D1's SQL range).

## What this feature implements

`resolution.go` lines 42, 55, 58 compare `*candidate.Status` against
`string(templatesdomain.VersionStatusPublished)` / `string(templatesdomain.VersionStatusObsolete)`
instead of the bare literals `"published"` / `"obsolete"`, importing
`templatesdomain "metaldocs/internal/modules/templates/domain"`. No other change.

## Non-goals (mandatory)

- **No** change to `Resolve`'s behavior, return values, or error mapping (D6 parity).
- **No** change to the `TemplateVersionCandidate.Status` field type (stays `*string`) or to the
  `GetTemplateVersionState` port signature — that is M2/HS-2 territory.
- **No** SQL, no read-port, no view, no migration (M2–M4).
- **No** CD-local duplicate status constants.
- **No** touching CD's own `CDStatus` enum or the generated `api.gen.go`.
- **No** edits to any file other than `resolution.go` (+ optionally its `_test.go` for a regression guard).

## Validation Gate (concrete — approved before code)

| Acceptance criterion | Named test / proof command | Real vs fixture |
|----------------------|----------------------------|-----------------|
| Behavior unchanged — every resolution branch identical | `go test ./internal/modules/controlleddocuments/domain/ -run TestResolve -v` → 9/9 PASS, **unchanged** from baseline | real (in-memory domain logic; no provider) |
| Regression guard: resolution agrees with the **templates** vocabulary by reference | new `TestResolve_UsesTemplatesVocabulary` feeding `templatesdomain.VersionStatusPublished/Obsolete` as inputs → accept/`ErrDefaultObsolete` as contracted | real |
| Build clean | `go build ./...` → exit 0 | — |
| 0 bare status literals remain in resolution.go | `Select-String -Path …\resolution.go -Pattern '"published"','"obsolete"'` → **0 matches** | real |
| H-G guard unaffected (not an SQL site) | `go run ./tools/cilint ./...` → exit 0 | real |

> TDD note: this is a **behavior-preserving refactor**, not new behavior — true RED-first does not
> apply (the constant values equal the literals, so any behavior test is green before and after). The
> existing `resolution_test.go` (9 tests) is the **characterization/parity lock**, green at baseline
> (captured) and green post-change. One **new regression-guard** test is added that wires the
> *templates constants themselves* as inputs, so a future drift between CD's resolver and the templates
> vocabulary would fail at CD's boundary. Disposition labeled honestly per the milestone QA rules.

## ADR needed?

- [x] No durable decision — skip. The governing decision is already ADR-0039 (§"Worked classification"
  row 3 names this exact fix) + ADR-0030 (templates owns the vocabulary). M1 executes them; it adds none.
