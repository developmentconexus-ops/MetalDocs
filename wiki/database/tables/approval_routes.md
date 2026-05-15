# public.approval_routes

> **Source:** `db/baseline/0001_current_schema.sql`
> **Schema:** `public`
> **Owner:** approval

## Purpose
Current curated-baseline table owned by `approval`. See the owning module wiki and runtime repositories for business behavior.

## Columns

| Column | Type | Nullable | Meaning |
|---|---|---|---|
| `id` | `uuid` | no | Baseline column. |
| `tenant_id` | `uuid` | no | Baseline column. |
| `profile_code` | `text` | no | Baseline column. |
| `name` | `text` | no | Baseline column. |
| `version` | `integer` | no | Baseline column. |
| `created_at` | `timestamp with time zone` | no | Baseline column. |
| `created_by` | `text` | no | Baseline column. |
| `active` | `boolean` | no | Baseline column. |

## Baseline Definition

```sql
CREATE TABLE public.approval_routes (
id uuid DEFAULT gen_random_uuid() NOT NULL,
    tenant_id uuid NOT NULL,
    profile_code text NOT NULL,
    name text NOT NULL,
    version integer DEFAULT 1 NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    created_by text NOT NULL,
    active boolean DEFAULT true NOT NULL
);
```

## Runtime Usage

Use `rg -n "approval_routes" internal apps` and the owning module wiki to verify readers/writers before changing this table.

## Seed or Reference Data

Check `db/reference-data/0001_product_reference_data.sql` and `db/dev-seeds/0001_local_dev_seed.sql` before adding rows.

## Notes and Debt

Retained in `public` because current runtime/baseline truth still uses it. Do not move schemas without an approved migration plan.
