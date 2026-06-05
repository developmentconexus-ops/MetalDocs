-- Plan 9R Task 4: seed review capability for template lifecycle.
BEGIN;

INSERT INTO metaldocs.role_capabilities (role, capability, created_at)
VALUES
  ('approver', 'template.review', NOW()),
  ('system_admin', 'template.review', NOW()),
  ('qms_admin', 'template.review', NOW())
ON CONFLICT (role, capability) DO NOTHING;

INSERT INTO public.schema_migrations (version, description)
VALUES ('0192', 'seed review capability for template lifecycle')
ON CONFLICT (version) DO NOTHING;

COMMIT;
