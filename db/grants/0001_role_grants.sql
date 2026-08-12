-- MetalDocs role + privilege posture.
--
-- Bootstrap stage 4, applied AFTER db/baseline/0001_current_schema.sql and
-- db/reference-data/0001_product_reference_data.sql (the GRANT ... ON ALL
-- TABLES statements below need every table to already exist), and AFTER
-- 0000_identity_roles.sql (the CREATE ROLE + ownership-transfer stage this
-- file's GRANT/REVOKE statements depend on).
--
-- Role CREATE statements used to live in this file. As of issue #88 / A6.1's
-- re-cut they live in 0000_identity_roles.sql instead, and this file only
-- grants/revokes privileges on an ALREADY-CREATED role -- see that file for
-- why the split.
--
-- Ownership/execution identity (A6.1 re-cut): this file is applied ONLY by
-- apps/dbprovision/cmd/metaldocs-dbprovision (the one-shot provisioning
-- binary), connected as the bootstrap superuser, once per environment
-- start -- NOT at every metaldocs-api startup any more. Every GRANT/REVOKE
-- below still degrades cleanly (skip, never error) when a named grantee role
-- is absent, so this file stays safe to re-run by hand against an older
-- volume mid-migration.
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
-- Role creation lives in 0000_identity_roles.sql now; this block only grants.
-- Conditional on the role existing so a partial/older environment (grants
-- file re-run by hand before 0000 has ever run) skips cleanly with a NOTICE
-- instead of failing the whole transaction over a test-only role.
DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'metaldocs_ci') THEN
    RAISE NOTICE 'metaldocs_ci role does not exist -- skipping CI grants (expected until 0000_identity_roles.sql has run)';
    RETURN;
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

  -- Future objects created by the app/owner role inherit the same grants.
  -- Both defaults are set: metaldocs_app is the legacy pre-A6.1 identity that
  -- may still own objects on an in-place upgraded volume until this file's
  -- ownership transfer (0000_identity_roles.sql) has run against it;
  -- metaldocs_owner is the identity that creates every object from A6.1
  -- onward (db/migrations run under SET ROLE metaldocs_owner -- see
  -- apps/dbprovision). Setting both means a table created either identity's
  -- way still auto-grants to metaldocs_ci.
  IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'metaldocs_app') THEN
    EXECUTE 'ALTER DEFAULT PRIVILEGES FOR ROLE metaldocs_app IN SCHEMA metaldocs GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO metaldocs_ci';
    EXECUTE 'ALTER DEFAULT PRIVILEGES FOR ROLE metaldocs_app IN SCHEMA public   GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO metaldocs_ci';
    EXECUTE 'ALTER DEFAULT PRIVILEGES FOR ROLE metaldocs_app IN SCHEMA metaldocs GRANT USAGE, SELECT ON SEQUENCES TO metaldocs_ci';
    EXECUTE 'ALTER DEFAULT PRIVILEGES FOR ROLE metaldocs_app IN SCHEMA public   GRANT USAGE, SELECT ON SEQUENCES TO metaldocs_ci';
  END IF;
  IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'metaldocs_owner') THEN
    EXECUTE 'ALTER DEFAULT PRIVILEGES FOR ROLE metaldocs_owner IN SCHEMA metaldocs GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO metaldocs_ci';
    EXECUTE 'ALTER DEFAULT PRIVILEGES FOR ROLE metaldocs_owner IN SCHEMA public   GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO metaldocs_ci';
    EXECUTE 'ALTER DEFAULT PRIVILEGES FOR ROLE metaldocs_owner IN SCHEMA metaldocs GRANT USAGE, SELECT ON SEQUENCES TO metaldocs_ci';
    EXECUTE 'ALTER DEFAULT PRIVILEGES FOR ROLE metaldocs_owner IN SCHEMA public   GRANT USAGE, SELECT ON SEQUENCES TO metaldocs_ci';
  END IF;

  -- CURRENT_USER (CI-only gap, issue #88 A6.1 follow-up): the two blocks
  -- above only cover objects later created BY metaldocs_app or
  -- metaldocs_owner BY NAME. That is complete for apps/dbprovision's
  -- production path (Stage 2/this file always runs as the bootstrap
  -- superuser, and Stage 3 DDL always runs under SET ROLE metaldocs_owner --
  -- see apps/dbprovision/cmd/metaldocs-dbprovision/main.go), and for local
  -- dev (docker-compose's bootstrap superuser is literally named
  -- metaldocs_app). It is NOT complete for
  -- tests/integration/testdb.ApplyCuratedBootstrap, which applies this same
  -- file plus the forward db/migrations/ tail back-to-back via a single
  -- connection with no SET ROLE -- so any migration-tail table is created
  -- by whatever role DATABASE_URL names. In CI that role is "metaldocs"
  -- (.github/workflows/ci.yml POSTGRES_USER), a name neither hardcoded
  -- block above ever mentions. Without this, a forward migration adding a
  -- table reproduces SQLSTATE 42501 for metaldocs_ci/metaldocs_runtime in
  -- CI only, never on a developer machine -- found and diagnosed by lane B
  -- (PR #111) while wiring #112's generated handler bindings; out of their
  -- write-set (db/grants/**), handed off here. CURRENT_USER covers every
  -- caller in one block: metaldocs_app locally, metaldocs_owner mid-Stage-3,
  -- metaldocs (or any other name) in CI, with no environment-specific
  -- enumeration to keep in sync.
  EXECUTE format('ALTER DEFAULT PRIVILEGES FOR ROLE %I IN SCHEMA metaldocs GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO metaldocs_ci', current_user);
  EXECUTE format('ALTER DEFAULT PRIVILEGES FOR ROLE %I IN SCHEMA public   GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO metaldocs_ci', current_user);
  EXECUTE format('ALTER DEFAULT PRIVILEGES FOR ROLE %I IN SCHEMA metaldocs GRANT USAGE, SELECT ON SEQUENCES TO metaldocs_ci', current_user);
  EXECUTE format('ALTER DEFAULT PRIVILEGES FOR ROLE %I IN SCHEMA public   GRANT USAGE, SELECT ON SEQUENCES TO metaldocs_ci', current_user);
END
$$;

-- ── metaldocs_runtime: non-owner, non-bypass application role (A6.1) ────────
-- Role creation lives in 0000_identity_roles.sql now; this block only grants.
-- Conditional on the role existing so a partial/older environment (grants
-- file re-run by hand before 0000 has ever run) skips cleanly with a NOTICE.
--
-- Grant surface mirrors metaldocs_ci's DML posture (SELECT/INSERT/UPDATE/
-- DELETE on every existing and future table in both schemas, USAGE/SELECT on
-- sequences) but, unlike metaldocs_ci, also inherits the audit_events and
-- outbox_events hardening below -- metaldocs_runtime is the identity
-- metaldocs-api/worker/jobs connect as from A6.1 onward, where the
-- audit-immutability guarantee must hold; metaldocs_ci is a test role that
-- deliberately keeps full audit_events DML for audit-chain test assertions.
-- Deliberately NOCREATEDB NOCREATEROLE and granted no DDL: this role cannot
-- run schema migrations, and never owns any table (see 0000_identity_roles.sql
-- header) -- migrations run as metaldocs_owner via apps/dbprovision.
DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'metaldocs_runtime') THEN
    RAISE NOTICE 'metaldocs_runtime role does not exist -- skipping runtime grants (expected until 0000_identity_roles.sql has run)';
    RETURN;
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
  -- metaldocs_owner is the identity that creates every object from A6.1
  -- onward (db/migrations run under SET ROLE metaldocs_owner). Without this,
  -- a table added by a future forward migration would not auto-grant DML to
  -- metaldocs_runtime, and the API would start failing closed against its
  -- own new tables.
  IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'metaldocs_owner') THEN
    EXECUTE 'ALTER DEFAULT PRIVILEGES FOR ROLE metaldocs_owner IN SCHEMA metaldocs GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO metaldocs_runtime';
    EXECUTE 'ALTER DEFAULT PRIVILEGES FOR ROLE metaldocs_owner IN SCHEMA public   GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO metaldocs_runtime';
    EXECUTE 'ALTER DEFAULT PRIVILEGES FOR ROLE metaldocs_owner IN SCHEMA metaldocs GRANT USAGE, SELECT ON SEQUENCES TO metaldocs_runtime';
    EXECUTE 'ALTER DEFAULT PRIVILEGES FOR ROLE metaldocs_owner IN SCHEMA public   GRANT USAGE, SELECT ON SEQUENCES TO metaldocs_runtime';
  END IF;

  -- CURRENT_USER (CI-only gap): same rationale as the matching block in the
  -- metaldocs_ci DO block above -- see that comment.
  EXECUTE format('ALTER DEFAULT PRIVILEGES FOR ROLE %I IN SCHEMA metaldocs GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO metaldocs_runtime', current_user);
  EXECUTE format('ALTER DEFAULT PRIVILEGES FOR ROLE %I IN SCHEMA public   GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO metaldocs_runtime', current_user);
  EXECUTE format('ALTER DEFAULT PRIVILEGES FOR ROLE %I IN SCHEMA metaldocs GRANT USAGE, SELECT ON SEQUENCES TO metaldocs_runtime', current_user);
  EXECUTE format('ALTER DEFAULT PRIVILEGES FOR ROLE %I IN SCHEMA public   GRANT USAGE, SELECT ON SEQUENCES TO metaldocs_runtime', current_user);
END
$$;

-- ── audit_events insert/select-only hardening (from 0266 part a) ────────────
-- Ordering constraint for this whole file: it must run AFTER the baseline (and
-- reference data), because every GRANT/REVOKE below names tables that must
-- already exist. Ordering WITHIN the file is NOT free for metaldocs_runtime:
-- the blanket "GRANT ... ON ALL TABLES" above (metaldocs_runtime block)
-- includes metaldocs.audit_events, and this REVOKE strips UPDATE/DELETE/
-- TRUNCATE back off it. This block MUST stay AFTER the metaldocs_runtime
-- grant block; if it is ever moved above it, the grant re-adds the
-- privileges and the audit hardening is silently lost. metaldocs_ci is
-- unaffected by ordering -- it deliberately keeps DML on audit_events (it is
-- a test role and audit-chain tests need it); metaldocs_app and
-- metaldocs_runtime must never mutate or truncate the hash chain.
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
