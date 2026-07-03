# Milestone 2 — Authz enforcement generation & cap-name coherence

> **Program:** global-maximum-remediation  ·  **Governing spec:** `../mission.md` (§7 M2)
> **Status:** **IN-PROGRESS** — spec + `validation-contract.md` authored 2026-07-03, *before any implementation*.
> **Authored:** 2026-07-03 — *before any feature in this milestone began.*

> This file is a **spec**, authored up front. It says **what** M2 is, **which features** it
> contains, **what each feature implements**, and **what gets validated**. It contains **no
> execution steps** — the "how" of each feature lives in that feature's `plan.md`. The end-of-
> milestone QA (`qa/milestone-qa.md`) validates the milestone against *this* document and against
> the binding `validation-contract.md` (D4). Drift between implementation and that contract is
> **HS-7**.

## Objective

Make the **capability tripwire** — the last-line DB trigger `enforce_capability_asserted()` — stop
being a **hand-synced** per-table cap list and become **generated from a single Go source of truth**,
with a **CI drift check** that fails when a gated table has an asserted capability with no matching
arm. This closes the exact defect class that shipped **two production incidents in the last month**
(migration 0269: missing `template.review`; migration 0270: missing `template.archive` — each a
`P0001` 500 for *every* actor, found only by live QA) and — as this spec's investigation proves —
is **currently shipping two more latent instances of the same bug** (see Discovered runtime truth).

The bar this milestone advances: the review's **Dimension 2 (AuthZ enforcement)** cross-cutting
finding #1 ("the hand-sync defect class — derive the arms from the same registry the service reads;
CI-fail on drift") moves from **DEBT** toward **CONFIRMED**. The generated migration is byte-faithful
to the current trigger except for the arm literals it is *supposed* to change; the drift check is
green on a clean tree **and** has a recorded negative proof (RED on a synthetic gated write whose
asserted cap has no arm); the two existing `tripwire_caps_test.go` integration drives stay green; and
the two newly-discovered latent incidents are fixed **by the generation**, not by another hand-patch.

Alongside, close the **tier-1 ↔ tier-2 capability-name divergences** the 2026-07-03 review flagged
(force-release; approval-route management) so the route→capability table and the in-tx `authz.Require`
assertions **agree**, or the difference is a **written, intentional** coarse/fine split — proven by a
regression pin so the agreement cannot silently rot.

Source findings: review §Dimension 2; §Cross-cutting item 1; §Priorities P1 (hand-sync drift-proofing).

## Discovered runtime truth (recorded before implementation — HS-6 surface, not silent expansion)

Investigation while authoring this spec (2026-07-03; full census in `validation-contract.md §Census`)
established ground truth that **refines the mission's per-feature sizing**. Recorded here so the
expansion is visible to the validator and the operator, not absorbed silently:

1. **The tripwire is currently shipping TWO more instances of the 0269/0270 bug — not "a third", a
   third and a fourth.** The `documents (UPDATE)` arm is `{document.edit}` (migration 0270, line 71),
   but two production write-paths assert a *different* capability and never defensively re-assert
   `document.edit`:
   - **`membership.manage`** — force-release of a stuck editing session
     (`documents/repository/repository.go:798` `ForceReleaseSession`, `:828` `ForceReleaseSessionTx`)
     UPDATEs `documents.active_session_id` asserting **only** `CapMembershipManage` (deliberate, per
     ADR 0022 Phase 11 F4). Recorded cap `membership.manage` ∉ `{document.edit}` → `P0001` 500.
   - **`document.obsolete`** — `documents/approval/application/obsolete_service.go:88` asserts **only**
     `CapDocumentObsolete`, then `:93` UPDATEs `documents SET status='obsolete'`. Recorded cap
     `document.obsolete` ∉ `{document.edit}` → `P0001` 500, *unconditionally, for every actor*
     (the trigger checks the recorded cap set, not the role — identical mechanism to 0269/0270).

   Both are **function-local** assert-then-write (the `authz.Require` and the mutating SQL are in the
   same Go function), i.e. exactly the shape the F2.1 drift check is built to catch, and exactly the
   shape a correct **generated** arm fixes for free: the arm becomes the census union of caps that
   solely-authorize a write to that table. The corrected `documents (UPDATE)` arm is therefore
   `{document.edit, membership.manage, document.obsolete}`. This is **additive** (matches the
   0269/0270 convention: widen the arm to admit a legitimately-asserted cap; never a tightening),
   so no currently-passing path can regress.

