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

**2026-07-14:** legacy per-version role-routing columns `pending_reviewer_role`/`pending_approver_role` dropped by migration `db/migrations/0306_drop_templates_version_pending_roles.sql` (write-never since ROADMAP unit 3.1a S1/S4, read-never; superseded by the approval kernel route model, ADR 0082 phase c). `db/baseline/0001_current_schema.sql` has not yet been regenerated to reflect the drop — treat the migration, not the baseline dump, as schema truth until it is.

**2026-08-04 ([ADR 0088](../../decisions/0088-template-version-content-always-materialized.md)):** `content_hash` was always `text NOT NULL` at the column level (unchanged above), but its CHECK constraint tightened. Migration `db/migrations/0317_template_version_content_hash_always.sql` replaces the conditional `chk_template_version_content_hash_non_draft` CHECK (`status = 'draft' OR length(content_hash) = 64`) with an unconditional `length(content_hash) = 64` — a content-less version row (empty `content_hash`) can no longer exist in **any** status, not just non-draft. Pre-existing content-less draft rows were purged by the same migration (repair was impossible from SQL alone — it needs object-store access to materialize real bytes). `db/baseline/0001_current_schema.sql` has not yet been regenerated to reflect this constraint change — treat the migration, not the baseline dump above, as schema truth for this CHECK until it is.
