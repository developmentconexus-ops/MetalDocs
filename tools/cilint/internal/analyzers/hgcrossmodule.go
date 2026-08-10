package analyzers

import (
	"fmt"
	"go/ast"
	"go/token"
	"regexp"
	"strings"
)

// hgCrossModuleAllow is the inline directive that suppresses an H-G finding on
// the offending SQL line (a deliberate, rationale-commented exception). Sibling
// of //cilint:allow-responsemap.
const hgCrossModuleAllow = "//cilint:allow-hgcrossmodule"

// Ownership (hgOwnerByTable) is loaded from table-ownership.json — see
// table_ownership.go. It is deliberately NOT a map literal here: an
// architectural fact carried in analyzer source drifts silently from the ADRs
// that decide it.

// Access modes. An allowlist entry authorises exactly one of them.
const (
	hgModeReads  = "reads"
	hgModeWrites = "writes"
)

// hgSite is a (file-suffix, table, mode) allowlist key. The suffix is matched
// against the slash-normalized path; the table must match the access; the mode
// must match the SQL verb. File+table (not line) keeps entries stable under
// line drift.
//
// The mode is load-bearing (#87/A1 review B5). Entries used to be keyed by
// (file, table) only, so a file exempted for a foreign READ could start
// UPDATE/INSERT/DELETE-ing that same table with no new finding — the guard's
// central claim, that a foreign write is blocked from day one, was false for
// every exempted file. Reading another module's table breaks its read contract;
// writing it breaks its invariants, so the two are not one decision. A file
// that genuinely needs both takes two entries, each with its own rationale.
type hgSite struct {
	fileSuffix string
	table      string
	mode       string
	why        string
}

// hgPendingRemediation is the H-G DEBT LEDGER: the exact F0.2 in-scope sites
// (Categories A-skipped/B/C/C4 + N1) that are still raw cross-module reads today
// and are scheduled for porting in M1–M4. Each milestone DELETES its entries
// here when it ports the site (and its raw SQL). The guard is green now because
// every live in-scope read is listed; it fails the build on any NEW read. At the
// mission's terminal acceptance (mission.md §8) this slice MUST be empty.
var hgPendingRemediation = []hgSite{
	// Category B — foreign point-reads → owner read-ports (M2)
	// B1 ported (M2/F2.1): documents/repository reads profile_code via
	// controlleddocuments/domain.CDFieldReader (ADR-0039 D3(b)); raw read deleted.
	// B2/B3/B4 ported (M2/F2.2): GetActiveInstance reads the active/published
	// documents + in-progress approval_instances projection through the
	// documents-owned ActiveInstanceReader port (ADR-0039 D3(b)); the inline
	// documents/document_revisions/approval_instances reads were deleted.
	// B5 ported (M2/F2.1): LoadDocumentAreaCode resolves the CD area via
	// controlleddocuments/domain.CDFieldReader (tx-aware); LEFT JOIN deleted.
	// B6 ported (M2/F2.1): loadInstanceAreaCode resolves the CD area via
	// controlleddocuments/domain.CDFieldReader (tx-aware); LEFT JOIN deleted.
	// B7/B8 ported (M2/F2.3): documents (in-tx area name) and iam (off-tx area
	// existence) read metaldocs.document_process_areas through the taxonomy-owned
	// AreaCatalogReader port (ADR-0039 D3(b)); both raw reads were deleted.
	// N1 ported (M2/F2.4): LoadFillInSchema resolves template_version_id on the
	// documents-owned table, then reads templates_template_version.placeholder_schema
	// through the templates-owned TemplateVersionPort (ADR-0030 extended; ADR-0039
	// D3(b)); the cross-module JOIN was deleted.
	// Category C — authz-visibility membership reads → published view (M3)
	// C1+C2 ported (M3/F3.2): controlleddocuments List + CanRead read iam's published
	// metaldocs.v_active_user_areas view (ADR-0039 D3a); the raw user_process_areas
	// reads were deleted.
	// C3 ported (M3/F3.3): documents/approval ResolveEligibleActors reads the same
	// published view in-tx (H-PRE-1 preserved); the interval-predicate read of
	// metaldocs.user_process_areas was deleted.
	// C4a/C4b/C4c/C4d/C4e ported (M4/F4.3): search/v2documents/reader.go reads the
	// published contracts metaldocs.v_document_search_facts (documents),
	// metaldocs.v_cd_search_facts + metaldocs.v_cd_grantee (controlleddocuments,
	// the latter joined to iam's v_active_user_areas) instead of the five base
	// tables public.documents / controlled_documents / controlled_document_area_grants
	// / controlled_document_user_grants / user_process_areas; all five raw reads were
	// deleted (ADR-0039 D3a). This was the last in-scope debt site: the ledger is now
	// EMPTY, satisfying mission.md §8 terminal acceptance.
}

