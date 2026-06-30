-- db/migrations/0222_iam_mfa_and_failed_login_metadata.sql
-- Retro-land of PR-7 (commit e408d7578). The original file was authored as
-- migrations/0210_iam_mfa_and_failed_login_metadata.sql under the LEGACY
-- migrations/ archive directory, which the API runtime does NOT replay
-- (runtime applies db/migrations/*.sql only — see
-- apps/api/cmd/metaldocs-api/main.go:188 and wiki/database/migration-policy.md).
-- Result: live iam_users was missing mfa_enabled + mfa_enrolled_at and live
-- auth_identities was missing last_failed_login_at + last_failed_login_ip,
-- which caused /iam/kpi to 500 and made RecordFailedLogin a latent 500.
--
-- iam_users  : MFA enrollment columns. Stub-only until a real MFA flow ships
--              (see wiki/modules/security-tech-debt.md). Both default-false so
--              coverage reports report 0% honestly.
-- auth_identities : extra failure-side metadata for the security signals tab.
--              RecordSuccessfulLogin already zeroes failed_login_attempts; the
--              two new columns are populated by RecordFailedLogin (mirror of
--              the iam_users.last_login_* success-side write).

BEGIN;

ALTER TABLE metaldocs.iam_users
  ADD COLUMN IF NOT EXISTS mfa_enabled BOOLEAN NOT NULL DEFAULT FALSE,
  ADD COLUMN IF NOT EXISTS mfa_enrolled_at TIMESTAMPTZ;

ALTER TABLE metaldocs.auth_identities
  ADD COLUMN IF NOT EXISTS last_failed_login_at TIMESTAMPTZ,
  ADD COLUMN IF NOT EXISTS last_failed_login_ip TEXT;

-- Drives /security/lockouts (locked_until > now()) ordered by lockout end so
-- the UI can show "unlocks in X minutes" cards.
CREATE INDEX IF NOT EXISTS idx_auth_identities_locked_until
  ON metaldocs.auth_identities (locked_until)
  WHERE locked_until IS NOT NULL;

INSERT INTO public.schema_migrations (version, description)
VALUES ('0222', 'PR-7 retro-land: iam_users mfa metadata + auth_identities last_failed_login_* (originally orphan migrations/0210)')
ON CONFLICT (version) DO NOTHING;

COMMIT;
