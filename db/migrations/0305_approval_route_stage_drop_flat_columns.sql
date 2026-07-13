-- 0305_approval_route_stage_drop_flat_columns.sql
-- ROADMAP unit 3.2 slice 6b (M4 ActorSelector, contract step). Ratified design:
-- docs/superpowers/specs/2026-07-13-unit-3.2-actor-selector-design.md sec8b.
--
-- CONTRACT step: drops the flat public.approval_route_stages.required_role
-- and area_code columns now that resolver/route-admin/submit code is fully
-- cut over to public.approval_route_stage_selectors (0303) as the sole
-- source of truth for a route stage's actor pool. The flat->selector
-- synthesis moves to the HTTP boundary (route_admin_handler.go); the
-- instance-side required_role_snapshot/area_code_snapshot audit columns and
-- ResolveEligibleActors are untouched -- they remain pure audit/instance-side.

BEGIN;

-- Pre-drop verification (fail-loud, mirrors 0303's end-of-backfill guard).
-- Dropping the flat columns is only safe once every stage has been migrated to
-- >=1 child selector; if any stage still lacks one, the selectors table is not
-- yet the complete source of truth and dropping the flat columns would silently
-- orphan that stage's actor pool. Fail the migration rather than proceed
-- (no-fallback principle). Guarded before the DROP so the flat columns survive
-- for diagnosis if the assertion trips.
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM approval_route_stages s
         WHERE NOT EXISTS (
            SELECT 1 FROM approval_route_stage_selectors x
             WHERE x.route_stage_id = s.id
         )
    ) THEN
        RAISE EXCEPTION 'cannot drop flat columns: stage without selector (0303 backfill incomplete or a stage was inserted without a selector)';
    END IF;
END $$;

ALTER TABLE public.approval_route_stages DROP COLUMN IF EXISTS required_role;
ALTER TABLE public.approval_route_stages DROP COLUMN IF EXISTS area_code;

-- ── schema_migrations ledger ─────────────────────────────────────────────────
INSERT INTO public.schema_migrations (version, description)
VALUES ('0305', 'Unit 3.2 slice 6b (M4 ActorSelector, contract step): drop flat public.approval_route_stages.required_role/area_code columns now that approval_route_stage_selectors (0303) is the sole source of truth for a route stage''s actor pool; flat<->selector synthesis relocates to the HTTP boundary (route_admin_handler.go). Instance-side required_role_snapshot/area_code_snapshot audit columns and ResolveEligibleActors unchanged.')
ON CONFLICT (version) DO NOTHING;

COMMIT;
