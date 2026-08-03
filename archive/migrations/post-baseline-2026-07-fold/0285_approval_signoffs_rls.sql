-- 0285_approval_signoffs_rls.sql
-- M7 F7.4 (global-maximum-remediation, milestone-7-tenant-lifecycle,
-- validation-contract.md §4.4): close the one tenant-scoped table that the
-- FORCE-RLS census missed. public.approval_signoffs is keyed by
-- actor_tenant_id (not tenant_id) — which is exactly why the earlier sweep,
-- keyed on the tenant_id column name, skipped it. It carries e-signature PII
-- (actor identity, display-name snapshot, signature payload) and MUST be
-- tenant-isolated at the DB tier like the other 33 tenant tables.
--
-- Idiom copied VERBATIM from 0281_tenant_keys.sql / 0278_tenant_lifecycle_jobs
-- (GUC metaldocs.tenant_id, policy name tenant_isolation, NULLIF(...) IS NULL
-- null-GUC escape-hatch branch preserved for janitor/bootstrap parity — see
-- ADR 0027 amendment). Only the schema (public) and the tenant column
-- (actor_tenant_id) differ.
--
-- No cross-tenant co-sign semantics exist (a signoff's actor_tenant_id is the
-- signer's own tenant, FK-anchored to iam_users(tenant_id,user_id)), so the
-- strict tenant_isolation policy is correct — no ADR 0070 amendment / lint
-- carve-out needed (validation-contract §4.4 default expectation).
--
-- Idempotent (ENABLE/FORCE are no-ops if already set; DROP POLICY IF EXISTS
-- before CREATE). Down (manual, if ever needed):
--   DROP POLICY IF EXISTS tenant_isolation ON public.approval_signoffs;
--   ALTER TABLE ONLY public.approval_signoffs NO FORCE ROW LEVEL SECURITY;
--   ALTER TABLE public.approval_signoffs DISABLE ROW LEVEL SECURITY;

BEGIN;

ALTER TABLE public.approval_signoffs ENABLE ROW LEVEL SECURITY;
ALTER TABLE ONLY public.approval_signoffs FORCE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS tenant_isolation ON public.approval_signoffs;
CREATE POLICY tenant_isolation ON public.approval_signoffs
  USING (
    (NULLIF(current_setting('metaldocs.tenant_id'::text, true), ''::text) IS NULL)
    OR (actor_tenant_id = (NULLIF(current_setting('metaldocs.tenant_id'::text, true), ''::text))::uuid)
  );

INSERT INTO public.schema_migrations (version, description)
VALUES ('0285', 'M7 F7.4 (validation-contract §4.4): add FORCE ROW LEVEL SECURITY + tenant_isolation policy (on actor_tenant_id) to public.approval_signoffs -- the one tenant table the tenant_id-keyed FORCE-RLS census missed (it keys on actor_tenant_id). Idiom verbatim from 0281/0278 incl. null-GUC escape-hatch branch (ADR 0027). No cross-tenant co-sign, so strict policy is correct.')
ON CONFLICT (version) DO NOTHING;

COMMIT;
