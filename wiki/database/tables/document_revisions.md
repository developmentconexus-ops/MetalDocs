# public.document_revisions

> **Source:** `db/baseline/0001_current_schema.sql`
> **Schema:** `public`
> **Owner:** documents

## Purpose
Technical/autosave artifact-revision storage owned by `documents`. This table is not governed business revision history; business/governed lineage remains sourced from `public.documents` revision state.

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
| `file_size_bytes` | `bigint` | yes | Server-authoritative size in bytes of the saved DOCX artifact for this technical revision row. |
| `page_count` | `integer` | yes | Client- or server-derived rendered page count for this saved technical revision row. |
| `page_count_source` | `text` | yes | Provenance for `page_count`; allowed values are `eigenpal_client` and `server_renderer`. |
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
