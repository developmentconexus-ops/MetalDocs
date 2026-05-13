-- migrations/0191_document_placeholder_values_revision_fk.sql
-- Plan 9R Task 3: fix placeholder values revision FK to point at document_revisions(id).
BEGIN;

ALTER TABLE document_placeholder_values
  DROP CONSTRAINT IF EXISTS document_placeholder_values_revision_id_fkey;

ALTER TABLE document_placeholder_values
  ADD CONSTRAINT document_placeholder_values_revision_id_fkey
  FOREIGN KEY (revision_id) REFERENCES document_revisions(id) ON DELETE CASCADE;

COMMIT;
