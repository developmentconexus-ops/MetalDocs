# metaldocs.job_leases

> **Source:** `db/baseline/0001_current_schema.sql`
> **Schema:** `metaldocs`
> **Owner:** platform/workers

## Purpose
Current curated-baseline table owned by `platform/workers`. See the owning module wiki and runtime repositories for business behavior.

## Columns

| Column | Type | Nullable | Meaning |
|---|---|---|---|
| `job_name` | `text` | no | Baseline column. |
| `leader_id` | `text` | no | Baseline column. |
| `lease_epoch` | `bigint` | no | Baseline column. |
| `acquired_at` | `timestamp with time zone` | no | Baseline column. |
| `heartbeat_at` | `timestamp with time zone` | no | Baseline column. |
| `expires_at` | `timestamp with time zone` | no | Baseline column. |

## Baseline Definition

```sql
CREATE TABLE metaldocs.job_leases (
job_name text NOT NULL,
    leader_id text NOT NULL,
    lease_epoch bigint DEFAULT 0 NOT NULL,
    acquired_at timestamp with time zone DEFAULT now() NOT NULL,
    heartbeat_at timestamp with time zone DEFAULT now() NOT NULL,
    expires_at timestamp with time zone NOT NULL
);
```

## Runtime Usage

Use `rg -n "job_leases" internal apps` and the owning module wiki to verify readers/writers before changing this table.

## Seed or Reference Data

Check `db/reference-data/0001_product_reference_data.sql` and `db/dev-seeds/0001_local_dev_seed.sql` before adding rows.

## Notes and Debt

Curated baseline table in the `metaldocs` schema.
