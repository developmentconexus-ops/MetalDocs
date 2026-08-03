-- 0282_user_process_areas_erasure_delete.sql
-- M7 F7.3 Task E (global-maximum-remediation, milestone-7-tenant-lifecycle,
-- ADR 0070): public.user_process_areas carries
-- trg_user_process_areas_no_delete -> public.reject_user_process_areas_delete(),
-- which unconditionally RAISE EXCEPTIONs on any DELETE ("rows cannot be
-- deleted (revoke via UPDATE effective_to)", see db/baseline/0001_current_schema.sql
-- line 881). The F7.3 erase workflow's iam TenantDataPort
-- (internal/modules/iam/infrastructure/postgres/tenant_data_port.go)
-- hard-DELETEs public.user_process_areas rows per contract §3.2 (retention
-- via UPDATE is physically impossible — the rows FK to iam_users and
-- document_process_areas, both of which erasure also deletes). Without this
-- amendment every tenant erasure would P0001 on this table unconditionally.
--
-- Fix: amend the function so it allows the DELETE exactly when the
-- transaction has set the tenant-erasure lifecycle context GUC
-- (metaldocs.tenant_erasure — a NEW, erase-workflow-only marker, orthogonal
-- to metaldocs.bypass_authz which the tripwire's cap-assertion check uses).
-- Every other caller (normal operation — no erasure GUC set) still gets the
-- identical rejection, same error message, same ERRCODE. The function body
-- is otherwise unchanged.
--
-- Why a dedicated GUC rather than reusing metaldocs.bypass_authz: bypass_authz
-- answers "is a capability asserted for this write", a wholly different
-- question from "is this DELETE part of an authorized tenant-erasure
-- lifecycle transaction". Conflating them would let any bypass_authz=scheduler
-- background job silently delete process-area grants outside of erasure,
-- which is not the intended blast radius of that GUC. The erase orchestrator
-- (internal/modules/iam/application/tenant_lifecycle_service.go) is the only
-- caller expected to ever set metaldocs.tenant_erasure.

BEGIN;

CREATE OR REPLACE FUNCTION public.reject_user_process_areas_delete() RETURNS trigger
    LANGUAGE plpgsql
    SET search_path TO 'pg_catalog', 'pg_temp'
    AS $$
BEGIN
  IF NULLIF(current_setting('metaldocs.tenant_erasure', true), '') IS NOT NULL THEN
    RETURN OLD;
  END IF;
  RAISE EXCEPTION 'user_process_areas rows cannot be deleted (revoke via UPDATE effective_to)'
    USING ERRCODE = 'check_violation';
END;
$$;

-- ── schema_migrations ledger ─────────────────────────────────────────────────

INSERT INTO public.schema_migrations (version, description)
VALUES ('0282', 'M7 F7.3 Task E: amend public.reject_user_process_areas_delete() to allow DELETE on public.user_process_areas when the tx-local metaldocs.tenant_erasure GUC is set (tenant erasure lifecycle context); all other callers still P0001 with the identical message.')
ON CONFLICT (version) DO NOTHING;

COMMIT;
