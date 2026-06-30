-- 0217_view_grade_capabilities.sql
-- Seed View-grade capabilities (ADR 0016) into metaldocs.role_capabilities.
--
-- Closes the registry gap surfaced by F-001 audit: Tier-1 declarative authz
-- rows for /metrics, /access-policies, /iam/users, /iam/area-memberships, and
-- /taxonomy/* read paths had no correct read cap to point at, so they fell
-- back to Manage-grade caps and collapsed read/write separation.
--
-- Adds four new caps:
--   metrics.view, membership.view, user.view, taxonomy.view
--
-- Grant matrix per ADR 0016 (product-confirmed):
--   system_admin:  all 4
--   area_admin:    membership.view, user.view, taxonomy.view
--   approver:      membership.view, taxonomy.view
--   author:        membership.view, taxonomy.view
--   editor:        membership.view, taxonomy.view
--   viewer:        membership.view, taxonomy.view
--
-- Idempotent: ON CONFLICT (role, capability) DO NOTHING.
-- Does NOT modify permissions.go; the Tier-1 rule rewrite lands in the F-001
-- follow-up PR.

BEGIN;

-- system_admin — all four view caps
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES
    ('system_admin', 'metrics.view',     'Read /api/v1/metrics Prometheus scrape endpoint'),
    ('system_admin', 'membership.view',  'Read access policies and area memberships'),
    ('system_admin', 'user.view',        'Read user directory and IAM admin overview'),
    ('system_admin', 'taxonomy.view',    'Read taxonomy profiles, areas, families')
ON CONFLICT (role, capability) DO NOTHING;

-- area_admin — admin-ish reads, no metrics
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES
    ('area_admin', 'membership.view', 'Read access policies and area memberships'),
    ('area_admin', 'user.view',       'Read user directory'),
    ('area_admin', 'taxonomy.view',   'Read taxonomy profiles, areas, families')
ON CONFLICT (role, capability) DO NOTHING;

-- approver / author / editor / viewer — membership + taxonomy reads only
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES
    ('approver', 'membership.view', 'Read access policies and area memberships'),
    ('approver', 'taxonomy.view',   'Read taxonomy profiles, areas, families'),
    ('author',   'membership.view', 'Read access policies and area memberships'),
    ('author',   'taxonomy.view',   'Read taxonomy profiles, areas, families'),
    ('editor',   'membership.view', 'Read access policies and area memberships'),
    ('editor',   'taxonomy.view',   'Read taxonomy profiles, areas, families'),
    ('viewer',   'membership.view', 'Read access policies and area memberships'),
    ('viewer',   'taxonomy.view',   'Read taxonomy profiles, areas, families')
ON CONFLICT (role, capability) DO NOTHING;

INSERT INTO public.schema_migrations (version, description)
VALUES ('0217', 'seed view-grade caps metrics.view/membership.view/user.view/taxonomy.view per ADR 0016')
ON CONFLICT (version) DO NOTHING;

COMMIT;
