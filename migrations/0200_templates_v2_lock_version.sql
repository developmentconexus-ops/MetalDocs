BEGIN;

ALTER TABLE public.templates_v2_template_version
  ADD COLUMN IF NOT EXISTS lock_version integer NOT NULL DEFAULT 0;

INSERT INTO public.schema_migrations (version, description)
VALUES ('0200', 'add lock_version to templates_v2_template_version')
ON CONFLICT (version) DO NOTHING;

COMMIT;
