# public.document_checkpoints

> **Source:** `db/baseline/0001_current_schema.sql`
> **Schema:** `public`
> **Owner:** documents

## Purpose
Current curated-baseline table owned by `documents`. See the owning module wiki and runtime repositories for business behavior.

## Columns

| Column | Type | Nullable | Meaning |
|---|---|---|---|
| `id` | `uuid` | no | Baseline column. |
| `document_id` | `uuid` | no | Baseline column. |
| `revision_id` | `uuid` | no | Baseline column. |
| `version_num` | `integer` | no | Baseline column. |
| `label` | `text` | yes | Baseline column. |
| `created_at` | `timestamp with time zone` | no | Baseline column. |
| `created_by` | `text` | no | Baseline column. |

## Baseline Definition

```sql
CREATE TABLE public.document_checkpoints (
id uuid DEFAULT gen_random_uuid() NOT NULL,
    document_id uuid NOT NULL,
    revision_id uuid NOT NULL,
    version_num integer NOT NULL,
    label text,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    created_by text NOT NULL
);
```

## Runtime Usage

Use `rg -n "document_checkpoints" internal apps` and the owning module wiki to verify readers/writers before changing this table.

## Seed or Reference Data

Check `db/reference-data/0001_product_reference_data.sql` and `db/dev-seeds/0001_local_dev_seed.sql` before adding rows.

## Notes and Debt

Retained in `public` because current runtime/baseline truth still uses it. Do not move schemas without an approved migration plan.
