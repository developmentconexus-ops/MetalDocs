-- MetalDocs local development seed data.
-- Optional. Never apply this file in production/shared environments.
-- Credentials in this file are local-only and exist to support manual developer flows.

BEGIN;

CREATE EXTENSION IF NOT EXISTS pgcrypto WITH SCHEMA public;

-- Local-only auth identities.
INSERT INTO metaldocs.auth_identities (user_id, username, display_name, password_hash, password_algo)
VALUES
  ('approver', 'approver', 'Approver Dev', crypt('ApproverMetalDocs123!', gen_salt('bf', 12)), 'bcrypt'),
  ('author-test', 'author-test', 'Author Test', crypt('AuthorTest123!', gen_salt('bf', 12)), 'bcrypt'),
  ('approver-test', 'approver-test', 'Approver Test', crypt('ApproverMetalDocs456!@', gen_salt('bf', 12)), 'bcrypt')
ON CONFLICT (user_id) DO NOTHING;

-- Local-only IAM users.
INSERT INTO metaldocs.iam_users (user_id, display_name, tenant_id)
VALUES
  ('approver', 'Approver Dev', 'ffffffff-ffff-ffff-ffff-ffffffffffff'),
  ('author-test', 'Author Test', 'ffffffff-ffff-ffff-ffff-ffffffffffff'),
  ('approver-test', 'Approver Test', 'ffffffff-ffff-ffff-ffff-ffffffffffff'),
  ('reviewer-1', 'Reviewer One', 'ffffffff-ffff-ffff-ffff-ffffffffffff'),
  ('admin-local', 'Administrator', 'ffffffff-ffff-ffff-ffff-ffffffffffff')
ON CONFLICT (user_id) DO NOTHING;

-- Local-only role assignments.
INSERT INTO metaldocs.iam_user_roles (user_id, role_code, assigned_by)
VALUES
  ('approver', 'approver', 'admin-local'),
  ('author-test', 'system_admin', 'admin-local'),
  ('approver-test', 'approver', 'admin-local')
ON CONFLICT (user_id, role_code) DO NOTHING;

-- Local process-area convenience memberships.
INSERT INTO public.user_process_areas (user_id, tenant_id, area_code, role, effective_from, granted_by)
VALUES
  ('reviewer-1', 'ffffffff-ffff-ffff-ffff-ffffffffffff', 'quality', 'reviewer', now(), 'admin-local')
ON CONFLICT DO NOTHING;

DO $$
BEGIN
  IF EXISTS (
    SELECT 1
    FROM public.user_process_areas
    WHERE user_id = 'admin-local'
      AND tenant_id = 'ffffffff-ffff-ffff-ffff-ffffffffffff'
      AND area_code = 'quality'
      AND effective_to IS NULL
  ) THEN
    UPDATE public.user_process_areas
    SET role = 'qms_admin'
    WHERE user_id = 'admin-local'
      AND tenant_id = 'ffffffff-ffff-ffff-ffff-ffffffffffff'
      AND area_code = 'quality'
      AND effective_to IS NULL;
  ELSE
    INSERT INTO public.user_process_areas (user_id, tenant_id, area_code, role, effective_from, granted_by)
    VALUES ('admin-local', 'ffffffff-ffff-ffff-ffff-ffffffffffff', 'quality', 'qms_admin', now(), 'admin-local');
  END IF;
END $$;

COMMIT;
