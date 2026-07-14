-- 0307_drop_dead_grant_area_membership_fn.sql
-- ROADMAP unit 4.4 -- schema hygiene, hub-added item.
--
-- Drops the dead SQL function metaldocs.grant_area_membership(uuid, text,
-- text, text, text) (db/baseline/0001_current_schema.sql:188).
--
-- ============================================================================
-- PREMISE VERIFICATION (runtime truth, not assumed from docs)
-- ============================================================================
--   - DEAD: zero product-RUNTIME callers. The full repo reference set is
--     THREE occurrences (rg "grant_area_membership" over internal apps
--     frontend tests db, excluding baseline+this migration), NONE of them
--     product-runtime code:
--       1. internal/test/e2e_seed.go:579 -- a live e2e-seed HTTP endpoint
--          helper, build-tagged `//go:build integration && !production`,
--          reached only via the seed HTTP handler and NOT executed by any Go
--          test. (Non-`_test.go` filename, but test-only infrastructure, not
--          product runtime.) Track B (unit 4.3) removes this call.
--       2. tests/integration/scenarios/membership_fn_test.go -- an
--          integration-tagged Go test (TestGrantAreaMembershipFn +
--          TestGrantAreaMembershipIdempotent) that DOES call
--          `SELECT metaldocs.grant_area_membership(...)`. It is RED today
--          precisely because of the UNSATISFIABLE contradiction below (its
--          own header documents the escalated defect, "leaving RED and
--          reporting up") -- 2 of the 9-accepted integration baseline REDs.
--          It degrades gracefully once the function is dropped: the call
--          errors with "does not exist", which the test matches and
--          t.Skip()s -- so dropping the function turns these two baseline-RED
--          tests into SKIPs (a net baseline improvement), never new RED.
--          Track B removes these tests.
--       3. tests/integration/testdb/db.go:422 -- a `grant_area_membership`
--          key in the `metaldocsOwnedObjects` name->schema resolver map
--          (Qualified()/runtimeSchemaName() prefix objects in this set with
--          the `metaldocs` schema instead of `public`). It is a passive
--          lookup entry, consulted ONLY when test code qualifies that name
--          (i.e. the membership_fn_test.go path above); NO assertion iterates
--          the map expecting the object to exist. A stale entry after the
--          drop is therefore harmless. Track B (owner of tests/integration)
--          prunes it alongside the test removal. This migration does NOT edit
--          it (cross-track file ownership).
--     No product Go/TS runtime code calls this function.
--   - UNSATISFIABLE: the function's input validation
--     (db/baseline/0001_current_schema.sql:204) requires an UPPERCASE
--     area_code (`_area_code !~ '^[A-Z0-9_]+$'`, raises 22023 otherwise).
--     The row it must INSERT into public.user_process_areas carries FK
--     user_process_areas_tenant_id_area_code_fkey (tenant_id, area_code) ->
--     metaldocs.document_process_areas(tenant_id, code), and THAT table's
--     own CHECK area_code_format (db/baseline/0001_current_schema.sql:1060)
--     requires the OPPOSITE, a LOWERCASE-first code (`^[a-z][a-z0-9_-]{1,63}$`).
--     No string satisfies both regexes, so any area_code passing the
--     function's uppercase gate is rejected by the FK's referenced-table
--     CHECK -- the function can never succeed for ANY input. The lowercase
--     CHECK lives on document_process_areas.code (reached via the FK), NOT
--     directly on user_process_areas. Dead by construction, independent of
--     the zero-product-callers finding above.
--
-- ============================================================================
-- SAFETY
-- ============================================================================
-- Forward-only; safe to run exactly once and to no-op on re-run or in an
-- environment already lacking the function (existence-guarded assert + DROP
-- IF EXISTS + ledger ON CONFLICT DO NOTHING). The pre-drop assert is the
-- DB-level analog of the "0 product code depends on it" finding above: it
-- fails loud (P0001) if any normal/auto (deptype 'n'/'a') catalogued
-- dependent that pg_depend records against this function's OID still exists
-- (a view/default/constraint referencing the fn), rather
-- than silently cascade-dropping it. NOTE on coverage: pg_depend records
-- hard/auto/normal dependency EDGES between catalog objects; it does NOT
-- parse the *bodies* of other PL/pgSQL routines, so a plain
-- `SELECT metaldocs.grant_area_membership(...)` buried inside some other
-- function body is invisible to this assert (a known blind spot). That
-- residual risk is acceptable here: the grep sweep already established zero
-- such SQL callers, and the function is unsatisfiable regardless.
-- Deliberately NO CASCADE: a surprise dependent means the "dead" premise is
-- wrong for that environment and must be handled deliberately, never
-- destroyed along with it.

BEGIN;

-- ── Pre-drop dependency assert ──────────────────────────────────────────────
DO $$
DECLARE
  v_fn_oid oid;
  v_dependent text;
BEGIN
  v_fn_oid := to_regprocedure('metaldocs.grant_area_membership(uuid, text, text, text, text)');

  IF v_fn_oid IS NOT NULL THEN
    SELECT pg_describe_object(d.classid, d.objid, d.objsubid)
      INTO v_dependent
      FROM pg_catalog.pg_depend d
      WHERE d.refobjid = v_fn_oid
        AND d.refclassid = 'pg_catalog.pg_proc'::regclass
        AND d.deptype IN ('n', 'a')
      LIMIT 1;

    IF v_dependent IS NOT NULL THEN
      RAISE EXCEPTION '0307 pre-drop assert: metaldocs.grant_area_membership(uuid,text,text,text,text) still has a dependent (%) -- the dead-function premise does not hold in this environment; handle the dependent deliberately before dropping.', v_dependent;
    END IF;
  END IF;
END $$;

-- ── Drop the dead function (no CASCADE) ─────────────────────────────────────
DROP FUNCTION IF EXISTS metaldocs.grant_area_membership(uuid, text, text, text, text);

-- ── schema_migrations ledger ─────────────────────────────────────────────────
INSERT INTO public.schema_migrations (version, description)
VALUES ('0307', 'Unit 4.4 (schema hygiene, hub-added item): drop dead metaldocs.grant_area_membership(uuid,text,text,text,text). Zero product-runtime callers; the 3 repo references are all test-only infra: internal/test/e2e_seed.go:579 (integration seed helper, never run by Go tests), tests/integration/scenarios/membership_fn_test.go (integration test, RED today, skip-graceful on drop), tests/integration/testdb/db.go:422 (passive metaldocsOwnedObjects name->schema resolver entry, harmless when stale) -- Track B (unit 4.3) prunes all three. Also unsatisfiable by construction: the function requires an UPPERCASE area_code but the FK-referenced metaldocs.document_process_areas.code CHECK area_code_format requires lowercase, so no INSERT could ever succeed. Behind a pre-drop pg_depend assert (normal/auto catalogued-edge coverage only; PL/pgSQL bodies opaque), no CASCADE.')
ON CONFLICT (version) DO NOTHING;

COMMIT;
