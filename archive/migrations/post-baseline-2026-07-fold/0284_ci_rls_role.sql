-- 0284_ci_rls_role.sql
-- M7 F7.4 (global-maximum-remediation, milestone-7-tenant-lifecycle,
-- validation-contract.md §4.1/§4.2): create a dedicated, NON-OWNER,
-- NOSUPERUSER + NOBYPASSRLS application/CI database role so the integration
-- suite exercises RLS *for real* instead of running under the dev
-- metaldocs_app superuser (SUPERUSER + BYPASSRLS + owner of every table →
-- RLS triply inert → false-green isolation tests). This is the prod-faithful
-- posture: in production the app role is already NOSUPERUSER + NOBYPASSRLS and
-- a non-owner (baseline "Owner: -"); metaldocs_app stays the dev/bootstrap
-- owner identity.
--
-- DEVIATION (surfaced for HS-1, spec.md interview delta 1): the F7.4 contract
-- §4.2 suggests *reassigning* ownership of all 63 tables to a distinct owner
-- role. We instead add a fresh role that OWNS NOTHING (non-owner by
-- construction) with DML-only grants. This upholds §4.2's actual requirement
-- (the connecting proof role is a non-owner under NOBYPASSRLS — RLS applies to
-- it under plain ENABLE; FORCE is only needed for a table's owner) without a
-- schema-wide ownership migration that would break dev bootstrap / migrations
-- / the leader-elected janitors, all of which run as metaldocs_app.
--
-- Password: the dev/CI value below is a NON-SECRET fixture for a DML-only,
-- non-superuser, non-bypass role — same trust posture as the dev metaldocs_app
-- password. It is NOT an application secret and is deliberately not sourced
-- from .env (the integration harness defaults METALDOCS_CI_DB_PASSWORD to this
-- exact literal). Production MUST override it out-of-band:
--   ALTER ROLE metaldocs_ci PASSWORD '<deployment-secret>';   -- from a vault, never committed
--
-- Grants are baked into the template database when the integration harness
-- rebuilds it, so every `CREATE DATABASE … TEMPLATE` per-test clone inherits
-- them. ALTER DEFAULT PRIVILEGES covers tables/sequences created by future
-- migrations (which run as metaldocs_app).
--
-- Idempotent (guarded CREATE ROLE + IF NOT EXISTS-style grants); safe to
-- re-run. Down (manual, if ever needed):
--   REVOKE ALL ON ALL TABLES IN SCHEMA metaldocs, public FROM metaldocs_ci;
--   REVOKE ALL ON ALL SEQUENCES IN SCHEMA metaldocs, public FROM metaldocs_ci;
--   REVOKE USAGE ON SCHEMA metaldocs, public FROM metaldocs_ci;
--   ALTER DEFAULT PRIVILEGES FOR ROLE metaldocs_app IN SCHEMA metaldocs, public REVOKE ALL ON TABLES FROM metaldocs_ci;
--   DROP ROLE metaldocs_ci;   -- (roles are cluster-global)

BEGIN;

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
-- table + everything else — the app role reads/writes ordinary tables too;
-- RLS is what enforces tenant isolation, not the grant surface).
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA metaldocs TO metaldocs_ci;
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public   TO metaldocs_ci;

-- Sequences (serial/identity defaults on INSERT).
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA metaldocs TO metaldocs_ci;
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public   TO metaldocs_ci;

-- Future objects created by metaldocs_app inherit the same grants.
ALTER DEFAULT PRIVILEGES FOR ROLE metaldocs_app IN SCHEMA metaldocs
  GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO metaldocs_ci;
ALTER DEFAULT PRIVILEGES FOR ROLE metaldocs_app IN SCHEMA public
  GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO metaldocs_ci;
ALTER DEFAULT PRIVILEGES FOR ROLE metaldocs_app IN SCHEMA metaldocs
  GRANT USAGE, SELECT ON SEQUENCES TO metaldocs_ci;
ALTER DEFAULT PRIVILEGES FOR ROLE metaldocs_app IN SCHEMA public
  GRANT USAGE, SELECT ON SEQUENCES TO metaldocs_ci;

INSERT INTO public.schema_migrations (version, description)
VALUES ('0284', 'M7 F7.4 (validation-contract §4.1/§4.2): create dedicated non-owner NOSUPERUSER+NOBYPASSRLS metaldocs_ci role with DML-only grants (USAGE + SELECT/INSERT/UPDATE/DELETE on all tables + sequences in metaldocs/public, ALTER DEFAULT PRIVILEGES for future objects) so the integration suite exercises RLS for real. DEVIATION (HS-1): dedicated non-owner role instead of literal 63-table ownership-reassignment -- same non-owner+non-bypass property, avoids a dev-bootstrap-breaking owner migration. Dev password is a non-secret DML-only fixture; prod overrides via ALTER ROLE.')
ON CONFLICT (version) DO NOTHING;

COMMIT;
