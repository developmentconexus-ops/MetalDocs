-- 0183_documents_name_not_empty.sql
-- Enforce documents.name as a non-empty string.
-- Backfills any existing empty/null name from the linked controlled_documents.title
-- before tightening the column with NOT NULL + CHECK length(trim(name)) > 0.
-- Revert path: drop the CHECK constraint and (if needed) re-allow NULL via
--   ALTER TABLE documents DROP CONSTRAINT documents_name_not_empty;
--   ALTER TABLE documents ALTER COLUMN name DROP NOT NULL;

BEGIN;

DO $$ BEGIN
  IF EXISTS (SELECT 1 FROM public.schema_migrations WHERE version = '0183') THEN
    RAISE NOTICE '0183 already applied — no-op';
    RETURN;
  END IF;
END $$;

-- Backfill any orphan empty/null name from CD title (must run before constraint).
UPDATE documents d
SET name = cd.title
FROM controlled_documents cd
WHERE d.controlled_document_id = cd.id
  AND (d.name IS NULL OR length(trim(d.name)) = 0);

-- Then enforce: NOT NULL is already declared in 0110, but we re-assert it for
-- defense-in-depth and add a CHECK to forbid whitespace-only values.
ALTER TABLE documents
  ALTER COLUMN name SET NOT NULL,
  ADD CONSTRAINT documents_name_not_empty CHECK (length(trim(name)) > 0);

INSERT INTO public.schema_migrations (version, description)
VALUES ('0183', 'documents.name NOT NULL + CHECK length>0 + backfill from CD title')
ON CONFLICT (version) DO NOTHING;

COMMIT;
