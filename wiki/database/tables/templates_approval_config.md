# public.templates_approval_config

> **Source:** `db/baseline/0001_current_schema.sql`
> **Schema:** `public`
> **Owner:** templates

## Purpose
Current curated-baseline table owned by `templates`. See the owning module wiki and runtime repositories for business behavior.

## Columns

| Column | Type | Nullable | Meaning |
|---|---|---|---|
| `template_id` | `uuid` | no | Baseline column. |
| `reviewer_role` | `text` | yes | Baseline column. |
| `approver_role` | `text` | no | Baseline column. |

## Baseline Definition

```sql
CREATE TABLE public.templates_approval_config (
template_id uuid NOT NULL,
    reviewer_role text,
    approver_role text NOT NULL
);
```

## Runtime Usage

Use `rg -n "templates_approval_config" internal apps` and the owning module wiki to verify readers/writers before changing this table.

## Seed or Reference Data

Check `db/reference-data/0001_product_reference_data.sql` and `db/dev-seeds/0001_local_dev_seed.sql` before adding rows.

## Notes and Debt

Retained in `public` because current runtime/baseline truth still uses it. Do not move schemas without an approved migration plan.
