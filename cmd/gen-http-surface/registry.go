package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strconv"
)

// loadCapabilityRegistry reads the Capability constants declared in
// internal/modules/iam/domain/model.go and returns wire string -> Go constant
// name (e.g. "metrics.view" -> "CapMetricsView"). It parses the source
// directly rather than importing the iam package, because a Go constant's
// value carries no reflectable link back to its declared name; the source is
// the only place both are visible together, and it is already documented
// there as the single source of truth (ADR 0022).
func loadCapabilityRegistry(path string) (map[string]string, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, path, nil, 0)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}

	registry := map[string]string{}
	for _, decl := range file.Decls {
		gen, ok := decl.(*ast.GenDecl)
		if !ok || gen.Tok != token.CONST {
			continue
		}
		for _, spec := range gen.Specs {
			vs, ok := spec.(*ast.ValueSpec)
			if !ok {
				continue
			}
			ident, ok := vs.Type.(*ast.Ident)
			if !ok || ident.Name != "Capability" {
				continue
			}
			for i, name := range vs.Names {
				if i >= len(vs.Values) {
					continue
				}
				lit, ok := vs.Values[i].(*ast.BasicLit)
				if !ok || lit.Kind != token.STRING {
					continue
				}
				wire, err := strconv.Unquote(lit.Value)
				if err != nil {
					continue
				}
				registry[wire] = name.Name
			}
		}
	}
	return registry, nil
}
