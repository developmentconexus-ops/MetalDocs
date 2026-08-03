-- 0295_profile_governance_class.sql
-- ROADMAP unit 2.1 (G1 — per-profile signature policy). Ratified spec:
-- docs/superpowers/specs/2026-07-10-review-approval-workflow-model.md (R1 + G1);
-- design: docs/superpowers/specs/2026-07-10-g1-per-profile-signature-policy-design.md.
--
-- Adds the governance-class dimension to document profiles and its DB last-line
-- enforcement. A profile's governance_class decides its route-signature policy:
--   controlado -> the active approval route MUST contain >=1 approval-kind stage
--   simples    -> a review-only route is permitted (no approval stage required)
--   livre      -> no approval route is permitted at all
--
-- Expand-only, forward-only, idempotent (ADD COLUMN IF NOT EXISTS; DROP
-- CONSTRAINT/TRIGGER IF EXISTS before ADD — idiom from 0286/0287/0290).
-- Behavior-preserving backfill: every existing profile becomes 'controlado'
-- (the pre-G1 world where every profile implicitly required a signature), so no
-- live route silently becomes invalid.
--
-- Enforcement is a DEFERRABLE INITIALLY DEFERRED constraint trigger firing in
-- BOTH directions against one shared assert_route_shape() helper:
--   A. approval_route_stages writes  -> validate the affected active route's
--      shape against its profile's class. Deferred so a multi-statement route
--      write (INSERT route, INSERT stages) is validated once against the FULL
--      committed stage set.
--   A'. approval_routes writes (insert/activation) -> validate the route row
--      itself when it is (or becomes) active. Required because the stage trigger
--      only fires when a stage row is written: a zero-stage active route
--      (controlado with no approval stage, or ANY livre route) and an
--      insert-inactive-then-activate flip (which touches only approval_routes)
--      would otherwise never reach the last line. Same deferred helper, so a
--      route + its stages committed together are validated once against the
--      final committed shape.
--   B. document_profiles.governance_class change (reclassification guard) ->
--      re-validate every active route of the reclassified profile against the
--      NEW class.
-- Only ACTIVE routes are considered, so historical/superseded route versions
-- never trip the guard and in-flight instances (RouteVersionSnapshot-pinned)
-- are untouched. The app friendly-first check is domain Route.Validate(policy);
-- this trigger is the authoritative last line.
--
-- Error tokens (ERRCODE P0001) are prefix-stable so the app can map them:
--   'ErrRouteViolatesProfilePolicy:'  -> route-shape violation (direction A)
--   'ErrClassChangeRouteConflict:'    -> reclassify conflict     (direction B, -> app 409)

BEGIN;

-- 1. Column + CHECK -----------------------------------------------------------

ALTER TABLE metaldocs.document_profiles
    ADD COLUMN IF NOT EXISTS governance_class text NOT NULL DEFAULT 'controlado';

ALTER TABLE metaldocs.document_profiles
    DROP CONSTRAINT IF EXISTS document_profiles_governance_class_check;
ALTER TABLE metaldocs.document_profiles
    ADD CONSTRAINT document_profiles_governance_class_check
    CHECK (governance_class IN ('controlado', 'simples', 'livre'));

-- 2. Shared shape-assert helper ----------------------------------------------
-- Raises P0001 with the given prefix when route p_route_id violates p_class.
-- livre  -> the very existence of the (active) route is the violation.
-- controlado -> the route must have >=1 approval-kind stage.
-- simples -> always valid.

CREATE OR REPLACE FUNCTION public.assert_route_shape(
    p_route_id uuid,
    p_class    text,
    p_prefix   text
) RETURNS void
    LANGUAGE plpgsql
    SET search_path TO 'pg_catalog', 'pg_temp'
    AS $$
DECLARE
    v_has_approval boolean;
BEGIN
    IF p_class = 'livre' THEN
        RAISE EXCEPTION '%: profile is livre; no approval route is permitted (route %)',
            p_prefix, p_route_id USING ERRCODE = 'P0001';
    ELSIF p_class = 'controlado' THEN
        SELECT EXISTS (
            SELECT 1 FROM public.approval_route_stages s
             WHERE s.route_id = p_route_id AND s.stage_kind = 'approval'
        ) INTO v_has_approval;
        IF NOT v_has_approval THEN
            RAISE EXCEPTION '%: profile is controlado; route % must contain at least one approval-kind stage',
                p_prefix, p_route_id USING ERRCODE = 'P0001';
        END IF;
    END IF;
    -- simples (and any unexpected value): no route-shape constraint.
END;
$$;

-- 3. Bidirectional enforcement function --------------------------------------

CREATE OR REPLACE FUNCTION public.enforce_profile_route_policy() RETURNS trigger
    LANGUAGE plpgsql
    SET search_path TO 'pg_catalog', 'pg_temp'
    AS $$
