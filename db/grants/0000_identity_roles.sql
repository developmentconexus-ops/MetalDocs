-- MetalDocs identity-split bootstrap (issue #88 / A6.1 re-cut).
--
-- Runs FIRST in db/grants (lexical order), and — unlike 0001_role_grants.sql
-- — is intended to run ONLY under the bootstrap superuser identity: at fresh
-- compose bootstrap (docker-entrypoint-initdb.d, connected as POSTGRES_USER)
-- and at every invocation of the metaldocs-dbprovision one-shot binary
-- (apps/dbprovision/cmd/metaldocs-dbprovision), which is the ONLY thing that
-- opens a connection as the bootstrap superuser going forward. Application
-- processes (metaldocs-api/worker/jobs) never see this identity and never
-- run this file.
--
-- THE THREE IDENTITIES (A6.1 re-cut ruling):
--   1. bootstrap superuser (today's metaldocs_app / ${POSTGRES_USER}) --
--      creates roles and extensions. Used once, by provisioning only.
--   2. metaldocs_owner -- owns the schema, runs all DDL migrations
--      (db/migrations via internal/platform/migrate.Apply, plus the River
--      job-queue schema). NOSUPERUSER, NOBYPASSRLS, and deliberately NOLOGIN:
--      it is never a login identity in its own right, only ever reached via
--      `SET ROLE metaldocs_owner` from the bootstrap superuser session inside
--      metaldocs-dbprovision (see apps/dbprovision). NOLOGIN closes off the
--      one attack surface a LOGIN owner role would add (direct authentication
--      as the object owner) without costing anything -- provisioning never
--      needs to authenticate as it directly.
--   3. metaldocs_runtime -- DML only (see 0001_role_grants.sql for the grant
--      surface). The ONLY identity internal/platform/db/postgres.
--      AssertSafeIdentity is meant to accept in a correctly provisioned
--      environment; metaldocs-api/worker/jobs connect as this role exclusively.
--
-- HARD CONSTRAINT: the serving identity (metaldocs_runtime) must never own a
-- table. Postgres RLS does not apply to the table owner unless FORCE ROW
-- LEVEL SECURITY is set, so ownership on any non-FORCE table would be a
-- silent full RLS bypass -- precisely the defect A6.1 exists to close. This
-- file only ever grants metaldocs_runtime DML privileges (via
-- 0001_role_grants.sql); it never makes it an owner of anything.
--
-- OWNERSHIP TRANSFER: REASSIGN OWNED BY CURRENT_USER TO metaldocs_owner below
-- moves every object the bootstrap superuser currently owns (the entire
-- schema, on every pre-A6.1 volume) to metaldocs_owner in one statement. It
-- is unconditionally idempotent: once ownership has moved, CURRENT_USER (the
-- bootstrap superuser) owns nothing in this database, so a repeat run
-- reassigns an empty set and is a no-op. This does not depend on knowing the
-- bootstrap role's name (metaldocs_app in this repo's compose, potentially
-- different in another deployment) -- CURRENT_USER is whichever identity the
-- provisioning connection authenticated as, which by construction is always
-- the bootstrap superuser.
--
-- Idempotent and safe to re-run, exactly like 0001_role_grants.sql.

BEGIN;

-- ── metaldocs_owner: schema-owning, DDL-capable, NOLOGIN role ──────────────
DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'metaldocs_owner') THEN
    IF NOT EXISTS (
      SELECT 1 FROM pg_roles
       WHERE rolname = current_user AND (rolcreaterole OR rolsuper)
    ) THEN
      RAISE NOTICE 'metaldocs_owner is absent and % lacks CREATEROLE -- skipping owner role provisioning (expected outside the bootstrap-superuser provisioning path)', current_user;
      RETURN;
    END IF;
    EXECUTE 'CREATE ROLE metaldocs_owner NOSUPERUSER NOBYPASSRLS NOCREATEDB NOCREATEROLE NOINHERIT NOLOGIN';
  END IF;
END
$$;

-- ── metaldocs_runtime: non-owner, non-bypass application role (A6.1) ───────
-- Role creation only (moved here from 0001_role_grants.sql, which now runs
-- with metaldocs_owner's privileges and therefore cannot itself CREATE ROLE
-- -- see that file's header). Grant surface stays in 0001.
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
END
$$;

-- ── metaldocs_ci: non-owner, non-bypass DML test role (from 0284) ──────────
-- Role creation only, moved here for the same reason as metaldocs_runtime
-- above. Grant surface stays in 0001.
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
END
$$;

-- ── ownership transfer: bootstrap superuser -> metaldocs_owner ─────────────
-- Schema ownership first (grants metaldocs_owner CREATE on both schemas, so
-- migrate.Apply's forward migrations can create new objects that are owned
-- by metaldocs_owner from birth, no per-object reassignment ever needed
-- again), then every object the bootstrap superuser currently owns. Both
-- statements require superuser (or matching role membership); this file must
-- only ever run under the bootstrap superuser connection -- see header.
DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'metaldocs_owner')
     AND (SELECT rolsuper FROM pg_roles WHERE rolname = current_user) THEN
    EXECUTE 'ALTER SCHEMA metaldocs OWNER TO metaldocs_owner';
    EXECUTE 'ALTER SCHEMA public OWNER TO metaldocs_owner';
    EXECUTE 'REASSIGN OWNED BY CURRENT_USER TO metaldocs_owner';
  ELSE
    RAISE NOTICE 'skipping ownership transfer to metaldocs_owner: role absent or % is not superuser', current_user;
  END IF;
END
$$;

COMMIT;
