package tripwire

import (
	"fmt"
	"strings"
)

// RenderMigration renders the full 0300 forward-only migration SQL,
// regenerated from the prior tripwire migration (0299) with every CASE branch
// preserved byte-for-byte EXCEPT approval_signoffs/INSERT, which becomes a
// parent-lookup nested CASE (ADR 0083 follow-on, M3 P3.S2b-3b-iii-a):
// approval_signoffs has no direct subject_kind column, so its subject is
// resolved via the parent approval_instances row (NEW.approval_instance_id) —
// parent subject_kind 'document' requires exactly document.signoff, parent
// subject_kind 'template' requires exactly template.approve, never unioned.
// Branch order follows M2 validation-contract.md §1.2 as extended by M6 §3
// (documents/UPDATE gains document.review), M7 §5 (tenants/INSERT arm), M7 §5
// (tenant_lifecycle_jobs/INSERT arm), ADR 0083 (approval_instances/INSERT
// subject discrimination), and ADR 0083's follow-on (approval_signoffs/INSERT
// parent-lookup subject discrimination). Determinism: TripwireArms is a fixed
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
		"  v_parent_subject_kind TEXT;\n" +
		"BEGIN\n" +
		"  CASE\n" +
		"    WHEN TG_TABLE_NAME = 'approval_instances' AND TG_OP = 'INSERT' THEN\n" +
		"      -- 0299 (ADR 0083): approval_instances is a shared (subject_kind,\n" +
		"      -- subject_key) kernel table (ADR 0082); the required capability is\n" +
		"      -- subject-discriminated so a flat match-one arm cannot express\n" +
		"      -- \"document rows require document.submit; template rows require\n" +
		"      -- template.submit\" without a cross-subject security regression\n" +
		"      -- (ADR 0083 \"why the obvious widen is a security regression\").\n" +
		"      -- Nested CASE on NEW.subject_kind; the two subjects' capability\n" +
		"      -- sets are never unioned.\n" +
		renderDiscriminatedCase(findArms("approval_instances", OpInsert)) +
		"      v_tenant_id     := COALESCE(NEW.tenant_id, OLD.tenant_id);\n" +
		"    WHEN TG_TABLE_NAME = 'approval_signoffs' AND TG_OP = 'INSERT' THEN\n" +
		"      -- 0300 (ADR 0083 follow-on, M3 P3.S2b-3b-iii-a): approval_signoffs has\n" +
		"      -- no direct subject_kind column of its own; the subject is resolved\n" +
		"      -- via the parent approval_instances row (NEW.approval_instance_id).\n" +
		"      -- Nested CASE on the looked-up parent subject_kind; the two\n" +
		"      -- subjects' capability sets are never unioned.\n" +
		renderParentLookupDiscriminatedCase(findArms("approval_signoffs", OpInsert)) +
		"      v_tenant_id     := NEW.actor_tenant_id;\n" +
		"    WHEN TG_TABLE_NAME = 'iam_user_roles' THEN\n" +
		"      v_required_caps := " + renderArray(findArm("iam_user_roles", OpAny)) + ";\n" +
		"      v_tenant_id     := COALESCE(NEW.tenant_id, OLD.tenant_id);\n" +
		"    WHEN TG_TABLE_NAME = 'user_process_areas' THEN\n" +
		"      v_required_caps := " + renderArray(findArm("user_process_areas", OpAny)) + ";\n" +
		"      v_tenant_id     := COALESCE(NEW.tenant_id, OLD.tenant_id);\n" +
		"    WHEN TG_TABLE_NAME = 'documents' AND TG_OP = 'INSERT' THEN\n" +
		"      v_required_caps := " + renderArray(findArm("documents", OpInsert)) + ";\n" +
		"      v_tenant_id     := COALESCE(NEW.tenant_id, OLD.tenant_id);\n" +
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
		"      v_tenant_id     := COALESCE(NEW.tenant_id, OLD.tenant_id);\n" +
		"    WHEN TG_TABLE_NAME = 'controlled_documents' AND TG_OP = 'INSERT' THEN\n" +
		"      v_required_caps := " + renderArray(findArm("controlled_documents", OpInsert)) + ";\n" +
		"      v_tenant_id     := COALESCE(NEW.tenant_id, OLD.tenant_id);\n" +
		"    WHEN TG_TABLE_NAME = 'controlled_documents' AND TG_OP = 'UPDATE' THEN\n" +
		"      v_required_caps := " + renderArray(findArm("controlled_documents", OpUpdate)) + ";\n" +
		"      v_tenant_id     := COALESCE(NEW.tenant_id, OLD.tenant_id);\n" +
		"    WHEN TG_TABLE_NAME = 'cd_sequence_counters' THEN\n" +
		"      v_required_caps := " + renderArray(findArm("cd_sequence_counters", OpAny)) + ";\n" +
		"      v_tenant_id     := COALESCE(NEW.tenant_id, OLD.tenant_id);\n" +
		"    WHEN TG_TABLE_NAME = 'document_profiles' THEN\n" +
		"      v_required_caps := " + renderArray(findArm("document_profiles", OpAny)) + ";\n" +
		"      v_tenant_id     := COALESCE(NEW.tenant_id, OLD.tenant_id);\n" +
		"    WHEN TG_TABLE_NAME = 'document_process_areas' THEN\n" +
		"      v_required_caps := " + renderArray(findArm("document_process_areas", OpAny)) + ";\n" +
		"      v_tenant_id     := COALESCE(NEW.tenant_id, OLD.tenant_id);\n" +
		"    WHEN TG_TABLE_NAME = 'document_families' THEN\n" +
		"      v_required_caps := " + renderArray(findArm("document_families", OpAny)) + ";\n" +
		"      v_tenant_id     := COALESCE(NEW.tenant_id, OLD.tenant_id);\n" +
		"    WHEN TG_TABLE_NAME = 'templates_template' THEN\n" +
		"      -- 0270: 'template.archive' added — Service.ArchiveTemplate updates this\n" +
		"      -- table under CapTemplateArchive; see file header.\n" +
		"      v_required_caps := " + renderArray(findArm("templates_template", OpAny)) + ";\n" +
		"      v_tenant_id     := COALESCE(NEW.tenant_id, OLD.tenant_id);\n" +
		"    WHEN TG_TABLE_NAME = 'templates_template_version' THEN\n" +
		"      -- 0269: 'template.review' added — the reviewer stage (Service.Review)\n" +
		"      -- writes this table under CapTemplateReview; see file header.\n" +
		"      v_required_caps := " + renderArray(findArm("templates_template_version", OpAny)) + ";\n" +
		"      v_tenant_id     := COALESCE(NEW.tenant_id, OLD.tenant_id);\n" +
		"    -- ── New branches (SEC-05 / T-004 residual) ──────────────────────────────\n" +
		"    WHEN TG_TABLE_NAME = 'iam_users' THEN\n" +
		"      v_required_caps := " + renderArray(findArm("iam_users", OpAny)) + ";\n" +
		"      v_tenant_id     := COALESCE(NEW.tenant_id, OLD.tenant_id);\n" +
		"    WHEN TG_TABLE_NAME = 'iam_groups' THEN\n" +
		"      v_required_caps := " + renderArray(findArm("iam_groups", OpAny)) + ";\n" +
		"      v_tenant_id     := COALESCE(NEW.tenant_id, OLD.tenant_id);\n" +
		"    WHEN TG_TABLE_NAME = 'iam_group_members' THEN\n" +
		"      v_required_caps := " + renderArray(findArm("iam_group_members", OpAny)) + ";\n" +
		"      v_tenant_id     := COALESCE(NEW.tenant_id, OLD.tenant_id);\n" +
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
		"      v_tenant_id     := COALESCE(NEW.tenant_id, OLD.tenant_id);\n" +
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
		"          COALESCE(NEW.id::TEXT, OLD.id::TEXT, 'unknown'),\n" +
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
		"      -- BEFORE-trigger return contract: returning NEW on a DELETE returns\n" +
		"      -- NULL (NEW is unassigned for DELETE), which SILENTLY CANCELS the\n" +
		"      -- row deletion. 0283 root-cause fix: return OLD for DELETE.\n" +
		"      IF TG_OP = 'DELETE' THEN\n" +
		"        RETURN OLD;\n" +
		"      END IF;\n" +
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
		"  -- Same BEFORE-trigger return contract as the bypass path above (0283).\n" +
		"  IF TG_OP = 'DELETE' THEN\n" +
		"    RETURN OLD;\n" +
		"  END IF;\n" +
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
		"VALUES ('0300', '" + ledgerDescription + "')\n" +
		"ON CONFLICT (version) DO NOTHING;\n" +
		"\n" +
		"COMMIT;\n"
}

