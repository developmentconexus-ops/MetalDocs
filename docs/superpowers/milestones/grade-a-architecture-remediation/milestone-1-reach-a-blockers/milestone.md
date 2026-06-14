# Milestone 1 — Reach-A Blockers (Wave A)

> **Program:** grade-a-architecture-remediation  ·  **Governing spec:** `docs/superpowers/specs/2026-06-14-grade-a-architecture-remediation-design.md` (§6 M1)
> **Status:** Spec approved
> **Authored:** 2026-06-14 — *before any feature in this milestone began.*

> This file is a **spec**, authored up front. It says **what** this milestone is, **which features** it
> contains, **what each feature implements**, and **what gets validated**. It contains **no execution
> steps** — the "how" of each feature lives in that feature's `plan.md`. The end-of-milestone validation
> (`qa/milestone-qa.md`) validates the milestone against *this* document.

## Objective

Close **all 4 Grade-A blockers** plus the error-contract tail, by killing the offending **classes**, not
just the instances. After M1: zero bare-`405` anywhere in the swept delivery packages, `GET
/iam/presence/stream` consistent across runtime/openapi/codegen/FE (tri-source), and the cross-module
`iam_users` read in the approval signoff path contained out of the lock-holding transaction.

**Quality bar this advances:** the **contract** and **error-contract** facets. Exact criteria that prove
it moved: (a) a `grep` across the swept packages finds **0** bare-`405` writes — root cause, class-level;
(b) the presence/stream route is tri-source-consistent on a built-route-truth-table; (c) the signoff
display-name read no longer executes inside the advisory-lock atomic tx (H-PRE-1 honored).

This is a **code** milestone — first one in the program. Skill routing: backend delivery/OpenAPI/codegen
work → `metaldocs-backend-api`; FE generated types → `metaldocs-tanstack-query`. Contract-first regen
order applies to F1.2.

## Features

| Feature id | Slug / folder | What to implement | What to validate (acceptance) |
|------------|---------------|-------------------|-------------------------------|
| F1.1 | `f1.1-bare-405-sweep` | Replace every bare `w.WriteHeader(405)` / hand-rolled 405 with a canonical `problem+json` 405 via the shared helper, across the delivery packages that emit them: `auth/delivery/http/handler.go` (login/logout/me/change-password), `iam/delivery/http/admin_handler.go` (admin/overview), `iam/.../sessions_handler.go`, `featureflags/handler.go`, `observability/http.go`. Kill the **class**, not the 7 listed sites only. | `grep` proves **0** bare-`405` / direct `WriteHeader(405)` in all swept packages. Hitting a known route with the wrong method returns **405** with `Content-Type: application/problem+json` and the canonical problem body shape (type/title/status/detail). backend-api-qa-checklist clean for the touched handlers. No declared route's success path changed. |
| F1.2 | `f1.2-presence-stream-spec` | Declare the already-live, FE-consumed `GET /iam/presence/stream` in `openapi.yaml`; regenerate the server stub + FE type via the contract-first flow; the 101 WebSocket upgrade path stays live (statusWriter `Unwrap()` already merged — do not redesign it). | Built **route-truth-table** shows runtime ↔ `openapi.yaml` ↔ generated server ↔ FE type all agree for `/iam/presence/stream` (tri-source-consistent, was: runtime-only). `oapi-codegen` + FE `gen:api` run clean; FE `tsc` 0. Runtime: the stream endpoint still upgrades to 101 and streams (observable proof, by us). No other route's contract shifted in the regen. |
| F1.3 | `f1.3-approval-displayname-reach` | **Contained** fix only (Approach-3 **step 1**): move the raw `SELECT … FROM metaldocs.iam_users` (signoff display-name read, `documents/approval/application/decision_service.go:266-272`) out of the lock-holding signoff transaction into a contained method on the approval module's own repository. **Do not** generalize to a shared port — that is M4/F4.1 and generalizing here is scope drift. | The raw `iam_users` SELECT no longer executes **inside** the advisory-lock atomic tx (H-PRE-1 honored — read is off-tx / pre-flight). Signoff still returns the correct approver display name (runtime proof). CD-create / signoff stays fast, no deadlock (runtime-observed). The read lives on `ApprovalRepository`, not inline in `decision_service`. |

