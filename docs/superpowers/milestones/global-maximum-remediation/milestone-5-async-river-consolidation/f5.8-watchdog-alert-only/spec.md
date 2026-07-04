# F5.8 — watchdog alert-only collapse (spec)

> **Milestone:** M5 · **Status:** in progress · **Type:** production fix + ADR (durable decision).
> **Origin:** F5.7 proof run surfaced a failing watchdog equivalence test; root-cause audit
> (2026-07-04, three read-only investigators) proved it is **not** a fixture bug but a genuine,
> pre-existing production defect — an orphaned dead-code branch from an abandoned concept. Operator
> directed the global-maximum fix: **collapse the watchdog to alert-only and remove the orphans.**
> **Rails:** ADR 0068 (this feature's decision record).

## Root cause (audited, not assumed)

The `on_eligibility_drift` column carried two orthogonal concepts under one name. Concept A
(eligibility-drift quorum policy `{reduce_quorum, fail_stage, keep_snapshot}`) won — it is the only
CHECK-allowed value set and the only thing signoff evaluates. Concept B (a stuck-timeout action
`{auto_cancel, alert_only, none}`) was abandoned before ever reaching schema. The frontend was
corrected to Concept A (`14e48071`); the Go watchdog was never synced. So the watchdog reads
`on_eligibility_drift_snapshot` and compares it to `"auto_cancel"` — a value the column can never
hold. Full detail + evidence anchors in ADR 0068. **Not introduced by M5** (origin phase-8
`84b0507f`); M5 F5.2 migrated the read source but preserved the orphan.

## Consumer contract (who consumes the output, and the shape required)

- **Consumer 1 — the `metaldocs-jobs` runtime (watchdog wiring).** Requires `NewWorker` to no longer
  need a `CancelService`: after the dead branch is removed, the watchdog's only side effect is
  emitting governance alerts. The composition-root call site that constructs the watchdog must be
  updated to the new signature. Watchdog behavior: every `approval_instances` row `in_progress` with
  `submitted_at < now() - 7d` emits **exactly one** `approval.instance.stuck_alert` — no cancel path.
- **Consumer 2 — the milestone-validator (honest proof).** Requires the F5.2 watchdog equivalence
  proof to assert **real** behavior on a real Postgres: a stuck instance seeded with any CHECK-valid
  `on_eligibility_drift_snapshot` stays `in_progress`, its document stays `under_review`, and exactly
  one `stuck_alert` governance event is written. No test may seed `"auto_cancel"` (a CHECK violation)
  or assert a cancellation.
- **Consumer 3 — the approval module's public surface.** Requires the user-facing
  `CancelService.CancelInstance` (HTTP `cancel_handler.go` → `document.edit`-gated, ADR 0022 P10) to
  remain **byte-for-byte behaviorally unchanged**. Only the unreachable `SystemCancelInstance` +
  `system` bypass path is removed.

## What to implement

1. **Watchdog body** (`internal/modules/jobs/stuck_instance_watchdog/job.go`):
   - Remove the `if inst.DriftPolicy == "auto_cancel" { … SystemCancelInstance … }` block (lines
     100–116) and the `autoCancelled` counter + its log field.
   - Remove `cancelSvc` end to end: the `cancelSvcInterface` type, the `cancelSvc` + `runner` struct
     fields (both become unused — `runner` was used only to pass to `SystemCancelInstance`), the
     `NewWorker` `cancelSvc` parameter, and `run`'s `cancelSvc`/`runner` params. `listStuckInstances`
     and `emitStuckAlert` use raw `db.BeginTx`, not `runner`.
   - Keep `authz.WithBackgroundBypass(ctx)` in `Work` (still needed for `listStuckInstances` /
     `emitStuckAlert`'s own `authz.BypassSystem` reads) — update its comment (no longer references
     `SystemCancelInstance`). Keep `StuckInstance.DriftPolicy` (still emitted in the alert payload as
     informational context) and the `on_eligibility_drift_snapshot` read.
   - Update the doc comments on the worker/`run` that reference auto-cancel semantics.
2. **CancelService** (`internal/modules/documents/approval/application/cancel_service.go`):
   - Remove `SystemCancelInstance` (lines 50–54).
   - Fold `cancelInstance(…, system bool)` into `CancelInstance`: drop the `system` param, delete the
     `if system { BypassSystem + SeedTxIdentity }` block (66–80), and make the `authz.Require`
     (114–118) unconditional. Confirm `authz` import stays used (`WithCapCache`, `Require`); remove
     now-unused `authz.BypassSystem`/`SeedTxIdentity` calls only.
   - Everything else (GUC set, instance/stage cancel, doc→draft OCC, governance event) is unchanged —
     it is the shared body the live path already uses.
3. **Watchdog tests** (`job_test.go` + `job_integration_test.go`):
   - `job_test.go`: remove the `mockCancelService` (or its `SystemCancelInstance` method) and the
     auto-cancel unit cases; update the `NewWorker` call sites to the new signature. If the
     read-source invariant test (watchdog reads `asi.on_eligibility_drift_snapshot`) still holds
     (the alert payload still selects it), keep it; otherwise adjust its rationale — do **not** delete
     a genuine guard without cause.
   - `job_integration_test.go`: replace `TestIntegration_Watchdog_P1_AutoCancelEquivalence` with an
     honest alert-only proof (stuck instance, valid drift value → `in_progress` + `under_review` +
     one `stuck_alert`). `TestIntegration_Watchdog_P1_AlertOnlyEquivalence` already asserts this shape
     — merge/rename so there is one clear alert-only proof, no fiction. Drop now-unused helpers
     (`seedActiveStageSnapshot`'s `auto_cancel` usage, cancelled/draft asserts) only if truly unused.
4. **Wiki** (`wiki/modules/approval.md`): correct the watchdog description (~lines 77–79, 207–208)
   from "auto-cancels or emits governance alert" to alert-only; cite ADR 0068.
5. **ADR 0068** — already written (`wiki/decisions/0068-stuck-instance-watchdog-alert-only.md`),
   indexed. No further ADR work beyond confirming the cite from `approval.md`.

## Non-goals

- **No new `on_timeout_action` config.** Building a real auto-cancel feature is explicitly future
  work (ADR 0068 §Future) — a successor ADR + new column + snapshot + openapi + frontend. Not here.
- **No change to the user-facing `CancelInstance` path**, its HTTP route, its authz gate, or its
  governance event. No OpenAPI edit, no capability-registry change, no migration.
- **No change to the 7-day threshold** or batch size. No metrics/Prometheus.
- **No touching eligibility-drift semantics** (`domain/drift.go`, `decision_service.go`) — that
  concept is correct and unrelated.

## Validation Gate (acceptance — all must hold)

1. **Build:** `go build ./...` green; the `metaldocs-jobs` composition root compiles with the new
   `NewWorker` signature.
2. **Vet:** `go vet ./...` clean — no unused field/param/import left by the removal.
3. **Grep census = 0:** no `auto_cancel` / `SystemCancelInstance` / `stuck_watchdog_auto_cancel`
   references remain in `internal/**` non-test **or** test code (record the grep in evidence).
4. **Honest alert-only integration proof (testdb, targeted `-run`, real Postgres):** a stuck instance
   with a CHECK-valid `on_eligibility_drift_snapshot` → instance stays `in_progress`, document stays
   `under_review`, exactly one `approval.instance.stuck_alert`. A fresh (<7d) instance is untouched.
   This proof must **pass on real Postgres** (no schema-bypassing mock seeding an impossible value).
5. **User-cancel path regression:** `go test -tags=integration -run Cancel
   ./internal/modules/documents/approval/...` green — `CancelInstance` unchanged.
6. **Bypass-surface check:** `authz.BypassSystem` caller count in the approval module drops by one
   (the removed system-cancel path); the watchdog still roots in `WithBackgroundBypass` for its
   list/alert reads. Record before/after.
7. ADR 0068 Accepted + indexed; `wiki/modules/approval.md` reflects alert-only.

## Interview record

The operator interview was the root-cause AskUserQuestion (2026-07-04): after the three-investigator
audit established the semantic-collision root cause, the operator was presented three global-maximum
options and selected **"Collapse to alert-only, remove orphans"** over "build stuck-timeout action as
a real feature" and "investigate further". That selection IS the locked contract this spec distills.

| Question the audit forced | Answer (evidence) |
|---|---|
| Fixture bug or production defect? | Production defect — dead branch, unreachable literal (audit A/B/C; ADR 0068 Context). |
| Does a timeout-action concept exist to wire correctly? | No — no schema/domain/config/wiki concept anywhere (audit C verdict). |
| Was auto-cancel ever real product behavior? | No — never schema-backed, never fired in production, abandoned pre-schema (audit B git archaeology). |
| Safer default for a regulated doc system? | Human-in-the-loop alert, not silent janitor-cancel (ADR 0068 §Decision 4). |
| Global maximum? | Collapse to alert-only + remove orphans (operator-selected). |

## ADR

**ADR 0068** — `wiki/decisions/0068-stuck-instance-watchdog-alert-only.md` (Accepted 2026-07-04).