2. **F2.2's two divergences are already CODE-resolved; the residual defect is doc-truth drift.** The
   2026-07-03 review flagged force-release (tier-1 `membership.manage` vs tier-2 `document.edit`) and
   approval-route management (tier-1 `document.submit`) as **open** tier-1↔tier-2 divergences. Runtime
   truth contradicts the review: **both were closed in ADR 0022 Phase 11 F4 (2026-06-04)** —
   force-release tier-2 was changed to `CapMembershipManage` (so tier-1 `permissions.go:157` ==
   tier-2 `repository.go:798/828` == `membership.manage`), and four explicit `/approval/routes` tier-1
   rows were added resolving to `CapRouteManage` (== the tier-2 `route.manage` assertion), ordered
   before the generic `/approval/` fallback and pinned by `permissions_test.go`. The review read the
   **stale Phase 7/8 ⚠️-follow-up lines** of ADR 0022 (≈198, 236–237, 250) that were never
   back-annotated as resolved when Phase 11 closed them. So **F2.2's premise is contradicted by
   runtime truth**: the divergences are closed. F2.2's honest content is therefore **verify the
   current alignment + add a regression pin so it cannot silently reopen + restore the ADR/wiki
   doc-truth** — not "close two live divergences." The general tier-1-coarse / tier-2-fine differences
   that remain on other `/approval/` operations (e.g. tier-1 `document.submit` gate, tier-2
   `document.signoff` on decision) are the **deliberate** coarse/fine PDP split the review's Dimension
   2 explicitly vindicated — by-design, not a defect, and out of scope to "align".

None of the above crosses a feature's boundary into redesign (no HS-2): F2.1 is the assigned
generation feature executed against runtime truth (the arm it generates is simply *correct*), and F2.2
is the assigned coherence feature executed against runtime truth (the coherence already holds; the
feature verifies, pins, and records it). Both expansions are reported at the HS-1 operator gate. The
arm-pruning of harmless supersets (e.g. `template.submit` on `templates_template`, which no writer
asserts) is a **deliberate non-goal** here (tightening carries residual regression risk the census
cannot fully exclude) and is recorded as a bounded defer to M9 arm-hygiene.

## Features

| Feature id | Slug / folder | What to implement | What to validate (acceptance) |
|------------|---------------|-------------------|-------------------------------|
| F2.1 | `f2.1-tripwire-generation` | A single Go **`TripwireArms`** source-of-truth map (gated `(table, op)` → required-cap set), from which the trigger migration is **generated** (codegen the `CREATE OR REPLACE enforce_capability_asserted()` arm literals — SQL no longer hand-typed) as forward-only migration **0271**, regenerated from 0270 byte-for-byte except the arm literals. Plus a **blocking CI drift check** in `scripts/api-lint` (Go AST, same framework as the 5 authz lints + the existing `checkTripwirePairing`) that (a) asserts every registry-real cap referenced by `TripwireArms` exists in the iam capability registry and the generated SQL matches `TripwireArms` (golden/parity), and (b) scans `authz.Require` call sites and fails when a **function-local** gated-table write asserts a capability absent from that table's arm. The corrected `documents (UPDATE)` arm (`{document.edit, membership.manage, document.obsolete}`) fixes the two latent incidents. | Drift check **RED** on a synthetic new asserted cap on a gated-table write lacking an arm (negative proof, captured output); **GREEN** on the clean tree; the generated 0271 is **byte-faithful** to 0270 except the documents(UPDATE) arm (diff shown); `TripwireArms`↔SQL parity holds; `TestCapabilityRegistrySize` (34) and the existing `tripwire_caps_test.go` drives (0269/0270) stay **green**; two new integration drives (force-release, obsolete) **pass** post-0271 (each proven RED pre-0271 → GREEN post). |
| F2.2 | `f2.2-cap-name-divergence` | **Verify** the tier-1 route→capability table (`permissions.go`) and the in-tx `authz.Require` assertions agree for the two historically-divergent sites (force-release → `membership.manage`; approval-route management → `route.manage`); add a **regression pin** binding tier-1 cap == tier-2 asserted cap for exactly those reconciled routes so a future edit that re-diverges them reddens CI; **restore doc-truth** — back-annotate the stale ADR 0022 Phase 7/8 ⚠️-follow-up lines as RESOLVED-in-Phase-11 and correct any wiki page that still describes them as open. | The two sites' tier-1 and tier-2 capability names **agree** (shown from source: `permissions.go:157`==`repository.go:798/828`; `/approval/routes` rows==the `route.manage` assertion); the regression pin **fails** on a synthetic re-divergence and **passes** on HEAD; **all 5 authz CI lints stay green**; ADR 0022 no longer contains an un-annotated "open divergence" claim for either site; the deliberate coarse/fine `/approval/` differences are recorded as intentional (written exception), not "fixed". |

For each feature, "what to validate" is objectively checkable — a gate that fails on the negative
fixture and passes on the clean tree, with captured command output as evidence (positive + negative
proof per D4). F2.1 additionally requires **integration proof** on a tripwire-enforced DB (the two new
latent-incident drives) — that is the only class of test that can pin a `P0001` arm regression
(application tests are sqlmock).

## Milestone validation definition

The close gate is run by the **`milestone-validator` subagent** (separation of powers — it judges and
writes `qa/milestone-qa.md`; the main session flips status only on its PASS), following the binding
C1–C7 checklist in `.claude/skills/milestone/references/milestone-end-validation.md`. For M2:

1. **Per-feature acceptance** — every feature meets its declared "what to validate", each demonstrated
   **failing-then-passing from clean state** (negative fixture/probe RED → clean tree GREEN) with real
   captured output. F2.1's two new latent-incident drives are each shown RED against 0270 and GREEN
   against 0271 (real integration run, labeled real-DB not sqlmock).
