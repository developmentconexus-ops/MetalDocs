-- 0184_auth_sessions_tenant_id.sql
-- Bind each session to the tenant chosen at login. Backfills existing rows
-- using the user's lone IAM role tenant; rows where the user has no role (or
-- multiple) fall through to DevTenantID so dev/test data continues to resolve.

ALTER TABLE metaldocs.auth_sessions
  ADD COLUMN IF NOT EXISTS tenant_id TEXT;

UPDATE metaldocs.auth_sessions s
SET tenant_id = sub.tenant_id
FROM (
  SELECT user_id, MIN(tenant_id) AS tenant_id
  FROM metaldocs.iam_user_roles
  GROUP BY user_id
  HAVING COUNT(DISTINCT tenant_id) = 1
) sub
WHERE s.user_id = sub.user_id AND s.tenant_id IS NULL;

UPDATE metaldocs.auth_sessions
SET tenant_id = 'ffffffff-ffff-ffff-ffff-ffffffffffff'
WHERE tenant_id IS NULL;

ALTER TABLE metaldocs.auth_sessions
  ALTER COLUMN tenant_id SET NOT NULL;

CREATE INDEX IF NOT EXISTS idx_auth_sessions_tenant_user
  ON metaldocs.auth_sessions (tenant_id, user_id);
