package main

// TRIPWIRE-ARM-PARITY and TRIPWIRE-ARM-DRIFT — M2 F2.1 Stage 2a
// (docs/superpowers/milestones/global-maximum-remediation/
// milestone-2-authz-enforcement-generation/validation-contract.md §1.5 —
// binding). Both are blocking; there is no reported-only tier (main.go).
//
// TRIPWIRE-ARM-PARITY (a) binds internal/platform/tripwire.TripwireArms to the
// registry (every cap real) and to the generated migration (RenderMigration()
// byte-equal to the latest committed tripwire migration, db/migrations/0279_*.sql
// as of M7 F7.3) — a Go map / registry / generated-SQL three-way drift catcher.
//
// TRIPWIRE-ARM-DRIFT (b) is the function-local generalization of
// checkTripwirePairing (code_rules.go): pairing checks *presence* of
// authz.Require + mutating SQL in the same function; this rule checks *arm
// membership* — the asserted cap(s) must include at least one cap in the
// gated table's TripwireArms entry for the SQL statement's op (match-one,
// mirroring the DB trigger). Scope is function-local by design (contract
// §1.5.b coverage note); cross-layer (assert-in-service / write-in-repository)
// drift is out of scope here and covered by the integration drives (§1.6).

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	iamdomain "metaldocs/internal/modules/iam/domain"
	"metaldocs/internal/platform/tripwire"
)

// tripwireMigrationPath is the committed golden file TRIPWIRE-ARM-PARITY
// compares tripwire.RenderMigration() against. Advances with the latest
// rendered tripwire migration: M2 pinned 0271; M6 F6.2 re-rendered it to 0275
// (documents/UPDATE gains document.review); M7 F7.2 re-rendered it to 0277
// (tenants/INSERT arm gated on tenant.onboard + attachment, ADR 0070); M7 F7.3
// re-rendered it to 0279 (tenant_lifecycle_jobs/INSERT arm gated on
// tenant.export OR tenant.erase + attachment, ADR 0070), then to 0283 (DELETE
// paths RETURN OLD — RETURN NEW on BEFORE DELETE silently cancelled DELETEs),
// then to 0299 (ADR 0083, M3 P3.S2b-3b-0: approval_instances/INSERT arm
// splits into two subject-discriminated entries — document rows require
// document.submit, template rows require template.submit, via a nested CASE
// NEW.subject_kind, never unioned; numbered 0299 to avoid the 0284_ci_rls_role
// prefix collision — migrate.Apply dedupes by 4-digit prefix only), then to
// 0300 (ADR 0083 follow-on, M3 P3.S2b-3b-iii-a: approval_signoffs/INSERT arm
// splits into two parent-lookup-discriminated entries — approval_signoffs has
// no direct subject_kind column, so the trigger SELECTs the parent
// approval_instances row's subject_kind and CASEs on the looked-up value;
// document-parent rows require document.signoff, template-parent rows
// require template.approve, never unioned), then to 0301 (ADR 0082 phase c,
// unit 3.1a S5: templates_template_version arm drops template.review — the
// legacy reviewer stage was deleted in S4 and the capability is retired from
// the IAM registry in the same change-set).
const tripwireMigrationPath = "db/migrations/0301_tripwire_template_review_retired.sql"

// gatedTableSet returns the set of table names present in TripwireArms — the
// tables TRIPWIRE-ARM-DRIFT restricts its attention to (contract §1.5.b: "only
// flag writes to GATED tables").
func gatedTableSet(arms []tripwire.Arm) map[string]struct{} {
	out := make(map[string]struct{}, len(arms))
	for _, a := range arms {
		out[a.Table] = struct{}{}
	}
	return out
}

// armFor returns the TripwireArms entry (if any) matching table+op, where
// OpAny arms match any op (mirrors the trigger's CASE branches: some branches
// test TG_OP, some test table name only).
func armFor(arms []tripwire.Arm, table string, op tripwire.Op) (tripwire.Arm, bool) {
	for _, a := range arms {
		if a.Table != table {
			continue
		}
		if a.Op == tripwire.OpAny || a.Op == op {
			return a, true
		}
	}
	return tripwire.Arm{}, false
}

