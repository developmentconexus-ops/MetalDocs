package analyzers_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"metaldocs/tools/cilint/internal/analyzers"
)

func hgFixture(t *testing.T, relPath, src string) string {
	t.Helper()
	dir := t.TempDir()
	full := filepath.Join(dir, filepath.FromSlash(relPath))
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(full, []byte(src), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	return full
}

func TestHGCrossModule_Positive_CrossModuleRead(t *testing.T) {
	// search reads documents' base table directly — H-G violation (ADR-0039 D1).
	src := "package search\n" +
		"func q() string {\n" +
		"\treturn `SELECT d.id FROM documents d WHERE d.tenant_id = $1`\n" +
		"}\n"
	path := hgFixture(t, "internal/modules/search/infrastructure/v2documents/probe.go", src)
	findings := analyzers.HGCrossModule([]string{path})
	if len(findings) == 0 {
		t.Fatal("expected finding for search reading documents (cross-module base-table read)")
	}
}

func TestHGCrossModule_Negative_OwnTable(t *testing.T) {
	// controlleddocuments reading its OWN controlled_documents is compliant (D3c).
	src := "package infrastructure\n" +
		"func q() string {\n" +
		"\treturn `SELECT code FROM controlled_documents WHERE tenant_id = $1`\n" +
		"}\n"
	path := hgFixture(t, "internal/modules/controlleddocuments/infrastructure/repo.go", src)
	findings := analyzers.HGCrossModule([]string{path})
	if len(findings) != 0 {
		t.Fatalf("own-table read must not be flagged, got %d: %+v", len(findings), findings)
	}
}

func TestHGCrossModule_Negative_SubpackageSameModule(t *testing.T) {
	// ADR 0082 promoted approval to its own top-level module (superseding ADR 0072's
	// nested documents/approval ruling), so approval_instances is now owned by
	// "approval", not "documents" (tools/cilint/internal/analyzers/hgcrossmodule.go
	// hgOwnerByTable). The scenario this test guards — a file in one SUBPACKAGE of a
	// module reading a base table owned by that same top-level module — is now
	// expressed inside approval itself: approval/application (a different subpackage
	// from wherever the table is canonically written) reading approval's own
	// approval_instances is INTRA-module, not cross-module. Must not flag.
	src := "package application\n" +
		"func q() string {\n" +
		"\treturn `SELECT id FROM approval_instances WHERE tenant_id = $1`\n" +
		"}\n"
	path := hgFixture(t, "internal/modules/approval/application/context_builder.go", src)
	findings := analyzers.HGCrossModule([]string{path})
	if len(findings) != 0 {
		t.Fatalf("intra-module (approval/application reading approval's own approval_instances) read must not be flagged, got %d: %+v", len(findings), findings)
	}
}

func TestHGCrossModule_Negative_CommentMention(t *testing.T) {
	// A foreign table named only in a COMMENT (not a SQL string literal) must not flag —
	// the census found real instances (people_service.go:690, observability_repository.go:164).
	src := "package application\n" +
		"// D1 defer: pushing filters into SQL requires IAM to join auth_identities later.\n" +
		"func q() string {\n" +
		"\treturn `SELECT user_id FROM iam_users WHERE tenant_id = $1`\n" +
		"}\n"
	path := hgFixture(t, "internal/modules/iam/application/people_service.go", src)
	findings := analyzers.HGCrossModule([]string{path})
	if len(findings) != 0 {
		t.Fatalf("foreign table named only in a comment must not be flagged, got %d: %+v", len(findings), findings)
	}
}

func TestHGCrossModule_LedgerDrained_EmptyAtMissionEnd(t *testing.T) {
	// End-state guard: M4/F4.3 drained the last in-scope debt entries (the C4
	// search sites), so hgPendingRemediation is now EMPTY (mission.md §8 terminal
	// acceptance). The formerly-pending C4d baseline — search/v2documents/reader.go
	// reading iam's user_process_areas — must therefore NO LONGER be suppressed: with
	// the ledger drained, a re-introduced raw read flags again. (The suppression
	// MECHANISM, hgListed, stays covered by TestHGCrossModule_Negative_Exempt.)
	src := "package v2documents\n" +
		"func q() string {\n" +
		"\treturn `SELECT area_code FROM user_process_areas WHERE user_id = $1 AND tenant_id = $2`\n" +
		"}\n"
	path := hgFixture(t, "internal/modules/search/infrastructure/v2documents/reader.go", src)
	findings := analyzers.HGCrossModule([]string{path})
	if len(findings) == 0 {
		t.Fatal("ledger is drained (empty at mission end): the formerly-pending C4d site must now flag, got 0 findings")
	}
}

func TestHGCrossModule_Negative_Exempt(t *testing.T) {
	// An X-site (security reading audit_events — platform append-sink, ADR-0039 D3d) is
	// permanently exempt via hgExempt.
	src := "package postgres\n" +
		"func q() string {\n" +
		"\treturn `SELECT id FROM metaldocs.audit_events e WHERE e.actor_id = $1`\n" +
		"}\n"
	path := hgFixture(t, "internal/modules/security/infrastructure/postgres/repository.go", src)
	findings := analyzers.HGCrossModule([]string{path})
	if len(findings) != 0 {
		t.Fatalf("exempt X-site (platform sink) must not be flagged, got %d: %+v", len(findings), findings)
	}
}

func TestHGCrossModule_Negative_AllowDirective(t *testing.T) {
	src := "package search\n" +
		"func q() string {\n" +
		"\treturn `SELECT d.id FROM documents d` //cilint:allow-hgcrossmodule deliberate probe\n" +
		"}\n"
	path := hgFixture(t, "internal/modules/search/infrastructure/v2documents/probe.go", src)
	findings := analyzers.HGCrossModule([]string{path})
	if len(findings) != 0 {
		t.Fatalf("inline allow directive must suppress finding, got %d: %+v", len(findings), findings)
	}
}

// #87/A1 review F1's negative fixture. document_images is a live base table with
// a governed doc (wiki/database/tables/document_images.md, "**Owner:** documents")
// that was NOT in table-ownership.json until the completeness pass. Before that,
// this exact write — approval mutating a documents-owned table — produced zero
// findings, because hgViolation reports nothing for a table it does not know.
// The bug was never in the rule; it was in the rule's input being a subset.
//
// The review named document_attachments as the example. That one is not usable
// as a fixture and the reason is itself the F1 property: it is documented but
// NOT in the live schema (archive/migrations/0014 created it, the 2026-07-29
// baseline fold did not carry it forward), so it is now on the catalog's
// `retired` list and TestGovernedDocsReconcileWithLiveAndRetired is what holds
// that classification honest. document_images is the same class of omission with
// a table that actually exists.
func TestHGCrossModule_Positive_ForeignWriteToPreviouslyOmittedTable(t *testing.T) {
	src := "package application\n" +
		"func q() string {\n" +
		"\treturn `UPDATE document_images SET checksum = $1 WHERE document_id = $2 AND tenant_id = $3`\n" +
		"}\n"
	path := hgFixture(t, "internal/modules/approval/application/attach_service.go", src)
	findings := analyzers.HGCrossModule([]string{path})
	if len(findings) == 0 {
		t.Fatal("approval writing documents' document_images must flag: an omitted table is an unguarded table")
	}
	if !strings.Contains(findings[0].Message, "writes") || !strings.Contains(findings[0].Message, "document_images") {
		t.Fatalf("finding must name the verb and the table, got: %s", findings[0].Message)
	}
}

// The same shape one class over: a table classified out of module enforcement
// carries no owner, so an access to it cannot be foreign. This is what keeps the
// non-module classes honest as classifications rather than as silence — the
// table IS in the catalog, and the completeness test can see it.
func TestHGCrossModule_Negative_PlatformClassTableHasNoOwner(t *testing.T) {
	src := "package infrastructure\n" +
		"func q() string {\n" +
		"\treturn `INSERT INTO outbox_events (aggregate_id, payload) VALUES ($1, $2)`\n" +
		"}\n"
	path := hgFixture(t, "internal/modules/documents/infrastructure/postgres/outbox.go", src)
	findings := analyzers.HGCrossModule([]string{path})
	if len(findings) != 0 {
		t.Fatalf("outbox_events is class platform (every module writes it by design), got %d: %+v", len(findings), findings)
	}
}

func TestHGCrossModule_OutOfScope_NonModuleFile(t *testing.T) {
	// A file outside internal/modules/** is out of scope (platform, tools, etc.).
	src := "package platform\n" +
		"func q() string {\n" +
		"\treturn `SELECT id FROM documents`\n" +
		"}\n"
	path := hgFixture(t, "internal/platform/db/probe.go", src)
	findings := analyzers.HGCrossModule([]string{path})
	if len(findings) != 0 {
		t.Fatalf("non-module file is out of scope, got %d: %+v", len(findings), findings)
	}
}
