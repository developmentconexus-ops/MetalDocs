# public.schema_migrations

> **Source:** `db/baseline/0001_current_schema.sql`
> **Schema:** `public`
> **Owner:** platform/db tooling

## Purpose
Current curated-baseline table owned by `platform/db tooling`. See the owning module wiki and runtime repositories for business behavior.

## Columns

| Column | Type | Nullable | Meaning |
|---|---|---|---|
| `version` | `text` | no | Baseline column. |
| `applied_at` | `timestamp with time zone` | no | Baseline column. |
| `description` | `text` | yes | Baseline column. |

## Baseline Definition

```sql
CREATE TABLE public.schema_migrations (
version text NOT NULL,
    applied_at timestamp with time zone DEFAULT now() NOT NULL,
    description text
);
```

## Runtime Usage

Use `rg -n "schema_migrations" internal apps` and the owning module wiki to verify readers/writers before changing this table.

## Seed or Reference Data

Check `db/reference-data/0001_product_reference_data.sql` and `db/dev-seeds/0001_local_dev_seed.sql` before adding rows.

## Notes and Debt

Retained in `public` because current runtime/baseline truth still uses it. Do not move schemas without an approved migration plan.
