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

## Seed or Reference Data

Check `db/reference-data/0001_product_reference_data.sql` and `db/dev-seeds/0001_local_dev_seed.sql` before adding rows.

## Notes and Debt

Curated baseline table in the `metaldocs` schema.
