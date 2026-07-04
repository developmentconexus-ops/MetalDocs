# F5.5 — fanout ordering guarantee (plan)

> Executes `spec.md` / contract §5. Single-task feature (proof-only) — engine:
> `superpowers:subagent-driven-development` (one fresh subagent, sonnet, main session reviews+commits).

## Load-bearing facts (verified 2026-07-04)

1. **Worker under test:** `internal/modules/notifications/infrastructure/fanout_worker.go` —
   `NotificationsFanoutWorker.Work(ctx, job *river.Job[documentsdomain.LifecycleEventArgs])`. Each call
   is single-event, single-tenant, its own seeded tx (`authz.SeedTxTenant`).
2. **Dedup key:** `metaldocs.notifications (recipient_user_id, source_event_id)` partial unique index
   (`WHERE source_event_id IS NOT NULL`), `source_event_id = args.EventID` (minted per lifecycle event
   at emit time, `documents/domain/notification_events.go:15`).
3. **Existing test file to extend or sibling:** `internal/modules/notifications/infrastructure/fanout_worker_integration_test.go`
   (`TestNotificationsFanoutWorker`, `//go:build integration`) — read its testdb setup (obligated-reader
   view seeding, tenant/document/user fixtures) and reuse the same fixture helpers rather than
   reinventing document/reader seeding.
4. **Obligated readers source:** `fanoutToReaders` queries `metaldocs.v_cd_obligated_readers` — the
   race test needs at least 2 distinct obligated readers on one controlled document so the "every
   recipient has both rows" assertion is non-trivial (a single-reader race is a weaker proof).
5. **`published` vs `superseded` both fan out via `fanoutToReaders`** (same code path, different
   `EventType`/`EventID`) — the two racing `Work()` calls differ only in `args.EventType` +
   `args.EventID` (+ whatever the emit path would set — check `documentsdomain.LifecycleEventArgs`'s
   fields to construct valid Args for both event kinds directly, without needing to drive it through
   the full emit pipeline).

## Task breakdown

### T1 (sonnet) — concurrent race integration test
- New test in the existing `fanout_worker_integration_test.go` (or a new
  `fanout_worker_race_integration_test.go` sibling file, same package, same build tag — your call,
  document the choice) proving contract §5.2's 3 assertions:
  - Seed one tenant, one controlled document with ≥2 obligated readers (reuse existing fixture
    helpers from the current integration test file).
  - Construct two `documentsdomain.LifecycleEventArgs` for the SAME document/tenant: one
    `EventTypeDocumentPublished` (`EventID` A), one `EventTypeDocumentSuperseded` (`EventID` B).
  - Run BOTH interleavings (e.g. loop `i := 0; i < 2; i++` swapping which goroutine's tx a barrier
    releases first, or simply run both orderings as two separate subtests with a start-gate
    channel/`sync.WaitGroup` forcing genuine concurrent `BeginTx` before either commits) using 2
    REAL goroutines each calling `worker.Work(ctx, ...)` on the SAME testdb, not simulated
    sequentially.
  - After both complete (join via `errgroup`/`sync.WaitGroup` + collected errors), assert: total
    notification-row count = readers × 2 (one row per reader per event); every reader has exactly
    one row keyed `source_event_id=A` and one keyed `source_event_id=B`; row set is IDENTICAL between
    the two interleaving runs (compare sorted (recipient, source_event_id, event_type) tuples).
  - Re-run one of the two `Work()` calls again post-race (redelivery) → assert row count unchanged
    (idempotency, `ON CONFLICT ... DO NOTHING` holds).
- `go build -tags=integration ./...` clean, `go vet -tags=integration ./...` clean.
- Attempt to run (`go test -tags=integration -run <name> ./internal/modules/notifications/infrastructure/... -v -race`)
  — use `-race` too since this test is specifically about concurrency; expect a `SKIP` (no DB DSN),
  same precedent as F5.2/F5.3/F5.4. If it fails for any other reason, fix it.
- If (and only if) the test reveals a shared mutable projection breaking commutativity: STOP, do not
  patch inside this task — report BLOCKED with the concrete failure, that is an HS-6 scope surface for
  the main session/operator to triage, not a silent fix.
- Commit.

## Test strategy
This feature IS the test — no separate TDD red/green cycle beyond writing the race test itself
(the production code is not expected to change). Targeted `-run`, not the full suite.

## Files touched (census)
NEW test file (or extension of the existing integration test file) under
`internal/modules/notifications/infrastructure/`. No production code changes anticipated.
