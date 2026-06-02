-- 0220_iam_users_last_seen_at.sql
-- PR-9 of the IAM Admin Center rebuild (ADR 0019).
--
-- Adds last_seen_at to metaldocs.iam_users so the new WebSocket presence
-- stream (GET /api/v1/iam/presence/stream) and HTTP snapshot fallback
-- (GET /api/v1/iam/presence/snapshot) can compute online/idle status
-- from a per-user, tenant-scoped heartbeat instead of the legacy
-- 10-minute wall-clock window derived from metaldocs.auth_sessions.
--
-- Why iam_users and not auth_sessions:
--   * auth_sessions.last_seen_at tracks per-session liveness; multiple
--     concurrent sessions per user inflate the "online" count.
--   * Presence is governance metadata (tenant-scoped, scoped to the
--     managed user record) and therefore belongs on iam_users.
--   * PR-9 hybrid (Option C): authenticated request middleware bumps
--     last_seen_at debounced once per 60s per user; WS heartbeat also
--     bumps. Status: online <60s, idle 60s-5min, off >5min.
--
-- Idempotent: ADD COLUMN IF NOT EXISTS / CREATE INDEX IF NOT EXISTS.
-- Forward-only per wiki/database/migration-policy.md.

BEGIN;

ALTER TABLE metaldocs.iam_users
    ADD COLUMN IF NOT EXISTS last_seen_at timestamp with time zone NOT NULL DEFAULT now();

COMMENT ON COLUMN metaldocs.iam_users.last_seen_at IS
    'Most recent authenticated-request or WS heartbeat timestamp (PR-9). Drives presence online/idle classification.';

CREATE INDEX IF NOT EXISTS ix_iam_users_tenant_last_seen
    ON metaldocs.iam_users (tenant_id, last_seen_at DESC);

INSERT INTO public.schema_migrations (version, description)
VALUES ('0220', 'PR-9: iam_users.last_seen_at + tenant/last_seen_at index for WS presence stream')
ON CONFLICT (version) DO NOTHING;

COMMIT;
