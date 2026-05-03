-- 0167_documents_bridge_and_state_columns.sql
-- Repairs two earlier migration bugs:
--   0126 added bridge cols to documents_v2 (wrong table — code writes to documents)
--   0131 failed entirely because its unique index referenced the missing col
-- This migration applies both sets of changes to public.documents.

BEGIN;

-- Idempotent guard (matches schema_migrations insert pattern)
DO $$ BEGIN
  IF EXISTS (SELECT 1 FROM public.schema_migrations WHERE version = '0167') THEN
    RAISE NOTICE '0167 already applied — no-op';
    RETURN;
  END IF;
END $$;

-- 1. Bridge columns (0126 equivalent — correct table)
ALTER TABLE documents
  ADD COLUMN IF NOT EXISTS controlled_document_id UUID
    REFERENCES controlled_documents(id),
  ADD COLUMN IF NOT EXISTS profile_code_snapshot TEXT,
  ADD COLUMN IF NOT EXISTS process_area_code_snapshot TEXT,
  ADD COLUMN IF NOT EXISTS code TEXT;

CREATE INDEX IF NOT EXISTS ix_documents_controlled_doc
  ON documents (controlled_document_id)
  WHERE controlled_document_id IS NOT NULL;

-- 2. State columns (0131 equivalent — transaction was rolled back)
ALTER TABLE documents
  DROP CONSTRAINT IF EXISTS documents_status_check,
  ADD  CONSTRAINT documents_status_check
    CHECK (status IN (
      'draft','finalized','archived',
      'under_review','approved','rejected',
      'scheduled','published','superseded','obsolete'
    ));

ALTER TABLE documents
  ADD COLUMN IF NOT EXISTS effective_from          TIMESTAMPTZ,
  ADD COLUMN IF NOT EXISTS effective_to            TIMESTAMPTZ,
  ADD COLUMN IF NOT EXISTS revision_number         INT NOT NULL DEFAULT 1,
  ADD COLUMN IF NOT EXISTS revision_version        INT NOT NULL DEFAULT 0,
  ADD COLUMN IF NOT EXISTS locked_at               TIMESTAMPTZ,
  ADD COLUMN IF NOT EXISTS content_hash_at_submit  TEXT;

CREATE UNIQUE INDEX IF NOT EXISTS ux_documents_v2_cd_revision
  ON documents (controlled_document_id, revision_number)
  WHERE controlled_document_id IS NOT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS ux_documents_v2_cd_active
  ON documents (controlled_document_id)
  WHERE controlled_document_id IS NOT NULL
    AND status IN ('draft','under_review','approved','rejected','scheduled');

INSERT INTO public.schema_migrations (version, description)
VALUES ('0167', 'add bridge + state columns to documents (repairs 0126/0131 targeting wrong table)')
ON CONFLICT (version) DO NOTHING;

COMMIT;
