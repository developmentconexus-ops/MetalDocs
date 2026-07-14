# public.templates — RETIRED 2026-07-03 (DB-01)

> **Source:** `db/baseline/0001_current_schema.sql` (history) + `db/migrations/0268_drop_legacy_template_family.sql` (retirement)
> **Schema:** `public`
> **Owner:** templates

## Status: dropped

Dropped by `db/migrations/0268_drop_legacy_template_family.sql`. The canonical, and now sole, template store is `public.templates_template` (see `wiki/database/tables/templates_template.md`) — live since ARC-01 (commit `4160018e`). This page is retained as historical record of the legacy table's prior shape; do not resurrect it for new code.

**Evidence for the drop:** `internal/platform/docgenv2.LegacyTemplateReadCount()` and API/worker log grep both observed zero legacy fallback reads across the full Goal-3 QA run window; the legacy tables held zero rows on a fresh canonical bootstrap at drop time. See `wiki/modules/templates-tech-debt.md` DB-01 (closed) and `wiki/architecture/data-model.md`.

## Purpose (historical)
Former curated-baseline table owned by `templates`, superseded by the canonical `public.templates_template` store.

## Columns

| Column | Type | Nullable | Meaning |
|---|---|---|---|
| `id` | `uuid` | no | Baseline column. |
| `tenant_id` | `uuid` | no | Baseline column. |
| `key` | `text` | no | Baseline column. |
| `name` | `text` | no | Baseline column. |
| `description` | `text` | yes | Baseline column. |
| `current_published_version_id` | `uuid` | yes | Baseline column. |
| `created_at` | `timestamp with time zone` | no | Baseline column. |
| `updated_at` | `timestamp with time zone` | no | Baseline column. |
| `created_by` | `text` | no | Baseline column. |

## Baseline Definition

```sql
CREATE TABLE public.templates (
id uuid DEFAULT gen_random_uuid() NOT NULL,
    tenant_id uuid NOT NULL,
    key text NOT NULL,
    name text NOT NULL,
    description text,
    current_published_version_id uuid,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    created_by text NOT NULL
);
```

## Runtime Usage

None. All application readers/writers moved to `public.templates_template` at the ARC-01 canonical-first cutover; the `internal/platform/docgenv2` legacy fallback reader (the last consumer of this table) is deleted in the same change-set that promotes migration 0268 out of `_pending/`.

## Seed or Reference Data

None — `db/dev-seeds/0001_local_dev_seed.sql` seeds only the canonical `templates_template`/`templates_template_version` tables.

## Notes and Debt

Retained here only as historical record. Do not move schemas or resurrect this table without an approved migration plan.
