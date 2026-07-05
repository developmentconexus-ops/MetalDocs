package tripwire

import (
	"fmt"
	"strings"
)

// RenderMigration renders the full 0275 forward-only migration SQL,
// regenerated from the prior tripwire migration (0271) with every CASE branch
// preserved byte-for-byte except the documents/UPDATE arm, whose v_required_caps
// literal is produced from TripwireArms in the same order as M2
// validation-contract.md §1.2 as extended by M6 §3 (gains document.review).
// Determinism: TripwireArms is a fixed slice, so this returns identical output
// every call.
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
		"-- ── schema_migrations ledger ─────────────────────────────────────────────────────────────\n" +
		"\n" +
		"INSERT INTO public.schema_migrations (version, description)\n" +
		"VALUES ('0275', '" + ledgerDescription + "')\n" +
		"ON CONFLICT (version) DO NOTHING;\n" +
		"\n" +
		"COMMIT;\n"
}

// migrationHeader is the file-header comment block for 0275, in 0269/0270/0271
// house style: goal/incident framing, root cause, writer inventory, fix
// statement.
const migrationHeader = `-- 0275_documents_update_tripwire_review_cap.sql
-- M6 F6.2 (global-maximum-remediation, milestone-6-eqms-review-reason):
-- the eQMS periodic-review mark-reviewed workflow asserts ONLY document.review
-- (a new capability, ADR 0069) then UPDATEs public.documents (SET
-- last_reviewed_at, review_due_at). enforce_capability_asserted()'s
-- documents/UPDATE branch (post-0271) accepts {document.edit, document.obsolete,
-- membership.manage} — document.review is absent — so every mark-reviewed
-- UPDATE would fail-closed with SQLSTATE P0001 for EVERY actor (the trigger
-- checks the recorded asserted-cap set, not the role; authz.Require records
-- whatever cap was actually asserted). This is the exact defect class as the
-- 0269/0270/0271 incidents: a single non-arm capability asserted in the same
-- function as the mutating UPDATE.
--
-- Never caught for the same reason as 0269/0270/0271: application tests are
-- sqlmock and cannot exercise the live trigger; only an integration drive
-- against a tripwire-enforced DB can pin this
-- (tests/integration/documents/tripwire_documents_test.go, extended alongside
-- this migration).
--
-- Fix: widen the documents/UPDATE arm to
-- {document.edit, document.obsolete, membership.manage, document.review}.
-- Additive only — no cap is removed from any arm. This migration is
-- machine-generated from internal/platform/tripwire (TripwireArms +
-- RenderMigration) — see docs/superpowers/milestones/
-- global-maximum-remediation/milestone-6-eqms-review-reason/
-- validation-contract.md §3 (touchpoint 6) and the M2 regeneration protocol
-- (milestone-2-authz-enforcement-generation/validation-contract.md §1.2/§1.4).
-- Every other CASE branch is reproduced byte-for-byte from 0271; this is a
-- function-only swap, no trigger-attachment change, no backfill. Supersedes
-- 0271 as the latest definition of public.enforce_capability_asserted().

`

const ledgerDescription = "M6 F6.2: widen documents/UPDATE tripwire arm to {document.edit, document.obsolete, membership.manage, document.review} -- the mark-reviewed workflow asserts only document.review (ADR 0069) then UPDATEs documents (last_reviewed_at, review_due_at) and would be fail-closed P0001 for every actor without the arm, same defect class as 0269/0270/0271. Additive-only; all other branches preserved from 0271; machine-generated from internal/platform/tripwire."

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