// migrationHeader is the file-header comment block for 0300, in
// 0269/0270/0271/0275/0277/0283/0299 house style: goal/incident framing, root
// cause, writer inventory, fix statement.
const migrationHeader = `-- 0300_tripwire_signoff_parent_discriminator.sql
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

`

const ledgerDescription = "M3 P3.S2b-3b-iii-a (ADR 0083 follow-on): approval_signoffs/INSERT arm split into two parent-lookup-discriminated entries (parent approval_instances.subject_kind=document -> document.signoff; subject_kind=template -> template.approve, never unioned) via a SELECT of the parent approval_instances row subject_kind + a nested CASE with a fail-closed ELSE. No other arm/cap change; every other branch byte-for-byte from 0299; machine-generated from internal/platform/tripwire."

// findArm returns the single, undiscriminated Arm for (table, op), panicking
// if absent or if the pair is discriminated (multiple entries share the
// (table,op) — use findArms + renderDiscriminatedCase for those instead;
// every branch rendered above must correspond to a TripwireArms entry, parity
// is structural, not just tested).
func findArm(table string, op Op) Arm {
	arms := findArms(table, op)
	if len(arms) != 1 {
		panic(fmt.Sprintf("tripwire: findArm(%s, %s) expects exactly 1 undiscriminated arm, found %d — use findArms for a discriminated (table,op) pair", table, op, len(arms)))
	}
	return arms[0]
}