// unionArmCapsFor returns the UNION of Caps across ALL arms matching table+op,
// and whether any arm matched. A subject-discriminated table (ADR 0083:
// approval_instances/INSERT splits into a document arm requiring document.submit
// and a template arm requiring template.submit) has more than one arm for the
// same (table, op); the DB trigger picks the precise arm per-row via
// CASE NEW.subject_kind, but this STATIC function-local lint cannot know
// subject_kind at compile time. Unioning the arm caps for the static membership
// check is sound: it lets a function asserting EITHER subject's cap pass, while
// a function asserting NEITHER still fails. Runtime security is unchanged — the
// trigger's per-subject CASE remains the exact enforcer.
func unionArmCapsFor(arms []tripwire.Arm, table string, op tripwire.Op) ([]iamdomain.Capability, bool) {
	var union []iamdomain.Capability
	seen := map[iamdomain.Capability]struct{}{}
	matched := false
	for _, a := range arms {
		if a.Table != table {
			continue
		}
		if a.Op != tripwire.OpAny && a.Op != op {
			continue
		}
		matched = true
		for _, c := range a.Caps {
			if _, dup := seen[c]; dup {
				continue
			}
			seen[c] = struct{}{}
			union = append(union, c)
		}
	}
	return union, matched
}

// ---------------------------------------------------------------------------
// TRIPWIRE-ARM-PARITY
// ---------------------------------------------------------------------------

// checkTripwireArmParity implements contract §1.5.a: every capability
// referenced by tripwire.TripwireArms is a real registry capability, and
// tripwire.RenderMigration() byte-equals the committed 0271 migration under
// modulesRoot. modulesRoot must be the repo root (RenderMigration is a pure
// function of the imported TripwireArms — always the real, compiled-in arms —
// so this rule cannot be pointed at a synthetic arm set; the NEGATIVE fixtures
// for this rule construct a local, non-registry capability check + a
// tree-with-a-mutated-migration-file check instead, see
// tripwire_arm_rules_test.go).
//
// A missing 0271 file follows the same requireCoreFile convention as every
// other core-registry helper in this package (parseCapabilityConsts,
// checkSeedRegistryParity, ...): under strict (production/CI root) it is a
// hard error, not a silent pass; under non-strict (unit-test/table-driven
// fixture roots that never plant a migrations/ tree) it is skipped so
// pre-existing fixtures are unaffected.
func checkTripwireArmParity(modulesRoot string, strict bool) ([]Violation, error) {
	out := []Violation{}

	// (i) every cap in every arm must be a real registry capability.
	for _, arm := range tripwire.TripwireArms {
		for _, cap := range arm.Caps {
			if !iamdomain.IsValidCapability(cap) {
				out = append(out, Violation{
					File:    "internal/platform/tripwire/arms.go",
					Line:    0,
					Rule:    "TRIPWIRE-ARM-PARITY",
					Message: fmt.Sprintf("TripwireArms entry (%s, %s) references capability %q which is not a registered iam capability (IsValidCapability); every arm cap must originate in internal/modules/iam/domain/model.go", arm.Table, arm.Op, cap),
				})
			}
		}
	}

	// (ii) the generated migration must byte-equal the committed 0271 file.
	migPath := filepath.Join(modulesRoot, filepath.FromSlash(tripwireMigrationPath))
	committed, err := os.ReadFile(migPath)
	if err != nil {
		if os.IsNotExist(err) {
			if cerr := requireCoreFile(strict, "db/migrations/0271 (tripwire golden migration)", migPath); cerr != nil {
				return nil, cerr
			}
			return out, nil
		}
		return nil, err
	}
	rendered := tripwire.RenderMigration()
	if rendered != string(committed) {
		out = append(out, Violation{
			File:    migPath,
			Line:    0,
			Rule:    "TRIPWIRE-ARM-PARITY",
			Message: fmt.Sprintf("%s does not byte-equal internal/platform/tripwire.RenderMigration(); the committed SQL has drifted from the Go TripwireArms source of truth (hand-edit or stale regeneration) — regenerate and recommit", tripwireMigrationPath),
		})
	}

	return out, nil
}

