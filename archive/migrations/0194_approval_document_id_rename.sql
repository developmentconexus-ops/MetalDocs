BEGIN;

DO $$
BEGIN
  IF EXISTS (
    SELECT 1
    FROM information_schema.columns
    WHERE table_schema = 'public'
      AND table_name = 'approval_instances'
      AND column_name = 'document_v2_id'
  ) THEN
    ALTER TABLE public.approval_instances
      RENAME COLUMN document_v2_id TO document_id;
  ELSIF NOT EXISTS (
    SELECT 1
    FROM information_schema.columns
    WHERE table_schema = 'public'
      AND table_name = 'approval_instances'
      AND column_name = 'document_id'
  ) THEN
    RAISE EXCEPTION 'expected column not found: public.approval_instances.document_v2_id or document_id';
  END IF;

  IF EXISTS (
    SELECT 1
    FROM pg_constraint
    WHERE conname = 'approval_instances_document_v2_id_idempotency_key_key'
      AND conrelid = 'public.approval_instances'::regclass
  ) THEN
    ALTER TABLE public.approval_instances
      RENAME CONSTRAINT approval_instances_document_v2_id_idempotency_key_key
      TO approval_instances_document_id_idempotency_key_key;
  ELSIF NOT EXISTS (
    SELECT 1
    FROM pg_constraint
    WHERE conname = 'approval_instances_document_id_idempotency_key_key'
      AND conrelid = 'public.approval_instances'::regclass
  ) THEN
    RAISE EXCEPTION 'expected unique constraint not found on public.approval_instances document idempotency key';
  END IF;

  IF EXISTS (
    SELECT 1
    FROM pg_constraint
    WHERE conname = 'approval_instances_document_v2_id_fkey'
      AND conrelid = 'public.approval_instances'::regclass
  ) THEN
    ALTER TABLE public.approval_instances
      RENAME CONSTRAINT approval_instances_document_v2_id_fkey
      TO approval_instances_document_id_fkey;
  END IF;

  IF EXISTS (
    SELECT 1
    FROM pg_class c
    JOIN pg_namespace n ON n.oid = c.relnamespace
    WHERE n.nspname = 'public'
      AND c.relkind = 'i'
      AND c.relname = 'ux_approval_instances_active'
  ) THEN
    ALTER INDEX public.ux_approval_instances_active
      RENAME TO ux_approval_instances_active_document_id;
  ELSIF NOT EXISTS (
    SELECT 1
    FROM pg_class c
    JOIN pg_namespace n ON n.oid = c.relnamespace
    WHERE n.nspname = 'public'
      AND c.relkind = 'i'
      AND c.relname = 'ux_approval_instances_active_document_id'
  ) THEN
    RAISE EXCEPTION 'expected active approval_instances index not found';
  END IF;

  IF EXISTS (
    SELECT 1
    FROM pg_class c
    JOIN pg_namespace n ON n.oid = c.relnamespace
    WHERE n.nspname = 'public'
      AND c.relkind = 'i'
      AND c.relname = 'ix_approval_instances_tenant_doc'
  ) THEN
    ALTER INDEX public.ix_approval_instances_tenant_doc
      RENAME TO ix_approval_instances_tenant_document_id;
  ELSIF NOT EXISTS (
    SELECT 1
    FROM pg_class c
    JOIN pg_namespace n ON n.oid = c.relnamespace
    WHERE n.nspname = 'public'
      AND c.relkind = 'i'
      AND c.relname = 'ix_approval_instances_tenant_document_id'
  ) THEN
    RAISE EXCEPTION 'expected tenant/document approval_instances index not found';
  END IF;
END $$;

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

INSERT INTO public.schema_migrations (version, description)
VALUES ('0194', 'rename approval_instances document_v2_id to document_id')
ON CONFLICT (version) DO NOTHING;

COMMIT;
