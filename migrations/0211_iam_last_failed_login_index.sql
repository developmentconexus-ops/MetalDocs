-- migrations/0211_iam_last_failed_login_index.sql
-- PR-7b hardening. Sessions & Security tab calls
-- CountRecentFailedLoginsByUser, which WHERE-filters on
-- last_failed_login_at over a short rolling window. Migration 0210 added the
-- column but only indexed locked_until. Partial index keeps the working set
-- small (excludes accounts that have never failed).

BEGIN;

CREATE INDEX IF NOT EXISTS idx_auth_identities_last_failed_login_at
  ON metaldocs.auth_identities (last_failed_login_at)
  WHERE last_failed_login_at IS NOT NULL;

INSERT INTO public.schema_migrations (version, description)
VALUES ('0211', 'auth_identities partial index on last_failed_login_at (PR-7b)')
ON CONFLICT (version) DO NOTHING;

COMMIT;
