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
  ),
  (
    'qualidade',
    'ffffffff-ffff-ffff-ffff-ffffffffffff',
    'Qualidade',
    'Area seeded para QA do diretório de memberships.',
    TRUE,
    NULL,
    'admin',
    NULL,
    NULL
  ),
  (
    'producao',
    'ffffffff-ffff-ffff-ffff-ffffffffffff',
    'Producao',
    'Area seeded sem membership de author-test para exercitar o grant new-pair.',
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
    'approver',
    now(),
    NULL,
    'admin-local',
    NULL
  ),
  -- F-IAM1: 'approver' (the wiki QA approval account) needs a tier-2 membership.
  -- It had iam_user_roles but no user_process_areas, so authz.Require (which
  -- reads only UPA) returned 403 FORBIDDEN_CAPABILITY on every approve/publish
  -- even though /auth/me listed the capability.
  -- TST-03 (corrected to runtime truth 2026-07-03): the row's role is
  -- 'qms_admin', not 'approver'. Runtime facts, live-proven: Approve(accept)
  -- PUBLISHES DIRECTLY, including on reviewer-stage templates (lifecycle.go
  -- Approve accept branch) — there is no separate publish call in the
  -- workflow path, and the direct POST /publish route's handler gate reads
  -- iam_user_roles x role_capabilities (CapabilityService.CanDo), which this
  -- UPA row cannot influence (only 'admin'/system_admin passes it among
  -- seeded users). This row's job is tier-2: authz.Require reads only UPA,
  -- and 'approver' had NO UPA row at all (403 on every approve). A second
  -- 'approver'-role row is impossible: partial unique index
  -- ux_user_process_areas_one_active allows ONE active row per
  -- (user, tenant, area). qms_admin's capability set is a superset of
  -- approver's document/template caps (template.approve,
  -- document.signoff, ...), so the superset role is the safe single choice.
  -- domain.CheckSegregation is untouched (SoD compares user_id, not role);
  -- role='approver' coverage stays exercised by approver-test above.
  (
    'approver',
    'ffffffff-ffff-ffff-ffff-ffffffffffff',
    'rh',
    'qms_admin',
    now(),
    NULL,
    'admin-local',
    NULL
  ),
  -- Qualidade: varied roles so the tenant directory shows multiple users/areas.
  (
    'admin',
    'ffffffff-ffff-ffff-ffff-ffffffffffff',
    'qualidade',
    'qms_admin',
    now(),
    NULL,
    'admin-local',
    NULL
  ),
  (
    'approver-test',
    'ffffffff-ffff-ffff-ffff-ffffffffffff',
    'qualidade',
    'approver',
    now(),
    NULL,
    'admin-local',
    NULL
  ),
  (
    'reviewer-1',
    'ffffffff-ffff-ffff-ffff-ffffffffffff',
    'qualidade',
    'viewer',
    now(),
    NULL,
    'admin-local',
    NULL
  )
  -- NOTE: 'producao' is intentionally left with no memberships, and 'author-test'
  -- stays only in 'rh', so QA can exercise both grant paths: role-change
  -- (author-test rh/author -> rh/viewer, GrantAtomic) and new-pair
  -- (author-test -> producao/author, Insert).
ON CONFLICT (tenant_id, user_id, area_code, role) WHERE effective_to IS NULL DO NOTHING;

COMMIT;
