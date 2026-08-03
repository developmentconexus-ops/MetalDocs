-- 0264_process_areas_composite_pk.sql
-- DB-08 (P3) — HALF A: promote metaldocs.document_process_areas PK from
-- (code) to (tenant_id, code), retiring the now-redundant
-- ux_process_areas_tenant_code unique index (wiki/modules/taxonomy-tech-debt.md
-- T-015; wiki/backlog/taxonomy-refactor.md R-015).
--
-- Scope note: R-015/T-015 names BOTH document_profiles and
-- document_process_areas together, but the FK inventory below shows they are
-- architecturally different — this migration does document_process_areas
-- only. document_profiles is deliberately deferred with NO DDL written; see
-- the design note appended to wiki/modules/taxonomy-tech-debt.md T-015 for
-- why (also mirrored on wiki/database/tables/document_profiles.md). This
-- repo's convention is docs-only for a deferred half, not a placeholder
-- migration file — 0265 in this batch is used for an unrelated finding
-- (documents_status_check tightening), not a document_profiles design note.
--
-- FK/dependency inventory performed before writing this migration (grep
-- against db/baseline/0001_current_schema.sql for
-- "REFERENCES metaldocs.document_process_areas"):
--   - fk_area_parent_tenant (document_process_areas.parent_code self-FK)
--   - cd_sequence_counters_tenant_id_process_area_code_fkey
--   - controlled_document_area_grants_tenant_id_area_code_fkey
--   - controlled_documents_tenant_id_process_area_code_fkey
--   - user_process_areas_tenant_id_area_code_fkey
--   ALL FIVE inbound FKs are already composite (tenant_id, area_code) /
--   (tenant_id, parent_code) referencing the EXISTING
--   ux_process_areas_tenant_code (tenant_id, code) unique index — zero
--   single-column FKs reference document_process_areas(code) alone. Every
--   dependent table already carries tenant_id. Application code
--   (internal/modules/taxonomy/infrastructure/repository.go AreaRepository)
--   already reads/writes with tenant_id + code together in every query — no
--   code assumes cross-tenant single-column code uniqueness. This table is
--   therefore a clean composite-PK promotion with no re-pointing required.
--
-- RLS: FORCE ROW LEVEL SECURITY + the tenant_isolation USING policy on
-- document_process_areas key off the tenant_id column value, not the PK
-- constraint shape — swapping the PK does not touch RLS policy definitions
-- and the policy continues to apply identically after this migration.
--
-- Mechanics: drop document_process_areas_pkey (code), then PROMOTE the
-- existing ux_process_areas_tenant_code unique index into the new PK via
-- ADD CONSTRAINT ... PRIMARY KEY USING INDEX. A plain DROP INDEX +
-- ADD PRIMARY KEY sequence is NOT applyable: all five inbound composite FKs
-- are bound to that index (pg_constraint.conindid), so DROP INDEX fails with
-- SQLSTATE 2BP01 ("other objects depend on it") — proven on a fresh
-- baseline replay 2026-07-03. USING INDEX keeps the index OID (FK bindings
-- survive untouched) and Postgres renames the index to the constraint name,
-- so the end state is identical to a fresh composite-PK creation: PK
-- document_process_areas_pkey on (tenant_id, code), no redundant second
-- index, no gap in query-plan support for tenant-filtered code lookups.
-- Both columns are already NOT NULL in the baseline (0001 line 1050/1055).
--
-- Safety class: ACCESS EXCLUSIVE lock on metaldocs.document_process_areas
-- for the duration of the migration (DROP CONSTRAINT + ADD CONSTRAINT
-- PRIMARY KEY both take ACCESS EXCLUSIVE; table is small — process-area
-- taxonomy rows are low-cardinality reference data, not a hot high-volume
-- table — so the lock window is expected to be sub-second in practice, but
-- callers should still expect brief write blocking during deploy).
--
-- Rollback note: re-run in reverse — drop the composite PK, recreate the
-- single-column PK on (code), and recreate ux_process_areas_tenant_code
-- unique index on (tenant_id, code). Not scripted here (no automatic down
-- migrations in this repo's convention per db/migrations/README.md); if
-- rollback is ever needed, mirror this file's DDL with old/new swapped.
-- Pre-check: rollback to a single-column code PK is only valid if no two
-- tenants have adopted the same process-area code since this migration
-- applied (unlikely given the table is reference data, but not physically
-- prevented once the composite PK is live) — verify with
-- `SELECT code, count(DISTINCT tenant_id) FROM metaldocs.document_process_areas
--  GROUP BY code HAVING count(DISTINCT tenant_id) > 1` before rolling back.

BEGIN;

ALTER TABLE ONLY metaldocs.document_process_areas
    DROP CONSTRAINT document_process_areas_pkey;

ALTER TABLE ONLY metaldocs.document_process_areas
    ADD CONSTRAINT document_process_areas_pkey PRIMARY KEY USING INDEX ux_process_areas_tenant_code;

INSERT INTO public.schema_migrations (version, description)
VALUES ('0264', 'DB-08 (T-015/R-015) half A: promote metaldocs.document_process_areas PK from (code) to (tenant_id, code) by converting the ux_process_areas_tenant_code unique index into the PK''s backing index (PRIMARY KEY USING INDEX — keeps the index OID so the 5 inbound composite FKs bound to it stay valid; index renamed to document_process_areas_pkey). document_profiles half deferred (no DDL), see wiki/modules/taxonomy-tech-debt.md T-015 design note.')
ON CONFLICT (version) DO NOTHING;

COMMIT;
