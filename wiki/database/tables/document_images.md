# metaldocs.document_images

> **Source:** `db/baseline/0001_current_schema.sql`
> **Schema:** `metaldocs`
> **Owner:** documents

## Purpose
Current curated-baseline table owned by `documents`. See the owning module wiki and runtime repositories for business behavior.

## Columns

| Column | Type | Nullable | Meaning |
|---|---|---|---|
| `id` | `uuid` | no | Baseline column. |
| `sha256` | `text` | no | Baseline column. |
| `mime_type` | `text` | no | Baseline column. |
| `byte_size` | `integer` | no | Baseline column. |
| `bytes` | `bytea` | no | Baseline column. |
| `created_at` | `timestamp with time zone` | no | Baseline column. |

## Baseline Definition

```sql
CREATE TABLE metaldocs.document_images (
id uuid DEFAULT gen_random_uuid() NOT NULL,
    sha256 text NOT NULL,
    mime_type text NOT NULL,
    byte_size integer NOT NULL,
    bytes bytea NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT document_images_byte_size_check CHECK ((byte_size > 0))
);
```

## Runtime Usage

Use `rg -n "document_images" internal apps` and the owning module wiki to verify readers/writers before changing this table.

## Seed or Reference Data

Check `db/reference-data/0001_product_reference_data.sql` and `db/dev-seeds/0001_local_dev_seed.sql` before adding rows.

## Notes and Debt

Curated baseline table in the `metaldocs` schema.
