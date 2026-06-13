-- 0238: drop orphan metaldocs.documents.subject_code (+index) — FK was CASCADE-dropped by 0236 (CD T-010).
DROP INDEX IF EXISTS metaldocs.idx_documents_subject_code;
ALTER TABLE metaldocs.documents DROP COLUMN IF EXISTS subject_code;
