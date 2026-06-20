package analyzers

import (
	"go/ast"
	"go/token"
	"strings"
)

// noResponseMapAllow is the inline directive to suppress a finding on the
// offending writer-call line (e.g. a deliberately off-spec route).
const noResponseMapAllow = "//cilint:allow-responsemap"

// responseWriters are the 2xx body-writer function names a response literal can
// reach. Matched on the callee leaf name, so both the bare helper (writeJSON)
// and the exported/aliased forms (WriteJSON, writeFillInJSON) are covered.
// problem.Write is intentionally excluded — it takes a *Problem, never a map.
var responseWriters = map[string]bool{
	"writeJSON":       true,
	"writeFillInJSON": true,
	"WriteJSON":       true,
}

// noResponseMapExemptFiles are recorded, rationale-backed exemptions (H-D
// allowlist, api-contract.md §5b). health.go's liveness/readiness probes are
// infra endpoints (k8s/LB), not the typed FE resource API, and the readiness
// body is genuinely dynamic (variable dependency-check array) — declared-dynamic
// like the F8.2 metrics envelope. Recorded here, not silently passed.
var noResponseMapExemptFiles = []string{
	"internal/platform/observability/health.go",
}

// NoResponseMap flags a map[string]any RESPONSE LITERAL passed to a 2xx body
// writer on a registered public route — the H-D class (api-contract.md §5b).
// It is the laundering-resistant, path-widened successor to the historical
// `grep writeJSON.*map[string]any` gate, which was blind to (a) the WriteJSON /
// writeFillInJSON writer aliases and (b) built-then-written locals
// (`page := map[string]any{...}; writeJSON(w, 200, page)`).
//
// SCOPE — every package that registers a public route, not just delivery/http:
//
//	internal/modules/<m>/delivery/http/   (the original H-D scope)
//	internal/modules/documents/approval/http/
//	internal/modules/iam/presence/
//	internal/platform/observability/
//
// DETECTION — per function: a writer call whose argument is either a
// `map[string]any` composite literal directly, or an identifier bound earlier
// in the same function to such a literal. Non-response map[string]any (audit
// payloads, command FormData, security Evidence, the declared-dynamic metrics
// envelope) never reach a 2xx writer, so they are not flagged.
//
// EXEMPTIONS — recorded files (noResponseMapExemptFiles) and the inline
// //cilint:allow-responsemap directive on the writer-call line.
func NoResponseMap(files []string) []Finding {
	var out []Finding
	fset := token.NewFileSet()

	for _, path := range files {
		if !inRegisteredRoutePackage(path) || isResponseMapExempt(path) {
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
			// Pass 1: collect identifiers bound to a map[string]any literal in
			// this function body (built-then-written locals).
			boundMapIdents := map[string]bool{}
			ast.Inspect(fn.Body, func(m ast.Node) bool {
				as, ok := m.(*ast.AssignStmt)
				if !ok {
					return true
				}
				for i, rhs := range as.Rhs {
					if i >= len(as.Lhs) {
						break
					}
					if !isMapStringAnyLiteral(rhs) {
						continue
					}
					if id, ok := as.Lhs[i].(*ast.Ident); ok && id.Name != "_" {
						boundMapIdents[id.Name] = true
					}
				}
				return true
			})

			// Pass 2: flag writer calls whose args are a map literal or a bound ident.
			ast.Inspect(fn.Body, func(m ast.Node) bool {
				call, ok := m.(*ast.CallExpr)
				if !ok || !isResponseWriter(call.Fun) {
					return true
				}
				pos := fset.Position(call.Pos())
				if strings.Contains(getLine(src, pos.Line), noResponseMapAllow) {
					return true
				}
				for _, arg := range call.Args {
					flagged := isMapStringAnyLiteral(arg)
					if !flagged {
						if id, ok := arg.(*ast.Ident); ok && boundMapIdents[id.Name] {
							flagged = true
						}
					}
					if flagged {
						out = append(out, Finding{
							Analyzer: "noresponsemap",
							File:     path,
							Line:     pos.Line,
							Message:  "map[string]any response literal passed to a 2xx body writer on a registered route (H-D, api-contract.md §5b): every 200/201 body must be a typed struct; convert it or record an explicit exemption",
						})
						return true
					}
				}
				return true
			})
			return true
		})
	}
	return out
}

// inRegisteredRoutePackage reports whether the file lives in a package that
// registers a public HTTP route (the full H-D surface).
func inRegisteredRoutePackage(path string) bool {
	s := strings.ReplaceAll(path, "\\", "/")
	switch {
	case strings.Contains(s, "internal/modules/") && strings.Contains(s, "/delivery/http/"):
		return true
	case strings.Contains(s, "internal/modules/documents/approval/http/"):
		return true
	case strings.Contains(s, "internal/modules/iam/presence/"):
		return true
	case strings.Contains(s, "internal/platform/observability/"):
		return true
	}
	return false
}

func isResponseMapExempt(path string) bool {
	s := strings.ReplaceAll(path, "\\", "/")
	for _, ex := range noResponseMapExemptFiles {
		if strings.HasSuffix(s, ex) {
			return true
		}
	}
	return false
}

// isResponseWriter reports whether a call expression's callee is one of the
// known 2xx body writers, matched on the leaf name (Ident or selector .Sel).
func isResponseWriter(fn ast.Expr) bool {
	switch v := fn.(type) {
	case *ast.Ident:
		return responseWriters[v.Name]
	case *ast.SelectorExpr:
		return responseWriters[v.Sel.Name]
	}
	return false
}

// isMapStringAnyLiteral reports whether e is a composite literal of type
// map[string]any or map[string]interface{}.
func isMapStringAnyLiteral(e ast.Expr) bool {
	cl, ok := e.(*ast.CompositeLit)
	if !ok {
		return false
	}
	mt, ok := cl.Type.(*ast.MapType)
	if !ok {
		return false
	}
	key, ok := mt.Key.(*ast.Ident)
	if !ok || key.Name != "string" {
		return false
	}
	switch val := mt.Value.(type) {
	case *ast.Ident:
		return val.Name == "any"
	case *ast.InterfaceType:
		return val.Methods == nil || len(val.Methods.List) == 0
	}
	return false
}
