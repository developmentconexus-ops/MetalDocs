-- db/migrations/0236_dead_schema_drop.sql
-- Drops verified-dead schema after Wave 2 single-mode refactor (FE-5).
-- Re-verified zero production references at Wave 2.13 (2026-06-12):
--   * templates.areas / visibility / specific_areas (CreateTemplateTx stopped
--     writing them in this same commit; domain Template has no backing fields;
--     reference-data 0001_product_reference_data.sql write also removed same commit)
--   * document_profiles.is_active (superseded by archived_at in 0122; only the
--     e2e test seed wrote it — that write removed in this same commit)
--   * document_subjects table (zero Go references since creation)
-- Column drops are CASCADE-safe: no index, FK, or RLS policy references the
-- columns being dropped. The DROP TABLE uses CASCADE to remove the FK constraint
-- fk_documents_subject_code from metaldocs.documents (the column and its index
-- are dead too; left for a separate pass per surgical-change policy).
-- Precedent: 0231_db_hardening_tripwire_and_dead_schema.sql.

BEGIN;

ALTER TABLE public.templates_template DROP COLUMN IF EXISTS areas;
ALTER TABLE public.templates_template DROP COLUMN IF EXISTS visibility;
ALTER TABLE public.templates_template DROP COLUMN IF EXISTS specific_areas;

ALTER TABLE metaldocs.document_profiles DROP COLUMN IF EXISTS is_active;

DROP TABLE IF EXISTS metaldocs.document_subjects CASCADE;

INSERT INTO public.schema_migrations (version, description)
VALUES ('0236', 'dead schema drop (FE-5): templates.areas/visibility/specific_areas, document_profiles.is_active, document_subjects table')
ON CONFLICT (version) DO NOTHING;

COMMIT;
