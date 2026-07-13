-- 0303_approval_route_stage_selectors.sql
-- ROADMAP unit 3.2 slice 2 (M4 ActorSelector, B2 — DB expand + backfill).
-- Ratified design: docs/superpowers/specs/2026-07-13-unit-3.2-actor-selector-design.md
-- §3. ADR 0084 (pending, tracked slice 9) records the ActorSelector model this
-- migration lands the storage for.
--
-- EXPAND phase only, forward-only, idempotent (CREATE TABLE IF NOT EXISTS;
-- backfill uses WHERE NOT EXISTS so a re-run is a no-op). Flat
-- approval_route_stages.required_role/area_code columns are RETAINED this
-- slice — the contract (column-drop) step is migration 0304, landed only
-- after the resolver/route-admin/submit code is fully cut over to selectors
-- (spec §3, slice 6).
--
-- New child table public.approval_route_stage_selectors is the discriminated
-- union of "who may act on this stage" (BPMN candidateUsers/candidateGroups/
-- assignee alignment, spec §2): one stage may carry several selectors, the
-- resolver unions their pools (B3, a later slice). kind discriminates which
-- of user_id/role/area_code are meaningful; the per-kind CHECK enforces exact
-- field presence/absence, mirroring domain ActorSelector.Validate()
-- (internal/modules/approval/domain/selector.go) as the DB-side last line.
--
-- Multi-tenant: tenant_id is NOT NULL and every new index is tenant-scoped
-- (tenant_id leads both new indexes) per the standing invariant. No RLS is
-- added here: approval_route_stages itself (the FK parent) carries no RLS
-- either (route *config* written on the route.admin-gated route-admin path,
-- not the regulated approval_instances/approval_signoffs write path whose
-- tripwires are unchanged — spec §3).
--
-- BACKFILL: every existing approval_route_stages row gets exactly one
-- role_in_fixed_area(required_role, area_code) selector at selector_order=1,
-- carrying the parent route's tenant_id. Idempotent via WHERE NOT EXISTS. An
-- end-of-migration verification DO block fails the migration loudly if any
-- stage is left without a selector (fail-closed, no-fallback principle).

BEGIN;

-- 1. New table -------------------------------------------------------------

CREATE TABLE IF NOT EXISTS public.approval_route_stage_selectors (
    id uuid DEFAULT gen_random_uuid() NOT NULL PRIMARY KEY,
    tenant_id uuid NOT NULL,
    route_stage_id uuid NOT NULL REFERENCES public.approval_route_stages(id) ON DELETE CASCADE,
    selector_order integer NOT NULL,
    kind text NOT NULL,
    user_id text,
    role text,
    area_code text,
    CONSTRAINT approval_route_stage_selectors_kind_check
        CHECK (kind IN ('named_user', 'role_in_fixed_area', 'role_in_document_area', 'submit_choice')),
    CONSTRAINT approval_route_stage_selectors_fields_consistent
        CHECK (
            (kind = 'named_user' AND user_id IS NOT NULL AND role IS NULL AND area_code IS NULL)
            OR (kind = 'role_in_fixed_area' AND role IS NOT NULL AND area_code IS NOT NULL AND user_id IS NULL)
            OR (kind = 'role_in_document_area' AND role IS NOT NULL AND area_code IS NULL AND user_id IS NULL)
            OR (kind = 'submit_choice' AND role IS NOT NULL AND area_code IS NOT NULL AND user_id IS NULL)
        ),
    CONSTRAINT approval_route_stage_selectors_order_uq UNIQUE (route_stage_id, selector_order)
);

CREATE INDEX IF NOT EXISTS ix_approval_route_stage_selectors_tenant_id
    ON public.approval_route_stage_selectors (tenant_id);

CREATE INDEX IF NOT EXISTS ix_approval_route_stage_selectors_route_stage_id
    ON public.approval_route_stage_selectors (route_stage_id);

-- 2. Backfill ----------------------------------------------------------------
-- One role_in_fixed_area selector per existing stage row, sourced from the
-- stage's own required_role/area_code and the owning route's tenant_id.
-- WHERE NOT EXISTS makes this idempotent across re-runs.

INSERT INTO public.approval_route_stage_selectors
    (tenant_id, route_stage_id, selector_order, kind, role, area_code)
SELECT r.tenant_id, s.id, 1, 'role_in_fixed_area', s.required_role, s.area_code
  FROM public.approval_route_stages s
  JOIN public.approval_routes r
    ON r.id = s.route_id
 WHERE NOT EXISTS (
    SELECT 1 FROM public.approval_route_stage_selectors x
     WHERE x.route_stage_id = s.id
 );

-- 3. Verification --------------------------------------------------------
-- Fail the migration loudly rather than silently leaving a stage with no
-- resolvable actor pool (no-fallback principle).

DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM approval_route_stages s
         WHERE NOT EXISTS (
            SELECT 1 FROM approval_route_stage_selectors x
             WHERE x.route_stage_id = s.id
         )
    ) THEN
        RAISE EXCEPTION 'backfill incomplete: stage without selector';
    END IF;
END $$;

-- ── schema_migrations ledger ─────────────────────────────────────────────────
INSERT INTO public.schema_migrations (version, description)
VALUES ('0303', 'Unit 3.2 slice 2 (M4 ActorSelector, B2 expand): new child table public.approval_route_stage_selectors (tenant_id, route_stage_id FK CASCADE, selector_order, kind, user_id/role/area_code with per-kind CHECK mirroring domain ActorSelector.Validate) + tenant-scoped indexes; backfill one role_in_fixed_area(required_role, area_code) selector per existing approval_route_stages row; end-of-migration verification asserts every stage has >=1 selector. Flat required_role/area_code columns retained this slice -- dropped in migration 0304 (contract step, slice 6) only after resolver/route-admin/submit cut over.')
ON CONFLICT (version) DO NOTHING;

COMMIT;
