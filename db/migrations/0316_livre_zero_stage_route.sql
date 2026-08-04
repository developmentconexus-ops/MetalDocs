-- 0316_livre_zero_stage_route.sql
-- ADR 0087 — livre governance class: configured no-approval route.
--
-- ============================================================================
-- WHY
-- ============================================================================
--   Until now `governance_class='livre'` meant "no approval route is permitted"
--   (ex-migration 0295's assert_route_shape livre arm, folded into the
--   2026-07-29 baseline). Since config-first became universal that stance was
--   self-contradictory: documents hard-block creation without an active route
--   (D2 gate) and templates hard-block on has_active_template_route (ADR 0086),
--   so a livre profile could own nothing at all.
--
--   Operator ruling (2026-07-29, ratified as ADR 0087): livre is a profile
--   whose route is EXPLICITLY CONFIGURED to require no approval. Absence of a
--   route is misconfiguration for every governance class, without exception.
--
--   The livre arm of assert_route_shape therefore flips from "RAISE always" to
--   "the active route MUST carry ZERO stages". The route SHAPE is the whole
--   truth — there is no auto_approve flag anywhere, by design (two sources of
--   truth that can disagree; no-fallback principle).
--
-- ============================================================================
-- WHAT CHANGES
-- ============================================================================
--   CREATE OR REPLACE public.assert_route_shape(uuid, text, text):
--     livre       -> the route must have zero rows in approval_route_stages;
--                    any stage raises P0001 with the caller's prefix.
--     controlado  -> >=1 stage AND >=1 approval-kind stage.
--     simples     -> >=1 stage (NEW at the DB line: previously an app-only
--                    structural floor).
--     other/unknown -> fails closed onto the controlado rule.
--
--   Zero-stage is now DB-EXCLUSIVE to livre (ADR 0087 §4). That is the point of
--   this arm split: a stageless route AUTO-APPROVES on submit, so the shape may
--   never exist on a governed class. The app-side floor cannot be the last line
--   — RouteAdminService.resolvePolicy yields "" whenever the policy reader is
--   unwired, and Validate("") is structural-only by design.
--
--   enforce_profile_route_policy() is NOT touched: it already calls
--   assert_route_shape in all three directions (stage write, route row
--   write/activation, profile reclassification) and passes the class through,
--   so replacing the helper changes every direction at once. Its two constraint
--   triggers (trg_route_profile_policy, trg_route_stage_profile_policy) and the
--   reclassification trigger (trg_profile_reclassify_route_policy) stay
--   DEFERRABLE INITIALLY DEFERRED and are deliberately NOT recreated — that
--   preserves the intra-tx-downgrade guarantee (the final committed stage set
--   is what is validated, never an intermediate).
--
--   Consequences that fall out of the replacement, all intended:
--     * livre profile + zero-stage ACTIVE route  -> accepted (was: rejected);
--     * livre profile + any staged ACTIVE route  -> rejected (unchanged);
--     * reclassify controlado->livre while a staged active route exists
--       -> rejected at COMMIT with ErrClassChangeRouteConflict;
--     * reclassify livre->controlado while a zero-stage active route exists
--       -> rejected at COMMIT (controlado arm finds no approval-kind stage).
--
-- ============================================================================
-- DATA
-- ============================================================================
--   No purge, no backfill. Under the OLD rule a livre profile could not own an
--   active route at all, so by construction there is nothing to migrate: every
--   livre profile has zero routes today, which is exactly the state the NEW
--   rule calls "unconfigured" (the app's creation gates already surface that as
--   a missing-route error). Because nothing is deleted, the SQLSTATE 55006
--   lesson from ADR 0086 (deferred policy triggers must be disabled around a
--   route purge) does not apply here — no DELETE touches approval_routes.
--
-- Forward-only, idempotent (CREATE OR REPLACE), self-registering.

BEGIN;

