# public.controlled_documents

> **Source:** `db/baseline/0001_current_schema.sql`
> **Schema:** `public`
> **Owner:** registry

## Purpose
Current curated-baseline table owned by `registry`. See the owning module wiki and runtime repositories for business behavior.

## Columns

| Column | Type | Nullable | Meaning |
|---|---|---|---|
| `id` | `uuid` | no | Baseline column. |
| `tenant_id` | `uuid` | no | Baseline column. |
| `profile_code` | `text` | no | Baseline column. |
| `process_area_code` | `text` | no | Baseline column. |
| `department_code` | `text` | yes | Baseline column. |
| `code` | `text` | no | Baseline column. |
| `sequence_num` | `integer` | yes | Baseline column. |
| `title` | `text` | no | Baseline column. |
| `owner_user_id` | `text` | no | Baseline column. |
| `override_template_version_id` | `uuid` | yes | Baseline column. |
| `status` | `text` | no | Baseline column. |
| `created_at` | `timestamp with time zone` | no | Baseline column. |
| `updated_at` | `timestamp with time zone` | no | Baseline column. |
| `visibility_scope` | `text` | no | Baseline column. |

## Baseline Definition

```sql
CREATE TABLE public.controlled_documents (
id uuid DEFAULT gen_random_uuid() NOT NULL,
    tenant_id uuid NOT NULL,
    profile_code text NOT NULL,
    process_area_code text NOT NULL,
    department_code text,
    code text NOT NULL,
    sequence_num integer,
    title text NOT NULL,
    owner_user_id text NOT NULL,
    override_template_version_id uuid,
    status text DEFAULT 'active'::text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    visibility_scope text DEFAULT 'company'::text NOT NULL,
    CONSTRAINT controlled_document_code_format CHECK (((length(code) >= 2) AND (length(code) <= 100))),
    CONSTRAINT controlled_documents_status_check CHECK ((status = ANY (ARRAY['active'::text, 'obsolete'::text, 'superseded'::text]))),
    CONSTRAINT controlled_documents_visibility_scope_check CHECK ((visibility_scope = ANY (ARRAY['company'::text, 'restricted'::text])))
);
```

## Runtime Usage

Use `rg -n "controlled_documents" internal apps` and the owning module wiki to verify readers/writers before changing this table.

## Seed or Reference Data

Check `db/reference-data/0001_product_reference_data.sql` and `db/dev-seeds/0001_local_dev_seed.sql` before adding rows.

## Notes and Debt

Retained in `public` because current runtime/baseline truth still uses it. Do not move schemas without an approved migration plan.
