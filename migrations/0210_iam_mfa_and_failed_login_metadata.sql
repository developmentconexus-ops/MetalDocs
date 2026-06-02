-- migrations/0210_iam_mfa_and_failed_login_metadata.sql
-- PR-7 Sessions & Security tab backend.
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
VALUES ('0210', 'iam_users mfa metadata + auth_identities last_failed_login_* (PR-7)')
ON CONFLICT (version) DO NOTHING;

COMMIT;
