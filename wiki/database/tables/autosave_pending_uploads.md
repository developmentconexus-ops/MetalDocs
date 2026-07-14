# public.autosave_pending_uploads

> **Source:** `db/baseline/0001_current_schema.sql`
> **Schema:** `public`
> **Owner:** documents

## Purpose
Current curated-baseline table owned by `documents`. See the owning module wiki and runtime repositories for business behavior.

## Columns

| Column | Type | Nullable | Meaning |
|---|---|---|---|
| `id` | `uuid` | no | Baseline column. |
| `session_id` | `uuid` | no | Baseline column. |
| `document_id` | `uuid` | no | Baseline column. |
| `base_revision_id` | `uuid` | no | Baseline column. |
| `content_hash` | `text` | no | Baseline column. |
| `storage_key` | `text` | no | Baseline column. |
| `presigned_at` | `timestamp with time zone` | no | Baseline column. |
| `expires_at` | `timestamp with time zone` | no | Baseline column. |
| `consumed_at` | `timestamp with time zone` | yes | Baseline column. |

## Baseline Definition

```sql
CREATE TABLE public.autosave_pending_uploads (
id uuid DEFAULT gen_random_uuid() NOT NULL,
    session_id uuid NOT NULL,
    document_id uuid NOT NULL,
    base_revision_id uuid NOT NULL,
    content_hash text NOT NULL,
    storage_key text NOT NULL,
    presigned_at timestamp with time zone DEFAULT now() NOT NULL,
    expires_at timestamp with time zone NOT NULL,
    consumed_at timestamp with time zone
);
```

## Runtime Usage

Use `rg -n "autosave_pending_uploads" internal apps` and the owning module wiki to verify readers/writers before changing this table.

## Seed or Reference Data

Check `db/reference-data/0001_product_reference_data.sql` and `db/dev-seeds/0001_local_dev_seed.sql` before adding rows.

## Notes and Debt

Retained in `public` because current runtime/baseline truth still uses it. Do not move schemas without an approved migration plan.
