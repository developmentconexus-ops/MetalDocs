-- 0219_audit_export_jobs_pr6.sql
-- PR-6 of the 12-PR IAM Admin Center rebuild (ADR 0019).
--
-- Adds the persistence table for audit export jobs (Audit Trail tab,
-- POST /api/v1/audit/events/export + GET /api/v1/audit/events/export/{id}).
--
-- For PR-6 the sync path stores the rendered CSV/JSONL blob inline in the
-- payload BYTEA column. An async worker (deferred to a later PR) can later
-- push the blob to object storage and only keep metadata + object_key here.
--
-- Tenant + actor scoping is enforced by the application layer and reinforced
-- by the (tenant_id, actor_id, created_at DESC) index used by status polling.
-- The download path matches on (id, download_token); the token is a random
-- 24-byte hex string and the row carries an expires_at column so the URL is
-- effectively short-lived even before any object-storage signing exists.

BEGIN;

CREATE TABLE IF NOT EXISTS metaldocs.audit_export_jobs (
    id              TEXT PRIMARY KEY,
    tenant_id       UUID NOT NULL,
    actor_id        TEXT NOT NULL,
    format          TEXT NOT NULL CHECK (format IN ('csv','jsonl')),
    filter_json     JSONB NOT NULL,
    status          TEXT NOT NULL DEFAULT 'pending'
                    CHECK (status IN ('pending','running','ready','failed')),
    object_key      TEXT,
    download_token  TEXT,
    expires_at      TIMESTAMPTZ,
    error_message   TEXT,
    estimated_rows  BIGINT NOT NULL DEFAULT 0,
    actual_rows     BIGINT NOT NULL DEFAULT 0,
    payload         BYTEA,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    completed_at    TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS ix_audit_export_jobs_tenant_actor
    ON metaldocs.audit_export_jobs (tenant_id, actor_id, created_at DESC);

CREATE INDEX IF NOT EXISTS ix_audit_export_jobs_download_token
    ON metaldocs.audit_export_jobs (download_token)
    WHERE download_token IS NOT NULL;

COMMIT;
