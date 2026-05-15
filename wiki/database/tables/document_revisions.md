# public.document_revisions

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
| `revision_num` | `bigint` | no | Baseline column. |
| `parent_revision_id` | `uuid` | yes | Baseline column. |
| `session_id` | `uuid` | no | Baseline column. |
| `storage_key` | `text` | no | Baseline column. |
| `content_hash` | `text` | no | Baseline column. |
| `form_data_snapshot` | `jsonb` | yes | Baseline column. |
| `created_at` | `timestamp with time zone` | no | Baseline column. |

## Baseline Definition

```sql
CREATE TABLE public.document_revisions (
id uuid DEFAULT gen_random_uuid() NOT NULL,
    document_id uuid NOT NULL,
    revision_num bigint NOT NULL,
    parent_revision_id uuid,
    session_id uuid NOT NULL,
    storage_key text NOT NULL,
    content_hash text NOT NULL,
    form_data_snapshot jsonb,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);
```

## Runtime Usage

Use `rg -n "document_revisions" internal apps` and the owning module wiki to verify readers/writers before changing this table.

## Seed or Reference Data

Check `db/reference-data/0001_product_reference_data.sql` and `db/dev-seeds/0001_local_dev_seed.sql` before adding rows.

## Notes and Debt

Retained in `public` because current runtime/baseline truth still uses it. Do not move schemas without an approved migration plan.
