CREATE TABLE IF NOT EXISTS metaldocs.document_types (
  type_key TEXT PRIMARY KEY,
  name TEXT NOT NULL,
  description TEXT NOT NULL DEFAULT '',
  family_key TEXT NOT NULL DEFAULT '',
  active_version INTEGER NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

ALTER TABLE metaldocs.document_types
  ADD COLUMN IF NOT EXISTS type_key TEXT;

ALTER TABLE metaldocs.document_types
  ADD COLUMN IF NOT EXISTS family_key TEXT NOT NULL DEFAULT '';

ALTER TABLE metaldocs.document_types
  ADD COLUMN IF NOT EXISTS active_version INTEGER NOT NULL DEFAULT 1;

UPDATE metaldocs.document_types
SET type_key = COALESCE(NULLIF(type_key, ''), code)
WHERE type_key IS NULL OR type_key = '';

DO $$
BEGIN
  IF EXISTS (
    SELECT 1
    FROM information_schema.columns
    WHERE table_schema = 'metaldocs'
      AND table_name = 'document_types'
      AND column_name = 'family_code'
  ) THEN
    EXECUTE '
      UPDATE metaldocs.document_types
      SET family_key = COALESCE(NULLIF(family_key, ''''), family_code, '''')
      WHERE family_key = ''''';
  ELSE
    EXECUTE '
      UPDATE metaldocs.document_types
      SET family_key = COALESCE(NULLIF(family_key, ''''), '''')
      WHERE family_key = ''''';
  END IF;
END $$;

CREATE UNIQUE INDEX IF NOT EXISTS uq_document_types_type_key
  ON metaldocs.document_types (type_key);

CREATE TABLE IF NOT EXISTS metaldocs.document_type_schema_versions (
  type_key TEXT NOT NULL,
  version INTEGER NOT NULL,
  schema_json JSONB NOT NULL,
  governance_json JSONB NOT NULL DEFAULT '{}'::jsonb,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  PRIMARY KEY (type_key, version)
);

DO $$
BEGIN
  IF EXISTS (
    SELECT 1
    FROM information_schema.columns
    WHERE table_schema = 'metaldocs'
      AND table_name = 'document_types'
      AND column_name = 'type_key'
  ) THEN
    EXECUTE '
      ALTER TABLE metaldocs.document_type_schema_versions
      DROP CONSTRAINT IF EXISTS document_type_schema_versions_type_key_fkey';
    EXECUTE '
      ALTER TABLE metaldocs.document_type_schema_versions
      ADD CONSTRAINT document_type_schema_versions_type_key_fkey
      FOREIGN KEY (type_key) REFERENCES metaldocs.document_types(type_key)';
  END IF;
END $$;

ALTER TABLE metaldocs.documents
  ADD COLUMN IF NOT EXISTS document_type_key TEXT NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS document_type_version INTEGER NOT NULL DEFAULT 1;

UPDATE metaldocs.documents
SET document_type_key = COALESCE(NULLIF(document_type_key, ''), document_type_code)
WHERE document_type_key IS NULL
   OR document_type_key = '';

CREATE INDEX IF NOT EXISTS idx_documents_document_type_key
  ON metaldocs.documents (document_type_key);

ALTER TABLE metaldocs.document_versions
  ADD COLUMN IF NOT EXISTS values_json JSONB NOT NULL DEFAULT '{}'::jsonb;
