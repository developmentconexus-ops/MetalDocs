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
-- OWNERSHIP TRANSFER: moves every object the bootstrap superuser currently
-- owns in the public/metaldocs schemas (the entire application schema, on
-- every pre-A6.1 volume) to metaldocs_owner. This is a scoped, per-object
-- ALTER ... OWNER TO loop below -- NOT the blanket `REASSIGN OWNED BY
-- CURRENT_USER` statement an earlier revision of this file used. That
-- blanket form is unusable here: CURRENT_USER, when this file runs under
-- the bootstrap superuser as its header requires, is the actual cluster
-- initdb role, which (structurally, in every Postgres database) also "owns"
-- pg_catalog and information_schema. REASSIGN OWNED BY has no per-schema
-- filter -- it tries to move EVERYTHING CURRENT_USER owns, so it always
-- hits those two system schemas and always fails with
-- "cannot reassign ownership of objects owned by role ... because they are
-- required by the database system" (SQLSTATE 2BP01), rolling back this
-- entire transaction -- including the CREATE ROLE statements above -- on
-- every single run. Verified empirically against a fresh scratch database:
-- REASSIGN OWNED BY CURRENT_USER fails for the literal bootstrap role even
-- with zero other objects in play, and succeeds for any ordinary superuser
-- role that isn't the bootstrap identity. The scoped loop below reassigns
-- only what public/metaldocs actually contain (relations, sequences, views,
-- materialized views, functions/procedures, enum/domain types), so it never
-- touches pg_catalog/information_schema and cannot hit 2BP01.
--
-- Idempotent and safe to re-run, exactly like 0001_role_grants.sql: each
-- loop iteration only matches objects still owned by CURRENT_USER, so a
-- repeat run (once ownership has already moved) matches nothing and is a
-- no-op.

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
-- Role creation only (moved here from 0001_role_grants.sql). Both this file
-- and 0001 run under the bootstrap superuser -- apps/dbprovision does not
-- SET ROLE metaldocs_owner until its Stage 3, after the whole grants
-- directory (0000 + 0001) has applied -- so the split is not about which
-- role has CREATEROLE; it is about sequencing: 0001's grants name
-- metaldocs_owner/metaldocs_runtime/metaldocs_ci as grantees, so those roles
-- must already exist when 0001 runs. Grant surface stays in 0001.
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
-- Role creation only, moved here for the same sequencing reason as
-- metaldocs_runtime above. Grant surface stays in 0001.
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
-- again), then every object in those schemas the bootstrap superuser
-- currently owns, one object at a time (see the comment above for why this
-- cannot be a single blanket REASSIGN OWNED BY). All of it requires
-- superuser (or matching role membership); this file must only ever run
-- under the bootstrap superuser connection -- see header.
DO $$
DECLARE
  r RECORD;
  bootstrap_role oid;
BEGIN
  IF NOT (
    EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'metaldocs_owner')
    AND (SELECT rolsuper FROM pg_roles WHERE rolname = current_user)
  ) THEN
    RAISE NOTICE 'skipping ownership transfer to metaldocs_owner: role absent or % is not superuser', current_user;
    RETURN;
  END IF;

  EXECUTE 'ALTER SCHEMA metaldocs OWNER TO metaldocs_owner';
  EXECUTE 'ALTER SCHEMA public OWNER TO metaldocs_owner';

  bootstrap_role := (SELECT oid FROM pg_roles WHERE rolname = current_user);

  -- Relations: ordinary/partitioned tables, sequences, views, materialized
  -- views. Sequences take ALTER SEQUENCE, everything else ALTER TABLE
  -- (Postgres's generic spelling for OWNER TO on tables/views/matviews).
  --
  -- Sequences OWNED BY a table column (every SERIAL/IDENTITY/nextval-default
  -- sequence this baseline schema defines, e.g.
  -- metaldocs.audit_events_audit_sequence_seq) are excluded here: Postgres
  -- refuses ALTER SEQUENCE ... OWNER TO on them directly ("cannot change
  -- owner of sequence", SQLSTATE 0A000) because their ownership is derived
  -- from, and auto-follows, their owning table's ownership -- verified
  -- empirically that ALTER TABLE ... OWNER TO on the parent table alone
  -- already flips the linked sequence (and any index) to the new owner, no
  -- separate statement needed or permitted. The pg_depend filter below
  -- detects that link (deptype 'a'/'i' = auto/internal dependency) and skips
  -- exactly those sequences; a sequence with no such dependency (declared
  -- via bare CREATE SEQUENCE, not tied to a column) still needs the manual
  -- ALTER SEQUENCE below.
  FOR r IN
    SELECT c.relname, n.nspname, c.relkind
    FROM pg_class c
    JOIN pg_namespace n ON n.oid = c.relnamespace
    WHERE n.nspname IN ('public', 'metaldocs')
      AND c.relowner = bootstrap_role
      AND c.relkind IN ('r', 'p', 'S', 'v', 'm')
      AND NOT (
        c.relkind = 'S'
        AND EXISTS (
          SELECT 1 FROM pg_depend d
          WHERE d.objid = c.oid AND d.deptype IN ('a', 'i')
        )
      )
  LOOP
    IF r.relkind = 'S' THEN
      EXECUTE format('ALTER SEQUENCE %I.%I OWNER TO metaldocs_owner', r.nspname, r.relname);
    ELSE
      EXECUTE format('ALTER TABLE %I.%I OWNER TO metaldocs_owner', r.nspname, r.relname);
    END IF;
  END LOOP;

  -- Functions and procedures (e.g. the trigger functions this baseline
  -- schema defines). Identity arguments disambiguate overloads.
  FOR r IN
    SELECT p.oid, p.proname, p.prokind, n.nspname,
           pg_get_function_identity_arguments(p.oid) AS args
    FROM pg_proc p
    JOIN pg_namespace n ON n.oid = p.pronamespace
    WHERE n.nspname IN ('public', 'metaldocs')
      AND p.proowner = bootstrap_role
  LOOP
    IF r.prokind = 'p' THEN
      EXECUTE format('ALTER PROCEDURE %I.%I(%s) OWNER TO metaldocs_owner', r.nspname, r.proname, r.args);
    ELSE
      EXECUTE format('ALTER FUNCTION %I.%I(%s) OWNER TO metaldocs_owner', r.nspname, r.proname, r.args);
    END IF;
  END LOOP;

  -- Types: enums and domains (e.g. metaldocs.mddm_version_status). Composite
  -- row types ride along with the table that owns them, already handled
  -- above; base scalar types are never user-created in this schema.
  FOR r IN
    SELECT t.typname, n.nspname
    FROM pg_type t
    JOIN pg_namespace n ON n.oid = t.typnamespace
    WHERE n.nspname IN ('public', 'metaldocs')
      AND t.typowner = bootstrap_role
      AND t.typtype IN ('e', 'd')
  LOOP
    EXECUTE format('ALTER TYPE %I.%I OWNER TO metaldocs_owner', r.nspname, r.typname);
  END LOOP;
END
$$;

COMMIT;
