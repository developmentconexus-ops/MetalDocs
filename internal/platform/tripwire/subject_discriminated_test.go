package tripwire

import (
	"os"
	"path/filepath"
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

// TestRenderMigration_UntouchedArmsByteIdenticalToPriorGolden pins the
// critical ADR 0083 invariant: every non-approval_instances CASE branch must
// render BYTE-IDENTICAL to the prior golden migration (0283). We prove this
// by stripping the approval_instances branch out of both the prior golden
// file and the freshly rendered migration, then diffing the remainder.
func TestRenderMigration_UntouchedArmsByteIdenticalToPriorGolden(t *testing.T) {
	repoRoot, err := findRepoRoot()
	if err != nil {
		t.Fatalf("locate repo root: %v", err)
	}
	priorPath := filepath.Join(repoRoot, "db", "migrations", "0283_tripwire_delete_return_old.sql")
	prior, err := os.ReadFile(priorPath)
	if err != nil {
		t.Fatalf("read prior golden %s: %v", priorPath, err)
	}

	rendered := RenderMigration()

	priorMinusApproval := stripBranch(t, string(prior), "approval_instances")
	renderedMinusApproval := stripBranch(t, rendered, "approval_instances")

	// The function header/DECLARE block differs only by the migration-header
	// comment and schema_migrations version/description literal, which we
	// normalize out before comparing the CASE-branch bodies.
	priorBody := extractCaseBody(t, priorMinusApproval)
	renderedBody := extractCaseBody(t, renderedMinusApproval)

	if priorBody != renderedBody {
		t.Errorf("non-approval_instances CASE branches changed vs prior golden 0283 (ADR 0083 requires byte-identity for untouched arms).\n--- prior ---\n%s\n--- rendered ---\n%s", priorBody, renderedBody)
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

// stripBranch removes the WHEN TG_TABLE_NAME = '<table>' ... branch (and its
// body) from sql entirely, returning the remainder.
func stripBranch(t *testing.T, sql string, table string) string {
	t.Helper()
	marker := "WHEN TG_TABLE_NAME = '" + table + "'"
	start := strings.Index(sql, marker)
	if start < 0 {
		t.Fatalf("branch marker %q not found in SQL", marker)
	}
	rest := sql[start:]
	end := nextBranchOffset(rest)
	return sql[:start] + sql[start+end:]
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

// extractCaseBody returns the substring between "BEGIN\n  CASE" and
// "END CASE;" so header-comment/version-literal differences outside the CASE
// don't pollute the byte-identity comparison.
func extractCaseBody(t *testing.T, sql string) string {
	t.Helper()
	start := strings.Index(sql, "BEGIN\n  CASE")
	if start < 0 {
		t.Fatalf("CASE body start marker not found")
	}
	end := strings.Index(sql, "END CASE;")
	if end < 0 {
		t.Fatalf("CASE body end marker not found")
	}
	return sql[start:end]
}
