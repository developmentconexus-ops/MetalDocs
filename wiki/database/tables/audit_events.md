# metaldocs.audit_events

> **Source:** `db/baseline/0001_current_schema.sql`
> **Schema:** `metaldocs`
> **Owner:** audit

## Purpose
Current curated-baseline table owned by `audit`. See the owning module wiki and runtime repositories for business behavior.

## Columns

| Column | Type | Nullable | Meaning |
|---|---|---|---|
| `id` | `text` | no | Baseline column. |
| `occurred_at` | `timestamp with time zone` | no | Baseline column. |
| `actor_id` | `text` | no | Baseline column. |
| `action` | `text` | no | Baseline column. |
| `resource_type` | `text` | no | Baseline column. |
| `resource_id` | `text` | no | Baseline column. |
| `payload` | `jsonb` | no | Baseline column. |
| `trace_id` | `text` | no | Baseline column. |
| `tenant_id` | `text` | no | Baseline column. |
| `audit_sequence` | `bigint` | no | Baseline column. |
| `prev_hash` | `text` | no | Baseline column. |
| `row_hash` | `text` | no | Baseline column. |

## Baseline Definition

```sql
CREATE TABLE metaldocs.audit_events (
id text NOT NULL,
    occurred_at timestamp with time zone NOT NULL,
    actor_id text NOT NULL,
    action text NOT NULL,
    resource_type text NOT NULL,
    resource_id text NOT NULL,
    payload jsonb DEFAULT '{}'::jsonb NOT NULL,
    trace_id text NOT NULL,
    tenant_id text DEFAULT ''::text NOT NULL,
    audit_sequence bigint NOT NULL,
    prev_hash text DEFAULT ''::text NOT NULL,
    row_hash text DEFAULT ''::text NOT NULL
);
```

## Runtime Usage

Use `rg -n "audit_events" internal apps` and the owning module wiki to verify readers/writers before changing this table.

## Seed or Reference Data

Check `db/reference-data/0001_product_reference_data.sql` and `db/dev-seeds/0001_local_dev_seed.sql` before adding rows.

## Notes and Debt

Curated baseline table in the `metaldocs` schema.
