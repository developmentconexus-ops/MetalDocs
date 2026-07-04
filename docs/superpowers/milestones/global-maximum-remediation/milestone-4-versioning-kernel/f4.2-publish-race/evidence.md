# F4.2 evidence — scheduled-vs-manual publish race

> **Contract:** `../validation-contract.md` §2. Commit `9c9ad946`.
> **Outcome:** race proven **safe by construction** — no `PublishRevision` choke point added, no
> production code changed (contract §0.4/§2.3 expected outcome).

## What shipped

`internal/modules/documents/approval/application/publish_race_integration_test.go` (new;
`//go:build integration`; package `application`; testdb factory, **NOT sqlmock**).

Real concurrency harness: `runPublishRaceInterleaving(t, seed, manualFirst)` — both goroutines block on
one `start` channel closed once (`close(start)`), so the race is genuinely concurrent; launch order only
nudges the Go scheduler, the DB decides the winner.

**4 subtests** over the two mutually-exclusive seed states × both launch orders (covers §2.2's full
winner/loser table — both loser sentinels):

| Subtest | Seed | Winner | Loser + sentinel | Half proven |
|---|---|---|---|---|
| ScheduledSeed_ManualFirst / _SchedulerFirst | `status='scheduled'` | scheduler (`nil`) | manual → `repository.ErrStaleRevision` (its `WHERE status='approved'` matches 0 rows) | predicate divergence |
| ApprovedSeed_ManualFirst / _SchedulerFirst | `status='approved'` | manual (`nil`) | scheduler → no-op-to-`nil` | genuine `FOR UPDATE` vs `UPDATE` lock contention |

Each subtest also asserts: terminal `status='published'`, `revision_version == 4` (3+1, exactly once),
and exactly **1** `document_published` governance event (single side effect, no double-emit).

**Scheduler no-op sentinel (read, not guessed):** `RunScheduledPublishJob` maps **both** no-op paths —
`scheduledJobMatchesState==false` and the `errScheduledPublishNoOp` OCC sentinel — to a successful `nil`
(`scheduler_service.go`); the sentinel never escapes to the caller. So the approved-seed assertion is
`schedErr == nil`, and the no-double-bump / single-event checks prove the scheduler did not mutate the
row. Asserting `errors.Is(schedErr, errScheduledPublishNoOp)` would be wrong (swallowed internally).

## Verdict (contract §2.3)

Safe by construction: the two OCC UPDATEs have **mutually-exclusive `status` predicates** (`approved` vs
`scheduled`) plus a `revision_version` CAS — a single row satisfies at most one, so no interleaving yields
two winners. **No `PublishRevision` choke point warranted; none added. No production code changed.**

## Commands (real output)

```
$ go vet -tags integration ./internal/modules/documents/approval/application/   → clean, VET_EXIT=0
```

## Bounded defer — live green run (honest disclosure)

A live green run is **not** observed. Root cause is a **pre-existing** platform gap, NOT a defect in this
deliverable: `authz.MustActorID` / `MustTenantID` (`internal/modules/iam/authz/context.go`) `Scan` a
`SELECT current_setting('metaldocs.actor_id', true)` result into a plain `string`; on a cold pooled
connection never touched by `set_config`, Postgres returns SQL NULL and the `Scan` fails with
`converting NULL to string is unsupported` instead of the intended `ErrActorContextMissing` sentinel.

Verified this is environmental, not test-quality: the **reference** integration test
(`TestPublishApproved_DoesNotAutoCreateNextVersion`, the file this one mirrors) fails **byte-identically**
on the same disposable `postgres:16` container. The project's real CI/test-DB image seeds GUCs (M3 TxRunner
auto-seed) and masks this on cold connections; the disposable container bypassed that path.

| Defer | Rationale | Trigger / owner |
|---|---|---|
| Live `-tags integration` green run | 20-min box (contract §5/§6) + pre-existing authz NULL-GUC scan gap blocks cold-connection runs | Run on project CI/real test-DB image before program close-out; drive authored + committed (M1–M3 precedent) |
| Fix authz NULL-GUC scan → `ErrActorContextMissing` on cold connection | Out of M4 boundary (iam/authz platform robustness, not the publish race); flagged as `task_e03a4383` | M8 ops-readiness or a platform-hardening pass |
