package tripwire

import (
	"fmt"
	"strings"
)

// RenderMigration renders the full 0279 forward-only migration SQL,
// regenerated from the prior tripwire migration (0277) with every CASE branch
// preserved byte-for-byte plus one new branch — tenant_lifecycle_jobs/INSERT
// (M7 F7.3, ADR 0070) — whose v_required_caps literal is produced from
// TripwireArms entry #20, and the one-time trigger attachment on
// metaldocs.tenant_lifecycle_jobs (0259/0277 attachment precedent). Branch
// order follows M2 validation-contract.md §1.2 as extended by M6 §3
// (documents/UPDATE gains document.review), M7 §5 (tenants/INSERT arm) and M7
// §5 (tenant_lifecycle_jobs/INSERT arm). Determinism: TripwireArms is a fixed
// slice, so this returns identical output every call.
func RenderMigration() string {
	return migrationHeader +
		"BEGIN;\n" +
		"\n" +
		"CREATE OR REPLACE FUNCTION public.enforce_capability_asserted() RETURNS trigger\n" +
		"    LANGUAGE plpgsql SECURITY DEFINER\n" +
		"    SET search_path TO 'pg_catalog', 'pg_temp'\n" +
		"    AS $$\n" +
		"DECLARE\n" +
		"  v_bypass        TEXT;\n" +
		"  v_asserted_raw  TEXT;\n" +
		"  v_asserted      JSONB;\n" +
		"  v_required_caps TEXT[];\n" +
		"  v_tenant_id     UUID;\n" +
		"  v_cap_found     BOOLEAN := FALSE;\n" +
		"  v_element       JSONB;\n" +
		"BEGIN\n" +
		"  CASE\n" +
		"    WHEN TG_TABLE_NAME = 'approval_instances' AND TG_OP = 'INSERT' THEN\n" +
		"      v_required_caps := " + renderArray(findArm("approval_instances", OpInsert)) + ";\n" +
		"      v_tenant_id     := NEW.tenant_id;\n" +
		"    WHEN TG_TABLE_NAME = 'approval_signoffs' AND TG_OP = 'INSERT' THEN\n" +
		"      v_required_caps := " + renderArray(findArm("approval_signoffs", OpInsert)) + ";\n" +
		"      v_tenant_id     := NEW.actor_tenant_id;\n" +
		"    WHEN TG_TABLE_NAME = 'iam_user_roles' THEN\n" +
		"      v_required_caps := " + renderArray(findArm("iam_user_roles", OpAny)) + ";\n" +
		"      v_tenant_id     := NEW.tenant_id;\n" +
		"    WHEN TG_TABLE_NAME = 'user_process_areas' THEN\n" +
		"      v_required_caps := " + renderArray(findArm("user_process_areas", OpAny)) + ";\n" +
		"      v_tenant_id     := NEW.tenant_id;\n" +
		"    WHEN TG_TABLE_NAME = 'documents' AND TG_OP = 'INSERT' THEN\n" +
		"      v_required_caps := " + renderArray(findArm("documents", OpInsert)) + ";\n" +
		"      v_tenant_id     := NEW.tenant_id;\n" +
		"    WHEN TG_TABLE_NAME = 'documents' AND TG_OP = 'UPDATE' THEN\n" +
		"      -- 0271: 'document.obsolete' and 'membership.manage' added — two\n" +
		"      -- function-local write-paths assert only one of these caps (no\n" +
		"      -- co-asserted document.edit) and were fail-closed P0001 for every\n" +
		"      -- actor: ForceReleaseSession/ForceReleaseSessionTx\n" +
		"      -- (documents/repository/repository.go:798,828) and MarkObsolete\n" +
		"      -- (documents/approval/application/obsolete_service.go:88->93).\n" +
		"      -- 0275 (M6 F6.2): 'document.review' added — the mark-reviewed\n" +
		"      -- workflow asserts only document.review then UPDATEs documents\n" +
		"      -- (last_reviewed_at + review_due_at); see file header.\n" +
		"      v_required_caps := " + renderArray(findArm("documents", OpUpdate)) + ";\n" +
		"      v_tenant_id     := NEW.tenant_id;\n" +
		"    WHEN TG_TABLE_NAME = 'controlled_documents' AND TG_OP = 'INSERT' THEN\n" +
		"      v_required_caps := " + renderArray(findArm("controlled_documents", OpInsert)) + ";\n" +
		"      v_tenant_id     := NEW.tenant_id;\n" +
		"    WHEN TG_TABLE_NAME = 'controlled_documents' AND TG_OP = 'UPDATE' THEN\n" +
		"      v_required_caps := " + renderArray(findArm("controlled_documents", OpUpdate)) + ";\n" +
		"      v_tenant_id     := NEW.tenant_id;\n" +
		"    WHEN TG_TABLE_NAME = 'cd_sequence_counters' THEN\n" +
		"      v_required_caps := " + renderArray(findArm("cd_sequence_counters", OpAny)) + ";\n" +
		"      v_tenant_id     := NEW.tenant_id;\n" +
		"    WHEN TG_TABLE_NAME = 'document_profiles' THEN\n" +
		"      v_required_caps := " + renderArray(findArm("document_profiles", OpAny)) + ";\n" +
		"      v_tenant_id     := NEW.tenant_id;\n" +
		"    WHEN TG_TABLE_NAME = 'document_process_areas' THEN\n" +
		"      v_required_caps := " + renderArray(findArm("document_process_areas", OpAny)) + ";\n" +
		"      v_tenant_id     := NEW.tenant_id;\n" +
		"    WHEN TG_TABLE_NAME = 'document_families' THEN\n" +
		"      v_required_caps := " + renderArray(findArm("document_families", OpAny)) + ";\n" +
		"      v_tenant_id     := NEW.tenant_id;\n" +
		"    WHEN TG_TABLE_NAME = 'templates_template' THEN\n" +
		"      -- 0270: 'template.archive' added — Service.ArchiveTemplate updates this\n" +
		"      -- table under CapTemplateArchive; see file header.\n" +
		"      v_required_caps := " + renderArray(findArm("templates_template", OpAny)) + ";\n" +
		"      v_tenant_id     := NEW.tenant_id;\n" +
		"    WHEN TG_TABLE_NAME = 'templates_template_version' THEN\n" +
		"      -- 0269: 'template.review' added — the reviewer stage (Service.Review)\n" +
		"      -- writes this table under CapTemplateReview; see file header.\n" +
		"      v_required_caps := " + renderArray(findArm("templates_template_version", OpAny)) + ";\n" +
		"      v_tenant_id     := NEW.tenant_id;\n" +
		"    -- ── New branches (SEC-05 / T-004 residual) ──────────────────────────────\n" +
		"    WHEN TG_TABLE_NAME = 'iam_users' THEN\n" +
		"      v_required_caps := " + renderArray(findArm("iam_users", OpAny)) + ";\n" +
		"      v_tenant_id     := NEW.tenant_id;\n" +
		"    WHEN TG_TABLE_NAME = 'iam_groups' THEN\n" +
		"      v_required_caps := " + renderArray(findArm("iam_groups", OpAny)) + ";\n" +
		"      v_tenant_id     := NEW.tenant_id;\n" +
		"    WHEN TG_TABLE_NAME = 'iam_group_members' THEN\n" +
		"      v_required_caps := " + renderArray(findArm("iam_group_members", OpAny)) + ";\n" +
		"      v_tenant_id     := NEW.tenant_id;\n" +
		"    WHEN TG_TABLE_NAME = 'iam_group_roles' THEN\n" +
		"      v_required_caps := " + renderArray(findArm("iam_group_roles", OpAny)) + ";\n" +
		"      -- iam_group_roles has no tenant_id column (scoped transitively via\n" +
		"      -- group_id -> iam_groups.tenant_id); mirrors the document_families\n" +
		"      -- (pre-0258) / templates_v2_template_version (0188) precedent for\n" +
		"      -- tenant_id-less tables.\n" +
		"      v_tenant_id     := NULL;\n" +
		"    WHEN TG_TABLE_NAME = 'tenants' AND TG_OP = 'INSERT' THEN\n" +
		"      -- 0277 (M7 F7.2, ADR 0070): tenant onboarding — OnboardTenant asserts\n" +
		"      -- tenant.onboard then INSERTs metaldocs.tenants. tenants has no\n" +
		"      -- tenant_id column: NEW.id IS the tenant id (the row being\n" +
		"      -- provisioned is the tenant itself).\n" +
		"      v_required_caps := " + renderArray(findArm("tenants", OpInsert)) + ";\n" +
		"      v_tenant_id     := NEW.id;\n" +
		"    WHEN TG_TABLE_NAME = 'tenant_lifecycle_jobs' AND TG_OP = 'INSERT' THEN\n" +
		"      -- 0279 (M7 F7.3, ADR 0070): the export/erase handlers each assert\n" +
		"      -- exactly one of tenant.export / tenant.erase then INSERT\n" +
		"      -- metaldocs.tenant_lifecycle_jobs (kind discriminates which).\n" +
		"      v_required_caps := " + renderArray(findArm("tenant_lifecycle_jobs", OpInsert)) + ";\n" +
		"      v_tenant_id     := NEW.tenant_id;\n" +
		"    ELSE\n" +
		"      -- Fail-closed: a table carrying this trigger with no capability mapping is a\n" +
		"      -- wiring error, not a license to pass through. Refuse the write loudly.\n" +
		"      RAISE EXCEPTION 'ErrCapabilityNotAsserted: no capability mapping for table % (op %); trg_require_cap_asserted attached without an enforce_capability_asserted CASE branch', TG_TABLE_NAME, TG_OP\n" +
		"        USING ERRCODE = 'P0001';\n" +
		"  END CASE;\n" +
		"\n" +
		"  v_bypass := pg_catalog.current_setting('metaldocs.bypass_authz', true);\n" +
		"  IF v_bypass IS NOT NULL AND v_bypass <> '' THEN\n" +
		"    IF v_bypass = 'scheduler' THEN\n" +
		"      BEGIN\n" +
		"        INSERT INTO public.governance_events\n" +
		"          (tenant_id, event_type, actor_user_id, resource_type, resource_id, reason, payload_json)\n" +
		"        VALUES (\n" +
		"          v_tenant_id,\n" +
		"          'authz.bypass_used',\n" +
		"          'system:scheduler',\n" +
		"          TG_TABLE_NAME,\n" +
		"          COALESCE(NEW.id::TEXT, 'unknown'),\n" +
		"          'scheduler bypass for ' || v_required_caps[1],\n" +
		"          pg_catalog.to_jsonb(jsonb_build_object(\n" +
		"            'required_caps', to_jsonb(v_required_caps),\n" +
		"            'bypass_token',  v_bypass,\n" +
		"            'table',         TG_TABLE_NAME,\n" +
		"            'op',            TG_OP\n" +
		"          ))\n" +
		"        );\n" +
		"      EXCEPTION WHEN others THEN\n" +
		"        RAISE NOTICE 'enforce_capability_asserted: governance_events insert failed: %', SQLERRM;\n" +
		"      END;\n" +
		"      RETURN NEW;\n" +
		"    ELSE\n" +
		"      RAISE EXCEPTION 'ErrCapabilityNotAsserted: unrecognised bypass token; caps % required on %', v_required_caps, TG_TABLE_NAME\n" +
		"        USING ERRCODE = 'P0001';\n" +
		"    END IF;\n" +
		"  END IF;\n" +
		"\n" +
		"  v_asserted_raw := pg_catalog.current_setting('metaldocs.asserted_caps', true);\n" +
		"  IF v_asserted_raw IS NULL OR v_asserted_raw = '' THEN\n" +
		"    RAISE EXCEPTION 'ErrCapabilityNotAsserted: one of % required but metaldocs.asserted_caps is not set on %', v_required_caps, TG_TABLE_NAME\n" +
		"      USING ERRCODE = 'P0001';\n" +
		"  END IF;\n" +
		"\n" +
		"  BEGIN\n" +
		"    v_asserted := v_asserted_raw::JSONB;\n" +
		"  EXCEPTION WHEN invalid_text_representation OR others THEN\n" +
		"    RAISE EXCEPTION 'ErrCapabilityNotAsserted: metaldocs.asserted_caps is not valid JSONB (caps % required)', v_required_caps\n" +
		"      USING ERRCODE = 'P0001';\n" +
		"  END;\n" +
		"\n" +
		"  IF jsonb_typeof(v_asserted) <> 'array' THEN\n" +
		"    RAISE EXCEPTION 'ErrCapabilityNotAsserted: metaldocs.asserted_caps must be a JSONB array (caps % required)', v_required_caps\n" +
		"      USING ERRCODE = 'P0001';\n" +
		"  END IF;\n" +
		"\n" +
		"  FOR v_element IN SELECT * FROM jsonb_array_elements(v_asserted) LOOP\n" +
		"    IF (v_element->>'cap') = ANY(v_required_caps) THEN\n" +
		"      v_cap_found := TRUE;\n" +
		"      EXIT;\n" +
		"    END IF;\n" +
		"  END LOOP;\n" +
		"\n" +
		"  IF NOT v_cap_found THEN\n" +
		"    RAISE EXCEPTION 'ErrCapabilityNotAsserted: none of % present in asserted_caps on %', v_required_caps, TG_TABLE_NAME\n" +
		"      USING ERRCODE = 'P0001';\n" +
		"  END IF;\n" +
		"\n" +
		"  RETURN NEW;\n" +
		"END;\n" +
		"$$;\n" +
		"\n" +
		"-- ── trigger attachment ───────────────────────────────────────────────────────────────────\n" +
		"\n" +
		"-- One-time attachment on metaldocs.tenants (0259 attachment precedent:\n" +
		"-- DROP TRIGGER IF EXISTS + CREATE TRIGGER trg_require_cap_asserted). INSERT\n" +
		"-- only, matching the arm: the trigger's fail-closed ELSE branch would P0001\n" +
		"-- every UPDATE/DELETE if attached for ops the CASE does not map, and no\n" +
		"-- tenants update/offboard surface exists yet (F7.3+ owns lifecycle).\n" +
		"DROP TRIGGER IF EXISTS trg_require_cap_asserted ON metaldocs.tenants;\n" +
		"CREATE TRIGGER trg_require_cap_asserted\n" +
		"  BEFORE INSERT ON metaldocs.tenants\n" +
		"  FOR EACH ROW EXECUTE FUNCTION public.enforce_capability_asserted();\n" +
		"\n" +
		"-- One-time attachment on metaldocs.tenant_lifecycle_jobs (0259/0277\n" +
		"-- attachment precedent). INSERT only, matching the arm — the table's only\n" +
		"-- mutation surface at this stage is the enqueue INSERT (Task A); no\n" +
		"-- UPDATE/DELETE writer exists yet (the worker's status transitions are a\n" +
		"-- separate later task and would need their own arm/attachment then).\n" +
		"DROP TRIGGER IF EXISTS trg_require_cap_asserted ON metaldocs.tenant_lifecycle_jobs;\n" +
		"CREATE TRIGGER trg_require_cap_asserted\n" +
		"  BEFORE INSERT ON metaldocs.tenant_lifecycle_jobs\n" +
		"  FOR EACH ROW EXECUTE FUNCTION public.enforce_capability_asserted();\n" +
		"\n" +
		"-- ── schema_migrations ledger ─────────────────────────────────────────────────────────────\n" +
		"\n" +
		"INSERT INTO public.schema_migrations (version, description)\n" +
		"VALUES ('0279', '" + ledgerDescription + "')\n" +
		"ON CONFLICT (version) DO NOTHING;\n" +
		"\n" +
		"COMMIT;\n"
}

