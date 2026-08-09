package analyzers

import (
	"go/ast"
	"go/token"
	"regexp"
	"strings"
)

// noDualModeAllow is the inline directive to suppress a finding on the
// offending if-statement line.
const noDualModeAllow = "//cilint:allow-dualmode"

// dualModeIdentRe matches leaf identifier names that are db/port-shaped:
//   - literal: db, sqldb, tx
//   - suffix:  *db, *port, *dispatcher, *limiter, *writer, *logger
//
// Matching is case-insensitive so deps.SQLDB, s.db, govLogger, etc. all hit.
var dualModeIdentRe = regexp.MustCompile(`(?i)^(db|sqldb|tx|.*db|.*port|.*dispatcher|.*limiter|.*writer|.*logger)$`)

// NoDualMode flags nil-comparison if-statements that form a dual-mode branch
// in application-layer files under internal/modules/<m>/application/.
//
// A dual-mode branch is one where a nil check on a db/port-shaped identifier
// guards an ALTERNATE EXECUTION PATH rather than failing loudly. Two patterns
// are flagged:
//
//  1. if x == nil { ... } else { ... }  — explicit alternate path via else.
//  2. if x != nil { ... }               — conditional skip-a-write (no else),
//     meaning the block is silently bypassed when x is nil.
//
// EXEMPTION — required-dep / precondition guard (NOT flagged):
//
//	if x == nil { panic(...) }   ← no else, body terminates
//	if x == nil { return ... }   ← no else, body terminates
//
// These are the FE-4 "fail-loud" pattern that ELIMINATES the alternate path
// rather than providing one. The live example is auth/application/service.go:
//
//	if loginCtxPort == nil { panic("auth.NewService: loginCtxPort is required") }
//
// Detection: the if-body must have no else AND its last (only meaningful)
// statement must be a panic call or a return statement — both are terminating.
// _test.go files are excluded entirely.
func NoDualMode(files []string) []Finding {
	var out []Finding
	fset := token.NewFileSet()

	for _, path := range files {
		if !inApplicationLayer(path) {
			continue
		}
		// Exclude test files.
		if strings.HasSuffix(path, "_test.go") {
			continue
		}
		_, raw := parseFile(fset, path)
		if raw == nil {
			continue
		}
		f := raw.(*ast.File)
		src := readSource(path)

		ast.Inspect(f, func(n ast.Node) bool {
			ifStmt, ok := n.(*ast.IfStmt)
			if !ok {
				return true
			}
			if finding, flagged := dualModeFinding(fset, path, src, ifStmt); flagged {
				out = append(out, finding)
			}
			return true
		})
	}
	return out
}

// dualModeFinding evaluates a single if-statement against the dual-mode
// nil-check pattern (see NoDualMode doc comment for the two flag rules and
// the fail-loud exemption) and returns the Finding to emit, if any.
func dualModeFinding(fset *token.FileSet, path, src string, ifStmt *ast.IfStmt) (Finding, bool) {
	// The condition must be a binary expression: X == nil or X != nil.
	bin, ok := ifStmt.Cond.(*ast.BinaryExpr)
	if !ok {
		return Finding{}, false
	}
	isEqNil, isNeqNil := classifyNilCheck(bin)
	if !isEqNil && !isNeqNil {
		return Finding{}, false
	}

	// Determine the variable/selector being checked.
	subject := nilCheckSubject(bin)
	if subject == "" {
		return Finding{}, false
	}
	// Apply the db/port identifier heuristic to the leaf name.
	if !dualModeIdentRe.MatchString(leafName(subject)) {
		return Finding{}, false
	}

	// Check allow-directive on the if-statement line.
	pos := fset.Position(ifStmt.Pos())
	line := getLine(src, pos.Line)
	if strings.Contains(line, noDualModeAllow) {
		return Finding{}, false
	}

	// ── Exemption: fail-loud guard ────────────────────────────────────
	// Pattern: if x == nil { panic(...) or return ... } with NO else.
	// This is a required-dep precondition guard — not a dual-mode branch.
	if isEqNil && ifStmt.Else == nil && isTerminatingBody(ifStmt.Body) {
		return Finding{}, false
	}

	// ── Flag rule 1: explicit else (alternate path) ────────────────────
	if ifStmt.Else != nil {
		return Finding{
			Analyzer: "nodualmode",
			File:     path,
			Line:     pos.Line,
			Message:  "dual-mode branch on db/port identifier '" + subject + "': nil/non-nil if-else forms alternate execution paths; application services must be single-mode (Task 6, F-10)",
		}, true
	}

	// ── Flag rule 2: if x != nil { ... } — conditional skip ────────────
	// No else, but the != nil form means the block is silently skipped
	// when the dependency is absent — that is the conditional-skip-a-write
	// dual-mode anti-pattern.
	if isNeqNil {
		return Finding{
			Analyzer: "nodualmode",
			File:     path,
			Line:     pos.Line,
			Message:  "conditional skip on db/port identifier '" + subject + "': if x != nil { ... } with no else silently skips work when the dependency is nil; application services must be single-mode (Task 6, F-10)",
		}, true
	}

	return Finding{}, false
}

