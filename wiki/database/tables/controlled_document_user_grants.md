# public.controlled_document_user_grants

> **Source:** `db/baseline/0001_current_schema.sql`
> **Schema:** `public`
> **Owner:** controlled-documents

## Purpose
Current curated-baseline table owned by the `controlled-documents` module. See the owning module wiki and runtime repositories for business behavior.

## Columns

| Column | Type | Nullable | Meaning |
|---|---|---|---|
| `tenant_id` | `uuid` | no | Baseline column. |
| `controlled_document_id` | `uuid` | no | Baseline column. |
| `user_id` | `text` | no | Baseline column. |
| `created_at` | `timestamp with time zone` | no | Baseline column. |

## Baseline Definition

```sql
CREATE TABLE public.controlled_document_user_grants (
tenant_id uuid NOT NULL,
    controlled_document_id uuid NOT NULL,
    user_id text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);
```

## Runtime Usage

Use `rg -n "controlled_document_user_grants" internal apps` and the owning module wiki to verify readers/writers before changing this table.

## Seed or Reference Data

Check `db/reference-data/0001_product_reference_data.sql` and `db/dev-seeds/0001_local_dev_seed.sql` before adding rows.

## Notes and Debt

Retained in `public` because current runtime/baseline truth still uses it. Do not move schemas without an approved migration plan.
