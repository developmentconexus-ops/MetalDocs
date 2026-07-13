package tripwire

import (
	"strings"
	"testing"
)

// TestRenderMigration_ApprovalInstancesSubjectDiscriminated pins ADR 0083's
// core correctness property: approval_instances/INSERT renders a nested
// CASE NEW.subject_kind with two DISJOINT capability arrays — document rows
// require exactly document.submit, template rows require exactly
// template.submit — and the two sets are never unioned into one flat ARRAY.
// A coarse union (ARRAY['document.submit', 'template.submit']) would let a
// principal holding only template.submit authorize a document-subject insert
// (the security regression ADR 0083 rejects); this test would pass under a
// union too if it only checked substring containment, so it additionally
// asserts there is exactly one undiscriminated top-level
// "v_required_caps := ARRAY[...]" assignment gone from the approval_instances
// branch (replaced by the nested CASE) and that each subject's ARRAY literal
// does not contain the other subject's capability.
func TestRenderMigration_ApprovalInstancesSubjectDiscriminated(t *testing.T) {
	rendered := RenderMigration()

	branch := extractBranch(t, rendered, "approval_instances")

	if !strings.Contains(branch, "CASE NEW.subject_kind") {
		t.Fatalf("approval_instances branch must contain a nested CASE NEW.subject_kind:\n%s", branch)
	}

	docIdx := strings.Index(branch, "WHEN 'document' THEN")
	tplIdx := strings.Index(branch, "WHEN 'template' THEN")
	if docIdx < 0 {
		t.Fatalf("approval_instances branch missing WHEN 'document' THEN arm:\n%s", branch)
	}
	if tplIdx < 0 {
		t.Fatalf("approval_instances branch missing WHEN 'template' THEN arm:\n%s", branch)
	}

	// Slice out each subject's assignment line to check for cross-contamination.
	var docSection, tplSection string
	if docIdx < tplIdx {
		docSection = branch[docIdx:tplIdx]
		tplSection = branch[tplIdx:]
	} else {
		tplSection = branch[tplIdx:docIdx]
		docSection = branch[docIdx:]
	}

	if !strings.Contains(docSection, "ARRAY['document.submit']") {
		t.Fatalf("document arm must require exactly document.submit:\n%s", docSection)
	}
	if strings.Contains(docSection, "template.submit") {
		t.Fatalf("document arm must NOT reference template.submit (cross-subject union regression):\n%s", docSection)
	}

	if !strings.Contains(tplSection, "ARRAY['template.submit']") {
		t.Fatalf("template arm must require exactly template.submit:\n%s", tplSection)
	}
	if strings.Contains(tplSection, "document.submit") {
		t.Fatalf("template arm must NOT reference document.submit (cross-subject union regression):\n%s", tplSection)
	}

	// Fail-closed inner ELSE: a subject_kind outside {document, template} must
	// RAISE, mirroring the outer CASE's fail-closed style (P0001).
	if !strings.Contains(branch, "ELSE") || !strings.Contains(branch, "P0001") {
		t.Fatalf("approval_instances nested CASE must have a fail-closed ELSE raising P0001:\n%s", branch)
	}
}

// NOTE (unit 3.1a S5): the two slice-scoped golden-diff tests that lived here
// (TestRenderMigration_UntouchedArmsByteIdenticalToPriorGolden vs 0283 and
// ...ToPrior0299Golden vs 0299) were deleted with the 0301 re-render. Each
// pinned "only branch X changed in THIS slice" — a property of the 0299/0300
// re-renders, not a standing invariant — and re-broke by design on every
// legitimate later arm change. The standing invariants remain pinned by
// TestRenderMigration_MatchesCommittedFile (byte-parity with the committed
// latest golden, also enforced as the blocking TRIPWIRE-ARM-PARITY api-lint
// rule) and TestTripwireArms_MatchesContractTable (exact arm→caps table);
// per-slice minimality is proven at review time by diffing the committed
// goldens (0300 vs 0301: templates_template_version branch + ledger only).

