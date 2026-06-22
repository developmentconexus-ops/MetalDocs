# Feature F3.3 — Evidence

> **Milestone:** 3 — Notifications (full-stack)  ·  **Feature:** `f3.3-lifecycle-emitter`  ·  **Closed:** 2026-06-22
> **Contract:** `spec.md` (consumer contract + Validation Gate this proves against).
> **ADR:** ADR-0044 (domain-event pattern; River dispatch).

## What was implemented

Seven tasks per `plan.md`, all delivered in one session:

**T1 — Domain contract (`documents/domain`)**
- **CREATED** `internal/modules/documents/domain/notification_events.go` — `LifecycleEventArgs` River-job envelope (8 fields); 5 `EventType*` constants (`document.published/superseded/obsoleted/approved/rejected`); `LifecycleEventEnqueuer` port interface (`EnqueueLifecycleEventTx(ctx, db.Tx, LifecycleEventArgs) error`). No `river` import — domain stays infra-free.
- **CREATED** `internal/modules/documents/domain/notification_events_test.go` — compile-time tests: `Kind()` returns `"notification_fanout"`, all 5 constants non-empty, interface exists.

**T2 — LoadDocumentControlledDocumentID helper (`documents/application`)**
- **CREATED** `internal/modules/documents/application/document_cdid.go` — `LoadDocumentControlledDocumentID(ctx, *sql.Tx, tenantID, documentID) (string, error)`. Returns `""` (not an error) for no-CD-link or `ErrNoRows`. Used by reader-event emit sites.

**T3 — RiverLifecycleEventEnqueuer adapter (`approval/jobs`)**
- **CREATED** `internal/modules/documents/approval/jobs/lifecycle_event_enqueuer.go` — `RiverLifecycleEventEnqueuer` + `NewLifecycleEventEnqueuer`. Mirrors `RiverScheduledPublishEnqueuer` pattern exactly. One localized `tx.(*sql.Tx)` assertion (required by River; `db.Tx` interface has no `*sql.Tx` guarantee). `InsertTx` with `nil` opts (default queue).
- **CREATED** `internal/modules/documents/approval/jobs/lifecycle_event_enqueuer_test.go` — interface-check compile test + wrong-tx-type error path test.

**T4 — Additive enqueue at 5 emit sites + Services.WithLifecycleEnqueuer**
- **MODIFIED** `internal/modules/documents/approval/application/publish_service.go` — `lifecycleEnqueuer` field + `WithLifecycleEnqueuer` + reader-event enqueue block after audit emit in `PublishApproved`. Calls `LoadDocumentControlledDocumentID` in-tx; skips enqueue silently if field nil.
- **MODIFIED** `internal/modules/documents/approval/application/supersede_service.go` — same pattern; reader event on `req.NewDocumentID`.
- **MODIFIED** `internal/modules/documents/approval/application/obsolete_service.go` — reader event on `req.DocumentID`.
- **MODIFIED** `internal/modules/documents/approval/application/decision_service.go` — author events (`document.approved` / `document.rejected`) gated on `result.InstanceApproved` / `result.InstanceRejected`; uses `instance.SubmittedBy`; `now` variable in scope from original line 297.
- **MODIFIED** `internal/modules/documents/approval/application/services.go` — `WithLifecycleEnqueuer(*Services)` wires all 4 services (Publish, Supersede, Obsolete, Decision) and returns `*Services`.
- **CREATED** `internal/modules/documents/approval/application/lifecycle_emit_test.go` — `spyLifecycleEnqueuer`; `TestWithLifecycleEnqueuer_Services` (all 4 fields wired, same pointer returned); `TestWithLifecycleEnqueuer_NilServices` (nil receiver safe).

**T5 — NotificationsFanoutWorker + integration tests**
- **CREATED** `internal/modules/notifications/infrastructure/fanout_worker.go` — `NotificationsFanoutWorker` (`river.Worker[LifecycleEventArgs]` + `river.WorkerDefaults`). `Work` dispatches: reader events (`published/superseded/obsoleted`) → `fanoutToReaders` (queries `metaldocs.v_cd_obligated_readers`; skips if `ControlledDocumentID=="""`); author events (`approved/rejected`) → `fanoutToAuthor` (inserts for `args.SubmittedBy`). `insertRow` uses `ON CONFLICT (recipient_user_id, source_event_id) WHERE source_event_id IS NOT NULL DO NOTHING` (partial unique index from migration 0247). pt-BR title/message map for all 5 event types.
- **CREATED** `internal/modules/notifications/infrastructure/fanout_worker_integration_test.go` — `//go:build integration`; 7 subtests: `published_to_obligated_readers`, `superseded_to_obligated_readers`, `obsoleted_to_obligated_readers`, `approved_to_submitter`, `rejected_to_submitter`, `redelivery_is_noop`, `no_cd_link_skips_readers`. Seeds via `testdb.NewControlledDoc` + direct INSERT into `public.controlled_document_user_grants`. Calls `worker.Work` directly (no River runtime needed).

**T6 — Wire binaries**
- **MODIFIED** `apps/api/cmd/metaldocs-api/main.go` — `approvalServices.WithLifecycleEnqueuer(approvaljobs.NewLifecycleEventEnqueuer(riverBundle.Client))` added immediately after `WithScheduledPublishEnqueuer` (same `riverBundle.Client`, same `deps.SQLDB != nil` guard, line ~494).
- **MODIFIED** `apps/jobs/cmd/metaldocs-jobs/main.go` — added `notificationsinfra` import; in worker factory lambda: `river.AddWorker(workers, notificationsinfra.NewNotificationsFanoutWorker(db))` after `approvaljobs.NewWorkers`.

