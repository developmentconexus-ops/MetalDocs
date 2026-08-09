package analyzers

import (
	"go/ast"
	"go/token"
	"strings"
)

// postCommitAuditSinks is the set of non-Tx method names that write audit or
// governance events. A call to any of these AFTER a Commit() in the same
// function body is a violation (REQ-ASYNC-1, F-07, D-01, D-2): a crash between
// the commit and the write silently drops the audit record.
//
// Names are selector method names (the part after the dot), not full qualified
// names.  Tx-safe variants (RecordTx, WriteTx, AppendAuditTx, LogTx) are
// intentionally absent — they are the correct form.
var postCommitAuditSinks = map[string]bool{
	"Log":         true, // domain.GovernanceLogger.Log
	"Write":       true, // documents Audit.Write
	"Record":      true, // audit.Writer.Record
	"AppendAudit": true, // templates repo.AppendAudit
}

// postCommitAuditAllowComment is the inline directive to suppress a finding on
// a specific call site.  Place it on the same line as the call.
const postCommitAuditAllowComment = "//cilint:allow-post-commit-audit"

// PostCommitAudit flags non-Tx audit/governance write calls that occur
// lexically after a tx.Commit() call inside the same function body in
// internal/modules/** files.
//
// Design: simple linear scan of CallExpr nodes inside each FuncDecl body.
// We collect the first line position of any Commit() call, then flag any
// audit-sink call whose position is greater.  This gives zero false positives
// for the in-Tx pattern (audit before Commit) and flags every post-commit
// plain write reliably.
func PostCommitAudit(files []string) []Finding {
	var out []Finding
	fset := token.NewFileSet()

	for _, path := range files {
		slashed := strings.ReplaceAll(path, "\\", "/")
		if !strings.Contains(slashed, "internal/modules/") {
			continue
		}
		_, raw := parseFile(fset, path)
		if raw == nil {
			continue
		}
		f := raw.(*ast.File)
		src := readSource(path)

		ast.Inspect(f, func(n ast.Node) bool {
			fn, ok := n.(*ast.FuncDecl)
			if !ok || fn.Body == nil {
				return true
			}
			checkFuncBody(fset, fn.Body, path, src, &out)
			return true
		})
	}
	return out
}

// deliveryAuditFields names struct fields / identifiers that hold an audit or
// governance event writer. Pairing a sink method (Record/Write/Log/AppendAudit)
// with one of these receiver names is what distinguishes a real audit write
// (h.audit.Record(...)) from the pervasive non-audit .Write/.Log calls in the
// delivery layer: problem.Write(w, p), w.Write(b), mac.Write(b), slog.Log(...).
var deliveryAuditFields = map[string]bool{
	"audit":       true,
	"auditWriter": true,
	"auditor":     true,
	"govLogger":   true,
}

// DeliveryAuditSink flags any non-Tx audit/governance write performed directly
// in the delivery layer (internal/modules/<m>/delivery/**). Audit for a business
// mutation must be emitted in the application/service layer inside the same tx
// that performs the mutation (H-3b, REQ-ASYNC-1, F-07): a delivery-layer
// audit.Record runs outside any business tx, so a crash between the committed
// mutation and the write silently drops the record. The existing intra-function
// PostCommitAudit rule cannot see this case because the delivery handler has no
// Commit() of its own — the commit happened in a different function (the service).
//
// Detection is name-based and single-file (no call graph): a CallExpr whose
// method name is in postCommitAuditSinks AND whose receiver's trailing identifier
// is an audit-writer field name (deliveryAuditFields). Tx-safe variants
// (RecordTx, WriteTx, ...) never match because the method-name set holds only the
// non-Tx names.
//
// Legitimately best-effort sinks — authentication events that have no business tx
// to attach to (e.g. auth.login.failed), and bulk-action envelopes whose per-item
// mutations are already atomically audited at the application layer — are
// suppressed with //cilint:allow-post-commit-audit on the call line.
func DeliveryAuditSink(files []string) []Finding {
	var out []Finding
	fset := token.NewFileSet()

	for _, path := range files {
		slashed := strings.ReplaceAll(path, "\\", "/")
		if !strings.Contains(slashed, "internal/modules/") || !strings.Contains(slashed, "/delivery/") {
			continue
		}
		_, raw := parseFile(fset, path)
		if raw == nil {
			continue
		}
		f := raw.(*ast.File)
		src := readSource(path)

		ast.Inspect(f, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			if finding, flagged := deliveryAuditFinding(fset, path, src, call); flagged {
				out = append(out, finding)
			}
			return true
		})
	}
	return out
}

// deliveryAuditFinding checks a single call expression for the delivery-layer
// non-Tx audit sink pattern (see DeliveryAuditSink doc comment), returning
// the Finding to emit if it matches and is not explicitly allowed.
func deliveryAuditFinding(fset *token.FileSet, path, src string, call *ast.CallExpr) (Finding, bool) {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return Finding{}, false
	}
	if !postCommitAuditSinks[sel.Sel.Name] {
		return Finding{}, false
	}
	if !isAuditReceiver(sel.X) {
		return Finding{}, false
	}
	pos := fset.Position(call.Pos())
	line := getLine(src, pos.Line)
	if strings.Contains(line, postCommitAuditAllowComment) {
		return Finding{}, false
	}
	return Finding{
		Analyzer: "postcommitaudit",
		File:     path,
		Line:     pos.Line,
		Message:  "non-Tx audit/governance write (" + sel.Sel.Name + ") in the delivery layer; emit audit in the application/service layer inside the business tx (use the *Tx variant), or suppress with //cilint:allow-post-commit-audit for a legitimately best-effort event (H-3b, REQ-ASYNC-1, F-07)",
	}, true
}

// isAuditReceiver reports whether the receiver expression of a sink call resolves
// to an audit-writer field/identifier: bare `audit.Record(...)` (Ident) or
// `h.audit.Record(...)` (SelectorExpr whose trailing field is audit-named).
func isAuditReceiver(recv ast.Expr) bool {
	switch r := recv.(type) {
	case *ast.Ident:
		return deliveryAuditFields[r.Name]
	case *ast.SelectorExpr:
		return deliveryAuditFields[r.Sel.Name]
	}
	return false
}

// checkFuncBody scans a single function body for the post-commit-audit pattern.
func checkFuncBody(fset *token.FileSet, body *ast.BlockStmt, path, src string, out *[]Finding) {
	// Pass 1: find the minimum line of any Commit() call in this function.
	commitLine := -1
	ast.Inspect(body, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		if sel.Sel.Name != "Commit" {
			return true
		}
		pos := fset.Position(call.Pos())
		if commitLine < 0 || pos.Line < commitLine {
			commitLine = pos.Line
		}
		return true
	})
	if commitLine < 0 {
		return // no Commit() in this function
	}

	// Pass 2: flag audit sink calls after the first Commit().
	ast.Inspect(body, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		if !postCommitAuditSinks[sel.Sel.Name] {
			return true
		}
		pos := fset.Position(call.Pos())
		if pos.Line <= commitLine {
			return true // before or at commit — fine
		}
		// Allow-directive on the same line suppresses the finding.
		line := getLine(src, pos.Line)
		if strings.Contains(line, postCommitAuditAllowComment) {
			return true
		}
		*out = append(*out, Finding{
			Analyzer: "postcommitaudit",
			File:     path,
			Line:     pos.Line,
			Message:  "non-Tx audit/governance write (" + sel.Sel.Name + ") called after Commit() in the same function; use the *Tx variant before Commit to keep the audit record atomic (REQ-ASYNC-1, F-07)",
		})
		return true
	})
}
