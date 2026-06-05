-- 0175_documents_area_name_snapshot.sql
BEGIN;

ALTER TABLE public.documents
    ADD COLUMN IF NOT EXISTS area_name_snapshot TEXT;

UPDATE public.documents d
   SET area_name_snapshot = pa.name
  FROM metaldocs.document_process_areas pa
 WHERE pa.tenant_id = d.tenant_id
   AND pa.code      = d.process_area_code_snapshot
   AND d.area_name_snapshot IS NULL;

INSERT INTO public.schema_migrations (version, description)
VALUES ('0175', 'add area_name_snapshot to documents')
ON CONFLICT (version) DO NOTHING;

COMMIT;
