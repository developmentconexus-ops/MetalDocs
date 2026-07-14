# public.document_comments

> **Source:** `db/baseline/0001_current_schema.sql`
> **Schema:** `public`
> **Owner:** documents

## Purpose
Current curated-baseline table owned by `documents`. See the owning module wiki and runtime repositories for business behavior.

## Columns

| Column | Type | Nullable | Meaning |
|---|---|---|---|
| `id` | `uuid` | no | Baseline column. |
| `tenant_id` | `uuid` | no | Baseline column. |
| `document_id` | `uuid` | no | Baseline column. |
| `library_comment_id` | `integer` | no | Baseline column. |
| `parent_library_id` | `integer` | yes | Baseline column. |
| `author_id` | `text` | no | Baseline column. |
| `author_display` | `text` | no | Baseline column. |
| `content_json` | `jsonb` | no | Baseline column. |
| `resolved_at` | `timestamp with time zone` | yes | Baseline column. |
| `resolved_by` | `text` | yes | Baseline column. |
| `created_at` | `timestamp with time zone` | no | Baseline column. |
| `updated_at` | `timestamp with time zone` | no | Baseline column. |

## Baseline Definition

```sql
CREATE TABLE public.document_comments (
id uuid DEFAULT gen_random_uuid() NOT NULL,
    tenant_id uuid NOT NULL,
    document_id uuid NOT NULL,
    library_comment_id integer NOT NULL,
    parent_library_id integer,
    author_id text NOT NULL,
    author_display text NOT NULL,
    content_json jsonb NOT NULL,
    resolved_at timestamp with time zone,
    resolved_by text,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);
```

## Runtime Usage

Use `rg -n "document_comments" internal apps` and the owning module wiki to verify readers/writers before changing this table.

## Seed or Reference Data

Check `db/reference-data/0001_product_reference_data.sql` and `db/dev-seeds/0001_local_dev_seed.sql` before adding rows.

## Notes and Debt

Retained in `public` because current runtime/baseline truth still uses it. Do not move schemas without an approved migration plan.