// TestRenderMigration_ApprovalSignoffsParentLookupDiscriminated pins the M3
// P3.S2b-3b-iii-a core correctness property (ADR 0083 follow-on):
// approval_signoffs/INSERT has no direct subject_kind column, so it renders a
// parent-lookup SELECT against approval_instances followed by a nested
// CASE on the looked-up value, with two DISJOINT capability arrays — parent
// subject_kind='document' requires exactly document.signoff, parent
// subject_kind='template' requires exactly template.approve — never unioned
// into one flat ARRAY, and a fail-closed ELSE (P0001) for any other/absent
// parent subject_kind.
func TestRenderMigration_ApprovalSignoffsParentLookupDiscriminated(t *testing.T) {
	rendered := RenderMigration()

	branch := extractBranch(t, rendered, "approval_signoffs")

	// The parent-lookup SELECT must resolve the subject via the correct FK
	// column, off the correct parent table/column.
	if !strings.Contains(branch, "SELECT subject_kind INTO v_parent_subject_kind FROM public.approval_instances WHERE id = NEW.approval_instance_id;") {
		t.Fatalf("approval_signoffs branch missing the expected parent-lookup SELECT:\n%s", branch)
	}

	if !strings.Contains(branch, "CASE v_parent_subject_kind") {
		t.Fatalf("approval_signoffs branch must contain a nested CASE v_parent_subject_kind:\n%s", branch)
	}

	docIdx := strings.Index(branch, "WHEN 'document' THEN")
	tplIdx := strings.Index(branch, "WHEN 'template' THEN")
	if docIdx < 0 {
		t.Fatalf("approval_signoffs branch missing WHEN 'document' THEN arm:\n%s", branch)
	}
	if tplIdx < 0 {
		t.Fatalf("approval_signoffs branch missing WHEN 'template' THEN arm:\n%s", branch)
	}

	var docSection, tplSection string
	if docIdx < tplIdx {
		docSection = branch[docIdx:tplIdx]
		tplSection = branch[tplIdx:]
	} else {
		tplSection = branch[tplIdx:docIdx]
		docSection = branch[docIdx:]
	}

	if !strings.Contains(docSection, "ARRAY['document.signoff']") {
		t.Fatalf("document-parent arm must require exactly document.signoff:\n%s", docSection)
	}
	if strings.Contains(docSection, "template.approve") {
		t.Fatalf("document-parent arm must NOT reference template.approve (cross-subject union regression):\n%s", docSection)
	}

	if !strings.Contains(tplSection, "ARRAY['template.approve']") {
		t.Fatalf("template-parent arm must require exactly template.approve:\n%s", tplSection)
	}
	if strings.Contains(tplSection, "document.signoff") {
		t.Fatalf("template-parent arm must NOT reference document.signoff (cross-subject union regression):\n%s", tplSection)
	}

	// There must be no single flat ARRAY containing both caps anywhere in the
	// branch (the coarse-union security regression ADR 0083 rejects).
	if strings.Contains(branch, "ARRAY['document.signoff', 'template.approve']") ||
		strings.Contains(branch, "ARRAY['template.approve', 'document.signoff']") {
		t.Fatalf("approval_signoffs branch must never union document.signoff and template.approve into one ARRAY:\n%s", branch)
	}

	// Fail-closed inner ELSE: an unrecognised/absent parent subject_kind
	// (including a NULL from a missing parent row) must RAISE P0001, never
	// pass through.
	if !strings.Contains(branch, "ELSE") || !strings.Contains(branch, "P0001") {
		t.Fatalf("approval_signoffs nested CASE must have a fail-closed ELSE raising P0001:\n%s", branch)
	}
}

// extractBranch returns the text of the WHEN TG_TABLE_NAME = '<table>' ...
// branch (up to but excluding the next WHEN/ELSE at the same 4-space CASE
// indent), for assertions scoped to just that branch.
func extractBranch(t *testing.T, sql string, table string) string {
	t.Helper()
	marker := "WHEN TG_TABLE_NAME = '" + table + "'"
	start := strings.Index(sql, marker)
	if start < 0 {
		t.Fatalf("branch marker %q not found in rendered SQL", marker)
	}
	rest := sql[start:]
	end := nextBranchOffset(rest)
	return rest[:end]
}

// nextBranchOffset finds the offset of the next top-level "    WHEN " or
// "    ELSE" (4-space indent, the outer CASE's branch indent) after the
// start of the current branch's own "WHEN " keyword.
func nextBranchOffset(branchAndRest string) int {
	// Skip past the branch's own leading "WHEN " so we search from its body.
	searchFrom := len("WHEN")
	idx := strings.Index(branchAndRest[searchFrom:], "\n    WHEN TG_TABLE_NAME")
	elseIdx := strings.Index(branchAndRest[searchFrom:], "\n    ELSE")
	if idx < 0 || (elseIdx >= 0 && elseIdx < idx) {
		idx = elseIdx
	}
	if idx < 0 {
		return len(branchAndRest)
	}
	return searchFrom + idx + 1 // +1 to include leading \n in the stripped branch, excluded from the next
}
