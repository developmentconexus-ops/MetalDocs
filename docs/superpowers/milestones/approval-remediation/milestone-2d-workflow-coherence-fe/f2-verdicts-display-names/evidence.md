# Feature F2d.2 — Evidence

> **Milestone:** 2d  ·  **Feature:** `f2-verdicts-display-names`  ·  **Closed:** 2026-07-09
> **Contract:** `spec.md` (consumer contract + Validation Gate this proves against).

## What was implemented

By outcome (producer matches the consumer contract in `spec.md` — the by-document DTO gains a required
`verdicts[]`; the code was built to that shape, not the reverse):

- **D1 — DB-enforced snapshot invariant (root cause).** `db/migrations/0294_actor_display_name_snapshot_not_null.sql`:
  idempotent, forward-only. Backfills any `NULL`/`''` `actor_display_name_snapshot` from
  `metaldocs.iam_users.display_name` (joined on user_id+tenant_id) on **both**
  `approval_review_verdicts` and `approval_signoffs`, then `SET NOT NULL` + `ADD CONSTRAINT
  ..._display_name_nonempty CHECK (actor_display_name_snapshot <> '')`; ledger insert `'0294'`
  ON CONFLICT DO NOTHING. Insert path binds the column **unconditionally** in `InsertVerdict` +
  `InsertSignoff` (removed the `if name != ""` omission). Decision logic unchanged.
- **D2 — read is a pure projection.** Removed `coalesce(v.actor_display_name_snapshot,'')` in
  `LoadStageVerdicts`, the two signoff loaders, and `loadVerdictByStageActor` (scan a plain non-null
  string, guaranteed by D1). Collapsed the signoff read's snapshot-else-live branch in
  `get_instance_handler.go` (`mapInstanceResponse` records loop + `buildStageActors` signed branch) to
  **snapshot-only**. No `displayNameReader` fallback survives on any snapshotted-action read path.
- **By-instance read** `LoadInstanceVerdicts(ctx, tx, tenantID, instanceID)` — mirrors
  `LoadStageVerdicts` but `WHERE v.approval_instance_id = $1`, `ORDER BY v.verdict_at ASC`,
  tenant-scoped (`ai.tenant_id = $2 AND v.actor_tenant_id = ai.tenant_id`); shared `scanVerdicts` helper.
- **App read** `LoadInstanceByDocumentForViewWithViewer` → 4-value return `(*Instance, ViewerFacts,
  []ReviewVerdict, error)`; loads verdicts in the SAME view tx (plain SELECT, no lock — H-PRE-1 trivially
  satisfied, snapshot names inline so no off-tx join).
- **Handler** projects `[]domain.ReviewVerdict` → `[]contracts.VerdictRecord` in `mapInstanceResponse`,
  gated on `viewer != nil` (by-document only; by-id passes `nil, nil`). `Verdicts *[]VerdictRecord`
  pointer → nil omitted (by-id), non-nil empty slice serializes `[]` (by-document empty). `reason` is
  `*string` nil when comment empty; `verdict_at` RFC3339 UTC.
- **OpenAPI + regen.** `ApprovalInstanceByDocumentResponse` gains required `verdicts`;
  new `ApprovalVerdictRecordResponse` schema (verdict enum [ready, request_changes], reason nullable).
  `oapi-codegen` Go (`api.gen.go`) + TS (`index.d.ts`) regenerated — no hand-written consumer.
- **ADR 0079** written (verdict-history contract + immutable actor-name, no read fallback);
  `wiki/decisions/index.md` rows added (0078 was also missing); `milestone.md` HS-2 amendment recorded.

Not yet committed.

## Verification

