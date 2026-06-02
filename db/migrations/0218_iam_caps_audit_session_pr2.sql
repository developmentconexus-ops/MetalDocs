-- 0218_iam_caps_audit_session_pr2.sql
-- PR-2 of the 12-PR IAM Admin Center rebuild (ADR 0019).
--
-- Broadens audit.read grants beyond system_admin and introduces a new
-- session.manage capability for force-logout / session-revoke operations.
--
-- Caps land DORMANT — no handler consumes them yet:
--   * PR-6 (audit trail) will wire handlers against audit.read.
--   * PR-7 (sessions & security) will wire handlers against session.manage.
--
-- Pre-existing state (DO NOT re-seed; here only for context):
--   * audit.read for system_admin was seeded by migrations/0189_audit_capability_seed.sql.
--
-- Grant matrix this migration applies (per ADR 0019):
--   audit.read     -> qms_admin, area_admin, approver
--                     (system_admin already holds it from 0189)
--   session.manage -> system_admin only
--
-- Rationale (see ADR 0019 §"Role grant matrix"):
--   * qms_admin   — ISO/QMS compliance lead, needs the trail for audits.
--   * area_admin  — needs visibility into events scoped to their area.
--   * approver    — needs audit on documents they approve (signoff integrity).
--   * session.manage restricted to system_admin — security-sensitive write.
--
-- Idempotent: ON CONFLICT (role, capability) DO NOTHING.
-- Forward-only per wiki/database/migration-policy.md.
-- Does NOT modify apps/api/cmd/metaldocs-api/permissions.go (Tier-1 (method,path)
-- mapping lands with PR-6 and PR-7).

BEGIN;

INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES
    ('qms_admin',    'audit.read',     'Read governance / audit events'),
    ('area_admin',   'audit.read',     'Read governance / audit events'),
    ('approver',     'audit.read',     'Read governance / audit events'),
    ('system_admin', 'session.manage', 'Force-logout / revoke auth sessions')
ON CONFLICT (role, capability) DO NOTHING;

INSERT INTO public.schema_migrations (version, description)
VALUES ('0218', 'PR-2: broaden audit.read grants + add session.manage capability (ADR 0019)')
ON CONFLICT (version) DO NOTHING;

COMMIT;
