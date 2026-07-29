# metaldocs.outbox_events

> **Source:** `db/baseline/0001_current_schema.sql`
> **Schema:** `metaldocs`
> **Owner:** platform/workers

## Purpose
Current curated-baseline table owned by `platform/workers`. See the owning module wiki and runtime repositories for business behavior.

## Columns

| Column | Type | Nullable | Meaning |
|---|---|---|---|
| `event_id` | `text` | no | Baseline column. |
| `event_type` | `text` | no | Baseline column. |
| `aggregate_type` | `text` | no | Baseline column. |
| `aggregate_id` | `text` | no | Baseline column. |
| `occurred_at` | `timestamp with time zone` | no | Baseline column. |
| `version` | `integer` | no | Baseline column. |
| `idempotency_key` | `text` | no | Baseline column. |
| `producer` | `text` | no | Baseline column. |
| `trace_id` | `text` | no | Baseline column. |
| `payload` | `jsonb` | no | Baseline column. |
| `published_at` | `timestamp with time zone` | yes | Baseline column. |
| `attempt_count` | `integer` | no | Baseline column. |
| `last_error` | `text` | yes | Baseline column. |
| `last_attempt_at` | `timestamp with time zone` | yes | Baseline column. |
| `next_attempt_at` | `timestamp with time zone` | yes | Baseline column. |
| `dead_lettered_at` | `timestamp with time zone` | yes | Baseline column. |

## Baseline Definition

```sql
CREATE TABLE metaldocs.outbox_events (
event_id text NOT NULL,
    event_type text NOT NULL,
    aggregate_type text NOT NULL,
    aggregate_id text NOT NULL,
    occurred_at timestamp with time zone NOT NULL,
    version integer NOT NULL,
    idempotency_key text NOT NULL,
    producer text NOT NULL,
    trace_id text NOT NULL,
    payload jsonb DEFAULT '{}'::jsonb NOT NULL,
    published_at timestamp with time zone,
    attempt_count integer DEFAULT 0 NOT NULL,
    last_error text,
    last_attempt_at timestamp with time zone,
    next_attempt_at timestamp with time zone,
    dead_lettered_at timestamp with time zone
);
```

## Runtime Usage

Use `rg -n "outbox_events" internal apps` and the owning module wiki to verify readers/writers before changing this table.

All SQL against this table lives in `internal/platform/messaging/outbox/postgres` — `publisher.go` (insert), `consumer.go` (claim / mark published / mark failed), `retention.go` (purge). No other package may query it directly.

## Retention

Terminal rows are purged by the `outbox-events-retention` River periodic job (`internal/modules/jobs/outbox_retention`, executed by `metaldocs-jobs` per ADR 0067; enqueued by whichever of `metaldocs-api` / `metaldocs-jobs` wins leader election). It ticks every 24 h and runs two independent bounded-batch deletes:

| Class | Predicate | Window |
|---|---|---|
| Published | `published_at IS NOT NULL AND published_at < cutoff AND dead_lettered_at IS NULL` | 7 days |
| Dead-lettered | `dead_lettered_at IS NOT NULL AND dead_lettered_at < cutoff` | 90 days |

**Fail-closed:** a row that is neither published nor dead-lettered is ineligible at any age — it is still claimable work. The two classes use separate statements with separate cutoffs so the short published window can never reach the DLQ, and the dead-letter window is deliberately long because a dead-lettered row is the only durable record that an event never reached its consumer.

Retention is not merely a size control. `idempotency_key` is UNIQUE and `publisher.go` inserts with `ON CONFLICT (idempotency_key) DO NOTHING`, so a retained terminal row pins its key and silently swallows any later re-publish on it; before this job existed nothing ever deleted from this table, so keys were pinned forever. See `wiki/backend/platform/async-messaging.md` §2.2 and §3.1.

`metaldocs_app` was granted `DELETE` on this table by `db/migrations/0314_outbox_events_retention_grant.sql`; prior grants were `INSERT` (archived migration 0008) and `SELECT`, `UPDATE` (archived migration 0019).

## Seed or Reference Data

Check `db/reference-data/0001_product_reference_data.sql` and `db/dev-seeds/0001_local_dev_seed.sql` before adding rows.

## Notes and Debt

Curated baseline table in the `metaldocs` schema.