// ---------------------------------------------------------------------------
// TRIPWIRE-ARM-DRIFT
// ---------------------------------------------------------------------------

// mutatingSQLWrite is one Exec/ExecContext call site's parsed (table, op).
type mutatingSQLWrite struct {
	table string
	op    tripwire.Op
	line  int
}

var (
	sqlUpdateRE     = regexp.MustCompile(`(?i)UPDATE\s+([a-zA-Z_][a-zA-Z0-9_.]*)`)
	sqlInsertIntoRE = regexp.MustCompile(`(?i)INSERT\s+INTO\s+([a-zA-Z_][a-zA-Z0-9_.]*)`)
	sqlDeleteFromRE = regexp.MustCompile(`(?i)DELETE\s+FROM\s+([a-zA-Z_][a-zA-Z0-9_.]*)`)
)

// parseMutatingTable extracts (table, op) from a SQL statement string, per
// contract §1.5.b: "table name parsed from the SQL string literal after
// INTO / UPDATE / DELETE FROM". Schema-qualified names (public.documents) are
// reduced to the bare table name to match TripwireArms.Table. Returns ok=false
// if the statement is not a recognized single mutating form.
func parseMutatingTable(sql string) (table string, op tripwire.Op, ok bool) {
	if m := sqlUpdateRE.FindStringSubmatch(sql); m != nil {
		return bareTableName(m[1]), tripwire.OpUpdate, true
	}
	if m := sqlInsertIntoRE.FindStringSubmatch(sql); m != nil {
		return bareTableName(m[1]), tripwire.OpInsert, true
	}
	if m := sqlDeleteFromRE.FindStringSubmatch(sql); m != nil {
		// DELETE has no TripwireArms Op const beyond Insert/Update/Any in the
		// contract's 18 entries; a DELETE against a gated table only matches an
		// OpAny arm (armFor handles that), never an Insert/Update-only arm.
		return bareTableName(m[1]), tripwire.Op("DELETE"), true
	}
	return "", "", false
}

func bareTableName(name string) string {
	if i := strings.LastIndex(name, "."); i >= 0 {
		return name[i+1:]
	}
	return name
}

// scanFuncAuthzAndWrites walks one function body and collects (a) every
// capability resolved from an authz.Require(ctx, tx, <cap>, <area>) call site
// (constName->value via constToValue, unwrapping string(...) — the
// capConstName technique from registry_rules.go) and (b) every mutating SQL
// write (table, op) from an Exec/ExecContext call, regardless of target table
// (filtering to gated tables happens in the caller).
func scanFuncAuthzAndWrites(fn *ast.FuncDecl, fset *token.FileSet, constToValue map[string]iamdomain.Capability) (caps []iamdomain.Capability, writes []mutatingSQLWrite) {
	seenCap := map[iamdomain.Capability]struct{}{}
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}

		// authz.Require(ctx, tx, <cap>, <area>)
		if pkg, ok := sel.X.(*ast.Ident); ok && pkg.Name == "authz" && sel.Sel.Name == "Require" {
			if len(call.Args) >= 3 {
				if constName, ok := capConstName(call.Args[2]); ok {
					if capVal, known := constToValue[constName]; known {
						if _, dup := seenCap[capVal]; !dup {
							seenCap[capVal] = struct{}{}
							caps = append(caps, capVal)
						}
					}
				}
			}
			return true
		}

		// db.Exec(sql, args...) / db.ExecContext(ctx, sql, args...)
		var sqlIdx int
		switch sel.Sel.Name {
		case "Exec":
			sqlIdx = 0
		case "ExecContext":
			sqlIdx = 1
		default:
			return true
		}
		if len(call.Args) <= sqlIdx {
			return true
		}
		raw, ok := call.Args[sqlIdx].(*ast.BasicLit)
		if !ok || raw.Kind != token.STRING {
			return true
		}
		sqlText, err := strconv.Unquote(raw.Value)
		if err != nil {
			sqlText = raw.Value
		}
		if table, op, ok := parseMutatingTable(sqlText); ok {
			writes = append(writes, mutatingSQLWrite{table: table, op: op, line: fset.Position(call.Pos()).Line})
		}
		return true
	})
	return caps, writes
}

