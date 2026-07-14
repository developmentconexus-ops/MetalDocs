# public.template_versions — RETIRED 2026-07-03 (DB-01)

> **Source:** `db/baseline/0001_current_schema.sql` (history) + `db/migrations/0268_drop_legacy_template_family.sql` (retirement)
> **Schema:** `public`
> **Owner:** templates

## Status: dropped

Dropped by `db/migrations/0268_drop_legacy_template_family.sql`. The canonical, and now sole, template-version store is `public.templates_template_version` (see `wiki/database/tables/templates_template_version.md`) — live since ARC-01 (commit `4160018e`). This page is retained as historical record of the legacy table's prior shape; do not resurrect it for new code.

**Evidence for the drop:** `internal/platform/docgenv2.LegacyTemplateReadCount()` and API/worker log grep both observed zero legacy fallback reads across the full Goal-3 QA run window; the legacy tables held zero rows on a fresh canonical bootstrap at drop time. See `wiki/modules/templates-tech-debt.md` DB-01 (closed) and `wiki/architecture/data-model.md`.

## Purpose (historical)
Former curated-baseline table owned by `templates`, superseded by the canonical `public.templates_template_version` store.

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

None. All application readers/writers moved to `public.templates_template_version` at the ARC-01 canonical-first cutover; the `internal/platform/docgenv2` legacy fallback reader (the last consumer of this table) is deleted in the same change-set that promotes migration 0268 out of `_pending/`.

## Seed or Reference Data

None — `db/dev-seeds/0001_local_dev_seed.sql` seeds only the canonical `templates_template`/`templates_template_version` tables.

## Notes and Debt

Retained here only as historical record. Do not move schemas or resurrect this table without an approved migration plan.