| Check | Command / action | Result (evidence) | Real vs fixture |
|-------|------------------|-------------------|-----------------|
| TDD — failing tests first, then green | 6 real-DB tests in `read_service_instance_verdicts_integration_test.go` written before impl (Step 0), red pre-migration | `ok metaldocs/internal/modules/documents/approval/application 342.463s` — 6/6 PASS | real |
| Static (build/vet/test-compile) | `go build ./internal/modules/documents/approval/...`; `go vet -tags integration ./...application/` | BUILD_EXIT=0; VET_EXIT=0 | — |
| Targeted real-DB test | `go test -tags integration -run TestInstanceVerdicts_* ...` | Ready 223.80s, RequestChanges 64.82s, MultiStageOrdered 19.68s, None 4.28s — PASS | real |
| D1 fail-closed invariant | `TestInsertVerdict_EmptySnapshot_Rejected` (5.68s), `TestInsertSignoff_EmptySnapshot_Rejected` (17.54s) | empty snapshot → DB CHECK/NOT NULL rejection, both PASS | real |
| D2 grep gate — no snapshot coalesce | `grep coalesce(*.actor_display_name_snapshot infrastructure/` | NONE | real |
| D2 grep gate — no hand-written consumer | `grep body.data.verdicts frontend/apps/web/src` | NONE | real |
| Regression — D1 blast radius (F2d.1 signoff seed) | `go test -tags integration -run TestViewerBlock_AlreadySigned ...` after fixing `seedSignoff` to set snapshot | `ok ... 7.414s` PASS | real |

> Regression note: D1's NOT NULL broke exactly one pre-existing real-DB seed (`seedSignoff` in the F2d.1
> viewer integration test, which omitted the snapshot). Fixed the seed (added `ActorDisplayNameSnapshot`)
> — a required harness consequence of the invariant, not a scope change. All other `NewSignoff`/`NewVerdict`
> sites are fakes (phase5, decision_service) or domain/HTTP unit tests (no real DB), unaffected.

## Acceptance vs spec Validation Gate

| Acceptance criterion (from spec.md) | Met? | Evidence |
|-------------------------------------|------|----------|
| verdict history shape after `ready` (single-stage) | yes | `TestInstanceVerdicts_Ready` |
| verdict history after `request_changes` (reason present) | yes | `TestInstanceVerdicts_RequestChanges` |
| verdicts span multiple stages, chronological asc | yes | `TestInstanceVerdicts_MultiStageOrdered` |
| display names resolve for seeded users (snapshot, as-cast) | yes | asserted in the three above (snapshot = seeded name) |
| no in-tx display lookup; no read-side fallback | yes | `LoadInstanceVerdicts` reads snapshot inline; review 0 findings; D2 grep gate NONE |
| empty instance ⇒ `verdicts: []` (present, not omitted) | yes | `TestInstanceVerdicts_None` + pointer-to-empty-slice serialization |
| D1 — NOT NULL + CHECK(<>''); insert binds unconditionally | yes | migration 0294 + `TestInsert{Verdict,Signoff}_EmptySnapshot_Rejected` |
| D1 — backfill leaves zero null/empty; SET NOT NULL succeeds | yes | migration runs clean in testdb bootstrap (all 6 tests apply it) |
| D2 — no fallback on snapshotted-action reads | yes | grep gate NONE + review confirmed |
| D2 — signoff name still projects (snapshot-only) | yes | `TestViewerBlock_AlreadySigned` green |
| regen: no hand-written DTO consumer; build clean | yes | oapi regen + build EXIT=0 + grep gate NONE |

## Review disposition

- Spec-compliance review: **PASS** — producer matches the consumer contract (required `verdicts[]` on the
  by-document DTO; by-id omits; enum/reason/verdict_at shapes match spec §Consumer contract).
- Code-quality review (independent `caveman:cavecrew-reviewer`, sonnet, 7 files): **APPROVE — 0 findings**
  (`totals: 0🔴 0🟡 0❓`). Confirmed unconditional binds, no coalesce/live fallback, correct pointer
  semantics, `reason` nil-for-empty, `verdict_at` ASC, tenant scoping, migration idempotency + both-table
  constraints + ledger.

## Bounded defers

| Defer | Why bounded | Written trigger / owner |
|-------|-------------|-------------------------|
| **D3** — surface `on_behalf_of` delegator ("X on behalf of Y") in `verdicts[]` | Delegator name is not snapshotted; as-cast display needs a NEW snapshot column (separate additive migration + write-path capture) — distinct from D1's constraint hardening | Trigger: when F2d.5 timeline design calls for delegation attribution. Owner: approval module + F2d.5. |
