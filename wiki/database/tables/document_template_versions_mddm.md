# metaldocs.document_template_versions_mddm

> **Source:** `db/baseline/0001_current_schema.sql`
> **Schema:** `metaldocs`
> **Owner:** templates

## Purpose
Current curated-baseline table owned by `templates`. See the owning module wiki and runtime repositories for business behavior.

## Columns

| Column | Type | Nullable | Meaning |
|---|---|---|---|
| `id` | `uuid` | no | Baseline column. |
| `template_id` | `uuid` | no | Baseline column. |
| `version` | `integer` | no | Baseline column. |
| `mddm_version` | `integer` | no | Baseline column. |
| `content_blocks` | `jsonb` | no | Baseline column. |
| `content_hash` | `text` | no | Baseline column. |
| `is_published` | `boolean` | no | Baseline column. |
| `created_at` | `timestamp with time zone` | no | Baseline column. |

## Baseline Definition

```sql
CREATE TABLE metaldocs.document_template_versions_mddm (
id uuid DEFAULT gen_random_uuid() NOT NULL,
    template_id uuid NOT NULL,
    version integer NOT NULL,
    mddm_version integer NOT NULL,
    content_blocks jsonb NOT NULL,
    content_hash text NOT NULL,
    is_published boolean DEFAULT false NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT document_template_versions_mddm_mddm_version_check CHECK ((mddm_version >= 1)),
    CONSTRAINT document_template_versions_mddm_version_check CHECK ((version >= 1))
);
```

## Runtime Usage

Use `rg -n "document_template_versions_mddm" internal apps` and the owning module wiki to verify readers/writers before changing this table.

## Seed or Reference Data

Check `db/reference-data/0001_product_reference_data.sql` and `db/dev-seeds/0001_local_dev_seed.sql` before adding rows.

## Notes and Debt

Curated baseline table in the `metaldocs` schema.
