-- migrations/0169_role_capabilities_process_areas.sql
-- Migration 0165 truncated role_capabilities and only re-seeded
-- viewer/editor/author/approver/system_admin. Legacy process-area roles
-- signer/area_admin/qms_admin had zero capabilities, causing authz.Require
-- to return 403 for every non-admin user.
-- Idempotent via ON CONFLICT (assumes unique constraint on role+capability).

BEGIN;

INSERT INTO metaldocs.role_capabilities (role, capability) VALUES
  ('signer', 'doc.view'),
  ('signer', 'doc.signoff'),
  ('area_admin', 'doc.view'),
  ('area_admin', 'doc.create'),
  ('area_admin', 'doc.edit'),
  ('area_admin', 'doc.submit'),
  ('area_admin', 'doc.signoff'),
  ('area_admin', 'membership.manage'),
  ('qms_admin', 'doc.view'),
  ('qms_admin', 'doc.create'),
  ('qms_admin', 'doc.edit'),
  ('qms_admin', 'doc.submit'),
  ('qms_admin', 'doc.signoff'),
  ('qms_admin', 'template.view'),
  ('qms_admin', 'template.approve'),
  ('qms_admin', 'template.publish'),
  ('qms_admin', 'route.manage'),
  ('qms_admin', 'taxonomy.manage')
ON CONFLICT (role, capability) DO NOTHING;

COMMIT;
