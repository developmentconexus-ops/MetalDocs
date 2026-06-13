-- 0238: drop orphan metaldocs.documents.subject_code (+index) — FK was CASCADE-dropped by 0236 (CD T-010).
-- Wave Z Z-25. Zero code references verified (git grep + information_schema).
-- Forward-only, idempotent.

BEGIN;

DROP INDEX IF EXISTS metaldocs.idx_documents_subject_code;
ALTER TABLE metaldocs.documents DROP COLUMN IF EXISTS subject_code;

INSERT INTO public.schema_migrations (version, description)
VALUES ('0238', 'Drop orphan metaldocs.documents.subject_code column + index (Wave Z Z-25, CD T-010, 0236 residual)')
ON CONFLICT (version) DO NOTHING;

COMMIT;
