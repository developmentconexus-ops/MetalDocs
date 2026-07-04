# Milestone 4 — Versioning kernel correctness

> **Program:** global-maximum-remediation  ·  **Governing spec:** `../mission.md` (§7 M4)
> **Status:** Spec (drafting)
> **Authored:** 2026-07-04 — *before any feature in this milestone began.*

> This file is a **spec**, authored up front. It says **what** this milestone is,
> **which features** it contains, **what each feature implements**, and **what gets
> validated**. It contains **no execution steps** — the "how" of each feature lives in
> that feature's `plan.md`. The end-of-milestone QA (`qa/milestone-qa.md`) validates
> the milestone against *this* document. The full expected behaviors are pinned in
> `validation-contract.md` (D4), authored and committed **before** the first feature's
> implementation.

## Objective

Close the two versioning-kernel correctness risks the 2026-07-03 architecture review
(§6, dimension 6) named RE-LITIGATE, plus its minor concurrency-idiom DEBT:

1. **The 9-status document lifecycle is not one state machine.** `CanTransitionDocument`
   (`internal/modules/documents/domain/state.go`) covers only 3 of 9 statuses; the real
   approve/publish/supersede/obsolete/schedule/reject transitions live as scattered
   `if status != X` guards across the approval services. Templates already has the right
   pattern (`TemplateVersion.CanTransition`). **Bar moved:** one exhaustive transition
   function covers all 9 statuses × all transitions, proven by a coverage test; zero
   scattered lifecycle status-equality guards remain in the approval services.

2. **Scheduled-publish vs manual-publish race is unverified.** Scheduler takes
   `FOR UPDATE` + `status == scheduled`; the manual path checks `status == approved`
   independently — no demonstrated shared choke point. **Bar moved:** a real concurrent
   integration test exercises both interleavings on one revision; at most one publish
   wins with a correct terminal state — proven safe, or made safe by a single
   `PublishRevision` choke point both paths route through.

3. **Two concurrency-transport idioms for one mechanism** (documents `If-Match "vN"`
   header vs templates body `expected_lock_version`). **Bar moved:** one idiom across
   documents + templates, or an ADR-recorded exception.

Coherent slice: it is exactly the versioning-kernel correctness cluster (findings 8, 9,
10), no more. Enforcement-gate work (M1–M3) precedes it; async consolidation (M5)
follows. DB triggers/constraints stay the last line of enforcement — this milestone
consolidates the **app-layer friendly-first-line** checks into one function; it does not
move invariants out of the DB.

## Features