// classifyNilCheck reports whether bin is an X == nil or X != nil comparison,
// checking both operand orders (nil may appear on either side).
func classifyNilCheck(bin *ast.BinaryExpr) (isEqNil, isNeqNil bool) {
	isEqNil = bin.Op.String() == "==" && isNilIdent(bin.Y)
	isNeqNil = bin.Op.String() == "!=" && isNilIdent(bin.Y)
	if !isEqNil && !isNeqNil {
		isEqNil = bin.Op.String() == "==" && isNilIdent(bin.X)
		isNeqNil = bin.Op.String() == "!=" && isNilIdent(bin.X)
	}
	return isEqNil, isNeqNil
}

// inApplicationLayer reports whether the file path is under an
// internal/modules/<module>/application/ directory.
func inApplicationLayer(path string) bool {
	slashed := strings.ReplaceAll(path, "\\", "/")
	return strings.Contains(slashed, "internal/modules/") &&
		strings.Contains(slashed, "/application/")
}

// isNilIdent reports whether node is the bare identifier "nil".
func isNilIdent(n ast.Expr) bool {
	id, ok := n.(*ast.Ident)
	return ok && id.Name == "nil"
}

// nilCheckSubject returns a human-readable representation of the non-nil side
// of a nil comparison (e.g. "s.db", "loginCtxPort", "deps.SQLDB").
func nilCheckSubject(bin *ast.BinaryExpr) string {
	var subject ast.Expr
	if isNilIdent(bin.Y) {
		subject = bin.X
	} else {
		subject = bin.Y
	}
	return exprString(subject)
}

// exprString returns a concise string for an expression — handles plain
// identifiers and single-level selectors (x.y). Deeper expressions return "".
func exprString(e ast.Expr) string {
	switch v := e.(type) {
	case *ast.Ident:
		return v.Name
	case *ast.SelectorExpr:
		if base, ok := v.X.(*ast.Ident); ok {
			return base.Name + "." + v.Sel.Name
		}
		return v.Sel.Name
	}
	return ""
}

// leafName returns the rightmost identifier name from a dotted expression
// string such as "s.db" → "db", "loginCtxPort" → "loginCtxPort".
func leafName(s string) string {
	if idx := strings.LastIndex(s, "."); idx >= 0 {
		return s[idx+1:]
	}
	return s
}

// isTerminatingBody reports whether a block statement's effective last
// statement is a panic call or a return — making it a fail-loud guard body.
// We scan all statements and consider the block terminating if any statement
// is a panic call-expr or a return statement (covering single-statement bodies
// as well as blocks that may have a leading comment-triggered blank line).
func isTerminatingBody(body *ast.BlockStmt) bool {
	if body == nil || len(body.List) == 0 {
		return false
	}
	last := body.List[len(body.List)-1]
	return isTerminatingStmt(last)
}

// isTerminatingStmt reports whether a statement is a panic call or return.
func isTerminatingStmt(s ast.Stmt) bool {
	switch stmt := s.(type) {
	case *ast.ReturnStmt:
		return true
	case *ast.ExprStmt:
		call, ok := stmt.X.(*ast.CallExpr)
		if !ok {
			return false
		}
		fn, ok := call.Fun.(*ast.Ident)
		return ok && fn.Name == "panic"
	}
	return false
}