// hgExempt is the PERMANENT allowlist: cross-module reads that are compliant by a
// principled ADR-0039 D3(d)–(f) exemption (M0/F0.2 HS-6 operator ruling). Unlike
// hgPendingRemediation, these are NOT scheduled for porting — they are recorded,
// justified carve-outs. A new read here must be added with its rationale.
// Every entry below is mode hgModeReads. That is the whole set as ruled: each
// one was justified as a read projection or a dispositioned read. None of them
// authorises a write, and none ever should have.
var hgExempt = []hgSite{
	// D3(d) — platform append-sink (audit_events): cross-cutting telemetry sink every
	// module writes via AppendAudit[Tx]; read projections are a distinct class.
	{"security/infrastructure/postgres/repository.go", "audit_events", hgModeReads, "X3 — D3(d) read projection over the append-sink"},
	{"iam/infrastructure/postgres/observability_repository.go", "audit_events", hgModeReads, "X4 — D3(d) read projection over the append-sink"},
	// X7 path reconciled 2026-08-07: F9.5 renamed templates/repository/ →
	// templates/infrastructure/. scripts/check-test-discipline.sh:59 reconciled
	// the same rename on 2026-07-06; this list did not, so the exemption had
	// silently stopped matching and the read fell into the baseline instead.
	{"templates/infrastructure/postgres.go", "audit_events", hgModeReads, "X7 — D3(d) read projection over the append-sink"},
	// D3(e) — parent grade-a-completion M4 dispositioned auth reads (ADR 0029/0031):
	// auth_identities has no tenant_id, scoped via = ANY(ids); re-porting re-litigates 0031.
	{"security/infrastructure/postgres/repository.go", "auth_identities", hgModeReads, "X1 — D3(e) dispositioned auth read (ADR 0029/0031)"},
	{"security/infrastructure/postgres/repository.go", "auth_sessions", hgModeReads, "X2 — D3(e) dispositioned auth read (ADR 0029/0031)"},
	{"iam/infrastructure/postgres/observability_repository.go", "auth_identities", hgModeReads, "X5 — D3(e) dispositioned auth read (ADR 0029/0031)"},
	{"iam/presence/repository.go", "auth_identities", hgModeReads, "X6 — D3(e) dispositioned auth read (ADR 0029/0031)"},
	// D3(f) — worker-layer (jobs): infrastructure operating on the approval domain;
	// jobs-boundary rule deferred to a future pass.
	{"jobs/stuck_instance_watchdog/job.go", "approval_instances", hgModeReads, "X8 — D3(f) worker-layer read; the watchdog is alert-only (ADR 0068), it does not mutate"},
	{"jobs/stuck_instance_watchdog/job.go", "approval_stage_instances", hgModeReads, "X8 — D3(f) worker-layer read; the watchdog is alert-only (ADR 0068), it does not mutate"},
}

// hgFromJoin matches a table token following FROM or JOIN (covering plain reads,
// JOINs, subqueries, and EXISTS — all of which contain a FROM/JOIN), with an
// optional public./metaldocs. schema prefix. Case-insensitive; \s spans newlines
// so `FROM\n  table` is caught.
var hgFromJoin = regexp.MustCompile(`(?i)\b(?:from|join)\s+(?:public\.|metaldocs\.)?([a-z_][a-z0-9_]*)`)

