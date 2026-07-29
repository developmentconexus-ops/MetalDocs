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
    -- owner_user_id above is the 'admin' USERNAME; editable_by_role below is a
    -- ROLE and is now a bound vocabulary (F-QA4-2): the taxonomy domain
    -- validates it against internal/platform/iamtypes, so the legacy 'admin'
    -- phantom would make this seeded profile un-editable through the API.
    -- system_admin is the canonical role the 'admin' phantom migrated to.
    'system_admin',
    NULL
  ),
  (
    'it',
    'ffffffff-ffff-ffff-ffff-ffffffffffff',
    'quality',
    'Instrucao de Trabalho',
    'Perfil seeded para as jornadas de template (ADR 0086).',
    'IT',
    365,
    NULL,
    'admin',
    'system_admin',
    NULL
  )
ON CONFLICT (tenant_id, code) DO NOTHING;

-- ADR 0086: a template can only be created under a profile that already has an
-- ACTIVE TEMPLATE approval route (subject_kind='template', subject_key = the
-- profile code). Without these rows every local POST /templates would fail
-- closed with 409 APPROVAL_ROUTE_MISSING and the template QA journeys could not
-- start. Both seeded profiles are governance_class='controlado' (the column
-- default), so each route needs >= 1 approval-kind stage or
-- enforce_profile_route_policy rejects it.
INSERT INTO public.approval_routes (
  id, tenant_id, profile_code, name, version, created_by, active, subject_kind, subject_key
)
VALUES
  (
    '00000000-0000-0000-0000-0000000003a1',
    'ffffffff-ffff-ffff-ffff-ffffffffffff',
    'po',
    'Aprovacao de template - PO',
    1,
    'admin',
    TRUE,
    'template',
    'po'
  ),
  (
    '00000000-0000-0000-0000-0000000003a2',
    'ffffffff-ffff-ffff-ffff-ffffffffffff',
    'it',
    'Aprovacao de template - IT',
    1,
    'admin',
    TRUE,
    'template',
    'it'
  )
ON CONFLICT (id) DO NOTHING;

INSERT INTO public.approval_route_stages (
  id, route_id, stage_order, name, required_capability, quorum, quorum_m, on_eligibility_drift, stage_kind
)
VALUES
  (
    '00000000-0000-0000-0000-0000000003b1',
    '00000000-0000-0000-0000-0000000003a1',
    1,
    'Aprovacao',
    'template.approve',
    'any_1_of',
    NULL,
    'reduce_quorum',
    'approval'
  ),
  (
    '00000000-0000-0000-0000-0000000003b2',
    '00000000-0000-0000-0000-0000000003a2',
    1,
    'Aprovacao',
    'template.approve',
    'any_1_of',
    NULL,
    'reduce_quorum',
    'approval'
  )
ON CONFLICT (id) DO NOTHING;

-- Actor pool: role x fixed area. 'rh' holds the seeded approver memberships
-- ('approver' as qms_admin, 'approver-test' as approver), so both stages
-- resolve to a non-empty eligible pool and submit does not 422.
INSERT INTO public.approval_route_stage_selectors (
  id, tenant_id, route_stage_id, selector_order, kind, user_id, role, area_code
)
VALUES
  (
    '00000000-0000-0000-0000-0000000003c1',
    'ffffffff-ffff-ffff-ffff-ffffffffffff',
    '00000000-0000-0000-0000-0000000003b1',
    1,
    'role_in_fixed_area',
    NULL,
    'approver',
    'rh'
  ),
  (
    '00000000-0000-0000-0000-0000000003c2',
    'ffffffff-ffff-ffff-ffff-ffffffffffff',
    '00000000-0000-0000-0000-0000000003b2',
    1,
    'role_in_fixed_area',
    NULL,
    'approver',
    'rh'
  )
ON CONFLICT (id) DO NOTHING;

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
