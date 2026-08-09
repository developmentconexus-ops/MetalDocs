package analyzers

import (
	"go/ast"
	"go/token"
	"strings"
)

// mutatingRepos are repo method prefixes that indicate state mutation on approval tables.
var mutatingRepos = []string{
	"UpdateApproval",
	"UpdateDocument",
	"UpdateSignoff",
	"UpdateInstance",
	"CreateSignoff",
	"CreateInstance",
}

// OutboxPair reports methods in approval/application that mutate approval state
// without a paired events.Emit call (outbox pairing requirement).
func OutboxPair(files []string) []Finding {
	var out []Finding
	fset := token.NewFileSet()

	for _, path := range files {
		if !strings.Contains(strings.ReplaceAll(path, "\\", "/"), "approval/application") {
			continue
		}
		_, raw := parseFile(fset, path)
		if raw == nil {
			continue
		}
		f := raw.(*ast.File)

		for _, decl := range f.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Body == nil {
				continue
			}
			// Only check exported methods with receivers
			if fn.Recv == nil || fn.Name == nil || !fn.Name.IsExported() {
				continue
			}

			if finding, ok := outboxPairFinding(fset, path, fn); ok {
				out = append(out, finding)
			}
		}
	}
	return out
}

// outboxPairFinding checks a single exported approval/application method for
// a repo-mutation call unpaired with a governance events.Emit call, returning
// the Finding to emit if the pairing is missing and not explicitly allowed.
func outboxPairFinding(fset *token.FileSet, path string, fn *ast.FuncDecl) (Finding, bool) {
	hasMutation, hasEmit := scanOutboxCalls(fn.Body)
	if !hasMutation || hasEmit {
		return Finding{}, false
	}

	pos := fset.Position(fn.Pos())
	src := readSource(path)
	line := getLine(src, pos.Line)
	// Skip if explicitly allowed
	if strings.Contains(line, "cilint:allow-no-outbox") ||
		(fn.Doc != nil && funcDocContains(fn.Doc, "cilint:allow-no-outbox")) {
		return Finding{}, false
	}
	return Finding{
		Analyzer: "outboxpair",
		File:     path,
		Line:     pos.Line,
		Message:  "method " + fn.Name.Name + " mutates approval state without paired events.Emit call; add governance event or //cilint:allow-no-outbox with justification",
	}, true
}

// scanOutboxCalls scans a function body for repo-mutation calls (mutatingRepos
// prefixes) and a paired governance event emitter Emit call, accepting both
// the bare-identifier form (events.Emit / emitter.Emit / e.Emit) and the
// struct-field form used throughout the codebase (s.emitter.Emit), where
// sel.X is itself a selector ending in the emitter field.
func scanOutboxCalls(body *ast.BlockStmt) (hasMutation, hasEmit bool) {
	ast.Inspect(body, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		if isMutatingCallName(sel.Sel.Name) {
			hasMutation = true
		}
		if isEmitCall(sel) {
			hasEmit = true
		}
		return true
	})
	return hasMutation, hasEmit
}

// isMutatingCallName reports whether name carries one of the mutatingRepos
// prefixes (repo method names that mutate approval state).
func isMutatingCallName(name string) bool {
	for _, prefix := range mutatingRepos {
		if strings.HasPrefix(name, prefix) {
			return true
		}
	}
	return false
}

// isEmitCall reports whether sel is a governance events.Emit call, accepting
// both the bare-identifier form (events.Emit / emitter.Emit / e.Emit) and the
// struct-field form used throughout the codebase (s.emitter.Emit), where
// sel.X is itself a selector ending in the emitter field.
func isEmitCall(sel *ast.SelectorExpr) bool {
	if sel.Sel.Name != "Emit" {
		return false
	}
	switch base := sel.X.(type) {
	case *ast.Ident:
		return base.Name == "events" || base.Name == "emitter" || base.Name == "e"
	case *ast.SelectorExpr:
		return base.Sel.Name == "events" || base.Sel.Name == "emitter" || base.Sel.Name == "e"
	}
	return false
}

func funcDocContains(doc *ast.CommentGroup, substr string) bool {
	for _, c := range doc.List {
		if strings.Contains(c.Text, substr) {
			return true
		}
	}
	return false
}
