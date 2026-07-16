# metaldocs.document_profiles

> **Source:** `db/baseline/0001_current_schema.sql`
> **Schema:** `metaldocs`
> **Owner:** taxonomy
> **Last verified:** 2026-07-16 (migration 0308 — PK promoted from `code` alone to composite `(tenant_id, code)`; baseline file unchanged, see Post-baseline note below) | **Prior:** 2026-06-12 (migration 0236 — `is_active` dropped as dead schema, superseded by `archived_at`; baseline file unchanged, the column is removed by the forward migration)

## Purpose
Current curated-baseline table owned by `taxonomy`. See the owning module wiki and runtime repositories for business behavior.

## Columns

| Column | Type | Nullable | Meaning |
|---|---|---|---|
| `code` | `text` | no | Baseline column. |
| `family_code` | `text` | no | Baseline column. |
| `name` | `text` | no | Baseline column. |
| `description` | `text` | no | Baseline column. |
| `review_interval_days` | `integer` | no | Baseline column. |
| `created_at` | `timestamp with time zone` | no | Baseline column. |
| `alias` | `text` | no | Baseline column. |
| `tenant_id` | `uuid` | no | Baseline column. |
| `default_template_version_id` | `uuid` | yes | Baseline column. |
| `owner_user_id` | `text` | yes | Baseline column. |
| `editable_by_role` | `text` | no | Baseline column. |
| `archived_at` | `timestamp with time zone` | yes | Baseline column. |

## Baseline Definition

```sql
CREATE TABLE metaldocs.document_profiles (
code text NOT NULL,
    family_code text NOT NULL,
    name text NOT NULL,
    description text DEFAULT ''::text NOT NULL,
    review_interval_days integer NOT NULL,
    is_active boolean DEFAULT true NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    alias text NOT NULL,
    tenant_id uuid DEFAULT 'ffffffff-ffff-ffff-ffff-ffffffffffff'::uuid NOT NULL,
    default_template_version_id uuid,
    owner_user_id text,
    editable_by_role text DEFAULT 'admin'::text NOT NULL,
    archived_at timestamp with time zone,
    CONSTRAINT chk_document_profiles_alias_length CHECK (((char_length(alias) >= 1) AND (char_length(alias) <= 24))),
    CONSTRAINT document_profiles_review_interval_days_check CHECK ((review_interval_days > 0)),
    CONSTRAINT profile_code_format CHECK ((code ~ '^[a-z][a-z0-9_-]{1,63}$'::text))
);
```

**Post-baseline note (migration `0308`, 2026-07-16, ROADMAP unit 4.5, T-015/R-015 half B):** the PRIMARY KEY was promoted from `code` alone to composite `(tenant_id, code)`. The previously separate `ux_document_profiles_tenant_code` unique index on `(tenant_id, code)` was **promoted into the PK** via `ADD CONSTRAINT ... PRIMARY KEY USING INDEX` (keeps the index OID, so the 3 inbound composite FKs bound to it — `approval_routes_document_profile_fk`, `cd_sequence_counters_tenant_id_profile_code_fkey`, `controlled_documents_tenant_id_profile_code_fkey` — stayed valid throughout, no re-pointing needed); the index no longer exists under its old name, renamed to `document_profiles_pkey`. This migration also dropped the 4 tables that previously blocked the promotion (see Notes and Debt) in the same transaction, before the PK swap. This closes the `document_profiles` half of T-015 that the design note below had deferred.

## Runtime Usage

Use `rg -n "document_profiles" internal apps` and the owning module wiki to verify readers/writers before changing this table.

## Seed or Reference Data

Check `db/reference-data/0001_product_reference_data.sql` and `db/dev-seeds/0001_local_dev_seed.sql` before adding rows.

## Notes and Debt

Curated baseline table in the `metaldocs` schema.

**PK promoted (2026-07-16, migration `0308`, ROADMAP unit 4.5, T-015/R-015 half B):** PK is now composite `(tenant_id, code)` — see the Post-baseline note above. This supersedes the 2026-07-02 "PK left as-is" ruling below, which is retained for historical context.

**Historical — PK left as-is (2026-07-02, DB-08/T-015 audit; RESOLVED 2026-07-16 by migration 0308):** PK previously remained `code` alone (cross-tenant unique), NOT promoted to `(tenant_id, code)` despite the redundant-looking `(tenant_id, code)` unique index (`ux_document_profiles_tenant_code`). Reason at the time: `document_profiles(code)` had 4 inbound single-column FKs (`document_profile_governance`, `document_profile_schema_versions`, `document_sequences`, and formerly `template_drafts`) from tables that had **no `tenant_id` column and no RLS** — global registries keyed purely by `profile_code`, not per-tenant data. Migration 0308 closed this by dropping all 4 dead tables (zero readers/writers repo-wide — see `wiki/database/tables/document_profile_governance.md`, `document_profile_schema_versions.md`, `document_sequences.md`, `document_profile_template_defaults.md` for each table's retirement record) and then promoting the PK, since `template_drafts` had already been dropped by migration 0260. See `wiki/modules/taxonomy-tech-debt.md` T-015 for the closure record. Contrast with `document_process_areas` (same T-015 finding), which was promoted earlier — see `wiki/database/tables/document_process_areas.md`.
