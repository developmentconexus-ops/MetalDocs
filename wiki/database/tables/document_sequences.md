# metaldocs.document_sequences

> **Source:** `db/baseline/0001_current_schema.sql`
> **Schema:** `metaldocs`
> **Owner:** unknown-owner
> **Last verified:** 2026-07-16 (migration 0308 — table dropped)

## Purpose
**DROPPED** by forward migration `db/migrations/0308_document_profiles_tenant_pk.sql` (ROADMAP unit 4.5, T-015/R-015 half B). This table still appears in the curated baseline snapshot (`db/baseline/0001_current_schema.sql`) below, but that snapshot predates migration 0308; a database migrated forward past 0308 no longer has this table. Zero Go references repo-wide (`grep "document_sequences"` across `internal/` `apps/` finds no hits outside the migration and this baseline DDL dump) — an archive-era dead legacy counter table with no reader, no writer, anywhere. Not to be confused with the live, tenant-scoped `public.cd_sequence_counters` (owned by `controlleddocuments`) — see `wiki/database/tables/cd_sequence_counters.md`. See `wiki/modules/taxonomy-tech-debt.md` T-015 and `wiki/database/tables/document_profiles.md` for the PK-promotion context this drop unblocked.

## Columns

| Column | Type | Nullable | Meaning |
|---|---|---|---|
| `profile_code` | `text` | no | Baseline column. |
| `next_value` | `integer` | no | Baseline column. |

## Baseline Definition

```sql
CREATE TABLE metaldocs.document_sequences (
profile_code text NOT NULL,
    next_value integer NOT NULL,
    CONSTRAINT document_sequences_next_value_check CHECK ((next_value > 0))
);
```

## Runtime Usage

Use `rg -n "document_sequences" internal apps` and the owning module wiki to verify readers/writers before changing this table.

## Seed or Reference Data

Check `db/reference-data/0001_product_reference_data.sql` and `db/dev-seeds/0001_local_dev_seed.sql` before adding rows.

## Notes and Debt

**Dropped 2026-07-16** by `db/migrations/0308_document_profiles_tenant_pk.sql` (ROADMAP unit 4.5), in the same transaction as 3 sibling dead tables, immediately before `document_profiles`' PK was promoted from `code` alone to composite `(tenant_id, code)`. This table's sole inbound FK (`document_sequences_profile_code_fkey` → `document_profiles(code)`) was one of the things keeping the old single-column PK alive; dropping it (and its 3 siblings) let the PK swap proceed with no re-pointing needed. The table remains in the curated baseline snapshot (`db/baseline/0001_current_schema.sql`) as a pre-0308 artifact — do not resurrect it without reversing that decision.
