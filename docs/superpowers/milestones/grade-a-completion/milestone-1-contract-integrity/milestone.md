# Milestone 1 — Contract / API integrity

> **Program:** grade-a-completion  ·  **Governing spec:** `../mission.md`
> **Status:** planned
> **Authored:** 2026-06-15 — *before any feature in this milestone began.*

> This file is a **spec**, authored up front. It says **what** this milestone is,
> **which features** it contains, **what each feature implements**, and **what gets
> validated**. It contains **no execution steps** — the "how" of each feature lives in
> that feature's `plan.md`. The end-of-milestone QA (`qa/milestone-qa.md`) validates
> the milestone against *this* document.

## Objective

Close the 8 skeptic-confirmed contract/API defects in mission §5 (A1–A8) and drive the
**contract-api** dimension to **≥ A−** and the **H-D class to 0**. Every public-route handler emits
its declared generated response type with the spec-declared status code; no raw domain structs and
no `map[string]any` on public routes; FE codegen is regenerated in the canonical contract-first
order (mission §10: route truth-table → compare runtime/spec/codegen/wiki → regen `api.gen.go` then
FE codegen).

**Bar:** for each defect, a regression/contract test fails before and passes after the fix; the
report §6 grep commands return **0** for the H-D class at milestone close; FE codegen regen is
clean; no public route emits raw domain or `map[string]any`. The fix lands at the contract surface
(handler returns the generated type), not via downstream FE shims or stringly typed translators.

## Features

| Feature id | Slug / folder | What to implement | What to validate (acceptance) |
|------------|---------------|-------------------|-------------------------------|
| F1.1 | `f1.1-checkpoints-typed` | List/create checkpoint endpoints serialize the generated snake_case response type — `documents/delivery/http/handler.go:881` (A1, Critical). | Wire JSON keys are snake_case matching FE codegen (`index.d.ts`); contract test asserts response ⊆ spec schema; H-D-adjacent grep for the call sites is 0. |
| F1.2 | `f1.2-status-and-body-conformance` | `renameDocument` → 200 no-content per spec (drops raw domain body) — `documents/delivery/http/handler.go:519` (A2); `templates` `createNextVersion` 201 → 200 — `templates/delivery/http/routes_create.go:36` (A4, H-D); `presignTemplateAutosave` 201 → 200 — `templates/delivery/http/routes_autosave.go:42` (A5, H-D). | Handler tests assert status code + body shape equal the OpenAPI declaration for all three routes; H-D grep on the three sites returns 0. |
| F1.3 | `f1.3-declared-fields-only` | `createTemplate` returns only the declared schema (drop undeclared `id` / `version_id`) — `templates/delivery/http/routes_generated.go:64` (A3, H-D). | Response is a strict subset of the OpenAPI schema (no extra keys); FE codegen consumers get the declared shape; H-D grep on the site returns 0. |
| F1.4 | `f1.4-typed-responses-class` | Replace all `map[string]any` / raw-domain emits with generated response types across `documents/delivery/http/handler.go:816(+5)` (A6), `taxonomy/delivery/http/routes_profiles.go:67,111,126,169` (A7, H-D), `documents/delivery/http/handler.go:317` (raw stats, A8), and add the `Allow` header on `audit/delivery/http/handler.go:81` 405 (A8). Regenerate FE codegen last. | **H-D grep (mission report §6 commands) returns 0**; FE codegen regen is clean; no public route emits raw `domain.*` or `map[string]any`; the audit 405 response includes a correct `Allow` header. |

For each feature, "what to validate" is objectively checkable — a named test that passes plus an
observed runtime/grep result. No "works" / "looks right".

## Milestone validation definition

The close gate is run by the **`milestone-validator` subagent** (separation of powers — it judges
and writes `qa/milestone-qa.md`; the main session flips status only on its PASS), following the
binding C1–C7 checklist in `.claude/skills/milestone/references/milestone-end-validation.md`. What
that gate enforces for this milestone:

1. **Per-feature acceptance** — every feature above meets its declared "what to validate", and each
   feature's consumer contract (`spec.md`) was honored.
2. **Workflow-class QA checklist** — [`wiki/quality/backend-api-qa-checklist.md`](../../../../wiki/quality/backend-api-qa-checklist.md)
   with a **contract-truth lens**: route truth-table reconciled across runtime / spec / codegen /
   wiki; regen order honored; no hand-edits to generated wiring; OpenAPI shape unchanged except
   where a feature explicitly amends it.
3. **H-D class-zero proof** — the mission report §6 grep commands return **0** at milestone close.
4. **Regression** — whole-repo `go test ./...` green; M0's authz / session corpus stays green; FE
   codegen regen is clean (no manual patches).
5. **Quality-bar / root-cause check** — fixes land at the handler/contract surface, not by FE-side
   shims, stringly typed translators, or symptom patches around `map[string]any`.
6. **No unplanned scope** — anything implemented beyond these four features (esp. anything touching
   M2/M3/M4 finding classes) is recorded with rationale.

## Dependencies & constraints

- Depends on: M0 passed (HS-1 approved 2026-06-15). HEAD includes the M0 close commits; every
  `file:line` above is from mission §5 / report §3 — re-verify at feature start.
- Architectural constraints respected:
  - **Contract-first regen order (mission §10):** build route truth-table → compare
    runtime/spec/codegen/wiki → regen `api.gen.go` then FE codegen. **No editing generated wiring
    or OpenAPI shape from memory.**
  - **No OpenAPI shape drift** beyond what a feature explicitly amends (e.g. F1.3 drops undeclared
    fields — that is a shape correction, not a shape change).
  - **No FE-side shims** to hide a wrong server contract; the fix lives at the handler.
  - **No schema/migration redesign** (mission Non-Goals).
  - **Skill routing:** backend HTTP/handler/contract → `metaldocs-backend-api`; FE codegen / query
    types → `metaldocs-tanstack-query`; prereq repair → `runtime-contract-prereq`; module-wiki sync
    after structural change → `metaldocs-module-doc-sync`.

## Applicable hard-stops

| ID | What would trip it here |
|----|--------------------------|
| HS-1 | This milestone's boundary — operator review gate after the validator PASS; no next milestone / no merge without approval. |
| HS-2 | If fixing a contract defect implies redesigning a shared API contract beyond the named site (e.g. a checkpoint type change that ripples into unrelated domain APIs, or a status-code change that would force a versioned route) — stop, report the boundary + minimum prerequisite plan, do not symptom-patch via FE shims. |
| HS-3 | If a prerequisite boundary fails (build / runnable / auth-session / route / contract truth, e.g. codegen drift before fixes start) — repair via `runtime-contract-prereq`, rerun the failed checkpoint, resume the feature. |
| HS-4 | If `milestone-validator` returns FAIL — open the named fix feature, re-run its lifecycle, re-dispatch the validator. |
| HS-6 | If a fix uncovers a contract defect F5.1 missed, or scope drifts off these four features (e.g. an observability/quality finding) — stop, surface the deviation, replan before continuing. |