// findArms returns every TripwireArms entry for (table, op), in TripwireArms
// order, panicking if none exist. A (table,op) pair with more than one entry
// is discriminated (ADR 0083): every entry for that pair MUST carry a
// WhenColumn/WhenValue pair.
func findArms(table string, op Op) []Arm {
	var out []Arm
	for _, arm := range TripwireArms {
		if arm.Table == table && arm.Op == op {
			out = append(out, arm)
		}
	}
	if len(out) == 0 {
		panic(fmt.Sprintf("tripwire: no TripwireArms entry for (%s, %s)", table, op))
	}
	return out
}

// renderDiscriminatedCase renders a nested CASE NEW.<WhenColumn> ... END CASE
// block that assigns v_required_caps per discriminated arm (ADR 0083),
// followed by a fail-closed ELSE (RAISE P0001) mirroring the outer CASE's
// ELSE style for any subject value with no arm. Panics if the arms disagree
// on WhenColumn, or if any arm is undiscriminated (WhenColumn == "") — a
// programming error: this function is only called for (table,op) pairs with
// more than one TripwireArms entry, which are discriminated by construction.
func renderDiscriminatedCase(arms []Arm) string {
	whenColumn := arms[0].WhenColumn
	table := arms[0].Table
	if whenColumn == "" {
		panic(fmt.Sprintf("tripwire: renderDiscriminatedCase called for (%s, %s) with an undiscriminated arm (empty WhenColumn)", table, arms[0].Op))
	}
	var b strings.Builder
	b.WriteString("      CASE NEW." + whenColumn + "\n")
	for _, arm := range arms {
		if arm.WhenColumn != whenColumn {
			panic(fmt.Sprintf("tripwire: discriminated arms for (%s, %s) disagree on WhenColumn (%q vs %q)", arm.Table, arm.Op, whenColumn, arm.WhenColumn))
		}
		if arm.WhenValue == "" {
			panic(fmt.Sprintf("tripwire: discriminated arm for (%s, %s) has an empty WhenValue", arm.Table, arm.Op))
		}
		b.WriteString("        WHEN '" + arm.WhenValue + "' THEN\n")
		b.WriteString("          v_required_caps := " + renderArray(arm) + ";\n")
	}
	b.WriteString("        ELSE\n")
	b.WriteString("          RAISE EXCEPTION 'ErrCapabilityNotAsserted: no capability mapping for " + table + " " + whenColumn + " %, op %; enforce_capability_asserted has no discriminated arm for this subject', NEW." + whenColumn + ", TG_OP\n")
	b.WriteString("            USING ERRCODE = 'P0001';\n")
	b.WriteString("      END CASE;\n")
	return b.String()
}

