-- 0176_pdf_dispatch_outbox.sql
CREATE TABLE IF NOT EXISTS metaldocs.pdf_dispatch_outbox (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID NOT NULL,
    revision_id     UUID NOT NULL,
    content_hash    BYTEA NOT NULL,
    status          TEXT NOT NULL DEFAULT 'pending'
                    CHECK (status IN ('pending','processing','dispatched','failed')),
    attempts        INT  NOT NULL DEFAULT 0,
    last_error      TEXT,
    claimed_at      TIMESTAMPTZ,
    next_retry_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    dispatched_at   TIMESTAMPTZ,
    CONSTRAINT ux_pdf_dispatch_outbox_revision UNIQUE (tenant_id, revision_id)
);

CREATE INDEX IF NOT EXISTS ix_pdf_dispatch_outbox_pending
    ON metaldocs.pdf_dispatch_outbox (next_retry_at)
 WHERE status IN ('pending','processing');

INSERT INTO public.schema_migrations (version, description)
VALUES ('0176', 'add metaldocs.pdf_dispatch_outbox transactional outbox table')
ON CONFLICT (version) DO NOTHING;
