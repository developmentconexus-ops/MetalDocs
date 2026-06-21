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

// hgOwnerByTable maps every owned base table to the TOP-LEVEL module that owns
// it (holds its writes), per the F0.2 binding census
// (docs/superpowers/milestones/backend-module-boundary-hardening/milestone-0-adr-and-census/f0.2-binding-census/census.md).
// "Top-level" means the first segment under internal/modules/: documents/approval
// ⊂ documents, iam/presence ⊂ iam — so an intra-context read across sub-packages
// is NOT cross-module. This is the data ADR-0039 D1 (base table = violation)
// classifies against.
var hgOwnerByTable = map[string]string{
	// controlleddocuments
	"controlled_documents":            "controlleddocuments",
	"controlled_document_area_grants": "controlleddocuments",
	"controlled_document_user_grants": "controlleddocuments",
	"cd_sequence_counters":            "controlleddocuments",
	// documents (incl. the approval sub-context)
	"documents":                   "documents",
	"document_revisions":          "documents",
	"document_comments":           "documents",
	"document_checkpoints":        "documents",
	"document_exports":            "documents",
	"document_placeholder_values": "documents",
	"editor_sessions":             "documents",
	"autosave_pending_uploads":    "documents",
	"auth_failure_counters":       "documents", // owned by documents/approval signature limiter (census false-positive note)
	"approval_instances":          "documents",
	"approval_routes":             "documents",
	"approval_route_stages":       "documents",
	"approval_stage_instances":    "documents",
	"approval_signoffs":           "documents",
	"governance_events":           "documents",
	// taxonomy
	"document_process_areas": "taxonomy",
	"document_profiles":      "taxonomy",
	"document_families":      "taxonomy",
	// iam
	"iam_users":          "iam",
	"iam_user_roles":     "iam",
	"user_process_areas": "iam",
	// auth
	"auth_identities": "auth",
	"auth_sessions":   "auth",
	// audit (cross-cutting platform append-sink — read projections exempt via D3d)
	"audit_events":      "audit",
	"audit_export_jobs": "audit",
	// templates
	"templates_template":         "templates",
	"templates_template_version": "templates",
	"templates_approval_config":  "templates",
	// jobs
	"idempotency_keys": "jobs",
	"job_leases":       "jobs",
}

// hgSite is a (file-suffix, table) allowlist key. The suffix is matched against
// the slash-normalized path; the table must match the read. File+table (not
// line) keeps entries stable under line drift.
type hgSite struct {
	fileSuffix string
	table      string
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
var hgExempt = []hgSite{
	// D3(d) — platform append-sink (audit_events): cross-cutting telemetry sink every
	// module writes via AppendAudit[Tx]; read projections are a distinct class.
	{"security/infrastructure/postgres/repository.go", "audit_events"},          // X3
	{"iam/infrastructure/postgres/observability_repository.go", "audit_events"}, // X4
	{"templates/repository/postgres.go", "audit_events"},                        // X7
	// D3(e) — parent grade-a-completion M4 dispositioned auth reads (ADR 0029/0031):
	// auth_identities has no tenant_id, scoped via = ANY(ids); re-porting re-litigates 0031.
	{"security/infrastructure/postgres/repository.go", "auth_identities"},          // X1
	{"security/infrastructure/postgres/repository.go", "auth_sessions"},            // X2
	{"iam/infrastructure/postgres/observability_repository.go", "auth_identities"}, // X5
	{"iam/presence/repository.go", "auth_identities"},                              // X6
	// D3(f) — worker-layer (jobs): infrastructure operating on the approval domain;
	// jobs-boundary rule deferred to a future pass.
	{"jobs/stuck_instance_watchdog/job.go", "approval_instances"},       // X8
	{"jobs/stuck_instance_watchdog/job.go", "approval_stage_instances"}, // X8
}

// hgFromJoin matches a table token following FROM or JOIN (covering plain reads,
// JOINs, subqueries, and EXISTS — all of which contain a FROM/JOIN), with an
// optional public./metaldocs. schema prefix. Case-insensitive; \s spans newlines
// so `FROM\n  table` is caught.
var hgFromJoin = regexp.MustCompile(`(?i)\b(?:from|join)\s+(?:public\.|metaldocs\.)?([a-z_][a-z0-9_]*)`)

// HGCrossModule flags a raw cross-module base-table read (ADR-0039 D1): a FROM/JOIN
// against another top-level module's owned base table, in a non-owner package,
// outside the recorded allowlists. It is the H-G sibling of the H-D noResponseMap
// guard and mechanizes ADR-0039 D6.
//
// SCOPE — files under internal/modules/<module>/** (non-test; _test.go excluded by
// the collector). The reader module is the first path segment under internal/modules/.
//
// DETECTION — only *ast.BasicLit STRING nodes are scanned (SQL lives in string
// literals), so a foreign table named in a comment or Go identifier never flags —
// the census found real such comments (people_service.go:690,
// observability_repository.go:164). For each FROM/JOIN <table> whose owner ≠ the
// reader module, a finding is emitted unless the (file,table) pair is on
// hgPendingRemediation (the M1–M4 debt ledger) or hgExempt (permanent D3(d)–(f)),
// or the source line carries //cilint:allow-hgcrossmodule.
//
// RESIDUAL — dynamically-assembled or aliased table names behind Go variables are
// invisible to the literal-token scan (recorded in the F0.2 census coverage
// statement, same residual as the H-D guard).
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
		seen := map[string]bool{} // dedupe per (line,table) within a file

		ast.Inspect(f, func(n ast.Node) bool {
			lit, ok := n.(*ast.BasicLit)
			if !ok || lit.Kind != token.STRING {
				return true
			}
			val := lit.Value
			for _, m := range hgFromJoin.FindAllStringSubmatchIndex(val, -1) {
				table := strings.ToLower(val[m[2]:m[3]])
				owner, owned := hgOwnerByTable[table]
				if !owned || owner == reader {
					continue // unknown table or own-table read (D3c) — compliant
				}
				line := fset.Position(lit.Pos()).Line + strings.Count(val[:m[0]], "\n")
				if hgListed(path, table, hgPendingRemediation) || hgListed(path, table, hgExempt) {
					continue
				}
				if strings.Contains(getLine(src, line), hgCrossModuleAllow) {
					continue
				}
				key := fmt.Sprintf("%d|%s", line, table)
				if seen[key] {
					continue
				}
				seen[key] = true
				out = append(out, Finding{
					Analyzer: "hgcrossmodule",
					File:     path,
					Line:     line,
					Message: fmt.Sprintf(
						"module %q reads %q's base table %q with raw SQL (H-G, ADR-0039 D1): port to the owner's published view/read-port (D3a/b), or record an explicit exemption",
						reader, owner, table),
				})
			}
			return true
		})
	}
	return out
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

func hgListed(path, table string, list []hgSite) bool {
	s := strings.ReplaceAll(path, "\\", "/")
	for _, site := range list {
		if site.table == table && strings.HasSuffix(s, site.fileSuffix) {
			return true
		}
	}
	return false
}
