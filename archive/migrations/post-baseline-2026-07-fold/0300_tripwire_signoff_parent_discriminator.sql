-- 0300_tripwire_signoff_parent_discriminator.sql
-- M3 P3.S2b-3b-iii-a (approval-remediation, kernel extraction, ADR 0083
-- follow-on -- extends ADR 0083, which itself extends ADR 0082):
-- approval_signoffs is INSERTed by both document and template signoff flows,
-- but has NO subject_kind column of its own (unlike approval_instances) --
-- its subject lives on the parent approval_instances row via
-- NEW.approval_instance_id. The capability tripwire
-- enforce_capability_asserted() hardcoded approval_signoffs INSERT ->
-- ARRAY['document.signoff'] (its Go source of truth, TripwireArms arm #2, had
-- no discriminator at all). A template signoff therefore fail-closed P0001
-- for every actor, and the "obvious" fix -- widening the arm to
-- ARRAY['document.signoff', 'template.approve'] -- would be a SECURITY
-- REGRESSION, exactly as ADR 0083 rejected for approval_instances: a
-- principal holding only template.approve would authorize a document
-- signoff (and vice versa), cross-contaminating the two subjects'
-- capability requirements on the shared table.
--
-- Fix, at the generator (internal/platform/tripwire/arms.go + render.go),
-- regenerated here: TripwireArms arm #2 splits into two
-- parent-lookup-discriminated entries (WhenParentTable "approval_instances",
-- WhenParentKeyCol "approval_instance_id", WhenParentDiscrimCol
-- "subject_kind", WhenValue "document" / "template"); RenderMigration emits,
-- inside the approval_signoffs/INSERT branch, a SELECT of the parent row's
-- subject_kind followed by a nested CASE on the looked-up value, assigning
-- v_required_caps := ARRAY['document.signoff'] or ARRAY['template.approve']
-- per parent subject (never unioned), with a fail-closed ELSE (RAISE P0001)
-- for any other/absent parent subject_kind.
--
-- No other arm/cap change of any kind -- every other CASE branch and its
-- v_required_caps literal are reproduced byte-for-byte from 0299. This
-- migration is machine-generated from internal/platform/tripwire
-- (TripwireArms + RenderMigration) per the M2 regeneration protocol
-- (milestone-2-authz-enforcement-generation/validation-contract.md §1.2/§1.4)
-- as amended by ADR 0083 and its follow-on. Supersedes 0299 as the latest
-- definition of public.enforce_capability_asserted().

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
  v_parent_subject_kind TEXT;
