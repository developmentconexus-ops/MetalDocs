# public.template_versions

> **Source:** `db/baseline/0001_current_schema.sql`
> **Schema:** `public`
> **Owner:** templates

## Purpose
Current curated-baseline table owned by `templates`. See the owning module wiki and runtime repositories for business behavior.

## Columns

| Column | Type | Nullable | Meaning |
|---|---|---|---|
| `id` | `uuid` | no | Baseline column. |
| `template_id` | `uuid` | no | Baseline column. |
| `version_num` | `integer` | no | Baseline column. |
| `status` | `text` | no | Baseline column. |
| `grammar_version` | `integer` | no | Baseline column. |
| `docx_storage_key` | `text` | no | Baseline column. |
| `schema_storage_key` | `text` | no | Baseline column. |
| `docx_content_hash` | `text` | no | Baseline column. |
| `schema_content_hash` | `text` | no | Baseline column. |
| `published_at` | `timestamp with time zone` | yes | Baseline column. |
| `published_by` | `text` | yes | Baseline column. |
| `deprecated_at` | `timestamp with time zone` | yes | Baseline column. |
| `lock_version` | `integer` | no | Baseline column. |
| `created_at` | `timestamp with time zone` | no | Baseline column. |
| `updated_at` | `timestamp with time zone` | no | Baseline column. |
| `created_by` | `text` | no | Baseline column. |

## Baseline Definition

```sql
CREATE TABLE public.template_versions (
id uuid DEFAULT gen_random_uuid() NOT NULL,
    template_id uuid NOT NULL,
    version_num integer NOT NULL,
    status text NOT NULL,
    grammar_version integer DEFAULT 1 NOT NULL,
    docx_storage_key text NOT NULL,
    schema_storage_key text NOT NULL,
    docx_content_hash text NOT NULL,
    schema_content_hash text NOT NULL,
    published_at timestamp with time zone,
    published_by text,
    deprecated_at timestamp with time zone,
    lock_version integer DEFAULT 0 NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    created_by text NOT NULL,
    CONSTRAINT template_versions_status_check CHECK ((status = ANY (ARRAY['draft'::text, 'published'::text, 'deprecated'::text])))
);
```

## Runtime Usage

Use `rg -n "template_versions" internal apps` and the owning module wiki to verify readers/writers before changing this table.

## Seed or Reference Data

Check `db/reference-data/0001_product_reference_data.sql` and `db/dev-seeds/0001_local_dev_seed.sql` before adding rows.

## Notes and Debt

Retained in `public` because current runtime/baseline truth still uses it. Do not move schemas without an approved migration plan.
