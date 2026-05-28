-- MetalDocs local development seed data.
-- Optional. Never apply this file in production/shared environments.
-- Credentials in this file are local-only and exist to support manual developer flows.

BEGIN;

CREATE EXTENSION IF NOT EXISTS pgcrypto WITH SCHEMA public;

SELECT set_config(
  'metaldocs.asserted_caps',
  '[{"cap":"user.manage"},{"cap":"membership.manage"},{"cap":"taxonomy.manage"},{"cap":"document.create"},{"cap":"document.submit"},{"cap":"document.signoff"}]',
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

-- Local-only taxonomy records so /documents/new is testable after baseline bootstrap.
INSERT INTO metaldocs.document_families (code, name, description, is_active)
VALUES
  ('quality', 'Qualidade', 'Familia dev local para smoke do wizard.', TRUE)
ON CONFLICT (code) DO NOTHING;

INSERT INTO metaldocs.document_process_areas (
  code, tenant_id, name, description, is_active, parent_code, owner_user_id, default_approver_role, archived_at
)
VALUES
  (
    'rh',
    'ffffffff-ffff-ffff-ffff-ffffffffffff',
    'Recursos Humanos',
    'Area seeded para smoke do novo documento.',
    TRUE,
    NULL,
    'admin',
    NULL,
    NULL
  )
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO metaldocs.document_profiles (
  code,
  tenant_id,
  family_code,
  name,
  description,
  alias,
  review_interval_days,
  default_template_version_id,
  owner_user_id,
  editable_by_role,
  archived_at
)
VALUES
  (
    'po',
    'ffffffff-ffff-ffff-ffff-ffffffffffff',
    'quality',
    'Procedimento Operacional',
    'Perfil seeded para smoke do novo documento.',
    'PO',
    365,
    '00000000-0000-0000-0000-000000000102',
    'admin',
    'admin',
    NULL
  )
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.user_process_areas (
  user_id,
  tenant_id,
  area_code,
  role,
  effective_from,
  effective_to,
  granted_by,
  revoked_by
)
VALUES
  (
    'admin',
    'ffffffff-ffff-ffff-ffff-ffffffffffff',
    'rh',
    'qms_admin',
    now(),
    NULL,
    'admin-local',
    NULL
  ),
  (
    'author-test',
    'ffffffff-ffff-ffff-ffff-ffffffffffff',
    'rh',
    'author',
    now(),
    NULL,
    'admin-local',
    NULL
  ),
  (
    'approver-test',
    'ffffffff-ffff-ffff-ffff-ffffffffffff',
    'rh',
    'reviewer',
    now(),
    NULL,
    'admin-local',
    NULL
  )
ON CONFLICT (tenant_id, user_id, area_code, role) WHERE effective_to IS NULL DO NOTHING;

COMMIT;
