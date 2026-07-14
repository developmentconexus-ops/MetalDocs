# public.templates_template_version

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
| `version_number` | `integer` | no | Baseline column. |
| `status` | `text` | no | Baseline column. |
| `docx_storage_key` | `text` | no | Baseline column. |
| `content_hash` | `text` | no | Baseline column. |
| `metadata_schema` | `jsonb` | no | Baseline column. |
| `placeholder_schema` | `jsonb` | no | Baseline column. |
| `author_id` | `text` | no | Baseline column. |
| `pending_reviewer_role` | `text` | yes | Baseline column. |
| `pending_approver_role` | `text` | no | Baseline column. |
| `reviewer_id` | `text` | yes | Baseline column. |
| `approver_id` | `text` | yes | Baseline column. |
| `submitted_at` | `timestamp with time zone` | yes | Baseline column. |
| `reviewed_at` | `timestamp with time zone` | yes | Baseline column. |
| `approved_at` | `timestamp with time zone` | yes | Baseline column. |
| `published_at` | `timestamp with time zone` | yes | Baseline column. |
| `obsoleted_at` | `timestamp with time zone` | yes | Baseline column. |
| `created_at` | `timestamp with time zone` | no | Baseline column. |
| `lock_version` | `integer` | no | Baseline column. |

## Baseline Definition

```sql
CREATE TABLE public.templates_template_version (
id uuid NOT NULL,
    template_id uuid NOT NULL,
    version_number integer NOT NULL,
    status text NOT NULL,
    docx_storage_key text NOT NULL,
    content_hash text NOT NULL,
    metadata_schema jsonb NOT NULL,
    placeholder_schema jsonb NOT NULL,
    author_id text NOT NULL,
    pending_reviewer_role text,
    pending_approver_role text DEFAULT ''::text NOT NULL,
    reviewer_id text,
    approver_id text,
    submitted_at timestamp with time zone,
    reviewed_at timestamp with time zone,
    approved_at timestamp with time zone,
    published_at timestamp with time zone,
    obsoleted_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    lock_version integer DEFAULT 0 NOT NULL
);
```

## Runtime Usage

Use `rg -n "templates_template_version" internal apps` and the owning module wiki to verify readers/writers before changing this table.

## Seed or Reference Data

Check `db/reference-data/0001_product_reference_data.sql` and `db/dev-seeds/0001_local_dev_seed.sql` before adding rows.

## Notes and Debt

Retained in `public` because current runtime/baseline truth still uses it. Do not move schemas without an approved migration plan.
