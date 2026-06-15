# Milestone 2 — Contract Tail (H-D class) (Wave B)

> **Program:** grade-a-architecture-remediation  ·  **Governing spec:** `docs/superpowers/specs/2026-06-14-grade-a-architecture-remediation-design.md` (§6 M2)
> **Status:** PASSED — milestone-validator C1–C7 PASS (`qa/milestone-qa.md`, 2026-06-14); awaiting HS-1 operator gate to open M3
> **Authored:** 2026-06-14 — *before any feature in this milestone began.*

> **Execution log (not spec):** F2.1 `f2.1-usage-plantier` committed `69ad234d`; F2.2 `f2.2-presence-status`
> committed `0f0fb1ee`; F2.3 `f2.3-templates-envelope` committed `4ab670b1`. Single milestone-batched FE
> `gen:api` (the one regen) committed `728783f6` — `openapi-typescript` surfaced `plan_tier`,
> `OnlinePresenceItem.status?`, `ListTemplatesResponse`; FE `tsc --noEmit` exit 0. HS-2 (FE eigenpal
> `file:` path) did not block: `gen:api` uses the present `openapi-typescript` binary only, no `pnpm install`.
> Next: dispatch the `milestone-validator` subagent (Phase 4).

> This file is a **spec**, authored up front. It says **what** this milestone is, **which features** it
> contains, **what each feature implements**, and **what gets validated**. It contains **no execution
> steps** — the "how" of each feature lives in that feature's `plan.md`. The end-of-milestone validation
> (`qa/milestone-qa.md`) validates the milestone against *this* document.

## Objective

Eliminate the **H-D class**: delivery handlers that **emit a JSON field the OpenAPI contract never
declares** (handler-emits-undeclared-field drift). Three known instances drive the milestone, but the
target is the **class** — after M2, **0** delivery handlers emit a response field absent from the
contract (tri-source: runtime ↔ `openapi.yaml` ↔ generated type ↔ FE consumer all agree). All three
contract changes are batched behind **one** FE `gen:api` regen so the frontend takes a single,
deliberate type bump rather than three.

**Quality bar this advances:** the **contract** dimension (one of the three formerly-C dimensions).
Exact criteria that prove it moved: (a) each of the three drifting fields/shapes is now **declared** in
`openapi.yaml`, regenerated, and consumed by the FE through a generated type — no hand-typed shim; (b) a
**focused audit slice** across delivery handlers finds **0** remaining emitted-but-undeclared response
fields (root cause, class-level — not just these three); (c) the single FE regen leaves `tsc` at **0**
and changes no unrelated route's contract.

This is a **contract-first code** milestone. Skill routing: backend delivery/OpenAPI/codegen work →
`metaldocs-backend-api`; FE generated types + the screen that consumes them →
`metaldocs-tanstack-query` / `metaldocs-frontend`. The contract-first regen order (build route-truth /
field-truth table → compare runtime / spec / codegen / FE → then regen) applies to every feature.

## Features

| Feature id | Slug / folder | What to implement | What to validate (acceptance) |
|------------|---------------|-------------------|-------------------------------|
| F2.1 | `f2.1-usage-plantier` | Declare the `planTier` field — currently emitted by the `/iam/usage` handler but absent from its OpenAPI response schema — in `openapi.yaml`, regenerate the server type, and have `UsageGauges.tsx` consume it through the **generated** FE type (not a hand-typed local shape). The consumer contract (shape/enum/optionality of `planTier`) is **read from what the handler emits and the FE already consumes** — never invented. | The `/iam/usage` response schema in `openapi.yaml` declares `planTier`. Field-truth table shows runtime ↔ spec ↔ generated server type ↔ FE generated type all agree for `planTier`. `UsageGauges.tsx` reads `planTier` off the generated type with no local cast/shim. Runtime: `GET /iam/usage` returns `planTier` and the gauge renders with a real value (observable proof, by us). No other field on `/iam/usage` changed. |
| F2.2 | `f2.2-presence-status` | Declare the `status` field — emitted on `OnlinePresenceItem` but absent from the declared schema — in `openapi.yaml`; regenerate; FE consumes via generated type. | `OnlinePresenceItem` in `openapi.yaml` declares `status` with the shape/enum the handler actually emits (read from runtime, not guessed). Field-truth table agrees across runtime ↔ spec ↔ codegen ↔ FE. Runtime: the presence payload carries `status` and the generated type matches. No other `OnlinePresenceItem` field shifted. |
| F2.3 | `f2.3-templates-envelope` | Resolve the `/templates` list **envelope-vs-bare-array** drift: the handler returns one shape, the contract declares another. Make `openapi.yaml` declare the shape the handler actually returns (or, if the consumer contract requires the envelope, align both to it) — decided **contract-first from the FE consumer**, recorded in `spec.md`. Regenerate; FE consumes the generated list type. | `/templates` runtime body, `openapi.yaml`, generated type, and FE consumer all agree on **one** shape (envelope or array — chosen from the consumer contract, documented). Route-truth table for `/templates` is tri-source-consistent. Runtime: `GET /templates` returns the declared shape and the templates screen renders. No other route's contract moved. |

