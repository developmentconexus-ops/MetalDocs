-- MetalDocs product reference data.
-- Data in this file is required for every environment.

BEGIN;

INSERT INTO public.schema_migrations (version, description)
VALUES ('baseline-2026-05-14', 'curated current-state database baseline')
ON CONFLICT (version) DO NOTHING;

COMMIT;
