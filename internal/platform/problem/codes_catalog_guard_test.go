package problem

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
)

// errorWriterCodeArg maps an error-writing function name to the zero-based index
// of its problem.Code argument. These are the only sanctioned ways the documents
// and templates HTTP packages emit a Problem.code, so the code position must
// never be a raw string literal (which would reintroduce an ad-hoc code).
var errorWriterCodeArg = map[string]int{
	"New":                1, // problem.New(status, code, title)
	"httpErr":            2, // httpErr(w, status, code)
	"httpErrDetail":      2, // httpErrDetail(w, status, code, detail)
	"writeErr":           2, // writeErr(w, status, code, message)
	"writePDFWebhookErr": 2, // writePDFWebhookErr(w, status, code, detail)
	"writeErrJSON":       2, // idempotency: writeErrJSON(w, status, code, msg)
	"WriteError":         2, // httpresponse.WriteError(w, status, code, message)
}

// guardExcludedFiles names files inside an otherwise-guarded directory that own a
// separate, documented dotted-code taxonomy and are exempt from the canonical-catalog
// check. fillin_handler.go defines the documents fill-in taxonomy (not_found.revision,
// validation.failed, internal.unknown, ...) shared by the view/reconstruct/
// placeholder-options handlers; H-1e relocated it into documents/delivery/http but it
// retains its own taxonomy.
var guardExcludedFiles = map[string]bool{
	"fillin_handler.go": true,
}

// guardedPackages are the HTTP packages whose error vocabulary Family 4
// standardized onto the canonical catalog (codes.go). The dotted-taxonomy
// packages (documents fill-in (see guardExcludedFiles), documents/approval/http) are intentionally
// excluded: they own a separate, documented code taxonomy.
var guardedPackages = []string{
	filepath.Join("internal", "modules", "documents", "delivery", "http"),
	filepath.Join("internal", "modules", "templates", "delivery", "http"),
	filepath.Join("internal", "platform", "idempotency"), // F-09: middleware raw codes closed Wave 1.4
	// H-4: identity/auth/audit/security/search delivery packages standardized onto
	// the canonical catalog. Every Problem.code in these packages is now a typed
	// problem.Code constant (the membership domain codes were promoted into the
	// catalog with their exact existing string values, so the wire contract and
	// the frontend error-code map are unchanged).
	filepath.Join("internal", "modules", "iam", "delivery", "http"),
	filepath.Join("internal", "modules", "auth", "delivery", "http"),
	filepath.Join("internal", "modules", "audit", "delivery", "http"),
	filepath.Join("internal", "modules", "security", "delivery", "http"),
	filepath.Join("internal", "modules", "search", "delivery", "http"),
	"pdf_webhook", // documents/delivery/http: only the pdf-complete webhook handler
	// F-09 Wave 2.4: only handler.go is on the canonical catalog. The sibling
	// routes.go (writeDomainError) owns a separate CD/template domain taxonomy
	// — including dotted codes (template.artifact_missing) — that is intentionally
	// excluded, mirroring the documents/approval/http exclusion above.
	"controlleddocuments_handler",
}

func repoRoot(t *testing.T) string {
	t.Helper()
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	// .../internal/platform/problem/codes_catalog_guard_test.go -> repo root
	return filepath.Clean(filepath.Join(filepath.Dir(thisFile), "..", "..", ".."))
}

// canonicalCodeNames parses codes.go and returns the set of exported Code
// constant names (e.g. "CodeNotFoundResource").
func canonicalCodeNames(t *testing.T, root string) map[string]bool {
	t.Helper()
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, filepath.Join(root, "internal", "platform", "problem", "codes.go"), nil, 0)
	if err != nil {
		t.Fatalf("parse codes.go: %v", err)
	}
	names := map[string]bool{}
	for _, decl := range f.Decls {
		gd, ok := decl.(*ast.GenDecl)
		// ADR 0089 step 1: the catalog is `var` now, not `const` — Code became a
		// closed struct type, which cannot be a Go constant. Accept both tokens so
		// this guard keeps working until ADR 0089's later step deletes it outright
		// (it is superseded by the registry, which enforces the same invariant at
		// init time instead of by AST scan).
		if !ok || (gd.Tok != token.CONST && gd.Tok != token.VAR) {
			continue
		}
		for _, spec := range gd.Specs {
			vs, ok := spec.(*ast.ValueSpec)
			if !ok {
				continue
			}
			for _, name := range vs.Names {
				if strings.HasPrefix(name.Name, "Code") || strings.HasPrefix(name.Name, "FieldCode") {
					names[name.Name] = true
				}
			}
		}
	}
	if len(names) == 0 {
		t.Fatal("no canonical Code constants found in codes.go")
	}
	return names
}