## Verification

| Check | Command / action | Result | Real vs fixture |
|-------|------------------|--------|-----------------|
| `go build ./...` (full) | `go build ./...` | exit 0, no output | — |
| `go build -tags integration ./internal/modules/notifications/...` | integration tag build | exit 0, no output | — |
| `go test ./internal/modules/documents/domain/...` | unit suite | ok (cached) | — |
| `go test ./internal/modules/documents/approval/jobs/...` | unit suite | ok 0.953s | — |
| `go test ./internal/modules/documents/approval/application/...` | unit suite incl. spy wiring tests | ok (cached) | — |
| `go test ./internal/modules/documents/application/...` | unit suite | ok (cached) | — |
| Module boundary check | `.\scripts\check-module-boundaries.ps1` | Pre-existing violations only; **zero new violations** from F3.3 files | — |
| Integration tests — compile | `go build -tags integration ./internal/modules/notifications/...` | exit 0 | — |
| Integration tests — runtime | Deferred to CI / live-PG run (no live DB in this session) | Compile-verified | fixture (compile only) |

### Integration test subtests (compile-verified; runtime deferred)

| Subtest | Assert |
|---------|--------|
| `published_to_obligated_readers` | 2 users granted → 2 `metaldocs.notifications` rows for that `source_event_id` |
| `superseded_to_obligated_readers` | 1 granted reader → 1 row |
| `obsoleted_to_obligated_readers` | 1 granted reader → 1 row |
| `approved_to_submitter` | `submitted_by` → 1 row |
| `rejected_to_submitter` | `submitted_by` → 1 row |
| `redelivery_is_noop` | Two `Work` calls → still 1 row (ON CONFLICT DO NOTHING) |
| `no_cd_link_skips_readers` | `ControlledDocumentID=""` → 0 rows |

## Acceptance vs spec Validation Gate

| Acceptance criterion (from spec.md) | Met? | Evidence |
|-------------------------------------|------|----------|
| `LifecycleEventArgs.Kind() == "notification_fanout"` | yes | `TestKind` compile-time test; `go test ./documents/domain/...` PASS |
| 5 event-type constants defined + non-empty | yes | `TestEventTypeConstants` PASS |
| `LifecycleEventEnqueuer` port in `documents/domain` (no `river` import) | yes | `go build ./...` exit 0; `river` absent from domain package imports |
| `db.Tx` interface used (not `*sql.Tx`) for port; single assertion in adapter | yes | Port uses `db.Tx`; `RiverLifecycleEventEnqueuer` is the only assertion point |
| Enqueue is additive (after audit emit, before result assignment) | yes | Source inspection: all 4 service files; `emitter.Emit` → `lifecycleEnqueuer.EnqueueLifecycleEventTx` ordering verified |
| HS-2 constraint — publish/approval semantics unchanged | yes | Only additive enqueue added; no conditional or state-change path altered |
| H-PRE-1 — fan-out worker is off-tx (no authz read inside lock) | yes | `NotificationsFanoutWorker.Work` uses `*sql.DB` (pool), never a `*sql.Tx`; worker runs post-commit |
| Reader events query `v_cd_obligated_readers` (not CD base tables) | yes | `fanoutToReaders` SQL uses `metaldocs.v_cd_obligated_readers` |
| `ControlledDocumentID==""` → skip (no error) | yes | `fanoutToReaders` guards `if args.ControlledDocumentID == ""`; `no_cd_link_skips_readers` subtest |
| Idempotency — redelivery inserts 0 duplicate rows | yes | `ON CONFLICT … DO NOTHING` on partial unique index; `redelivery_is_noop` subtest |
| `Services.WithLifecycleEnqueuer` wires all 4 services | yes | `TestWithLifecycleEnqueuer_Services` PASS; `TestWithLifecycleEnqueuer_NilServices` PASS |
| API binary: enqueuer injected via `riverBundle.Client` | yes | `apps/api/cmd/metaldocs-api/main.go` line ~494; `go build ./...` exit 0 |
| Jobs binary: `NotificationsFanoutWorker` registered | yes | `apps/jobs/cmd/metaldocs-jobs/main.go`; `river.AddWorker` call; `go build ./...` exit 0 |
| Module boundary: `notifications/infrastructure` → `documents/domain` only (no /application cross-module) | yes | Module boundary check: zero new violations from F3.3 files |
| `go build ./...` green | yes | exit 0 |

## Review disposition

- **Spec-compliance:** All 5 domain events implemented exactly as specced. Port in `documents/domain` (infra-free, 47× precedent). `db.Tx` interface preserved at port; `*sql.Tx` assertion localized to adapter. HS-2 additive-only constraint met. H-PRE-1 met (worker is off-tx). Idempotency via pre-existing partial unique index. Module boundary clean. Non-goals hold: no FE wire, no delivery channel, no SMTP.
- **Code-quality:** `go build ./...` + `go vet` clean. Pattern mirrors `RiverScheduledPublishEnqueuer` exactly (no novel patterns). `ptBRMessages` map is the only business-logic table; all 5 entries populated. No TODOs or stubs in delivered files.

## Bounded defers

| Defer | Why bounded | Written trigger / owner |
|-------|-------------|-------------------------|
| Integration test runtime pass (live PG) | No live DB in this session; compile pass + subtest structure verified | CI on first push with `METALDOCS_TEST_DB_URL` set |
| F3.4 FE notifications screen wire | Intentionally not in F3.3 (non-goal) | F3.4 execution |
