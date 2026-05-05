-- 0181_drop_documents_locked_at.sql
-- Drops orphaned documents.locked_at column — zero references in apps/api/,
-- internal/, or frontend/ confirmed before this migration was written.
-- Rollback: ALTER TABLE public.documents ADD COLUMN locked_at TIMESTAMPTZ NULL;

BEGIN;

ALTER TABLE public.documents DROP COLUMN IF EXISTS locked_at;

INSERT INTO public.schema_migrations (version, description)
VALUES ('0181', 'drop orphaned documents.locked_at column')
ON CONFLICT (version) DO NOTHING;

COMMIT;
