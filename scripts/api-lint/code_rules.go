package main

// Code-side lints are intentionally single-file scans; they do not build a call graph.

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"unicode"

	"gopkg.in/yaml.v3"
)

func RunCodeRules(specPath, modulesRoot string) ([]Violation, error) {
	if modulesRoot == "" {
		return nil, nil
	}

	data, err := os.ReadFile(specPath)
	if err != nil {
		return nil, err
	}

	var doc yaml.Node
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return nil, err
	}

	root := doc.Content[0]
	paths := mapGet(root, "paths")
	if paths == nil {
		return nil, fmt.Errorf("%s: missing paths", specPath)
	}

	fset := token.NewFileSet()
	index, err := indexModuleFuncs(modulesRoot, fset)
	if err != nil {
		return nil, err
	}

	out := []Violation{}
	for i := 0; i+1 < len(paths.Content); i += 2 {
		pathVal := paths.Content[i+1]
		for j := 0; j+1 < len(pathVal.Content); j += 2 {
			op := pathVal.Content[j+1]
			opID := scalarValue(mapGet(op, "operationId"))
			if opID == "" {
				continue
			}
			out = append(out, checkAuthzCallPresent(specPath, opID, op, index, fset)...)
		}
	}

	tripwire, err := checkTripwirePairing(modulesRoot, fset)
	if err != nil {
		return nil, err
	}
	out = append(out, tripwire...)

	return out, nil
}

type indexedFunc struct {
	File string
	Decl *ast.FuncDecl
}

func indexModuleFuncs(modulesRoot string, fset *token.FileSet) (map[string][]indexedFunc, error) {
	out := map[string][]indexedFunc{}
	err := filepath.WalkDir(modulesRoot, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(strings.ToLower(path), ".go") || strings.HasSuffix(strings.ToLower(path), "_test.go") {
			return nil
		}
		file, err := parser.ParseFile(fset, path, nil, parser.SkipObjectResolution)
		if err != nil {
			return err
		}
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Recv == nil {
				continue
			}
			out[fn.Name.Name] = append(out[fn.Name.Name], indexedFunc{File: path, Decl: fn})
		}
		return nil
	})
	return out, err
}

func checkAuthzCallPresent(specPath, opID string, op *yaml.Node, index map[string][]indexedFunc, fset *token.FileSet) []Violation {
	// Gate: rule applies only when the op declares x-authz-area or x-authz-custom: true.
	// x-authz-skip-area silences the rule explicitly (also implies the gate is open).
	if mapGet(op, "x-authz-skip-area") != nil {
		return nil
	}
	hasArea := mapGet(op, "x-authz-area") != nil
	custom := strings.EqualFold(scalarValue(mapGet(op, "x-authz-custom")), "true")
	if !hasArea && !custom {
		return nil
	}

	handlerName := pascalCase(opID)
	candidates := index[handlerName]
	if len(candidates) == 0 {
		return []Violation{{
			File:    specPath,
			Line:    op.Line,
			Rule:    "authz-call-present",
			Message: fmt.Sprintf("handler %s not found for operation %s", handlerName, opID),
		}}
	}
	if len(candidates) > 1 {
		fmt.Printf("warning: multiple handler matches for %s (%s), using first\n", opID, handlerName)
	}

	fn := candidates[0].Decl
	expectedFn, expectedExpr := expectedAuthzCall(opID, op)
	hasAny, hasExpectedFn, matchedArg, actualArg := inspectAuthzCalls(fn, fset, expectedFn, expectedExpr)

	if custom {
		if !hasAny {
			return []Violation{{
				File:    candidates[0].File,
				Line:    fset.Position(fn.Pos()).Line,
				Rule:    "authz-call-present",
				Message: fmt.Sprintf("handler %s does not call authz.Require or authz.RequireAll", handlerName),
			}}
		}
		return nil
	}

	if !hasExpectedFn {
		return []Violation{{
			File:    candidates[0].File,
			Line:    fset.Position(fn.Pos()).Line,
			Rule:    "authz-call-present",
			Message: fmt.Sprintf("handler %s does not call authz.%s", handlerName, expectedFn),
		}}
	}

	if matchedArg {
		return nil
	}

	if actualArg == "" {
		actualArg = "<missing>"
	}
	return []Violation{{
		File:    candidates[0].File,
		Line:    fset.Position(fn.Pos()).Line,
		Rule:    "authz-call-present",
		Message: fmt.Sprintf("handler %s calls authz.%s with arg %s; expected %s", handlerName, expectedFn, actualArg, expectedExpr),
	}}
}

