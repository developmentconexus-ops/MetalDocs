-- 0239: pg_trgm GIN indexes for controlled_documents ILIKE search (code, title) — Wave Z Z-24 (F-20b, REQ perf).
-- Accelerates leading-wildcard ILIKE '%x%' queries natively via trigram matching.
-- CREATE EXTENSION IF NOT EXISTS is idempotent. Indexes are non-concurrent inside txn (safe, small table).

BEGIN;

CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE INDEX IF NOT EXISTS idx_controlled_documents_code_trgm
    ON public.controlled_documents USING GIN (code gin_trgm_ops);

CREATE INDEX IF NOT EXISTS idx_controlled_documents_title_trgm
    ON public.controlled_documents USING GIN (title gin_trgm_ops);

INSERT INTO public.schema_migrations (version, description)
VALUES ('0239', 'pg_trgm GIN indexes on controlled_documents(code, title) for ILIKE search (Wave Z Z-24, F-20b)')
ON CONFLICT (version) DO NOTHING;

COMMIT;
