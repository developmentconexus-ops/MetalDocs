package analyzers_test

import (
	"os"
	"path/filepath"
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
	// documents/approval reading the documents-context approval_instances is INTRA-module
	// (approval ⊂ documents top-level module), not cross-module. Must not flag.
	src := "package application\n" +
		"func q() string {\n" +
		"\treturn `SELECT id FROM approval_instances WHERE tenant_id = $1`\n" +
		"}\n"
	path := hgFixture(t, "internal/modules/documents/application/context_builder.go", src)
	findings := analyzers.HGCrossModule([]string{path})
	if len(findings) != 0 {
		t.Fatalf("intra-module (documents/approval ⊂ documents) read must not be flagged, got %d: %+v", len(findings), findings)
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

func TestHGCrossModule_Negative_PendingBaseline(t *testing.T) {
	// A known in-scope debt site on hgPendingRemediation is suppressed until its
	// milestone ports it. Uses a STILL-PENDING site — the C4d search baseline
	// (search/infrastructure/v2documents/reader.go reading iam's user_process_areas),
	// scheduled for M4. The prior C1+C2 site this test used
	// (controlleddocuments/infrastructure/repository.go reading user_process_areas) was
	// ported and drained in M3/F3.2, so the fixture was realigned to a live ledger entry.
	src := "package v2documents\n" +
		"func q() string {\n" +
		"\treturn `SELECT area_code FROM user_process_areas WHERE user_id = $1 AND tenant_id = $2`\n" +
		"}\n"
	path := hgFixture(t, "internal/modules/search/infrastructure/v2documents/reader.go", src)
	findings := analyzers.HGCrossModule([]string{path})
	if len(findings) != 0 {
		t.Fatalf("pending-remediation baseline site must be suppressed, got %d: %+v", len(findings), findings)
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
