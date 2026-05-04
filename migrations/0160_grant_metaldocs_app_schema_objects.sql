-- migrations/0160_grant_metaldocs_app_schema_objects.sql
-- Grant metaldocs_app (the runtime DB user) access to metaldocs schema objects
-- needed by the background worker and the signoff idempotency store.
--
-- Root cause: migrations 0147–0149 granted
-- only to metaldocs_writer. The actual runtime user is metaldocs_app.

BEGIN;

GRANT USAGE ON SCHEMA metaldocs TO metaldocs_app;
GRANT USAGE ON SCHEMA metaldocs TO metaldocs_admin;

-- Job lease functions (background worker)
GRANT SELECT, INSERT, UPDATE, DELETE
    ON metaldocs.job_leases
    TO metaldocs_app;

-- USAGE grant is needed for SECURITY DEFINER functions owned by metaldocs_admin
-- to resolve schema objects in hardened environments where PUBLIC lacks schema USAGE.

GRANT EXECUTE ON FUNCTION metaldocs.acquire_lease(text, text, interval)   TO metaldocs_app;
GRANT EXECUTE ON FUNCTION metaldocs.heartbeat_lease(text, text, bigint)   TO metaldocs_app;
GRANT EXECUTE ON FUNCTION metaldocs.release_lease(text, text, bigint)     TO metaldocs_app;
GRANT EXECUTE ON FUNCTION metaldocs.assert_lease_epoch(text, bigint)      TO metaldocs_app;

-- Idempotency keys table (signoff replay store)
GRANT SELECT, INSERT, UPDATE
    ON metaldocs.idempotency_keys
    TO metaldocs_app;


COMMIT;
