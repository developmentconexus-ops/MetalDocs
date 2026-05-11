-- Plan 6: add tenant isolation to audit_events (T-007)
ALTER TABLE metaldocs.audit_events
    ADD COLUMN IF NOT EXISTS tenant_id TEXT NOT NULL DEFAULT '';

CREATE INDEX IF NOT EXISTS idx_audit_events_tenant_id
    ON metaldocs.audit_events (tenant_id);
