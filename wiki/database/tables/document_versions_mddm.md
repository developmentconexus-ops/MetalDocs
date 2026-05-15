# metaldocs.document_versions_mddm

> **Source:** `db/baseline/0001_current_schema.sql`
> **Schema:** `metaldocs`
> **Owner:** documents

## Purpose
Current curated-baseline table owned by `documents`. See the owning module wiki and runtime repositories for business behavior.

## Columns

| Column | Type | Nullable | Meaning |
|---|---|---|---|
| `id` | `uuid` | no | Baseline column. |
| `document_id` | `text` | no | Baseline column. |
| `version_number` | `integer` | no | Baseline column. |
| `revision_label` | `text` | no | Baseline column. |
| `status` | `metaldocs.mddm_version_status` | no | Baseline column. |
| `content_blocks` | `jsonb` | yes | Baseline column. |
| `docx_bytes` | `bytea` | yes | Baseline column. |
| `template_ref` | `jsonb` | yes | Baseline column. |
| `content_hash` | `text` | yes | Baseline column. |
| `revision_diff` | `jsonb` | yes | Baseline column. |
| `change_summary` | `text` | yes | Baseline column. |
| `created_at` | `timestamp with time zone` | no | Baseline column. |
| `created_by` | `text` | no | Baseline column. |
| `approved_at` | `timestamp with time zone` | yes | Baseline column. |
| `approved_by` | `text` | yes | Baseline column. |

## Baseline Definition

```sql
CREATE TABLE metaldocs.document_versions_mddm (
id uuid DEFAULT gen_random_uuid() NOT NULL,
    document_id text NOT NULL,
    version_number integer NOT NULL,
    revision_label text NOT NULL,
    status metaldocs.mddm_version_status NOT NULL,
    content_blocks jsonb,
    docx_bytes bytea,
    template_ref jsonb,
    content_hash text,
    revision_diff jsonb,
    change_summary text,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    created_by text NOT NULL,
    approved_at timestamp with time zone,
    approved_by text,
    CONSTRAINT document_versions_mddm_version_number_check CHECK ((version_number >= 1))
);
```

## Runtime Usage

Use `rg -n "document_versions_mddm" internal apps` and the owning module wiki to verify readers/writers before changing this table.

## Seed or Reference Data

Check `db/reference-data/0001_product_reference_data.sql` and `db/dev-seeds/0001_local_dev_seed.sql` before adding rows.

## Notes and Debt

Curated baseline table in the `metaldocs` schema.