| Feature id | Slug / folder | What to implement | What to validate (acceptance) |
|------------|---------------|-------------------|-------------------------------|
| F4.1 | `f4.1-state-machine-unification` | Single exhaustive transition function over all 9 document statuses (pattern = templates' `CanTransition`); the approval services route every lifecycle transition through it. | Table-coverage test proves all 9 statuses × transitions are handled (every legal edge allowed, every illegal edge rejected, per the `validation-contract.md` §1 table); **no scattered `if status !=` / status-equality lifecycle guards remain in the approval services** (grep census = 0 or explicitly allowlisted with reason); `go build ./...`; targeted go tests green. |
| F4.2 | `f4.2-publish-race` | Concurrent integration test: scheduled-publish vs manual-publish racing ONE revision, BOTH interleavings. If the paths are not already safe against one choke, add a single `PublishRevision` method both route through. | Real concurrent integration test (testdb factory, NOT sqlmock) green; deterministically exercises both orders; **at most one publish wins**; loser is rejected/no-ops with the correct terminal state (exactly one published revision, no double-publish, no lost transition). Contract §2 states the expected single-winner + terminal state before implementation. |
| F4.3 | `f4.3-concurrency-idiom` | Unify the optimistic-concurrency transport across documents + templates (decide `If-Match` header vs body `expected_lock_version`; migrate the minority contract-first), **or** ADR-record the split as an intentional exception. | One idiom across documents + templates (contract-first: `openapi.yaml` + regen, zero hand-edits to generated files; consumers updated), **or** a written ADR exception with the trade-off. Contract §3 states the decision (which idiom, migrate-vs-ADR) before implementation. |

For each feature, "what to validate" is objectively checkable: a coverage test that
passes, a real concurrent race test that yields exactly one winner, a single wire idiom
(or an ADR). No "works"/"looks right".

## Milestone validation definition

The close gate is run by the **`milestone-validator` subagent** (separation of powers —
it judges and writes `qa/milestone-qa.md`; the main session flips status only on its
PASS), per the binding C1–C7 checklist in
`.claude/skills/milestone/references/milestone-end-validation.md`. For M4:

1. **Per-feature acceptance** — every feature meets its declared "what to validate", and
   each feature's consumer contract (`spec.md`) was honored. F4.1's coverage test
   matches the `validation-contract.md` §1 nine-status table cell-for-cell; F4.2's race
   test matches §2's expected single-winner + terminal state; F4.3 matches §3's idiom
   decision.
2. **Workflow-class QA checklist** — `wiki/quality/backend-api-checklist.md` (route/
   contract + handler paths) and `wiki/quality/test-discipline.md` (testdb factory for
   the DB race integration; targeted `-run` only, no full suite).
3. **Regression** — M0–M3 gates still pass; the VersionRef contract (M0) and the tenancy
   chokepoint (M3) are not regressed by any approval-service change; `go build ./...`
   green.
4. **Quality-bar / root-cause check** — the fragmentation is fixed at the **root** (one
   transition function the services route through), NOT symptom-patched (a 4th scattered
   guard, or a coverage test over the old scattered guards, is a FAIL). Publish safety is
   proven by a **real** concurrent test, not asserted.
5. **No unplanned scope** — anything beyond F4.1–F4.3 recorded with rationale.

**HS-7 (mission-specific):** the implementation is compared **section-by-section** against
the committed `validation-contract.md`. Any drift = stop; fix the code to the contract, or
re-open the contract **with operator approval** — never silently edit the contract to match
the code.

## Dependencies & constraints

- **Depends on:** M0 (VersionRef contract — F4.3 idiom decision must respect the nested
  version-ref wire shape); M3 (tenancy chokepoint — approval-service refactors must keep
  the auto-seeded GUC / RLS behavior intact). M0–M3 committed, HS-1 pending on M3.
- **DB enforces invariants (binding):** the DB triggers/constraints on document status
  (publish uniqueness, immutability, SoD guards) stay the **last line**. This milestone's
  transition function is the **friendly first line** — it must not remove, weaken, or
  duplicate-as-authoritative any DB enforcement. Verified in `validation-contract.md` §4.
- **Contract-first:** any route/wire change (F4.3 if migrating) is `openapi.yaml` +
  `oapi-codegen` regen only — zero hand-edits to generated files.
- **ADR 0013** governs revision numbering — semantics unchanged by this milestone.
- **Test discipline:** testdb factory for the F4.2 DB race integration (real concurrent,
  not sqlmock); targeted `-run` filters only; the full integration suite is NOT run
  locally (20+ min box) — bounded defers recorded in evidence.
- **Model policy:** implement via subagents (sonnet implement/review, haiku mechanical,
  never fable, ≤15 concurrent); main session orchestrates/reviews/commits.
- **Commit after verified work; never push; never commit `docs/release/`; plans dir
  gitignored (never force-add).**

## Applicable hard-stops

| ID | What would trip it here |
|----|-------------------------|
| HS-1 | Milestone boundary — operator review gate after validator PASS; no M5, no merge/push without approval. |
| HS-2 | A fix implies redesign outside M4's boundary — e.g. F4.2 reveals the publish paths must move enforcement into or out of the DB, or the choke point requires a cross-module (jobs↔documents) contract change beyond routing both through one method. Stop, report the boundary + minimum prerequisite, no symptom-patch. |
| HS-3 | A prerequisite boundary fails (build / runnable / route / contract truth) — repair first, rerun, resume. |
| HS-4 | `milestone-validator` returns FAIL — open the named fix feature, re-run its lifecycle, re-dispatch the validator. |
| HS-6 | Scope drift / off-plan discovery mid-milestone (e.g. a 10th status found, or a second publish race surface) — stop, surface, replan before continuing. |
| HS-7 | Implementation deviates from the committed `validation-contract.md` — fix code to contract, or re-open contract with operator approval; never silently adjust the contract. |
