-- 0242: iam publishes metaldocs.v_active_user_areas — the purpose-built, versioned
-- active-membership READ CONTRACT (mission backend-module-boundary-hardening, M3/F3.1).
--
-- ADR-0039 D3(a)/D4: a non-owner module reads this PUBLISHED VIEW, never iam's
-- public.user_process_areas base table. The view encodes EXACTLY the Model-A
-- active-now predicate `effective_to IS NULL` (ADR 0037 D1) — no interval
-- reinterpretation (`effective_to > now()` stays refuted, ADR 0037 D2). It projects
-- the four columns the three M3 consumers consume — (tenant_id, user_id, area_code,
-- role); `role` is required by documents/approval ResolveEligibleActors (C3).
-- effective_from/effective_to are deliberately NOT exposed: encoding active-now IS the
-- view's contract, so consumers never re-derive the temporal predicate.
--
-- This is distinct from the existing 1:1 passthrough metaldocs.user_process_areas
-- (baseline 0001_current_schema.sql:1653), a schema-mirror exposing all columns with no
-- active predicate — NOT a published contract, and left untouched here. Security posture
-- matches that view (no security_invoker), so RLS over public.user_process_areas is
-- identical.

BEGIN;

CREATE VIEW metaldocs.v_active_user_areas AS
  SELECT tenant_id,
         user_id,
         area_code,
         role
    FROM public.user_process_areas
   WHERE effective_to IS NULL;

COMMENT ON VIEW metaldocs.v_active_user_areas IS
  'Published active-membership read contract (ADR-0039 D3a/D4; ADR 0037 D1): active-now user-area-role rows (effective_to IS NULL). Non-owner modules read THIS view, never public.user_process_areas. Columns (tenant_id, user_id, area_code, role). Mission backend-module-boundary-hardening M3/F3.1.';

INSERT INTO public.schema_migrations (version, description)
VALUES ('0242', 'iam publishes metaldocs.v_active_user_areas active-membership view (M3/F3.1, ADR-0039 D3a)')
ON CONFLICT (version) DO NOTHING;

COMMIT;