For each feature, "what to validate" is **objectively checkable** — a grep that returns zero, a route-truth
table that agrees, a 405 response with the contracted media type and body, a runtime stream upgrade, a
display name returned with no deadlock. No "works" / "looks right".

## Milestone validation definition

The close gate is run by the **`milestone-validator` subagent** (separation of powers — it judges and
writes `qa/milestone-qa.md`; the main session flips status only on its PASS), following the binding
C1–C7 checklist in `.claude/skills/milestone/references/milestone-end-validation.md`. What the gate
enforces for M1:

1. **Per-feature acceptance** — every feature meets its declared "what to validate", and each feature's
   **consumer contract** (`spec.md`) was honored (producer matches consumer). For F1.2 the consumer is
   the FE presence-stream client and the generated type; the contract shape comes from what the FE/runtime
   already consume — **read, never guessed**.
2. **Workflow-class QA** — `wiki/quality/backend-api-qa-checklist.md` for F1.1/F1.2/F1.3; the
   contract-first route-truth-table flow for F1.2.
3. **Focused audit slice** (non-terminal) on the **contract** + **error-contract** facets: **0** bare-`405`
   across swept packages, `/iam/presence/stream` tri-source-consistent — confirmed **fixed at root cause,
   not symptom-patched** (e.g. not a single-route patch leaving the class alive elsewhere).
4. **Regression** — M0 still passes its gate (docs unbroken); no swept handler's success path regressed;
   no unrelated route's contract moved in the F1.2 regen.
5. **No unplanned scope** — specifically, F1.3 stays **contained** (no shared-port generalization leaked in
   from M4); anything beyond this list is recorded with rationale or is an HS-6 stop.

## Dependencies & constraints

- **Depends on:** M0 passed (clean docs / single roadmap / reconciled ledger).
- **Runnable prerequisite:** M1 is the first code milestone — a fresh, runnable API on `:8081` via
  `.\scripts\start-api.ps1 -Build` is a prerequisite-truth boundary. If build / runnable / auth-session /
  route / contract truth fails, that is **HS-3** (switch to `runtime-contract-prereq`, repair, rerun the
  checkpoint, return to the feature) — not something to patch around.
- **Contract-first** (F1.2): build the route-truth-table first → compare runtime / spec / codegen / wiki →
  then regen in canonical order. No editing generated wiring or OpenAPI shape from memory.
- **H-PRE-1** (F1.3): never call an authz-recording read on a fresh connection inside the audit-lock atomic
  tx; fix the hazard by hoisting the read **off-tx**, never by adding a Tx-variant inside the lock.
- **Contained-before-generalized** (D4 / Approach-3): F1.3 is the contained step; the generalized
  `UserDisplayNameReader` port is **M4/F4.1**. Reads stay live; no migrations; no snapshot semantics.
- **No-merge:** the operator merges; the agent never does.

## Applicable hard-stops

| HS | What would trip it here |
|----|-------------------------|
| HS-1 | This milestone's close boundary — operator review gate; no next milestone and no merge without approval. |
| HS-2 | A fix turns out to require redesign outside its boundary — e.g. the canonical 405 helper itself needs a shared-API redesign, `/iam/presence/stream` needs a cross-module auth/authz-model change, or F1.3's contained read can't be contained without a storage/provider redesign. Stop; report the boundary + minimum prerequisite plan; do not symptom-patch. |
| HS-3 | A prerequisite boundary fails: build, runnable API on `:8081`, dev login/session, target route, or contract truth. Repair via `runtime-contract-prereq`, rerun the failed checkpoint, then resume the feature. |
| HS-4 | The focused audit slice finds a symptom-patch (a bare-`405` still alive in a swept package, or presence/stream still not tri-source-consistent) or the contract facet not trending A−. Stop; replan the offending feature; re-run its close-out. |
| HS-6 | Scope drift — most likely **F1.3 generalizing** into the M4 shared port, or the F1.2 regen quietly changing another route's contract. Stop; surface the deviation; replan before continuing. |
