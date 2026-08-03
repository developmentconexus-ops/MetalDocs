-- 0287_approval_route_versioning.sql
-- M2b F2 (approval-kernel-backend, approval remediation design spec.md/plan.md W1/W6):
-- closes the real route-versioning gap. Runtime-truth verification this session found
-- `approval_routes.version` and `.active` already existed (migrations 0134/0146) and
-- `route_admin_service.go` Update() already mutates the SAME row in place. The actual
-- defect: `enforce_route_immutable` (added by an earlier migration, function body
-- reproduced below) blocks ANY UPDATE on a route once referenced by an
-- approval_instances row, with no column-level exception — so a route becomes
-- PERMANENTLY frozen after first use, with no path to edit it again. This migration:
--   1. Adds `superseded_at timestamptz` (the only genuinely new column; `version`/
--      `active` already exist).
--   2. Replaces the single UNIQUE(tenant_id, profile_code) constraint (which allowed
--      only ever one row per profile) with a partial UNIQUE(tenant_id, profile_code)
--      WHERE active (exactly one active row per profile) plus a UNIQUE(tenant_id,
--      profile_code, version) covering the full version history.
--   3. Replaces enforce_route_immutable()'s body with a column-scoped exception: an
--      UPDATE that touches ONLY `active`/`superseded_at` is allowed even on an in-use
--      row (needed so the application-layer supersede path can retire the old row in
--      the same statement shape); an UPDATE touching any other column (name,
--      profile_code, version, or anything else) remains blocked exactly as before
--      whenever the row is in use OR already inactive (a superseded/historical row
--      must never be edited at all, even before this migration existing tests
--      exercised only the in-use case).
--
-- No behavior change for any route that has never been superseded and stays active —
-- in-place mutation continues to work exactly as before. Service-layer wiring for the
-- new supersede-on-in-use branch is F2's Go changes (route_admin_service.go), not this
-- migration.

BEGIN;

-- approval_routes -------------------------------------------------------------

ALTER TABLE public.approval_routes
    ADD COLUMN IF NOT EXISTS superseded_at timestamp with time zone;

-- Replace the single-row-per-profile unique constraint with two: partial-active
-- (exactly one live version per profile) and full version history (every version
-- of a profile keeps a distinct, permanently addressable row).
ALTER TABLE public.approval_routes
    DROP CONSTRAINT IF EXISTS approval_routes_tenant_profile_key;

DROP INDEX IF EXISTS public.approval_routes_active_profile_uq;
CREATE UNIQUE INDEX approval_routes_active_profile_uq
    ON public.approval_routes (tenant_id, profile_code)
    WHERE active;

DROP INDEX IF EXISTS public.approval_routes_profile_version_uq;
CREATE UNIQUE INDEX approval_routes_profile_version_uq
    ON public.approval_routes (tenant_id, profile_code, version);

-- enforce_route_immutable(): column-scoped exception. Mirrors the pre-existing
-- function's style (SET search_path, ERRCODE P0001, same message prefix
-- "ErrRouteInUse:" that internal/modules/documents/approval/infrastructure/errors.go
-- string-matches on) — only the blocking condition changes.
CREATE OR REPLACE FUNCTION public.enforce_route_immutable()
    RETURNS trigger
    LANGUAGE plpgsql
    SET search_path = pg_catalog, pg_temp
AS $$
BEGIN
    -- Always allow an UPDATE that touches only active/superseded_at (or neither,
    -- e.g. a no-op UPDATE) — this is the supersede step that retires a row.
    IF NEW.name IS NOT DISTINCT FROM OLD.name
       AND NEW.profile_code IS NOT DISTINCT FROM OLD.profile_code
       AND NEW.version IS NOT DISTINCT FROM OLD.version
       AND NEW.tenant_id IS NOT DISTINCT FROM OLD.tenant_id
       AND NEW.created_at IS NOT DISTINCT FROM OLD.created_at
       AND NEW.created_by IS NOT DISTINCT FROM OLD.created_by
    THEN
        RETURN NEW;
    END IF;

    -- Any other column change (definition columns) is blocked once the row is
    -- in use by an instance OR already inactive (a historical/superseded row
    -- must never be edited).
    IF NOT OLD.active OR EXISTS (
        SELECT 1
          FROM public.approval_instances
         WHERE route_id = OLD.id
    ) THEN
        RAISE EXCEPTION 'ErrRouteInUse: route % is referenced by one or more approval instances (or is already inactive) and cannot be modified', OLD.id
            USING ERRCODE = 'P0001';
    END IF;

    RETURN NEW;
END;
$$;

-- ── schema_migrations ledger ─────────────────────────────────────────────────────────────

INSERT INTO public.schema_migrations (version, description)
VALUES ('0287', 'M2b F2: route versioning (W1) + empty-pool 422 substrate (W6). Adds approval_routes.superseded_at; replaces single tenant_id+profile_code unique with partial-active-unique + full version-history-unique; enforce_route_immutable() gains a column-scoped exception permitting active/superseded_at-only updates on in-use rows while continuing to block definition-column edits on in-use or inactive rows. No behavior change for never-superseded active routes.')
ON CONFLICT (version) DO NOTHING;

COMMIT;