2. **Validation-contract conformance (D4)** — implementation is checked against `validation-contract.md`
   section-by-section, including the **exact `TripwireArms` table** (all 18 gated `(table,op)` arms)
   and the drift-check RED/GREEN definitions; any divergence is **HS-7** (fix code to the contract, or
   re-open the contract WITH operator approval — never silently adjust the contract to match code).
3. **Workflow-class QA** — backend-api authz/DB-invariant class. QA re-runs the deterministic gates
   from a clean tree: `go run ./scripts/api-lint -strict api/openapi/v1/openapi.yaml .` (the new drift
   rule blocking, zero live violations), `go test ./scripts/api-lint/...`, the migration-generation
   golden/parity check, `go build ./...`, and the **targeted** integration drives
   (`go test -tags integration -run 'Tripwire' ./tests/integration/templates/...` and the new
   documents drives) — **not** the full 20-min suite (mission §10).
4. **Regression** — M0 (VersionRef) and M1 (contract/FE gates) still pass their gates; the 5 authz CI
   lints stay green; `TestCapabilityRegistrySize` (34) unchanged; no route/contract shape regresses.
5. **Root-cause check** — the hand-sync class is **structurally** closed: the trigger arm literals are
   now **generated** from `TripwireArms` (a hand-edit to the SQL that drifts from the map is caught by
   the parity check; a new gated write with an unarmed cap is caught by the AST drift check). The two
   latent incidents are fixed **by the generation**, not by a one-off arm patch. The drift check's
   power is proven live (synthetic unarmed cap → RED), not gutted to pass.
6. **No unplanned scope** — anything implemented beyond this list is recorded with rationale. The
   superset-arm pruning defer (M9) and any AST cross-layer coverage limitation are recorded bounded
   defers with triggers.

## Dependencies & constraints

- **Depends on:** M0 + M1 passed and committed. F2.1 operates on the current trigger (migration 0270)
  and the iam capability registry (`internal/modules/iam/domain/model.go`, 34 caps).
- **Tripwire = DB migration, forward-only (no down).** 0271 is `CREATE OR REPLACE` regenerated from
  0270 **byte-for-byte except the arm literals it is meant to change** (mission workflow-pt-4). No
  trigger attachments change; no backfill; ledger insert + `BEGIN/COMMIT` wrapper preserved.
- **Contract-first / registry-first (non-negotiable):** capability names come **only** from the iam
  registry; `TripwireArms` references registry consts, never raw strings the linter would reject
  (`no-rawstring-capability`). The generated SQL is the *output*; the Go map is the *source*.
- **CI-truth:** every gate this milestone adds is **blocking** by construction (api-lint `main.go`:
  "the model is bound by CI, not by discipline"). No reported-only tier.
- **Targeted tests only:** the full integration suite is **not** run (20+ min box, mission §10).
  F2.1's integration proof is scoped with `-run 'Tripwire'` + the two new drives. If the box cannot
  run integration locally, the drive is authored + the block recorded as a bounded defer with the
  run-trigger (matching M1's env-risk precedent).
- **Model policy:** sonnet implement/review; haiku mechanical sweeps; **never fable** workers; ≤15
  concurrent. Subagent-driven implementation (`superpowers:subagent-driven-development`); the main
  session orchestrates, reviews between features, and commits — it does **not** implement inline.
- **Commit after verified work** (standing auth); **never push**; **never commit `docs/release/`**;
  plans dir is gitignored (never force-add).
- **Advisory-lock constraint (H-PRE-1) holds:** no authz-recording read inside a lock-holding atomic
  tx. F2.1 touches the trigger definition + a lint, not the tx-lock discipline.

## Applicable hard-stops

| ID | What would trip it here |
|----|-------------------------|
| HS-1 | Milestone boundary — operator review gate after validator PASS; no M3 and no push without approval. |
| HS-2 | A fix implies redesign outside its boundary — e.g. the drift check would require call-graph analysis to resolve cross-layer assert→write edges (that is a bigger static-analysis effort), or the arm generation would require rewriting the trigger to a generic `cap_table_requirements` lookup (a load-bearing security-trigger redesign). Stop; report; do not patch across the boundary. The function-local drift check + additive generation exist precisely to stay inside the boundary. |
| HS-3 | A prerequisite fails from clean state — `go build ./...` red, `go generate` drift, the tripwire DB not applyable, or an M0/M1 gate regressed. Repair the prerequisite first; rerun; resume. |
| HS-4 | `milestone-validator` returns FAIL — open the named fix feature, re-run its lifecycle, re-dispatch the validator. |
| HS-6 | Scope drift / off-plan discovery beyond the two recorded latent incidents (F2.1) and the doc-truth-restore reframing (F2.2). Stop; surface; replan. |
| HS-7 (mission) | Implementation deviates from `validation-contract.md` (esp. the `TripwireArms` arm table) — fix code to the contract, or re-open the contract WITH operator approval; never silently adjust the contract to match the code. |
