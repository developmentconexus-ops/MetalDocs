# metaldocs.document_attachments

> **Source:** `db/baseline/0001_current_schema.sql`
> **Schema:** `metaldocs`
> **Owner:** documents

## Purpose
Current curated-baseline table owned by `documents`. See the owning module wiki and runtime repositories for business behavior.

## Columns

| Column | Type | Nullable | Meaning |
|---|---|---|---|
| `id` | `text` | no | Baseline column. |
| `document_id` | `text` | no | Baseline column. |
| `file_name` | `text` | no | Baseline column. |
| `content_type` | `text` | no | Baseline column. |
| `size_bytes` | `bigint` | no | Baseline column. |
| `storage_key` | `text` | no | Baseline column. |
| `uploaded_by` | `text` | no | Baseline column. |
| `created_at` | `timestamp with time zone` | no | Baseline column. |

## Baseline Definition

```sql
CREATE TABLE metaldocs.document_attachments (
id text NOT NULL,
    document_id text NOT NULL,
    file_name text NOT NULL,
    content_type text NOT NULL,
    size_bytes bigint NOT NULL,
    storage_key text NOT NULL,
    uploaded_by text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT document_attachments_size_bytes_check CHECK ((size_bytes > 0))
);
```

## Runtime Usage

Use `rg -n "document_attachments" internal apps` and the owning module wiki to verify readers/writers before changing this table.

## Seed or Reference Data

Check `db/reference-data/0001_product_reference_data.sql` and `db/dev-seeds/0001_local_dev_seed.sql` before adding rows.

## Notes and Debt

Curated baseline table in the `metaldocs` schema.
