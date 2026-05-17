# public.templates_template

> **Source:** `db/baseline/0001_current_schema.sql`
> **Schema:** `public`
> **Owner:** templates

## Purpose
Current curated-baseline table owned by `templates`. See the owning module wiki and runtime repositories for business behavior.

## Columns

| Column | Type | Nullable | Meaning |
|---|---|---|---|
| `id` | `uuid` | no | Baseline column. |
| `tenant_id` | `text` | no | Baseline column. |
| `doc_type_code` | `text` | no | Baseline column. |
| `key` | `text` | no | Baseline column. |
| `name` | `text` | no | Baseline column. |
| `description` | `text` | no | Baseline column. |
| `areas` | `text[]` | no | Baseline column. |
| `visibility` | `text` | no | Baseline column. |
| `specific_areas` | `text[]` | no | Baseline column. |
| `latest_version` | `integer` | no | Baseline column. |
| `published_version_id` | `uuid` | yes | Baseline column. |
| `created_by` | `text` | no | Baseline column. |
| `created_at` | `timestamp with time zone` | no | Baseline column. |
| `archived_at` | `timestamp with time zone` | yes | Baseline column. |
| `system_owned` | `boolean` | no | Baseline column. |

## Baseline Definition

```sql
CREATE TABLE public.templates_template (
id uuid NOT NULL,
    tenant_id text NOT NULL,
    doc_type_code text NOT NULL,
    key text NOT NULL,
    name text NOT NULL,
    description text DEFAULT ''::text NOT NULL,
    areas text[] DEFAULT '{}'::text[] NOT NULL,
    visibility text NOT NULL,
    specific_areas text[] DEFAULT '{}'::text[] NOT NULL,
    latest_version integer DEFAULT 0 NOT NULL,
    published_version_id uuid,
    created_by text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    archived_at timestamp with time zone,
    system_owned boolean DEFAULT false NOT NULL
);
```

## Runtime Usage

Use `rg -n "templates_template" internal apps` and the owning module wiki to verify readers/writers before changing this table.

## Seed or Reference Data

Check `db/reference-data/0001_product_reference_data.sql` and `db/dev-seeds/0001_local_dev_seed.sql` before adding rows.

## Notes and Debt

Retained in `public` because current runtime/baseline truth still uses it. Do not move schemas without an approved migration plan.
