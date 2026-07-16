-- 0308_document_profiles_tenant_pk.sql
-- ROADMAP unit 4.5 -- T-015/R-015 deferred HALF B (wiki/modules/taxonomy-tech-debt.md
-- T-015; wiki/backlog/taxonomy-refactor.md R-015): promote
-- metaldocs.document_profiles' PK from (code) to (tenant_id, code), the
-- document_profiles half that 0264 (HALF A, document_process_areas)
-- deliberately deferred with no DDL. See 0264's header for why the two
-- tables were split.
--
-- ============================================================================
-- DEAD-TABLE DROPS (evidence, against db/baseline/0001_current_schema.sql,
-- current as of this migration's authoring)
-- ============================================================================
-- Four profile-adjacent tables are dropped in the same transaction, before
-- the PK swap, per hub ruling ("same-migration-or-never" -- these tables'
-- only remaining purpose was as document_profiles(code) FK children, and
-- carrying them through the PK swap as-is would require giving them a
-- tenant_id column they have never had):
--
--   - metaldocs.document_profile_governance (PK profile_code; FK
--     profile_code -> document_profiles(code)): archive-era 0027/0040s
--     table. Zero Go references repo-wide (grep "document_profile_governance"
--     across internal/ apps/ tests/ -- zero hits outside this migration and
--     the baseline DDL dump). No reader, no writer, anywhere.
--   - metaldocs.document_profile_schema_versions (PK (profile_code,
--     version); FK profile_code -> document_profiles(code)): same era, same
--     verdict -- zero Go references repo-wide.
--   - metaldocs.document_sequences (PK profile_code; FK profile_code ->
--     document_profiles(code)): same era, same verdict -- zero Go
--     references repo-wide. Not to be confused with
--     public.cd_sequence_counters (live, tenant-scoped, owned by the
--     controlleddocuments module) or metaldocs.document_profile_schema_versions
--     above -- this is a distinct dead legacy counter table.
--   - metaldocs.document_profile_template_defaults (PK profile_code; NO FK
--     to document_profiles despite the profile_code column -- confirmed by
--     grep against the baseline's FK CONSTRAINT block, zero
--     "document_profile_template_defaults" hits): zero READERS repo-wide.
--     Its sole WRITER is apps/api/cmd/metaldocs-e2e-seed/main.go's
--     ensurePODefaultTemplateBinding, which a sibling slice (unit 4.5 track
--     B) is deleting outright -- superseded by
--     document_profiles.default_template_version_id (baseline line ~1125),
--     the live default-template-binding column. Hub-ratified: drop here
--     rather than carry a soon-to-be-unwritten table through the PK swap.
--
-- None of the four has any INBOUND foreign key from any other table (grep
-- "REFERENCES metaldocs\.(document_profile_governance|
-- document_profile_schema_versions|document_sequences|
-- document_profile_template_defaults)" across the baseline: zero hits), so a
-- flat IF EXISTS ... CASCADE list is safe -- no drop-order dependency among
-- them, and CASCADE only ever needs to clean up each table's own outbound FK
-- to document_profiles(code), which DROP TABLE removes for free.
--
-- ============================================================================
-- PK SWAP MECHANICS -- FK-dependency finding that changes the originally
-- planned drop/recreate sequence
-- ============================================================================
-- Before writing this migration, the full FK inventory was re-run against
-- the CURRENT baseline (grep "REFERENCES metaldocs\.document_profiles"):
-- after the four dead-table drops above, the ONLY remaining reference to
-- document_profiles(code) singularly was metaldocs.template_drafts'
-- profile_code FK -- and that table was already dropped by archived 0260
-- (DB-05, template_drafts CASCADE), so document_profiles_pkey (on code
-- alone) has ZERO inbound FKs by the time this migration runs. The three
-- composite FKs this migration must preserve --
-- approval_routes_document_profile_fk (public.approval_routes),
-- cd_sequence_counters_tenant_id_profile_code_fkey
-- (public.cd_sequence_counters), and
-- controlled_documents_tenant_id_profile_code_fkey
-- (public.controlled_documents) -- are ALL already bound to
-- ux_document_profiles_tenant_code (tenant_id, code), the existing unique
-- index, not to document_profiles_pkey.
--
-- This means the plain single-column PK can be dropped outright (nothing
-- depends on it), and the exact 0264 technique applies cleanly: PROMOTE the
-- existing ux_document_profiles_tenant_code unique index into the new PK via
-- ADD CONSTRAINT ... PRIMARY KEY USING INDEX. USING INDEX keeps the index
-- OID, so the three composite FKs (bound to that index via
-- pg_constraint.conindid) stay valid and untouched throughout -- no DROP
-- CONSTRAINT / re-ADD CONSTRAINT round-trip needed for any of them, and no
-- separate "drop the now-redundant unique index" step either: Postgres
-- renames the index to document_profiles_pkey as part of USING INDEX, so
-- ux_document_profiles_tenant_code stops existing under that name and there
-- is no leftover duplicate index. End state is identical to a fresh
-- composite-PK creation with the three FKs re-pointed -- just reached without
-- ever invalidating them.
--
-- Both tenant_id and code are already NOT NULL in the baseline (line
-- ~1117/1124), a precondition PRIMARY KEY USING INDEX enforces.
--
-- Idempotency: both ALTER TABLE statements are wrapped in DO-block guards
-- (pg_constraint / pg_class introspection) so a re-run -- including a re-run
-- against a FUTURE baseline that already bakes the composite PK in directly,
-- with no ux_document_profiles_tenant_code index left to promote -- is a
-- clean no-op. Guard 1 only fires while document_profiles_pkey exists AND is
-- backed by exactly (code); guard 2 only fires while document_profiles has
-- no primary key AND ux_document_profiles_tenant_code still exists as a
-- distinct index.
--
-- Safety class: ACCESS EXCLUSIVE lock on metaldocs.document_profiles for the
-- duration (DROP CONSTRAINT + ADD CONSTRAINT PRIMARY KEY both take it);
-- document_profiles is reference/taxonomy data, not a hot high-volume table,
-- so the lock window is expected to be brief.
--
-- RLS: FORCE ROW LEVEL SECURITY + the tenant_isolation USING policy on
-- document_profiles key off the tenant_id column value, not the PK
-- constraint shape -- unaffected by this migration.
--
-- Rollback note (manual, no down migrations in this repo's convention): drop
-- the composite PK, recreate the single-column PK on (code), and recreate
-- ux_document_profiles_tenant_code as a plain unique index on (tenant_id,
-- code). Only valid if no two tenants have adopted the same profile code
-- since this migration applied -- verify with
-- `SELECT code, count(DISTINCT tenant_id) FROM metaldocs.document_profiles
--  GROUP BY code HAVING count(DISTINCT tenant_id) > 1` before rolling back.

BEGIN;

DROP TABLE IF EXISTS metaldocs.document_profile_governance CASCADE;
DROP TABLE IF EXISTS metaldocs.document_profile_schema_versions CASCADE;
DROP TABLE IF EXISTS metaldocs.document_sequences CASCADE;
DROP TABLE IF EXISTS metaldocs.document_profile_template_defaults CASCADE;

DO $$
DECLARE
  pk_cols text;
BEGIN
  SELECT string_agg(a.attname, ',' ORDER BY array_position(c.conkey, a.attnum))
    INTO pk_cols
    FROM pg_constraint c
    JOIN pg_attribute a ON a.attrelid = c.conrelid AND a.attnum = ANY(c.conkey)
   WHERE c.conrelid = 'metaldocs.document_profiles'::regclass
     AND c.contype = 'p';

  IF pk_cols = 'code' THEN
    EXECUTE 'ALTER TABLE ONLY metaldocs.document_profiles DROP CONSTRAINT document_profiles_pkey';
  END IF;
END
$$;

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_constraint
     WHERE conrelid = 'metaldocs.document_profiles'::regclass
       AND contype = 'p'
  ) AND EXISTS (
    SELECT 1 FROM pg_class rc
    JOIN pg_namespace n ON n.oid = rc.relnamespace
     WHERE rc.relname = 'ux_document_profiles_tenant_code'
       AND n.nspname = 'metaldocs'
  ) THEN
    EXECUTE 'ALTER TABLE ONLY metaldocs.document_profiles ADD CONSTRAINT document_profiles_pkey PRIMARY KEY USING INDEX ux_document_profiles_tenant_code';
  END IF;
END
$$;

INSERT INTO public.schema_migrations (version, description)
VALUES ('0308', 'ROADMAP unit 4.5 (T-015/R-015 half B): drop 4 dead document_profiles-adjacent tables (document_profile_governance, document_profile_schema_versions, document_sequences -- zero Go refs, archive-era 0027/0040s; document_profile_template_defaults -- zero readers, sole writer being deleted in sibling slice, superseded by document_profiles.default_template_version_id), then promote metaldocs.document_profiles PK from (code) to (tenant_id, code) via PRIMARY KEY USING INDEX ux_document_profiles_tenant_code (0264 technique -- keeps the index OID so the 3 inbound composite FKs bound to it, approval_routes_document_profile_fk / cd_sequence_counters_tenant_id_profile_code_fkey / controlled_documents_tenant_id_profile_code_fkey, stay valid untouched; index renamed to document_profiles_pkey, no leftover redundant index).')
ON CONFLICT (version) DO NOTHING;

COMMIT;