DECLARE
    v_route_id  uuid;
    v_tenant_id uuid;
    v_profile   text;
    v_active    boolean;
    v_class     text;
BEGIN
    IF TG_TABLE_NAME = 'approval_route_stages' THEN
        -- Direction A: a stage write. Re-validate the affected route.
        IF TG_OP = 'DELETE' THEN
            v_route_id := OLD.route_id;
        ELSE
            v_route_id := NEW.route_id;
        END IF;

        SELECT r.tenant_id, r.profile_code, r.active
          INTO v_tenant_id, v_profile, v_active
          FROM public.approval_routes r
         WHERE r.id = v_route_id;
        IF NOT FOUND OR NOT v_active THEN
            RETURN NULL; -- route removed or inactive: nothing to enforce.
        END IF;

        SELECT p.governance_class INTO v_class
          FROM metaldocs.document_profiles p
         WHERE p.tenant_id = v_tenant_id AND p.code = v_profile;
        IF NOT FOUND THEN
            RETURN NULL; -- FK should prevent this; be lenient.
        END IF;

        PERFORM public.assert_route_shape(v_route_id, v_class, 'ErrRouteViolatesProfilePolicy');
        RETURN NULL;

    ELSIF TG_TABLE_NAME = 'approval_routes' THEN
        -- Direction A' (route row): an insert or activation. Validate the route
        -- shape whenever the committed row is active. Covers the zero-stage and
        -- activate-later paths the stage trigger cannot see.
        IF TG_OP = 'DELETE' THEN
            RETURN NULL; -- route removed: nothing to enforce.
        END IF;
        IF NOT NEW.active THEN
            RETURN NULL; -- inactive route: no shape constraint.
        END IF;

        SELECT p.governance_class INTO v_class
          FROM metaldocs.document_profiles p
         WHERE p.tenant_id = NEW.tenant_id AND p.code = NEW.profile_code;
        IF NOT FOUND THEN
            RETURN NULL; -- FK should prevent this; be lenient.
        END IF;

        PERFORM public.assert_route_shape(NEW.id, v_class, 'ErrRouteViolatesProfilePolicy');
        RETURN NULL;

    ELSIF TG_TABLE_NAME = 'document_profiles' THEN
        -- Direction B: reclassification guard. Only when the class changed.
        IF NEW.governance_class IS NOT DISTINCT FROM OLD.governance_class THEN
            RETURN NULL;
        END IF;
        FOR v_route_id IN
            SELECT r.id
              FROM public.approval_routes r
             WHERE r.tenant_id = NEW.tenant_id
               AND r.profile_code = NEW.code
               AND r.active
        LOOP
            PERFORM public.assert_route_shape(v_route_id, NEW.governance_class, 'ErrClassChangeRouteConflict');
        END LOOP;
        RETURN NULL;
    END IF;

    RETURN NULL;
END;
$$;

-- 4. Deferred constraint triggers (both directions) --------------------------

DROP TRIGGER IF EXISTS trg_route_stage_profile_policy ON public.approval_route_stages;
CREATE CONSTRAINT TRIGGER trg_route_stage_profile_policy
    AFTER INSERT OR UPDATE OR DELETE ON public.approval_route_stages
    DEFERRABLE INITIALLY DEFERRED
    FOR EACH ROW EXECUTE FUNCTION public.enforce_profile_route_policy();

DROP TRIGGER IF EXISTS trg_route_profile_policy ON public.approval_routes;
CREATE CONSTRAINT TRIGGER trg_route_profile_policy
    AFTER INSERT OR UPDATE OR DELETE ON public.approval_routes
    DEFERRABLE INITIALLY DEFERRED
    FOR EACH ROW EXECUTE FUNCTION public.enforce_profile_route_policy();

DROP TRIGGER IF EXISTS trg_profile_reclassify_route_policy ON metaldocs.document_profiles;
CREATE CONSTRAINT TRIGGER trg_profile_reclassify_route_policy
    AFTER UPDATE ON metaldocs.document_profiles
    DEFERRABLE INITIALLY DEFERRED
    FOR EACH ROW EXECUTE FUNCTION public.enforce_profile_route_policy();

-- ── schema_migrations ledger ─────────────────────────────────────────────────────────────

INSERT INTO public.schema_migrations (version, description)
VALUES ('0295', 'G1 per-profile signature policy: metaldocs.document_profiles.governance_class text NOT NULL DEFAULT ''controlado'' + CHECK (controlado/simples/livre); shared assert_route_shape() helper + enforce_profile_route_policy() DEFERRABLE INITIALLY DEFERRED constraint triggers on approval_route_stages (stage-shape), approval_routes (route-row insert/activation — closes zero-stage + activate-later paths), and document_profiles (reclassify guard). controlado routes need >=1 approval stage, livre routes forbidden, simples review-only allowed; only active routes enforced (superseded versions and in-flight snapshots untouched).')
ON CONFLICT (version) DO NOTHING;

COMMIT;