BEGIN
  CASE
    WHEN TG_TABLE_NAME = 'approval_instances' AND TG_OP = 'INSERT' THEN
      -- 0299 (ADR 0083): approval_instances is a shared (subject_kind,
      -- subject_key) kernel table (ADR 0082); the required capability is
      -- subject-discriminated so a flat match-one arm cannot express
      -- "document rows require document.submit; template rows require
      -- template.submit" without a cross-subject security regression
      -- (ADR 0083 "why the obvious widen is a security regression").
      -- Nested CASE on NEW.subject_kind; the two subjects' capability
      -- sets are never unioned.
      CASE NEW.subject_kind
        WHEN 'document' THEN
          v_required_caps := ARRAY['document.submit'];
        WHEN 'template' THEN
          v_required_caps := ARRAY['template.submit'];
        ELSE
          RAISE EXCEPTION 'ErrCapabilityNotAsserted: no capability mapping for approval_instances subject_kind %, op %; enforce_capability_asserted has no discriminated arm for this subject', NEW.subject_kind, TG_OP
            USING ERRCODE = 'P0001';
      END CASE;
      v_tenant_id     := COALESCE(NEW.tenant_id, OLD.tenant_id);
    WHEN TG_TABLE_NAME = 'approval_signoffs' AND TG_OP = 'INSERT' THEN
      -- 0300 (ADR 0083 follow-on, M3 P3.S2b-3b-iii-a): approval_signoffs has
      -- no direct subject_kind column of its own; the subject is resolved
      -- via the parent approval_instances row (NEW.approval_instance_id).
      -- Nested CASE on the looked-up parent subject_kind; the two
      -- subjects' capability sets are never unioned.
      SELECT subject_kind INTO v_parent_subject_kind FROM public.approval_instances WHERE id = NEW.approval_instance_id;
      CASE v_parent_subject_kind
        WHEN 'document' THEN
          v_required_caps := ARRAY['document.signoff'];
        WHEN 'template' THEN
          v_required_caps := ARRAY['template.approve'];
        ELSE
          RAISE EXCEPTION 'ErrCapabilityNotAsserted: no capability mapping for approval_signoffs parent subject_kind %, op %; enforce_capability_asserted has no discriminated arm for this parent subject (parent lookup via % = NEW.%)', v_parent_subject_kind, TG_OP, 'approval_instances', 'approval_instance_id'
            USING ERRCODE = 'P0001';
      END CASE;
      v_tenant_id     := NEW.actor_tenant_id;
    WHEN TG_TABLE_NAME = 'iam_user_roles' THEN
      v_required_caps := ARRAY['user.manage'];
      v_tenant_id     := COALESCE(NEW.tenant_id, OLD.tenant_id);
    WHEN TG_TABLE_NAME = 'user_process_areas' THEN
      v_required_caps := ARRAY['membership.manage'];
      v_tenant_id     := COALESCE(NEW.tenant_id, OLD.tenant_id);
    WHEN TG_TABLE_NAME = 'documents' AND TG_OP = 'INSERT' THEN
      v_required_caps := ARRAY['document.create'];
      v_tenant_id     := COALESCE(NEW.tenant_id, OLD.tenant_id);
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
      v_tenant_id     := COALESCE(NEW.tenant_id, OLD.tenant_id);
    WHEN TG_TABLE_NAME = 'controlled_documents' AND TG_OP = 'INSERT' THEN
      v_required_caps := ARRAY['controlled_documents.create'];
      v_tenant_id     := COALESCE(NEW.tenant_id, OLD.tenant_id);
    WHEN TG_TABLE_NAME = 'controlled_documents' AND TG_OP = 'UPDATE' THEN
      v_required_caps := ARRAY['controlled_documents.obsolete', 'controlled_documents.supersede'];
      v_tenant_id     := COALESCE(NEW.tenant_id, OLD.tenant_id);
    WHEN TG_TABLE_NAME = 'cd_sequence_counters' THEN
      v_required_caps := ARRAY['controlled_documents.create'];
      v_tenant_id     := COALESCE(NEW.tenant_id, OLD.tenant_id);
    WHEN TG_TABLE_NAME = 'document_profiles' THEN
      v_required_caps := ARRAY['taxonomy.manage'];
      v_tenant_id     := COALESCE(NEW.tenant_id, OLD.tenant_id);
    WHEN TG_TABLE_NAME = 'document_process_areas' THEN
      v_required_caps := ARRAY['taxonomy.manage'];
      v_tenant_id     := COALESCE(NEW.tenant_id, OLD.tenant_id);
    WHEN TG_TABLE_NAME = 'document_families' THEN
      v_required_caps := ARRAY['taxonomy.manage'];
      v_tenant_id     := COALESCE(NEW.tenant_id, OLD.tenant_id);
    WHEN TG_TABLE_NAME = 'templates_template' THEN
      -- 0270: 'template.archive' added — Service.ArchiveTemplate updates this
      -- table under CapTemplateArchive; see file header.
      v_required_caps := ARRAY['template.create', 'template.edit', 'template.submit', 'template.approve', 'template.publish', 'template.archive'];
      v_tenant_id     := COALESCE(NEW.tenant_id, OLD.tenant_id);
    WHEN TG_TABLE_NAME = 'templates_template_version' THEN
      -- 0269: 'template.review' added — the reviewer stage (Service.Review)
      -- writes this table under CapTemplateReview; see file header.
      v_required_caps := ARRAY['template.create', 'template.edit', 'template.submit', 'template.review', 'template.approve', 'template.publish'];
      v_tenant_id     := COALESCE(NEW.tenant_id, OLD.tenant_id);
    -- ── New branches (SEC-05 / T-004 residual) ──────────────────────────────
    WHEN TG_TABLE_NAME = 'iam_users' THEN
      v_required_caps := ARRAY['user.manage'];
      v_tenant_id     := COALESCE(NEW.tenant_id, OLD.tenant_id);
    WHEN TG_TABLE_NAME = 'iam_groups' THEN
      v_required_caps := ARRAY['user.manage'];
      v_tenant_id     := COALESCE(NEW.tenant_id, OLD.tenant_id);
    WHEN TG_TABLE_NAME = 'iam_group_members' THEN
      v_required_caps := ARRAY['user.manage'];
      v_tenant_id     := COALESCE(NEW.tenant_id, OLD.tenant_id);
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
      v_tenant_id     := COALESCE(NEW.tenant_id, OLD.tenant_id);
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
          COALESCE(NEW.id::TEXT, OLD.id::TEXT, 'unknown'),
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
      -- BEFORE-trigger return contract: returning NEW on a DELETE returns
      -- NULL (NEW is unassigned for DELETE), which SILENTLY CANCELS the
      -- row deletion. 0283 root-cause fix: return OLD for DELETE.
      IF TG_OP = 'DELETE' THEN
        RETURN OLD;
      END IF;
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

  -- Same BEFORE-trigger return contract as the bypass path above (0283).
  IF TG_OP = 'DELETE' THEN
    RETURN OLD;
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
VALUES ('0300', 'M3 P3.S2b-3b-iii-a (ADR 0083 follow-on): approval_signoffs/INSERT arm split into two parent-lookup-discriminated entries (parent approval_instances.subject_kind=document -> document.signoff; subject_kind=template -> template.approve, never unioned) via a SELECT of the parent approval_instances row subject_kind + a nested CASE with a fail-closed ELSE. No other arm/cap change; every other branch byte-for-byte from 0299; machine-generated from internal/platform/tripwire.')
ON CONFLICT (version) DO NOTHING;

COMMIT;
