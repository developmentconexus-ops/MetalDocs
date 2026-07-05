-- 0279_tenant_lifecycle_jobs_tripwire_export_erase_cap.sql
-- M7 F7.3 Task A (global-maximum-remediation, milestone-7-tenant-lifecycle,
-- ADR 0070): the tenant export/erase workflows each assert exactly one of
-- tenant.export / tenant.erase (both new capabilities, ADR 0070) then INSERT
-- metaldocs.tenant_lifecycle_jobs (kind discriminates which). Before this
-- migration metaldocs.tenant_lifecycle_jobs (created by 0278) carried NO
-- trg_require_cap_asserted trigger at all — the enqueue INSERT was a
-- privileged provisioning write with no DB tripwire backstop.
--
-- Fix: (1) add a tenant_lifecycle_jobs/INSERT CASE branch requiring
-- {tenant.export, tenant.erase} (match-one / OR semantics, mirroring the
-- documents/UPDATE (#6) and controlled_documents/UPDATE (#8) multi-cap arm
-- precedent); (2) one-time trigger attachment on
-- metaldocs.tenant_lifecycle_jobs (0259/0277 attachment precedent), BEFORE
-- INSERT only to match the arm — the fail-closed ELSE branch would P0001
-- every UPDATE/DELETE if the trigger fired for ops the CASE does not map, and
-- no UPDATE/DELETE writer exists yet at this task boundary (the async
-- worker's status-transition writes are a separate later task).
--
-- Additive only — no cap is removed from any arm. This migration is
-- machine-generated from internal/platform/tripwire (TripwireArms +
-- RenderMigration) — see docs/superpowers/milestones/
-- global-maximum-remediation/milestone-7-tenant-lifecycle/
-- validation-contract.md §5 (touchpoint 6) and the M2 regeneration protocol
-- (milestone-2-authz-enforcement-generation/validation-contract.md §1.2/§1.4).
-- Every other CASE branch is reproduced byte-for-byte from 0277. Supersedes
-- 0277 as the latest definition of public.enforce_capability_asserted().

BEGIN;

CREATE OR REPLACE FUNCTION public.enforce_capability_asserted() RETURNS trigger
    LANGUAGE plpgsql SECURITY DEFINER
    SET search_path TO 'pg_catalog', 'pg_temp'
    AS $$
DECLARE
  v_bypass        TEXT;
  v_asserted_raw  TEXT;
  v_asserted      JSONB;
  v_required_caps TEXT[];
  v_tenant_id     UUID;
  v_cap_found     BOOLEAN := FALSE;
  v_element       JSONB;
BEGIN
  CASE
    WHEN TG_TABLE_NAME = 'approval_instances' AND TG_OP = 'INSERT' THEN
      v_required_caps := ARRAY['document.submit'];
      v_tenant_id     := NEW.tenant_id;
    WHEN TG_TABLE_NAME = 'approval_signoffs' AND TG_OP = 'INSERT' THEN
      v_required_caps := ARRAY['document.signoff'];
      v_tenant_id     := NEW.actor_tenant_id;
    WHEN TG_TABLE_NAME = 'iam_user_roles' THEN
      v_required_caps := ARRAY['user.manage'];
      v_tenant_id     := NEW.tenant_id;
    WHEN TG_TABLE_NAME = 'user_process_areas' THEN
      v_required_caps := ARRAY['membership.manage'];
      v_tenant_id     := NEW.tenant_id;
    WHEN TG_TABLE_NAME = 'documents' AND TG_OP = 'INSERT' THEN
      v_required_caps := ARRAY['document.create'];
      v_tenant_id     := NEW.tenant_id;
    WHEN TG_TABLE_NAME = 'documents' AND TG_OP = 'UPDATE' THEN
      -- 0271: 'document.obsolete' and 'membership.manage' added — two
      -- function-local write-paths assert only one of these caps (no
      -- co-asserted document.edit) and were fail-closed P0001 for every
      -- actor: ForceReleaseSession/ForceReleaseSessionTx
      -- (documents/repository/repository.go:798,828) and MarkObsolete
      -- (documents/approval/application/obsolete_service.go:88->93).
      -- 0275 (M6 F6.2): 'document.review' added — the mark-reviewed
      -- workflow asserts only document.review then UPDATEs documents
      -- (last_reviewed_at + review_due_at); see file header.
      v_required_caps := ARRAY['document.edit', 'document.obsolete', 'membership.manage', 'document.review'];
      v_tenant_id     := NEW.tenant_id;
    WHEN TG_TABLE_NAME = 'controlled_documents' AND TG_OP = 'INSERT' THEN
      v_required_caps := ARRAY['controlled_documents.create'];
      v_tenant_id     := NEW.tenant_id;
    WHEN TG_TABLE_NAME = 'controlled_documents' AND TG_OP = 'UPDATE' THEN
      v_required_caps := ARRAY['controlled_documents.obsolete', 'controlled_documents.supersede'];
      v_tenant_id     := NEW.tenant_id;
    WHEN TG_TABLE_NAME = 'cd_sequence_counters' THEN
      v_required_caps := ARRAY['controlled_documents.create'];
      v_tenant_id     := NEW.tenant_id;
    WHEN TG_TABLE_NAME = 'document_profiles' THEN
      v_required_caps := ARRAY['taxonomy.manage'];
      v_tenant_id     := NEW.tenant_id;
    WHEN TG_TABLE_NAME = 'document_process_areas' THEN
      v_required_caps := ARRAY['taxonomy.manage'];
      v_tenant_id     := NEW.tenant_id;
    WHEN TG_TABLE_NAME = 'document_families' THEN
      v_required_caps := ARRAY['taxonomy.manage'];
      v_tenant_id     := NEW.tenant_id;
    WHEN TG_TABLE_NAME = 'templates_template' THEN
      -- 0270: 'template.archive' added — Service.ArchiveTemplate updates this
      -- table under CapTemplateArchive; see file header.
      v_required_caps := ARRAY['template.create', 'template.edit', 'template.submit', 'template.approve', 'template.publish', 'template.archive'];
      v_tenant_id     := NEW.tenant_id;
    WHEN TG_TABLE_NAME = 'templates_template_version' THEN
      -- 0269: 'template.review' added — the reviewer stage (Service.Review)
      -- writes this table under CapTemplateReview; see file header.
      v_required_caps := ARRAY['template.create', 'template.edit', 'template.submit', 'template.review', 'template.approve', 'template.publish'];
      v_tenant_id     := NEW.tenant_id;
    -- ── New branches (SEC-05 / T-004 residual) ──────────────────────────────
    WHEN TG_TABLE_NAME = 'iam_users' THEN
      v_required_caps := ARRAY['user.manage'];
      v_tenant_id     := NEW.tenant_id;
    WHEN TG_TABLE_NAME = 'iam_groups' THEN
      v_required_caps := ARRAY['user.manage'];
      v_tenant_id     := NEW.tenant_id;
    WHEN TG_TABLE_NAME = 'iam_group_members' THEN
      v_required_caps := ARRAY['user.manage'];
      v_tenant_id     := NEW.tenant_id;
    WHEN TG_TABLE_NAME = 'iam_group_roles' THEN
      v_required_caps := ARRAY['user.manage'];
      -- iam_group_roles has no tenant_id column (scoped transitively via
      -- group_id -> iam_groups.tenant_id); mirrors the document_families
      -- (pre-0258) / templates_v2_template_version (0188) precedent for
      -- tenant_id-less tables.
      v_tenant_id     := NULL;
    WHEN TG_TABLE_NAME = 'tenants' AND TG_OP = 'INSERT' THEN
      -- 0277 (M7 F7.2, ADR 0070): tenant onboarding — OnboardTenant asserts
      -- tenant.onboard then INSERTs metaldocs.tenants. tenants has no
      -- tenant_id column: NEW.id IS the tenant id (the row being
      -- provisioned is the tenant itself).
      v_required_caps := ARRAY['tenant.onboard'];
      v_tenant_id     := NEW.id;
    WHEN TG_TABLE_NAME = 'tenant_lifecycle_jobs' AND TG_OP = 'INSERT' THEN
      -- 0279 (M7 F7.3, ADR 0070): the export/erase handlers each assert
      -- exactly one of tenant.export / tenant.erase then INSERT
      -- metaldocs.tenant_lifecycle_jobs (kind discriminates which).
      v_required_caps := ARRAY['tenant.export', 'tenant.erase'];
      v_tenant_id     := NEW.tenant_id;
    ELSE
      -- Fail-closed: a table carrying this trigger with no capability mapping is a
      -- wiring error, not a license to pass through. Refuse the write loudly.
      RAISE EXCEPTION 'ErrCapabilityNotAsserted: no capability mapping for table % (op %); trg_require_cap_asserted attached without an enforce_capability_asserted CASE branch', TG_TABLE_NAME, TG_OP
        USING ERRCODE = 'P0001';
  END CASE;

  v_bypass := pg_catalog.current_setting('metaldocs.bypass_authz', true);
  IF v_bypass IS NOT NULL AND v_bypass <> '' THEN
    IF v_bypass = 'scheduler' THEN
      BEGIN
        INSERT INTO public.governance_events
          (tenant_id, event_type, actor_user_id, resource_type, resource_id, reason, payload_json)
        VALUES (
          v_tenant_id,
          'authz.bypass_used',
          'system:scheduler',
          TG_TABLE_NAME,
          COALESCE(NEW.id::TEXT, 'unknown'),
          'scheduler bypass for ' || v_required_caps[1],
          pg_catalog.to_jsonb(jsonb_build_object(
            'required_caps', to_jsonb(v_required_caps),
            'bypass_token',  v_bypass,
            'table',         TG_TABLE_NAME,
            'op',            TG_OP
          ))
        );
      EXCEPTION WHEN others THEN
        RAISE NOTICE 'enforce_capability_asserted: governance_events insert failed: %', SQLERRM;
      END;
      RETURN NEW;
    ELSE
      RAISE EXCEPTION 'ErrCapabilityNotAsserted: unrecognised bypass token; caps % required on %', v_required_caps, TG_TABLE_NAME
        USING ERRCODE = 'P0001';
    END IF;
  END IF;

  v_asserted_raw := pg_catalog.current_setting('metaldocs.asserted_caps', true);
  IF v_asserted_raw IS NULL OR v_asserted_raw = '' THEN
    RAISE EXCEPTION 'ErrCapabilityNotAsserted: one of % required but metaldocs.asserted_caps is not set on %', v_required_caps, TG_TABLE_NAME
      USING ERRCODE = 'P0001';
  END IF;

  BEGIN
    v_asserted := v_asserted_raw::JSONB;
  EXCEPTION WHEN invalid_text_representation OR others THEN
    RAISE EXCEPTION 'ErrCapabilityNotAsserted: metaldocs.asserted_caps is not valid JSONB (caps % required)', v_required_caps
      USING ERRCODE = 'P0001';
  END;

  IF jsonb_typeof(v_asserted) <> 'array' THEN
    RAISE EXCEPTION 'ErrCapabilityNotAsserted: metaldocs.asserted_caps must be a JSONB array (caps % required)', v_required_caps
      USING ERRCODE = 'P0001';
  END IF;

  FOR v_element IN SELECT * FROM jsonb_array_elements(v_asserted) LOOP
    IF (v_element->>'cap') = ANY(v_required_caps) THEN
      v_cap_found := TRUE;
      EXIT;
    END IF;
  END LOOP;

  IF NOT v_cap_found THEN
    RAISE EXCEPTION 'ErrCapabilityNotAsserted: none of % present in asserted_caps on %', v_required_caps, TG_TABLE_NAME
      USING ERRCODE = 'P0001';
  END IF;

  RETURN NEW;
END;
$$;

-- ── trigger attachment ───────────────────────────────────────────────────────────────────

-- One-time attachment on metaldocs.tenants (0259 attachment precedent:
-- DROP TRIGGER IF EXISTS + CREATE TRIGGER trg_require_cap_asserted). INSERT
-- only, matching the arm: the trigger's fail-closed ELSE branch would P0001
-- every UPDATE/DELETE if attached for ops the CASE does not map, and no
-- tenants update/offboard surface exists yet (F7.3+ owns lifecycle).
DROP TRIGGER IF EXISTS trg_require_cap_asserted ON metaldocs.tenants;
CREATE TRIGGER trg_require_cap_asserted
  BEFORE INSERT ON metaldocs.tenants
  FOR EACH ROW EXECUTE FUNCTION public.enforce_capability_asserted();

-- One-time attachment on metaldocs.tenant_lifecycle_jobs (0259/0277
-- attachment precedent). INSERT only, matching the arm — the table's only
-- mutation surface at this stage is the enqueue INSERT (Task A); no
-- UPDATE/DELETE writer exists yet (the worker's status transitions are a
-- separate later task and would need their own arm/attachment then).
DROP TRIGGER IF EXISTS trg_require_cap_asserted ON metaldocs.tenant_lifecycle_jobs;
CREATE TRIGGER trg_require_cap_asserted
  BEFORE INSERT ON metaldocs.tenant_lifecycle_jobs
  FOR EACH ROW EXECUTE FUNCTION public.enforce_capability_asserted();

-- ── schema_migrations ledger ─────────────────────────────────────────────────────────────

INSERT INTO public.schema_migrations (version, description)
VALUES ('0279', 'M7 F7.3 Task A (ADR 0070): add tenant_lifecycle_jobs/INSERT tripwire arm requiring {tenant.export, tenant.erase} (match-one) and attach trg_require_cap_asserted to metaldocs.tenant_lifecycle_jobs (BEFORE INSERT) -- the export/erase enqueue workflows assert only one of tenant.export/tenant.erase then INSERT tenant_lifecycle_jobs, previously with no DB tripwire backstop. Additive-only; all other branches preserved from 0277; machine-generated from internal/platform/tripwire.')
ON CONFLICT (version) DO NOTHING;

COMMIT;
