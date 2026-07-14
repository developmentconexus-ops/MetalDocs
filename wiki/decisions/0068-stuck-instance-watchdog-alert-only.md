# ADR 0068 — Stuck-instance watchdog is alert-only; the auto-cancel timeout-action concept is withdrawn

- **Status:** Accepted 2026-07-04
- **Module(s):** `jobs` (stuck-instance-watchdog) · `documents/approval` (CancelService)
- **REQ IDs:** `backend-target-architecture.md` REQ-ASYNC-* (janitors), REQ-AUTHZ-* (bypass surface)
- **Supersedes / amends:** corrects the original phase-8 watchdog design (`84b0507f`, Task 8.3);
  makes stale ADR 0067 §H-PRE-1's incidental reference to `SystemCancelInstance` as a
  background-bypass caller.

## Context

An M5-close proof run (F5.7) surfaced a failing watchdog equivalence test. Root-cause audit
(2026-07-04, three read-only investigators) established the failure is **not** a fixture bug but the
last surviving fragment of a **semantic collision**, resolved everywhere except the watchdog:

- The column `approval_route_stages.on_eligibility_drift` (snapshotted to
  `approval_stage_instances.on_eligibility_drift_snapshot`) was born (phase 8) carrying **two
  orthogonal concepts under one name**:
  - **A — eligibility-drift quorum policy** `{reduce_quorum, fail_stage, keep_snapshot}`: what to do
    when an approver *loses eligibility mid-approval*; evaluated at signoff (`domain/drift.go`).
  - **B — stuck-timeout action** `{auto_cancel, alert_only, none}`: what to do when an instance
    *sits past the 7-day threshold*; read by the watchdog.
- The database CHECK constraint has **always** allowed only Concept A (no `auto_cancel` in any
  migration, baseline, or archive). The frontend enum was corrected to Concept A in `14e48071`
  ("align frontend DriftPolicy to backend"). The Go watchdog was **never synced** — it still reads
  `on_eligibility_drift_snapshot` and compares it to `"auto_cancel"` (`job.go:100`).
- **Consequences, all fictions:**
  1. The `auto_cancel` branch is **unreachable** — the column can never hold that value.
  2. `CancelService.SystemCancelInstance` (authz-bypass system cancel) is reachable **only** via that
     dead branch → a fully **orphaned** capability in production.
  3. `TestIntegration_Watchdog_P1_AutoCancelEquivalence` and `wiki/modules/approval.md` assert/document
     an auto-cancel that **has never once fired in production**; the test passed only via mocks that
     bypass the schema CHECK.
- **No timeout-action configuration concept exists anywhere** — schema, domain model, route/stage
  config, or wiki intent. The 7-day threshold is a hard-coded constant. Auto-cancel was never a
  specced product feature; it is a pre-schema design fragment. M5 F5.2 migrated the watchdog read
  source (route → snapshot) but preserved the orphan literal; **it did not introduce the defect.**

Actual production behavior today: every stuck instance emits `approval.instance.stuck_alert`; a human
decides. The watchdog is *de facto* alert-only.

## Decision

1. **The stuck-instance watchdog is alert-only.** Every instance `in_progress` past the 7-day
   threshold emits exactly one `approval.instance.stuck_alert` governance event. There is **no**
   automated cancellation path.
2. **Withdraw the `auto_cancel` timeout-action concept.** It was never schema-backed and collided
   with the eligibility-drift semantics of `on_eligibility_drift`. Remove the dead branch.
3. **Remove `SystemCancelInstance` and the `system` authz-bypass path** in
   `CancelService.cancelInstance` as an unreachable capability. This strictly **reduces the
   authz-bypass surface** (one fewer `authz.BypassSystem` caller). The user-facing
   `CancelInstance` — gated by `document.edit` (ADR 0022 P10) — is **unchanged**.
4. **Rationale (global maximum).** For a regulated controlled-documents system, a janitor silently
   cancelling an in-flight approval after a timer is a dangerous default: cancellation reverts the
   document to draft and discards approval progress. Stuck approvals warrant **human judgment**, not
   automatic reversal. Alert-only is the safer design *and* matches the behavior the system has
   actually exhibited since the schema/frontend correction. Collapsing to alert-only removes the
   patch entirely rather than optimizing inside it.

## Consequences

- F5.2's `TestIntegration_Watchdog_P1_AutoCancelEquivalence` (a fiction — asserted a cancel that never
  happens, green only under schema-bypassing mocks) is **replaced** by an honest alert-only proof:
  a stuck instance with any CHECK-valid drift policy stays `in_progress` / document stays
  `under_review`, and exactly one `stuck_alert` is emitted.
- `wiki/modules/approval.md` watchdog description ("auto-cancels or emits governance alert") is
  corrected to alert-only.
- ADR 0067 §H-PRE-1's incidental mention of `SystemCancelInstance`'s `authz.BypassSystem` becomes
  stale; the watchdog still roots its tick in `authz.WithBackgroundBypass` for its own
  `listStuckInstances` / `emitStuckAlert` bypass reads, so the background-root itself stays.
- No wire-contract change, no OpenAPI change, no capability-registry change, no migration. The
  removed system-cancel path had no route.

## Future work (explicitly out of scope)

If auto-cancel-on-timeout (or escalation/SLA actions) is ever wanted as a **product** feature, it
requires a **first-class `on_timeout_action` configuration**: a new route-stage column with its own
CHECK enum, snapshotted to the instance like drift, surfaced in OpenAPI + frontend, and read by the
watchdog from *that* field. It MUST NOT reuse `on_eligibility_drift`. That is a successor ADR and a
new feature, not a revival of this dead branch.
