# metaldocs.document_profile_template_defaults

> **Source:** `db/baseline/0001_current_schema.sql`
> **Schema:** `metaldocs`
> **Owner:** taxonomy
> **Last verified:** 2026-07-16 (migration 0308 — table dropped)

## Purpose
**DROPPED** by forward migration `db/migrations/0308_document_profiles_tenant_pk.sql` (ROADMAP unit 4.5, T-015/R-015 half B). This table still appears in the curated baseline snapshot (`db/baseline/0001_current_schema.sql`) below, but that snapshot predates migration 0308; a database migrated forward past 0308 no longer has this table. Zero readers repo-wide; its sole writer was `apps/api/cmd/metaldocs-e2e-seed/main.go`'s `ensurePODefaultTemplateBinding`, deleted in a sibling unit-4.5 slice — superseded by `document_profiles.default_template_version_id`, the live default-template-binding column (see `wiki/database/tables/document_profiles.md`). See `wiki/modules/taxonomy-tech-debt.md` T-015 and `wiki/backend/binaries/api.md` §7 for the e2e-seed history this drop reflects.

## Columns

| Column | Type | Nullable | Meaning |
|---|---|---|---|
| `profile_code` | `text` | no | Baseline column. |
| `template_key` | `text` | no | Baseline column. |
| `template_version` | `integer` | no | Baseline column. |
| `assigned_at` | `timestamp with time zone` | no | Baseline column. |

## Baseline Definition

```sql
CREATE TABLE metaldocs.document_profile_template_defaults (
profile_code text NOT NULL,
    template_key text NOT NULL,
    template_version integer NOT NULL,
    assigned_at timestamp with time zone DEFAULT now() NOT NULL
);
```

## Runtime Usage

Use `rg -n "document_profile_template_defaults" internal apps` and the owning module wiki to verify readers/writers before changing this table.

## Seed or Reference Data

Check `db/reference-data/0001_product_reference_data.sql` and `db/dev-seeds/0001_local_dev_seed.sql` before adding rows.

## Notes and Debt

**Dropped 2026-07-16** by `db/migrations/0308_document_profiles_tenant_pk.sql` (ROADMAP unit 4.5), in the same transaction as 3 sibling dead tables, immediately before `document_profiles`' PK was promoted from `code` alone to composite `(tenant_id, code)`. Unlike its 3 siblings, this table had **no FK** to `document_profiles` (confirmed by grep against the baseline's FK CONSTRAINT block), so it did not block the PK swap directly — it was dropped in the same pass because its only writer (`ensurePODefaultTemplateBinding`) was being deleted outright by a sibling slice, leaving it permanently unwritten. The table remains in the curated baseline snapshot (`db/baseline/0001_current_schema.sql`) as a pre-0308 artifact — do not resurrect it without reversing that decision.
