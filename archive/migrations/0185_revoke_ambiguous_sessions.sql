-- 0185_revoke_ambiguous_sessions.sql
-- Migration 0184 set tenant_id to the DevTenantID sentinel for sessions
-- belonging to users with >1 tenant in iam_user_roles. Those sessions cannot
-- be unambiguously bound to a real tenant. Revoke them so affected users
-- re-login through resolveLoginTenant, which will correctly require a tenant
-- selection claim.
--
-- Users with zero tenants (bootstrap admin, service accounts) are left as-is;
-- AllowDevTenantFallback handles them at runtime.
--
-- This migration is idempotent: already-revoked sessions are not touched.

BEGIN;

UPDATE metaldocs.auth_sessions
SET revoked_at = NOW()
WHERE revoked_at IS NULL
  AND tenant_id = 'ffffffff-ffff-ffff-ffff-ffffffffffff'
  AND user_id IN (
    SELECT user_id
    FROM metaldocs.iam_user_roles
    GROUP BY user_id
    HAVING COUNT(DISTINCT tenant_id) > 1
  );

INSERT INTO public.schema_migrations (version, description)
VALUES ('0185', 'revoke sessions ambiguously pinned to DevTenantID by migration 0184')
ON CONFLICT (version) DO NOTHING;

COMMIT;
