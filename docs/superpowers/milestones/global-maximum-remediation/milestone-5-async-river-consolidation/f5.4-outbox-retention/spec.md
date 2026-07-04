# F5.4 — outbox retention (spec)

> **Milestone:** M5 · **Status:** in progress
> **Binding parent:** `../validation-contract.md` §4 (retention policy + bounded-purge proof).
> On any conflict the **contract wins**; divergence is HS-7.
> **Rails:** ADR 0067 (River is the single async primitive; retention is a River-native mechanism +
> one new periodic purge job — no new triggers, no new business logic in triggers, contract §6).

## Consumer contract (who consumes the output, and the shape required)

- **Consumer 1 — operators inspecting the DLQ / disk growth.** Require `pdf_dispatch_outbox` and
  `materialize_dispatch_outbox` to NOT grow unbounded: `status='dispatched'` rows older than 7 days
  are purged; `dead_lettered_at IS NOT NULL` rows are **retained** (never auto-purged — inspectable
  DLQ, REQ-ASYNC-3); recent dispatched rows are untouched.
- **Consumer 2 — River's own job table.** Requires `CompletedJobRetentionPeriod=24h`,
  `CancelledJobRetentionPeriod=24h`, `DiscardedJobRetentionPeriod=7*24h` set on the `metaldocs-jobs`
  client config — River's native cleaner does the rest; **no new code**, config only.
- **Consumer 3 — the milestone-validator.** Requires a retention integration proof (testdb): seed
  old-dispatched + recent-dispatched + dead-lettered rows → run the purge job → assert exactly the
  old-dispatched rows are gone, everything else survives; assert the DELETE is batch-bounded (no
  single unbounded statement).

## What to implement (per contract §4.1)

| Rows | Policy | Mechanism |
|---|---|---|
| River's own job rows | Completed 24h / Cancelled 24h / Discarded 7×24h | River client config only (`riverjobs.Config` fields already threaded in F5.2 T1) — set in `apps/jobs/cmd/metaldocs-jobs/main.go`'s client construction |
| `*_dispatch_outbox` — `status='dispatched'` | delete `dispatched_at < now() - 7d` | new River periodic job `ID:"staging-outbox-purge"`, `PeriodicInterval(24*time.Hour)`, queue `maintenance`; body: bounded batch **5000 rows × ≤10 iterations/run** per table (pdf + materialize) |
| `*_dispatch_outbox` — `dead_lettered_at IS NOT NULL` | retained, never purged | purge query's `WHERE` clause excludes these by construction (`status='dispatched' AND dispatched_at < now()-7d`; a dead-lettered row's status is `'failed'`, never `'dispatched'` — excluded automatically, no extra guard needed beyond matching the status filter) |

Values (7d retention, 5000×≤10 batch, 24h cadence) are **binding** per contract §4.1 — changing them
is HS-7.

## Non-goals

- **No dead-lettered row pruning.** Contract §8 bounded defer (ops-readiness decision beyond M5).
- **No Prometheus / metrics for purge activity.** Contract §8 (M8 defer).
- **No new triggers, no schema change to add a retention column** — `dispatched_at`/`dead_lettered_at`
  already exist (used unmodified since before F5.3).
- **No change to River's own row purge mechanism** — config only, no hand-rolled cleaner.

## Validation Gate (acceptance — all must hold)

1. **Retention integration proof (testdb, targeted `-run`)** (contract §4.2): seed (a) dispatched rows
   with `dispatched_at` > 7d old, (b) recent dispatched rows, (c) a dead-lettered row → run the purge
   job once → assert (a) gone, (b) and (c) survive, for **both** pdf and materialize tables.
2. **Batch-bounded proof:** the purge query is provably bounded (LIMIT/batch loop capped at 5000×10,
   not a single unbounded DELETE) — either by seeding > 5000 eligible rows and asserting the job caps
   its own single-run deletion count, or by direct code inspection cited in evidence with the exact
   bound.
3. **River native retention config set:** `metaldocs-jobs`' River client config carries
   `CompletedJobRetentionPeriod: 24*time.Hour`, `CancelledJobRetentionPeriod: 24*time.Hour`,
   `DiscardedJobRetentionPeriod: 7*24*time.Hour` — confirmed by reading the constructed `river.Config`
   (no proof needed beyond config-value assertion; River's own cleaner is out-of-repo tested code).
4. **`go build ./...` green; all 4 binaries build.**
5. **Section-by-section match to contract §4.** No divergence, or divergence surfaced as HS-7.

## Interview record

No operator interview — contract §4 (operator-committed at D4) is the locked spec; F5.4 is its
execution. No design freedom: retention values, mechanism (River periodic purge job vs. e.g. a
separate cron), and batch bound are all pre-decided in the contract.

## ADR

No new ADR — F5.4 executes ADR 0067's accepted decision (River as the single async primitive
including retention).