Single FE `pnpm gen:api` (or canonical regen command) runs **once**, after all three contract edits land;
FE `tsc` must be **0** afterward. For each feature, "what to validate" is **objectively checkable** — a
field present in the spec, a field-truth table that agrees, a generated type consumed with no shim, a
runtime payload carrying the declared field, `tsc 0`. No "works" / "looks right".

## Milestone validation definition

The close gate is run by the **`milestone-validator` subagent** (separation of powers — it judges and
writes `qa/milestone-qa.md`; the main session flips status only on its PASS), following the binding
C1–C7 checklist in `.claude/skills/milestone/references/milestone-end-validation.md`. What the gate
enforces for M2:

1. **Per-feature acceptance** — every feature meets its declared "what to validate", and each feature's
   **consumer contract** (`spec.md`) was honored: the declared field/shape matches what the FE consumer
   and runtime already use. The contract shape is **read from the consumer (FE generated-type usage +
   runtime payload), never guessed** — this is the H-D defect's own root cause, so guessing here is a fail.
2. **Workflow-class QA** — `wiki/quality/backend-api-qa-checklist.md` for all three features, plus the
   **FE screen-impact check** specifically on `UsageGauges.tsx` (F2.1's consumer) and the templates-list
   screen (F2.3's consumer): the single regen did not break either screen; `tsc` 0.
3. **Focused audit slice** (non-terminal) on the **contract** dimension: a sweep across delivery handlers
   confirms **0** emitted-but-undeclared response fields remain — the H-D **class** is closed at root
   cause, not three symptom patches leaving the class alive elsewhere.
4. **Regression** — M0 and M1 still pass their gates (docs unbroken; bare-405 sweep / presence-stream
   tri-source still hold); the single FE regen moved **no unrelated route's contract**; no swept handler's
   success path changed beyond the three declared fields/shapes.
5. **No unplanned scope** — only these three contract surfaces change; the H-G class (M4) and mechanical
   quality (M3) are not touched. Anything beyond this list is recorded with rationale or is an HS-6 stop.

## Dependencies & constraints

- **Depends on:** M1 passed (operator-approved 2026-06-14) and its test-infra-rebaseline condition
  discharged. Clean, runnable API on `:8081` via `.\scripts\start-api.ps1 -Build` is a prerequisite-truth
  boundary.
- **HS-2 watch (carry-forward):** the FE eigenpal `file:` path defer (program README HS-2) must be
  resolved **before any FE `pnpm install` / regen** in this milestone. The single `gen:api` is the regen —
  do not run it until that defer is cleared, or it is an HS-3 prerequisite repair.
- **Contract-first** (all features): build the field-truth / route-truth table first → compare runtime /
  spec / codegen / FE → then regen in canonical order. No editing generated wiring or OpenAPI shape from
  memory. The OpenAPI surface is the source; the generated types follow it.
- **One regen, batched:** all three contract edits land before the single FE `gen:api`. Do not regen per
  feature — the milestone's contract is *one* deliberate FE type bump.
- **Direction of alignment:** where the FE already consumes the field (F2.1 `planTier`, F2.2 `status`),
  the contract is aligned **to the live runtime/consumer** (declare the field). F2.3's direction
  (envelope vs array) is decided **contract-first from the consumer** and recorded in its `spec.md` —
  do not pick a direction from memory.
- **No-merge:** the operator merges; the agent never does.

## Applicable hard-stops

| HS | What would trip it here |
|----|-------------------------|
| HS-1 | This milestone's close boundary — operator review gate; no next milestone and no merge without approval. |
| HS-2 | A contract fix turns out to require redesign outside its boundary — e.g. declaring `planTier` / `status` / the `/templates` envelope forces a cross-module schema or auth-model change, or the FE consumer needs a structural rework rather than a regen. Stop; report the boundary + minimum prerequisite plan; do not symptom-patch. **Also:** the carry-forward FE eigenpal `file:`-path HS-2 must be cleared before the regen. |
| HS-3 | A prerequisite boundary fails: build, runnable API on `:8081`, dev login/session, target route, or contract/codegen truth (e.g. `gen:api` or `oapi-codegen` fails, `tsc` ≠ 0). Repair via `runtime-contract-prereq`, rerun the failed checkpoint, then resume the feature. |
| HS-4 | The focused audit slice finds a symptom-patch — an emitted-but-undeclared field still alive in a delivery handler after the three named fixes (the H-D class not closed at root), or the contract dimension not trending A−. Stop; replan the offending feature; re-run its close-out. |
| HS-6 | Scope drift — the regen quietly changing another route's contract, an H-G port (M4) or mechanical-quality item (M3) leaking into this milestone, or a fourth contract surface added without rationale. Stop; surface the deviation; replan before continuing. |
