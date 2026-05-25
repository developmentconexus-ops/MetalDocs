-- 0123_taxonomy_extend_process_areas.sql
BEGIN;

ALTER TABLE metaldocs.document_process_areas
  ADD COLUMN IF NOT EXISTS tenant_id UUID NOT NULL DEFAULT 'ffffffff-ffff-ffff-ffff-ffffffffffff';

ALTER TABLE metaldocs.document_process_areas
  ADD COLUMN IF NOT EXISTS parent_code TEXT;

ALTER TABLE metaldocs.document_process_areas
  ADD COLUMN IF NOT EXISTS owner_user_id TEXT;

ALTER TABLE metaldocs.document_process_areas
  ADD COLUMN IF NOT EXISTS default_approver_role TEXT;

ALTER TABLE metaldocs.document_process_areas
  ADD COLUMN IF NOT EXISTS archived_at TIMESTAMPTZ;

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1
    FROM pg_indexes
    WHERE schemaname = 'metaldocs'
      AND indexname = 'ux_process_areas_tenant_code'
  ) THEN
    CREATE UNIQUE INDEX ux_process_areas_tenant_code
      ON metaldocs.document_process_areas (tenant_id, code);
  END IF;
END $$;

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1
    FROM pg_constraint
    WHERE conname = 'fk_area_parent_tenant'
      AND conrelid = 'metaldocs.document_process_areas'::regclass
  ) THEN
    ALTER TABLE metaldocs.document_process_areas
      ADD CONSTRAINT fk_area_parent_tenant
      FOREIGN KEY (tenant_id, parent_code)
      REFERENCES metaldocs.document_process_areas (tenant_id, code);
  END IF;
END $$;

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1
    FROM pg_constraint
    WHERE conname = 'area_code_format'
      AND conrelid = 'metaldocs.document_process_areas'::regclass
  ) THEN
    ALTER TABLE metaldocs.document_process_areas
      ADD CONSTRAINT area_code_format
      CHECK (code ~ '^[a-z][a-z0-9_-]{1,63}$');
  END IF;
END $$;

-- idempotent: safe to re-create, identical to definition in 0122
CREATE OR REPLACE FUNCTION reject_code_update() RETURNS trigger AS $$
BEGIN
  IF NEW.code IS DISTINCT FROM OLD.code THEN
    RAISE EXCEPTION 'code column is immutable (table=%, old=%, new=%)',
      TG_TABLE_NAME, OLD.code, NEW.code
      USING ERRCODE = 'check_violation';
  END IF;
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_process_areas_code_immutable ON metaldocs.document_process_areas;
CREATE TRIGGER trg_process_areas_code_immutable
  BEFORE UPDATE ON metaldocs.document_process_areas
  FOR EACH ROW EXECUTE FUNCTION reject_code_update();

COMMIT;