func inspectAuthzCalls(fn *ast.FuncDecl, fset *token.FileSet, expectedFn, expectedExpr string) (hasAny, hasExpectedFn, matchedArg bool, actualArg string) {
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		x, ok := sel.X.(*ast.Ident)
		if !ok || x.Name != "authz" {
			return true
		}
		if sel.Sel.Name != "Require" && sel.Sel.Name != "RequireAll" {
			return true
		}
		hasAny = true
		if sel.Sel.Name != expectedFn {
			return true
		}
		hasExpectedFn = true
		if len(call.Args) == 0 {
			if actualArg == "" {
				actualArg = "<missing>"
			}
			return true
		}
		arg := renderExpr(fset, call.Args[len(call.Args)-1])
		if arg == expectedExpr {
			matchedArg = true
			return false
		}
		if actualArg == "" {
			actualArg = arg
		}
		return true
	})
	return hasAny, hasExpectedFn, matchedArg, actualArg
}

func expectedAuthzCall(opID string, op *yaml.Node) (fnName, expr string) {
	area := mapGet(op, "x-authz-area")
	source := strings.ToLower(strings.TrimSpace(scalarValue(mapGet(area, "source"))))
	if source == "" {
		source = "body"
	}
	field := strings.TrimSpace(scalarValue(mapGet(area, "field")))
	multi := strings.EqualFold(scalarValue(mapGet(area, "multi")), "true")

	fnName = "Require"
	if multi {
		fnName = "RequireAll"
	}

	if source == "path" {
		expr = "req." + pascalCase(opID) + "Params"
		if field != "" {
			expr += "." + pascalCasePath(field)
		}
		return fnName, expr
	}

	expr = "req.Body"
	if field != "" {
		expr += "." + pascalCasePath(field)
	}
	return fnName, expr
}

func pascalCasePath(path string) string {
	parts := strings.Split(path, ".")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if strings.TrimSpace(p) == "" {
			continue
		}
		out = append(out, pascalCase(p))
	}
	return strings.Join(out, ".")
}

// pascalCase mirrors oapi-codegen's method-name derivation for operation IDs.
//
// Behaviour:
//   - If the input already has no separators (`_`, `-`, `.`), preserve internal
//     case and just uppercase the leading rune. This matches camelCase op-ids
//     like `listTemplatesV2` → `ListTemplatesV2` and initialism-rich names like
//     `recordMDDMShadowDiff` → `RecordMDDMShadowDiff`.
//   - If the input has separators (e.g. `area_code`, `zone.area_code`), split,
//     lowercase each segment, then uppercase the first rune of each.
//
// TODO(initialisms): for snake_case fields like `template_id`, real codegen
// can emit `TemplateID` rather than `TemplateId` depending on initialism
// config. Revisit once a real handler trips this.
func pascalCase(s string) string {
	if s == "" {
		return ""
	}
	if !strings.ContainsAny(s, "_-.") {
		r := []rune(s)
		r[0] = unicode.ToUpper(r[0])
		return string(r)
	}
	parts := strings.FieldsFunc(s, func(r rune) bool {
		return r == '_' || r == '-' || r == '.'
	})
	var b strings.Builder
	for _, p := range parts {
		if p == "" {
			continue
		}
		r := []rune(strings.ToLower(p))
		r[0] = unicode.ToUpper(r[0])
		b.WriteString(string(r))
	}
	return b.String()
}

func renderExpr(fset *token.FileSet, expr ast.Expr) string {
	var b bytes.Buffer
	_ = printer.Fprint(&b, fset, expr)
	return b.String()
}

func checkTripwirePairing(modulesRoot string, fset *token.FileSet) ([]Violation, error) {
	out := []Violation{}
	err := filepath.WalkDir(modulesRoot, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}
		name := strings.ToLower(filepath.Base(path))
		if !strings.Contains(name, "repository") {
			return nil
		}
		if !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
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
			hasAuthz, hasMutatingSQL := scanTripwireFunc(fn)
			if hasMutatingSQL && !hasAuthz {
				out = append(out, Violation{
					File:    path,
					Line:    fset.Position(fn.Pos()).Line,
					Rule:    "tripwire-pairing",
					Message: fmt.Sprintf("mutating SQL in %s without authz.Require call", fn.Name.Name),
				})
			}
		}
		return nil
	})
	return out, err
}

func scanTripwireFunc(fn *ast.FuncDecl) (hasAuthz, hasMutatingSQL bool) {
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		if x, ok := sel.X.(*ast.Ident); ok && x.Name == "authz" && (sel.Sel.Name == "Require" || sel.Sel.Name == "RequireAll") {
			hasAuthz = true
		}
		// db.Exec(sql, args...)            → SQL is Args[0]
		// db.ExecContext(ctx, sql, args...)→ SQL is Args[1]
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
		s := strings.ToUpper(sqlText)
		if strings.Contains(s, "INSERT INTO") || strings.Contains(s, "UPDATE") || strings.Contains(s, "DELETE FROM") {
			hasMutatingSQL = true
		}
		return true
	})
	return hasAuthz, hasMutatingSQL
}