// renderParentLookupDiscriminatedCase renders a parent-lookup SELECT
// followed by a nested CASE <looked-up value> ... END CASE block that assigns
// v_required_caps per parent-lookup-discriminated arm (ADR 0083 follow-on, M3
// P3.S2b-3b-iii-a), followed by a fail-closed ELSE (RAISE P0001) mirroring
// renderDiscriminatedCase's ELSE style. A SELECT that finds no parent row (or
// a parent row with a NULL/unrecognised discriminator value) leaves
// v_parent_subject_kind unmatched by every WHEN, which Postgres CASE routes
// to the ELSE — fail-closed by construction, no special-case NULL handling
// needed. Panics if the arms disagree on the parent-lookup shape, or if any
// arm is missing WhenParentTable/WhenParentKeyCol/WhenParentDiscrimCol/
// WhenValue — a programming error: this function is only called for
// (table,op) pairs with more than one TripwireArms entry using this
// discriminator form.
func renderParentLookupDiscriminatedCase(arms []Arm) string {
	table := arms[0].Table
	parentTable := arms[0].WhenParentTable
	parentKeyCol := arms[0].WhenParentKeyCol
	parentDiscrimCol := arms[0].WhenParentDiscrimCol
	if parentTable == "" || parentKeyCol == "" || parentDiscrimCol == "" {
		panic(fmt.Sprintf("tripwire: renderParentLookupDiscriminatedCase called for (%s, %s) with an arm missing WhenParentTable/WhenParentKeyCol/WhenParentDiscrimCol", table, arms[0].Op))
	}
	var b strings.Builder
	b.WriteString("      SELECT " + parentDiscrimCol + " INTO v_parent_subject_kind FROM public." + parentTable + " WHERE id = NEW." + parentKeyCol + ";\n")
	b.WriteString("      CASE v_parent_subject_kind\n")
	for _, arm := range arms {
		if arm.Table != table {
			panic(fmt.Sprintf("tripwire: renderParentLookupDiscriminatedCase called with mixed tables (%q vs %q)", table, arm.Table))
		}
		if arm.WhenParentTable != parentTable || arm.WhenParentKeyCol != parentKeyCol || arm.WhenParentDiscrimCol != parentDiscrimCol {
			panic(fmt.Sprintf("tripwire: parent-lookup-discriminated arms for (%s, %s) disagree on parent-lookup shape", arm.Table, arm.Op))
		}
		if arm.WhenValue == "" {
			panic(fmt.Sprintf("tripwire: parent-lookup-discriminated arm for (%s, %s) has an empty WhenValue", arm.Table, arm.Op))
		}
		b.WriteString("        WHEN '" + arm.WhenValue + "' THEN\n")
		b.WriteString("          v_required_caps := " + renderArray(arm) + ";\n")
	}
	b.WriteString("        ELSE\n")
	b.WriteString("          RAISE EXCEPTION 'ErrCapabilityNotAsserted: no capability mapping for " + table + " parent " + parentDiscrimCol + " %, op %; enforce_capability_asserted has no discriminated arm for this parent subject (parent lookup via % = NEW.%)', v_parent_subject_kind, TG_OP, '" + parentTable + "', '" + parentKeyCol + "'\n")
	b.WriteString("            USING ERRCODE = 'P0001';\n")
	b.WriteString("      END CASE;\n")
	return b.String()
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
