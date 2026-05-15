-- MetalDocs local development seed data.
-- Optional. Never apply this file in production/shared environments.
-- Credentials in this file are local-only and exist to support manual developer flows.

BEGIN;

CREATE EXTENSION IF NOT EXISTS pgcrypto WITH SCHEMA public;

SELECT set_config(
  'metaldocs.asserted_caps',
  '[{"cap":"user.manage"},{"cap":"membership.manage"},{"cap":"document.create"},{"cap":"document.submit"},{"cap":"document.signoff"}]',
  true
);

-- Local-only auth identities.
INSERT INTO metaldocs.auth_identities (user_id, username, display_name, password_hash, password_algo)
VALUES
  ('admin', 'admin', 'Administrator', crypt('AdminMetalDocs123!', gen_salt('bf', 12)), 'bcrypt'),
  ('approver', 'approver', 'Approver Dev', crypt('ApproverMetalDocs123!', gen_salt('bf', 12)), 'bcrypt'),
  ('author-test', 'author-test', 'Author Test', crypt('AuthorTest123!', gen_salt('bf', 12)), 'bcrypt'),
  ('approver-test', 'approver-test', 'Approver Test', crypt('ApproverMetalDocs456!@', gen_salt('bf', 12)), 'bcrypt')
ON CONFLICT (user_id) DO NOTHING;

-- Local-only IAM users.
INSERT INTO metaldocs.iam_users (user_id, display_name, tenant_id)
VALUES
  ('admin', 'Administrator', 'ffffffff-ffff-ffff-ffff-ffffffffffff'),
  ('approver', 'Approver Dev', 'ffffffff-ffff-ffff-ffff-ffffffffffff'),
  ('author-test', 'Author Test', 'ffffffff-ffff-ffff-ffff-ffffffffffff'),
  ('approver-test', 'Approver Test', 'ffffffff-ffff-ffff-ffff-ffffffffffff'),
  ('reviewer-1', 'Reviewer One', 'ffffffff-ffff-ffff-ffff-ffffffffffff'),
  ('admin-local', 'Administrator', 'ffffffff-ffff-ffff-ffff-ffffffffffff')
ON CONFLICT (user_id) DO NOTHING;

-- Local-only role assignments.
INSERT INTO metaldocs.iam_user_roles (user_id, role_code, assigned_by)
VALUES
  ('admin', 'system_admin', 'bootstrap'),
  ('approver', 'approver', 'admin-local'),
  ('author-test', 'author', 'admin-local'),
  ('approver-test', 'approver', 'admin-local')
ON CONFLICT (user_id, role_code) DO NOTHING;

COMMIT;
