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
	{"documents/repository/repository.go", "controlled_documents"},                   // B1
	{"documents/repository/repository.go", "document_process_areas"},                 // B7
	{"controlleddocuments/infrastructure/repository.go", "document_revisions"},       // B2
	{"controlleddocuments/infrastructure/repository.go", "documents"},                // B3
	{"controlleddocuments/infrastructure/repository.go", "approval_instances"},       // B4
	{"documents/application/document_area.go", "controlled_documents"},               // B5
	{"documents/approval/application/read_service.go", "controlled_documents"},       // B6
	{"iam/infrastructure/postgres/area_catalog_reader.go", "document_process_areas"}, // B8
	{"documents/application/fillin_service.go", "templates_template_version"},        // N1 (M0/HS-6 → M2)
	// Category C — authz-visibility membership reads → published view (M3)
	{"controlleddocuments/infrastructure/repository.go", "user_process_areas"},              // C1+C2
	{"documents/approval/repository/postgres_approval_repository.go", "user_process_areas"}, // C3
	// Category C4 — search visibility-contract redesign (M4)
	{"search/infrastructure/v2documents/reader.go", "documents"},                       // C4a
	{"search/infrastructure/v2documents/reader.go", "controlled_documents"},            // C4b
	{"search/infrastructure/v2documents/reader.go", "controlled_document_area_grants"}, // C4c
	{"search/infrastructure/v2documents/reader.go", "controlled_document_user_grants"}, // C4e
	{"search/infrastructure/v2documents/reader.go", "user_process_areas"},              // C4d
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
