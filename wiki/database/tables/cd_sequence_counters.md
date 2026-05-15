# public.cd_sequence_counters

> **Source:** `db/baseline/0001_current_schema.sql`
> **Schema:** `public`
> **Owner:** registry

## Purpose
Current curated-baseline table owned by `registry`. See the owning module wiki and runtime repositories for business behavior.

## Columns

| Column | Type | Nullable | Meaning |
|---|---|---|---|
| `tenant_id` | `uuid` | no | Baseline column. |
| `profile_code` | `text` | no | Baseline column. |
| `process_area_code` | `text` | no | Baseline column. |
| `next_seq` | `integer` | no | Baseline column. |

## Baseline Definition

```sql
CREATE TABLE public.cd_sequence_counters (
tenant_id uuid NOT NULL,
    profile_code text NOT NULL,
    process_area_code text NOT NULL,
    next_seq integer DEFAULT 1 NOT NULL
);
```

## Runtime Usage

Use `rg -n "cd_sequence_counters" internal apps` and the owning module wiki to verify readers/writers before changing this table.

## Seed or Reference Data

Check `db/reference-data/0001_product_reference_data.sql` and `db/dev-seeds/0001_local_dev_seed.sql` before adding rows.

## Notes and Debt

Retained in `public` because current runtime/baseline truth still uses it. Do not move schemas without an approved migration plan.
