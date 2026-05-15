# metaldocs.template_drafts

> **Source:** `db/baseline/0001_current_schema.sql`
> **Schema:** `metaldocs`
> **Owner:** templates

## Purpose
Current curated-baseline table owned by `templates`. See the owning module wiki and runtime repositories for business behavior.

## Columns

| Column | Type | Nullable | Meaning |
|---|---|---|---|
| `template_key` | `text` | no | Baseline column. |
| `profile_code` | `text` | no | Baseline column. |
| `base_version` | `integer` | no | Baseline column. |
| `name` | `text` | no | Baseline column. |
| `theme_json` | `jsonb` | no | Baseline column. |
| `meta_json` | `jsonb` | no | Baseline column. |
| `blocks_json` | `jsonb` | no | Baseline column. |
| `lock_version` | `integer` | no | Baseline column. |
| `has_stripped_fields` | `boolean` | no | Baseline column. |
| `stripped_fields_json` | `jsonb` | yes | Baseline column. |
| `created_by` | `text` | no | Baseline column. |
| `created_at` | `timestamp with time zone` | no | Baseline column. |
| `updated_at` | `timestamp with time zone` | no | Baseline column. |
| `published_html` | `text` | yes | Baseline column. |
| `draft_status` | `text` | no | Baseline column. |

## Baseline Definition

```sql
CREATE TABLE metaldocs.template_drafts (
template_key text NOT NULL,
    profile_code text NOT NULL,
    base_version integer DEFAULT 0 NOT NULL,
    name text NOT NULL,
    theme_json jsonb DEFAULT '{}'::jsonb NOT NULL,
    meta_json jsonb DEFAULT '{}'::jsonb NOT NULL,
    blocks_json jsonb NOT NULL,
    lock_version integer DEFAULT 1 NOT NULL,
    has_stripped_fields boolean DEFAULT false NOT NULL,
    stripped_fields_json jsonb,
    created_by text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    published_html text,
    draft_status text DEFAULT 'draft'::text NOT NULL,
    CONSTRAINT template_drafts_draft_status_check CHECK ((draft_status = ANY (ARRAY['draft'::text, 'pending_review'::text, 'published'::text])))
);
```

## Runtime Usage

Use `rg -n "template_drafts" internal apps` and the owning module wiki to verify readers/writers before changing this table.

## Seed or Reference Data

Check `db/reference-data/0001_product_reference_data.sql` and `db/dev-seeds/0001_local_dev_seed.sql` before adding rows.

## Notes and Debt

Curated baseline table in the `metaldocs` schema.
