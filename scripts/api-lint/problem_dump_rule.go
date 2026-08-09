package main

// PROBLEM-DUMP-IMPORT (ADR 0089 step 12).
//
// The Problem code catalog is a RUNTIME value: every code is created by a
// problem.Register call at package init, and cmd/problem-codes-dump asks
// problem.Registrations() for the whole set. That is strictly better than the
// regex scraper it replaced — except for one failure mode the scraper did not
// have. A registering package that nobody imports never runs its init, so its
// codes are simply absent from the registry, and every generated artifact is
// silently short: the frontend coverage test passes, the wiki table looks
// complete, and a whole module's errors reach users as untranslated raw codes.
//
// The failure is invisible by construction — there is nothing to see, only
// something missing — which is exactly the shape of defect a lint has to carry,
// because no test can assert about codes it never learned exist.
//
// So: AST-scan the tree for problem.Register* calls, collect the packages that
// make them, read cmd/problem-codes-dump's own import block, and fail when a
// registering package is not linked in. The dumper's import list stops being a
// thing anyone has to remember.

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// dumperRelPath is the command whose import block must cover every registering
// package. It is the ONLY consumer that needs the whole catalog linked in.
var dumperRelPath = filepath.Join("cmd", "problem-codes-dump", "main.go")

// modulePrefix is the Go module path, used to turn an import string back into a
// repo-relative directory.
const modulePrefix = "metaldocs/"

// registerFuncs are the constructors that put a code in the registry. A package
// calling any of them contributes to problem.Registrations() and must therefore
// be linked into the dumper.
//
// RegisterField is deliberately included: field codes live in a separate
// namespace and are not emitted to the FE snapshot today, but they are still
// registry state the dumper reports on, and a package that only registers field
// codes is exactly as invisible if it is not imported.
var registerFuncs = map[string]struct{}{
	"Register":           {},
	"RegisterWithStatus": {},
	"RegisterLegacy":     {},
	"RegisterField":      {},
}

// checkProblemDumpImports reports every package that registers Problem codes but
// is not imported by cmd/problem-codes-dump.
//
// root is the REPO root. The scan covers internal/ and apps/ — the two trees
// that hold HTTP delivery code. The platform package that DEFINES Register is
// skipped: it is the registry itself, always linked in transitively, and its own
// calls are the shared catalog.
func checkProblemDumpImports(root string, fset *token.FileSet, strict bool) ([]Violation, error) {
	dumper := filepath.Join(root, dumperRelPath)
	imported, err := dumperImports(dumper, fset)
	if err != nil {
		if os.IsNotExist(err) && !strict {
			// Non-strict runs tolerate a missing dumper (a partial checkout, a
			// wrong root). -strict — what CI uses — does not, so a wrong root can
			// never masquerade as a clean pass.
			return nil, nil
		}
		return nil, fmt.Errorf("PROBLEM-DUMP-IMPORT: read %s: %w", dumperRelPath, err)
	}

	registering, err := registeringPackages(root, fset)
	if err != nil {
		return nil, err
	}

	var missing []string
	for pkg := range registering {
		if _, ok := imported[pkg]; !ok {
			missing = append(missing, pkg)
		}
	}
	sort.Strings(missing)

	out := make([]Violation, 0, len(missing))
	for _, pkg := range missing {
		out = append(out, Violation{
			File: dumperRelPath,
			Line: 1,
			Rule: "PROBLEM-DUMP-IMPORT",
			Message: fmt.Sprintf(
				"package %q calls problem.Register* but is not imported by cmd/problem-codes-dump; "+
					"its codes never reach the registry, so the generated FE snapshot and wiki table "+
					"silently omit them. Add `_ %q` to the dumper's import block.",
				filepath.ToSlash(pkg), modulePrefix+filepath.ToSlash(pkg)),
		})
	}
	return out, nil
}

// dumperImports returns the repo-relative package directories the dumper
// imports, keyed for lookup. Only metaldocs/-prefixed imports are relevant.
func dumperImports(path string, fset *token.FileSet) (map[string]struct{}, error) {
	src, err := os.ReadFile(path) // #nosec G304 -- path is root (CLI-supplied repo root) joined with the fixed literal dumperRelPath constant; not external input.
	if err != nil {
		return nil, err
	}
	f, err := parser.ParseFile(fset, path, src, parser.ImportsOnly)
	if err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	out := map[string]struct{}{}
	for _, imp := range f.Imports {
		p, err := strconv.Unquote(imp.Path.Value)
		if err != nil || !strings.HasPrefix(p, modulePrefix) {
			continue
		}
		out[filepath.FromSlash(strings.TrimPrefix(p, modulePrefix))] = struct{}{}
	}
	return out, nil
}

// registeringPackages walks internal/ and apps/ and returns the repo-relative
// directory of every package containing a problem.Register* call.
//
// Test files are skipped: a code registered only from a _test.go file is not
// part of the shipped catalog, and the platform package's own compile-fail
// fixtures would otherwise be reported.
func registeringPackages(root string, fset *token.FileSet) (map[string]struct{}, error) {
	skip := filepath.Join("internal", "platform", "problem")
	out := map[string]struct{}{}

	for _, tree := range []string{"internal", "apps"} {
		base := filepath.Join(root, tree)
		if _, err := os.Stat(base); os.IsNotExist(err) {
			continue
		}
		err := filepath.Walk(base, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				if info.Name() == "testdata" || info.Name() == "vendor" {
					return filepath.SkipDir
				}
				return nil
			}
			if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
				return nil
			}
			rel, relErr := filepath.Rel(root, filepath.Dir(path))
			if relErr != nil {
				return relErr
			}
			if rel == skip {
				return nil
			}
			if _, done := out[rel]; done {
				return nil
			}
			registers, parseErr := fileRegistersCode(path, fset)
			if parseErr != nil {
				return parseErr
			}
			if registers {
				out[rel] = struct{}{}
			}
			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("PROBLEM-DUMP-IMPORT: walk %s: %w", tree, err)
		}
	}
	return out, nil
}

// fileRegistersCode reports whether the file contains a `problem.Register*`
// call. Matched on the selector shape rather than resolved types: api-lint is a
// single-file AST scan by design (no call graph, no type checker), and the
// qualifier is unambiguous — nothing else in this tree is named `problem`.
func fileRegistersCode(path string, fset *token.FileSet) (bool, error) {
	src, err := os.ReadFile(path) // #nosec G304 -- path is produced by filepath.Walk over root/internal and root/apps (CLI-supplied repo root); not external input.
	if err != nil {
		return false, err
	}
	f, err := parser.ParseFile(fset, path, src, 0)
	if err != nil {
		return false, fmt.Errorf("parse %s: %w", path, err)
	}
	found := false
	ast.Inspect(f, func(n ast.Node) bool {
		if found {
			return false
		}
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		ident, ok := sel.X.(*ast.Ident)
		if !ok || ident.Name != "problem" {
			return true
		}
		if _, ok := registerFuncs[sel.Sel.Name]; ok {
			found = true
			return false
		}
		return true
	})
	return found, nil
}
