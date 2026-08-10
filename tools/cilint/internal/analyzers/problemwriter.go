package analyzers

import (
	"go/ast"
	"go/token"
	"strings"
)

// problemWriterAllow is the inline directive to suppress a finding on the
// offending line.
const problemWriterAllow = "//cilint:allow-problem-writer"

// problemMediaType is the RFC 9457 media type. Outside the platform problem
// package, naming it means the file is hand-rolling an error envelope.
const problemMediaType = "application/problem+json"

// canonicalWriterPkg is the ONLY package allowed to serialize a Problem.
const canonicalWriterPkg = "internal/platform/problem/"

// ProblemWriter flags any attempt to build a second HTTP error-serialization
// path (G-07, rulebook R-ERR-1). MetalDocs has exactly one: problem.Respond /
// problem.RespondCause in internal/platform/problem.
//
// Two shapes are refused, because a local writer must take one of them:
//
//  1. A function that accepts BOTH an http.ResponseWriter and a *problem.Problem.
//     That is precisely the signature of the twelve `writeProblem(w, p)` clones
//     this rule exists to keep from coming back. A legitimate error-TRANSLATION
//     helper maps error -> *problem.Problem and does not touch the writer; a
//     legitimate handler takes the writer and builds its Problem inline. Only a
//     serializer needs both at once.
//
//  2. The literal "application/problem+json". Emission itself is already
//     unreachable — problem.Write is unexported — so the remaining way to clone
//     the envelope is to hand-write the media type onto a ResponseWriter.
//
// Together these make the clone unrepresentable rather than merely discouraged:
// a would-be writer can neither reach the serializer nor rebuild it.
//
// Suppress a genuinely exceptional line with //cilint:allow-problem-writer and
// say why.
//
// *.gen.go is skipped, and the reason is structural rather than convenient:
// oapi-codegen re-emits those files verbatim from api/openapi/v1/openapi.yaml on
// every regen, so a finding there is unfixable by hand and would reappear at the
// next `go generate`. The enforcement point for generated emission is the
// generator configuration, not this lint. That is not a free pass: the
// strict-server `VisitXxxResponse` methods really are a second emission path
// wherever a module implements StrictServerInterface (today: distribution,
// notifications). That is recorded as A3-adjacent debt on the generated-code
// axis, not silently ignored here.
func ProblemWriter(files []string) []Finding {
	var out []Finding
	fset := token.NewFileSet()

	for _, path := range files {
		slashed := strings.ReplaceAll(path, "\\", "/")
		if !inRuntimeTree(slashed) ||
			strings.Contains(slashed, canonicalWriterPkg) ||
			strings.HasSuffix(slashed, ".gen.go") {
			continue
		}
		_, raw := parseFile(fset, path)
		if raw == nil {
			continue
		}
		f := raw.(*ast.File)
		src := readSource(path)

		ast.Inspect(f, func(n ast.Node) bool {
			switch node := n.(type) {
			case *ast.FuncDecl:
				if node.Type != nil && isProblemWriterSignature(node.Type.Params) {
					out = appendProblemFinding(out, fset, src, path, node.Pos(),
						"function "+node.Name.Name+" takes both an http.ResponseWriter and a *problem.Problem, "+
							"which is the local problem-writer shape (G-07, R-ERR-1); call problem.Respond / "+
							"problem.RespondCause instead — serialization has exactly one owner")
				}
			case *ast.FuncLit:
				if node.Type != nil && isProblemWriterSignature(node.Type.Params) {
					out = appendProblemFinding(out, fset, src, path, node.Pos(),
						"function literal takes both an http.ResponseWriter and a *problem.Problem, "+
							"which is the local problem-writer shape (G-07, R-ERR-1); call problem.Respond / "+
							"problem.RespondCause instead")
				}
			case *ast.BasicLit:
				if node.Kind == token.STRING && strings.Contains(node.Value, problemMediaType) {
					out = appendProblemFinding(out, fset, src, path, node.Pos(),
						"the "+problemMediaType+" media type is written by internal/platform/problem and "+
							"nowhere else (G-07, R-ERR-1); naming it here means a second error envelope")
				}
			}
			return true
		})
	}
	return out
}

// inRuntimeTree limits the rule to code that can actually serve an HTTP
// response: the modular monolith (internal/) and the binaries (apps/). Lint and
// codegen tooling under scripts/ and tools/ names the media type as the SUBJECT
// of an inspection — scripts/api-lint asserts the spec declares problem+json,
// and this analyzer holds it as a constant. Flagging those would be a category
// error: they describe the envelope, they never emit one.
func inRuntimeTree(slashed string) bool {
	for _, tooling := range []string{"tools/", "scripts/"} {
		if strings.HasPrefix(slashed, tooling) || strings.Contains(slashed, "/"+tooling) {
			return false
		}
	}
	return true
}

func appendProblemFinding(out []Finding, fset *token.FileSet, src, path string, pos token.Pos, msg string) []Finding {
	p := fset.Position(pos)
	if strings.Contains(getLine(src, p.Line), problemWriterAllow) {
		return out
	}
	return append(out, Finding{
		Analyzer: "problemwriter",
		File:     path,
		Line:     p.Line,
		Message:  msg,
	})
}

// isProblemWriterSignature reports whether a parameter list carries both an
// http.ResponseWriter and a *problem.Problem.
func isProblemWriterSignature(params *ast.FieldList) bool {
	if params == nil {
		return false
	}
	var hasWriter, hasProblem bool
	for _, field := range params.List {
		switch {
		case isSelector(field.Type, "http", "ResponseWriter"):
			hasWriter = true
		case isPointerToSelector(field.Type, "problem", "Problem"):
			hasProblem = true
		}
	}
	return hasWriter && hasProblem
}

func isSelector(expr ast.Expr, pkg, name string) bool {
	sel, ok := expr.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	ident, ok := sel.X.(*ast.Ident)
	return ok && ident.Name == pkg && sel.Sel.Name == name
}

func isPointerToSelector(expr ast.Expr, pkg, name string) bool {
	star, ok := expr.(*ast.StarExpr)
	if !ok {
		return false
	}
	return isSelector(star.X, pkg, name)
}
