# public.document_exports

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
| `composite_hash` | `bytea` | no | Baseline column. |
| `storage_key` | `text` | no | Baseline column. |
| `size_bytes` | `bigint` | no | Baseline column. |
| `paper_size` | `text` | no | Baseline column. |
| `landscape` | `boolean` | no | Baseline column. |
| `docgen_v2_ver` | `text` | no | Baseline column. |
| `created_at` | `timestamp with time zone` | no | Baseline column. |

## Baseline Definition

```sql
CREATE TABLE public.document_exports (
id uuid DEFAULT gen_random_uuid() NOT NULL,
    document_id uuid NOT NULL,
    revision_id uuid NOT NULL,
    composite_hash bytea NOT NULL,
    storage_key text NOT NULL,
    size_bytes bigint NOT NULL,
    paper_size text DEFAULT 'A4'::text NOT NULL,
    landscape boolean DEFAULT false NOT NULL,
    docgen_v2_ver text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT document_exports_composite_hash_check CHECK ((octet_length(composite_hash) = 32)),
    CONSTRAINT document_exports_size_bytes_check CHECK ((size_bytes > 0))
);
```

## Runtime Usage

Use `rg -n "document_exports" internal apps` and the owning module wiki to verify readers/writers before changing this table.

## Seed or Reference Data

Check `db/reference-data/0001_product_reference_data.sql` and `db/dev-seeds/0001_local_dev_seed.sql` before adding rows.

## Notes and Debt

Retained in `public` because current runtime/baseline truth still uses it. Do not move schemas without an approved migration plan.
