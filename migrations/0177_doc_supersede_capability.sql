-- 0177_doc_supersede_capability.sql
-- Migration 0165 truncated role_capabilities and did not re-seed doc.supersede.
-- Re-add it for roles that need to publish new revisions.

BEGIN;

INSERT INTO metaldocs.role_capabilities (role, capability) VALUES
  ('area_admin',   'doc.supersede'),
  ('qms_admin',    'doc.supersede'),
  ('system_admin', 'doc.supersede')
ON CONFLICT (role, capability) DO NOTHING;

INSERT INTO public.schema_migrations (version, description)
VALUES ('0177', 'restore doc.supersede capability for area_admin, qms_admin, system_admin')
ON CONFLICT (version) DO NOTHING;

COMMIT;
