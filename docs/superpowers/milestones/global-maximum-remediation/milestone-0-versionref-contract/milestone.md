# Milestone 0 — VersionRef contract refactor

> **Program:** global-maximum-remediation  ·  **Governing spec:** `../mission.md` (§7 M0)
> **Status:** Spec approved
> **Authored:** 2026-07-03 — *before any feature in this milestone began.*

> This file is a **spec**, authored up front. It says **what** this milestone is, **which
> features** it contains, **what each feature implements**, and **what gets validated**. It
> contains **no execution steps** — the "how" lives in each feature's `plan.md` (here: the
> committed `docs/superpowers/plans/2026-07-03-versionref-template-contract.md`). The
> end-of-milestone QA (`qa/milestone-qa.md`) validates the milestone against *this* document
> and against the committed `validation-contract.md` (mission D4).

## Objective

Land the already-planned nested version-reference contract cutover for **templates**, proving
the plan→review→implement loop on a ready workpiece as M0 of the mission (mission.md §7 M0,
finding 1). The wire contract shape is the local-maximum artifact being replaced: `TemplateDTO`
carries five parallel coupled version/revision scalars with an implicit tri-field null-coupling
invariant that no consumer can see in the schema — the class that produced the 2026-07-03 HIGH
bug (unpublished template selectable in the controlled-document wizard, fixed 9f86828b).

**Bar moved:** ADR 0035's optional-vs-null drift subclass becomes **structurally closed** for
templates — the null-coupling invariant is made unrepresentable on the wire (`published_version`
is one required-and-nullable nested object; consumers gate on the object, never on inner fields).
Proof criterion: marshal-shape pin tests assert `published_version` present-and-null for a
never-published template and the exact nested ref field-set, and the four removed flat keys are
absent everywhere (backend pins + FE consumers + live drive).

Documents-side migration of the same *pattern* (not schema) is a bounded defer to Plan 2
(mission scope note: M0 executes Plan 1 only; the documents `DocumentRevisionRef` follow-up is
explicitly time-boxed to pre-v1 by ADR 0065 and Plan Task 13).

## Features

| Feature id | Slug / folder | What to implement | What to validate (acceptance) |
|------------|---------------|-------------------|-------------------------------|
| F0.1 | `f0.1-versionref-cutover` | Execute the 13-task plan: `TemplateVersionRef` component schema; `TemplateDTO` reshaped to `latest_version: TemplateVersionRef` (required) + `published_version: TemplateVersionRef \| null` (required-nullable), four flat scalars removed; BE+FE regeneration (zero hand-edits to generated files); domain `VersionRef` value object + `TemplateRead` read model; repository projects refs; delivery mapper emits nested objects; FE consumers gate on the single nullable object; ADR 0065 authored (F0.2) | Marshal-shape pin tests pass (`published_version` present-and-null; nested ref field-set `{id,number,revision_number,status}`; four removed keys absent); `go build ./...` green; targeted `go test ./internal/modules/templates/...`; `pnpm exec tsc --noEmit` clean; `pnpm exec vitest run src/features/{documents,templates,taxonomy}` green; openapi lint valid; live drive `GET /api/v1/templates` shows nested refs + `published_version:null` key-present, and `/documents/new` wizard Step 3 shows unpublished template NOT selectable with status-precise badge |
| F0.2 | `f0.2-adr-0065` | ADR 0065 "Version references are nested value objects in wire contracts" incl. pre-v1 atomic-cutover exception; annotate ADR 0035 subclass structurally closed | ADR 0065 exists under `wiki/decisions/`, indexed, cited by F0.1 cutover commit; ADR 0035 memory/doc annotated closed for this class |

"What to validate" for each feature is enumerated in exact wire/behavior detail in
`./validation-contract.md` (mission D4) — that file is the binding acceptance contract;
implementation drift is compared against it, never rationalized after (HS-7).

## Milestone validation definition

The close gate is run by the **`milestone-validator` subagent** (separation of powers — it judges
and writes `qa/milestone-qa.md`; the main session flips status only on its PASS), following the
binding C1–C7 checklist in `.claude/skills/milestone/references/milestone-end-validation.md`. What
that gate enforces for this milestone:

1. **Per-feature acceptance** — F0.1 and F0.2 each meet their declared "what to validate", and each
   feature's consumer contract (`spec.md`) was honored (producer matches the consumer shape the FE +
   pin tests require).
2. **Validation-contract compliance (D4/HS-7)** — the shipped wire shapes, pin-test behaviors, FE
   gating, and live-drive outputs match `./validation-contract.md` exactly; any deviation is HS-7.
3. **Workflow-class QA checklist** — backend-api contract subset (contract gate + docs gate per the
   gate artifact §8; authz/multi-tenant/async/DB-invariant gates N/A — untouched).
4. **Regression** — `go build ./...` green from clean state; targeted templates tests green; no
   prior milestone exists to regress (M0 is first).
5. **Quality-bar / root-cause check** — ADR 0035 optional-vs-null subclass confirmed **structurally
   closed** (the invariant is unrepresentable on the wire), not symptom-patched; the 9f86828b
   null-serialization guarantee carries forward and is pinned.
6. **No unplanned scope** — anything implemented beyond the 13-task plan is recorded with rationale;
   documents-side migration stays deferred to Plan 2.

## Dependencies & constraints

- **Depends on:** the committed Yellow gate artifact
  `docs/superpowers/analysis/2026-07-03-versionref-contract-refactor-system-impact.md` (8 locked
  constraints, binding) and the committed plan
  `docs/superpowers/plans/2026-07-03-versionref-template-contract.md` (gitignored — never force-add).
- **Architectural constraints respected:**
  - Contract-first: every shape change lands in `api/openapi/v1/openapi.yaml` first, then regenerate
    (`go generate` per module api pkg + FE `pnpm run gen:api`); **zero hand-edits** to generated files.
  - Two schemas, one pattern: `TemplateVersionRef` ≠ `DocumentRevisionRef`; no shared cross-context
    schema or Go package (locked constraint 2).
  - `published_version` required-and-nullable (present-and-null, never absent) — locked constraint 3.
  - No DB migration; SQL projections gain columns only (locked constraint 6, as amended by plan Task 1).
  - Repo builds Go-green at every commit; breaking wire change is a single atomic cutover (pre-v1
    exception, ADR 0065).
- **Test discipline:** contract/invariant guards repaired onto the new shape; one-off scaffolding
  tests that break may be deleted (legacy-test policy); full integration suite NOT run (20+ min box
  constraint) — targeted `-run` filters; bounded defer recorded in evidence.

## Applicable hard-stops

| ID | What would trip it here |
|----|-------------------------|
| HS-1 | Milestone boundary — operator review gate after validator PASS; no M1, no push without approval. |
| HS-2 | If the cutover implies redesign outside templates/documents wire contracts (e.g. a shared Go DTO package, an authz or DB-schema change) — stop, report, don't patch around. |
| HS-3 | Prerequisite boundary fails (build/runnable/auth-session/route/contract truth) — e.g. `start-api.ps1 -Build` fails, or oapi-codegen renders the nullable ref as an inline anon struct instead of a named pointer — repair prerequisite first, rerun checkpoint, resume. |
| HS-4 | `milestone-validator` returns FAIL — open the named fix feature, re-run its lifecycle, re-dispatch. |
| HS-6 | Scope drift / off-plan discovery mid-milestone (e.g. documents-side work creeping in) — stop, surface, replan. |
| HS-7 (mission) | Implementation deviates from the committed `validation-contract.md` — stop; fix the code to the contract, or re-open the contract WITH operator approval — never silently adjust the contract to match the code. |
