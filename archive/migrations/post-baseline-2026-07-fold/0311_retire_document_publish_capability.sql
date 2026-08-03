-- 0311_retire_document_publish_capability.sql
-- ADR 0085 Stage B — retire the `document.publish` capability.
--
-- Publication stopped being a human-invoked action in Stage A: the release
-- coordinator releases an approved revision once every readiness fact holds
-- (approval fact x artifact fact x effective-date gate x supersession head).
-- With POST /documents/{id}/publish, /schedule-publish and /supersede deleted
-- in Stage B, no route and no service asserts `document.publish` any more, and
-- the capability is gone from the Go registry
-- (internal/modules/iam/domain/model.go) and from the curated reference-data
-- seed (db/reference-data/0001_product_reference_data.sql).
--
-- This migration removes the grants that already-bootstrapped databases carry,
-- so the DB agrees with the registry. Leaving them would be a live grant for a
-- capability nothing can assert: harmless today, but it re-arms the moment any
-- future code reuses the string, and it breaks seed<->registry parity.
--
-- `document.supersede` is deliberately NOT touched. ADR 0085 retains it and
-- re-homes it to submission time: naming a CROSS-document supersede target in
-- the publication plan requires it on the TARGET's process area, checked in
-- the submit transaction.
--
-- Scope: exactly one table, metaldocs.role_capabilities. There is no separate
-- capability catalog table — role_capabilities (role, capability, description)
-- is the only place a capability string is persisted.
--
-- No tripwire arm change is required: no arm in internal/platform/tripwire
-- carries `document.publish`, so the arm/registry parity lints stay green.
--
-- Stage B also deletes the `scheduled_publish_cutover` River job kind
-- (internal/modules/approval/jobs/scheduled_publish_args.go and its worker).
-- After deploy no worker registers that kind, so any row still parked in a
-- runnable state would be picked up by River's producer, fail to resolve a
-- worker, and churn as erroring/retryable noise forever. Those rows are strand
-- garbage the moment the binaries roll: the release coordinator owns the future
-- publication of every one of those documents, keyed on its own generation, so
-- replaying them is not merely impossible but wrong.
--
-- Idempotent: bare DELETEs by capability value / job kind; re-running is a
-- no-op.

BEGIN;

-- Role grants (seeded to area_admin and qms_admin before Stage B).
DELETE FROM metaldocs.role_capabilities
 WHERE capability = 'document.publish';

-- ── strand-safe retirement of the `scheduled_publish_cutover` River kind ────
--
-- River provisions its own schema at binary startup (rivermigrate, via
-- internal/platform/bootstrap.MigrateRiverSchema), NOT through this migration
-- series — so on a bare bootstrap public.river_job does not exist yet when this
-- file runs. to_regclass keeps that path a clean no-op instead of a 42P01.
--
-- Only NON-TERMINAL rows are removed. `completed`/`cancelled`/`discarded` rows
-- are finished history and stay: River will never run them again, and deleting
-- them would erase the audit trail of scheduled publications that actually
-- happened. `running` is deliberately excluded too — a row in that state
-- belongs to a live producer's lease; it drains on its own once the old binary
-- stops, and deleting it out from under a still-running worker is the one way
-- to turn a retired job into an error.
DO $$
BEGIN
  IF to_regclass('public.river_job') IS NOT NULL THEN
    -- state::text, not a bare enum comparison: `pending` only entered
    -- river_job_state in River's own migration 004, and an enum literal that
    -- does not exist in the installed type raises 22P02 instead of matching
    -- nothing. Comparing as text keeps this file safe against every River
    -- schema version.
    DELETE FROM public.river_job
     WHERE kind = 'scheduled_publish_cutover'
       AND state::text IN ('available', 'scheduled', 'retryable', 'pending');
  END IF;
END $$;

-- ── schema_migrations ledger ────────────────────────────────────────────────

INSERT INTO public.schema_migrations (version, description)
VALUES ('0311', 'ADR 0085 Stage B: retire the document.publish capability — delete every metaldocs.role_capabilities grant for it, matching its removal from the Go capability registry and the curated reference-data seed. Publication is no longer a human-invoked action (the release coordinator reacts to durable readiness facts), so no route or service asserts the capability. document.supersede is retained and re-homed to submit-time cross-document plan authorization. Also strand-safe: deletes every NON-TERMINAL public.river_job row of the deleted kind scheduled_publish_cutover (available/scheduled/retryable/pending), guarded by to_regclass because River provisions its own schema at binary startup rather than through this migration series; completed/cancelled/discarded history and leased running rows are left untouched.')
ON CONFLICT (version) DO NOTHING;

COMMIT;
