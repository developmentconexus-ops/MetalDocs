# Plan 2 — documents `DocumentRevisionRef` (BOUNDED DEFER, pre-v1)

Status: **DEFERRED** (not started). Sibling of F0.1's `TemplateVersionRef`. This
doc is the written defer of record — scope, trigger, and system-impact framing so
the work can be picked up cold without re-discovery. Governed by the
`developing-new-work` pre-design gate (must run before implementation).

## Why deferred, not done now
M0's charter is the **templates** version-pointer surface — the exact shape the
9f86828b bug lived on. The documents module carries a structurally identical
coupling (parallel revision/version scalars on document read models), but:
- It is a **separate bounded context** (own module, own DTOs, own consumers). Folding it into M0 would widen the milestone past its appetite and past the single feature the validator is scoped to judge.
- Documents wire shape is currently **unchanged and correct** (Drive 2 proved `GET /api/v1/documents` `{items,page,total}` stable) — there is no live defect forcing it now.
- Doing it here would couple two independent cutovers into one non-atomic contract change (violates one-feature-at-a-time + contract-first atomicity).

Per CLAUDE.md "Global Maximum, not Local Maximum" and the milestone HS-6 (scope
drift) rule: surface it as its own planned slice rather than patch it into M0.

## Scope when picked up (a) module (b) invariants (c) owning wiki
- **(a) Module:** `internal/modules/documents` (owner). Cross-module read only through its application service — never its repo/SQL directly.
- **(b) Invariants it must satisfy:** contract-first (openapi → oapi-codegen + openapi-typescript, zero hand-edits); present-and-null for any nullable ref (no omitempty); multi-tenant pooled (tenant_id predicates on the twin joins); RFC 9457 errors unchanged; fixed request lifecycle inherited.
- **(c) Read first:** `wiki/modules/documents.md`, `wiki/architecture/api-contract.md`, ADR 0065 (the governing decision — documents reuses it, no new ADR unless the shape genuinely diverges).

## Shape (mirror of ADR 0065, documents-side)
- Introduce `DocumentRevisionRef {id, version, number}` (all required) as a nested value object on the documents read model / summary DTO.
- Split a `domain.DocumentRead` (aggregate + refs) vs the write aggregate, mirroring `TemplateRead`/`Template`.
- Any nullable revision pointer → required-and-nullable (present-and-null).
- Remove whatever parallel scalar revision/version fields the documents DTOs currently expose in favor of the nested ref.

## Acceptance (when done)
Same rigor as F0.1: openapi diff + regen (no hand-edits), pin guard for present-and-null + exact ref field-set + removed-scalar absence, `go build`/`vet`(+integration)/targeted tests, FE `tsc` + vitest + zero-hit sweep, and a live drive of `GET /api/v1/documents` proving the nested ref + present-and-null on real data.

## Trigger to un-defer
Any of:
1. Start of the next global-maximum-remediation milestone that owns documents contract work; **or**
2. A documents consumer is found reading a parallel revision/version scalar in a way that can drift (the documents-side analog of the 9f86828b class); **or**
3. Pre-v1 contract-freeze checklist reaches the documents module.

When triggered: run `developing-new-work` for the written system-impact + Green/Yellow/Red verdict FIRST (Red hard-blocks), then execute as its own feature under the milestone workflow.