// hgWrite matches a table token in write position: UPDATE <t>, INSERT INTO <t>,
// DELETE FROM <t>. Same schema-prefix and case/newline handling as hgFromJoin.
//
// A4.0 (#87/A1 infrastructure, #93/A4 property owner). Until this existed the
// guard was write-blind: hgFromJoin matches "FROM" and so caught DELETE FROM by
// accident while UPDATE <t> SET and INSERT INTO <t> — the two shapes that
// mutate another module's state — were invisible. The audit
// (docs/superpowers/analysis/audit-2026-08-09/final-synthesis.md G-02) counted
// 12 foreign-table writes the guard could not see, against 55 reads it could.
// Reading another module's table breaks its read contract; writing it breaks
// its invariants, so the blind half was the dangerous half.
//
// DELETE FROM is deliberately claimed by this pattern rather than left to
// hgFromJoin: it is a write, and reporting it as a read is a wrong statement
// about what the code does. hgScanLiteral resolves the overlap by classifying
// write-position hits first.
var hgWrite = regexp.MustCompile(`(?i)\b(?:update|insert\s+into|delete\s+from)\s+(?:public\.|metaldocs\.)?([a-z_][a-z0-9_]*)`)

// HGCrossModule flags raw cross-module base-table ACCESS (ADR-0039 D1) in a
// non-owner package, outside the recorded allowlists:
//
//	read  — FROM / JOIN <table>
//	write — UPDATE <table> / INSERT INTO <table> / DELETE FROM <table>
//
// It is the H-G sibling of the H-D noResponseMap guard and mechanizes ADR-0039 D6.
//
// The write half is A4.0. The property is not "grep found a keyword" — it is
// that a module cannot emit SQL against a business table owned by another
// module unless that access is a deliberately classified exemption. A guard
// that saw only reads asserted half of that and, worse, asserted it about the
// harmless half: a foreign read breaks the owner's read contract, a foreign
// write breaks its invariants.
//
// SCOPE — files under internal/modules/<module>/** (non-test; _test.go excluded by
// the collector). The reader module is the first path segment under internal/modules/.
//
// DETECTION — only *ast.BasicLit STRING nodes are scanned (SQL lives in string
// literals), so a foreign table named in a comment or Go identifier never flags —
// the census found real such comments (people_service.go:690,
// observability_repository.go:164). For each read- or write-position <table>
// whose owner ≠ the reader module, a finding is emitted unless the
// (file,table,mode) triple is on hgPendingRemediation (the M1–M4 debt ledger) or
// hgExempt (permanent D3(d)–(f)), or the source line carries
// //cilint:allow-hgcrossmodule. Mode is part of the key: a read exemption does
// not authorise a write.
//
// RESIDUAL — dynamically-assembled or aliased table names behind Go variables are
// invisible to the literal-token scan (recorded in the F0.2 census coverage
// statement, same residual as the H-D guard). Also invisible: writes routed
// through a query builder rather than a SQL literal, and CTE-form writes whose
// target follows neither UPDATE/INSERT INTO/DELETE FROM nor FROM/JOIN.
func HGCrossModule(files []string) []Finding {
	var out []Finding
	fset := token.NewFileSet()

	for _, path := range files {
		reader := hgModuleOf(path)
		if reader == "" {
			continue // only internal/modules/** is in scope
		}
		_, raw := parseFile(fset, path)
		if raw == nil {
			continue
		}
		f := raw.(*ast.File)
		src := readSource(path)
		out = append(out, hgScanFile(fset, path, reader, f, src)...)
	}
	return out
}

// hgScanFile scans a single parsed file for cross-module base-table reads,
// returning one Finding per (line,table) violation found in its string
// literals.
func hgScanFile(fset *token.FileSet, path, reader string, f *ast.File, src string) []Finding {
	var out []Finding
	seen := map[string]bool{} // dedupe per (line,table) within a file

	ast.Inspect(f, func(n ast.Node) bool {
		lit, ok := n.(*ast.BasicLit)
		if !ok || lit.Kind != token.STRING {
			return true
		}
		out = append(out, hgScanLiteral(fset, path, reader, lit, src, seen)...)
		return true
	})
	return out
}

