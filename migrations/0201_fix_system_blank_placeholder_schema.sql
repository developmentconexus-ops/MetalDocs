BEGIN;

UPDATE public.templates_v2_template_version
SET placeholder_schema = '[]'::jsonb
WHERE id = '00000000-0000-0000-0000-000000000102'
  AND jsonb_typeof(placeholder_schema) = 'object';

INSERT INTO public.schema_migrations (version, description)
VALUES ('0201', 'fix system blank template placeholder_schema shape')
ON CONFLICT (version) DO NOTHING;

COMMIT;