// checkTripwireArmDrift implements contract §1.5.b: for every function that
// both asserts an authz.Require capability and writes a gated table (a table
// present in tripwire.TripwireArms), require that at least one asserted cap
// in that function is a member of the write's arm (match-one, mirroring the
// DB trigger's match-one semantics). A gated write whose function asserts no
// arm-member cap is a violation naming file:line, table, op, and the
// (non-empty) set of asserted caps that failed to satisfy the arm.
func checkTripwireArmDrift(modulesRoot string, fset *token.FileSet, strict bool) ([]Violation, error) {
	constToValue, err := parseCapabilityConsts(modulesRoot, fset, strict)
	if err != nil {
		return nil, err
	}
	if len(constToValue) == 0 {
		// No registry under this root (e.g. isolated lint fixture roots that
		// don't carry iam/domain/model.go); nothing to resolve against.
		return nil, nil
	}

	gated := gatedTableSet(tripwire.TripwireArms)

	out := []Violation{}
	walkErr := filepath.WalkDir(modulesRoot, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			switch d.Name() {
			case ".git", ".claude", "node_modules", "vendor":
				return filepath.SkipDir
			}
			return nil
		}
		lower := strings.ToLower(path)
		if !strings.HasSuffix(lower, ".go") || strings.HasSuffix(lower, "_test.go") || strings.HasSuffix(lower, ".gen.go") {
			return nil
		}
		file, err := parser.ParseFile(fset, path, nil, parser.SkipObjectResolution)
		if err != nil {
			return err
		}
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Body == nil {
				continue
			}
			caps, writes := scanFuncAuthzAndWrites(fn, fset, constToValue)
			if len(caps) == 0 || len(writes) == 0 {
				continue
			}
			for _, w := range writes {
				if _, isGated := gated[w.table]; !isGated {
					continue
				}
				armCaps, found := unionArmCapsFor(tripwire.TripwireArms, w.table, w.op)
				if !found {
					// Gated table but no arm matches this exact op (e.g. a DELETE
					// against a table only gated for INSERT/UPDATE by the contract's
					// entries) — not this rule's concern; the DB trigger's own
					// CASE branches are the runtime source of truth for op coverage.
					continue
				}
				if !anyCapInArm(caps, armCaps) {
					out = append(out, Violation{
						File: path,
						Line: w.line,
						Rule: "TRIPWIRE-ARM-DRIFT",
						Message: fmt.Sprintf(
							"function %s asserts capabilities %s and writes gated table %s (op %s), but none of the asserted caps are in TripwireArms[%s,%s]=%s; the DB tripwire will reject this write with P0001 for every actor — widen the arm in internal/platform/tripwire/arms.go or assert an arm-member capability",
							fn.Name.Name, capListString(caps), w.table, w.op, w.table, w.op, capListString(armCaps),
						),
					})
				}
			}
		}
		return nil
	})
	if walkErr != nil {
		return nil, walkErr
	}
	return out, nil
}

func anyCapInArm(caps []iamdomain.Capability, arm []iamdomain.Capability) bool {
	armSet := make(map[iamdomain.Capability]struct{}, len(arm))
	for _, c := range arm {
		armSet[c] = struct{}{}
	}
	for _, c := range caps {
		if _, ok := armSet[c]; ok {
			return true
		}
	}
	return false
}

func capListString(caps []iamdomain.Capability) string {
	s := make([]string, len(caps))
	for i, c := range caps {
		s[i] = string(c)
	}
	sort.Strings(s)
	return "{" + strings.Join(s, ", ") + "}"
}
