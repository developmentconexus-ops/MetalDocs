# F5.5 — fanout ordering guarantee (spec)

> **Milestone:** M5 · **Status:** in progress
> **Binding parent:** `../validation-contract.md` §5 (fanout commutativity proof).
> On any conflict the **contract wins**; divergence is HS-7.
> **Rails:** ADR 0067 §4 (lifecycle fanout is idempotent-commutative; no ordering guard needed —
> proof-only feature unless the proof reveals otherwise, contract §5.2).

## Consumer contract (who consumes the output, and the shape required)

- **Consumer 1 — the milestone-validator.** Requires a concurrent-race integration test (testdb, real
  Postgres, real concurrency — not sqlmock, contract §7) proving that racing a `published` fanout and
  a `superseded` fanout for the **same document** in both possible interleavings produces an
  **identical terminal notification-row set**, that redelivering either event is a no-op (no
  duplicate rows), and that no obligated recipient's row is lost or overwritten.
- **Consumer 2 — the ADR 0067 §4 commutativity argument.** Requires the argument (per-event distinct
  rows keyed by `source_event_id`, no shared mutable "current status" row, `ON CONFLICT
  (recipient_user_id, source_event_id) WHERE source_event_id IS NOT NULL DO NOTHING`) to be verified
  against the ACTUAL runtime code (`internal/modules/notifications/infrastructure/fanout_worker.go`),
  not merely asserted — read fully at spec time and cited below.

## Runtime verification of the claim (read at spec time, 2026-07-04)

`fanout_worker.go` confirms every clause of contract §5.1:
- `insertRow` (`fanout_worker.go:123-137`): `INSERT INTO metaldocs.notifications (tenant_id,
  recipient_user_id, event_type, resource_type, resource_id, title, message, source_event_id) ...
  ON CONFLICT (recipient_user_id, source_event_id) WHERE source_event_id IS NOT NULL DO NOTHING` —
  keyed on `args.EventID` (minted per lifecycle event, `documentsdomain.LifecycleEventArgs`).
  Idempotent and additive: no `UPDATE`, no shared row, no `ON CONFLICT ... DO UPDATE`.
- `Work` (`:47-86`) fans a `published`/`superseded`/`obsoleted` event out to every obligated reader
  (`fanoutToReaders`) or an `approved`/`rejected` event to the submitter (`fanoutToAuthor`) — each
  call is scoped to **one** `job.Args` (one event), so two different events (`published` vs
  `superseded`) for the same document run as two independent `Work` invocations, each its own seeded
  tx, each inserting its own `source_event_id`-keyed rows. There is no code path where one event's
  `Work` call reads or mutates a row inserted by a different event's `Work` call — confirming the
  "no shared mutable projection" premise.
- Conclusion: the commutativity claim holds by construction (additive, disjoint-key inserts). The
  race test below is the runtime proof, not a design decision — no code change is expected unless the
  proof surfaces something the read above missed (e.g. a trigger or view with hidden shared state —
  none found in this file, but the proof is real DB, not just code reading).

## What to implement (per contract §5.2)

A single concurrent-race integration test (testdb, real Postgres — genuine concurrency, not mocked):
for **one** document, race a `published` fanout `Work()` call and a `superseded` fanout `Work()` call
in **both** interleavings (deterministic barrier — e.g. a `sync.WaitGroup`/channel forcing both
goroutines to start their tx at the same instant, run twice with the barrier release order reversed,
or run truly concurrently via goroutines + a start-gate channel so both orderings are exercised across
repeated runs). Assert the 3 rows in contract §5.2's table.

No new production code is anticipated — this feature is proof-only unless the test reveals a defect.

## Non-goals

- **No ordering key/guard added preemptively.** Contract §5.2: only add one if the test proves it's
  needed (that would be an HS-6 scope surface — surfaced, not silently done).
- **No change to `fanout_worker.go`, `LifecycleEventArgs`, or the notifications schema** unless the
  proof fails.
- **No metrics/observability** (M8 defer, contract §8).

## Validation Gate (acceptance — all must hold)

1. **Terminal notification-row set identical across both interleavings** — same recipients, same
   `source_event_id`s, same `event_type`s, regardless of which fanout's tx commits first.
2. **Idempotency — redelivering either event inserts no duplicate row** (re-run the same `Work()` call
   a second time post-race; row count unchanged).
3. **No lost/inverted state** — every obligated recipient has both events' rows (2 rows per recipient:
   one `published`-keyed, one `superseded`-keyed); none dropped, none overwritten.
4. **Real concurrency, real Postgres** (testdb factory, actual goroutines racing actual transactions —
   not a sequential simulation).
5. **Section-by-section match to contract §5.** If the test reveals a shared mutable projection
   (breaking commutativity), that is an HS-6 scope surface: report it, do not silently add a fix
   inside this feature without surfacing the deviation first.

## Interview record

No operator interview — contract §5 (operator-committed at D4) is the locked spec; F5.5 is its proof.
No design freedom: the commutativity mechanism (additive per-event rows, `ON CONFLICT` dedup) is
already implemented (pre-dates M5); F5.5 only proves it under real concurrency.

## ADR

No new ADR — F5.5 proves ADR 0067 §4's already-accepted claim. If the proof fails, escalate per
contract §5.2 (HS-6), which may require a new ADR at that point — not anticipated.
