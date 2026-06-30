-- 0219_iam_users_last_login_context.sql
-- PR-4 of the 12-PR IAM Admin Center rebuild (ADR 0019).
--
-- Adds last-login context columns to metaldocs.iam_users so the People-tab
-- "Last login" drawer (PR-11/12) can surface IP / User-Agent / device label
-- alongside the timestamp that already lives on metaldocs.auth_identities.
--
-- Why iam_users and not auth_identities:
--   * auth_identities.last_login_at is part of the credential surface (set on
--     every successful bcrypt verify). The new columns are governance metadata
--     scoped to a tenant-managed user record and therefore belong with the
--     tenant-scoped iam_users row.
--   * PR-7 (Sessions tab) will derive last_login_device_label from User-Agent
--     parsing; PR-4 only populates ip + user_agent verbatim from the request.
--
-- Idempotent: ADD COLUMN IF NOT EXISTS.
-- Forward-only per wiki/database/migration-policy.md.

BEGIN;

ALTER TABLE metaldocs.iam_users
    ADD COLUMN IF NOT EXISTS last_login_ip           text,
    ADD COLUMN IF NOT EXISTS last_login_user_agent   text,
    ADD COLUMN IF NOT EXISTS last_login_device_label text;

COMMENT ON COLUMN metaldocs.iam_users.last_login_ip           IS 'Client IP recorded on most recent successful login (PR-4).';
COMMENT ON COLUMN metaldocs.iam_users.last_login_user_agent   IS 'Raw User-Agent header on most recent successful login (PR-4).';
COMMENT ON COLUMN metaldocs.iam_users.last_login_device_label IS 'Derived device label (browser + OS). Populated by PR-7; nullable until then.';

INSERT INTO public.schema_migrations (version, description)
VALUES ('0219', 'PR-4: iam_users last-login context columns (ip/user_agent/device_label)')
ON CONFLICT (version) DO NOTHING;

COMMIT;
