# metaldocs.document_template_versions

> **Source:** `db/baseline/0001_current_schema.sql`
> **Schema:** `metaldocs`
> **Owner:** templates

## Purpose
Current curated-baseline table owned by `templates`. See the owning module wiki and runtime repositories for business behavior.

## Columns

| Column | Type | Nullable | Meaning |
|---|---|---|---|
| `template_key` | `text` | no | Baseline column. |
| `version` | `integer` | no | Baseline column. |
| `profile_code` | `text` | no | Baseline column. |
| `schema_version` | `integer` | no | Baseline column. |
| `name` | `text` | no | Baseline column. |
| `definition_json` | `jsonb` | no | Baseline column. |
| `created_at` | `timestamp with time zone` | no | Baseline column. |
| `editor` | `text` | no | Baseline column. |
| `content_format` | `text` | no | Baseline column. |
| `body_html` | `text` | no | Baseline column. |
| `export_config` | `jsonb` | yes | Baseline column. |
| `status` | `text` | no | Baseline column. |

## Baseline Definition

```sql
CREATE TABLE metaldocs.document_template_versions (
template_key text NOT NULL,
    version integer NOT NULL,
    profile_code text NOT NULL,
    schema_version integer NOT NULL,
    name text NOT NULL,
    definition_json jsonb NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    editor text DEFAULT 'ckeditor5'::text NOT NULL,
    content_format text DEFAULT 'html'::text NOT NULL,
    body_html text DEFAULT ''::text NOT NULL,
    export_config jsonb,
    status text DEFAULT 'published'::text NOT NULL
);
```

## Runtime Usage

Use `rg -n "document_template_versions" internal apps` and the owning module wiki to verify readers/writers before changing this table.

## Seed or Reference Data

Check `db/reference-data/0001_product_reference_data.sql` and `db/dev-seeds/0001_local_dev_seed.sql` before adding rows.

## Notes and Debt

Curated baseline table in the `metaldocs` schema.
