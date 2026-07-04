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

## Commands (real output) — LIVE GREEN on real Postgres

Ran against a disposable `postgres:16` container (`METALDOCS_DATABASE_URL` set for the run only; `.env`
never touched). The harness (`testdb.Open`) built the template DB from prerequisites + baseline +
reference-data + **all migrations including `0272`**, then cloned per test.

```
$ go vet -tags integration ./internal/modules/documents/approval/application/   → clean, VET_EXIT=0

$ METALDOCS_DATABASE_URL=postgres://…@localhost:55432/postgres?sslmode=disable \
    go test ./internal/modules/documents/approval/application/ -tags integration -run TestPublishRace -count=1 -v
=== RUN   TestPublishRace
--- PASS: TestPublishRace (130.42s)
    --- PASS: TestPublishRace/ScheduledSeed_ManualFirst (103.93s)   # first run builds the template DB
    --- PASS: TestPublishRace/ScheduledSeed_SchedulerFirst (7.04s)
    --- PASS: TestPublishRace/ApprovedSeed_ManualFirst (9.21s)
    --- PASS: TestPublishRace/ApprovedSeed_SchedulerFirst (10.24s)
PASS
ok  metaldocs/internal/modules/documents/approval/application  134.096s
```

All four subtests green: single winner per seed, terminal `status='published'`, `revision_version==4`
(exactly one bump), exactly one `document_published` governance event. The safe-by-construction verdict
is now **empirically confirmed under real concurrency**, not only argued.

## Bounded defer — live green run — **CLOSED (2026-07-04)**

The live green run the original evidence deferred is now **observed** (above). Getting there corrected two
distinct root causes; the original "masked by CI GUC-seeding" hypothesis was imprecise and is superseded:

1. **Test-harness identity gap (the real blocker, PRIMARY).** The harness drove the services with
   `context.Background()` — no tenant/actor — so the M3 `TxRunner` auto-seed
   (`seedTxIdentityFromContext`) no-op'd and the in-tx `authz.Require` correctly failed closed. Fixed by
   seeding the context **exactly as production middleware does**: manual path gets
   `platformtenant.WithTenantID`+`WithActorID` (human lifecycle); scheduler path gets
   `authz.WithBackgroundBypass` and self-seeds its tenant in-tx via `SeedTxTenant` (async lifecycle) —
   mirroring `scheduled_publish_job.go`. Plus two fixture-precision fixes: `effectiveAt` truncated to
   µs (Postgres `timestamptz` resolution, or `scheduledJobMatchesState` spuriously no-ops), and the
   `governance_events` count query compares `resource_id` as **TEXT** (the column is text, not uuid).
   These are test-only changes — **no production publish-path code was touched.**

2. **authz NULL-GUC scan gap (secondary, now fixed as F4.5).** On a cold connection the missing identity
   surfaced as an opaque driver error (`converting NULL to string is unsupported`) instead of the
   documented `ErrActorContextMissing` sentinel, because `MustActorID`/`MustTenantID` scanned
   `current_setting(...,true)` (which returns SQL **NULL** for a never-set placeholder GUC) into a plain
   `string`. Fixed in **[F4.5](../f4.5-authz-guc-null-hardening/evidence.md)** — one shared
   `readSoftGUC` helper (`sql.NullString`, NULL⇒missing) now backs all four GUC readers. Fail-closed
   behavior is unchanged (stricter-or-equal); it only replaces a driver crash with the correct sentinel.
   `task_e03a4383` is thereby resolved inside M4, not deferred.

**No defers remain open for F4.2.**
