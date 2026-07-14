# metaldocs.document_versions

> **Source:** `db/baseline/0001_current_schema.sql`
> **Schema:** `metaldocs`
> **Owner:** documents

## Purpose
Current curated-baseline table owned by `documents`. See the owning module wiki and runtime repositories for business behavior.

## Columns

| Column | Type | Nullable | Meaning |
|---|---|---|---|
| `document_id` | `text` | no | Baseline column. |
| `version_number` | `integer` | no | Baseline column. |
| `content` | `text` | no | Baseline column. |
| `created_at` | `timestamp with time zone` | no | Baseline column. |
| `content_hash` | `text` | no | Baseline column. |
| `change_summary` | `text` | no | Baseline column. |
| `content_source` | `text` | no | Baseline column. |
| `native_content` | `jsonb` | yes | Baseline column. |
| `docx_storage_key` | `text` | yes | Baseline column. |
| `pdf_storage_key` | `text` | yes | Baseline column. |
| `text_content` | `text` | yes | Baseline column. |
| `file_size_bytes` | `bigint` | yes | Baseline column. |
| `original_filename` | `text` | yes | Baseline column. |
| `page_count` | `integer` | yes | Baseline column. |
| `search_vector` | `tsvector GENERATED ALWAYS AS (to_tsvector('portuguese'::regconfig, COALESCE(text_content, ''::text))) STORED` | yes | Baseline column. |
| `body_blocks` | `jsonb` | yes | Baseline column. |
| `values_json` | `jsonb` | no | Baseline column. |
| `template_key` | `text` | yes | Baseline column. |
| `template_version` | `integer` | yes | Baseline column. |
| `renderer_pin` | `jsonb` | yes | Baseline column. |
| `release_artifact_key` | `text` | yes | Baseline column. |
| `canonical_mddm_snapshot` | `jsonb` | yes | Baseline column. |

## Baseline Definition

```sql
CREATE TABLE metaldocs.document_versions (
document_id text NOT NULL,
    version_number integer NOT NULL,
    content text DEFAULT ''::text NOT NULL,
    created_at timestamp with time zone NOT NULL,
    content_hash text NOT NULL,
    change_summary text DEFAULT ''::text NOT NULL,
    content_source text DEFAULT 'native'::text NOT NULL,
    native_content jsonb,
    docx_storage_key text,
    pdf_storage_key text,
    text_content text,
    file_size_bytes bigint,
    original_filename text,
    page_count integer,
    search_vector tsvector GENERATED ALWAYS AS (to_tsvector('portuguese'::regconfig, COALESCE(text_content, ''::text))) STORED,
    body_blocks jsonb DEFAULT '[]'::jsonb,
    values_json jsonb DEFAULT '{}'::jsonb NOT NULL,
    template_key text,
    template_version integer,
    renderer_pin jsonb,
    release_artifact_key text,
    canonical_mddm_snapshot jsonb
);
```

## Runtime Usage

Use `rg -n "document_versions" internal apps` and the owning module wiki to verify readers/writers before changing this table.

## Seed or Reference Data

Check `db/reference-data/0001_product_reference_data.sql` and `db/dev-seeds/0001_local_dev_seed.sql` before adding rows.

## Notes and Debt

Curated baseline table in the `metaldocs` schema.
