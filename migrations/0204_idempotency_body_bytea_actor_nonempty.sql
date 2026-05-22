-- migrations/0204_idempotency_body_bytea_actor_nonempty.sql
-- H11: three schema defects in metaldocs.idempotency_keys corrected.
--
-- 1. response_body JSONB → BYTEA.
--    JSONB coerces writes through a JSON parser; a non-JSON 2xx response body
--    (binary, plain-text, CSV) fails the insert. BYTEA stores raw bytes as-is,
--    which is the correct type for []byte from r.body.Bytes().
--    Existing rows: NULL (in_flight) stays NULL; completed rows hold JSON text —
--    converted to UTF-8 byte strings via the USING expression.
--
-- 2. 64 KiB size cap on response_body.
--    No prior limit meant a single multi-MB rendered-PDF response could inflate
--    the table unboundedly. The Go layer enforces the same cap before writing
--    (postgres_store.go CompleteReplay / RecordReplay), so this constraint is a
--    belt, not the primary gate.
--
-- 3. CHECK (actor_user_id <> '').
--    The column accepted any string, including the empty string. A UUID-type
--    column with FK is not viable: iam_users.user_id is TEXT and service
--    accounts ("system") are valid actors not stored in iam_users. The non-empty
--    CHECK is the narrowest feasible constraint; the Go layer validates this
--    before DB writes as well.
--
-- Deployment caveat: if prod data contains actor_user_id = '' rows, the
-- CHECK addition will fail. Verify with:
--   SELECT COUNT(*) FROM metaldocs.idempotency_keys WHERE actor_user_id = '';
-- before applying. Empty-string rows must be deleted or backfilled first.

BEGIN;

-- 1. Convert JSONB → BYTEA.
ALTER TABLE metaldocs.idempotency_keys
    ALTER COLUMN response_body TYPE BYTEA
        USING CASE
            WHEN response_body IS NULL THEN NULL
            ELSE convert_to(response_body::text, 'UTF8')
        END;

-- 2. 64 KiB cap (covers NULL for in_flight rows).
ALTER TABLE metaldocs.idempotency_keys
    DROP CONSTRAINT IF EXISTS idempotency_keys_response_body_size;
ALTER TABLE metaldocs.idempotency_keys
    ADD CONSTRAINT idempotency_keys_response_body_size
        CHECK (response_body IS NULL OR octet_length(response_body) <= 65536);

-- 3. Non-empty actor_user_id.
ALTER TABLE metaldocs.idempotency_keys
    DROP CONSTRAINT IF EXISTS idempotency_keys_actor_nonempty;
ALTER TABLE metaldocs.idempotency_keys
    ADD CONSTRAINT idempotency_keys_actor_nonempty
        CHECK (actor_user_id <> '');

INSERT INTO public.schema_migrations (version, description)
VALUES ('0204', 'idempotency: response_body JSONB→BYTEA + 64 KiB cap + non-empty actor_user_id')
ON CONFLICT (version) DO NOTHING;

COMMIT;