func resolvePkgFiles(t *testing.T, root, entry string) []string {
	t.Helper()
	if entry == "pdf_webhook" {
		return []string{filepath.Join(root, "internal", "modules", "documents", "delivery", "http", "pdf_webhook_handler.go")}
	}
	if entry == "controlleddocuments_handler" {
		return []string{filepath.Join(root, "internal", "modules", "controlleddocuments", "delivery", "http", "handler.go")}
	}
	dir := filepath.Join(root, entry)
	dirEntries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read dir %s: %v", dir, err)
	}
	var files []string
	for _, de := range dirEntries {
		name := de.Name()
		if de.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		if guardExcludedFiles[name] {
			continue
		}
		files = append(files, filepath.Join(dir, name))
	}
	return files
}

func reportAdHoc(t *testing.T, fset *token.FileSet, root, path string, lit *ast.BasicLit) {
	t.Helper()
	rel, _ := filepath.Rel(root, path)
	pos := fset.Position(lit.Pos())
	val, _ := strconv.Unquote(lit.Value)
	t.Errorf("ad-hoc string code %q at %s:%d — use a problem.Code constant from the canonical catalog", val, rel, pos.Line)
}

// isProblemCodeType reports whether a type expression is problem.Code.
func isProblemCodeType(expr ast.Expr) bool {
	sel, ok := expr.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	pkg, ok := sel.X.(*ast.Ident)
	return ok && pkg.Name == "problem" && sel.Sel.Name == "Code"
}

func codeArgIndex(call *ast.CallExpr) (int, bool) {
	switch fn := call.Fun.(type) {
	case *ast.Ident:
		// Package-local helpers: httpErr, httpErrDetail, writeErr, writePDFWebhookErr.
		if fn.Name == "New" {
			return 0, false // only problem.New (selector) carries a code arg
		}
		idx, ok := errorWriterCodeArg[fn.Name]
		return idx, ok
	case *ast.SelectorExpr:
		// problem.New(status, code, title) and httpresponse.WriteError(w, status,
		// code, message) carry a code arg; ignore other pkg.New (e.g. idempotency.New).
		if pkg, ok := fn.X.(*ast.Ident); ok && (pkg.Name == "problem" || pkg.Name == "httpresponse") {
			idx, ok := errorWriterCodeArg[fn.Sel.Name]
			return idx, ok
		}
	}
	return 0, false
}

// TestNoAdHocStringCodes fails if any sanctioned error-writer in a guarded
// package passes a raw string literal in the Problem.code position. This is the
// regression guard for Family 4: it prevents reintroducing ad-hoc codes such as
// "not_found" or "stale_base" that bypass the canonical catalog.
func TestNoAdHocStringCodes(t *testing.T) {
	root := repoRoot(t)
	fset := token.NewFileSet()

	for _, entry := range guardedPackages {
		for _, path := range resolvePkgFiles(t, root, entry) {
			f, err := parser.ParseFile(fset, path, nil, 0)
			if err != nil {
				t.Fatalf("parse %s: %v", path, err)
			}
			ast.Inspect(f, func(n ast.Node) bool {
				switch node := n.(type) {
				case *ast.CallExpr:
					idx, ok := codeArgIndex(node)
					if !ok || idx >= len(node.Args) {
						return true
					}
					if lit, ok := node.Args[idx].(*ast.BasicLit); ok && lit.Kind == token.STRING {
						reportAdHoc(t, fset, root, path, lit)
					}
				case *ast.ValueSpec:
					// Flag `const X problem.Code = "literal"` in guarded packages;
					// string-literal Code values belong only in the catalog.
					if !isProblemCodeType(node.Type) {
						return true
					}
					for _, v := range node.Values {
						if lit, ok := v.(*ast.BasicLit); ok && lit.Kind == token.STRING {
							reportAdHoc(t, fset, root, path, lit)
						}
					}
				}
				return true
			})
		}
	}
}

// TestCanonicalCodeSelectorsExist fails if a guarded package references a
// problem.CodeXxx selector that does not exist in codes.go (catches typos and
// stale codes after a catalog rename).
func TestCanonicalCodeSelectorsExist(t *testing.T) {
	root := repoRoot(t)
	catalog := canonicalCodeNames(t, root)
	fset := token.NewFileSet()

	for _, entry := range guardedPackages {
		for _, path := range resolvePkgFiles(t, root, entry) {
			f, err := parser.ParseFile(fset, path, nil, 0)
			if err != nil {
				t.Fatalf("parse %s: %v", path, err)
			}
			ast.Inspect(f, func(n ast.Node) bool {
				sel, ok := n.(*ast.SelectorExpr)
				if !ok {
					return true
				}
				pkg, ok := sel.X.(*ast.Ident)
				if !ok || pkg.Name != "problem" {
					return true
				}
				name := sel.Sel.Name
				// Skip the bare type names (problem.Code / problem.FieldCode used in
				// signatures); only constant references must resolve to the catalog.
				if name == "Code" || name == "FieldCode" {
					return true
				}
				if !strings.HasPrefix(name, "Code") && !strings.HasPrefix(name, "FieldCode") {
					return true
				}
				if !catalog[name] {
					rel, _ := filepath.Rel(root, path)
					pos := fset.Position(sel.Pos())
					t.Errorf("unknown canonical code problem.%s at %s:%d", name, rel, pos.Line)
				}
				return true
			})
		}
	}
}
