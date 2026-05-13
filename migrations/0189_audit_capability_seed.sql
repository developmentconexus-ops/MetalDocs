-- Plan 6: seed audit.read capability for system_admin
BEGIN;

INSERT INTO metaldocs.role_capabilities (role, capability, created_at)
VALUES ('system_admin', 'audit.read', NOW())
ON CONFLICT (role, capability) DO NOTHING;

INSERT INTO public.schema_migrations (version, description)
VALUES ('0189', 'seed audit.read capability for system_admin')
ON CONFLICT (version) DO NOTHING;

COMMIT;
