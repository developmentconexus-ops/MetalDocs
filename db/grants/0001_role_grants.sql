-- MetalDocs role + privilege posture.
--
-- Bootstrap stage 4, applied AFTER db/baseline/0001_current_schema.sql and
-- db/reference-data/0001_product_reference_data.sql (the GRANT ... ON ALL
-- TABLES statements below need every table to already exist).
--
-- ALSO re-applied unconditionally at every metaldocs-api startup, before
-- migrations, by internal/platform/migrate.ApplyGrants — no schema_migrations
-- ledger row, by design. Fresh-bootstrap-only application meant an edit to this
-- file never reached a long-lived volume. Every statement below is therefore
-- required to be idempotent AND to degrade cleanly (skip, never error) when the
-- executing role lacks CREATEROLE or does not own the target objects.
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
-- metaldocs_runtime (added for issue #88 / axis A6.1): today
-- deploy/compose/docker-compose.yml passes ${POSTGRES_USER} -- the Postgres
-- image's bootstrap superuser (also BYPASSRLS by construction) -- as PGUSER
-- to metaldocs-api/worker/jobs, so the boot-fatal identity assertion added by
-- A6.1 (internal/platform/db/postgres.AssertSafeIdentity) refuses to boot
-- against the dev compose's current PGUSER. This block provisions a
-- dedicated, non-superuser, non-bypassrls role that CAN satisfy that
-- assertion, so the assertion mechanism itself is provably correct and
-- testable. It is NOT yet wired into any compose/env default -- pointing
-- PGUSER at it, and resolving how schema migrations (which need DDL rights
-- this role deliberately does not have) run under a least-privilege identity,
-- is issue #88's A6.2 ("app role provisioning in compose"), out of A6.1's
-- scope. Until A6.2 lands, this role sits provisioned-but-unused, exactly
-- like metaldocs_ci sat between 0284 and its first consumer.
--
-- Everything here is idempotent and safe to re-run.

BEGIN;

-- ── metaldocs_ci: non-owner, non-bypass DML role (from 0284) ────────────────
-- Roles are cluster-global, so the CREATE is a no-op on a cluster that already
-- has it. The dev password is a non-secret DML-only fixture; a deployment
-- rotates it with `ALTER ROLE metaldocs_ci PASSWORD '<deployment-secret>'` and
-- points the suite at it via METALDOCS_CI_DB_PASSWORD. Re-running this file
-- NEVER resets an already-rotated password: the CREATE only runs when the role
-- is absent.
--
-- The whole block (create + grants) is conditional on the role existing or
-- being creatable. Under ApplyGrants this file runs at every API startup, and
-- in an environment where the app's DB user lacks CREATEROLE and the CI role
-- was never provisioned, a bare CREATE ROLE -- or the GRANTs that name a
-- non-existent grantee -- would turn startup into a hard failure over a
-- test-only role. That case skips cleanly with a NOTICE. It does not swallow
-- anything else: any other error still aborts the file (and the transaction).
DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'metaldocs_ci') THEN
    IF NOT EXISTS (
      SELECT 1 FROM pg_roles
       WHERE rolname = current_user AND (rolcreaterole OR rolsuper)
    ) THEN
      RAISE NOTICE 'metaldocs_ci is absent and % lacks CREATEROLE -- skipping CI role provisioning', current_user;
      RETURN;
    END IF;
    EXECUTE 'CREATE ROLE metaldocs_ci NOSUPERUSER NOBYPASSRLS NOCREATEDB NOCREATEROLE NOINHERIT LOGIN PASSWORD ''metaldocs_ci_dev''';
  END IF;

  -- Schema access.
  EXECUTE 'GRANT USAGE ON SCHEMA metaldocs TO metaldocs_ci';
  EXECUTE 'GRANT USAGE ON SCHEMA public   TO metaldocs_ci';

  -- DML on all existing tables in both schemas (covers every tenant-scoped
  -- table + everything else -- the app role reads/writes ordinary tables too;
  -- RLS is what enforces tenant isolation, not the grant surface).
  EXECUTE 'GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA metaldocs TO metaldocs_ci';
  EXECUTE 'GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public   TO metaldocs_ci';

  -- Sequences (serial/identity defaults on INSERT).
  EXECUTE 'GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA metaldocs TO metaldocs_ci';
  EXECUTE 'GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public   TO metaldocs_ci';

  -- Future objects created by the app role inherit the same grants.
  IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'metaldocs_app') THEN
    EXECUTE 'ALTER DEFAULT PRIVILEGES FOR ROLE metaldocs_app IN SCHEMA metaldocs GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO metaldocs_ci';
    EXECUTE 'ALTER DEFAULT PRIVILEGES FOR ROLE metaldocs_app IN SCHEMA public   GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO metaldocs_ci';
    EXECUTE 'ALTER DEFAULT PRIVILEGES FOR ROLE metaldocs_app IN SCHEMA metaldocs GRANT USAGE, SELECT ON SEQUENCES TO metaldocs_ci';
    EXECUTE 'ALTER DEFAULT PRIVILEGES FOR ROLE metaldocs_app IN SCHEMA public   GRANT USAGE, SELECT ON SEQUENCES TO metaldocs_ci';
  END IF;
END
$$;

