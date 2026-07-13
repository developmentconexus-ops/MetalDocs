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

ALTER TABLE public.approval_route_stages DROP COLUMN IF EXISTS required_role;
ALTER TABLE public.approval_route_stages DROP COLUMN IF EXISTS area_code;

-- ── schema_migrations ledger ─────────────────────────────────────────────────
INSERT INTO public.schema_migrations (version, description)
VALUES ('0305', 'Unit 3.2 slice 6b (M4 ActorSelector, contract step): drop flat public.approval_route_stages.required_role/area_code columns now that approval_route_stage_selectors (0303) is the sole source of truth for a route stage''s actor pool; flat<->selector synthesis relocates to the HTTP boundary (route_admin_handler.go). Instance-side required_role_snapshot/area_code_snapshot audit columns and ResolveEligibleActors unchanged.')
ON CONFLICT (version) DO NOTHING;

COMMIT;