// migrationHeader is the file-header comment block for 0279, in
// 0269/0270/0271/0275/0277 house style: goal/incident framing, root cause,
// writer inventory, fix statement.
const migrationHeader = `-- 0279_tenant_lifecycle_jobs_tripwire_export_erase_cap.sql
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

`

const ledgerDescription = "M7 F7.3 Task A (ADR 0070): add tenant_lifecycle_jobs/INSERT tripwire arm requiring {tenant.export, tenant.erase} (match-one) and attach trg_require_cap_asserted to metaldocs.tenant_lifecycle_jobs (BEFORE INSERT) -- the export/erase enqueue workflows assert only one of tenant.export/tenant.erase then INSERT tenant_lifecycle_jobs, previously with no DB tripwire backstop. Additive-only; all other branches preserved from 0277; machine-generated from internal/platform/tripwire."

// findArm returns the Arm for (table, op), panicking if absent — every
// branch rendered above must correspond to a TripwireArms entry (parity is
// structural, not just tested).
func findArm(table string, op Op) Arm {
	for _, arm := range TripwireArms {
		if arm.Table == table && arm.Op == op {
			return arm
		}
	}
	panic(fmt.Sprintf("tripwire: no TripwireArms entry for (%s, %s)", table, op))
}

// renderArray renders an Arm's Caps as a Postgres ARRAY[...] literal of
// single-quoted strings, in the same order as TripwireArms (contract §1.2
// order), matching 0270's literal style exactly (e.g.
// ARRAY['document.create']).
func renderArray(arm Arm) string {
	quoted := make([]string, len(arm.Caps))
	for i, cap := range arm.Caps {
		quoted[i] = "'" + string(cap) + "'"
	}
	return "ARRAY[" + strings.Join(quoted, ", ") + "]"
}
