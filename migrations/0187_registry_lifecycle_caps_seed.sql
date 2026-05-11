-- migrations/0187_registry_lifecycle_caps_seed.sql
-- Plan 5: rename doc.supersede -> registry.supersede (aligns with typed Capability namespace
-- established in Plan 4) and seed registry.obsolete for the same regulated roles.
-- Idempotent: UPDATE is a no-op if already renamed; INSERT uses ON CONFLICT DO NOTHING.

BEGIN;

-- Rename doc.supersede rows seeded by 0177 to match typed namespace.
UPDATE metaldocs.role_capabilities
   SET capability = 'registry.supersede'
 WHERE capability = 'doc.supersede';

-- Seed registry.obsolete for the same roles that may supersede.
INSERT INTO metaldocs.role_capabilities (role, capability) VALUES
  ('area_admin',   'registry.obsolete'),
  ('qms_admin',    'registry.obsolete'),
  ('system_admin', 'registry.obsolete')
ON CONFLICT (role, capability) DO NOTHING;

INSERT INTO public.schema_migrations (version, description)
VALUES ('0187', 'Plan 5: rename doc.supersede -> registry.supersede, seed registry.obsolete')
ON CONFLICT (version) DO NOTHING;

COMMIT;
