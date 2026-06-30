BEGIN;

ALTER TABLE public.documents
  ADD COLUMN IF NOT EXISTS revision_title text;

INSERT INTO public.schema_migrations (version, description)
VALUES ('0204', 'add documents.revision_title for governed revision metadata')
ON CONFLICT (version) DO NOTHING;

COMMIT;
