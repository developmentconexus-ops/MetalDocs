-- MetalDocs role + privilege posture.
--
-- Bootstrap stage 4, applied AFTER db/baseline/0001_current_schema.sql and
-- db/reference-data/0001_product_reference_data.sql (the GRANT ... ON ALL
-- TABLES statements below need every table to already exist).
--
-- WHY THIS FILE EXISTS
-- The curated baseline is regenerated with `pg_dump --schema-only --no-owner
-- --no-privileges`, which deliberately carries no ACLs, and role creation is
-- cluster-global so pg_dump never emits it at all. Before the 2026-07-29 fold
-- the privilege posture was supplied by three forward migrations:
--   - 0266_audit_events_hardening.sql (a)  REVOKE UPDATE/DELETE/TRUNCATE on
--     metaldocs.audit_events from the app role -- durably hardens the
--     insert/select-only audit posture (T-004/T-009).
--   - 0284_ci_rls_role.sql                 create the dedicated non-owner
--     NOSUPERUSER + NOBYPASSRLS metaldocs_ci role with DML-only grants, so the
--     integration suite exercises RLS for real instead of false-greening under
--     the owner/superuser app role (M7 F7.4; tests/integration/testdb/ci_role.go).
--   - 0314_outbox_events_retention_grant.sql  GRANT DELETE on
--     metaldocs.outbox_events to the app role so the outbox-retention
--     maintenance job can purge terminal rows (ADR 0067).
-- Folding those migrations into the schema baseline would have silently
-- dropped all three effects. They are re-homed here verbatim, in the same
-- guarded/idempotent form, so a fresh bootstrap reproduces the exact posture
-- of a fully-migrated database.
--
-- Everything here is idempotent and safe to re-run.

BEGIN;

-- ── metaldocs_ci: non-owner, non-bypass DML role (from 0284) ────────────────
-- Roles are cluster-global, so this is a no-op on a cluster that already has
-- it. The dev password is a non-secret DML-only fixture; a deployment rotates
-- it with `ALTER ROLE metaldocs_ci PASSWORD '<deployment-secret>'` and points
-- the suite at it via METALDOCS_CI_DB_PASSWORD.
DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'metaldocs_ci') THEN
    EXECUTE 'CREATE ROLE metaldocs_ci NOSUPERUSER NOBYPASSRLS NOCREATEDB NOCREATEROLE NOINHERIT LOGIN PASSWORD ''metaldocs_ci_dev''';
  END IF;
END
$$;

-- Schema access.
GRANT USAGE ON SCHEMA metaldocs TO metaldocs_ci;
GRANT USAGE ON SCHEMA public   TO metaldocs_ci;

-- DML on all existing tables in both schemas (covers every tenant-scoped
-- table + everything else -- the app role reads/writes ordinary tables too;
-- RLS is what enforces tenant isolation, not the grant surface).
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA metaldocs TO metaldocs_ci;
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public   TO metaldocs_ci;

-- Sequences (serial/identity defaults on INSERT).
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA metaldocs TO metaldocs_ci;
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public   TO metaldocs_ci;

-- Future objects created by the app role inherit the same grants.
DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'metaldocs_app') THEN
    EXECUTE 'ALTER DEFAULT PRIVILEGES FOR ROLE metaldocs_app IN SCHEMA metaldocs GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO metaldocs_ci';
    EXECUTE 'ALTER DEFAULT PRIVILEGES FOR ROLE metaldocs_app IN SCHEMA public   GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO metaldocs_ci';
    EXECUTE 'ALTER DEFAULT PRIVILEGES FOR ROLE metaldocs_app IN SCHEMA metaldocs GRANT USAGE, SELECT ON SEQUENCES TO metaldocs_ci';
    EXECUTE 'ALTER DEFAULT PRIVILEGES FOR ROLE metaldocs_app IN SCHEMA public   GRANT USAGE, SELECT ON SEQUENCES TO metaldocs_ci';
  END IF;
END
$$;

-- ── audit_events insert/select-only hardening (from 0266 part a) ────────────
-- Runs AFTER the metaldocs_ci block on purpose: the blanket
-- "GRANT ... ON ALL TABLES" above would otherwise re-grant UPDATE/DELETE on
-- audit_events to metaldocs_ci. metaldocs_ci keeps DML there (it is a test
-- role and audit-chain tests need it); the revoke targets the app role, which
-- must never mutate or truncate the hash chain.
DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'metaldocs_app') THEN
    EXECUTE 'REVOKE UPDATE, DELETE, TRUNCATE ON TABLE metaldocs.audit_events FROM metaldocs_app';
  END IF;
END
$$;

-- ── outbox retention purge (from 0314) ──────────────────────────────────────
-- The outbox-events-retention maintenance job (internal/modules/jobs/outbox_retention,
-- executed by metaldocs-jobs per ADR 0067) DELETEs terminal rows. Prior grants
-- on the table were INSERT (0008) and SELECT, UPDATE (0019) only; without
-- DELETE the purge fails closed wherever metaldocs_app is not the table owner.
DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'metaldocs_app') THEN
    EXECUTE 'GRANT DELETE ON TABLE metaldocs.outbox_events TO metaldocs_app';
  END IF;
END
$$;

COMMIT;
