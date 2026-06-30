-- 0254_staging_outbox_dead_lettered_at.sql
-- F-R3 [MED]: pdf_dispatch_outbox and materialize_dispatch_outbox had no
-- dead_lettered_at column, so permanently-failed rows were invisible to
-- operators and monitoring queries. The main outbox consumer surfaces this
-- via dead_lettered_at on metaldocs.outbox_events (consumer.go:152-177).
--
-- FIX (forward-only, idempotent):
--   Add dead_lettered_at timestamptz to both staging dispatch outbox tables.
--   Uses ADD COLUMN IF NOT EXISTS for safe re-run.
--
-- SIBLING PATTERN: mirrors the dead_lettered_at column on outbox_events
-- (internal/platform/messaging/outbox/postgres/consumer.go:166-177).

BEGIN;

ALTER TABLE metaldocs.pdf_dispatch_outbox
    ADD COLUMN IF NOT EXISTS dead_lettered_at timestamptz;

ALTER TABLE metaldocs.materialize_dispatch_outbox
    ADD COLUMN IF NOT EXISTS dead_lettered_at timestamptz;

INSERT INTO public.schema_migrations (version)
VALUES ('0254')
ON CONFLICT (version) DO NOTHING;

COMMIT;
