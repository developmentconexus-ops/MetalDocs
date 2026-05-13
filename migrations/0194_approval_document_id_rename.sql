BEGIN;

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1
    FROM information_schema.columns
    WHERE table_schema = 'public'
      AND table_name = 'approval_instances'
      AND column_name = 'document_v2_id'
  ) THEN
    RAISE EXCEPTION 'expected column not found: public.approval_instances.document_v2_id';
  END IF;

  IF NOT EXISTS (
    SELECT 1
    FROM pg_constraint
    WHERE conname = 'approval_instances_document_v2_id_idempotency_key_key'
      AND conrelid = 'public.approval_instances'::regclass
  ) THEN
    RAISE EXCEPTION 'expected constraint not found: public.approval_instances.approval_instances_document_v2_id_idempotency_key_key';
  END IF;

  IF NOT EXISTS (
    SELECT 1
    FROM pg_class c
    JOIN pg_namespace n ON n.oid = c.relnamespace
    WHERE n.nspname = 'public'
      AND c.relkind = 'i'
      AND c.relname = 'ux_approval_instances_active'
  ) THEN
    RAISE EXCEPTION 'expected index not found: public.ux_approval_instances_active';
  END IF;

  IF NOT EXISTS (
    SELECT 1
    FROM pg_class c
    JOIN pg_namespace n ON n.oid = c.relnamespace
    WHERE n.nspname = 'public'
      AND c.relkind = 'i'
      AND c.relname = 'ix_approval_instances_tenant_doc'
  ) THEN
    RAISE EXCEPTION 'expected index not found: public.ix_approval_instances_tenant_doc';
  END IF;
END $$;

ALTER TABLE approval_instances
  RENAME COLUMN document_v2_id TO document_id;

ALTER TABLE approval_instances
  RENAME CONSTRAINT approval_instances_document_v2_id_idempotency_key_key
  TO approval_instances_document_id_idempotency_key_key;

ALTER INDEX ux_approval_instances_active
  RENAME TO ux_approval_instances_active_document_id;

ALTER INDEX ix_approval_instances_tenant_doc
  RENAME TO ix_approval_instances_tenant_document_id;

CREATE OR REPLACE FUNCTION enforce_signoff_sod() RETURNS trigger AS $$
DECLARE
  author_id TEXT;
BEGIN
  SELECT d.created_by INTO author_id
    FROM public.approval_instances i
    JOIN public.documents d ON d.id = i.document_id
   WHERE i.id = NEW.approval_instance_id;

  IF NEW.actor_user_id = author_id THEN
    RAISE EXCEPTION 'SoD: author cannot sign own revision'
      USING ERRCODE = 'check_violation';
  END IF;

  RETURN NEW;
END;
$$ LANGUAGE plpgsql
   SET search_path = pg_catalog, pg_temp;

COMMIT;
