# Milestone 1 — Category A: typed status constants

> **Program:** backend-module-boundary-hardening  ·  **Governing spec:** `../mission.md` (§5 row 3, §7 M1)
> **Status:** Spec approved
> **Authored:** 2026-06-20 — *before any feature in this milestone began.*

> This file is a **spec**, authored up front. It says **what** this milestone is, **which features**
> it contains, **what each feature implements**, and **what gets validated**. It contains **no
> execution steps** — the "how" lives in the feature's `plan.md`. The end-of-milestone QA
> (`qa/milestone-qa.md`) validates the milestone against *this* document.

## Objective

Eliminate the **3 stringly-typed foreign domain-state literals** in controlled-documents (CD) domain
logic — the only **Category A** sites in the F0.2 binding census (A1–A3). CD's
`domain/resolution.go` compares a *template version* status (a value owned by the **templates**
module's vocabulary) against bare string literals `"published"`/`"obsolete"`. ADR-0039 classifies
these as **not base-table reads** (out of D1's SQL range) — an H-G-**adjacent** literal coupling the
mission remediates with **typed status constants** so the foreign vocabulary is referenced by name,
not duplicated as magic strings that can silently drift from the owner.

**Bar moved:** the Category-A literal-coupling class → **0 sites**. After this milestone, no CD code
hardcodes a templates-vocabulary status literal; the coupling is an explicit, compiler-checked
reference to the owner's published `VersionStatus` constants (precedent: ADR 0030 made templates the
owner of template-version state reads; the `"published"` hardcode at the infra reach was already
deleted in parent M4/F4.2 — this closes the residual literals in the domain resolver).

This is a **seam/clarity** change, **not** a logic change (mission §2 Non-Goals; D6 parity). Behavior
must be byte-identical: the existing `resolution_test.go` characterization suite is the parity lock.

## Features

| Feature id | Slug / folder | What to implement | What to validate (acceptance) |
|------------|---------------|-------------------|-------------------------------|
| F1.1 | `f1.1-resolution-constants` | `controlleddocuments/domain/resolution.go:42,55,58` reference the templates-owned `domain.VersionStatusPublished` / `VersionStatusObsolete` constants instead of bare `"published"` / `"obsolete"` literals. Pure refactor; no behavior, signature, or struct-shape change. | `go build ./...` exit 0; the existing `resolution_test.go` suite (9 tests) green **unchanged**; `Select-String '"published"','"obsolete"'` over `resolution.go` returns **0** bare-literal hits; `go run ./tools/cilint ./...` exit 0 (H-G guard unaffected — this is not an SQL site). |

## Milestone validation definition

The close gate is run by the **`milestone-validator` subagent** (separation of powers — it judges and
writes `qa/milestone-qa.md`; the main session flips status only on its PASS), following the binding
C1–C7 checklist in `.claude/skills/milestone/references/milestone-end-validation.md`. For M1 it
enforces:

1. **Per-feature acceptance** — F1.1 meets every cell of its "what to validate", and the feature's
   **consumer contract** (`spec.md`) was honored: the constants referenced are the **templates**
   module's published `VersionStatus*` vocabulary (the owner), not CD-local copies (which would
   re-introduce the drift the milestone removes).
2. **Workflow-class QA checklist** — backend code-quality / module-boundaries (DDD vocabulary
   ownership). No API, persistence, or migration surface is touched.
3. **Regression** — M0 still passes its gate: ADR-0039 unchanged, F0.3 cilint H-G guard still green on
   the full tree (M1 does not touch any SQL site, so the `hgPendingRemediation` debt ledger is
   untouched and must remain green).
4. **Quality-bar / root-cause check** — the Category-A class is re-measured: a tree-wide grep confirms
   the 3 literals are gone from `resolution.go` and were replaced by the **owner's** typed constants
   (root cause = duplicated foreign vocabulary), not by a CD-local literal alias (symptom patch).
5. **No unplanned scope** — only `resolution.go` (and, if a regression guard is added, its
   `_test.go`) change. No SQL, no ports, no views — those are M2–M4. Any other touched file is drift.

## Dependencies & constraints

- **Depends on:** M0 (ADR-0039 + binding census). M1 addresses census rows A1–A3 exactly.
- **Architectural constraints respected:**
  - **No SQL, no ports, no views, no migrations** — M1 is in-memory literals only. (Ports = M2;
    membership view = M3; search visibility contract = M4.)
  - **D6 parity (non-negotiable):** behavior identical. The constant values
    (`VersionStatusPublished = "published"`, `VersionStatusObsolete = "obsolete"`,
    `templates/domain/version.go:14-15`) equal the literals being replaced, so every existing
    `resolution_test.go` assertion holds unchanged. No raw SQL is deleted in M1, so D6's
    parity-before-delete clause is satisfied vacuously — but the characterization suite still gates.
  - **Vocabulary ownership (root cause):** the status vocabulary is **templates**-owned (ADR 0030).
    The fix references the owner's constants (`templates/domain`), imported under an alias to avoid
    the `domain`/`domain` package-name collision. CD-local duplicate constants are **forbidden** —
    they would re-create the drift this milestone removes.
  - No import cycle: `templates/domain` imports nothing from `controlleddocuments` (verified).

## Applicable hard-stops

| ID | What would trip it here |
|----|--------------------------|
| HS-1 | **The M1 boundary itself.** On validator PASS, the main session flips status and **stops** for operator review. No M2 start, no merge, without explicit approval. |
| HS-2 | If the literal→constant swap turned out to require reshaping the `TemplateVersionCandidate.Status` field type or the ADR-0030 `GetTemplateVersionState` port signature (a cross-module contract change) — **stop**, do not redesign the port under an M1 ticket. (Not expected: M1 keeps `Status *string` and converts at the comparison.) |
| HS-6 | If a Category-A literal-coupling site exists **beyond** census A1–A3 (e.g. another file hardcoding a templates status) such that M1's shape changes — **stop**, surface to the operator, replan before continuing. |
| HS-PRE-1 | N/A — no SQL, no tx, no lock. Recorded as not-applicable. |
