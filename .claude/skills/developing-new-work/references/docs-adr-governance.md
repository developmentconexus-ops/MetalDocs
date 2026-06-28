# Docs & ADR governance

**Last verified:** 2026-06-28

## Wiki module doc
A new module gets a `wiki/modules/<name>.md` following the standard 12-section structure (exemplar:
`wiki/modules/taxonomy.md`), plus:
- a sister `wiki/modules/<name>-tech-debt.md` (known gaps / deferred items), and
- an entry in `wiki/modules/index.md`.

A **feature** updates the owning module's existing doc and refreshes its `Last verified` header — it
does not create a new module doc.

## `Last verified` convention
Every wiki doc and every reference in this skill carries a `Last verified: YYYY-MM-DD` stamp
(`wiki/standards/documentation-governance.md`). Runtime truth beats docs: when code and doc disagree,
fix the doc and re-stamp. Don't cite a stale doc as truth.

## REQ-ID citation
The governing target spec is `wiki/architecture/backend-target-architecture.md`, which carries REQ
IDs. Cite the relevant REQ IDs in §9 of the artifact and in any review. A **MUST-deviation** from the
target spec requires an ADR.

## When an ADR is required
Write an ADR when the work:
- deviates from a MUST in the target spec, or
- changes a standing policy / supersedes an earlier decision (e.g. the tenant dictionary supersedes
  ADR 0008's "fixed catalog" stance).

ADRs live in `wiki/decisions/` with sequential numbering; follow the existing ADR template. Key ADRs to
know: **0007**, **0008** (fixed placeholder catalog — superseded when the token dictionary lands),
**0022** (capability-oriented authz — the authz boundary), **0034** (integration test fixture
framework). A routine, in-bounds feature usually needs **no** ADR — only flag one when a MUST-deviation
or policy change is real.
