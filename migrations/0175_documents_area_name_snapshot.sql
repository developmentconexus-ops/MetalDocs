-- 0175_documents_area_name_snapshot.sql
ALTER TABLE metaldocs.documents
    ADD COLUMN IF NOT EXISTS area_name_snapshot TEXT;

UPDATE metaldocs.documents d
   SET area_name_snapshot = pa.name
  FROM metaldocs.process_areas pa
 WHERE pa.tenant_id = d.tenant_id
   AND pa.code      = d.area_code_snapshot
   AND d.area_name_snapshot IS NULL;
