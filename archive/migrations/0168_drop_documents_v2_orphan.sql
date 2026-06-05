-- 0168_drop_documents_v2_orphan.sql
-- public.documents_v2 was the W1 scaffold table created in migration 0103.
-- Migration 0110 replaced it with public.documents (W3 schema).
-- No Go code writes to or reads from documents_v2. Always 0 rows.
-- Migrations 0126/0129 mistakenly added bridge columns to this table instead
-- of documents; that was corrected in 0167.

BEGIN;

DROP TABLE IF EXISTS public.documents_v2;

INSERT INTO public.schema_migrations (version, description)
VALUES ('0168', 'drop orphaned public.documents_v2 table (W1 scaffold, replaced by documents in 0110)')
ON CONFLICT (version) DO NOTHING;

COMMIT;
