-- migrations/0203_idempotency_two_phase.sql
-- C3/C4 fix: enable two-phase idempotency writes. The `in_flight` row inserted
-- by BeginReplay has no response yet, so `response_status` / `response_body`
-- must allow NULL until CompleteReplay populates them. Also adds a partial
-- index targeted at the janitor's orphan-sweep query.

BEGIN;

ALTER TABLE metaldocs.idempotency_keys
    ALTER COLUMN response_status DROP NOT NULL,
    ALTER COLUMN response_body   DROP NOT NULL;

-- Belt: a completed row must always carry a response.
ALTER TABLE metaldocs.idempotency_keys
    DROP CONSTRAINT IF EXISTS idempotency_keys_completed_has_response;
ALTER TABLE metaldocs.idempotency_keys
    ADD CONSTRAINT idempotency_keys_completed_has_response
    CHECK (status <> 'completed'
           OR (response_status IS NOT NULL AND response_body IS NOT NULL));

-- M9 belt: response_status is an HTTP status code; reject 0 / negative / >599.
ALTER TABLE metaldocs.idempotency_keys
    DROP CONSTRAINT IF EXISTS idempotency_keys_response_status_range;
ALTER TABLE metaldocs.idempotency_keys
    ADD CONSTRAINT idempotency_keys_response_status_range
    CHECK (response_status IS NULL
           OR (response_status BETWEEN 100 AND 599));

-- Janitor orphan sweep: in_flight rows past expires_at point to crashed
-- handlers. The partial index keeps the warn-log query cheap as the table
-- grows.
CREATE INDEX IF NOT EXISTS idx_idempotency_keys_in_flight_expires
    ON metaldocs.idempotency_keys (expires_at)
    WHERE status = 'in_flight';

INSERT INTO public.schema_migrations (version, description)
VALUES ('0203', 'idempotency two-phase writes: nullable response columns + orphan index')
ON CONFLICT (version) DO NOTHING;

COMMIT;
