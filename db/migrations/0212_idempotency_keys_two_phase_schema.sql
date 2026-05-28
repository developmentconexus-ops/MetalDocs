-- 0212_idempotency_keys_two_phase_schema.sql
-- Curated-tail mirror of the legacy idempotency schema repairs from the
-- historical migrations chain. Fresh baseline installs and existing databases
-- that only know the curated tail must converge on the same two-phase row
-- shape used by internal/platform/idempotency.

BEGIN;

ALTER TABLE metaldocs.idempotency_keys
    ALTER COLUMN response_status DROP NOT NULL,
    ALTER COLUMN response_body DROP NOT NULL;

DO $$
DECLARE
    response_body_type text;
BEGIN
    SELECT c.udt_name
      INTO response_body_type
      FROM information_schema.columns AS c
     WHERE c.table_schema = 'metaldocs'
       AND c.table_name = 'idempotency_keys'
       AND c.column_name = 'response_body';

    IF response_body_type IS DISTINCT FROM 'bytea' THEN
        EXECUTE $sql$
            ALTER TABLE metaldocs.idempotency_keys
                ALTER COLUMN response_body TYPE BYTEA
                    USING CASE
                        WHEN response_body IS NULL THEN NULL
                        ELSE convert_to(response_body::text, 'UTF8')
                    END
        $sql$;
    END IF;
END $$;

ALTER TABLE metaldocs.idempotency_keys
    DROP CONSTRAINT IF EXISTS idempotency_keys_completed_has_response;
ALTER TABLE metaldocs.idempotency_keys
    ADD CONSTRAINT idempotency_keys_completed_has_response
        CHECK (
            status <> 'completed'
            OR (response_status IS NOT NULL AND response_body IS NOT NULL)
        );

ALTER TABLE metaldocs.idempotency_keys
    DROP CONSTRAINT IF EXISTS idempotency_keys_response_status_range;
ALTER TABLE metaldocs.idempotency_keys
    ADD CONSTRAINT idempotency_keys_response_status_range
        CHECK (
            response_status IS NULL
            OR response_status BETWEEN 100 AND 599
        );

ALTER TABLE metaldocs.idempotency_keys
    DROP CONSTRAINT IF EXISTS idempotency_keys_response_body_size;
ALTER TABLE metaldocs.idempotency_keys
    ADD CONSTRAINT idempotency_keys_response_body_size
        CHECK (
            response_body IS NULL
            OR octet_length(response_body) <= 65536
        );

ALTER TABLE metaldocs.idempotency_keys
    DROP CONSTRAINT IF EXISTS idempotency_keys_actor_nonempty;
ALTER TABLE metaldocs.idempotency_keys
    ADD CONSTRAINT idempotency_keys_actor_nonempty
        CHECK (actor_user_id <> '');

CREATE INDEX IF NOT EXISTS idx_idempotency_keys_in_flight_expires
    ON metaldocs.idempotency_keys (expires_at)
    WHERE status = 'in_flight';

INSERT INTO public.schema_migrations(version, description)
VALUES ('0212', 'idempotency keys two-phase schema alignment for curated baseline')
ON CONFLICT (version) DO NOTHING;

COMMIT;
