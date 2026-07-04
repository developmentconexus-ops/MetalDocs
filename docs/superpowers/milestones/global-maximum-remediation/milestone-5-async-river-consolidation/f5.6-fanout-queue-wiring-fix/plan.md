# F5.6 — fanout queue-wiring fix (plan)

> Executes `spec.md`. Engine: `superpowers:subagent-driven-development` (single task, sonnet
> implement+review; main session reviews + commits + performs the live re-drive itself since it
> requires the running dev system, not something a subagent should own restarting).

## Load-bearing facts (verified 2026-07-04)

1. **The bug:** `internal/modules/documents/approval/jobs/lifecycle_event_enqueuer.go:33` —
   `e.Client.InsertTx(ctx, sqlTx, args, nil)` — `nil` `InsertOpts` means River defaults the job to
   queue `"default"` (`vendor/github.com/riverqueue/river/client.go:1635`). `metaldocs-jobs` only
   subscribes `"temporal"` (`internal/platform/config/jobs.go:24`) and `"maintenance"`
   (`apps/jobs/cmd/metaldocs-jobs/main.go:51`) — never `"default"`. Jobs land in River's job table
   and sit there forever, unconsumed.
2. **The fix:** change `nil` → `&river.InsertOpts{Queue: "temporal"}`, matching
   `scheduled_publish_job.go:87-89` and `dispatchjobs/enqueuer.go:96` exactly.
3. **Test pattern to mirror:** `internal/modules/render/fanout/dispatchjobs/enqueuer_test.go` — a
   `fakeRiverInserter` capturing `gotOpts` and asserting `gotOpts.Queue == "temporal"`. Apply the
   same pattern to a (new or existing) `lifecycle_event_enqueuer_test.go`.
4. **No queue-subscription change needed** — `"temporal"` already has `MaxWorkers: 10`
   (`internal/platform/config/jobs.go:24`), plenty of headroom; this just adds one more producer
   type to an already-subscribed queue.
5. **Live re-drive is the main session's job, not the subagent's** — it requires the already-running
   dev system to be rebuilt (`.\scripts\start-api.ps1 -Build`) and re-driven via HTTP, which is
   simplest to do directly rather than dispatching a second live-drive subagent for one narrow
   re-check.

## Task breakdown

### T1 (sonnet) — fix + unit test
- `internal/modules/documents/approval/jobs/lifecycle_event_enqueuer.go`: change the `InsertTx` call
  to pass `&river.InsertOpts{Queue: "temporal"}` instead of `nil`.
- Add/extend a unit test (new `lifecycle_event_enqueuer_test.go` if none exists — check first, this
  package may already have one) asserting the enqueuer passes `Queue: "temporal"` to `InsertTx`
  (mirror `dispatchjobs/enqueuer_test.go`'s `fakeRiverInserter` pattern — check if a shared fake
  already exists in this package before writing a new one).
- `go build ./...` green; `go test ./internal/modules/documents/approval/jobs/...` green (real unit
  test, no DB needed — this is a queue-routing assertion against a fake, not an integration test).
- Commit.

## Test strategy
Unit-level (fake River client capturing `InsertOpts`) — no DB/testdb needed, this is a pure
routing-parameter assertion. The live re-drive (Validation Gate item 3) is a separate, main-session
step after T1 lands and binaries rebuild.

## Files touched (census)
MODIFY `internal/modules/documents/approval/jobs/lifecycle_event_enqueuer.go`. NEW or EXTEND
`internal/modules/documents/approval/jobs/lifecycle_event_enqueuer_test.go`.
