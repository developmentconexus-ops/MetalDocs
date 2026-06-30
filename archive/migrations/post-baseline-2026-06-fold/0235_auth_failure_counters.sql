-- db/migrations/0235_auth_failure_counters.sql
-- Postgres-backed auth-failure rate-limiter counter table.
-- Replaces process-local InMemoryAuthFailureRateLimiter in production,
-- enabling lockout state to survive API restarts and to be shared across
-- replicas (F-20e, REQ-REL-3, OWASP ASVS §2.2.1, D-1: no Redis).
--
-- Design:
--   One row per actor_id (user UUID as text). The window is fixed (not sliding):
--   fail_count resets when the current clock exceeds window_start + 60s.
--   The application layer performs the reset check; the DB stores the counter.
--   Stale rows (window_start older than 2× windowDur) are pruned by the
--   RecordFailure UPSERT to avoid unbounded table growth without a janitor job.
--
-- Owner: approval / auth boundary (keyed on iam_users.id).
-- Schema: public (same as approval tables).

BEGIN;

CREATE TABLE IF NOT EXISTS public.auth_failure_counters (
    actor_id    TEXT        NOT NULL,
    fail_count  INTEGER     NOT NULL DEFAULT 0,
    window_start TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (actor_id)
);

-- Index is the PK itself (single-column lookup). No extra index needed.

-- Grant to the application role that owns the public schema objects.
-- (metaldocs_app is the PGUSER in .env — see .env for credentials.)
GRANT SELECT, INSERT, UPDATE, DELETE ON public.auth_failure_counters TO metaldocs_app;

INSERT INTO public.schema_migrations (version, description)
VALUES ('0235', 'auth_failure_counters: Postgres-backed auth-failure rate-limiter table (F-20e, REQ-REL-3, D-1)')
ON CONFLICT (version) DO NOTHING;

COMMIT;