-- ── metaldocs_runtime: non-owner, non-bypass application role (A6.1) ────────
-- Same guarded/idempotent shape as the metaldocs_ci block above: the CREATE
-- only runs when the role is absent (re-running this file never resets an
-- already-rotated password), and the whole block skips cleanly with a NOTICE
-- when the executing role lacks CREATEROLE, instead of turning startup into a
-- hard failure over a role no compose/env currently connects as.
--
-- Grant surface mirrors metaldocs_ci's DML posture (SELECT/INSERT/UPDATE/
-- DELETE on every existing and future table in both schemas, USAGE/SELECT on
-- sequences) but, unlike metaldocs_ci, also inherits the audit_events and
-- outbox_events hardening below -- metaldocs_runtime is meant to eventually
-- stand in for metaldocs_app on the request-serving path (A6.2), where the
-- audit-immutability guarantee must hold; metaldocs_ci is a test role that
-- deliberately keeps full audit_events DML for audit-chain test assertions.
-- Deliberately NOCREATEDB NOCREATEROLE and granted no DDL: this role cannot
-- run schema migrations. Which identity runs migrate.Apply/ApplyGrants once
-- compose stops connecting as the superuser owner is an open question left to
-- A6.2, not decided here.
DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'metaldocs_runtime') THEN
    IF NOT EXISTS (
      SELECT 1 FROM pg_roles
       WHERE rolname = current_user AND (rolcreaterole OR rolsuper)
    ) THEN
      RAISE NOTICE 'metaldocs_runtime is absent and % lacks CREATEROLE -- skipping runtime role provisioning', current_user;
      RETURN;
    END IF;
    EXECUTE 'CREATE ROLE metaldocs_runtime NOSUPERUSER NOBYPASSRLS NOCREATEDB NOCREATEROLE NOINHERIT LOGIN PASSWORD ''metaldocs_runtime_dev''';
  END IF;

  EXECUTE 'GRANT USAGE ON SCHEMA metaldocs TO metaldocs_runtime';
  EXECUTE 'GRANT USAGE ON SCHEMA public   TO metaldocs_runtime';

  EXECUTE 'GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA metaldocs TO metaldocs_runtime';
  EXECUTE 'GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public   TO metaldocs_runtime';

  EXECUTE 'GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA metaldocs TO metaldocs_runtime';
  EXECUTE 'GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public   TO metaldocs_runtime';

  IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'metaldocs_app') THEN
    EXECUTE 'ALTER DEFAULT PRIVILEGES FOR ROLE metaldocs_app IN SCHEMA metaldocs GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO metaldocs_runtime';
    EXECUTE 'ALTER DEFAULT PRIVILEGES FOR ROLE metaldocs_app IN SCHEMA public   GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO metaldocs_runtime';
    EXECUTE 'ALTER DEFAULT PRIVILEGES FOR ROLE metaldocs_app IN SCHEMA metaldocs GRANT USAGE, SELECT ON SEQUENCES TO metaldocs_runtime';
    EXECUTE 'ALTER DEFAULT PRIVILEGES FOR ROLE metaldocs_app IN SCHEMA public   GRANT USAGE, SELECT ON SEQUENCES TO metaldocs_runtime';
  END IF;
END
$$;

-- ── audit_events insert/select-only hardening (from 0266 part a) ────────────
-- Ordering constraint for this whole file: it must run AFTER the baseline (and
-- reference data), because every GRANT/REVOKE below names tables that must
-- already exist. Ordering WITHIN the file is free -- the blanket
-- "GRANT ... ON ALL TABLES" above targets metaldocs_ci/metaldocs_runtime,
-- while this REVOKE targets metaldocs_app/metaldocs_runtime, so they never
-- touch the same (grantee, object) pair and neither can undo the other.
-- metaldocs_ci deliberately keeps DML on audit_events (it is a test role and
-- audit-chain tests need it); metaldocs_app and metaldocs_runtime must never
-- mutate or truncate the hash chain.
DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'metaldocs_app') THEN
    EXECUTE 'REVOKE UPDATE, DELETE, TRUNCATE ON TABLE metaldocs.audit_events FROM metaldocs_app';
  END IF;
  IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'metaldocs_runtime') THEN
    EXECUTE 'REVOKE UPDATE, DELETE, TRUNCATE ON TABLE metaldocs.audit_events FROM metaldocs_runtime';
  END IF;
END
$$;

-- ── outbox retention purge (from 0314) ──────────────────────────────────────
-- The outbox-events-retention maintenance job (internal/modules/jobs/outbox_retention,
-- executed by metaldocs-jobs per ADR 0067) DELETEs terminal rows. Prior grants
-- on the table were INSERT (0008) and SELECT, UPDATE (0019) only; without
-- DELETE the purge fails closed wherever metaldocs_app is not the table owner.
-- metaldocs_runtime gets the same grant so it can eventually stand in for
-- metaldocs_app on the jobs binary too (A6.2).
DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'metaldocs_app') THEN
    EXECUTE 'GRANT DELETE ON TABLE metaldocs.outbox_events TO metaldocs_app';
  END IF;
  IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'metaldocs_runtime') THEN
    EXECUTE 'GRANT DELETE ON TABLE metaldocs.outbox_events TO metaldocs_runtime';
  END IF;
END
$$;

COMMIT;