// hgScanLiteral scans a single string literal for FROM/JOIN references to
// another module's owned base table, returning one Finding per newly-seen
// (line,table) violation not covered by an allowlist or inline directive.
func hgScanLiteral(fset *token.FileSet, path, reader string, lit *ast.BasicLit, src string, seen map[string]bool) []Finding {
	var out []Finding
	val := lit.Value

	// Writes are classified first. DELETE FROM matches both patterns, and it
	// is a write; letting hgFromJoin claim it would keep calling a deletion a
	// read. hgWritten records the (line,table) pairs already accounted for so
	// the read pass below skips them.
	hgWritten := map[string]bool{}
	for _, m := range hgWrite.FindAllStringSubmatchIndex(val, -1) {
		table := strings.ToLower(val[m[2]:m[3]])
		line := fset.Position(lit.Pos()).Line + strings.Count(val[:m[0]], "\n")
		hgWritten[fmt.Sprintf("%d|%s", line, table)] = true
		f, ok := hgViolation(path, reader, table, line, src, "writes", seen)
		if ok {
			out = append(out, f)
		}
	}

	for _, m := range hgFromJoin.FindAllStringSubmatchIndex(val, -1) {
		table := strings.ToLower(val[m[2]:m[3]])
		line := fset.Position(lit.Pos()).Line + strings.Count(val[:m[0]], "\n")
		if hgWritten[fmt.Sprintf("%d|%s", line, table)] {
			continue
		}
		f, ok := hgViolation(path, reader, table, line, src, "reads", seen)
		if ok {
			out = append(out, f)
		}
	}
	return out
}

// hgViolation applies the ownership rule and the three suppression layers to
// one (table, line, verb) hit, returning the Finding to emit if any.
//
// The verb is part of the dedupe key and of the message. Both matter: a single
// statement can read and write the same foreign table ("UPDATE documents SET x
// = (SELECT ... FROM documents)"), and collapsing the two would report only
// whichever the scanner reached first.
func hgViolation(path, reader, table string, line int, src, verb string, seen map[string]bool) (Finding, bool) {
	owner, owned := hgOwnerByTable[table]
	if !owned || owner == reader {
		return Finding{}, false // unknown table, or own-table access (D3c) — compliant
	}
	// The allowlists are keyed by (file, table, mode). A read exemption
	// authorises reads and nothing else: a file cleared to project over
	// audit_events that starts UPDATE-ing it is a new decision, and the guard
	// makes someone take it.
	if hgListed(path, table, verb, hgPendingRemediation) || hgListed(path, table, verb, hgExempt) {
		return Finding{}, false
	}
	if strings.Contains(getLine(src, line), hgCrossModuleAllow) {
		return Finding{}, false
	}
	key := fmt.Sprintf("%d|%s|%s", line, table, verb)
	if seen[key] {
		return Finding{}, false
	}
	seen[key] = true

	remedy := "port to the owner's published view/read-port (D3a/b), or record an explicit exemption"
	if verb == "writes" {
		remedy = "route the mutation through the owner's application service, or record an explicit exemption"
	}
	return Finding{
		Analyzer: "hgcrossmodule",
		File:     path,
		Line:     line,
		Message: fmt.Sprintf(
			"module %q %s %q's base table %q with raw SQL (H-G, ADR-0039 D1): %s",
			reader, verb, owner, table, remedy),
	}, true
}

// hgModuleOf returns the top-level module segment under internal/modules/, or ""
// if the file is not under internal/modules/.
func hgModuleOf(path string) string {
	s := strings.ReplaceAll(path, "\\", "/")
	const marker = "internal/modules/"
	i := strings.Index(s, marker)
	if i < 0 {
		return ""
	}
	rest := s[i+len(marker):]
	seg := rest
	if j := strings.IndexByte(rest, '/'); j >= 0 {
		seg = rest[:j]
	}
	return seg
}

// hgListed reports whether (path, table, verb) is authorised by list. All three
// must match: an entry for a different mode does not authorise this access.
func hgListed(path, table, verb string, list []hgSite) bool {
	s := strings.ReplaceAll(path, "\\", "/")
	for _, site := range list {
		if site.table == table && site.mode == verb && strings.HasSuffix(s, site.fileSuffix) {
			return true
		}
	}
	return false
}
