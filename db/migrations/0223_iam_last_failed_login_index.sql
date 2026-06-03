-- db/migrations/0223_iam_last_failed_login_index.sql
-- Retro-land of PR-7b. The original file was authored as
-- migrations/0211_iam_last_failed_login_index.sql under the LEGACY migrations/
-- archive directory, which the API runtime does NOT replay (runtime applies
-- db/migrations/*.sql only — see apps/api/cmd/metaldocs-api/main.go:188 and
-- wiki/database/migration-policy.md). Without this index,
-- CountRecentFailedLoginsByUser would either fail (missing column, fixed by
-- 0222) or seq-scan auth_identities.
--
-- PR-7b hardening. Sessions & Security tab calls
-- CountRecentFailedLoginsByUser, which WHERE-filters on
-- last_failed_login_at over a short rolling window. 0222 added the column but
-- only indexed locked_until. Partial index keeps the working set small
-- (excludes accounts that have never failed).

BEGIN;

CREATE INDEX IF NOT EXISTS idx_auth_identities_last_failed_login_at
  ON metaldocs.auth_identities (last_failed_login_at)
  WHERE last_failed_login_at IS NOT NULL;

INSERT INTO public.schema_migrations (version, description)
VALUES ('0223', 'PR-7b retro-land: auth_identities partial index on last_failed_login_at (originally orphan migrations/0211)')
ON CONFLICT (version) DO NOTHING;

COMMIT;
