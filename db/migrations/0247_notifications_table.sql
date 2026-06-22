-- 0247_notifications_table.sql
-- Notifications module (M3/F3.2): per-recipient inbox table owned by the
-- notifications module. Read surface = list / unread-count / mark-read.
-- ADR-0043. source_event_id + the partial unique index are the F3.3 projector
-- idempotency key, shipped now so F3.3 needs no ALTER (operator decision
-- 2026-06-22). RLS uses the 0237 NULL-permissive tenant_isolation pattern
-- verbatim. Forward-only, idempotent.

BEGIN;

CREATE TABLE IF NOT EXISTS metaldocs.notifications (
    id                uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id         uuid        NOT NULL,
    recipient_user_id text        NOT NULL,
    event_type        text        NOT NULL,
    resource_type     text        NOT NULL,
    resource_id       text        NOT NULL,
    title             text        NOT NULL,
    message           text        NOT NULL,
    status            text        NOT NULL DEFAULT 'PENDING'
                                  CHECK (status IN ('PENDING', 'SENT', 'READ')),
    created_at        timestamptz NOT NULL DEFAULT now(),
    read_at           timestamptz,
    source_event_id   uuid
);

-- Keyset list index: (tenant_id, recipient_user_id) equality + (created_at DESC, id DESC) order.
CREATE INDEX IF NOT EXISTS idx_notifications_recipient_created
    ON metaldocs.notifications (tenant_id, recipient_user_id, created_at DESC, id DESC);

-- Projector idempotency (F3.3): at most one row per (recipient, source event).
-- Partial so non-projector rows (source_event_id IS NULL, e.g. test fixtures) are unconstrained.
CREATE UNIQUE INDEX IF NOT EXISTS uq_notifications_recipient_event
    ON metaldocs.notifications (recipient_user_id, source_event_id)
    WHERE source_event_id IS NOT NULL;

-- RLS: 0237 pattern verbatim (ENABLE + FORCE + one NULL-permissive tenant_isolation policy).
ALTER TABLE metaldocs.notifications ENABLE ROW LEVEL SECURITY;
ALTER TABLE metaldocs.notifications FORCE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS tenant_isolation ON metaldocs.notifications;
CREATE POLICY tenant_isolation ON metaldocs.notifications
  USING (
    NULLIF(current_setting('metaldocs.tenant_id', true), '') IS NULL
    OR tenant_id = NULLIF(current_setting('metaldocs.tenant_id', true), '')::uuid
  );

COMMIT;
