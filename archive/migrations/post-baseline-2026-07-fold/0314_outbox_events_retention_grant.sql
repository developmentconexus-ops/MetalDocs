-- 0314_outbox_events_retention_grant.sql
-- Grant the runtime role DELETE on metaldocs.outbox_events so the
-- outbox-events-retention maintenance job can purge terminal rows.
--
-- ============================================================================
-- WHY
-- ============================================================================
--   metaldocs.outbox_events had no retention job of any kind: nothing deleted
--   published or dead-lettered rows, so the table grew without bound AND every
--   terminal row pinned its idempotency_key forever against the UNIQUE
--   constraint that Publisher.Publish dedupes on
--   (internal/platform/messaging/outbox/postgres/publisher.go — INSERT ...
--   ON CONFLICT (idempotency_key) DO NOTHING). A later legitimate re-publish on
--   the same key was therefore silently swallowed.
--
--   The new outbox-events-retention job (internal/modules/jobs/outbox_retention,
--   executing in metaldocs-jobs per ADR 0067) closes the growth half. It needs
--   DELETE, which metaldocs_app never had: the only grants this table ever
--   received were INSERT (archive/migrations/0008_grant_outbox_privileges.sql)
--   and SELECT, UPDATE (archive/migrations/0019_grant_worker_runtime_privileges.sql).
--   Without this grant the purge fails closed with "permission denied for table
--   outbox_events" in any environment whose runtime role is not the table owner
--   — i.e. everywhere except the local dev database, where metaldocs_app is a
--   superuser and the missing grant is invisible.
--
-- ============================================================================
-- SCOPE
-- ============================================================================
--   DELETE only, on this one table, for the runtime role. It does NOT widen
--   what may be deleted: the eligibility predicate (published_at IS NOT NULL
--   AND dead_lettered_at IS NULL, or dead_lettered_at IS NOT NULL) lives in
--   internal/platform/messaging/outbox/postgres/retention.go and is what keeps
--   unpublished, non-dead-lettered rows unreachable by the purge.
--
--   Role-existence guarded, following
--   archive/migrations/0160_grant_metaldocs_app_schema_objects.sql: the role is
--   provisioned outside the migration stream and is absent in some bootstrap
--   paths (notably the curated-baseline test template), where this is a no-op.
--
--   ROLLBACK: REVOKE DELETE ON TABLE metaldocs.outbox_events FROM metaldocs_app;
--   The retention job then fails closed on every tick (permission denied) and
--   the table resumes growing without bound; nothing else regresses.

BEGIN;

DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'metaldocs_app') THEN
        GRANT DELETE ON TABLE metaldocs.outbox_events TO metaldocs_app;
    END IF;
END $$;

INSERT INTO public.schema_migrations (version, description)
VALUES ('0314', 'Grant DELETE on metaldocs.outbox_events to metaldocs_app so the new outbox-events-retention maintenance job (internal/modules/jobs/outbox_retention, executed by metaldocs-jobs per ADR 0067) can purge terminal rows. The table previously had NO retention at all, so published/dead-lettered rows accumulated forever and each one pinned its idempotency_key against the publisher''s ON CONFLICT (idempotency_key) DO NOTHING dedup, silently swallowing later legitimate re-publishes on the same key. Prior grants were INSERT (0008) and SELECT, UPDATE (0019) only; DELETE was missing, which would have failed the purge closed in any environment where metaldocs_app is not the table owner.')
ON CONFLICT (version) DO NOTHING;

COMMIT;
