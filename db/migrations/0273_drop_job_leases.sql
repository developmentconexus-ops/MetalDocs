-- 0273_drop_job_leases.sql
-- Retired by M5 async-River consolidation (ADR 0067): drops the custom lease
-- scheduler's DB objects. River owns leader/lease lifecycle in `river_leader`
-- now — there is nothing left in this codebase that calls acquire_lease,
-- heartbeat_lease, release_lease, or assert_lease_epoch, nor anything that
-- reads/writes metaldocs.job_leases. The custom lease scheduler + lease-reaper
-- Go code was deleted in a prior task in this program; this migration removes
-- the corresponding DB-side functions and table.
--
-- Included beyond the ticket's original 3-function list: metaldocs.assert_lease_
-- epoch(text, bigint) also depends solely on metaldocs.job_leases (queries it
-- directly for lease_epoch fencing) and has zero live callers outside a now-
-- deleted integration test (tests/integration/scenarios/concurrency_test.go's
-- testLeaseEpochFencing, dead scaffolding for the retired lease scheduler,
-- removed in the same change-set as this migration). Dropping it alongside the
-- other three keeps the DB free of orphaned lease-scheduler functions.
--
-- Drop order: functions first (no inbound dependencies once the Go callers are
-- gone), then the table. job_leases_pkey (job_name) drops automatically with
-- the table. No RLS, no triggers, no inbound FK from any other table (job_leases
-- is a tenant-global system table with no tenant_id column — confirmed in
-- scripts/api-lint/async-tenant-tables.txt commentary). IF EXISTS makes this
-- idempotent and safe to re-run.
--
-- Rollback note: not applicable — forward-only per db/migrations/README.md.
-- Restoring the lease scheduler would require re-adding both the Go code
-- (already deleted in a prior task) and these DB objects from
-- db/baseline/0001_current_schema.sql (functions ~lines 59, 137, 262, 300;
-- table ~line 1354).

BEGIN;

DROP FUNCTION IF EXISTS metaldocs.acquire_lease(text, text, interval);
DROP FUNCTION IF EXISTS metaldocs.heartbeat_lease(text, text, bigint);
DROP FUNCTION IF EXISTS metaldocs.release_lease(text, text, bigint);
DROP FUNCTION IF EXISTS metaldocs.assert_lease_epoch(text, bigint);

DROP TABLE IF EXISTS metaldocs.job_leases;

INSERT INTO public.schema_migrations (version, description)
VALUES ('0273', 'Drop metaldocs.job_leases and its acquire_lease/heartbeat_lease/release_lease/assert_lease_epoch functions — custom lease scheduler retired by M5 async-River consolidation (ADR 0067); River owns leader/lease lifecycle in river_leader.')
ON CONFLICT (version) DO NOTHING;

COMMIT;