CREATE OR REPLACE FUNCTION public.assert_route_shape(p_route_id uuid, p_class text, p_prefix text)
RETURNS void
LANGUAGE plpgsql
SET search_path TO 'pg_catalog', 'pg_temp'
AS $$
DECLARE
    v_has_approval boolean;
    v_has_stage    boolean;
BEGIN
    IF p_class = 'livre' THEN
        -- ADR 0087: the route is REQUIRED and must carry no stages at all.
        -- Any stage kind counts: a livre route that reviews is not livre.
        SELECT EXISTS (
            SELECT 1 FROM public.approval_route_stages s
             WHERE s.route_id = p_route_id
        ) INTO v_has_stage;
        IF v_has_stage THEN
            RAISE EXCEPTION '%: profile is livre; route % must carry zero stages (a livre route is configured to require no approval)',
                p_prefix, p_route_id USING ERRCODE = 'P0001';
        END IF;
    ELSE
        -- ADR 0087 §4: zero-stage is DB-EXCLUSIVE to livre. A stageless route
        -- auto-approves on submit, so that shape may never exist on a governed
        -- class — not even transiently, and not even when the app-side policy
        -- reader is unwired (resolvePolicy yields "" ⇒ structural-only
        -- validation; this line is the authoritative one).
        SELECT EXISTS (
            SELECT 1 FROM public.approval_route_stages s
             WHERE s.route_id = p_route_id
        ) INTO v_has_stage;
        IF NOT v_has_stage THEN
            RAISE EXCEPTION '%: profile is %; route % must contain at least one stage (only a livre route may be stageless)',
                p_prefix, p_class, p_route_id USING ERRCODE = 'P0001';
        END IF;

        -- controlado additionally requires a SIGNATURE stage. Unknown/unexpected
        -- classes fail closed onto the same rule (never silently unconstrained).
        IF p_class <> 'simples' THEN
            SELECT EXISTS (
                SELECT 1 FROM public.approval_route_stages s
                 WHERE s.route_id = p_route_id AND s.stage_kind = 'approval'
            ) INTO v_has_approval;
            IF NOT v_has_approval THEN
                RAISE EXCEPTION '%: profile is %; route % must contain at least one approval-kind stage',
                    p_prefix, p_class, p_route_id USING ERRCODE = 'P0001';
            END IF;
        END IF;
    END IF;
END;
$$;

-- ── schema_migrations ledger ────────────────────────────────────────────────

INSERT INTO public.schema_migrations (version, description)
VALUES ('0316', 'ADR 0087: governance_class=''livre'' flips from ''no active route permitted'' to ''active route REQUIRED and it must carry ZERO stages''. CREATE OR REPLACE assert_route_shape(uuid,text,text) only: the livre arm stops raising unconditionally and instead raises P0001 when any approval_route_stages row exists for the route; every NON-livre class additionally gains a >=1-stage floor at the DB line, making zero-stage DB-exclusive to livre (a stageless route auto-approves on submit, so that shape may never exist on a governed class, and the app-side floor is not authoritative because the policy reader can be unwired). controlado keeps its >=1-approval-kind-stage rule; simples gains the >=1-any-stage rule; unknown classes fail closed onto the controlado rule. enforce_profile_route_policy() and its three constraint triggers (trg_route_profile_policy, trg_route_stage_profile_policy, trg_profile_reclassify_route_policy) are intentionally NOT recreated, so both enforcement directions and DEFERRABLE INITIALLY DEFERRED semantics are preserved and the replacement takes effect in all three call sites at once. No data purge and no backfill: under the superseded rule a livre profile could own no active route, so zero rows exist to migrate, and no DELETE touches approval_routes -- the ADR 0086 SQLSTATE 55006 deferred-trigger lesson therefore does not apply. First forward migration after the 2026-07-29 baseline fold; also exercises the empty-tail migration pipeline end-to-end.')
ON CONFLICT (version) DO NOTHING;

COMMIT;
