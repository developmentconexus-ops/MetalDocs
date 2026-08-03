-- 0304_approval_stage_instance_selectors_snapshot.sql
-- ROADMAP unit 3.2 slice 6a (M4 ActorSelector, Option C' drift-source cutover
-- -- ADDITIVE step). Ratified design:
-- docs/superpowers/specs/2026-07-13-unit-3.2-actor-selector-design.md sec8b.
--
-- ADDITIVE ONLY, forward-only, idempotent (ADD COLUMN IF NOT EXISTS; backfill
-- guarded by WHERE selectors_snapshot = '[]'::jsonb so a re-run is a no-op).
-- The flat approval_route_stages.required_role/area_code columns and the
-- instance-side required_role_snapshot/area_code_snapshot columns are NOT
-- touched or dropped here -- they remain pure audit; the contract step
-- (dropping the route-side flat columns) is a later slice (6b).
--
-- New column public.approval_stage_instances.selectors_snapshot freezes the
-- stage's actor selectors onto the instance at submit (mirrors the existing
-- eligible_actor_ids jsonb precedent). Post-cutover this is the SOLE source
-- the drift path (decision_service.go / review_verdict_service.go)
-- re-resolves eligibility from: pool kinds (role_in_fixed_area,
-- role_in_document_area) re-resolve live membership; named_user selectors
-- (the drift-exempt class, incl. materialized submit_choice picks) are
-- unioned unconditionally and never re-filtered.
--
-- BACKFILL: existing in-flight rows are backfilled to a single synthesized
-- role_in_fixed_area(required_role_snapshot, area_code_snapshot) selector --
-- NOT an empty array -- because an empty snapshot would make the rewired
-- drift path resolve currentEligible = empty set for an in-flight stage
-- under a re-evaluate drift policy and wrongly collapse its quorum
-- denominator. The synthesized flat selector re-resolves to the identical
-- set the pre-cutover flat path produced (this IS the parity mapping). Rows
-- whose required_role_snapshot is empty backfill to '[]' (no such row exists
-- pre-unit -- the flat model always carried a role). Idempotent via the
-- selectors_snapshot = '[]'::jsonb guard.

BEGIN;

-- 1. New column --------------------------------------------------------------

ALTER TABLE public.approval_stage_instances
    ADD COLUMN IF NOT EXISTS selectors_snapshot jsonb NOT NULL DEFAULT '[]'::jsonb;

-- 2. Backfill ----------------------------------------------------------------
-- Synthesize one role_in_fixed_area selector per in-flight row still at the
-- default '[]', sourced from the row's own flat snapshot scalars. Keys MUST
-- match the domain ActorSelector json tags exactly: kind, role, area_code.

UPDATE public.approval_stage_instances
   SET selectors_snapshot = jsonb_build_array(
         jsonb_build_object(
           'kind', 'role_in_fixed_area',
           'role', required_role_snapshot,
           'area_code', area_code_snapshot))
 WHERE selectors_snapshot = '[]'::jsonb
   AND required_role_snapshot <> ''
   AND area_code_snapshot <> '';

-- ── schema_migrations ledger ─────────────────────────────────────────────────
INSERT INTO public.schema_migrations (version, description)
VALUES ('0304', 'Unit 3.2 slice 6a (M4 ActorSelector, Option C prime drift-source cutover, additive): new column public.approval_stage_instances.selectors_snapshot jsonb NOT NULL DEFAULT [] freezing the stage''s actor selectors onto the instance at submit, materializing submit_choice into named_user at write time; backfill synthesizes one role_in_fixed_area(required_role_snapshot, area_code_snapshot) selector per existing in-flight row (not an empty array, to preserve quorum denominators under re-evaluate drift policies) -- this is the parity mapping to the pre-cutover flat resolver. Flat route-side and instance-side snapshot columns retained -- contract step (drop) is a later slice.')
ON CONFLICT (version) DO NOTHING;

COMMIT;
