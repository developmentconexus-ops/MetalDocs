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

  -- Future objects created by WHICHEVER ROLE IS RUNNING THIS FILE, keyed
  -- dynamically via current_user rather than the literal name 'metaldocs_app'
  -- above. This closes a real gap, not a hypothetical one: A8.1 (migration
  -- 0318) creates metaldocs.capability_bindings and metaldocs.roles, and both
  -- db/grants/0001_role_grants.sql (this file) and internal/platform/migrate.
  -- ApplyGrants's own doc comment establish that the grants stage ALWAYS runs
  -- BEFORE db/migrations, every single bootstrap/startup, with no second pass
  -- in between -- so the blanket "ON ALL TABLES" grant above can never see a
  -- table a migration is about to create moments later in the same run. The
  -- metaldocs_app block above is the intended forward-declaring fix for
  -- exactly that ordering gap (ALTER DEFAULT PRIVILEGES applies to objects
  -- created AFTER it runs, by the named role, regardless of order-of-existence),
  -- but it is keyed to the LITERAL role name 'metaldocs_app' -- correct for a
  -- real deployment (.env.example: POSTGRES_USER=metaldocs_app), wrong for
  -- CI's throwaway Postgres service, which names its connecting/bootstrapping/
  -- migration-running role "metaldocs" (POSTGRES_USER in
  -- .github/workflows/ci.yml) and never creates a role literally named
  -- metaldocs_app at all -- so the IF EXISTS guard above is false and the
  -- whole block is skipped there. That mismatch is what let
  -- TestRLSTruth_CapabilityBindingsEnforcesIsolation pass against a local dev
  -- DB (role genuinely named metaldocs_app) while failing 42501 permission
  -- denied for table capability_bindings against a fresh CI database (role
  -- named metaldocs). current_user is always exactly the role connected right
  -- now -- the same one that will run the migrations directly after this file,
  -- in every environment -- so this statement needs no name-existence guard
  -- and no environment-specific branching: altering your OWN default
  -- privileges is always permitted.
  EXECUTE format('ALTER DEFAULT PRIVILEGES FOR ROLE %I IN SCHEMA metaldocs GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO metaldocs_ci', current_user);
  EXECUTE format('ALTER DEFAULT PRIVILEGES FOR ROLE %I IN SCHEMA public   GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO metaldocs_ci', current_user);
  EXECUTE format('ALTER DEFAULT PRIVILEGES FOR ROLE %I IN SCHEMA metaldocs GRANT USAGE, SELECT ON SEQUENCES TO metaldocs_ci', current_user);
  EXECUTE format('ALTER DEFAULT PRIVILEGES FOR ROLE %I IN SCHEMA public   GRANT USAGE, SELECT ON SEQUENCES TO metaldocs_ci', current_user);
END
$$;

-- ── audit_events insert/select-only hardening (from 0266 part a) ────────────
-- Ordering constraint for this whole file: it must run AFTER the baseline (and
-- reference data), because every GRANT/REVOKE below names tables that must
-- already exist. Ordering WITHIN the file is free -- the blanket
-- "GRANT ... ON ALL TABLES" above targets metaldocs_ci, while this REVOKE
-- targets metaldocs_app, so the two never touch the same (grantee, object)
-- pair and neither can undo the other. metaldocs_ci deliberately keeps DML on
-- audit_events (it is a test role and audit-chain tests need it); the app role
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
