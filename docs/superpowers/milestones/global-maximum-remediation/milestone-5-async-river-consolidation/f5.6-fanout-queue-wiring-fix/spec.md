# F5.6 — fanout queue-wiring fix (spec)

> **Milestone:** M5 · **Status:** in progress
> **Origin:** M5 close-gate live QA drive (task #7) — HS-6 scope-surface discovery, operator-directed
> fix-in-place (2026-07-04). Not a new feature request; a defect found by the milestone's own
> required live drive (`milestone.md` validation-definition item 6).
> **Rails:** the established M5 enqueue convention — every periodic/request-triggered River job this
> milestone touches sets an explicit `Queue` in its `InsertOpts` matching a queue `metaldocs-jobs`
> actually subscribes to (`"temporal"` for request-triggered jobs — `scheduled_publish_job.go:88`,
> `dispatchjobs/enqueuer.go:96`; `"maintenance"` for periodic jobs — `maintenance/periodic.go`,
> `retention/periodic.go:30`). `RiverLifecycleEventEnqueuer` is the one enqueue site that predates
> this convention and passes `nil` `InsertOpts` — landing on River's implicit `"default"` queue,
> which `metaldocs-jobs` never subscribes to (`internal/platform/config/jobs.go:19-40`,
> `apps/jobs/cmd/metaldocs-jobs/main.go:45-51`).

## Consumer contract

- **Consumer 1 — `NotificationsFanoutWorker`.** Already correctly registered
  (`river.AddWorker(workers, notificationsinfra.NewNotificationsFanoutWorker(db))`,
  `apps/jobs/cmd/metaldocs-jobs/main.go:61`) and already proven correct in isolation (F5.5's race
  test). It has never run in the live system because the jobs it needs are never dequeued — the
  defect is purely queue-name mismatch at the enqueue site, not the worker or its SQL/view logic.
- **Consumer 2 — the milestone's own live QA drive (`milestone.md` validation-definition item 6).**
  Requires a live-driven proof that a `document.published` (or `superseded`) event actually produces
  a materialized notification row, end to end, on the running system.
- **Consumer 3 — the milestone-validator.** Requires this fixed and evidenced before M5 can PASS its
  live-drive criterion.

## What to implement

Set `Queue: "temporal"` explicitly on the `RiverLifecycleEventEnqueuer.EnqueueLifecycleEventTx`
insert (`internal/modules/documents/approval/jobs/lifecycle_event_enqueuer.go:33`), matching the
convention used by every other request-triggered M5 job (`scheduled_publish_job.go`,
`dispatchjobs/enqueuer.go`). `"temporal"` is the correct queue (not `"maintenance"`) because
lifecycle-event fanout is triggered by a user/business action (publish/supersede/obsolete/
approve/reject), the same class of request-triggered async work as scheduled-publish and staging
dispatch — not a periodic maintenance sweep.

No change to `NotificationsFanoutWorker`, `v_cd_obligated_readers`, the notifications schema, or
`metaldocs-jobs`'s queue subscription map (`"temporal"` is already subscribed with 10 workers,
`internal/platform/config/jobs.go:24` — no new queue subscription needed).

## Non-goals

- No change to fanout logic (already proven correct by F5.5).
- No new queue added to `JobsConfig.Queues` — `"temporal"` already exists and has capacity.
- No change to `metaldocs-api`'s enqueue-only config (it never subscribes any queue; unaffected).
- No retroactive re-processing of the two live-drive scratch documents' orphaned `"default"`-queue
  jobs from this session's investigation — those are inert (never processed, harmless dev-tenant
  rows in a local Postgres instance) and out of scope for a code fix.

## Validation Gate (acceptance — all must hold)

1. **Unit/compile proof:** `RiverLifecycleEventEnqueuer.EnqueueLifecycleEventTx` passes
   `&river.InsertOpts{Queue: "temporal"}` instead of `nil` — verified by reading the diff and by a
   focused unit test asserting the `InsertOpts.Queue` value passed to a fake/mock `InsertTx`
   (mirrors `dispatchjobs/enqueuer_test.go`'s existing pattern of asserting `gotOpts.Queue`).
2. **`go build ./...` green; all 4 binaries build.**
3. **Live re-drive:** repeat the notification-fanout live drive (direct `/publish` on a
   `visibility.scope=company` controlled document, `admin` as an obligated company-scope reader) on
   the rebuilt (`.\scripts\start-api.ps1 -Build`) system; `GET /api/v1/notifications` for the reader
   shows a materialized row for the driven event within a bounded poll window (River's `"temporal"`
   queue has active workers, so this should be near-immediate, not policy-interval-bound).
4. **No regression:** F5.2/F5.3/F5.4/F5.5's existing tests still compile/vet clean;
   `scheduled_publish_job.go`'s and `dispatchjobs`'s existing `"temporal"`-queue behavior (Step A of
   the live drive) is unaffected — same queue, just one more producer.
5. **Section-by-section match to this spec.** No HS-7 (this spec is the first documented contract
   for this exact defect — no prior contract to diverge from).

## Interview record

No operator interview needed beyond the HS-6 AskUserQuestion already answered (2026-07-04:
"Open bounded fix feature F5.6 now" — recommended and selected). The fix itself is fully determined
by the existing M5 convention (three sibling enqueue sites already do this correctly) — no design
freedom.

## ADR

No new ADR — this is a wiring-bug fix consistent with ADR 0067's already-accepted decision (River
is the single async primitive with per-queue subscription discipline); it does not change any
accepted architectural decision.
