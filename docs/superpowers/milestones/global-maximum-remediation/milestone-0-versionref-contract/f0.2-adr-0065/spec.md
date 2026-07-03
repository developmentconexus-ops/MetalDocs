# Feature F0.2 — adr-0065 — Spec

> **Milestone:** 0 — VersionRef contract refactor  ·  **Folder:** `f0.2-adr-0065`
> **Status:** Approved (pre-code)
> **Approved before code:** 2026-07-03 / Leandro (mission /goal directive; ADR text pre-drafted in plan Task 1) ✅

## Interview record (fail-closed gate)

| # | Question | Answer |
|---|----------|--------|
| 1 | Interview needed? | **None needed** — the decision, its context, and full ADR text are pre-authored in plan Task 1 (Step 1) and mandated by mission.md §7 F0.2 + gate artifact §9. Nothing to discover. This row is the C1 evidence. |

## Consumer contract (FIRST — before any producer)

- **Consumers:** the F0.1 cutover commit (cites ADR 0065); future reviewers/PRs (governance
  traceability); `wiki/decisions/index.md` (index consumer); the ADR-0035 memory + doc (annotated as
  structurally closed for this class).
- **Contract:** an ADR at `wiki/decisions/0065-version-references-are-nested-value-objects.md` with
  Status/Context/Decision/Consequences, stating: (1) version-reference field sets are one nested
  required object, never parallel scalars; (2) nullable-as-a-whole when the pointer may not exist,
  consumers gate on the object; (3) per-bounded-context schemas (`TemplateVersionRef` ≠
  `DocumentRevisionRef`), pattern unified not schema; (4) AIP view semantics (compact ref in list,
  full object in detail); (5) pre-v1 atomic-cutover exception to expand/contract. Complements ADR
  0035; closes its optional-vs-null drift subclass structurally.
- **Source of truth:** existing ADR format (e.g. `wiki/decisions/0034-integration-test-fixture-framework.md`).

## What this feature implements

ADR 0065 authored + indexed; ADR 0035 (memory `adr0035-flat-envelope-drift.md` + doc) annotated as
this-class-closed for templates (documents pending Plan 2); gate-artifact §10 constraints 5 and 6
amended per plan Task 1 Step 2. Plan Task 1.

## Non-goals (mandatory)

- No code change (this feature is docs/governance only).
- ADR 0022, 0013 not modified (0013 semantics unchanged, only referenced).
- Documents-side closure NOT claimed — ADR marks documents "pending migration".

## Validation Gate (concrete — approved before code)

| Acceptance criterion | Named test / proof command | Real vs fixture |
|----------------------|----------------------------|-----------------|
| ADR file exists, Accepted | `test -f wiki/decisions/0065-*.md` + read | real |
| Indexed | grep `0065` in `wiki/decisions/index.md` | real |
| Cited by cutover | F0.1 commit message references ADR 0065 | real |
| ADR 0035 annotated closed | `adr0035-flat-envelope-drift.md` memory updated | real |

## ADR needed?

- [x] This feature IS the ADR: `wiki/decisions/0065-version-references-are-nested-value-objects.md`.
