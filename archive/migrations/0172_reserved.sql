-- intentionally skipped during Group C/D split
INSERT INTO public.schema_migrations (version, description)
VALUES ('0172', 'reserved (intentionally skipped during Group C/D split)')
ON CONFLICT (version) DO NOTHING;
